// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cgi provides a cgi execution driver.

package cgi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/download"
	"github.com/drone/go-task/task/logger"
)

// Config provides the driver config.
type Config struct {
	Artifact *task.Artifact    `json:"artifact"`
	Method   string            `json:"method"`
	Endpoint string            `json:"endpoint"`
	Headers  map[string]string `json:"headers"`
	Path     string            `json:"path"`
	Args     []string          `json:"args"`
	Envs     []string          `json:"envs"`
	Dir      string            `json:"dir"`
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
		log    = logger.FromContext(ctx)
		conf   = new(Config)
		method = "POST"
		url    = "/"
	)

	// decode the task configuration
	err := json.Unmarshal(req.Task.Config, conf)
	if err != nil {
		return task.Error(err)
	}

	// download the artifact if provided (and if needed)
	if conf.Artifact != nil {
		if err := d.downloader.Download(ctx, conf.Artifact); err != nil {
			log.Error("artifact download failed: %s", err)
			return task.Error(err)
		}
	}

	// set the workdir if needed
	dir := conf.Dir
	if dir == "" {
		dir, _ = os.Getwd()
	}

	// replace XDG_CACHE_HOME if needed
	// lookup the path if needed
	path := expandCache(conf.Path)
	if p, err := exec.LookPath(path); err == nil {
		path = p
	}

	// replace XDG_CACHE_HOME if needed
	args := expandCacheSlice(conf.Args)

	// create the cgi handler
	handler := cgi.Handler{
		Dir:  dir,
		Path: path,
		Args: args,
		Env:  append(conf.Envs, os.Environ()...), // is this needed?

		Logger: nil, // TODO support optional logger
		Stderr: nil, // TODO support optional stderr
	}

	if conf.Method != "" {
		method = conf.Method
	}

	if conf.Endpoint != "" {
		url = conf.Endpoint
	}

	// create the cgi request
	r, err := http.NewRequestWithContext(ctx, method, url,
		bytes.NewReader(req.Task.Data),
	)
	if err != nil {
		log.Error("cannot invoke task: %s", err)
		return task.Error(err)
	}

	for key, value := range conf.Headers {
		r.Header.Set(key, value)
	}

	// create ethe cgi response
	rw := httptest.NewRecorder()

	log.With("cgi.dir", dir).
		With("cgi.path", path).
		With("cgi.args", conf.Args).
		With("cgi.method", method).
		With("cgi.url", url).
		Debug("invoke cgi task")

	// execute the request
	handler.ServeHTTP(rw, r)

	// check the error code and write the error
	// to the context, if applicable.
	// TODO should we unmarshal the response body to an error type?
	if code := rw.Code; code > 299 {
		return task.Errorf("received error code %d", code)
	}

	return task.Respond(
		rw.Body.Bytes(),
	)
}
