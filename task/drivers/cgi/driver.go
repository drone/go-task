// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgi

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/download"
	"github.com/drone/go-task/task/logger"
)

var (
	taskYmlPath = "task.yml"
)

// Config provides the driver config.
type Config struct {
	Repository *task.Repository  `json:"repository"`
	Method     string            `json:"method"`
	Endpoint   string            `json:"endpoint"`
	Headers    map[string]string `json:"headers"`
	Envs       []string          `json:"envs"`
}

// New returns the task execution driver.
func New(d download.Downloader) task.Handler {
	return &driver{downloader: d}
}

type driver struct {
	downloader download.Downloader
}

// Handle handles the task execution request.
func (d *driver) Handle(ctx context.Context, req *task.Request) task.Response {
	var (
		log  = logger.FromContext(ctx)
		conf = new(Config)
	)

	// decode the task configuration
	err := json.Unmarshal(req.Task.Config, conf)
	if err != nil {
		return task.Error(err)
	}

	path, err := d.downloader.Download(ctx, conf.Repository)
	if err != nil {
		log.With("error", err).Error("artifact download failed")
		return task.Error(err)
	}

	if conf.Method == "" {
		conf.Method = "POST"
	}

	if conf.Endpoint == "" {
		conf.Endpoint = "/"
	}

	execer := newExecer(filepath.Join(path, taskYmlPath), conf)
	resp, err := execer.Exec(ctx, req.Task.Data)
	if err != nil {
		log.With("error", err).Error("could not execute cgi task")
		return task.Error(err)
	}

	return task.Respond(resp)
}
