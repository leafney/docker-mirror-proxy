package main

import (
	"context"
	"fmt"
	"os"

	"github.com/leafney/docker-mirror-proxy/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Execute(context.Background(), cli.Config{
		Version: version,
		Out:     os.Stdout,
		Err:     os.Stderr,
	}, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "[dmp] 执行失败: %v\n", err)
		os.Exit(1)
	}
}
