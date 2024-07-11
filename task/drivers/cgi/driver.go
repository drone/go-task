// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cgi provides a cgi execution driver.

package cgi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/download"
	"github.com/drone/go-task/task/logger"
)

// Config provides the driver config.
type Config struct {
	Repository *task.Repository  `json:"repository"`
	Method     string            `json:"method"`
	Endpoint   string            `json:"endpoint"`
	Headers    map[string]string `json:"headers"`
	Envs       []string          `json:"envs"`
}

// getHashOfRepo constructs a hash from the repo config to figure out
// whether it should be re-cloned.
func getHashOfRepo(repo *task.Repository) string {
	data := fmt.Sprintf("%s|%s|%s|%s", repo.Clone, repo.Ref, repo.Sha, repo.Download)
	hash := sha256.New()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
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

	hash := getHashOfRepo(conf.Repository)
	cache, _ := getcache()
	path := filepath.Join(cache, ".harness", "cache", hash)
	_, err = os.Stat(path)
	if err != nil {
		// download the artifact to the destination directory
		artifact := &download.Artifact{
			Source:      conf.Repository,
			Destination: path,
		}
		if err := d.downloader.Download(ctx, artifact); err != nil {
			log.With("error", err).Error("artifact download failed")
			return task.Error(err)
		}
	}

	if conf.Method == "" {
		conf.Method = "POST"
	}

	if conf.Endpoint == "" {
		conf.Endpoint = "/"
	}

	execer := Execer{
		Source: path,
	}

	resp, err := execer.Exec(ctx, conf, req.Task.Data)
	if err != nil {
		log.With("error", err).Error("could not execute cgi task")
		return task.Error(err)
	}

	return task.Respond(resp)
}
