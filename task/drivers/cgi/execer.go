// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package cgi

import (
	"bytes"
	"cloud.google.com/go/logging"
	"context"
	"fmt"
	"github.com/drone/go-task/task/logger"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"io"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"
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
//func (e *Execer) Exec(ctx context.Context, in []byte) ([]byte, error) {
//	log := logger.FromContext(ctx)
//	conf := e.CGIConfig
//
//	// run the task using cgi
//	handler := cgi.Handler{
//		Path: e.Binpath,
//		Dir:  filepath.Dir(e.Binpath),
//		Env:  append(conf.Envs, os.Environ()...), // is this needed?
//
//		Logger: nil, // TODO support optional logger
//		Stderr: nil, // TODO support optional stderr
//	}
//
//	r, err := http.NewRequestWithContext(ctx, conf.Method, conf.Endpoint,
//		bytes.NewReader(in),
//	)
//	if err != nil {
//		return nil, fmt.Errorf("cannot invoke task: %s", err)
//	}
//
//	for key, value := range conf.Headers {
//		r.Header.Set(key, value)
//	}
//
//	// create ethe cgi response
//	rw := httptest.NewRecorder()
//
//	log.With("cgi.dir", filepath.Dir(e.Binpath)).
//		With("cgi.path", e.Binpath).
//		With("cgi.method", conf.Method).
//		With("cgi.url", conf.Endpoint).
//		Debug("invoke cgi task")
//
//	// execute the request
//	handler.ServeHTTP(rw, r)
//
//	// check the error code and write the error
//	// to the context, if applicable.
//	// TODO should we unmarshal the response body to an error type?
//	if code := rw.Code; code > 299 {
//		return nil, fmt.Errorf("received error code %d", code)
//	}
//
//	return rw.Body.Bytes(), nil
//}

//func (e *Execer) Exec(ctx context.Context, in []byte) ([]byte, error) {
//	conf := e.CGIConfig
//	log := logger.FromContext(ctx).WithFields(logrus.Fields{"cgi.dir": filepath.Dir(e.Binpath),
//		"cgi.path":   e.Binpath,
//		"cgi.method": conf.Method,
//		"cgi.url":    conf.Endpoint})
//
//	stderrPipe, stderrWriter, err := os.Pipe()
//	if err != nil {
//		return nil, fmt.Errorf("failed to create stderr pipe: %s", err)
//	}
//
//	handler := cgi.Handler{
//		Path: e.Binpath,
//		Dir:  filepath.Dir(e.Binpath),
//		Env:  append(conf.Envs, os.Environ()...),
//
//		Logger: nil,
//		Stderr: stderrWriter,
//	}
//
//	r, err := http.NewRequestWithContext(ctx, conf.Method, conf.Endpoint, bytes.NewReader(in))
//	if err != nil {
//		return nil, fmt.Errorf("cannot invoke task: %s", err)
//	}
//
//	for key, value := range conf.Headers {
//		r.Header.Set(key, value)
//	}
//
//	rw := httptest.NewRecorder()
//
//	log.Debug("invoke cgi task")
//
//	// execute the request
//	handler.ServeHTTP(rw, r)
//
//	// Close the writers after serving the request
//	err = stderrWriter.Close()
//	if err != nil {
//		return nil, err
//	}
//
//	var stderrBuf bytes.Buffer
//	stderrDone := make(chan error)
//
//	go func() {
//		_, err := io.Copy(&stderrBuf, stderrPipe)
//		stderrDone <- err
//	}()
//
//	<-stderrDone
//	// Wait for the response code and check for errors
//	if code := rw.Code; code > 299 {
//		return nil, fmt.Errorf("received error code %d", code)
//	}
//
//	log.Infof("Captured CGI output: %s", stderrBuf.String())
//
//	return rw.Body.Bytes(), nil
//}
//
//func (e *Execer) Exec(ctx context.Context, in []byte) ([]byte, error) {
//	conf := e.CGIConfig
//	log := logger.FromContext(ctx).WithFields(logrus.Fields{
//		"cgi.dir":    filepath.Dir(e.Binpath),
//		"cgi.path":   e.Binpath,
//		"cgi.method": conf.Method,
//		"cgi.url":    conf.Endpoint,
//	})
//
//	// Set up pipes for capturing stderr
//	stderrPipe, stderrWriter, err := os.Pipe()
//	if err != nil {
//		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
//	}
//	defer stderrPipe.Close()
//
//	// Set up the CGI handler
//	handler := cgi.Handler{
//		Path:   e.Binpath,
//		Dir:    filepath.Dir(e.Binpath),
//		Env:    append(conf.Envs, os.Environ()...),
//		Stderr: stderrWriter,
//	}
//
//	// Prepare the HTTP request for the CGI handler
//	req, err := http.NewRequestWithContext(ctx, conf.Method, conf.Endpoint, bytes.NewReader(in))
//	if err != nil {
//		return nil, fmt.Errorf("cannot create CGI request: %w", err)
//	}
//	for key, value := range conf.Headers {
//		req.Header.Set(key, value)
//	}
//
//	// Record the CGI handler’s response
//	respRecorder := httptest.NewRecorder()
//	log.Debug("Invoking CGI task")
//
//	// Execute the request
//	handler.ServeHTTP(respRecorder, req)
//	stderrWriter.Close() // Close stderr writer to signal end of stderr capture
//
//	// Capture stderr output
//	var stderrBuf bytes.Buffer
//	_, err = io.Copy(&stderrBuf, stderrPipe)
//	if err != nil {
//		return nil, fmt.Errorf("failed to read CGI output: %w", err)
//	}
//
//	// Check for errors in the response status code
//	if respRecorder.Code > 299 {
//		return nil, fmt.Errorf("received error code %d", respRecorder.Code)
//	}
//
//	log.Infof("Captured CGI output: %s", stderrBuf.String())
//	return respRecorder.Body.Bytes(), nil
//}

type WithToken struct {
	token *oauth2.Token
}

func (token *WithToken) Token() (*oauth2.Token, error) {
	return token.token, nil
}

func (e *Execer) Exec(ctx context.Context, in []byte) ([]byte, error) {
	conf := e.CGIConfig
	log := logger.FromContext(ctx).WithFields(logrus.Fields{
		"cgi.dir":    filepath.Dir(e.Binpath),
		"cgi.path":   e.Binpath,
		"cgi.method": conf.Method,
		"cgi.url":    conf.Endpoint,
	})

	tokenSource := &WithToken{
		token: &oauth2.Token{
			AccessToken: "receivedToken",
			Expiry:      time.UnixMilli(1730758382021),
		},
	}
	client, err := logging.NewClient(ctx, "qa-setup", option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %w", err)
	}
	defer client.Close()

	stackdriverLogger := client.Logger("runner.log").StandardLogger(logging.Debug)

	stderrPipe, stderrWriter, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	defer stderrPipe.Close()

	// Set up the CGI handler
	handler := cgi.Handler{
		Path:   e.Binpath,
		Dir:    filepath.Dir(e.Binpath),
		Env:    append(conf.Envs, os.Environ()...),
		Logger: stackdriverLogger,
		Stderr: stderrWriter,
	}

	// Prepare the HTTP request for the CGI handler
	req, err := http.NewRequestWithContext(ctx, conf.Method, conf.Endpoint, bytes.NewReader(in))
	if err != nil {
		return nil, fmt.Errorf("cannot create CGI request: %w", err)
	}
	for key, value := range conf.Headers {
		req.Header.Set(key, value)
	}

	// Record the CGI handler’s response
	responseRecorder := httptest.NewRecorder()
	log.Debug("Invoking CGI task")

	// Execute the request
	handler.ServeHTTP(responseRecorder, req)
	stderrWriter.Close()

	// Capture stderr output
	var stderrBuf bytes.Buffer
	_, err = io.Copy(&stderrBuf, stderrPipe)
	if err != nil {
		return nil, fmt.Errorf("failed to read CGI output: %w", err)
	}

	log.Infof("Captured CGI output: %s", stderrBuf.String())
	// check the error code and write the error
	// to the context, if applicable.
	// TODO should we unmarshal the response body to an error type?
	if responseRecorder.Code > 299 {
		return nil, fmt.Errorf("received error code %d", responseRecorder.Code)
	}

	return responseRecorder.Body.Bytes(), nil
}
