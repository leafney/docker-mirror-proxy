# dmp pull 代理配置命令优化讨论

## 原始需求

根据 `docs/docker-proxy.md`，为 `dmp pull` 命令给出优化方案。用户进一步提出希望把 Docker 代理设置操作融入命令中：

- `dmp pull proxy set xxx:7890`：为当前环境下的 Docker 写入代理配置并重启 Docker 服务
- `dmp pull proxy rm`：移除代理配置
- 后续补充第一版还需要 `dmp pull proxy sts`：查看当前 Docker 是否正确配置代理，并验证代理是否可用

## 问答记录

### Q1：`dmp pull` 优化目标应选哪条主线？

推荐答案：做“代理配置诊断 + 指引”，不自动修改 systemd/Docker Desktop 配置。原因是自动改系统配置需要管理员权限，平台差异大，风险较高。

用户答案：用户希望把代理设置操作融入命令中，命令负责写入 Docker 代理配置并重启 Docker 服务；同时询问需要 sudo 权限时是否可以给出提示。

### Q2：`dmp pull proxy set xxx:7890` 遇到需要管理员权限时，命令应怎么处理？

推荐答案：命令不隐式提权，只检测权限；无权限时打印 `sudo dmp pull proxy set ...` 或 `sudo dmp pull proxy rm`，让用户明确重新执行。

用户答案：接受。

### Q3：`set` 写入代理地址时，是否自动补协议？

推荐答案：自动补 `http://`。输入 `192.168.8.100:7890` 时写入 `http://192.168.8.100:7890`；输入 `http://...` 原样使用；输入 `https://...` 原样使用但提示 Docker daemon 代理常用 `http://`；输入 `socks5://...` 拒绝。

用户答案：确认。

### Q4：`proxy set/rm` 要支持哪些环境？

推荐答案：先支持 Linux systemd + Docker Desktop，先不做 WSL2 专门适配。

用户答案：默认先支持 Linux systemd，amd64 + arm64；其他系统提示暂不支持。

### Q5：`proxy set` 写入文件时，是否覆盖整个 `http-proxy.conf`？

推荐答案：只管理 dmp 专属文件，不碰用户已有代理文件。使用 `/etc/systemd/system/docker.service.d/dmp-http-proxy.conf`。`rm` 只删除这个文件，不删除用户手写的 `http-proxy.conf`。

用户答案：确认。

### Q6：`proxy` 子命令要怎么挂在 `pull` 下？

推荐答案：挂在 `dmp pull proxy ...` 下，不做顶层命令。

用户答案：确认。

### Q7：第一版支持哪些 `proxy` 子命令？`sts` 如何验证？

推荐答案：支持 `set`、`rm`、`status`，其中状态命令查看当前配置并可做验证。

用户答案：第一版支持三个命令：`set`、`rm`、`sts`。`sts` 用于查看当前 Docker 是否正确配置代理，最好能给出验证，看是否正常使用代理。

### Q8：`sts` 验证“代理是否正常使用”，验证到哪一层？

推荐答案：配置检查 + Docker 实拉探测，例如 `docker pull hello-world`。

用户答案：不做真实镜像拉取。采用两个验证：给出 `systemctl show` 提示；通过 `curl -x` 验证代理可用。

### Q9：`sts` 是否真正执行 `systemctl show` 和 `curl -x`，还是只打印命令给用户？

推荐答案：默认执行检查，同时打印对应命令。执行 `systemctl show --property=Environment docker` 判断 Docker 服务是否已加载代理；若配置文件中有代理地址，执行 `curl -x <proxy> https://registry-1.docker.io/v2/ -I` 验证代理连通性；不执行 `docker pull`。

用户答案：确认。

### Q10：`set` / `rm` 后是否自动执行 `systemctl daemon-reload` 和 `systemctl restart docker`？

推荐答案：默认自动执行。失败则给出手动命令。执行前提示“将重启 Docker 服务”，因为重启会影响当前运行容器。

用户答案：确认。执行过程中每一步都给出日志提示信息。

### Q11：`rm` 时如果 `dmp-http-proxy.conf` 不存在，要怎么处理？

推荐答案：幂等成功。文件不存在时直接输出无需移除并完成，不重启 Docker。

用户答案：确认。

## 最终共识

第一版新增 `dmp pull proxy` 子命令组，仅支持 Linux systemd 环境，架构限定 amd64 和 arm64。其他操作系统、非 systemd 环境、非支持架构统一提示暂不支持。

命令形态：

- `dmp pull proxy set <proxy>`：写入 dmp 专属 Docker daemon 代理配置，执行 `systemctl daemon-reload` 和 `systemctl restart docker`
- `dmp pull proxy rm`：删除 dmp 专属代理配置，若文件存在则 reload 并重启 Docker；若文件不存在则直接成功
- `dmp pull proxy sts`：查看 dmp 代理配置、检查 Docker 服务已加载的环境变量，并用 `curl -x` 验证代理地址连通性；不执行真实镜像拉取

代理配置文件固定为：

```text
/etc/systemd/system/docker.service.d/dmp-http-proxy.conf
```

`set` 写入内容：

```ini
[Service]
Environment="HTTP_PROXY=http://192.168.8.100:7890"
Environment="HTTPS_PROXY=http://192.168.8.100:7890"
Environment="NO_PROXY=localhost,127.0.0.1,::1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,*.local"
```

代理地址规则：

- 未带协议时自动补 `http://`
- `http://` 原样使用
- `https://` 原样使用，但提示 Docker daemon 代理常用 HTTP 代理地址
- `socks5://` 拒绝

权限规则：

- 命令不隐式调用 `sudo`
- 如果无权限写入、删除配置或重启 Docker，打印 `sudo dmp pull proxy ...` 或手动 `sudo systemctl ...` 提示
- 每一步输出 `[dmp]` 日志，包含检测环境、写入/删除配置、reload、restart、验证命令、验证结果
