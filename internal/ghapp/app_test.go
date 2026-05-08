package ghapp

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/leafney/docker-mirror-proxy/internal/ghdownload"
)

type downloadCall struct {
	Tool      ghdownload.Tool
	URL       string
	OutputDir string
	Timeout   time.Duration
}

type fakeDownloader struct {
	tool           ghdownload.Tool
	detectErr      error
	downloadErrors map[string]error
	calls          []downloadCall
}

func (f *fakeDownloader) Detect() (ghdownload.Tool, error) {
	if f.detectErr != nil {
		return "", f.detectErr
	}
	return f.tool, nil
}

func (f *fakeDownloader) Download(ctx context.Context, tool ghdownload.Tool, url, outputDir string, timeout time.Duration) error {
	f.calls = append(f.calls, downloadCall{
		Tool:      tool,
		URL:       url,
		OutputDir: outputDir,
		Timeout:   timeout,
	})
	if err, ok := f.downloadErrors[url]; ok {
		return err
	}
	return nil
}

func TestRunDownloadsWithNextMirrorAfterFailure(t *testing.T) {
	tmp := t.TempDir()
	downloader := &fakeDownloader{
		tool: ghdownload.ToolCurl,
		downloadErrors: map[string]error{
			"https://ghfast.top/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz": errors.New("failed"),
		},
	}
	var out bytes.Buffer

	err := Run(context.Background(), Options{
		URLs:       []string{"https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz"},
		Timeout:    time.Second,
		Output:     tmp,
		Out:        &out,
		Downloader: downloader,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := []downloadCall{
		{
			Tool:      ghdownload.ToolCurl,
			URL:       "https://ghfast.top/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
			OutputDir: tmp,
			Timeout:   time.Second,
		},
		{
			Tool:      ghdownload.ToolCurl,
			URL:       "https://gh-proxy.com/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz",
			OutputDir: tmp,
			Timeout:   time.Second,
		},
	}
	if !reflect.DeepEqual(downloader.calls, want) {
		t.Fatalf("calls = %#v, want %#v", downloader.calls, want)
	}

	output := out.String()
	if !strings.Contains(output, "下载失败，切换下一个加速地址") {
		t.Fatalf("log output missing failure switch message: %s", output)
	}
	if !strings.Contains(output, "下载目录: "+tmp) {
		t.Fatalf("log output missing absolute output dir: %s", output)
	}
	if !strings.Contains(output, "文件位置: "+filepath.Join(tmp, "dmp-linux-amd64.tar.gz")) {
		t.Fatalf("log output missing final file path: %s", output)
	}
}

func TestRunReturnsErrorWhenOutputDirDoesNotExist(t *testing.T) {
	downloader := &fakeDownloader{tool: ghdownload.ToolCurl}

	err := Run(context.Background(), Options{
		URLs:       []string{"https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz"},
		Timeout:    time.Second,
		Output:     filepath.Join(t.TempDir(), "missing"),
		Downloader: downloader,
	})
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if len(downloader.calls) != 0 {
		t.Fatalf("download calls = %#v, want none", downloader.calls)
	}
}

func TestRunReturnsErrorWhenDownloaderToolIsMissing(t *testing.T) {
	tmp := t.TempDir()
	downloader := &fakeDownloader{detectErr: errors.New("missing tool")}

	err := Run(context.Background(), Options{
		URLs:       []string{"https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz"},
		Timeout:    time.Second,
		Output:     tmp,
		Downloader: downloader,
	})
	if err == nil {
		t.Fatal("Run returned nil error")
	}
}
