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

var (
	errNoProgressTimeout = errors.New("拉取超过无进展超时时间")
	errMaxTimeTimeout    = errors.New("拉取超过总耗时上限")
)

type Docker interface {
	Pull(ctx context.Context, image string, onProgress func()) error
	Tag(ctx context.Context, source, target string) error
	Remove(ctx context.Context, image string) error
}

type Options struct {
	Images  []string
	Timeout time.Duration
	MaxTime time.Duration
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
	logger.Printf("无进展超时时间: %.0f 秒", opts.Timeout.Seconds())
	if opts.MaxTime > 0 {
		logger.Printf("单个加速地址总耗时上限: %.0f 秒", opts.MaxTime.Seconds())
	} else {
		logger.Printf("单个加速地址总耗时上限: 不限制")
	}
	logger.Printf("候选加速地址数量: %d", len(candidates))

	var pullErrs []error
	for i, candidate := range candidates {
		logger.Printf("尝试 %d/%d: %s", i+1, len(candidates), candidate)

		pullCtx, cancel := context.WithCancelCause(ctx)
		markProgress, stopMonitor := startPullMonitor(pullCtx, cancel, opts.Timeout, opts.MaxTime)
		err := opts.Docker.Pull(pullCtx, candidate, markProgress)
		stopMonitor()
		cause := context.Cause(pullCtx)

		if err != nil {
			switch {
			case errors.Is(cause, errNoProgressTimeout):
				logger.Printf("无进展超过 %.0f 秒，切换下一个加速地址", opts.Timeout.Seconds())
				pullErrs = append(pullErrs, fmt.Errorf("%s: %w", candidate, cause))
			case errors.Is(cause, errMaxTimeTimeout):
				logger.Printf("总耗时超过 %.0f 秒，切换下一个加速地址", opts.MaxTime.Seconds())
				pullErrs = append(pullErrs, fmt.Errorf("%s: %w", candidate, cause))
			case cause != nil:
				return fmt.Errorf("镜像拉取取消: %s: %w", candidate, cause)
			default:
				logger.Printf("拉取失败，切换下一个加速地址")
				pullErrs = append(pullErrs, fmt.Errorf("%s: %w", candidate, err))
			}
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

func startPullMonitor(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	noProgressTimeout time.Duration,
	maxTime time.Duration,
) (func(), func()) {
	progressCh := make(chan struct{}, 1)
	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)

		var noProgressTimer *time.Timer
		if noProgressTimeout > 0 {
			noProgressTimer = time.NewTimer(noProgressTimeout)
			defer stopTimer(noProgressTimer)
		}

		var maxTimeTimer *time.Timer
		if maxTime > 0 {
			maxTimeTimer = time.NewTimer(maxTime)
			defer stopTimer(maxTimeTimer)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-stopCh:
				return
			case <-progressCh:
				if noProgressTimer != nil {
					resetTimer(noProgressTimer, noProgressTimeout)
				}
			case <-timerChan(noProgressTimer):
				cancel(errNoProgressTimeout)
				return
			case <-timerChan(maxTimeTimer):
				cancel(errMaxTimeTimeout)
				return
			}
		}
	}()

	markProgress := func() {
		select {
		case progressCh <- struct{}{}:
		default:
		}
	}
	stopMonitor := func() {
		close(stopCh)
		<-doneCh
	}
	return markProgress, stopMonitor
}

func timerChan(timer *time.Timer) <-chan time.Time {
	if timer == nil {
		return nil
	}
	return timer.C
}

func resetTimer(timer *time.Timer, d time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(d)
}

func stopTimer(timer *time.Timer) {
	if timer == nil {
		return
	}
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}
