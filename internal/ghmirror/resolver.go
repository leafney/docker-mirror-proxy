package ghmirror

import "github.com/leafney/docker-mirror-proxy/internal/ghurl"

var DefaultProxies = []string{
	"https://ghfast.top",
	"https://gh-proxy.com",
	"https://gh-proxy.org",
	"https://fastgit.cc",
	"https://ghproxy.net",
	"https://ghproxylist.com",
	"https://hk.gh-proxy.org",
	"https://cdn.gh-proxy.org",
	"https://ghp.keleyaa.com",
	"https://gh.jasonzeng.dev",
}

func Candidates(ref ghurl.Ref) []string {
	candidates := make([]string, 0, len(DefaultProxies))
	for _, proxy := range DefaultProxies {
		candidates = append(candidates, proxy+"/"+ref.Original)
	}
	return candidates
}
