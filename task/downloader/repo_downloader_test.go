// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package downloader

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/drone/go-task/task"
	"github.com/drone/go-task/task/cloner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Cloner to use in tests
type MockCloner struct {
	mock.Mock
}

func (m *MockCloner) Clone(ctx context.Context, params cloner.Params) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

// Test for download method
func TestDownload(t *testing.T) {
	// Save the original httpGet to restore it later
	originalIsCacheHitFn := isCacheHitFn
	defer func() { isCacheHitFn = originalIsCacheHitFn }() // Restore after the test

	mockCloner := new(MockCloner)
	downloader := newRepoDownloader(mockCloner)

	tests := []struct {
		name     string
		repo     *task.Repository
		wantErr  bool
		cacheHit bool
		cloneErr bool
	}{
		{
			name:     "successful_clone",
			repo:     &task.Repository{Clone: "https://github.com/user/repo.git", Ref: "main"},
			wantErr:  false,
			cacheHit: false,
		},
		{
			name:     "no_repository",
			repo:     nil,
			wantErr:  true,
			cacheHit: false,
		},
		{
			name:     "cache_hit",
			repo:     &task.Repository{Clone: "https://github.com/user/repo.git", Ref: "main"},
			wantErr:  false,
			cacheHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the cache hit function
			isCacheHitFn = func(ctx context.Context, dest string) bool {
				return tt.cacheHit
			}
			if !tt.cacheHit {
				mockCloner.On("Clone", mock.Anything, mock.Anything).Return(nil)
			}
			_, err := downloader.download(context.Background(), "/tmp", tt.repo)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClone(t *testing.T) {
	mockCloner := new(MockCloner)
	downloader := newRepoDownloader(mockCloner)

	repo := &task.Repository{
		Clone: "https://github.com/user/repo.git",
		Ref:   "main",
		Sha:   "abc123",
	}

	mockCloner.On("Clone", mock.Anything, mock.Anything).Return(nil).Once()

	err := downloader.clone(context.Background(), repo, "/tmp/clone-dir")
	assert.NoError(t, err)

	mockCloner.AssertExpectations(t)
}

func TestGetDownloadDir(t *testing.T) {
	downloader := newRepoDownloader(nil)

	repo := &task.Repository{
		Clone: "https://github.com/user/repo.git",
		Ref:   "main",
		Sha:   "abc123",
	}

	hash := downloader.getHashOfRepo(repo)
	dir := downloader.getDownloadDir("/tmp", repo)
	expectedDir := filepath.Join("/tmp", hash)

	assert.Equal(t, expectedDir, dir)
}
