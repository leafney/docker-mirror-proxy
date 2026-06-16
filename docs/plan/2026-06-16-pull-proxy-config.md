Related discussion: [docs/discuss/2026-06-16-pull-proxy-config.md](../discuss/2026-06-16-pull-proxy-config.md)

# dmp pull 代理配置命令 PRD

## Problem Statement

用户在 Linux systemd 环境中使用 Docker 拉取镜像时，代理必须配置到 Docker daemon，而不是当前 shell。现有 `dmp pull` 只负责按镜像加速地址拉取镜像，不能帮助用户把 HTTP/HTTPS 代理写入 Docker daemon 配置，也不能验证 Docker 服务是否已加载代理配置。

用户希望把 `docs/docker-proxy.md` 中的手工配置步骤沉淀为命令能力，通过 `dmp pull proxy set/rm/sts` 完成代理配置写入、移除、Docker 服务重启和状态检查。

## Solution

在 `dmp pull` 下新增 `proxy` 子命令组，第一版只支持 Linux systemd 环境，架构限定 amd64 和 arm64。命令负责管理 dmp 专属 Docker daemon 代理配置文件，不修改用户已有代理文件。

新增命令：

1. `dmp pull proxy set <proxy>`：规范化代理地址，写入 dmp 专属 systemd drop-in 配置文件，执行 `systemctl daemon-reload` 和 `systemctl restart docker`。
2. `dmp pull proxy rm`：删除 dmp 专属配置文件，若有变更则 reload 并重启 Docker；若配置文件不存在则直接成功。
3. `dmp pull proxy sts`：检查当前 dmp 代理配置、检查 Docker 服务加载的环境变量，并通过 `curl -x` 验证代理连通性；不执行真实镜像拉取。

命令不隐式调用 `sudo`。权限不足时输出清晰提示，让用户使用 `sudo dmp pull proxy ...` 重新执行，或手动执行对应 `systemctl` 命令。

## User Stories

1. 作为 Linux 用户，我想用 `dmp pull proxy set 192.168.8.100:7890` 写入 Docker daemon 代理配置，从而不用手动编辑 systemd 配置文件。
2. 作为 Linux 用户，我想输入不带协议的代理地址时自动补 `http://`，从而减少命令输入成本。
3. 作为 Linux 用户，我想在输入 `socks5://` 代理时收到明确错误，从而避免写入 Docker daemon 不适用的代理格式。
4. 作为 Linux 用户，我想 `set` 成功后自动执行 `systemctl daemon-reload`，从而让 systemd 读取新配置。
5. 作为 Linux 用户，我想 `set` 成功后自动重启 Docker 服务，从而让代理配置立即生效。
6. 作为 Linux 用户，我想在 Docker 重启前看到明确日志提示，从而知道当前操作会影响正在运行的容器。
7. 作为普通用户，我想在权限不足时看到 `sudo dmp pull proxy set ...` 提示，从而知道下一步怎么做。
8. 作为普通用户，我想命令不要在内部弹出 sudo 密码交互，从而避免脚本或终端卡住。
9. 作为 Linux 用户，我想 `dmp pull proxy rm` 只移除 dmp 写入的配置，从而不破坏我手工维护的 Docker 代理文件。
10. 作为 Linux 用户，我想 `rm` 后自动 reload 并重启 Docker，从而让移除代理立即生效。
11. 作为 Linux 用户，我想 `rm` 在配置文件不存在时直接成功，从而可以放心重复执行。
12. 作为 Linux 用户，我想 `dmp pull proxy sts` 显示 dmp 配置文件是否存在，从而判断当前是否由 dmp 管理代理。
13. 作为 Linux 用户，我想 `sts` 显示 Docker 服务当前加载的 Environment，从而确认 systemd 是否已加载代理变量。
14. 作为 Linux 用户，我想 `sts` 执行并展示 `systemctl show --property=Environment docker`，从而能对照文档手动验证。
15. 作为 Linux 用户，我想 `sts` 使用 `curl -x <proxy> https://registry-1.docker.io/v2/ -I` 验证代理连通性，从而知道代理地址本身是否可用。
16. 作为 Linux 用户，我不想 `sts` 执行 `docker pull hello-world`，从而避免产生镜像拉取副作用。
17. 作为 macOS、Windows、非 systemd Linux 或非 amd64/arm64 用户，我想看到“暂不支持”的明确提示，从而不会误以为命令已配置成功。
18. 作为排障用户，我想每一步都有 `[dmp]` 日志，从而知道命令卡在哪一步或失败在哪一步。

## Implementation Decisions

- 新增 `pull proxy` 子命令组，挂在现有 `pull` 命令下，不新增顶层命令。
- 第一版支持命令名为 `set`、`rm`、`sts`，不使用 `status`。
- 只支持 Linux systemd 环境，并检查运行架构为 amd64 或 arm64。
- 非支持环境返回错误并输出暂不支持提示。
- dmp 只管理专属 drop-in 文件：`/etc/systemd/system/docker.service.d/dmp-http-proxy.conf`。
- `set` 会确保配置目录存在，并写入 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY`。
- 默认 `NO_PROXY` 使用文档中的局域网配置：`localhost,127.0.0.1,::1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,*.local`。
- 代理地址规范化规则：无协议自动补 `http://`；`http://` 原样使用；`https://` 原样使用并打印提示；`socks5://` 拒绝。
- 命令不调用 `sudo`。遇到写文件、删文件、reload 或 restart 权限不足时，打印重新执行建议或手动命令。
- `set` 写入配置后执行 `systemctl daemon-reload` 和 `systemctl restart docker`。
- `rm` 删除配置后执行 `systemctl daemon-reload` 和 `systemctl restart docker`。
- `rm` 在配置文件不存在时直接成功，不执行 reload 和 restart。
- `sts` 执行 `systemctl show --property=Environment docker` 并检查输出中是否包含代理环境变量。
- `sts` 从 dmp 配置文件读取代理地址，若存在则执行 `curl -x <proxy> https://registry-1.docker.io/v2/ -I` 验证代理连通性。
- `sts` 始终打印等价验证命令，方便用户手动复查。
- `sts` 不执行 `docker pull`。
- 系统命令执行应封装成可替换接口，便于测试命令参数、错误处理和日志输出。

## Testing Decisions

- 测试应覆盖外部行为，不依赖具体内部实现细节。
- CLI 测试覆盖 `dmp pull proxy set`、`rm`、`sts` 命令能被识别，参数缺失时返回清晰错误。
- 代理地址规范化逻辑需要单元测试，覆盖无协议、HTTP、HTTPS、SOCKS5 和空值。
- 环境支持检测需要单元测试，覆盖 Linux amd64、Linux arm64、非 Linux、非支持架构、非 systemd。
- 配置渲染逻辑需要单元测试，确保写入 `HTTP_PROXY`、`HTTPS_PROXY`、`NO_PROXY`。
- `set` 流程需要测试成功路径和权限不足路径，确认日志包含写入、reload、restart 提示。
- `rm` 流程需要测试文件存在和文件不存在两种路径，确认文件不存在时不执行 reload/restart。
- `sts` 流程需要测试 systemctl 输出匹配、未匹配、curl 成功和 curl 失败。
- 不做真实 `/etc` 写入测试，不真实重启 Docker，不真实执行 curl；使用 fake 文件系统或临时目录与 fake runner 验证行为。
- 现有 `pull` 拉取流程测试保持不变，避免代理配置命令影响镜像拉取主流程。

## Out of Scope

- 不支持 macOS Docker Desktop 自动配置。
- 不支持 Windows Docker Desktop 自动配置。
- 不做 WSL2 专门适配。
- 不修改用户已有的 `/etc/systemd/system/docker.service.d/http-proxy.conf`。
- 不支持 SOCKS5 代理写入 Docker daemon 配置。
- 不在命令内部隐式调用 sudo 或处理 sudo 密码输入。
- `sts` 不执行 `docker pull hello-world` 或其他真实镜像拉取。
- 不新增 `dmp pull --proxy` 参数。
- 不改变现有镜像加速地址池和 `dmp pull <image>` 拉取逻辑。

## Further Notes

该功能会重启 Docker 服务，执行前必须输出明确日志提示。实现时应优先保证错误信息清楚，避免用户在权限不足、非 systemd 环境或代理不可达时误判为配置成功。
