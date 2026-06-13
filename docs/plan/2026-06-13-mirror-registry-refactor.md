# PRD：镜像加速地址重构

Related discussion: docs/discuss/2026-06-13-mirror-registry-refactor.md

---

## Problem Statement

当前镜像候选地址列表存在三个问题：

1. 包含已失效或不再维护的加速地址（`registry.cyou`、`proxy.vvvv.ee`），徒增无效尝试
2. `docker.1ms.run` 对 GHCR 镜像的拼接方式有误——该服务要求 GHCR 镜像使用独立域名 `ghcr.1ms.run`，而当前代码统一用 `docker.1ms.run/ghcr.io/...`，导致拉取必然失败
3. 所有加速地址均失败时，没有原始地址兜底，用户需要手动改写命令才能回退到官方源

## Solution

重构 `mirror.Candidates()` 函数：

- 移除失效加速地址
- 对 `docker.1ms.run` 按镜像类型分支拼接：DockerHub 保持原逻辑，GHCR 换用 `ghcr.1ms.run/<path>`
- 在候选列表末尾追加对应的原始官方地址作为兜底

## User Stories

1. 作为用户，我想拉取 DockerHub 官方镜像（如 `nginx:latest`）时，系统不再尝试 `registry.cyou` 和 `proxy.vvvv.ee`，以减少无效等待时间
2. 作为用户，我想拉取 DockerHub 命名空间镜像（如 `lexiforest/curl-impersonate:latest`）时，系统不再尝试已失效的加速地址
3. 作为用户，我想拉取 GHCR 镜像（如 `ghcr.io/leafney/ai-signin:0.6.8`）时，`docker.1ms.run` 加速能够正确生效，而不是因为格式错误而失败
4. 作为用户，我想拉取 GHCR 镜像时，系统使用 `ghcr.1ms.run/<owner>/<image>:<tag>` 格式，符合 1ms.run 服务的实际接口规范
5. 作为用户，我想在所有加速地址均不可用时，系统自动回退到原始官方地址（DockerHub → `docker.io/...`，GHCR → `ghcr.io/...`），而无需我手动修改命令
6. 作为用户，我想兜底的官方地址排在所有加速地址之后，确保优先使用加速，仅在最后才回退官方源

## Implementation Decisions

### 模块：`mirror.Candidates()`

这是本次变更的唯一修改点，其余模块（`image.Parse`、`app.go`、CLI）无需改动。

**变更内容：**

- `DefaultRegistries` 移除 `registry.cyou`、`proxy.vvvv.ee`，保留 `docker.1ms.run` 和 `dockerproxy.net`
- `Candidates()` 对加速地址循环时，对 `docker.1ms.run` 做类型分支：
  - `ref.Type == TypeGHCR`：拼接为 `ghcr.1ms.run/<path>`（去掉 `ghcr.io/` 前缀）
  - 其他类型：保持 `docker.1ms.run/<original>` 原逻辑
- `dockerproxy.net` 及其他加速地址继续使用统一旧逻辑：`registry + "/" + ref.Original`
- 循环结束后，追加兜底地址：
  - `TypeGHCR`：直接追加 `ref.Original`（已含 `ghcr.io/` 域名）
  - `TypeDockerHubOfficial`：追加 `docker.io/` + `ref.Original`

**候选地址生成逻辑（伪代码，来自讨论共识）：**

```
for each registry in DefaultRegistries:
  if registry == "docker.1ms.run" and ref.Type == TypeGHCR:
    append "ghcr.1ms.run/" + stripPrefix(ref.Original, "ghcr.io/")
  else:
    append registry + "/" + ref.Original

// 兜底
if ref.Type == TypeGHCR:
  append ref.Original              // ghcr.io/leafney/ai-signin:0.6.8
else:
  append "docker.io/" + ref.Original  // docker.io/nginx:latest
```

### 变更后的候选列表示例

**DockerHub 官方镜像** `nginx:latest`：
1. `docker.1ms.run/nginx:latest`
2. `dockerproxy.net/nginx:latest`
3. `docker.io/nginx:latest`（兜底）

**DockerHub 命名空间镜像** `lexiforest/curl-impersonate:latest`：
1. `docker.1ms.run/lexiforest/curl-impersonate:latest`
2. `dockerproxy.net/lexiforest/curl-impersonate:latest`
3. `docker.io/lexiforest/curl-impersonate:latest`（兜底）

**GHCR 镜像** `ghcr.io/leafney/ai-signin:0.6.8`：
1. `ghcr.1ms.run/leafney/ai-signin:0.6.8`
2. `dockerproxy.net/ghcr.io/leafney/ai-signin:0.6.8`
3. `ghcr.io/leafney/ai-signin:0.6.8`（兜底）

## Testing Decisions

**好测试的标准：** 只测 `Candidates()` 的输出（外部行为），不关心内部分支实现细节。

**测试模块：** `mirror.Candidates()`

**先例：** 现有 `resolver_test.go` 中三个测试用例，结构清晰，直接更新 `want` 列表即可，无需新增测试文件。

需更新的测试用例：
- `TestCandidatesForDockerHubOfficialImage`：移除两个失效地址，末尾加 `docker.io/nginx:latest`
- `TestCandidatesForDockerHubNamespaceImage`：同上，末尾加 `docker.io/lexiforest/curl-impersonate:latest`
- `TestCandidatesForGHCRImage`：移除两个失效地址，GHCR 条目改为 `ghcr.1ms.run/leafney/ai-signin:0.6.8`，末尾加 `ghcr.io/leafney/ai-signin:0.6.8`

## Out of Scope

- `dockerproxy.net` 的 GHCR 格式适配（保持旧逻辑）
- 其他镜像源（`gcr.io`、`quay.io`、`k8s.gcr.io` 等）的支持
- 加速地址的健康检查或动态排序
- `ghmirror`（GitHub Release 代理）相关逻辑，不涉及本次变更

## Further Notes

- `docker.1ms.run` 免费通道仅支持 `docker.io` 和 `ghcr.io`，其余源（`gcr.io` 等）需付费，因此本次不扩展其他源的支持
- 参考文档：https://mdoc.cc/mliev/1ms/v1.0.0/3
