package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/leafney/docker-mirror-proxy/internal/app"
	"github.com/leafney/docker-mirror-proxy/internal/ghapp"
	"github.com/spf13/cobra"
)

type Config struct {
	Version   string
	GitBranch string
	GitCommit string
	BuildTime string
	Out       io.Writer
	Err       io.Writer
}

func New(cfg Config) *cobra.Command {
	if cfg.Out == nil {
		cfg.Out = os.Stdout
	}
	if cfg.Err == nil {
		cfg.Err = os.Stderr
	}

	root := &cobra.Command{
		Use:           "dmp",
		Short:         "Docker Mirror Proxy",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("未知命令: %s，请使用 dmp pull <镜像>", args[0])
			}
			return cmd.Help()
		},
	}
	root.SetOut(cfg.Out)
	root.SetErr(cfg.Err)
	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(newPullCommand(cfg.Out))
	root.AddCommand(newGhCommand(cfg.Out))
	root.AddCommand(newReservedCommand("pip", "Python 包加速"))
	root.AddCommand(newReservedCommand("npm", "Node 包加速"))
	root.AddCommand(newVersionCommand(cfg))

	return root
}

func newPullCommand(out io.Writer) *cobra.Command {
	var timeoutSeconds int

	cmd := &cobra.Command{
		Use:   "pull <镜像> [镜像...]",
		Short: "拉取 Docker 镜像",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if timeoutSeconds <= 0 {
				return fmt.Errorf("--timeout 必须大于 0")
			}
			return app.Run(cmd.Context(), app.Options{
				Images:  args,
				Timeout: time.Duration(timeoutSeconds) * time.Second,
				Out:     out,
			})
		},
	}

	cmd.Flags().IntVarP(&timeoutSeconds, "timeout", "t", 60, "单个加速地址的超时时间，单位为秒")
	return cmd
}

func newGhCommand(out io.Writer) *cobra.Command {
	var timeoutSeconds int
	var output string
	var proxy string

	cmd := &cobra.Command{
		Use:   "gh <url> [url...]",
		Short: "下载 GitHub 文件",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if timeoutSeconds <= 0 {
				return fmt.Errorf("--timeout 必须大于 0")
			}
			return ghapp.Run(cmd.Context(), ghapp.Options{
				URLs:    args,
				Timeout: time.Duration(timeoutSeconds) * time.Second,
				Output:  output,
				Out:     out,
				Proxy:   proxy,
			})
		},
	}

	cmd.Flags().IntVarP(&timeoutSeconds, "timeout", "t", 60, "单个加速地址的超时时间，单位为秒")
	cmd.Flags().StringVarP(&output, "output", "o", ".", "下载目录")
	cmd.Flags().StringVarP(&proxy, "proxy", "p", "", "代理地址，例如 http://127.0.0.1:7890")
	return cmd
}

func newReservedCommand(name, desc string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: desc,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "[dmp] %s 功能暂未实现\n", desc)
			return fmt.Errorf("%s 功能暂未实现", desc)
		},
	}
}

func newVersionCommand(cfg Config) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "Version: %s\n", cfg.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Git Branch: %s\n", cfg.GitBranch)
			fmt.Fprintf(cmd.OutOrStdout(), "Git Commit: %s\n", cfg.GitCommit)
			fmt.Fprintf(cmd.OutOrStdout(), "Build Time: %s\n", cfg.BuildTime)
		},
	}
}

func Execute(ctx context.Context, cfg Config, args []string) error {
	if len(args) > 0 && !isRootFlag(args[0]) && !isKnownCommand(args[0]) {
		return fmt.Errorf("未知命令: %s，请使用 dmp pull <镜像>", args[0])
	}
	cmd := New(cfg)
	cmd.SetArgs(args)
	return cmd.ExecuteContext(ctx)
}

func isRootFlag(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}

func isKnownCommand(arg string) bool {
	switch arg {
	case "pull", "gh", "pip", "npm", "version", "help":
		return true
	default:
		return false
	}
}
