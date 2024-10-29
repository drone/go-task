package downloader

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	globallogger "github.com/harness/runner/logger/logger"
)

// functions for mocking
var (
	mkdirAllFn     = os.MkdirAll
	httpGetFn      = http.Get
	createFn       = os.Create
	copyFn         = io.Copy
	getcacheFn     = os.UserCacheDir
	isCacheHitFn   = isCacheHit
	downloadFileFn = downloadFile
)

// downloadFile fetches the file from url and writes it to dest
func downloadFile(ctx context.Context, url, dest string) (string, error) {

	log := globallogger.FromContext(ctx).
		WithFields(logrus.Fields{
			"source":      url,
			"destination": dest,
		})
	log.Debug("downloading artifact")

	downloadDir := filepath.Dir(dest)
	// create the directory where the target is downloaded.
	if err := mkdirAllFn(downloadDir, 0777); err != nil {
		return "", err
	}

	resp, err := httpGetFn(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code > 299 {
		return "", fmt.Errorf("download error with status code %d", code)
	}

	outFile, err := createFn(dest)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	_, err = copyFn(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to file: %w", err)
	}

	log.Debug("downloaded artifact")

	return dest, nil
}

// getDownloadPath returns the full download path given the download url and the destination folder `dest`
func getDownloadPath(url, dest string) string {
	fileName := filepath.Base(url)
	return filepath.Join(dest, fileName)
}

// isCacheHit checks if the `dest` folder already exists
func isCacheHit(ctx context.Context, dest string) bool {
	log := globallogger.FromContext(ctx).
		WithFields(logrus.Fields{
			"target": dest,
		})

	if _, err := os.Stat(dest); err == nil {
		log.Debug("cache hit")
		return true
	}

	log.Debug("cache miss")
	return false
}

// ExpandCache returns the root directory where task
// downloads and repositories should be cached.
func ExpandCache(s string) string {
	cache, _ := getcacheFn()
	return strings.ReplaceAll(s, "$XDG_CACHE_HOME", cache)
}

// ExpandCacheSlice returns the root directory where task
// downloads and repositories should be cached.
func ExpandCacheSlice(items []string) []string {
	for i, s := range items {
		items[i] = ExpandCache(s)
	}
	return items
}

// IsRepository returns true if the provided download url
// is a git repository.
func IsRepository(s string) bool {
	u, _ := url.Parse(s)
	return strings.HasSuffix(u.Path, ".git")
}

// SplitRef splits the repository url and the commit ref.
func SplitRef(s string) (string, string) {
	u, err := url.Parse(s)
	if err != nil || u.Fragment == "" {
		return s, ""
	} else {
		ref := u.Fragment
		u.Fragment = ""
		u.RawFragment = ""
		return u.String(), ref
	}
}
