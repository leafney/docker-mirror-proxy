package image

import "testing"

func TestParseAllowsDockerHubOfficialImages(t *testing.T) {
	tests := []string{"nginx", "nginx:latest", "redis:7"}

	for _, input := range tests {
		ref, err := Parse(input)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", input, err)
		}
		if ref.Type != TypeDockerHubOfficial {
			t.Fatalf("Parse(%q) type = %q, want %q", input, ref.Type, TypeDockerHubOfficial)
		}
		if ref.Original != input {
			t.Fatalf("Parse(%q) original = %q, want %q", input, ref.Original, input)
		}
	}
}

func TestParseAllowsDockerHubNamespaceImages(t *testing.T) {
	tests := []string{
		"lexiforest/curl-impersonate:latest",
		"library/nginx:latest",
	}

	for _, input := range tests {
		ref, err := Parse(input)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", input, err)
		}
		if ref.Type != TypeDockerHubOfficial {
			t.Fatalf("Parse(%q) type = %q, want %q", input, ref.Type, TypeDockerHubOfficial)
		}
		if ref.Original != input {
			t.Fatalf("Parse(%q) original = %q, want %q", input, ref.Original, input)
		}
	}
}

func TestParseAllowsGHCRImages(t *testing.T) {
	input := "ghcr.io/leafney/ai-signin:0.6.8"

	ref, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", input, err)
	}
	if ref.Type != TypeGHCR {
		t.Fatalf("Parse(%q) type = %q, want %q", input, ref.Type, TypeGHCR)
	}
	if ref.Original != input {
		t.Fatalf("Parse(%q) original = %q, want %q", input, ref.Original, input)
	}
}

func TestParseRejectsUnsupportedImages(t *testing.T) {
	tests := []string{
		"quay.io/org/image:tag",
		"registry.k8s.io/pause:3.8",
		"localhost/org/image:tag",
		"repo.local/org/image:tag",
		"repo:5000/org/image:tag",
		"",
	}

	for _, input := range tests {
		if _, err := Parse(input); err == nil {
			t.Fatalf("Parse(%q) returned nil error", input)
		}
	}
}
