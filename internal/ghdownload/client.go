package ghdownload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Tool string

const (
	ToolCurl Tool = "curl"
	ToolWget Tool = "wget"
)

var errNotFound = errors.New("not found")

type Client struct {
	stdout     io.Writer
	stderr     io.Writer
	lookPath   func(string) (string, error)
	runCommand func(context.Context, string, ...string) error
}

func New(stdout, stderr io.Writer) *Client {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	client := &Client{stdout: stdout, stderr: stderr}
	client.lookPath = exec.LookPath
	client.runCommand = client.run
	return client
}

func (c *Client) Detect() (Tool, error) {
	lookPath := c.lookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if _, err := lookPath(string(ToolCurl)); err == nil {
		return ToolCurl, nil
	}
	if _, err := lookPath(string(ToolWget)); err == nil {
		return ToolWget, nil
	}
	return "", fmt.Errorf("当前系统未找到 curl 或 wget，请先安装其中一个下载工具")
}

func (c *Client) Download(ctx context.Context, tool Tool, url, outputDir string, timeout time.Duration) error {
	timeoutSeconds := strconv.Itoa(int(timeout.Seconds()))
	var args []string
	switch tool {
	case ToolCurl:
		args = []string{
			"-L",
			"--fail",
			"--connect-timeout", timeoutSeconds,
			"--max-time", timeoutSeconds,
			"-O",
			"--output-dir", outputDir,
			url,
		}
	case ToolWget:
		args = []string{
			"-T", timeoutSeconds,
			"-P", outputDir,
			url,
		}
	default:
		return fmt.Errorf("不支持的下载工具: %s", tool)
	}

	runCommand := c.runCommand
	if runCommand == nil {
		runCommand = c.run
	}
	return runCommand(ctx, string(tool), args...)
}

func (c *Client) run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	return cmd.Run()
}
