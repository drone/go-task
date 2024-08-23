// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/cloner"
	"github.com/drone/go-task/task/logger"
	"github.com/mholt/archiver"
)

// Downloader downloads a repository
// It also takes care of where to download the repository
type Downloader interface {
	// returns back the download directory
	Download(context.Context, *task.Repository, *task.ExecutableUrls) (string, error)
}

// New returns a downloader which downloads everything at the top-level
// dir directory
func New(cloner cloner.Cloner, dir string) Downloader {
	return &downloader{cloner: cloner, dir: dir}
}

type downloader struct {
	cloner cloner.Cloner
	dir    string // top-level cache directory
}

// getHashOfRepo constructs a hash from the repo config to figure out
// whether it should be re-cloned.
func getHashOfRepo(repo *task.Repository) string {
	data := fmt.Sprintf("%s|%s|%s|%s", repo.Clone, repo.Ref, repo.Sha, repo.Download)
	return getHash(data)
}

func getHash(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}

func (d *downloader) Download(ctx context.Context, repo *task.Repository, urls *task.ExecutableUrls) (string, error) {
	if urls != nil {
		return d.handleDownloadExecutable(ctx, urls)
	} else if repo != nil {
		return d.handleDownloadRepo(ctx, repo)
	}
	return "", errors.New("no repository or executable urls provided to download")
}

func (d *downloader) handleDownloadExecutable(ctx context.Context, urls *task.ExecutableUrls) (string, error) {
	osWithArch := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	url, ok := (*urls)[osWithArch]
	if !ok {
		return "", fmt.Errorf("os and architecture [%s] is not specified in task.yml file", osWithArch)
	}

	dest := filepath.Join(d.getBaseDownloadDir(), getHash(url))

	if cacheHit := d.isCacheHit(ctx, dest); cacheHit {
		// exit if the artifact destination already exists
		return d.getDownloadPath(url, dest), nil
	}

	binpath, err := d.downloadFile(ctx, url, dest)
	if err != nil {
		// remove the destination directory if downloading fails so it can be retried
		os.RemoveAll(dest)
		return "", err
	}

	err = os.Chmod(binpath, 0777)
	if err != nil {
		return "", fmt.Errorf("failed to set executable flag in task file [%s]: %w", binpath, err)
	}
	return binpath, nil
}

func (d *downloader) handleDownloadRepo(ctx context.Context, repo *task.Repository) (string, error) {
	dest := d.getDownloadDir(repo)

	if cacheHit := d.isCacheHit(ctx, dest); cacheHit {
		// exit if the artifact destination already exists
		return dest, nil
	}

	if repo.Download != "" {
		return dest, d.downloadRepo(ctx, repo, dest)
	}
	return dest, d.clone(ctx, repo, dest)
}

func (d *downloader) clone(ctx context.Context, repo *task.Repository, dest string) error {
	log := logger.FromContext(ctx)

	// extract the clone url, ref and sha
	url := repo.Clone
	ref := repo.Ref
	sha := repo.Sha

	log.With("source", url).
		With("revision", ref).
		With("sha", sha).
		With("target", dest).
		Debug("clone artifact")

	// clone the repository
	err := d.cloner.Clone(ctx, cloner.Params{
		Repo: url,
		Ref:  ref,
		Sha:  sha,
		Dir:  dest,
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *downloader) downloadRepo(ctx context.Context, repo *task.Repository, dest string) error {
	log := logger.FromContext(ctx)

	downloadPath, err := d.downloadFile(ctx, repo.Download, dest)
	if err != nil {
		// remove the destination directory if downloading fails so it can be retried
		os.RemoveAll(dest)
		return err
	}

	if err := d.unarchive(downloadPath, dest); err != nil {
		// remove the destination directory if unarchiving fails so it can be retried
		os.RemoveAll(dest)
		return err
	}

	log.With("source", repo.Download).
		With("destination", dest).
		Debug("extracted artifact")

	// delete the archive file after unpacking
	os.Remove(downloadPath)

	return nil
}

// unarchive unpacks srcPath into destDir. It unpacks everything directly into the
// destination directory and skips the top-level directory.
// For example, a github repo called "myrepo" with a file "task.yml" at the root
// will have an archive called "myrepo.zip" with the structure myrepo/task.yml.
// If destDir is "/tmp", this will extract the archive as /tmp/task.yml similar to the
// clone behavior.
func (d *downloader) unarchive(srcPath, destDir string) error {
	// create a custom walk function
	walkFn := func(f archiver.File) error {
		// skip directories
		if f.IsDir() {
			return nil
		}

		// get the relative path of the file within the archive
		relPath := f.Name()

		// split the path into components
		pathComponents := strings.Split(relPath, string(filepath.Separator))

		// if there's more than one component, remove the first one (top-level directory)
		if len(pathComponents) > 1 {
			relPath = filepath.Join(pathComponents[1:]...)
		}

		// construct the target file path
		targetFile := filepath.Join(destDir, relPath)

		// ensure the directory structure exists
		err := os.MkdirAll(filepath.Dir(targetFile), 0755)
		if err != nil {
			return fmt.Errorf("error creating directories: %w", err)
		}

		// create the target file
		outFile, err := os.Create(targetFile)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		defer outFile.Close()

		// copy the contents from the archive to the new file
		_, err = io.Copy(outFile, f)
		if err != nil {
			return fmt.Errorf("error copying file contents: %w", err)
		}

		return nil
	}

	// open and walk through the archive
	err := archiver.Walk(srcPath, walkFn)
	if err != nil {
		return fmt.Errorf("error walking through archive: %w", err)
	}

	return nil
}

// downloadFile fetches the file from url and writes it to dest
func (d *downloader) downloadFile(ctx context.Context, url, dest string) (string, error) {
	log := logger.FromContext(ctx)

	log.With("source", url).
		With("destination", dest).
		Debug("downloading artifact")

	// create the directory where the target is downloaded.
	if err := os.MkdirAll(dest, 0777); err != nil {
		return "", err
	}

	// determine the file name and download location
	downloadPath := d.getDownloadPath(url, dest)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code > 299 {
		return "", fmt.Errorf("download error with status code %d", code)
	}

	outFile, err := os.Create(downloadPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %w", err)
	}

	log.With("source", url).
		With("destination", downloadPath).
		Debug("downloaded artifact")

	return downloadPath, nil
}

// getDownloadPath returns the full download path given the download url and the destination folder `dest`
func (d *downloader) getDownloadPath(url, dest string) string {
	fileName := filepath.Base(url)
	return filepath.Join(dest, fileName)
}

// getDownloadDir returns the directory where the repository should be downloaded
// It joins the top-level directory with the hash of the repository config
func (d *downloader) getDownloadDir(repo *task.Repository) string {
	return filepath.Join(d.getBaseDownloadDir(), getHashOfRepo(repo))
}

// getBaseDownloadDir returns the top-level directory where all files should be downloaded
func (d *downloader) getBaseDownloadDir() string {
	return filepath.Join(d.dir, ".harness", "cache")
}

func (d *downloader) isCacheHit(ctx context.Context, dest string) bool {
	log := logger.FromContext(ctx)

	if _, err := os.Stat(dest); err == nil {
		log.With("target", dest).
			Debug("cache hit")
		return true
	}

	log.With("target", dest).
		Debug("cache miss")
	return false
}
