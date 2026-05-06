package cli

import (
	"bytes"
	"context"
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
	for _, name := range []string{"gh", "pip", "npm"} {
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
