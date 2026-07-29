# 云服务器部署教训

> 日期：2026-07-29 | 服务器：101.34.134.223 | 项目：bbs-go (Docker Compose)

## 致命错误：用 `docker cp` 替换单个文件

**错误做法：**
```
本地交叉编译 bbs-go-linux → 上传 → docker cp 进容器 → restart
```

**为什么失败：** bbs-go 是 Docker 全栈容器，包含 Go 二进制 + Node SSR + node_modules + 前端 build。只换了 Go 二进制，Node 端仍是旧代码，导致前端出现「旋转验证码 + 滑块验证码杂交」的混沌状态。

**正确做法：** 完整镜像部署。

## Docker 镜像的正确部署流程

```
1. 本机构建：docker compose build --no-cache
   （需要代理时：HTTP_PROXY=http://host.docker.internal:7897 docker compose build）
2. 导出镜像：docker save bbs-go:local -o bbs-go-images.tar
3. 上传：   sshexec push bbs-go-images.tar /root/bbs-go/bbs-go-images.tar
4. 加载：   docker load < /root/bbs-go/bbs-go-images.tar
5. 部署：   docker compose up -d --force-recreate bbs-go
```

**原则：对于 Docker 部署的项目，永远用完整镜像更新，不要试图「优化」为单文件替换。**

## Dockerfile 注意：代理只在构建阶段用

```dockerfile
# ✅ 构建阶段可以有代理（用于下载依赖）
FROM golang:alpine AS builder
ARG HTTP_PROXY
ENV HTTP_PROXY=${HTTP_PROXY}

# ❌ 最终镜像不能有代理配置
FROM node:bookworm AS app
# 不要 ENV HTTP_PROXY=... ！
```

`docker compose build` 时通过环境变量传入代理，仅影响构建阶段。最终镜像的 ENV 不包含代理。

## Git Bash 路径转换坑

Git Bash 会自动把 `/root/xxx` 转换为 `C:/Program Files/Git/root/xxx`。

**解决：** `MSYS_NO_PATHCONV=1 sshexec push file /root/path`

## 其他教训

- 部署前先在本地完整验证（本地 Docker 跑起来测）
- 不要把交叉编译的裸二进制直接用于 Docker 部署
- `docker compose down service-name` 可以只重建一个服务，不影响其他容器（如 MySQL）
- 云服务器上始终保留上一个正常镜像的 tar 备份（快速回滚用）
