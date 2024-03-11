package download

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
