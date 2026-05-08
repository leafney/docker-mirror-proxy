package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRootShowsHelpWithoutArgs(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs(nil)

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("ExecuteContext returned error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("pull")) {
		t.Fatalf("help output missing commands: %s", out.String())
	}
}

func TestRootRejectsImageShortcut(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"nginx:latest"})

	if err := cmd.ExecuteContext(context.Background()); err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
}

func TestReservedCommandsReturnNotImplemented(t *testing.T) {
	for _, name := range []string{"pip", "npm"} {
		var out bytes.Buffer
		cmd := New(Config{
			Version: "test",
			Out:     &out,
			Err:     &out,
		})
		cmd.SetArgs([]string{name})

		if err := cmd.ExecuteContext(context.Background()); err == nil {
			t.Fatalf("%s returned nil error", name)
		}
		if !bytes.Contains(out.Bytes(), []byte("暂未实现")) {
			t.Fatalf("%s output missing not implemented message: %s", name, out.String())
		}
	}
}

func TestGhCommandRequiresURL(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"gh"})

	if err := cmd.ExecuteContext(context.Background()); err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
}

func TestGhCommandSupportsShortFlags(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"gh", "-t", "0", "-o", ".", "https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz"})

	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
	if !strings.Contains(err.Error(), "--timeout 必须大于 0") {
		t.Fatalf("error = %q, want timeout validation", err.Error())
	}
}

func TestVersionCommandShowsBuildInfo(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version:   "v1.2.3",
		GitBranch: "main",
		GitCommit: "abc1234",
		BuildTime: "2026-05-06 10:11:12",
		Out:       &out,
		Err:       &out,
	})
	cmd.SetArgs([]string{"version"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("ExecuteContext returned error: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		"Version: v1.2.3",
		"Git Branch: main",
		"Git Commit: abc1234",
		"Build Time: 2026-05-06 10:11:12",
	} {
		if !bytes.Contains([]byte(output), []byte(want)) {
			t.Fatalf("version output missing %q: %s", want, output)
		}
	}
}
