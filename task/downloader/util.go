package downloader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/drone/go-task/task/logger"
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

// downloadFile fetches the file from a list of urls and writes it to dest.
// It tries urls one by one until a download is successful.
func downloadFile(ctx context.Context, urls []string, dest string) (string, error) {
	log := logger.FromContext(ctx)

	downloadDir := filepath.Dir(dest)
	// create the directory where the target is downloaded.
	if err := mkdirAllFn(downloadDir, 0777); err != nil {
		return "", err
	}

	var lastErr error
	for _, u := range urls {
		log.WithFields(map[string]interface{}{
			"source":      u,
			"destination": dest,
		}).Debug("attempting to download artifact")

		resp, err := httpGetFn(u)
		if err != nil {
			lastErr = fmt.Errorf("failed to download file from %s: %w", u, err)
			log.WithError(lastErr).Warn("download attempt failed")
			continue // try next url
		}

		if code := resp.StatusCode; code > 299 {
			resp.Body.Close()
			lastErr = fmt.Errorf("download error with status code %d for url %s", code, u)
			log.WithError(lastErr).Warn("download attempt failed")
			continue // try next url
		}

		outFile, err := createFn(dest)
		if err != nil {
			resp.Body.Close()
			return "", fmt.Errorf("failed to create file: %w", err)
		}

		_, err = copyFn(outFile, resp.Body)
		outFile.Close()
		resp.Body.Close()

		if err != nil {
			// This is a file writing error, not a download error. Fail immediately.
			return "", fmt.Errorf("failed to write to file: %w", err)
		}

		log.Debug("downloaded artifact successfully")
		return dest, nil // success
	}

	return "", fmt.Errorf("failed to download file from all provided urls: %w", lastErr)
}

// getDownloadPath returns the full download path given the download url and the destination folder `dest`
func getDownloadPath(url, dest string) string {
	fileName := filepath.Base(url)
	return filepath.Join(dest, fileName)
}

// isCacheHit checks if the `dest` folder already exists
func isCacheHit(ctx context.Context, dest string) bool {
	log := logger.FromContext(ctx).
		WithFields(map[string]interface{}{
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

func getHash(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}
