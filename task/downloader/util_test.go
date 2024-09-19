package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	// Save the original functions
	originalmkdirAll := mkdirAllFn
	originalHttpGet := httpGetFn
	originalCreateFn := createFn
	originalCopyFn := copyFn

	// Mock functions
	mkdirAllFn = func(s string, m os.FileMode) error {
		return nil
	}
	copyFn = func(w io.Writer, r io.Reader) (int64, error) {
		return 0, nil
	}

	// Restore original functions when test finishes
	defer func() { mkdirAllFn = originalmkdirAll }()
	defer func() { httpGetFn = originalHttpGet }()
	defer func() { createFn = originalCreateFn }()
	defer func() { copyFn = originalCopyFn }()

	tests := []struct {
		name          string
		url           string
		dest          string
		fileCreateErr bool
		wantErr       bool
		mockGetFn     func(string) (*http.Response, error)
	}{
		{
			name:    "successful_download",
			url:     "http://example.com/file.txt",
			dest:    "/tmp/testfile.txt",
			wantErr: false,
			mockGetFn: func(url string) (*http.Response, error) {
				body := io.NopCloser(strings.NewReader("mock file content"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       body,
				}, nil
			},
		},
		{
			name:    "http_error",
			url:     "http://example.com/nonexistent",
			dest:    "/tmp/testfile.txt",
			wantErr: true,
			mockGetFn: func(url string) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			},
		},
		{
			name:          "file_creation_error",
			url:           "http://example.com/file.txt",
			dest:          "/invalid/dir/testfile.txt",
			fileCreateErr: true,
			wantErr:       true,
			mockGetFn: func(url string) (*http.Response, error) {
				body := io.NopCloser(strings.NewReader("mock file content"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       body,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the functions
			httpGetFn = tt.mockGetFn
			if tt.fileCreateErr {
				createFn = func(s string) (*os.File, error) {
					return nil, fmt.Errorf("error creating file")
				}
			} else {
				createFn = func(s string) (*os.File, error) {
					return &os.File{}, nil
				}
			}

			_, err := downloadFile(context.Background(), tt.url, tt.dest)
			if gotErr := err != nil; gotErr != tt.wantErr {
				t.Errorf("downloadFile() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestGetDownloadPath(t *testing.T) {
	tests := []struct {
		url      string
		dest     string
		expected string
	}{
		{
			url:      "http://example.com/file.txt",
			dest:     "/downloads",
			expected: "/downloads/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("url: %s", tt.url), func(t *testing.T) {
			if got := getDownloadPath(tt.url, tt.dest); got != tt.expected {
				t.Errorf("getDownloadPath(%q, %q) = %q, want %q", tt.url, tt.dest, got, tt.expected)
			}
		})
	}
}

func TestIsCacheHit(t *testing.T) {
	tests := []struct {
		name    string
		dest    string
		setup   func(string) // function to create a file for "cache hit"
		wantHit bool
	}{
		{
			name: "cache_hit",
			dest: "/tmp/testfile.txt",
			setup: func(dest string) {
				_, _ = os.Create(dest)
			},
			wantHit: true,
		},
		{
			name:    "cache_miss",
			dest:    "/tmp/nonexistent.txt",
			setup:   func(dest string) {}, // no setup for cache miss
			wantHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.dest)

			if got := isCacheHit(context.Background(), tt.dest); got != tt.wantHit {
				t.Errorf("isCacheHit() = %v, want %v", got, tt.wantHit)
			}
		})
	}
}

func TestExpandCache(t *testing.T) {
	// provide a mock function to get the os cache
	getcacheFn = func() (string, error) {
		return "/home/ubuntu/.cache", nil
	}
	// reset to the original when the test completes
	defer func() {
		getcacheFn = os.UserCacheDir
	}()
	tests := []struct {
		before string
		after  string
	}{
		{
			before: "$XDG_CACHE_HOME/harness/task/slack-v1.0.0",
			after:  "/home/ubuntu/.cache/harness/task/slack-v1.0.0",
		},
		{
			before: "/var/harness/cache/harness/task/slack-v1.0.0",
			after:  "/var/harness/cache/harness/task/slack-v1.0.0",
		},
	}
	for _, test := range tests {
		if got, want := ExpandCache(test.before), test.after; got != want {
			t.Errorf("Want cache dir %s, got %s", want, got)
		}
	}
}

func TestSplitRef(t *testing.T) {
	tests := []struct {
		in  string
		url string
		ref string
	}{
		{
			in:  "https://github.com/octocat/hello-world.git#main",
			url: "https://github.com/octocat/hello-world.git",
			ref: "main",
		},
		{
			in:  "https://github.com/octocat/hello-world.git",
			url: "https://github.com/octocat/hello-world.git",
			ref: "",
		},
	}
	for _, test := range tests {
		url, ref := SplitRef(test.in)
		if got, want := url, test.url; got != want {
			t.Errorf("Expect url %s, got %s", got, want)
		}
		if got, want := ref, test.ref; got != want {
			t.Errorf("Expect ref %s, got %s", got, want)
		}
	}
}

func TestIsRepository(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{
			url:  "https://github.com/octocat/hello-world.git",
			want: true,
		},
		{
			url:  "https://github.com/octocat/hello-world/downloads/latest/release.tar.gz",
			want: false,
		},
	}
	for _, test := range tests {
		if got, want := IsRepository(test.url), test.want; got != want {
			t.Errorf("Expect %q is repository %v, got %v", test.url, got, want)
		}
	}
}
