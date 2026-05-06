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

```
Docker 官方镜像：
docker pull proxy.vvvv.ee/nginx
Docker 镜像：
docker pull proxy.vvvv.ee/user/image
ghcr.io 镜像：
docker pull proxy.vvvv.ee/ghcr.io/user/image
Quay.io 镜像：
docker pull proxy.vvvv.ee/quay.io/org/image
Kubernetes 镜像：
docker pull proxy.vvvv.ee/registry.k8s.io/pause:3.8
```

## 第三方镜像地址

1. `https://docker.1ms.run`
2. `https://dockerproxy.net`
3. `https://proxy.vvvv.ee`
4. `https://registry.cyou`
