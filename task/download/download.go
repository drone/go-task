// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package download

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/cloner"
	"github.com/drone/go-task/task/logger"
)

// Artifact provides the artifact used for custom
// task execution. It downloads the given source repository
// into the destination path.
type Artifact struct {
	Source      *task.Repository  `json:"source,omitempty"`
	Destination string            `json:"destination,omitempty"`
	Checksum    string            `json:"checksum,omitempty"`
	Insecure    bool              `json:"insecure,omitempty"`
	Header      map[string]string `json:"header,omitempty"`
}

// Downloader downloads a task artifact.
type Downloader interface {
	Download(context.Context, *Artifact) error
}

// New returns a downloader using the default
// system cache directory
func New(cloner cloner.Cloner) Downloader {
	return &downloader{cloner: cloner}
}

type downloader struct {
	cloner cloner.Cloner
}

func (d *downloader) Download(ctx context.Context, artifact *Artifact) error {
	if artifact.Source != nil {
		if artifact.Source.Download != "" {
			return d.download(ctx, artifact)
		}
		return d.clone(ctx, artifact)
	}
	return nil
}

func (d *downloader) clone(ctx context.Context, artifact *Artifact) error {
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

func (d *downloader) download(_ context.Context, artifact *Artifact) error {
	// get the target download location
	dest := artifact.Destination

	// exit if the artifact already exists
	if _, err := os.Stat(dest); err == nil {
		return nil
	}

	// create the directory where the target is downloaded.
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}

	// unpack the .tar.gz to the target directory
	return downloadAndUnpackTarGz(artifact.Source.Download, artifact.Destination)
}

// downloadAndUnpackTarGz downloads a .tar.gz file from a URL and unpacks it to a destination directory
func downloadAndUnpackTarGz(url, destDir string) error {
	// download the .tar.gz file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if code := resp.StatusCode; code > 299 {
		return fmt.Errorf("download error with status code %d", code)
	}

	// create a gzip reader
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	if gzr != nil {
		defer gzr.Close()
	}

	// create a tar reader
	tr := tar.NewReader(gzr)

	// unpack the tar archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // end of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Determine the proper path for the file/directory
		targetPath := filepath.Join(destDir, header.Name)

		// Check the type of the header
		switch header.Typeflag {
		case tar.TypeDir:
			// create directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// create file
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for file: %w", err)
			}
			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		default:
			return fmt.Errorf("unsupported tar entry type: %v", header.Typeflag)
		}
	}

	return nil
}
