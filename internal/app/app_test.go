package app

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type call struct {
	Name string
	Args []string
}

type fakeDocker struct {
	pullErrors   map[string]error
	pullBehavior func(ctx context.Context, image string, onProgress func()) error
	calls        []call
}

func (f *fakeDocker) Pull(ctx context.Context, image string, onProgress func()) error {
	f.calls = append(f.calls, call{Name: "pull", Args: []string{image}})
	if f.pullBehavior != nil {
		return f.pullBehavior(ctx, image, onProgress)
	}
	if err, ok := f.pullErrors[image]; ok {
		return err
	}
	return nil
}

func (f *fakeDocker) Tag(ctx context.Context, source, target string) error {
	f.calls = append(f.calls, call{Name: "tag", Args: []string{source, target}})
	return nil
}

func (f *fakeDocker) Remove(ctx context.Context, image string) error {
	f.calls = append(f.calls, call{Name: "rmi", Args: []string{image}})
	return nil
}

func TestRunTriesNextMirrorAfterPullFailure(t *testing.T) {
	docker := &fakeDocker{
		pullErrors: map[string]error{
			"docker.1ms.run/nginx:latest": errors.New("failed"),
		},
	}
	var out bytes.Buffer

	err := Run(context.Background(), Options{
		Images:  []string{"nginx:latest"},
		Timeout: time.Second,
		Out:     &out,
		Docker:  docker,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := []call{
		{Name: "pull", Args: []string{"docker.1ms.run/nginx:latest"}},
		{Name: "pull", Args: []string{"dockerproxy.net/nginx:latest"}},
		{Name: "tag", Args: []string{"dockerproxy.net/nginx:latest", "nginx:latest"}},
		{Name: "rmi", Args: []string{"dockerproxy.net/nginx:latest"}},
	}
	if !reflect.DeepEqual(docker.calls, want) {
		t.Fatalf("calls = %#v, want %#v", docker.calls, want)
	}
	if !strings.Contains(out.String(), "拉取失败，切换下一个加速地址") {
		t.Fatalf("log output missing failure switch message: %s", out.String())
	}
	if strings.Contains(out.String(), "拉取失败，切换下一个加速地址: docker.1ms.run/nginx:latest") {
		t.Fatalf("failure switch log should not repeat failed candidate: %s", out.String())
	}
}

func TestRunReturnsErrorWhenUnsupportedImageIsProvided(t *testing.T) {
	docker := &fakeDocker{}
	var out bytes.Buffer

	err := Run(context.Background(), Options{
		Images:  []string{"quay.io/org/image:tag"},
		Timeout: time.Second,
		Out:     &out,
		Docker:  docker,
	})
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if len(docker.calls) != 0 {
		t.Fatalf("docker calls = %#v, want none", docker.calls)
	}
}

func TestRunTriesNextMirrorAfterNoProgressTimeout(t *testing.T) {
	docker := &fakeDocker{
		pullBehavior: func(ctx context.Context, image string, onProgress func()) error {
			if image == "docker.1ms.run/nginx:latest" {
				<-ctx.Done()
				return ctx.Err()
			}
			return nil
		},
	}
	var out bytes.Buffer

	err := Run(context.Background(), Options{
		Images:  []string{"nginx:latest"},
		Timeout: 25 * time.Millisecond,
		Out:     &out,
		Docker:  docker,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(out.String(), "无进展超过") {
		t.Fatalf("log output missing no progress timeout message: %s", out.String())
	}
}

func TestRunDoesNotTimeoutWhileProgressContinues(t *testing.T) {
	docker := &fakeDocker{
		pullBehavior: func(ctx context.Context, image string, onProgress func()) error {
			if image != "docker.1ms.run/nginx:latest" {
				return nil
			}

			ticker := time.NewTicker(5 * time.Millisecond)
			defer ticker.Stop()
			done := time.After(80 * time.Millisecond)

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-ticker.C:
					onProgress()
				case <-done:
					return nil
				}
			}
		},
	}
	var out bytes.Buffer

	err := Run(context.Background(), Options{
		Images:  []string{"nginx:latest"},
		Timeout: 25 * time.Millisecond,
		Out:     &out,
		Docker:  docker,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if strings.Contains(out.String(), "无进展超过") {
		t.Fatalf("log output should not contain no progress timeout: %s", out.String())
	}
	want := []call{
		{Name: "pull", Args: []string{"docker.1ms.run/nginx:latest"}},
		{Name: "tag", Args: []string{"docker.1ms.run/nginx:latest", "nginx:latest"}},
		{Name: "rmi", Args: []string{"docker.1ms.run/nginx:latest"}},
	}
	if !reflect.DeepEqual(docker.calls, want) {
		t.Fatalf("calls = %#v, want %#v", docker.calls, want)
	}
}

func TestRunTriesNextMirrorAfterMaxTimeTimeout(t *testing.T) {
	docker := &fakeDocker{
		pullBehavior: func(ctx context.Context, image string, onProgress func()) error {
			if image == "docker.1ms.run/nginx:latest" {
				ticker := time.NewTicker(5 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-ticker.C:
						onProgress()
					}
				}
			}
			return nil
		},
	}
	var out bytes.Buffer

	err := Run(context.Background(), Options{
		Images:  []string{"nginx:latest"},
		Timeout: 200 * time.Millisecond,
		MaxTime: 30 * time.Millisecond,
		Out:     &out,
		Docker:  docker,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(out.String(), "总耗时超过") {
		t.Fatalf("log output missing max time timeout message: %s", out.String())
	}
}
