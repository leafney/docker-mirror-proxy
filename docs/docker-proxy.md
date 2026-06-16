# 本地代理

### **通过配置 systemd 的 `http-proxy.conf` 为 Docker 守护进程设置 HTTP/HTTPS 代理，即可让镜像拉取走 192.168.8.100:7890。**

下面给出在 Linux(systemd 管理的 Docker)环境下的完整操作步骤。Windows / macOS 桌面版 Docker Desktop 的设置方式不同,文末会单独说明。

---

### **一、原理说明**

Docker 拉取镜像的网络请求由后台守护进程 `dockerd` 发起,而不是当前 shell。所以即便你在终端里 `export http_proxy=...`,对 `docker pull` 也是无效的。正确做法是给 `docker.service` 这个 systemd 服务单独注入代理环境变量,让守护进程本身走代理。

需要注意:`192.168.8.100:7890` 是 ShellCrash 提供的 **HTTP 代理端口**(Clash/Mihomo 默认混合端口),所以 `http_proxy` 和 `https_proxy` 都填这个地址即可,不需要写成 `socks5://`。

---

### **二、为 Docker 守护进程配置代理**

#### **1. 创建 systemd 配置目录**

```bash
sudo mkdir -p /etc/systemd/system/docker.service.d
```

#### **2. 新建代理配置文件**

```bash
sudo vim /etc/systemd/system/docker.service.d/http-proxy.conf
```

写入以下内容(把局域网网段加入 `NO_PROXY`,避免内网访问也被代理):

```ini
[Service]
Environment="HTTP_PROXY=http://192.168.8.100:7890"
Environment="HTTPS_PROXY=http://192.168.8.100:7890"
Environment="NO_PROXY=localhost,127.0.0.1,::1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12,*.local"
```

#### **3. 重新加载并重启 Docker**

```bash
sudo systemctl daemon-reload
sudo systemctl restart docker
```

#### **4. 验证代理是否生效**

```bash
sudo systemctl show --property=Environment docker
```

应能看到刚才配置的三个变量。然后试拉一个官方镜像:

```bash
docker pull hello-world
```

如果能正常完成,说明守护进程已经走 ShellCrash 代理。

---

### **三、为 `docker build` 和容器运行时配置代理(可选)**

上面的配置只让 **守护进程拉镜像** 走代理。如果你在 `docker build` 过程中(例如 `RUN apt-get update`、`pip install`)也需要代理,要给 **客户端** 再配一份。

编辑(或新建) `~/.docker/config.json`:

```json
{
  "proxies": {
    "default": {
      "httpProxy": "http://192.168.8.100:7890",
      "httpsProxy": "http://192.168.8.100:7890",
      "noProxy": "localhost,127.0.0.1,192.168.0.0/16,10.0.0.0/8,172.16.0.0/12"
    }
  }
}
```

配置后,`docker run` 启动的容器和 `docker build` 过程会自动注入 `HTTP_PROXY` 等环境变量。无需重启 Docker,只对配置后新创建的容器生效。

---

### **四、关键排错点**

第一,确认 ShellCrash 代理对 **整个局域网** 开放,而不仅是本机回环。在路由器/主机上测试:

```bash
curl -x http://192.168.8.100:7890 https://www.google.com -I
```

返回 200/301 即代理可达。如果连不上,需要在 ShellCrash 里检查 "允许局域网连接" 是否开启,以及 Docker 主机所在网段是否在分流规则中走了代理出站(而不是 DIRECT)。

第二,如果 Docker 跑在 **WSL2** 里,`192.168.8.100` 通常是路由器或物理机地址,WSL2 默认能直接访问,无需特殊处理;但如果代理只监听在 Windows 主机的 `127.0.0.1`,那就要换成 Windows 主机在 WSL 网络中的 IP。

第三,之前配置的镜像加速器(`registry-mirrors`)在你已经能直连官方源时可以保留也可以删掉,二者不冲突。但要注意**镜像加速器不会代理 `docker login`、`docker push` 到 Docker Hub 的流量**,这类操作必须依赖上面的 HTTP 代理配置才能成功。

---

### **五、Docker Desktop(Windows / macOS)的配置方式**

如果你用的是桌面版 Docker Desktop,不要改 systemd(也没有)。打开 **Settings → Resources → Proxies**,选择 "Manual proxy configuration",把 Web Server (HTTP) 和 Secure Web Server (HTTPS) 都填 `http://192.168.8.100:7890`,Bypass 里加上 `localhost,127.0.0.1,*.local`,Apply & Restart 即可。

按以上步骤配置后,`docker pull` 应该就能稳定从 Docker Hub 直接拉取镜像了。