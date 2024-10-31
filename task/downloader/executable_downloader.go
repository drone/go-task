// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/drone/go-task/task/logger"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/drone/go-task/task"
)

// removeAllFn as a function for mocking
var removeAllFn = os.RemoveAll

// chmodFn as a function for mocking
var chmodFn = os.Chmod

// executableDownloader a binary executable file
// It also takes care of where to download the file
type executableDownloader struct{}

func newExecutableDownloader() *executableDownloader {
	return &executableDownloader{}
}

func (e *executableDownloader) download(ctx context.Context, dir string, taskType string, version string, exec *task.ExecutableConfig) (string, error) {
	if exec == nil {
		return "", errors.New("no executable urls provided to download")
	}
	operatingSystem := runtime.GOOS
	architecture := runtime.GOARCH
	url, ok := e.getExecutableUrl(exec, operatingSystem, architecture)
	if !ok {
		return "", fmt.Errorf("os [%s] and architecture [%s] are not specified in executable configuration", operatingSystem, architecture)
	}

	destDir := filepath.Join(dir, taskType, version)
	dest := getDownloadPath(url, destDir)

	if cacheHit := isCacheHitFn(ctx, destDir); cacheHit {
		// exit if the artifact destination already exists
		return dest, nil
	}

	// if no cache hit, remove all downloaded executables for this task's type
	// so that we don't keep multiple executables of the same type
	err := removeAllFn(filepath.Join(dir, taskType))
	if err != nil {
		return "", err
	}

	binpath, err := downloadFileFn(ctx, url, dest)
	if err != nil {
		// remove the destination directory if downloading fails so it can be retried
		removeAllFn(destDir)
		return "", err
	}
	e.logExecutableDownload(ctx, exec, operatingSystem, architecture)

	err = chmodFn(binpath, 0777)
	if err != nil {
		return "", fmt.Errorf("failed to set executable flag in task file [%s]: %w", binpath, err)
	}
	return binpath, nil
}

// getExecutableUrl fetches the download url for a task's executable file,
// given the current system's operating system and architecture
func (e *executableDownloader) getExecutableUrl(config *task.ExecutableConfig, operatingSystem, architecture string) (string, bool) {
	for _, exec := range config.Executables {
		if exec.Os == operatingSystem && exec.Arch == architecture {
			return exec.Url, true
		}
	}
	return "", false
}

// logExecutableDownload writes details about the Executable struct used to download a task's executable file
func (e *executableDownloader) logExecutableDownload(ctx context.Context, exec *task.ExecutableConfig, operatingSystem, architecture string) {
	log := logger.FromContext(ctx)
	filename := "executable_downloads.log"
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to open log file [%s]: %v", filename, err))
	}
	defer file.Close()

	// Convert the struct to JSON
	data, err := json.Marshal(exec)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to marshall Executable struct to json: %v", err))
	}

	entry := fmt.Sprintf("%s: dowloaded for os: [%s], arch: [%s] %s\n", time.Now().Format(time.RFC3339), operatingSystem, architecture, string(data))
	// Write the JSON string to the file, followed by a newline
	if _, err := file.WriteString(entry); err != nil {
		log.Error(fmt.Sprintf("Failed to write Executable struct to log file [%s]: %v", filename, err))
	}
}
