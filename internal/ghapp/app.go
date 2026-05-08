package ghapp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/leafney/docker-mirror-proxy/internal/ghdownload"
	"github.com/leafney/docker-mirror-proxy/internal/ghmirror"
	"github.com/leafney/docker-mirror-proxy/internal/ghurl"
	"github.com/leafney/docker-mirror-proxy/internal/logx"
)

type Downloader interface {
	Detect() (ghdownload.Tool, error)
	Download(ctx context.Context, tool ghdownload.Tool, url, outputDir string, timeout time.Duration) error
}

type Options struct {
	URLs       []string
	Timeout    time.Duration
	Output     string
	Out        io.Writer
	Downloader Downloader
}

func Run(ctx context.Context, opts Options) error {
	if opts.Timeout <= 0 {
		opts.Timeout = 60 * time.Second
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Output == "" {
		opts.Output = "."
	}
	if opts.Downloader == nil {
		opts.Downloader = ghdownload.New(opts.Out, os.Stderr)
	}

	logger := logx.New(opts.Out)

	outputDir, err := prepareOutputDir(opts.Output)
	if err != nil {
		logger.Printf("下载目录不可用: %s", opts.Output)
		logger.Printf("原因: %v", err)
		return err
	}

	tool, err := opts.Downloader.Detect()
	if err != nil {
		logger.Printf("%v", err)
		return err
	}

	var errs []error
	for _, rawURL := range opts.URLs {
		if err := runOne(ctx, opts, logger, tool, outputDir, rawURL); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func runOne(ctx context.Context, opts Options, logger logx.Logger, tool ghdownload.Tool, outputDir, rawURL string) error {
	ref, err := ghurl.Parse(rawURL)
	if err != nil {
		logger.Printf("跳过地址: %s", rawURL)
		logger.Printf("原因: %v", err)
		return err
	}

	candidates := ghmirror.Candidates(ref)
	logger.Printf("原始地址: %s", ref.Original)
	logger.Printf("下载目录: %s", outputDir)
	logger.Printf("下载工具: %s", tool)
	logger.Printf("超时时间: %.0f 秒", opts.Timeout.Seconds())
	logger.Printf("候选加速地址数量: %d", len(candidates))

	var downloadErrs []error
	for i, candidate := range candidates {
		logger.Printf("尝试 %d/%d: %s", i+1, len(candidates), candidate)

		downloadCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
		err := opts.Downloader.Download(downloadCtx, tool, candidate, outputDir, opts.Timeout)
		cancel()
		if err != nil {
			logger.Printf("下载失败，切换下一个加速地址")
			downloadErrs = append(downloadErrs, fmt.Errorf("%s: %w", candidate, err))
			continue
		}

		logger.Printf("下载成功")
		logger.Printf("文件位置: %s", filepath.Join(outputDir, ref.Filename))
		return nil
	}

	logger.Printf("所有加速地址均下载失败")
	for i, candidate := range candidates {
		logger.Printf("已尝试 %d/%d: %s", i+1, len(candidates), candidate)
	}
	return fmt.Errorf("文件下载失败: %s: %w", ref.Original, errors.Join(downloadErrs...))
}

func prepareOutputDir(input string) (string, error) {
	abs, err := filepath.Abs(input)
	if err != nil {
		return "", fmt.Errorf("解析下载目录失败: %w", err)
	}
	stat, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("下载目录不存在: %s", abs)
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("下载目录不是文件夹: %s", abs)
	}
	return abs, nil
}
