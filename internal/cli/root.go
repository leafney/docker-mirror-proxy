package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/leafney/docker-mirror-proxy/internal/app"
	"github.com/leafney/docker-mirror-proxy/internal/ghapp"
	"github.com/leafney/docker-mirror-proxy/internal/pullproxy"
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
	var showVersion bool

	root := &cobra.Command{
		Use:           "dmp",
		Short:         "Docker Mirror Proxy",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				printVersion(cmd.OutOrStdout(), cfg)
				return nil
			}
			if len(args) > 0 {
				return fmt.Errorf("未知命令: %s，请使用 dmp pull <镜像>", args[0])
			}
			return cmd.Help()
		},
	}
	root.SetOut(cfg.Out)
	root.SetErr(cfg.Err)
	root.SetHelpCommand(&cobra.Command{Hidden: true})
	root.CompletionOptions.DisableDefaultCmd = true
	root.Flags().BoolVarP(&showVersion, "version", "v", false, "显示版本信息")

	root.AddCommand(newPullCommand(cfg.Out))
	root.AddCommand(newGhCommand(cfg.Out))
	root.AddCommand(newReservedCommand("pip", "Python 包加速"))
	root.AddCommand(newReservedCommand("npm", "Node 包加速"))

	return root
}

func newPullCommand(out io.Writer) *cobra.Command {
	var timeoutSeconds int
	var maxTimeSeconds int

	cmd := &cobra.Command{
		Use:   "pull <镜像> [镜像...]",
		Short: "拉取 Docker 镜像",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if timeoutSeconds <= 0 {
				return fmt.Errorf("--timeout 必须大于 0")
			}
			if maxTimeSeconds < 0 {
				return fmt.Errorf("--max-time 不能小于 0")
			}
			return app.Run(cmd.Context(), app.Options{
				Images:  args,
				Timeout: time.Duration(timeoutSeconds) * time.Second,
				MaxTime: time.Duration(maxTimeSeconds) * time.Second,
				Out:     out,
			})
		},
	}

	cmd.Flags().IntVarP(&timeoutSeconds, "timeout", "t", 60, "单个加速地址的无进展超时时间，单位为秒")
	cmd.Flags().IntVarP(&maxTimeSeconds, "max-time", "m", 0, "单个加速地址的总耗时上限，单位为秒，0 表示不限制")
	cmd.AddCommand(newPullProxyCommand(out))
	return cmd
}

func newPullProxyCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "管理 Docker daemon 代理配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newPullProxySetCommand(out))
	cmd.AddCommand(newPullProxyRmCommand(out))
	cmd.AddCommand(newPullProxyStsCommand(out))
	return cmd
}

func newPullProxySetCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <proxy>",
		Short: "设置 Docker daemon 代理",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pullproxy.New(out).Set(cmd.Context(), args[0])
		},
	}
	return cmd
}

func newPullProxyRmCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm",
		Short: "移除 Docker daemon 代理",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pullproxy.New(out).Remove(cmd.Context())
		},
	}
	return cmd
}

func newPullProxyStsCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sts",
		Short: "查看 Docker daemon 代理状态",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pullproxy.New(out).Status(cmd.Context())
		},
	}
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

func printVersion(out io.Writer, cfg Config) {
	fmt.Fprintf(out, "Version: %s\n", cfg.Version)
	fmt.Fprintf(out, "Git Branch: %s\n", cfg.GitBranch)
	fmt.Fprintf(out, "Git Commit: %s\n", cfg.GitCommit)
	fmt.Fprintf(out, "Build Time: %s\n", cfg.BuildTime)
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
	case "pull", "gh", "pip", "npm":
		return true
	default:
		return false
	}
}
