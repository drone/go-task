package packaged

/**
 * @title PackageLoader

 * @desc PackageLoader is a struct that provides a method to get the path of a pre-package artifact
 * based on the task type and the executable name. This is used when artifacts are packaged with Runner
 * in a container.
 * It is assumed there is only one artifact per task type and executable name, because the OS and architecture
 * are pre-determined.
 */

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/drone/go-task/task"
)

type PackageLoader struct {
	dir string
}

func New(dir string) PackageLoader {
	return PackageLoader{dir: dir}
}

func (p *PackageLoader) GetPackagePath(ctx context.Context, taskType string, exec *task.ExecutableConfig) (string, error) {
	dir := filepath.Join(p.dir, taskType, exec.Name)
	return getFirstFile(dir)
}

func getFirstFile(directory string) (string, error) {
	// Open the directory
	files, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}

	// Iterate through files to find the first file (not a directory)
	for _, file := range files {
		if !file.IsDir() {
			return filepath.Join(directory, file.Name()), nil // Return the first file found
		}
	}

	// If no files are found, return an error
	return "", fmt.Errorf("no cgi found in directory: %s", directory)
}
