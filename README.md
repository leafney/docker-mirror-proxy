# docker-mirror-proxy

`dmp` 是一个多类型下载加速命令行工具。当前支持 Docker 镜像拉取加速和 GitHub 文件下载加速，后续预留 Python 包、Node 包加速命令。

## pull 命令 -- Docker 镜像拉取加速

### 支持镜像

- Docker Hub 官方镜像，例如 `nginx:latest`
- Docker Hub 自定义镜像，例如 `leafney/ai-signin:0.6.11`
- GHCR 镜像，例如 `ghcr.io/leafney/ai-signin:0.6.8`

`pull` 命令会按顺序尝试内置加速地址池中的地址。某个地址失败或超时后，会自动切换下一个地址。拉取成功后，会把镜像重新打回原始名称，并删除临时加速镜像标签。

这里的 `--timeout` 表示“无进展超时”，不是整个 `docker pull` 的总耗时。只要 Docker 仍有 stdout 或 stderr 输出，程序就会继续等待。需要限制单个加速地址总耗时时，可以配合 `--max-time` 使用。

### 支持参数

- `-t, --timeout int` 单个加速地址的无进展超时时间，单位为秒，默认 60
- `-m, --max-time int` 单个加速地址的总耗时上限，单位为秒，默认 0，表示不限制

### 使用方式

```bash
# dmp pull <DockerHub 镜像名称>
dmp pull nginx:latest

# dmp pull <GHCR 镜像名称>
dmp pull ghcr.io/leafney/ai-signin:0.6.8

# dmp pull -t/--timeout <超时时间> <镜像名称> （自定义无进展超时时间）
dmp pull --timeout 30 nginx:latest

# dmp pull -t/--timeout <无进展超时时间> -m/--max-time <总耗时上限> <镜像名称>
dmp pull --timeout 60 --max-time 1800 rabbitmq:4.3.1-management
```

### 操作示例

```bash
➜ dmp pull ghcr.io/leafney/ai-signin:0.6.11
[dmp] 原始镜像: ghcr.io/leafney/ai-signin:0.6.11
[dmp] 镜像类型: ghcr
[dmp] 无进展超时时间: 60 秒
[dmp] 单个加速地址总耗时上限: 不限制
[dmp] 候选加速地址数量: 4
[dmp] 尝试 1/4: docker.1ms.run/ghcr.io/leafney/ai-signin:0.6.11
Error response from daemon: manifest for docker.1ms.run/ghcr.io/leafney/ai-signin:0.6.11 not found: manifest unknown: 没有找到该资源，请检查镜像名称或者版本(tag)是否真实存在！(例如：拼写错误? 没有指定版本[tag]? 版本错误? AI胡诌?) 欢迎联系我们获得帮助 QQ群：1102523830
[dmp] 拉取失败，切换下一个加速地址
[dmp] 尝试 2/4: dockerproxy.net/ghcr.io/leafney/ai-signin:0.6.11
Error response from daemon: manifest for dockerproxy.net/ghcr.io/leafney/ai-signin:0.6.11 not found: manifest unknown: manifest unknown
[dmp] 拉取失败，切换下一个加速地址
[dmp] 尝试 3/4: proxy.vvvv.ee/ghcr.io/leafney/ai-signin:0.6.11
0.6.11: Pulling from ghcr.io/leafney/ai-signin
9b02e9fcb401: Pull complete 
d5babde5a207: Pull complete 
cf70f8595b3a: Pull complete 
bed64e9fefb7: Pull complete 
80d6ec238cdb: Pull complete 
Digest: sha256:d36187befddf7a8e0f01ce8da06d453c7f534bea713960e1a075a0cbfea02fa4
Status: Downloaded newer image for proxy.vvvv.ee/ghcr.io/leafney/ai-signin:0.6.11
proxy.vvvv.ee/ghcr.io/leafney/ai-signin:0.6.11
[dmp] 拉取成功: proxy.vvvv.ee/ghcr.io/leafney/ai-signin:0.6.11
[dmp] 打标签: proxy.vvvv.ee/ghcr.io/leafney/ai-signin:0.6.11 -> ghcr.io/leafney/ai-signin:0.6.11
[dmp] 清理临时标签: proxy.vvvv.ee/ghcr.io/leafney/ai-signin:0.6.11
Untagged: proxy.vvvv.ee/ghcr.io/leafney/ai-signin:0.6.11
Untagged: proxy.vvvv.ee/ghcr.io/leafney/ai-signin@sha256:d36187befddf7a8e0f01ce8da06d453c7f534bea713960e1a075a0cbfea02fa4
[dmp] 完成: ghcr.io/leafney/ai-signin:0.6.11
```

---

## gh 命令 -- GitHub 文件下载加速

### 支持参数

- `-t, --timeout int` 单个加速地址的超时时间，单位为秒，默认 60
- `-o, --output string` 下载目录，默认当前目录
- `-p, --proxy string` 代理地址，例如 `http://127.0.0.1:7890`，说明：指定代理后，程序会优先通过代理访问原始 GitHub 地址；代理请求失败或超时后，再回退到内置加速地址池。

### 使用方式

```bash
# dmp gh <文件 URL> （将文件下载到当前目录）
dmp gh https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz

# dmp gh -t/--timeout <超时时间> -o/--output <下载目录> <文件 URL> （将文件下载到指定目录下并自定义下载超时时间）
dmp gh -t 30 -o /tmp https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz

# dmp gh -p/--proxy <代理地址> <文件 URL> （优先使用指定代理下载）
dmp gh -p http://127.0.0.1:7890 https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
```

---

## 帮助

### 查看命令帮助

```bash
dmp
dmp --help
```

### 查看版本信息

```bash
dmp -v
dmp --version
```

---

## 命令规划

```text
dmp pull    Docker 镜像拉取加速
dmp gh      GitHub 文件下载加速
dmp pip     Python 包加速，暂未实现
dmp npm     Node 包加速，暂未实现
```
