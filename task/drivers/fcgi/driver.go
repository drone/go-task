// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fcgi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/download"
	"github.com/drone/go-task/task/logger"
)

var (
	taskYmlPath = "task.yml"
)

// Config provides the driver config.
type Config struct {
	Repository *task.Repository  `json:"repository"`
	Headers    map[string]string `json:"headers"`
	Envs       []string          `json:"envs"`
	Operation  string            `json:"operation"`
}

// New returns the task execution driver.
func New(d download.Downloader) task.Handler {
	return &Driver{downloader: d, fcgiServers: make(map[string]*exec.Cmd), fcgiPorts: make(map[string]int), nextPort: 9000}
}

type Driver struct {
	downloader    download.Downloader
	fcgiServers   map[string]*exec.Cmd
	fcgiPorts     map[string]int
	fcgiServersMu sync.Mutex
	nextPort      int
}

// Handle handles the task execution request.
func (d *Driver) Handle(ctx context.Context, req *task.Request) task.Response {
	var (
		log  = logger.FromContext(ctx)
		conf = new(Config)
	)

	// decode the task configuration
	err := json.Unmarshal(req.Task.Config, conf)
	if err != nil {
		return task.Error(err)
	}
	switch operation := conf.Operation; operation {
	case "ASSIGN":
		return d.handleAssignToFCGIServer(ctx, log, req.Task.Type, conf, req.Task.Data)
		//case "UNASSIGN":
		//return d.handleUnassignToFCGIServer(log, req.Task.Type, conf, req.Task.Data)
	default:
		return task.Errorf("unsupported operation %s", operation)

	}
}

// handlePrintMessage handles HTTP requests from the FastCGI server to print a message
func (d *Driver) handleAssignToFCGIServer(ctx context.Context, log *slog.Logger, taskType string, conf *Config, data []byte) task.Response {
	d.fcgiServersMu.Lock()
	defer d.fcgiServersMu.Unlock()
	_, ok := d.fcgiServers[taskType]
	if ok {
		// Just assign the task to to the fcgi server.
		port := d.fcgiPorts[taskType]
		log.Info(fmt.Sprintf("FCGI server exists already for taskType %s in port %d", taskType, port))
		return assignTask(port, data)
	}
	// otherwise, we need to spawn the fcgi server.
	log.Info(fmt.Sprintf("FCGI server does not exist for taskType %s. We are going to create it!", taskType))
	port := d.nextPort
	d.nextPort++
	cmd, err := d.spawnFCGIServer(ctx, port, conf)
	if err != nil {
		log.Error("Something went wrong!")
		return task.Error(err)
	}
	log.Info(fmt.Sprintf("Successfully spawned new FCGI server for taskType %s at port %d", taskType, port))
	d.fcgiServers[taskType] = cmd
	d.fcgiPorts[taskType] = port
	// Wait until server starts
	time.Sleep(3 * time.Second)
	log.Info(fmt.Sprintf("Assigning task to new FCGI server for taskType %s at port %d", taskType, port))
	return assignTask(port, data)
}

//func (d *driver) handleUnassignToFCGIServer(log *sdog.Logger, taskType string, conf *Config, data json.RawMessage) task.Response {
//// TODO: Implement this thing
//return task.Respond("Whatever")
//}

func assignTask(port int, data []byte) task.Response {
	url := fmt.Sprintf("http://localhost:%d/assign", port)
	resp, err := http.Post(url, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return task.Errorf("Something went wrong here")
	}
	defer resp.Body.Close()

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return task.Errorf("Something went wrong over here")
	}
	return task.Respond(string(body))
}

func (d *Driver) spawnFCGIServer(ctx context.Context, port int, conf *Config) (*exec.Cmd, error) {
	log := logger.FromContext(ctx)
	path, err := d.downloader.Download(ctx, conf.Repository)
	if err != nil {
		log.With("error", err).Error("artifact download failed")
		return nil, err
	}
	fullpath := filepath.Join(path, taskYmlPath)
	execer := newExecer(fullpath, conf)
	cmd, err := execer.Spawn(ctx, port)
	if err != nil {
		log.With("error", err).Error("could not spawn fcgi server")
		return nil, err
	}
	log.Info(fmt.Sprintf("Spawned the FCGI server on port %d and path %s", port, fullpath))
	return cmd, nil
}

func KillFCGIServer(id string) error {
	return nil
	//fcgiServersMu.Lock()
	//cmd, exists := fcgiServers[id]
	//port := fcgiPorts[id]
	//if !exists {
	//fcgiServersMu.Unlock()
	//return fmt.Errorf("FastCGI server not found")
	//}
	//delete(fcgiServers, id)
	//delete(fcgiPorts, id)
	//fcgiServersMu.Unlock()

	//err := cmd.Process.Kill()
	//if err != nil {
	//return err
	//}

	//fmt.Printf("Killed FastCGI server on port %d\n", port)
	//return nil
}
