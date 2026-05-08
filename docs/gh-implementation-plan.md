# dmp gh 功能实现规划

## 目标

`dmp gh` 用于加速下载 GitHub 文件，支持 GitHub 文件、Releases、Archive、Gist、`raw.githubusercontent.com` 文件下载，不支持整个项目文件夹下载。

命令示例：

```bash
dmp gh https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
dmp gh -t 30 https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
dmp gh --timeout 30 https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
dmp gh -o /tmp https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
dmp gh --output /tmp https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
dmp gh -p http://127.0.0.1:7890 https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
```

## 参数设计

- `-t, --timeout int`：单个加速地址下载超时时间，单位为秒，默认 `60`。
- `-o, --output string`：下载目录，默认当前目录。
- `-p, --proxy string`：代理地址，例如 `http://127.0.0.1:7890`。

不提供下载工具选择参数。程序内部自动检测当前系统环境：

1. 优先使用 `curl`
2. 如果没有 `curl`，再使用 `wget`
3. 如果二者都不存在，停止执行并提示用户先安装其中一个工具

## 下载目录和文件名

- 下载目录必须存在，不存在时直接报错，不自动创建。
- 下载目录需要转换为绝对路径后输出，效果类似 `pwd`。
- 下载文件名保持原始 URL 中的文件名，不自定义、不改名。
- 下载完成后输出文件所在的绝对路径，例如 `/tmp/test.zip`。

## 加速地址

内置 `docs/github.md` 中的代理地址，按顺序轮换：

```text
https://ghfast.top
https://gh-proxy.com
https://gh-proxy.org
https://fastgit.cc
https://ghproxy.net
https://ghproxylist.com
https://hk.gh-proxy.org
https://cdn.gh-proxy.org
https://ghp.keleyaa.com
https://gh.jasonzeng.dev
```

拼接规则：

```text
代理地址 + "/" + 原始 GitHub URL
```

示例：

```text
https://ghproxy.net/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
```

## 执行流程

单个 URL 下载流程：

```text
校验原始 URL
  ↓
解析原始文件名
  ↓
解析下载目录为绝对路径
  ↓
检查下载目录是否存在
  ↓
自动检测 curl 或 wget
  ↓
如果指定代理，优先通过代理下载原始 URL
  ↓
生成候选加速 URL
  ↓
按顺序尝试下载
  ↓
失败或超时后切换下一个加速地址
  ↓
成功后输出文件绝对路径
```

多个 URL 下载流程：

```text
按输入顺序逐个下载
某个 URL 失败后继续处理后续 URL
最后如果存在失败项，整体返回非零退出码
```

## 日志设计

```text
[dmp] 原始地址: https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
[dmp] 下载目录: /tmp
[dmp] 下载工具: curl
[dmp] 超时时间: 60 秒
[dmp] 代理地址: http://127.0.0.1:7890
[dmp] 候选加速地址数量: 10
[dmp] 优先尝试代理下载: https://github.com/...
[dmp] 尝试 1/10: https://ghfast.top/https://github.com/...
[dmp] 下载失败，切换下一个加速地址
[dmp] 尝试 2/10: https://gh-proxy.com/https://github.com/...
[dmp] 下载成功
[dmp] 文件位置: /tmp/dmp-linux-amd64.tar.gz
```

## 模块设计

GitHub 相关模块统一使用 `gh` 前缀：

```text
internal/
  ghurl/
    parser.go
    parser_test.go

  ghmirror/
    resolver.go
    resolver_test.go

  ghdownload/
    client.go
    client_test.go

  ghapp/
    app.go
    app_test.go
```

模块职责：

- `internal/ghurl`：校验 GitHub URL，解析原始文件名。
- `internal/ghmirror`：维护 GitHub 加速地址池，生成候选加速 URL。
- `internal/ghdownload`：检测 `curl` / `wget`，封装系统下载命令。
- `internal/ghapp`：串联完整下载流程，包括超时、失败轮换、日志和文件路径输出。
- `internal/cli`：把 `gh` 从预留命令改为真实命令，绑定 `-t/--timeout`、`-o/--output` 和 `-p/--proxy` 参数。

## 测试规划

- 支持 `https://github.com/.../releases/download/...`
- 支持 `https://raw.githubusercontent.com/...`
- 支持 `https://gist.githubusercontent.com/...`
- 拒绝非 GitHub 地址
- 拒绝没有明确文件名的 URL
- 生成 10 个候选加速 URL
- 自动优先选择 `curl`
- 没有 `curl` 时选择 `wget`
- 两者都不存在时报错
- 下载目录不存在时报错
- 下载目录输出为绝对路径
- 第一个加速地址失败后继续尝试第二个
- 下载成功后输出最终文件绝对路径
- `dmp gh -t` 和 `dmp gh --timeout` 均可用
- `dmp gh -o` 和 `dmp gh --output` 均可用
- `dmp gh -p` 和 `dmp gh --proxy` 均可用
