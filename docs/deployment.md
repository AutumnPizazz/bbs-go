# 云服务器部署指南

> 服务器：101.34.134.223 | 项目：bbs-go (Docker Compose)
> 最后更新：2026-07-30

## 前置条件

- 本地安装了 Docker Desktop
- 代理运行在 `localhost:7897`（Clash/V2Ray，用于 Docker 构建时下载依赖）
- `D:\bbs-go\secret\sshexec.exe` —— 自研 SSH 工具，已配置好密钥

## 一键部署流程

以下命令全部在项目根目录 `D:\bbs-go` 的 Git Bash 中执行。

```bash
# ============================================
# Step 1: 构建 Docker 镜像（带代理）
# ============================================
HTTP_PROXY=http://host.docker.internal:7897 \
HTTPS_PROXY=http://host.docker.internal:7897 \
docker compose build --no-cache bbs-go

# ============================================
# Step 2: 导出镜像为 tar
# ============================================
docker save bbs-go:local -o bbs-go-images.tar

# ============================================
# Step 3: 上传到云服务器
# ============================================
MSYS_NO_PATHCONV=1 secret/sshexec.exe push bbs-go-images.tar /root/bbs-go/bbs-go-images.tar

# ============================================
# Step 4: 加载镜像并重启容器
# ============================================
# 注意：sshexec 每次执行都是独立 shell，不能用 && 串联，
# 也不能依赖 cd，必须用完整路径
secret/sshexec.exe exec "docker load < /root/bbs-go/bbs-go-images.tar"
secret/sshexec.exe exec "docker compose -f /root/bbs-go/docker-compose.yml down bbs-go"
secret/sshexec.exe exec "docker compose -f /root/bbs-go/docker-compose.yml up -d --force-recreate bbs-go"

# ============================================
# Step 5: 验证部署
# ============================================
secret/sshexec.exe exec "docker compose -f /root/bbs-go/docker-compose.yml ps"
secret/sshexec.exe exec "docker compose -f /root/bbs-go/docker-compose.yml logs --tail=20 bbs-go"
```

## 关键教训

### 1. 永远用完整镜像更新，不要单文件替换

**错误：** 本地交叉编译 → 上传二进制 → `docker cp` → restart

bbs-go 是 Docker 全栈容器 = Go 二进制 + Node SSR + node_modules + 前端 build。单换一个组件必然导致不匹配。

### 2. sshexec 每个命令独立 shell

```bash
# ❌ 不行 —— && 不生效，cd 不保持
sshexec exec "cd /root/bbs-go && docker compose down"

# ✅ 正确 —— 用完整路径
sshexec exec "docker compose -f /root/bbs-go/docker-compose.yml down bbs-go"
```

### 3. Git Bash 路径转换

Git Bash 会把 `/root/xxx` 自动转换为 `C:/Program Files/Git/root/xxx`。

**解决：** 前面加 `MSYS_NO_PATHCONV=1`

### 4. Docker Desktop 代理配置

Docker Desktop 自身配置了 `host.docker.internal:7897` 代理。如果代理没运行：
- 症状：构建时 `dial tcp 10.8.1.150:7897: connectex: No connection could be made`
- 解决：启动 Clash/V2Ray 代理

构建时设置环境变量，仅影响构建阶段，不会进入最终镜像（Dockerfile 用 `ARG` + `ENV`，`--build-arg` 覆盖）。

### 5. 快速回滚

部署前始终保留上一个正常镜像：

```bash
# 部署前备份旧镜像
docker save bbs-go:local -o bbs-go-images-backup-$(date +%Y%m%d-%H%M).tar

# 上传并加载
sshexec push bbs-go-images-backup-*.tar /root/bbs-go/
sshexec exec "docker load < /root/bbs-go/bbs-go-images-backup-*.tar"
```

## 文件清单

| 文件 | 用途 |
|------|------|
| `Dockerfile` | 多阶段构建：web-builder → server-builder → app |
| `docker-compose.yml` | 本地开发/构建用 |
| `deploy/docker-compose.yml` | 服务器部署用（通过 `${BBSGO_IMAGE}` 指定镜像） |
| `secret/sshexec.exe` | 自研 SSH 工具，已配好服务器密钥 |
| `docker/entrypoint.sh` | 容器启动脚本 |
| `docker/bbs-go-docker.yaml` | 容器内默认配置文件 |

## 本地验证

部署前在本地跑一次完整测试：

```bash
# 启动本地 Docker 环境（含 MySQL）
docker compose up -d

# 访问 http://localhost:3000 验证功能
# 确认完后关闭
docker compose down
```
