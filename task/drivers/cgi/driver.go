// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgi

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/builder"
	"github.com/drone/go-task/task/downloader"
	"github.com/drone/go-task/task/logger"
)

var (
	taskYmlPath = "task.yml"
)

// Config provides the driver config.
type Config struct {
	ExecutableConfig *task.ExecutableConfig `json:"executable_config"`
	Repository       *task.Repository       `json:"repository"`
	Method           string                 `json:"method"`
	Endpoint         string                 `json:"endpoint"`
	Headers          map[string]string      `json:"headers"`
	Envs             []string               `json:"envs"`
	Version          string                 `json:"version"`
}

// New returns the task execution driver.
func New(d downloader.Downloader) task.Handler {
	return &driver{downloader: d}
}

type driver struct {
	downloader downloader.Downloader
}

// Handle handles the task execution request.
func (d *driver) Handle(ctx context.Context, req *task.Request) task.Response {
	log := logger.FromContext(ctx)

	conf := new(Config)
	// decode the task configuration
	err := json.Unmarshal(req.Task.Config, conf)
	if err != nil {
		return task.Error(err)
	}

	path, err := d.downloadArtifact(ctx, req.Task.Type, conf)
	if err != nil {
		log.With("error", err).Error("artifact download failed")
		return task.Error(err)
	}

	setDefaultConfigValues(conf)

	binPath, err := d.getBinaryPath(ctx, path, conf)
	if err != nil {
		log.With("error", err).Error("task build failed")
		return task.Error(err)
	}

	execer := newExecer(binPath, conf)
	resp, err := execer.Exec(ctx, req.Task.Data)
	if err != nil {
		log.With("error", err).Error("could not execute cgi task")
		return task.Error(err)
	}

	return task.Respond(resp)
}

func (d *driver) downloadArtifact(ctx context.Context, taskType string, conf *Config) (string, error) {
	if conf.ExecutableConfig != nil {
		return d.downloader.DownloadExecutable(ctx, taskType, conf.Version, conf.ExecutableConfig)
	}
	return d.downloader.DownloadRepo(ctx, conf.Repository)
}

func setDefaultConfigValues(conf *Config) {
	if conf.Method == "" {
		conf.Method = "POST"
	}
	if conf.Endpoint == "" {
		conf.Endpoint = "/"
	}
}

func (d *driver) getBinaryPath(ctx context.Context, path string, conf *Config) (string, error) {
	if conf.ExecutableConfig != nil {
		return path, nil
	}

	builder := builder.New(filepath.Join(path, taskYmlPath))
	return builder.Build(ctx)
}
