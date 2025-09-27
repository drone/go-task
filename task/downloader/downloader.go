// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package downloader

import (
	"context"

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
	repoDownloader := newRepoDownloader(cloner)
	executableDownloader := newExecutableDownloader()
	return Downloader{dir: dir, repoDownloader: repoDownloader, executableDownloader: executableDownloader}
}

func (d *Downloader) DownloadRepo(ctx context.Context, repo *task.Repository) (string, error) {
	return d.repoDownloader.download(ctx, d.dir, repo)
}

func (d *Downloader) DownloadExecutable(ctx context.Context, taskType string, exec *task.ExecutableConfig, fallbackEnabled bool) (string, error) {
	return d.executableDownloader.download(ctx, d.dir, taskType, exec, fallbackEnabled)
}
