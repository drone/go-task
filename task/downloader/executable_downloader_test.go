// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package downloader

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/drone/go-task/task"
	"github.com/stretchr/testify/assert"
)

func TestDownloadExecutable(t *testing.T) {
	// Save the original functions to restore them later
	originalIsCacheHitFn := isCacheHitFn
	defer func() { isCacheHitFn = originalIsCacheHitFn }() // Restore after the test

	originalDownloadFileFn := downloadFileFn
	defer func() { downloadFileFn = originalDownloadFileFn }() // Restore after the test

	originalChmodFn := chmodFn
	defer func() { chmodFn = originalChmodFn }() // Restore after the test

	originalRemoveAll := removeAllFn
	defer func() { removeAllFn = originalRemoveAll }() // Restore after the test
	removeAllFn = func(p string) error {
		return nil
	}

	downloader := newExecutableDownloader()

	tests := []struct {
		name        string
		dir         string
		taskType    string
		version     string
		exec        *task.ExecutableConfig
		cacheHit    bool
		downloadErr bool
		chmodErr    bool
		wantErr     bool
	}{
		{
			name:     "successful_download",
			dir:      "/tmp",
			taskType: "binary",
			version:  "v1.0.0",
			exec: &task.ExecutableConfig{
				Executables: []task.Executable{
					{Os: runtime.GOOS, Arch: runtime.GOARCH, Url: "valid_url"},
				},
			},
			cacheHit: false,
			wantErr:  false,
		},
		{
			name:     "cache_hit",
			dir:      "/tmp",
			taskType: "binary",
			version:  "v1.0.0",
			exec: &task.ExecutableConfig{
				Executables: []task.Executable{
					{Os: runtime.GOOS, Arch: runtime.GOARCH, Url: "valid_url"},
				},
			},
			cacheHit: true,
			wantErr:  false,
		},
		{
			name:     "download_error",
			dir:      "/tmp",
			taskType: "binary",
			version:  "v1.0.0",
			exec: &task.ExecutableConfig{
				Executables: []task.Executable{
					{Os: runtime.GOOS, Arch: runtime.GOARCH, Url: "invalid_url"},
				},
			},
			cacheHit:    false,
			downloadErr: true,
			wantErr:     true,
		},
		{
			name:     "chmod_error",
			dir:      "/tmp",
			taskType: "binary",
			version:  "v1.0.0",
			exec: &task.ExecutableConfig{
				Executables: []task.Executable{
					{Os: runtime.GOOS, Arch: runtime.GOARCH, Url: "valid_url"},
				},
			},
			cacheHit: false,
			chmodErr: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCacheHitFn = func(ctx context.Context, dest string) bool {
				return tt.cacheHit
			}

			chmodFn = func(path string, mode os.FileMode) error {
				if tt.chmodErr {
					return fmt.Errorf("chmod error")
				}
				return nil
			}

			downloadFileFn = func(ctx context.Context, url, dest string) (string, error) {
				if tt.downloadErr {
					return "", fmt.Errorf("download error")
				}
				return dest, nil
			}

			_, err := downloader.download(context.Background(), tt.dir, tt.taskType, tt.version, tt.exec)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetExecutableUrl(t *testing.T) {
	downloader := newExecutableDownloader()

	tests := []struct {
		name            string
		config          *task.ExecutableConfig
		operatingSystem string
		architecture    string
		expectedUrl     string
		found           bool
	}{
		{
			name: "valid_executable",
			config: &task.ExecutableConfig{
				Executables: []task.Executable{
					{Os: "linux", Arch: "amd64", Url: "https://example.com/executable"},
				},
			},
			operatingSystem: "linux",
			architecture:    "amd64",
			expectedUrl:     "https://example.com/executable",
			found:           true,
		},
		{
			name: "invalid_executable",
			config: &task.ExecutableConfig{
				Executables: []task.Executable{
					{Os: "linux", Arch: "amd64", Url: "https://example.com/executable"},
				},
			},
			operatingSystem: "windows",
			architecture:    "amd64",
			expectedUrl:     "",
			found:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, found := downloader.getExecutableUrl(tt.config, tt.operatingSystem, tt.architecture)
			assert.Equal(t, tt.expectedUrl, url)
			assert.Equal(t, tt.found, found)
		})
	}
}
