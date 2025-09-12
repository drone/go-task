// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/downloader"
	"github.com/drone/go-task/task/logger"
	"github.com/drone/go-task/task/packaged"
)


// Config provides the driver config.
type Config struct {
	ExecutableConfig *task.ExecutableConfig `json:"executable_config"`
	Envs             []string               `json:"envs"`
	WorkDir          string                 `json:"work_dir"`
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

	binPath, err := d.prepareArtifact(ctx, req.Task.Type, conf)
	if err != nil {
		log.WithError(err).Error("prepare artifact failed")
		return task.Error(err)
	}

	execer := newExecer(binPath, conf)
	if err := execer.Exec(ctx, req.Task.Data); err != nil {
		log.WithError(err).Error("binary execution failed")
		return task.Error(err)
	}

	return task.Respond(nil)
}

func (d *driver) prepareArtifact(ctx context.Context, taskType string, conf *Config) (string, error) {
	if conf.ExecutableConfig == nil {
		return "", errors.New("no executable config provided")
	}

	binPath, err := d.packageLoader.GetPackagePath(ctx, taskType, conf.ExecutableConfig)
	if err != nil {
		return d.downloader.DownloadExecutable(ctx, taskType, conf.ExecutableConfig)
	}

	log := logger.FromContext(ctx)
	log.WithField("path", binPath).Info("using prepackaged binary")
	return binPath, nil
}

