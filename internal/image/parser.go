package image

import (
	"fmt"
	"strings"
)

type Type string

const (
	TypeDockerHubOfficial Type = "dockerhub-official"
	TypeGHCR              Type = "ghcr"
)

type Ref struct {
	Original string
	Type     Type
}

func Parse(input string) (Ref, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return Ref{}, fmt.Errorf("镜像名称不能为空")
	}
	if strings.ContainsAny(trimmed, " \t\n\r") {
		return Ref{}, fmt.Errorf("镜像名称不能包含空白字符: %s", input)
	}

	if strings.HasPrefix(trimmed, "ghcr.io/") {
		rest := strings.TrimPrefix(trimmed, "ghcr.io/")
		if rest == "" || !strings.Contains(rest, "/") {
			return Ref{}, fmt.Errorf("ghcr.io 镜像必须包含 owner 和 image: %s", input)
		}
		return Ref{Original: trimmed, Type: TypeGHCR}, nil
	}

	if strings.Contains(trimmed, "/") {
		return Ref{}, fmt.Errorf("暂不支持该镜像类型: %s", input)
	}

	return Ref{Original: trimmed, Type: TypeDockerHubOfficial}, nil
}
