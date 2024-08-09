// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fcgi

import (
	"os"
	"strings"
)

// getcache as a function for mocking
var getcache = os.UserCacheDir

// expandCache expands the XDG_CACHE_HOME cache
// environment variable in the string.
func expandCache(s string) string {
	cache, _ := getcache()
	return strings.ReplaceAll(s, "$XDG_CACHE_HOME", cache)
}

// expandCacheSlice expands the XDG_CACHE_HOME cache
// environment variable in the slice.
func expandCacheSlice(items []string) []string {
	for i, s := range items {
		items[i] = expandCache(s)
	}
	return items
}
