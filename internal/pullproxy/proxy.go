package pullproxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DefaultConfigDir  = "/etc/systemd/system/docker.service.d"
	DefaultConfigFile = "dmp-http-proxy.conf"
	DefaultNoProxy    = "localhost,127.0.0.1,::1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,*.local"
	RegistryCheckURL  = "https://registry-1.docker.io/v2/"
)

var (
	ErrUnsupportedEnv = errors.New("当前环境暂不支持 dmp pull proxy，仅支持 Linux systemd amd64/arm64")
	ErrSocksProxy     = errors.New("Docker daemon 代理配置不支持 socks5://，请使用 HTTP 代理地址")
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

type RuntimeInfo struct {
	GOOS   string
	GOARCH string
}

type Service struct {
	ConfigDir string
	NoProxy   string
	Out       io.Writer
	Runner    Runner
	Runtime   RuntimeInfo
	StatPath  string
}

func New(out io.Writer) *Service {
	return &Service{
		ConfigDir: DefaultConfigDir,
		NoProxy:   DefaultNoProxy,
		Out:       out,
		Runner:    execRunner{},
		Runtime: RuntimeInfo{
			GOOS:   runtime.GOOS,
			GOARCH: runtime.GOARCH,
		},
		StatPath: "/run/systemd/system",
	}
}

func (s *Service) Set(ctx context.Context, rawProxy string) error {
	if err := s.ensureSupported(); err != nil {
		s.log("%v", err)
		return err
	}

	proxy, warning, err := NormalizeProxy(rawProxy)
	if err != nil {
		return err
	}
	if warning != "" {
		s.log("%s", warning)
	}

	s.log("代理地址: %s", proxy)
	s.blankLine()
	s.log("将写入 Docker daemon 代理配置: %s", s.configPath())
	if err := os.MkdirAll(s.ConfigDir, 0o755); err != nil {
		s.log("创建配置目录失败: %v", err)
		s.log("请使用管理员权限重新执行: sudo dmp pull proxy set %s", rawProxy)
		return err
	}

	content := RenderConfig(proxy, s.noProxy())
	if err := os.WriteFile(s.configPath(), []byte(content), 0o644); err != nil {
		s.log("写入代理配置失败: %v", err)
		s.log("请使用管理员权限重新执行: sudo dmp pull proxy set %s", rawProxy)
		return err
	}
	s.log("代理配置写入完成")
	s.blankLine()

	s.log("将重新加载 systemd 并重启 Docker 服务，正在运行的容器可能短暂受影响")
	if err := s.reloadAndRestart(ctx); err != nil {
		return err
	}
	s.log("完成")
	return nil
}

func (s *Service) Remove(ctx context.Context) error {
	if err := s.ensureSupported(); err != nil {
		s.log("%v", err)
		return err
	}

	path := s.configPath()
	s.log("检查 dmp 代理配置: %s", path)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.log("未发现 dmp 代理配置，无需移除")
			s.log("完成")
			return nil
		}
		return err
	}

	s.log("移除 dmp 代理配置")
	if err := os.Remove(path); err != nil {
		s.log("移除代理配置失败: %v", err)
		s.log("请使用管理员权限重新执行: sudo dmp pull proxy rm")
		return err
	}
	s.blankLine()

	s.log("将重新加载 systemd 并重启 Docker 服务，正在运行的容器可能短暂受影响")
	if err := s.reloadAndRestart(ctx); err != nil {
		return err
	}
	s.log("完成")
	return nil
}

func (s *Service) Status(ctx context.Context) error {
	if err := s.ensureSupported(); err != nil {
		s.log("%v", err)
		return err
	}

	path := s.configPath()
	s.log("检查 dmp 代理配置: %s", path)
	proxy, err := s.readProxy()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.log("未发现 dmp 代理配置")
		} else {
			s.log("读取代理配置失败: %v", err)
			return err
		}
	} else {
		s.log("发现 dmp 代理配置: %s", proxy)
	}
	s.blankLine()

	showCmd := "systemctl show --property=Environment docker"
	s.log("执行验证命令: %s", showCmd)
	showOut, showErr := s.runner().Run(ctx, "systemctl", "show", "--property=Environment", "docker")
	for _, line := range formatDockerEnvironment(showOut) {
		s.log("%s", line)
	}
	if showErr != nil {
		s.log("systemctl 验证失败: %v", showErr)
		return showErr
	}
	if proxy != "" && strings.Contains(showOut, proxy) {
		s.log("Docker 服务已加载 dmp 代理地址")
	} else if proxy != "" {
		s.log("Docker 服务尚未加载 dmp 代理地址，请确认是否已执行 daemon-reload 和 restart")
	}

	if proxy == "" {
		s.log("未配置代理地址，跳过 curl -x 验证")
		s.log("完成")
		return nil
	}

	s.blankLine()
	curlArgs := []string{"-sS", "-x", proxy, RegistryCheckURL, "-I"}
	s.log("执行验证命令: curl %s", strings.Join(curlArgs, " "))
	_, curlErr := s.runner().Run(ctx, "curl", curlArgs...)
	if curlErr != nil {
		s.log("代理连通性验证结果: fail")
		s.log("失败原因: %v", curlErr)
		return curlErr
	}
	s.log("代理连通性验证结果: success")
	s.log("完成")
	return nil
}

func NormalizeProxy(raw string) (proxy string, warning string, err error) {
	proxy = strings.TrimSpace(raw)
	if proxy == "" {
		return "", "", fmt.Errorf("代理地址不能为空")
	}
	lower := strings.ToLower(proxy)
	if strings.HasPrefix(lower, "socks5://") || strings.HasPrefix(lower, "socks5h://") {
		return "", "", ErrSocksProxy
	}
	if strings.Contains(proxy, "://") {
		if strings.HasPrefix(lower, "https://") {
			return proxy, "提示: Docker daemon 代理常用 HTTP 代理地址，请确认该 HTTPS 代理可用", nil
		}
		if strings.HasPrefix(lower, "http://") {
			return proxy, "", nil
		}
		return "", "", fmt.Errorf("不支持的代理协议: %s", proxy)
	}
	return "http://" + proxy, "", nil
}

func RenderConfig(proxy, noProxy string) string {
	return fmt.Sprintf(`[Service]
Environment="HTTP_PROXY=%s"
Environment="HTTPS_PROXY=%s"
Environment="NO_PROXY=%s"
`, proxy, proxy, noProxy)
}

func (s *Service) reloadAndRestart(ctx context.Context) error {
	if err := s.runStep(ctx, "重新加载 systemd", "systemctl", "daemon-reload"); err != nil {
		s.log("请手动执行: sudo systemctl daemon-reload")
		return err
	}
	s.blankLine()
	if err := s.runStep(ctx, "重启 Docker 服务", "systemctl", "restart", "docker"); err != nil {
		s.log("请手动执行: sudo systemctl restart docker")
		return err
	}
	return nil
}

func (s *Service) runStep(ctx context.Context, label, name string, args ...string) error {
	s.log("%s", label)
	out, err := s.runner().Run(ctx, name, args...)
	if strings.TrimSpace(out) != "" {
		for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
			s.log("%s", line)
		}
	}
	if err != nil {
		s.log("%s失败: %v", label, err)
	}
	return err
}

func (s *Service) ensureSupported() error {
	rt := s.runtimeInfo()
	if rt.GOOS != "linux" {
		return ErrUnsupportedEnv
	}
	if rt.GOARCH != "amd64" && rt.GOARCH != "arm64" {
		return ErrUnsupportedEnv
	}
	if _, err := os.Stat(s.systemdPath()); err != nil {
		return ErrUnsupportedEnv
	}
	return nil
}

func (s *Service) readProxy() (string, error) {
	content, err := os.ReadFile(s.configPath())
	if err != nil {
		return "", err
	}
	return ParseProxyConfig(string(content)), nil
}

func ParseProxyConfig(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		const prefix = `Environment="HTTP_PROXY=`
		if strings.HasPrefix(line, prefix) && strings.HasSuffix(line, `"`) {
			return strings.TrimSuffix(strings.TrimPrefix(line, prefix), `"`)
		}
	}
	return ""
}

func formatDockerEnvironment(output string) []string {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}
	output = strings.TrimPrefix(output, "Environment=")
	fields := splitEnvironmentFields(output)
	if len(fields) == 0 {
		return []string{"Environment="}
	}
	lines := make([]string, 0, len(fields)+1)
	lines = append(lines, "Environment:")
	for _, field := range fields {
		lines = append(lines, "  "+strings.Trim(field, `"`))
	}
	return lines
}

func splitEnvironmentFields(value string) []string {
	var fields []string
	var b strings.Builder
	inQuote := false
	for _, r := range value {
		switch r {
		case '"':
			inQuote = !inQuote
			b.WriteRune(r)
		case ' ', '\n', '\t':
			if inQuote {
				b.WriteRune(r)
				continue
			}
			if strings.TrimSpace(b.String()) != "" {
				fields = append(fields, strings.TrimSpace(b.String()))
				b.Reset()
			}
		default:
			b.WriteRune(r)
		}
	}
	if strings.TrimSpace(b.String()) != "" {
		fields = append(fields, strings.TrimSpace(b.String()))
	}
	return fields
}

func (s *Service) configPath() string {
	return filepath.Join(s.ConfigDir, DefaultConfigFile)
}

func (s *Service) systemdPath() string {
	if s.StatPath == "" {
		return "/run/systemd/system"
	}
	return s.StatPath
}

func (s *Service) noProxy() string {
	if s.NoProxy == "" {
		return DefaultNoProxy
	}
	return s.NoProxy
}

func (s *Service) runner() Runner {
	if s.Runner == nil {
		return execRunner{}
	}
	return s.Runner
}

func (s *Service) runtimeInfo() RuntimeInfo {
	if s.Runtime.GOOS == "" || s.Runtime.GOARCH == "" {
		return RuntimeInfo{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
	}
	return s.Runtime
}

func (s *Service) log(format string, args ...any) {
	if s.Out == nil {
		return
	}
	fmt.Fprintf(s.Out, "[dmp] "+format+"\n", args...)
}

func (s *Service) blankLine() {
	if s.Out == nil {
		return
	}
	fmt.Fprintln(s.Out)
}
