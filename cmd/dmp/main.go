package main

import (
	"context"
	"fmt"
	"os"

	"github.com/leafney/docker-mirror-proxy/internal/cli"
)

var (
	Version   = "dev"
	GitBranch = "unknown"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	if err := cli.Execute(context.Background(), cli.Config{
		Version:   Version,
		GitBranch: GitBranch,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		Out:       os.Stdout,
		Err:       os.Stderr,
	}, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "[dmp] 执行失败: %v\n", err)
		os.Exit(1)
	}
}
