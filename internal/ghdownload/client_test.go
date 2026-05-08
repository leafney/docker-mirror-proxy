package ghdownload

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestDetectPrefersCurl(t *testing.T) {
	client := Client{
		lookPath: func(name string) (string, error) {
			if name == "curl" {
				return "/usr/bin/curl", nil
			}
			if name == "wget" {
				return "/usr/bin/wget", nil
			}
			return "", errNotFound
		},
	}

	tool, err := client.Detect()
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if tool != ToolCurl {
		t.Fatalf("tool = %q, want %q", tool, ToolCurl)
	}
}

func TestDetectUsesWgetWhenCurlIsUnavailable(t *testing.T) {
	client := Client{
		lookPath: func(name string) (string, error) {
			if name == "wget" {
				return "/usr/bin/wget", nil
			}
			return "", errNotFound
		},
	}

	tool, err := client.Detect()
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if tool != ToolWget {
		t.Fatalf("tool = %q, want %q", tool, ToolWget)
	}
}

func TestDetectReturnsErrorWhenNoToolExists(t *testing.T) {
	client := Client{
		lookPath: func(name string) (string, error) {
			return "", errNotFound
		},
	}

	if _, err := client.Detect(); err == nil {
		t.Fatal("Detect returned nil error")
	}
}

func TestDownloadRunsCurlCommand(t *testing.T) {
	var gotName string
	var gotArgs []string
	client := Client{
		runCommand: func(ctx context.Context, name string, args ...string) error {
			gotName = name
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	err := client.Download(context.Background(), ToolCurl, "https://proxy/https://github.com/a/b/file.zip", "/tmp", 30*time.Second, "")
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}

	wantArgs := []string{
		"-L",
		"--fail",
		"--connect-timeout", "30",
		"--max-time", "30",
		"-O",
		"--output-dir", "/tmp",
		"https://proxy/https://github.com/a/b/file.zip",
	}
	if gotName != "curl" {
		t.Fatalf("command name = %q, want curl", gotName)
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestDownloadRunsWgetCommand(t *testing.T) {
	var gotName string
	var gotArgs []string
	client := Client{
		runCommand: func(ctx context.Context, name string, args ...string) error {
			gotName = name
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	err := client.Download(context.Background(), ToolWget, "https://proxy/https://github.com/a/b/file.zip", "/tmp", 30*time.Second, "")
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}

	wantArgs := []string{
		"-T", "30",
		"-P", "/tmp",
		"https://proxy/https://github.com/a/b/file.zip",
	}
	if gotName != "wget" {
		t.Fatalf("command name = %q, want wget", gotName)
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestDownloadRunsCurlCommandWithProxy(t *testing.T) {
	var gotArgs []string
	client := Client{
		runCommand: func(ctx context.Context, name string, args ...string) error {
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	err := client.Download(context.Background(), ToolCurl, "https://github.com/a/b/file.zip", "/tmp", 30*time.Second, "http://127.0.0.1:7890")
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}

	wantArgs := []string{
		"-L",
		"--fail",
		"--connect-timeout", "30",
		"--max-time", "30",
		"-O",
		"--output-dir", "/tmp",
		"--proxy", "http://127.0.0.1:7890",
		"https://github.com/a/b/file.zip",
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestDownloadRunsWgetCommandWithProxy(t *testing.T) {
	var gotArgs []string
	client := Client{
		runCommand: func(ctx context.Context, name string, args ...string) error {
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	err := client.Download(context.Background(), ToolWget, "https://github.com/a/b/file.zip", "/tmp", 30*time.Second, "http://127.0.0.1:7890")
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}

	wantArgs := []string{
		"-T", "30",
		"-P", "/tmp",
		"-e", "use_proxy=yes",
		"-e", "http_proxy=http://127.0.0.1:7890",
		"-e", "https_proxy=http://127.0.0.1:7890",
		"https://github.com/a/b/file.zip",
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}
