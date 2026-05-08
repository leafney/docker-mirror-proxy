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

func (c *Client) Pull(ctx context.Context, image, proxy string) error {
	return c.run(ctx, proxy, "pull", image)
}

func (c *Client) Tag(ctx context.Context, source, target string) error {
	return c.run(ctx, "", "tag", source, target)
}

func (c *Client) Remove(ctx context.Context, image string) error {
	return c.run(ctx, "", "rmi", image)
}

func (c *Client) run(ctx context.Context, proxy string, args ...string) error {
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	if proxy != "" {
		cmd.Env = append(os.Environ(),
			"HTTP_PROXY="+proxy,
			"HTTPS_PROXY="+proxy,
			"http_proxy="+proxy,
			"https_proxy="+proxy,
		)
	}
	return cmd.Run()
}
