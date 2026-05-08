# docker-mirror-proxy

`dmp` 是一个多类型下载加速命令行工具。当前支持 Docker 镜像拉取加速和 GitHub 文件下载加速，后续预留 Python 包、Node 包加速命令。

第一版只支持：

- Docker Hub 官方镜像，例如 `nginx:latest`
- GHCR 镜像，例如 `ghcr.io/leafney/ai-signin:0.6.8`

## 使用方式

```bash
dmp pull nginx:latest
dmp pull ghcr.io/leafney/ai-signin:0.6.8
dmp pull --timeout 30 nginx:latest
dmp pull -t 30 -p http://127.0.0.1:7890 nginx:latest
dmp gh https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
dmp gh -t 30 -o /tmp https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
dmp gh -p http://127.0.0.1:7890 https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
```

不带参数时会显示帮助：

```bash
dmp
dmp --help
```

第一版不支持以下快捷形式：

```bash
dmp nginx:latest
```

## 命令规划

```text
dmp pull    Docker 镜像拉取加速
dmp gh      GitHub 文件下载加速
dmp pip     Python 包加速，暂未实现
dmp npm     Node 包加速，暂未实现
dmp version 显示版本信息
```

## 工作方式

程序内置以下第三方镜像加速地址：

```text
docker.1ms.run
dockerproxy.net
proxy.vvvv.ee
registry.cyou
```

它们对 Docker Hub 官方镜像和 GHCR 镜像都使用统一前缀规则。

例如：

```text
nginx:latest
=> docker.1ms.run/nginx:latest
```

```text
ghcr.io/leafney/ai-signin:0.6.8
=> docker.1ms.run/ghcr.io/leafney/ai-signin:0.6.8
```

`dmp` 会按顺序尝试这些地址。某个地址失败或超时后，会自动切换下一个地址。拉取成功后，会把镜像重新打回原始名称，并删除临时加速镜像标签。

GitHub 文件下载会自动检测当前系统中的 `curl` 或 `wget`，优先使用 `curl`。下载目录必须存在，程序会输出下载目录和最终文件位置的绝对路径。

`pull` 和 `gh` 都支持通过 `-p, --proxy` 指定代理地址。指定代理后，程序会优先通过代理访问原始镜像或原始 GitHub 地址；代理请求失败或超时后，再回退到内置加速地址池。

## 参数

`pull` 命令支持：

```text
-t, --timeout int
    单个加速地址的超时时间，单位为秒，默认 60

-p, --proxy string
    代理地址，例如 http://127.0.0.1:7890
```

`gh` 命令支持：

```text
-t, --timeout int
    单个加速地址的超时时间，单位为秒，默认 60

-o, --output string
    下载目录，默认当前目录

-p, --proxy string
    代理地址，例如 http://127.0.0.1:7890
```

根命令支持：

```text
--help
    显示帮助信息
```

## 本地构建

```bash
make test
make build
make version
```

或直接执行：

```bash
go test ./...
go build -ldflags="-X 'main.Version=v0.1.0'" -o dmp ./cmd/dmp
```

`make build` 会自动注入版本信息，其中 `Build Time` 使用东八区时间：

```text
Version
Git Branch
Git Commit
Build Time
```

查看版本：

```bash
dmp version
```

## 发布构建

GitHub Actions 会在推送 `v*` 标签或手动触发时构建：

- `dmp-linux-amd64.tar.gz`
- `dmp-linux-arm64.tar.gz`
- `dmp-darwin-arm64.tar.gz`
- `checksums.txt`
