// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package fcgi

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/drone/go-task/task/logger"
)

type Execer struct {
	TaskYmlPath string  // path to the task.yml file which contains instructions for execution
	FCGIConfig  *Config // config for the fcgi execution
}

func newExecer(taskYmlPath string, cgiConfig *Config) *Execer {
	return &Execer{
		TaskYmlPath: taskYmlPath,
		FCGIConfig:  cgiConfig,
	}
}

// Exec parses the task.yml file and executes the task given the input
func (e *Execer) Spawn(ctx context.Context, port int) (*exec.Cmd, error) {
	log := logger.FromContext(ctx)
	conf := e.FCGIConfig

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
	// Get the directory of the binary
	bindir := filepath.Dir(binpath)

	// Create the command to run the executable with the -port flag
	cmd := exec.Command(binpath, "-port", fmt.Sprintf("%d", port))

	// Set the environment variables
	cmd.Env = append(os.Environ(), conf.Envs...)

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting the command: %v\n", err)
		return nil, err
	}

	log.With("dir", bindir).
		With("path", binpath).
		With("port", port).
		With("PID", cmd.Process.Pid).
		Info("started FCGI process!")

	return cmd, nil
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
