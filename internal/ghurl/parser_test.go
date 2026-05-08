package ghurl

import "testing"

func TestParseAllowsSupportedGitHubURLs(t *testing.T) {
	tests := []struct {
		input    string
		filename string
	}{
		{
			input:    "https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
			filename: "dmp-linux-amd64.tar.gz",
		},
		{
			input:    "http://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
			filename: "dmp-linux-amd64.tar.gz",
		},
		{
			input:    "https://github.com/stilleshan/dockerfiles/archive/master.zip",
			filename: "master.zip",
		},
		{
			input:    "https://raw.githubusercontent.com/leafney/docker-mirror-proxy/main/README.md",
			filename: "README.md",
		},
		{
			input:    "https://gist.githubusercontent.com/user/abc/raw/test.sh",
			filename: "test.sh",
		},
	}

	for _, tt := range tests {
		ref, err := Parse(tt.input)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
		}
		if ref.Original != tt.input {
			t.Fatalf("Parse(%q) original = %q, want %q", tt.input, ref.Original, tt.input)
		}
		if ref.Filename != tt.filename {
			t.Fatalf("Parse(%q) filename = %q, want %q", tt.input, ref.Filename, tt.filename)
		}
	}
}

func TestParseRejectsUnsupportedURLs(t *testing.T) {
	tests := []string{
		"https://example.com/file.zip",
		"https://github.com/leafney/docker-mirror-proxy/tree/main/docs",
		"https://github.com/leafney/docker-mirror-proxy/",
		"ftp://github.com/leafney/file.zip",
		"",
	}

	for _, input := range tests {
		if _, err := Parse(input); err == nil {
			t.Fatalf("Parse(%q) returned nil error", input)
		}
	}
}
