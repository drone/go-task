// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package cgi

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/drone/go-task/task/logger"
)

type Execer struct {
	Binpath   string  // path to the binary file for execution
	CGIConfig *Config // config for the cgi execution
}

func newExecer(binpath string, cgiConfig *Config) *Execer {
	return &Execer{
		Binpath:   binpath,
		CGIConfig: cgiConfig,
	}
}

// Exec parses the task.yml file and executes the task given the input
func (e *Execer) Exec(ctx context.Context, in []byte) ([]byte, error) {
	log := logger.FromContext(ctx)
	conf := e.CGIConfig

	// run the task using cgi
	handler := cgi.Handler{
		Path: e.Binpath,
		Dir:  filepath.Dir(e.Binpath),
		Env:  append(conf.Envs, os.Environ()...), // is this needed?

		Logger: nil, // TODO support optional logger
		Stderr: nil, // TODO support optional stderr
	}

	r, err := http.NewRequestWithContext(ctx, conf.Method, conf.Endpoint,
		bytes.NewReader(in),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot invoke task: %s", err)
	}

	for key, value := range conf.Headers {
		r.Header.Set(key, value)
	}

	// create ethe cgi response
	rw := httptest.NewRecorder()

	log.With("cgi.dir", filepath.Dir(e.Binpath)).
		With("cgi.path", e.Binpath).
		With("cgi.method", conf.Method).
		With("cgi.url", conf.Endpoint).
		Debug("invoke cgi task")

	// execute the request
	handler.ServeHTTP(rw, r)

	// check the error code and write the error
	// to the context, if applicable.
	// TODO should we unmarshal the response body to an error type?
	if code := rw.Code; code > 299 {
		return nil, fmt.Errorf("received error code %d", code)
	}

	return rw.Body.Bytes(), nil
}
