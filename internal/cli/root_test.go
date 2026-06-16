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

func TestPullCommandSupportsShortTimeoutFlag(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"pull", "-t", "0", "nginx:latest"})

	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
	if !strings.Contains(err.Error(), "--timeout 必须大于 0") {
		t.Fatalf("error = %q, want timeout validation", err.Error())
	}
}

func TestPullCommandSupportsShortMaxTimeFlag(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"pull", "-m", "-1", "nginx:latest"})

	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
	if !strings.Contains(err.Error(), "--max-time 不能小于 0") {
		t.Fatalf("error = %q, want max-time validation", err.Error())
	}
}

func TestPullCommandSupportsLongMaxTimeFlag(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"pull", "--max-time", "-1", "nginx:latest"})

	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
	if !strings.Contains(err.Error(), "--max-time 不能小于 0") {
		t.Fatalf("error = %q, want max-time validation", err.Error())
	}
}

func TestPullCommandRejectsProxyFlag(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"pull", "-p", "http://127.0.0.1:7890", "-t", "0", "nginx:latest"})

	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
	if !strings.Contains(err.Error(), "unknown shorthand flag: 'p' in -p") {
		t.Fatalf("error = %q, want unknown proxy flag", err.Error())
	}
}

func TestPullCommandRejectsNoCleanFlag(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"pull", "--no-clean", "-t", "0", "nginx:latest"})

	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
	if !strings.Contains(err.Error(), "unknown flag: --no-clean") {
		t.Fatalf("error = %q, want unknown no-clean flag", err.Error())
	}
}

func TestPullProxyCommandRequiresSubcommand(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"pull", "proxy"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("ExecuteContext returned error: %v", err)
	}
	if !strings.Contains(out.String(), "set") || !strings.Contains(out.String(), "rm") || !strings.Contains(out.String(), "sts") {
		t.Fatalf("proxy help missing subcommands: %s", out.String())
	}
}

func TestPullProxySetRequiresProxyArgument(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"pull", "proxy", "set"})

	if err := cmd.ExecuteContext(context.Background()); err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
}

func TestPullProxyRmRejectsArguments(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"pull", "proxy", "rm", "extra"})

	if err := cmd.ExecuteContext(context.Background()); err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
}

func TestPullProxyStsRejectsArguments(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"pull", "proxy", "sts", "extra"})

	if err := cmd.ExecuteContext(context.Background()); err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
}

func TestGhCommandSupportsProxyFlag(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"gh", "-p", "http://127.0.0.1:7890", "-t", "0", "https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz"})

	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
	if !strings.Contains(err.Error(), "--timeout 必须大于 0") {
		t.Fatalf("error = %q, want timeout validation", err.Error())
	}
}

func TestRootVersionFlagShowsBuildInfo(t *testing.T) {
	for _, args := range [][]string{{"-v"}, {"--version"}} {
		var out bytes.Buffer
		cmd := New(Config{
			Version:   "v1.2.3",
			GitBranch: "main",
			GitCommit: "abc1234",
			BuildTime: "2026-05-06 10:11:12",
			Out:       &out,
			Err:       &out,
		})
		cmd.SetArgs(args)

		if err := cmd.ExecuteContext(context.Background()); err != nil {
			t.Fatalf("ExecuteContext(%v) returned error: %v", args, err)
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
}

func TestVersionCommandIsRejected(t *testing.T) {
	var out bytes.Buffer
	cmd := New(Config{
		Version: "test",
		Out:     &out,
		Err:     &out,
	})
	cmd.SetArgs([]string{"version"})

	if err := cmd.ExecuteContext(context.Background()); err == nil {
		t.Fatal("ExecuteContext returned nil error")
	}
}
