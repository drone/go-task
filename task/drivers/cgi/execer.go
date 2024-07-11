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

// TODO: We can add more fields to the execer (like stdout, stderr, env, etc)
// but they are not being used at the moment.
type Execer struct {
	Source string // path to the task.yml file which contains instructions for execution
}

// Exec parses the task.yml file and executes the task given a CGI configuration.
func (e *Execer) Exec(ctx context.Context, conf *Config, input []byte) ([]byte, error) {
	log := logger.FromContext(ctx)

	out, err := ParseFile(e.Source)
	if err != nil {
		return nil, err
	}

	goos := runtime.GOOS

	// install linux dependencies
	if goos == "linux" {
		e.installAptDeps(ctx, out.Spec.Deps.Apt)
	}

	// install darwin dependencies
	if goos == "darwin" {
		e.installBrewDeps(ctx, out.Spec.Deps.Brew)
	}

	var binpath string

	// build go module if specified
	if out.Spec.Run.Go != nil {
		module := out.Spec.Run.Go.Module
		if module != "" {
			binpath, err = e.buildGoModule(ctx, module)
			if err != nil {
				return nil, fmt.Errorf("failed to build go module: %w", err)
			}
		}
	} else if out.Spec.Run.Bash != nil {
		binpath = filepath.Join(filepath.Dir(e.Source), out.Spec.Run.Bash.Script)
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
		bytes.NewReader(input),
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

// install any linux dependencies if specified
func (e *Execer) installAptDeps(ctx context.Context, deps []AptDep) {
	log := logger.FromContext(ctx)
	if len(deps) > 0 {
		log.Debug("apt-get update")

		cmd := exec.Command("sudo", "apt-get", "update")
		cmd.Dir = filepath.Dir(e.Source)
		cmd.Run()
	}

	for _, dep := range deps {
		log.Debug("apt-get install", slog.String("package", dep.Name))

		cmd := exec.Command("sudo", "apt-get", "install", dep.Name)
		cmd.Dir = filepath.Dir(e.Source)
		cmd.Run()
	}
}

// build go module and return back path to the binary
func (e *Execer) buildGoModule(ctx context.Context, module string) (string, error) {
	log := logger.FromContext(ctx)
	log.Debug("go build", slog.String("module", module))

	// create binary in the same directory as the task.yml file
	target := filepath.Join(filepath.Dir(e.Source), "task")

	// build the code
	cmd := exec.Command("go", "build", "-o", target, module)
	cmd.Dir = filepath.Dir(e.Source)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return target, nil
}

// install any brew dependencies if specified
func (e *Execer) installBrewDeps(ctx context.Context, deps []BrewDep) {
	log := logger.FromContext(ctx)
	for _, item := range deps {
		log.Debug("brew install", slog.String("package", item.Name))

		cmd := exec.Command("brew", "install", item.Name)
		cmd.Dir = filepath.Dir(e.Source)
		cmd.Run()
	}
}
