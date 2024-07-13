// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package cgi

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/drone/go-task/task/logger"
)

type Execer struct {
	TaskYmlPath string  // path to the task.yml file which contains instructions for execution
	CGIConfig   *Config // config for the cgi execution
}

func newExecer(taskYmlPath string, cgiConfig *Config) *Execer {
	return &Execer{
		TaskYmlPath: taskYmlPath,
		CGIConfig:   cgiConfig,
	}
}

// Exec parses the task.yml file and executes the task given the input
func (e *Execer) Exec(ctx context.Context, in []byte) ([]byte, error) {
	log := logger.FromContext(ctx)
	conf := e.CGIConfig

	out, err := ParseFile(e.TaskYmlPath)
	if err != nil {
		return nil, err
	}

	// install any dependencies for the task
	log.Info("installing dependencies")
	if err := e.installDeps(ctx, out.Spec.Deps); err != nil {
		return nil, fmt.Errorf("failed to install dependencies: %w", err)
	}
	log.Info("finished installing dependencies")

	var binpath string

	// build go binary if specified
	if out.Spec.Run.Go != nil {
		module := out.Spec.Run.Go.Module
		if module != "" {
			binName := "task.exe"
			err = e.buildGoModule(ctx, module, binName)
			if err != nil {
				return nil, fmt.Errorf("failed to build go module: %w", err)
			}
			binpath = filepath.Join(filepath.Dir(e.TaskYmlPath), binName)
		}
	} else if out.Spec.Run.Bash != nil {
		binpath = filepath.Join(filepath.Dir(e.TaskYmlPath), out.Spec.Run.Bash.Script)
	} else {
		return nil, fmt.Errorf("no execution specified in task.yml file")
	}

	// run the task using cgi
	handler := cgi.Handler{
		Path: binpath,
		Dir:  filepath.Dir(binpath),
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

	log.With("cgi.dir", filepath.Dir(binpath)).
		With("cgi.path", binpath).
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

// installDeps installs any dependencies for the task
func (e *Execer) installDeps(ctx context.Context, deps Deps) error {
	goos := runtime.GOOS

	// install linux dependencies
	if goos == "linux" {
		return e.installAptDeps(ctx, deps.Apt)
	}

	// install darwin dependencies
	if goos == "darwin" {
		return e.installBrewDeps(ctx, deps.Brew)
	}

	return nil
}

func (e *Execer) installAptDeps(ctx context.Context, deps []AptDep) error {
	log := logger.FromContext(ctx)
	var err error
	if len(deps) > 0 {
		log.Info("apt-get update")

		cmd := e.cmdRunner("sudo", "apt-get", "update")
		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	for _, dep := range deps {
		log.Info("apt-get install", slog.String("package", dep.Name))

		cmd := e.cmdRunner("sudo", "apt-get", "install", dep.Name)
		if err = cmd.Run(); err != nil {
			// TODO: perhaps errors can be logged as warnings instead of returning here,
			// but we can evaluate this in the future.
			return err
		}
	}

	return nil
}

// cmdRunner returns a new exec.Cmd with the given name and arguments
// It populates the working directory as the directory of the task.yml file.
func (e *Execer) cmdRunner(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Dir = filepath.Dir(e.TaskYmlPath)
	return cmd
}

func (e *Execer) buildGoModule(
	ctx context.Context,
	module string,
	binName string, // name of the target binary
) error {
	log := logger.FromContext(ctx)
	log.Info("go build", slog.String("module", module))

	// build the code
	cmd := e.cmdRunner("go", "build", "-o", binName, module)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (e *Execer) installBrewDeps(ctx context.Context, deps []BrewDep) error {
	log := logger.FromContext(ctx)
	for _, item := range deps {
		log.Info("brew install", slog.String("package", item.Name))

		cmd := e.cmdRunner("brew", "install", item.Name)
		if err := cmd.Run(); err != nil {
			// TODO: perhaps errors can be logged as a warning instead of returning here,
			// but we can evaluate this in the future.
			return err
		}
	}
	return nil
}
