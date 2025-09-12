// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package binary

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/drone/go-task/task/logger"
)

type Execer struct {
	binPath string
	config  *Config
}

func newExecer(binPath string, config *Config) *Execer {
	return &Execer{
		binPath: binPath,
		config:  config,
	}
}

// Exec executes the binary with the given configuration
func (e *Execer) Exec(ctx context.Context, taskData []byte) error {
	log := logger.FromContext(ctx).WithFields(map[string]interface{}{
		"binary.dir":  filepath.Dir(e.binPath),
		"binary.path": e.binPath,
	})

	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, e.binPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(e.config.Envs, os.Environ()...)

	if e.config.WorkDir != "" {
		cmd.Dir = e.config.WorkDir
	} else {
		cmd.Dir = filepath.Dir(e.binPath)
	}

	log.Debug("executing binary task")

	err := cmd.Run()

	// Log output for debugging
	if stdout.Len() > 0 {
		log.Infof("stdout: %s", stdout.String())
	}
	if stderr.Len() > 0 {
		log.Infof("stderr: %s", stderr.String())
	}

	return err
}
