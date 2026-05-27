package docker

import (
	"context"
	"io"
	"os"
	"os/exec"
)

type Client struct {
	stdout io.Writer
	stderr io.Writer
}

func New(stdout, stderr io.Writer) *Client {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &Client{stdout: stdout, stderr: stderr}
}

func (c *Client) Pull(ctx context.Context, image string, onProgress func()) error {
	return c.run(
		ctx,
		newProgressWriter(c.stdout, onProgress),
		newProgressWriter(c.stderr, onProgress),
		"pull",
		image,
	)
}

func (c *Client) Tag(ctx context.Context, source, target string) error {
	return c.run(ctx, c.stdout, c.stderr, "tag", source, target)
}

func (c *Client) Remove(ctx context.Context, image string) error {
	return c.run(ctx, c.stdout, c.stderr, "rmi", image)
}

func (c *Client) run(ctx context.Context, stdout, stderr io.Writer, args ...string) error {
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

type progressWriter struct {
	target     io.Writer
	onProgress func()
}

func newProgressWriter(target io.Writer, onProgress func()) io.Writer {
	if target == nil && onProgress == nil {
		return nil
	}
	return progressWriter{
		target:     target,
		onProgress: onProgress,
	}
}

func (w progressWriter) Write(p []byte) (int, error) {
	if len(p) > 0 && w.onProgress != nil {
		w.onProgress()
	}
	if w.target == nil {
		return len(p), nil
	}
	return w.target.Write(p)
}
