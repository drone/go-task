// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package cgi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/drone/go-task/task"
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

// Exec executes the task given the binary filepath and the configuration
func (e *Execer) Exec(ctx context.Context, in []byte) (*task.CGITaskResponse, error) {
	conf := e.CGIConfig
	log := logger.FromContext(ctx).WithFields(map[string]interface{}{
		"cgi.dir":    filepath.Dir(e.Binpath),
		"cgi.path":   e.Binpath,
		"cgi.method": conf.Method,
		"cgi.url":    conf.Endpoint,
	})

	// stderrPipe is the reading end of the pipe that will capture anything written to stderr
	// stderrWriter is the writing end of that pipe
	// we pass stderrWriter to the CGI handler and read the output from stderrPipe
	stderrPipe, stderrWriter, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	defer stderrPipe.Close()

	handler := cgi.Handler{
		Path: e.Binpath,
		Dir:  filepath.Dir(e.Binpath),
		Env:  append(conf.Envs, os.Environ()...),
		// the CGI handler reserves the stdout for the HTTP response (technically the application response) and all log messages are supposed to be written to stderr
		// by default frameworks like logrus, slog writes to stderr
		Stderr: stderrWriter,
	}

	// prepare the HTTP request for the handler
	req, err := http.NewRequestWithContext(ctx, conf.Method, conf.Endpoint, bytes.NewReader(in))
	if err != nil {
		return nil, fmt.Errorf("cannot create CGI request: %w", err)
	}
	for key, value := range conf.Headers {
		req.Header.Set(key, value)
	}

	// record the CGI handler’s response
	responseRecorder := httptest.NewRecorder()
	log.Debug("Invoking CGI task")

	// Execute the request
	handler.ServeHTTP(responseRecorder, req)
	stderrWriter.Close()

	// capture stderr output
	var stderrBuf bytes.Buffer
	_, err = io.Copy(&stderrBuf, stderrPipe)
	if err != nil {
		return nil, fmt.Errorf("failed to read CGI output: %w", err)
	}

	log.Infof("Captured CGI logs: %s", stderrBuf.String())
	return &task.CGITaskResponse{StatusCode: responseRecorder.Code, Body: responseRecorder.Body.Bytes(), Headers: headerToMap(responseRecorder.Header())}, nil
}

func headerToMap(header http.Header) map[string][]string {
	headerMap := make(map[string][]string)
	for key, values := range header {
		headerMap[key] = values
	}
	return headerMap
}
