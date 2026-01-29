// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/cloner"
	"github.com/drone/go-task/task/logger"
	"github.com/mholt/archives"
)

type repoDownloader struct {
	cloner cloner.Cloner
}

// repoDownloader downloads a repository
// It also takes care of where to download the repository
func newRepoDownloader(cloner cloner.Cloner) *repoDownloader {
	return &repoDownloader{cloner: cloner}
}

func (r *repoDownloader) download(ctx context.Context, dir string, repo *task.Repository) (string, error) {
	if repo == nil {
		return "", errors.New("no repository provided to download")
	}
	dest := r.getDownloadDir(dir, repo)

	if cacheHit := isCacheHitFn(ctx, dest); cacheHit {
		// exit if the destination already exists
		return dest, nil
	}
	if repo.Download != "" {
		return dest, r.downloadRepo(ctx, repo, dest)
	}
	return dest, r.clone(ctx, repo, dest)
}

func (r *repoDownloader) clone(ctx context.Context, repo *task.Repository, dest string) error {

	// extract the clone url, ref and sha
	url := repo.Clone
	ref := repo.Ref
	sha := repo.Sha

	log := logger.FromContext(ctx).
		WithFields(map[string]interface{}{
			"source":   url,
			"revision": ref,
			"sha":      sha,
			"target":   dest,
		})

	log.Debug("clone artifact")

	// clone the repository
	err := r.cloner.Clone(ctx, cloner.Params{
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

func (r *repoDownloader) downloadRepo(ctx context.Context, repo *task.Repository, destDir string) error {

	dest := getDownloadPath(repo.Download, destDir)
	downloadPath, err := downloadFile(ctx, []string{repo.Download}, dest)
	if err != nil {
		// remove the destination directory if downloading fails so it can be retried
		os.RemoveAll(dest)
		return err
	}

	if err := r.unarchive(downloadPath, dest); err != nil {
		// remove the destination directory if unarchiving fails so it can be retried
		os.RemoveAll(dest)
		return err
	}

	log := logger.FromContext(ctx).
		WithFields(map[string]interface{}{
			"source":      repo.Download,
			"destination": dest,
		})

	log.Debug("extracted artifact")

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
func (r *repoDownloader) unarchive(srcPath, destDir string) error {
	ctx := context.Background()

	// Create a virtual filesystem from the archive
	fsys, err := archives.FileSystem(ctx, srcPath, nil)
	if err != nil {
		return fmt.Errorf("error opening archive: %w", err)
	}

	// Walk through all files in the archive and extract them
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip root directory and directories themselves
		if path == "." || d.IsDir() {
			return nil
		}

		// Split the path into components
		pathComponents := strings.Split(path, "/")

		// Skip the top-level directory (e.g., "myrepo/task.yml" -> "task.yml")
		var relPath string
		if len(pathComponents) > 1 {
			relPath = filepath.Join(pathComponents[1:]...)
		} else {
			relPath = path
		}

		// Construct target file path
		targetFile := filepath.Join(destDir, relPath)

		// Path traversal protection: validate target is within destination
		if !strings.HasPrefix(filepath.Clean(targetFile), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", path)
		}

		// Ensure the directory structure exists
		if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
			return fmt.Errorf("error creating directories: %w", err)
		}

		// Open source file from archive
		srcFile, err := fsys.Open(path)
		if err != nil {
			return fmt.Errorf("error opening file in archive: %w", err)
		}
		defer srcFile.Close()

		// Create the target file
		outFile, err := os.Create(targetFile)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		defer outFile.Close()

		// Copy the contents from the archive to the new file
		if _, err = io.Copy(outFile, srcFile); err != nil {
			return fmt.Errorf("error copying file contents: %w", err)
		}

		return nil
	})
}

// getDownloadDir returns the directory where the repository should be downloaded
// It joins the top-level directory with the hash of the repository config
func (r *repoDownloader) getDownloadDir(dir string, repo *task.Repository) string {
	return filepath.Join(dir, r.getHashOfRepo(repo))
}

// getHashOfRepo constructs a hash from the repo config to figure out
// whether it should be re-cloned.
func (r *repoDownloader) getHashOfRepo(repo *task.Repository) string {
	data := fmt.Sprintf("%s|%s|%s|%s", repo.Clone, repo.Ref, repo.Sha, repo.Download)
	return getHash(data)
}
