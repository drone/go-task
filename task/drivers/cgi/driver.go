// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgi

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/drone/go-task/task/logger"
	"github.com/drone/go-task/task/packaged"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/builder"
	"github.com/drone/go-task/task/downloader"
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
}

// New returns the task execution driver.
func New(d downloader.Downloader, pl packaged.PackageLoader) task.Handler {
	return &driver{downloader: d, packageLoader: pl}
}

type driver struct {
	downloader    downloader.Downloader
	packageLoader packaged.PackageLoader
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

	path, err := d.prepareArtifact(ctx, req.Task.Type, conf)
	if err != nil {
		log.WithError(err).Error("Prepare artifact failed")
		return task.Error(err)
	}

	setDefaultConfigValues(conf)

	binPath, err := d.getBinaryPath(ctx, path, conf)
	if err != nil {
		log.WithError(err).Error("task build failed")
		return task.Error(err)
	}

	execer := newExecer(binPath, conf)
	resp, err := execer.Exec(ctx, req.Task.Data)
	if err != nil {
		log.WithError(err).Error("could not execute cgi task")
		return task.Error(err)
	}

	return task.Respond(resp)
}

func (d *driver) prepareArtifact(ctx context.Context, taskType string, conf *Config) (string, error) {
	// use binary artifact, packaged or downloaded
	if conf.ExecutableConfig != nil {
		cgiPath, err := d.packageLoader.GetPackagePath(ctx, taskType, conf.ExecutableConfig)
		if err != nil {
			return d.downloader.DownloadExecutable(ctx, taskType, conf.ExecutableConfig)
		} else {
			log := logger.FromContext(ctx)
			log.WithField("path", cgiPath).Info("using prepackaged binary")
			return cgiPath, nil
		}
	} else if conf.Repository != nil {
		return d.downloader.DownloadRepo(ctx, conf.Repository)
	} else {
		return "", errors.New("no executable or repository provided")
	}
}

func shouldUsePrepackagedBinary(conf *Config) bool {
	return len(conf.ExecutableConfig.Executables) == 0
}

func setDefaultConfigValues(conf *Config) {
	if conf.Method == "" {
		conf.Method = "POST"
	}
	if conf.Endpoint == "" {
		conf.Endpoint = "/"
	}
	// Always set RUN_AS_CGI=true for the CGI server process.
	// In case the cgi's binary has other modes for running
	// (like starting an HTTP or gRPC server) the cgi's application
	// can use this environment variable to decide if it needs to
	// start a CGI server.
	conf.Envs = append(conf.Envs, "RUN_AS_CGI=true")
}

func (d *driver) getBinaryPath(ctx context.Context, path string, conf *Config) (string, error) {
	if conf.ExecutableConfig != nil {
		return path, nil
	}

	builder := builder.New(filepath.Join(path, taskYmlPath))
	return builder.Build(ctx)
}
