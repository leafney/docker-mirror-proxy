package mirror

import "github.com/leafney/docker-mirror-proxy/internal/image"

var DefaultRegistries = []string{
	"docker.1ms.run",
	"dockerproxy.net",
	"proxy.vvvv.ee",
	"registry.cyou",
}

func Candidates(ref image.Ref) []string {
	candidates := make([]string, 0, len(DefaultRegistries))
	for _, registry := range DefaultRegistries {
		candidates = append(candidates, registry+"/"+ref.Original)
	}
	return candidates
}
