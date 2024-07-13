// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/cloner"
	"github.com/drone/go-task/task/logger"
	"github.com/mholt/archiver"
)

// Downloader downloads a task artifact.
type Downloader interface {
	Download(context.Context, *task.Artifact) error
}

// New returns a downloader using the default
// system cache directory
func New(cloner cloner.Cloner) Downloader {
	return &downloader{cloner: cloner}
}

type downloader struct {
	cloner cloner.Cloner
}

func (d *downloader) Download(ctx context.Context, artifact *task.Artifact) error {
	if artifact.Source != nil {
		if artifact.Source.Download != "" {
			return d.download(ctx, artifact)
		}
		return d.clone(ctx, artifact)
	}
	return nil
}

func (d *downloader) clone(ctx context.Context, artifact *task.Artifact) error {
	log := logger.FromContext(ctx)

	dir := artifact.Destination
	// exit if the artifact destination already exists
	if _, err := os.Stat(dir); err == nil {
		log.With("target", dir).
			Debug("cache hit")
		return nil
	} else {
		log.With("target", dir).
			Debug("cache miss")
	}

	// extract the clone url, ref and sha
	url := artifact.Source.Clone
	ref := artifact.Source.Ref
	sha := artifact.Source.Sha

	log.With("source", url).
		With("revision", ref).
		With("sha", sha).
		With("target", dir).
		Debug("clone artifact")

	// clone the repository
	err := d.cloner.Clone(ctx, cloner.Params{
		Repo: url,
		Ref:  ref,
		Sha:  sha,
		Dir:  dir,
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *downloader) download(ctx context.Context, artifact *task.Artifact) error {
	log := logger.FromContext(ctx)

	// get the target download location
	dest := artifact.Destination

	// exit if the directory already exists
	if _, err := os.Stat(dest); err == nil {
		log.With("target", dest).
			Debug("cache hit")
		return nil
	} else {
		log.With("target", dest).
			Debug("cache miss")
	}

	// create the directory where the target is downloaded.
	if err := os.MkdirAll(dest, 0777); err != nil {
		return err
	}

	// determine the file name and download location
	fileName := filepath.Base(artifact.Source.Download)
	downloadPath := filepath.Join(artifact.Destination, fileName)

	if err := downloadFile(artifact.Source.Download, downloadPath); err != nil {
		// remove the destination directory if downloading fails so it can be retried
		os.RemoveAll(artifact.Destination)
		return err
	}

	log.With("source", artifact.Source.Download).
		With("destination", artifact.Destination).
		Debug("downloaded artifact")

	if err := d.unarchive(downloadPath, artifact.Destination); err != nil {
		// remove the destination directory if unarchiving fails so it can be retried
		os.RemoveAll(artifact.Destination)
		return err
	}

	log.With("source", artifact.Source.Download).
		With("destination", artifact.Destination).
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
func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code > 299 {
		return fmt.Errorf("download error with status code %d", code)
	}

	outFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
