# github 文件下载

支持GitHub 文件、Releases、Archive、Gist、raw.githubusercontent.com文件加速下载，不支持整个项目文件夹的下载。

## 支持的站点

- https://ghfast.top
- https://gh-proxy.com
- https://gh-proxy.org
- https://fastgit.cc
- https://ghproxy.net
- https://ghproxylist.com
- https://hk.gh-proxy.org
- https://cdn.gh-proxy.org
- https://ghp.keleyaa.com
- https://gh.jasonzeng.dev


## 站点说明

- [GitHub Proxy 加速](https://ghfast.top/)
- [GitHub 文件加速](https://ghproxy.net/)
- [GitHub加速下载代理 - 快速访问 GitHub 文件](https://gh-proxy.com/)
- [GitHub 文件加速](https://fastgit.cc/)
- [GitHub 文件加速](https://ghp.keleyaa.com/)
- [GitHub 文件加速](https://gh.jasonzeng.dev/)
- [GitHub 代理下载服务合集 | 文件加速器](https://ghproxylist.com)

## 加速原理分析

原理：将加速地址拼接在原始 github 下载地址之前，实现加速下载。

原始 github release 下载地址：

```
https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
```

加速下载地址：

```
https://ghproxy.net/https://github.com/leafney/docker-mirror-proxy/releases/download/v0.0.4/dmp-linux-amd64.tar.gz
```

## 下载方式

通过当前系统自带的 `wget` 或者 `curl` 命令下载文件到当前目录。

```
wget https://ghfast.top/https://github.com/stilleshan/dockerfiles/archive/master.zip

curl -O https://ghfast.top/https://github.com/stilleshan/dockerfiles/archive/master.zip
```

---