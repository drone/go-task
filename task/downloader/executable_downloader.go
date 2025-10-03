// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"

	"github.com/drone/go-task/task/logger"

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

func (e *executableDownloader) download(ctx context.Context, dir string, taskType string, exec *task.ExecutableConfig, fallbackEnabled bool, envs map[string]string) (string, error) {
	if exec == nil {
		return "", errors.New("no executable urls provided to download")
	}
	operatingSystem := runtime.GOOS
	architecture := runtime.GOARCH
	urls, ok := e.getExecutableUrl(exec, operatingSystem, architecture, fallbackEnabled)
	if !ok {
		return "", fmt.Errorf("os [%s] and architecture [%s] are not specified in executable configuration", operatingSystem, architecture)
	}

	var destDir, dest string

	if exec.Target == "" {
		destDir = filepath.Join(dir, taskType, exec.Name)
		// {baseDir}/taskType/{name}/{name}-{version}-{os}-{arch}
		// download to a file named by this runner to make sure upstream changes doesn't affect the cache hit lookup
		dest = filepath.Join(destDir, exec.Name+"-"+exec.Version+"-"+operatingSystem+"-"+architecture)
	} else {
		dest = os.Expand(exec.Target, func(key string) string {
			if val, ok := envs[key]; ok {
				return val
			}
			return os.Getenv(key)
		})
		dest = os.ExpandEnv(dest)
	}
	if cacheHit := isCacheHitFn(ctx, dest); cacheHit {
		// exit if the artifact destination already exists
		return dest, nil
	}

	if exec.Compressed {
		dest = dest + ".zst"
	}

	// if no cache hit and destDir is set, remove all downloaded executables for this task's type
	// so that we don't keep multiple executables of the same type
	if destDir != "" {
		if err := removeAllFn(destDir); err != nil {
			return "", err
		}
	}

	binPath, err := downloadFileFn(ctx, urls, dest)
	if err != nil {
		// remove the destination directory if downloading fails so it can be retried
		if destDir != "" {
			removeAllFn(destDir)
		}
		return "", err
	}
	e.logExecutableDownload(ctx, exec, operatingSystem, architecture)

	if exec.Compressed {
		binPath, err = decompressFile(ctx, binPath)
		if err != nil {
			return "", fmt.Errorf("failed to decompress plugin [%s]: %w", binPath, err)
		}
	}

	if err = chmodFn(binPath, 0777); err != nil {
		return "", fmt.Errorf("failed to set executable flag in task file [%s]: %w", binPath, err)
	}
	return binPath, nil
}

// decompressFile decompresses a zstd file if needed
func decompressFile(ctx context.Context, filePath string) (string, error) {
	if !strings.HasSuffix(filePath, ".zst") {
		return filePath, nil // Not a zstd file, return original path
	}

	// Open the compressed file
	compressedFile, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open compressed file: %v", err)
	}
	defer compressedFile.Close()

	// Create the decompressed file (remove .zst extension)
	decompressedPath := strings.TrimSuffix(filePath, ".zst")
	decompressedFile, err := os.Create(decompressedPath)
	if err != nil {
		return "", fmt.Errorf("failed to create decompressed file: %v", err)
	}
	defer decompressedFile.Close()

	// Decompress
	if err := decompress(compressedFile, decompressedFile); err != nil {
		return "", fmt.Errorf("failed to decompress file: %v", err)
	}

	log := logger.FromContext(ctx)
	// Remove the original compressed file after successful decompression
	if err := os.Remove(filePath); err != nil {
		log.Error(fmt.Sprintf("Failed to remove compressed file [%s] after decompression: %v", filePath, err))
	}
	return decompressedPath, nil
}

// decompress decompresses a zstd compressed file
func decompress(in io.Reader, out io.Writer) error {
	d, err := zstd.NewReader(in)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(out, d)
	return err
}

// getExecutableUrl fetches download urls for a task's executable file,
// given the current system's operating system and architecture.
// If fallbackEnabled is true, it returns all matching URLs. Otherwise, it returns the first match.
func (e *executableDownloader) getExecutableUrl(config *task.ExecutableConfig, operatingSystem, architecture string, fallbackEnabled bool) ([]string, bool) {
	var urls []string
	for _, exec := range config.Executables {
		if exec.Os == operatingSystem && exec.Arch == architecture {
			urls = append(urls, exec.Url)
			if !fallbackEnabled {
				return urls, true // Return immediately with the first match
			}
		}
	}
	return urls, len(urls) > 0
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
