package download

import (
	"net/url"
	"os"
	"strings"
)

// getcache as a function for mocking
var getcache = os.UserCacheDir

// ExpandCache returns the root directory where task
// downloads and repositories should be cached.
func ExpandCache(s string) string {
	cache, _ := getcache()
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
