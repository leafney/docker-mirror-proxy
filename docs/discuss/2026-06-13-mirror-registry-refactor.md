# 讨论：镜像加速地址重构

日期：2026-06-13

## 原始需求

1. 移除 docker 加速地址 `registry.cyou`、`proxy.vvvv.ee`
2. 针对 `docker.1ms.run` 地址，按镜像类型改变拼接方式：
   - DockerHub (`docker.io`) 镜像：`docker.1ms.run/<original>`（原逻辑不变）
   - GHCR (`ghcr.io`) 镜像：`ghcr.1ms.run/<path>`（去掉 `ghcr.io/` 前缀，换域名）
3. 增加原始默认地址作为兜底，排在所有加速地址之后

参考文档：https://mdoc.cc/mliev/1ms/v1.0.0/3

## 问答记录

**Q1：`docker.1ms.run` 的 GHCR 格式如何处理？**

文档说 ghcr.io 镜像需将域名替换为 `ghcr.1ms.run`，例如：
- 输入：`ghcr.io/leafney/ai-signin:0.6.8`
- 加速：`ghcr.1ms.run/leafney/ai-signin:0.6.8`

当前代码对所有加速地址统一用 `registry + "/" + ref.Original`，会错误生成 `docker.1ms.run/ghcr.io/leafney/ai-signin:0.6.8`。

建议：`Candidates()` 按 `ref.Type` 分支处理 `docker.1ms.run`。

**A1：确认。**

---

**Q2：兜底地址格式？**

- DockerHub 兜底：`docker.io/<original>`（加前缀，因为 `ref.Original` 无域名）
- GHCR 兜底：直接用 `ref.Original`（已含 `ghcr.io/` 域名）

**A2：确认。**

---

**Q3：`dockerproxy.net` 对 GHCR 的处理方式？**

保持旧逻辑不变：`dockerproxy.net/ghcr.io/leafney/ai-signin:0.6.8`，不做特殊处理。

**A3：确认，保持旧逻辑。**

---

**Q4：兜底地址的完整范围？**

只加两个，与当前支持的镜像类型一致：
- DockerHub → `docker.io/<original>`
- GHCR → `ref.Original`（已含域名，直接使用）

**A4：确认，只加这两个。**

## 最终共识

### `resolver.go` 变更

`DefaultRegistries` 移除 `registry.cyou`、`proxy.vvvv.ee`，保留：
- `docker.1ms.run`（需特殊处理）
- `dockerproxy.net`（旧逻辑不变）

`Candidates()` 逻辑：

```
for each registry:
  if registry == "docker.1ms.run":
    if ref.Type == TypeGHCR:
      append "ghcr.1ms.run/" + stripPrefix(ref.Original, "ghcr.io/")
    else:
      append "docker.1ms.run/" + ref.Original
  else:
    append registry + "/" + ref.Original

// 兜底（最后追加）
if ref.Type == TypeGHCR:
  append ref.Original          // e.g. ghcr.io/leafney/ai-signin:0.6.8
else:
  append "docker.io/" + ref.Original  // e.g. docker.io/nginx:latest
```

### 测试变更

`resolver_test.go` 中三个测试用例的 `want` 列表需同步更新。
