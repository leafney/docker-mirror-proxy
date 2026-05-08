package ghurl

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

type Ref struct {
	Original string
	Filename string
}

func Parse(input string) (Ref, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return Ref{}, fmt.Errorf("下载地址不能为空")
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return Ref{}, fmt.Errorf("下载地址格式错误: %s", input)
	}
	if u.Scheme != "https" {
		return Ref{}, fmt.Errorf("仅支持 https 下载地址: %s", input)
	}
	if !isSupportedHost(u.Hostname()) {
		return Ref{}, fmt.Errorf("仅支持 GitHub 下载地址: %s", input)
	}
	if isGitHubDirectoryURL(u) {
		return Ref{}, fmt.Errorf("不支持整个项目文件夹下载: %s", input)
	}
	if isGitHubRepoRootURL(u) {
		return Ref{}, fmt.Errorf("不支持整个项目文件夹下载: %s", input)
	}

	filename := path.Base(u.Path)
	if filename == "." || filename == "/" || filename == "" {
		return Ref{}, fmt.Errorf("无法从下载地址解析文件名: %s", input)
	}

	return Ref{Original: trimmed, Filename: filename}, nil
}

func isSupportedHost(host string) bool {
	switch strings.ToLower(host) {
	case "github.com", "raw.githubusercontent.com", "gist.githubusercontent.com":
		return true
	default:
		return false
	}
}

func isGitHubDirectoryURL(u *url.URL) bool {
	if strings.ToLower(u.Hostname()) != "github.com" {
		return false
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	return len(parts) >= 5 && parts[2] == "tree"
}

func isGitHubRepoRootURL(u *url.URL) bool {
	if strings.ToLower(u.Hostname()) != "github.com" {
		return false
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	return len(parts) == 2
}
