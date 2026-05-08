package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	dockerclient "github.com/leafney/docker-mirror-proxy/internal/docker"
	"github.com/leafney/docker-mirror-proxy/internal/image"
	"github.com/leafney/docker-mirror-proxy/internal/logx"
	"github.com/leafney/docker-mirror-proxy/internal/mirror"
)

type Docker interface {
	Pull(ctx context.Context, image string) error
	Tag(ctx context.Context, source, target string) error
	Remove(ctx context.Context, image string) error
}

type Options struct {
	Images  []string
	Timeout time.Duration
	Out     io.Writer
	Docker  Docker
}

func Run(ctx context.Context, opts Options) error {
	if opts.Timeout <= 0 {
		opts.Timeout = 60 * time.Second
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Docker == nil {
		opts.Docker = dockerclient.New(opts.Out, os.Stderr)
	}

	logger := logx.New(opts.Out)
	var errs []error

	for _, original := range opts.Images {
		if err := runOne(ctx, opts, logger, original); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func runOne(ctx context.Context, opts Options, logger logx.Logger, original string) error {
	ref, err := image.Parse(original)
	if err != nil {
		logger.Printf("跳过镜像: %s", original)
		logger.Printf("原因: %v", err)
		return err
	}

	candidates := mirror.Candidates(ref)
	logger.Printf("原始镜像: %s", ref.Original)
	logger.Printf("镜像类型: %s", ref.Type)
	logger.Printf("超时时间: %.0f 秒", opts.Timeout.Seconds())
	logger.Printf("候选加速地址数量: %d", len(candidates))

	var pullErrs []error
	for i, candidate := range candidates {
		logger.Printf("尝试 %d/%d: %s", i+1, len(candidates), candidate)

		pullCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
		err := opts.Docker.Pull(pullCtx, candidate)
		cancel()
		if err != nil {
			logger.Printf("拉取失败，切换下一个加速地址")
			pullErrs = append(pullErrs, fmt.Errorf("%s: %w", candidate, err))
			continue
		}

		logger.Printf("拉取成功: %s", candidate)
		logger.Printf("打标签: %s -> %s", candidate, ref.Original)
		if err := opts.Docker.Tag(ctx, candidate, ref.Original); err != nil {
			return fmt.Errorf("打标签失败: %w", err)
		}

		logger.Printf("清理临时标签: %s", candidate)
		if err := opts.Docker.Remove(ctx, candidate); err != nil {
			logger.Printf("清理临时标签失败: %v", err)
		}

		logger.Printf("完成: %s", ref.Original)
		return nil
	}

	logger.Printf("所有加速地址均拉取失败")
	for i, candidate := range candidates {
		logger.Printf("已尝试 %d/%d: %s", i+1, len(candidates), candidate)
	}
	return fmt.Errorf("镜像拉取失败: %s: %w", ref.Original, errors.Join(pullErrs...))
}
