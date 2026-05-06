package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/leafney/docker-mirror-pull/internal/app"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	flags := flag.NewFlagSet("dmp", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)

	timeoutSeconds := flags.Int("timeout", 60, "单个加速地址的超时时间，单位为秒")
	noClean := flags.Bool("no-clean", false, "成功后不删除临时加速镜像标签")
	showHelp := flags.Bool("help", false, "显示帮助信息")
	showVersion := flags.Bool("version", false, "显示版本信息")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "dmp 是一个 Docker 镜像下载加速工具。")
		fmt.Fprintln(flags.Output())
		fmt.Fprintln(flags.Output(), "用法:")
		fmt.Fprintln(flags.Output(), "  dmp [选项] <镜像> [镜像...]")
		fmt.Fprintln(flags.Output())
		fmt.Fprintln(flags.Output(), "示例:")
		fmt.Fprintln(flags.Output(), "  dmp nginx:latest")
		fmt.Fprintln(flags.Output(), "  dmp ghcr.io/leafney/ai-signin:0.6.8")
		fmt.Fprintln(flags.Output(), "  dmp --timeout 30 nginx:latest")
		fmt.Fprintln(flags.Output())
		fmt.Fprintln(flags.Output(), "选项:")
		flags.PrintDefaults()
	}

	if len(args) == 0 {
		flags.Usage()
		return 0
	}

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *showHelp {
		flags.Usage()
		return 0
	}
	if *showVersion {
		fmt.Fprintf(os.Stdout, "dmp %s\n", version)
		return 0
	}
	if flags.NArg() == 0 {
		flags.Usage()
		return 0
	}
	if *timeoutSeconds <= 0 {
		fmt.Fprintln(os.Stderr, "[dmp] --timeout 必须大于 0")
		return 2
	}

	err := app.Run(context.Background(), app.Options{
		Images:  flags.Args(),
		Timeout: time.Duration(*timeoutSeconds) * time.Second,
		Out:     os.Stdout,
		NoClean: *noClean,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[dmp] 执行失败: %v\n", err)
		return 1
	}
	return 0
}
