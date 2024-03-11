// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgi

import (
	"os"
	"testing"
)

func TestExpandCache(t *testing.T) {
	// provide a mock function to get the os cache
	getcache = func() (string, error) {
		return "/home/ubuntu/.cache", nil
	}
	// reset to the original when the test completes
	defer func() {
		getcache = os.UserCacheDir
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
		if got, want := expandCache(test.before), test.after; got != want {
			t.Errorf("Want cache dir %s, got %s", want, got)
		}
	}
}
