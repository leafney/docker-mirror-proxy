Related discussion: [docs/discuss/2026-05-27-pull-progress-timeout.md](../discuss/2026-05-27-pull-progress-timeout.md)

# dmp pull 无进展超时优化 PRD

## Problem Statement

用户使用 `dmp pull` 拉取较大的 Docker 镜像时，镜像下载总耗时可能超过默认 60 秒。当前 `--timeout` 被用作单个候选加速地址的总耗时上限，导致 Docker 仍在正常输出下载进度时，也会被强制停止并切换到下一个加速地址。

用户需要的是识别“长时间无进展”的卡住场景，而不是限制正常下载的大镜像总耗时。

## Solution

将 `dmp pull` 的 `--timeout` 语义调整为“无进展超时”。当 `docker pull` 在指定秒数内没有任何 stdout 或 stderr 输出时，才取消当前候选地址并切换到下一个地址。

新增 `--max-time` 参数和 `-m` 短参数，用于设置单个候选地址的总耗时上限。默认值为 0，表示不限制总耗时。

## User Stories

1. 作为 `dmp pull` 用户，我希望大镜像只要仍在下载就不会因为超过 60 秒被中断，从而避免正常下载被误杀。
2. 作为 `dmp pull` 用户，我希望镜像源长时间无响应时可以自动切换到下一个源，从而避免命令一直卡住。
3. 作为 `dmp pull` 用户，我希望继续使用 `--timeout 60` 表达无进展超时，从而保持命令易懂。
4. 作为 `dmp pull` 用户，我希望可以设置总耗时上限，从而处理持续输出但耗时异常长的镜像源。
5. 作为 `dmp pull` 用户，我希望总耗时上限默认关闭，从而让大镜像下载不受默认总时间限制。
6. 作为 `dmp pull` 用户，我希望普通拉取失败仍会自动切换镜像源，从而保持现有容错体验。
7. 作为 `dmp pull` 用户，我希望日志能区分普通失败、无进展超时和总耗时超时，从而更容易判断失败原因。
8. 作为 `dmp pull` 用户，我希望帮助文档清楚描述 `--timeout` 和 `--max-time` 的区别，从而正确选择参数。

## Implementation Decisions

- 调整 Docker 拉取流程，取消把 `--timeout` 直接作为整个 `docker pull` 的 context deadline。
- Docker 命令 stdout 和 stderr 需要经过可观察的 writer。任意输出都视为有进展，并刷新最后进展时间。
- 拉取过程中启动无进展监控。超过 `--timeout` 秒没有输出时，取消当前 Docker 命令。
- 新增总耗时监控。`--max-time` 大于 0 时，超过该时间取消当前 Docker 命令。
- `--max-time` 使用 `-m` 作为短参数，不支持 `-mt`。
- Docker 命令非零退出时保持现有行为，记录失败并切换下一个候选地址。
- 日志需要明确输出不同失败原因：普通失败、无进展超时、总耗时超时。
- CLI 帮助文本和 README 需要同步更新参数语义。

## Testing Decisions

- 测试应覆盖外部行为，不绑定内部 goroutine 或 timer 实现细节。
- 应为应用层拉取流程增加测试，覆盖拉取失败后切换候选地址的既有行为仍保留。
- 应增加无进展超时测试，验证长期无输出时当前候选地址被取消并切换到下一个地址。
- 应增加有输出时不触发无进展超时的测试，验证大镜像持续输出不会被误杀。
- 应增加总耗时上限测试，验证 `--max-time` 大于 0 时可以取消耗时过长的候选地址。
- 应增加 CLI 参数测试，验证 `--max-time` 和 `-m` 可用，非正数参数按设计处理。

## Out of Scope

- 不解析 Docker 输出中的具体下载字节数。
- 不判断输出内容是否重复。
- 不对 `unknown blob`、`manifest unknown` 等 Docker 错误做分类策略。
- 不改变镜像候选地址生成逻辑。
- 不改变拉取成功后的 tag 和 rmi 流程。
- 不支持 `-mt` 多字符短参数。

## Further Notes

当前日志示例中的 `unknown blob` 属于 Docker pull 非零退出，不属于无进展超时。优化后该类错误仍会直接切换下一个源。

如果未来发现某些 Docker 版本在卡住时仍持续输出重复内容，可以再升级为解析下载字节数变化或层状态变化。本 PRD 先采用更简单稳定的输出活动检测。
