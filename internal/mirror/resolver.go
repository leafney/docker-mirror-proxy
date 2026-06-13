package mirror

import (
	"strings"

	"github.com/leafney/docker-mirror-proxy/internal/image"
)

var DefaultRegistries = []string{
	"docker.1ms.run",
	"dockerproxy.net",
}

func Candidates(ref image.Ref) []string {
	candidates := make([]string, 0, len(DefaultRegistries)+1)
	for _, registry := range DefaultRegistries {
		if registry == "docker.1ms.run" && ref.Type == image.TypeGHCR {
			path := strings.TrimPrefix(ref.Original, "ghcr.io/")
			candidates = append(candidates, "ghcr.1ms.run/"+path)
		} else {
			candidates = append(candidates, registry+"/"+ref.Original)
		}
	}
	// 兜底：原始官方地址
	if ref.Type == image.TypeGHCR {
		candidates = append(candidates, ref.Original)
	} else {
		candidates = append(candidates, "docker.io/"+ref.Original)
	}
	return candidates
}
