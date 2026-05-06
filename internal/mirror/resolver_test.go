package mirror

import (
	"reflect"
	"testing"

	"github.com/leafney/docker-mirror-pull/internal/image"
)

func TestCandidatesForDockerHubOfficialImage(t *testing.T) {
	ref, err := image.Parse("nginx:latest")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	got := Candidates(ref)
	want := []string{
		"docker.1ms.run/nginx:latest",
		"dockerproxy.net/nginx:latest",
		"proxy.vvvv.ee/nginx:latest",
		"registry.cyou/nginx:latest",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Candidates() = %#v, want %#v", got, want)
	}
}

func TestCandidatesForGHCRImage(t *testing.T) {
	ref, err := image.Parse("ghcr.io/leafney/ai-signin:0.6.8")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	got := Candidates(ref)
	want := []string{
		"docker.1ms.run/ghcr.io/leafney/ai-signin:0.6.8",
		"dockerproxy.net/ghcr.io/leafney/ai-signin:0.6.8",
		"proxy.vvvv.ee/ghcr.io/leafney/ai-signin:0.6.8",
		"registry.cyou/ghcr.io/leafney/ai-signin:0.6.8",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Candidates() = %#v, want %#v", got, want)
	}
}
