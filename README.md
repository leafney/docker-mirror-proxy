# docker-mirror-pull

`dmp` 是一个 Docker 镜像下载加速命令行工具。

第一版只支持：

- Docker Hub 官方镜像，例如 `nginx:latest`
- GHCR 镜像，例如 `ghcr.io/leafney/ai-signin:0.6.8`

## 使用方式

```bash
dmp nginx:latest
dmp ghcr.io/leafney/ai-signin:0.6.8
dmp --timeout 30 nginx:latest
```

不带参数时会显示帮助：

```bash
dmp
dmp --help
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

## 参数

```text
--timeout int
    单个加速地址的超时时间，单位为秒，默认 60

--no-clean
    成功后不删除临时加速镜像标签

--version
    显示版本信息

--help
    显示帮助信息
```

## 本地构建

```bash
make test
make build
```

或直接执行：

```bash
go test ./...
go build -o dmp ./cmd/dmp
```

## 发布构建

GitHub Actions 会在推送 `v*` 标签或手动触发时构建：

- `dmp-linux-amd64`
- `dmp-linux-arm64`
- `dmp-darwin-arm64`
- `checksums.txt`
