# 通过镜像地址前缀方式实现镜像拉取

## 前缀替换规则

原始镜像地址：

```
stilleshan/frpc:latest
```

通过代理拉取镜像：

```
docker pull dockerproxy.net/stilleshan/frpc:latest
```

重命名镜像：

```
docker tag dockerproxy.net/stilleshan/frpc:latest stilleshan/frpc:latest
```

删除代理镜像：

```
docker rmi dockerproxy.net/stilleshan/frpc:latest
```

## 支持多种镜像地址

不同加速地址对镜像类型的支持存在差异：

**docker.1ms.run（免费仅支持 docker.io 和 ghcr.io）**

```
Docker Hub 官方镜像：
docker pull docker.1ms.run/nginx

Docker Hub 自定义镜像：
docker pull docker.1ms.run/user/image

ghcr.io 镜像（需替换域名，不是加前缀）：
docker pull ghcr.1ms.run/user/image
```

**dockerproxy.net**

```
Docker Hub 官方镜像：
docker pull dockerproxy.net/nginx

Docker Hub 自定义镜像：
docker pull dockerproxy.net/user/image

ghcr.io 镜像：
docker pull dockerproxy.net/ghcr.io/user/image
```

## 第三方镜像地址

1. `https://docker.1ms.run` — DockerHub 和 ghcr.io（免费），其他源需付费
2. `https://dockerproxy.net`
