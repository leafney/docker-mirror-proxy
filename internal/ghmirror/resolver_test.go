package ghmirror

import (
	"reflect"
	"testing"

	"github.com/leafney/docker-mirror-proxy/internal/ghurl"
)

func TestCandidates(t *testing.T) {
	ref, err := ghurl.Parse("https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	got := Candidates(ref)
	want := []string{
		"https://ghfast.top/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
		"https://gh-proxy.com/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
		"https://gh-proxy.org/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
		"https://fastgit.cc/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
		"https://ghproxy.net/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
		"https://ghproxylist.com/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
		"https://hk.gh-proxy.org/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
		"https://cdn.gh-proxy.org/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
		"https://ghp.keleyaa.com/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
		"https://gh.jasonzeng.dev/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Candidates() = %#v, want %#v", got, want)
	}
}
