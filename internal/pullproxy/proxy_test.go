package pullproxy

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type runnerCall struct {
	Name string
	Args []string
}

type fakeRunner struct {
	calls   []runnerCall
	outputs map[string]string
	errors  map[string]error
}

func (f *fakeRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	call := runnerCall{Name: name, Args: append([]string(nil), args...)}
	f.calls = append(f.calls, call)
	key := commandKey(name, args...)
	return f.outputs[key], f.errors[key]
}

func commandKey(name string, args ...string) string {
	return name + " " + strings.Join(args, " ")
}

func newTestService(t *testing.T, runner *fakeRunner) (*Service, *bytes.Buffer) {
	t.Helper()
	tmp := t.TempDir()
	systemd := filepath.Join(tmp, "systemd")
	if err := os.MkdirAll(systemd, 0o755); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	return &Service{
		ConfigDir: filepath.Join(tmp, "docker.service.d"),
		NoProxy:   DefaultNoProxy,
		Out:       &out,
		Runner:    runner,
		Runtime:   RuntimeInfo{GOOS: "linux", GOARCH: "amd64"},
		StatPath:  systemd,
	}, &out
}

func TestNormalizeProxy(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantProxy   string
		wantWarning string
		wantErr     error
	}{
		{name: "without scheme", input: "192.168.8.100:7890", wantProxy: "http://192.168.8.100:7890"},
		{name: "http", input: "http://127.0.0.1:7890", wantProxy: "http://127.0.0.1:7890"},
		{name: "https", input: "https://proxy:7890", wantProxy: "https://proxy:7890", wantWarning: "提示"},
		{name: "socks5", input: "socks5://127.0.0.1:7890", wantErr: ErrSocksProxy},
		{name: "empty", input: " ", wantErr: errors.New("代理地址不能为空")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy, warning, err := NormalizeProxy(tt.input)
			if tt.wantErr != nil {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeProxy returned error: %v", err)
			}
			if proxy != tt.wantProxy {
				t.Fatalf("proxy = %q, want %q", proxy, tt.wantProxy)
			}
			if tt.wantWarning != "" && !strings.Contains(warning, tt.wantWarning) {
				t.Fatalf("warning = %q, want contains %q", warning, tt.wantWarning)
			}
		})
	}
}

func TestRenderConfigIncludesProxyVariables(t *testing.T) {
	content := RenderConfig("http://192.168.8.100:7890", "localhost")
	for _, want := range []string{
		`Environment="HTTP_PROXY=http://192.168.8.100:7890"`,
		`Environment="HTTPS_PROXY=http://192.168.8.100:7890"`,
		`Environment="NO_PROXY=localhost"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("config missing %q: %s", want, content)
		}
	}
}

func TestSetWritesConfigAndRestartsDocker(t *testing.T) {
	runner := &fakeRunner{outputs: map[string]string{}, errors: map[string]error{}}
	svc, out := newTestService(t, runner)

	if err := svc.Set(context.Background(), "192.168.8.100:7890"); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(svc.ConfigDir, DefaultConfigFile))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(content), `HTTP_PROXY=http://192.168.8.100:7890`) {
		t.Fatalf("config missing proxy: %s", string(content))
	}
	wantCalls := []runnerCall{
		{Name: "systemctl", Args: []string{"daemon-reload"}},
		{Name: "systemctl", Args: []string{"restart", "docker"}},
	}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
	if !strings.Contains(out.String(), "将重新加载 systemd 并重启 Docker 服务") {
		t.Fatalf("output missing restart warning: %s", out.String())
	}
	for _, want := range []string{
		"代理地址: http://192.168.8.100:7890\n\n[dmp] 将写入 Docker daemon 代理配置",
		"代理配置写入完成\n\n[dmp] 将重新加载 systemd",
		"重新加载 systemd\n\n[dmp] 重启 Docker 服务",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output missing blank line %q: %s", want, out.String())
		}
	}
}

func TestRemoveDoesNothingWhenConfigMissing(t *testing.T) {
	runner := &fakeRunner{outputs: map[string]string{}, errors: map[string]error{}}
	svc, out := newTestService(t, runner)

	if err := svc.Remove(context.Background()); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("calls = %#v, want none", runner.calls)
	}
	if !strings.Contains(out.String(), "无需移除") {
		t.Fatalf("output missing no-op message: %s", out.String())
	}
}

func TestRemoveDeletesConfigAndRestartsDocker(t *testing.T) {
	runner := &fakeRunner{outputs: map[string]string{}, errors: map[string]error{}}
	svc, _ := newTestService(t, runner)
	if err := os.MkdirAll(svc.ConfigDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(svc.ConfigDir, DefaultConfigFile)
	if err := os.WriteFile(path, []byte("config"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := svc.Remove(context.Background()); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config stat err = %v, want not exist", err)
	}
	wantCalls := []runnerCall{
		{Name: "systemctl", Args: []string{"daemon-reload"}},
		{Name: "systemctl", Args: []string{"restart", "docker"}},
	}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
}

func TestStatusRunsSystemctlAndCurl(t *testing.T) {
	proxy := "http://192.168.8.100:7890"
	runner := &fakeRunner{
		outputs: map[string]string{
			commandKey("systemctl", "show", "--property=Environment", "docker"): "Environment=HTTP_PROXY=" + proxy + " HTTPS_PROXY=" + proxy + "\n",
			commandKey("curl", "-x", proxy, RegistryCheckURL, "-I"):             "HTTP/2 401\n",
		},
		errors: map[string]error{},
	}
	svc, out := newTestService(t, runner)
	if err := os.MkdirAll(svc.ConfigDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(svc.ConfigDir, DefaultConfigFile), []byte(RenderConfig(proxy, DefaultNoProxy)), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := svc.Status(context.Background()); err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	wantCalls := []runnerCall{
		{Name: "systemctl", Args: []string{"show", "--property=Environment", "docker"}},
		{Name: "curl", Args: []string{"-x", proxy, RegistryCheckURL, "-I"}},
	}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, wantCalls)
	}
	if !strings.Contains(out.String(), "代理连通性验证成功") {
		t.Fatalf("output missing success: %s", out.String())
	}
}

func TestUnsupportedEnvironment(t *testing.T) {
	runner := &fakeRunner{outputs: map[string]string{}, errors: map[string]error{}}
	svc, _ := newTestService(t, runner)
	svc.Runtime = RuntimeInfo{GOOS: "darwin", GOARCH: "arm64"}

	if err := svc.Set(context.Background(), "127.0.0.1:7890"); !errors.Is(err, ErrUnsupportedEnv) {
		t.Fatalf("err = %v, want ErrUnsupportedEnv", err)
	}
}
