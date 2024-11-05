// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package downloader

import (
	"context"
	"path/filepath"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/cloner"
)

// Downloader is an interface for structs
// that handle downloading a task's implementation

type Downloader struct {
	dir                  string
	repoDownloader       *repoDownloader
	executableDownloader *executableDownloader
}

func New(cloner cloner.Cloner, dir string) Downloader {
	baseDir := getBaseDownloadDir(dir)
	repoDownloader := newRepoDownloader(cloner)
	executableDownloader := newExecutableDownloader()
	return Downloader{dir: baseDir, repoDownloader: repoDownloader, executableDownloader: executableDownloader}
}

// getBaseDownloadDir returns the top-level directory where all files should be downloaded
func getBaseDownloadDir(dir string) string {
	return filepath.Join(dir, ".harness", "cache")
}

func (d *Downloader) DownloadRepo(ctx context.Context, repo *task.Repository) (string, error) {
	return d.repoDownloader.download(ctx, d.dir, repo)
}

func (d *Downloader) DownloadExecutable(ctx context.Context, taskType string, exec *task.ExecutableConfig) (string, error) {
	return d.executableDownloader.download(ctx, d.dir, taskType, exec)
}
