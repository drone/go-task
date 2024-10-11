// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/common"
	"github.com/drone/go-task/task/masker"
)

type (
	// input for exec task handler
	execInput struct {
		Shell  string   `json:"shell"`
		Script []string `json:"script"`
		Envs   []string `json:"envs"`
	}

	// output for the exec task handler
	execOutput struct {
		Pid      int           `json:"pid"`
		Exited   bool          `json:"exited"`
		ExitCode int           `json:"exit_code"`
		UserTime time.Duration `json:"user_time"`
		SysTime  time.Duration `json:"sys_time"`
		Output   []string      `json:"out"`
	}

	// input for the file task handler
	fileInput struct {
		Path string `json:"path"`
	}
)

// Sample handler that exposes os/exec as a task. This is a
// sample only and is not meant for production use.
//
// Sample json input:
//
//	{
//	    "task": {
//	        "id": "67c0938c-9348-4c5e-8624-28218984e09f",
//	        "type": "sample/exec",
//	        "data": {
//	            "script": [ "echo hello world" ],
//	            "shell": "sh"
//	        }
//	    }
//	}
func execHandler(ctx context.Context, req *task.Request) task.Response {
	var conf = new(execInput)

	// decode the task configuration
	if err := json.Unmarshal(req.Task.Data, &conf); err != nil {
		return task.Error(err)
	}

	// create a buffer for stdout / stderr, wrapped
	// in the secret masker
	buf := new(bytes.Buffer)

	buf_ := masker.New(
		buf,
		masker.Slice(req.Secrets),
	)

	// create the command
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", strings.Join(conf.Script, "\n"))
	cmd.Stdout = buf_
	cmd.Stderr = buf_
	cmd.Env = conf.Envs

	// execute the command
	if err := cmd.Run(); err != nil {
		return task.Error(err)
	}

	// collect the output
	out := &execOutput{
		Pid:      cmd.ProcessState.Pid(),
		Exited:   cmd.ProcessState.Exited(),
		ExitCode: cmd.ProcessState.ExitCode(),
		SysTime:  cmd.ProcessState.SystemTime(),
		UserTime: cmd.ProcessState.UserTime(),

		// convert the buffer to log lines
		Output: strings.Split(buf.String(), "\n"),
	}

	return task.Respond(out)
}

// Sample handler that reads a file as a task. This is a
// sample only and is not meant for production use.
//
// Sample json input:
//
//	{
//	    "task": {
//	        "id": "67c0938c-9348-4c5e-8624-28218984e09f",
//	        "type": "sample/file",
//	        "data": {
//	            "path": "path/to/file.txt"
//	        }
//	    }
//	}
func fileHandler(ctx context.Context, req *task.Request) task.Response {
	conf := new(fileInput)

	// decode the task configuration.
	err := json.Unmarshal(req.Task.Data, conf)
	if err != nil {
		return task.Error(err)
	}

	// read the secret from the file.
	contents, err := os.ReadFile(conf.Path)
	if err != nil {
		return task.Error(err)
	}

	// write the secret to the response.
	return task.Respond(
		&common.Secret{
			Value: string(contents),
		},
	)
}
