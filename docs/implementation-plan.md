# dmp 第一版实现规划

## 目标

`dmp` 是一个 Docker 镜像下载加速命令行工具。第一版编译后只提供一个二进制文件，不读取配置文件，用户通过类似 `dmp nginx:latest` 或 `dmp ghcr.io/leafney/ai-signin:0.6.8` 的形式拉取镜像。

第一版只支持两类镜像：

- Docker Hub 官方镜像，例如 `nginx:latest`、`redis:7`
- GHCR 镜像，例如 `ghcr.io/leafney/ai-signin:0.6.8`

其他 registry 或 Docker Hub 用户镜像暂不支持。

## 核心规则

程序内置 `docs/first.md` 中提供的第三方镜像地址：

- `docker.1ms.run`
- `dockerproxy.net`
- `proxy.vvvv.ee`
- `registry.cyou`

这些地址对 Docker Hub 官方镜像和 GHCR 镜像都使用同一套前缀代理规则。

示例：

```text
nginx:latest
=> docker.1ms.run/nginx:latest
=> dockerproxy.net/nginx:latest
=> proxy.vvvv.ee/nginx:latest
=> registry.cyou/nginx:latest
```

```text
ghcr.io/leafney/ai-signin:0.6.8
=> docker.1ms.run/ghcr.io/leafney/ai-signin:0.6.8
=> dockerproxy.net/ghcr.io/leafney/ai-signin:0.6.8
=> proxy.vvvv.ee/ghcr.io/leafney/ai-signin:0.6.8
=> registry.cyou/ghcr.io/leafney/ai-signin:0.6.8
```

## CLI 行为

支持命令：

```bash
dmp
dmp --help
dmp --version
dmp --timeout 60 nginx:latest
dmp nginx:latest
dmp ghcr.io/leafney/ai-signin:0.6.8
```

规则：

- `dmp` 不带参数时等同于 `dmp --help`
- 默认单个镜像地址超时时间为 `60` 秒
- `--timeout` 支持临时设置超时时间，单位为秒
- 支持一次传入多个镜像，按顺序逐个处理
- 某个镜像处理失败后继续处理后续镜像，最后返回非零退出码

## 下载流程

单个镜像的处理流程：

```text
解析原始镜像
  ↓
生成候选代理镜像地址
  ↓
按顺序尝试 docker pull
  ↓
失败或超时后切换下一个地址
  ↓
成功后 docker tag 回原始镜像名
  ↓
docker rmi 删除临时代理标签
  ↓
输出完成日志
```

如果所有候选地址都失败，输出已尝试的地址列表，并返回错误。

## 日志设计

运行过程中输出必要日志，便于用户观察状态：

```text
[dmp] 原始镜像: nginx:latest
[dmp] 镜像类型: dockerhub-official
[dmp] 超时时间: 60 秒
[dmp] 候选加速地址数量: 4
[dmp] 尝试 1/4: docker.1ms.run/nginx:latest
[dmp] 拉取失败，切换下一个加速地址: docker.1ms.run/nginx:latest
[dmp] 尝试 2/4: dockerproxy.net/nginx:latest
[dmp] 拉取成功: dockerproxy.net/nginx:latest
[dmp] 打标签: dockerproxy.net/nginx:latest -> nginx:latest
[dmp] 清理临时标签: dockerproxy.net/nginx:latest
[dmp] 完成: nginx:latest
```

Docker 自身的输出直接透传到终端，方便用户看到真实拉取进度和错误原因。

## 项目结构

```text
cmd/
  dmp/
    main.go

internal/
  app/
    app.go
    app_test.go

  docker/
    client.go

  image/
    parser.go
    parser_test.go

  logx/
    logger.go

  mirror/
    resolver.go
    resolver_test.go

.github/
  workflows/
    release.yml

Makefile
go.mod
README.md
```

模块职责：

- `cmd/dmp`：程序入口，解析 CLI 参数，设置默认值和退出码。
- `internal/image`：解析和校验镜像名称，只放行业务允许的镜像类型。
- `internal/mirror`：维护内置镜像地址池，生成候选代理镜像。
- `internal/docker`：封装 `docker pull`、`docker tag`、`docker rmi`。
- `internal/app`：串联完整业务流程，包括失败切换、超时、日志和清理。
- `internal/logx`：统一日志输出格式。

## 测试规划

单元测试重点覆盖：

- `nginx` 被识别为 Docker Hub 官方镜像
- `nginx:latest` 被识别为 Docker Hub 官方镜像
- `ghcr.io/leafney/ai-signin:0.6.8` 被识别为 GHCR 镜像
- `quay.io/org/image:tag` 被拒绝
- `registry.k8s.io/pause:3.8` 被拒绝
- `user/image:tag` 第一版被拒绝
- Docker Hub 官方镜像生成 4 个候选代理地址
- GHCR 镜像生成 4 个候选代理地址
- 第一个代理地址失败时继续尝试下一个
- 拉取成功后执行 `docker tag`
- 默认执行 `docker rmi` 清理临时代理标签

## GitHub Actions 打包

新增 `.github/workflows/release.yml`。

触发方式：

- 推送 `v*` tag
- 手动 `workflow_dispatch`

构建产物：

- `dmp-linux-amd64`
- `dmp-linux-arm64`
- `dmp-darwin-arm64`
- `checksums.txt`

交叉编译命令：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/dmp-linux-amd64 ./cmd/dmp
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/dmp-linux-arm64 ./cmd/dmp
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/dmp-darwin-arm64 ./cmd/dmp
```

## 第一版验收标准

本地验收：

```bash
go test ./...
go build -o dmp ./cmd/dmp
./dmp
./dmp --help
./dmp --timeout 60 nginx:latest
./dmp ghcr.io/leafney/ai-signin:0.6.8
```

功能验收：

- 无参数显示帮助
- 默认超时时间为 60 秒
- 支持通过 `--timeout` 临时设置超时时间
- 只支持 Docker Hub 官方镜像和 GHCR 镜像
- 使用统一内置第三方镜像地址池
- 某个代理地址失败或超时后自动尝试下一个
- 成功后打回原始镜像标签
- 默认清理代理临时标签
- 运行过程中有清晰日志
- GitHub Actions 能构建 Linux amd64、Linux arm64、macOS arm64 三个二进制文件
