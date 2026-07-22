# 本地制作发行包

发行包在当前电脑直接生成，不依赖 GitHub Actions、Git 标签或镜像仓库。

包内包含：

- 当前源码构建出的 BBS-GO Docker 镜像
- 对应架构的 MySQL 8.4 镜像
- 当前工作区源码快照（包括未提交但未被 Git 忽略的文件）
- Docker Compose、环境变量模板、部署说明和清单

包内不会包含 `bbs-go.db`、`bbs-go.yaml`、上传文件、搜索索引、日志、`deploy/.env` 或其他被 Git 忽略的本地数据。

## 制作

启动 Docker，在项目目录执行：

```shell
go run ./cmd/package-release -version v1.0.0
```

默认生成：

```text
dist/bbs-go-v1.0.0-linux-amd64.zip
```

ARM64 设备使用：

```shell
go run ./cmd/package-release -version v1.0.0 -platform linux/arm64
```

命令会依次拉取 Node.js、Go、MySQL 基础镜像，再重新构建当前源码，因此制作阶段需要联网。输出 ZIP 包含运行所需的 BBS-GO 和 MySQL 镜像，体积会明显大于普通源码包。

## 部署

目标设备安装并启动 Docker，解压 ZIP，进入解压后的目录：

1. 将 `.env.example` 复制为 `.env`。
2. 设置两个不同的 MySQL 随机密码。
3. 加载离线镜像并启动：

```shell
docker load -i docker-images.tar
docker compose up -d
```

访问 `http://127.0.0.1:3000` 完成首次安装。目标设备不需要 Go、Node.js，也不需要联网拉取镜像。

## Docker Hub 连接失败

首次制作需要下载 Node.js、Go 和 MySQL 基础镜像。如果错误中包含 `failed to resolve source metadata`、`registry-1.docker.io` 或连接超时，应先配置 Docker Desktop 的代理或镜像加速器。

使用本机代理软件时，在 Docker Desktop 的 `Settings > Resources > Proxies` 中选择手动代理，并填写代理软件允许局域网访问的地址。例如本机代理监听 `7897` 时可尝试：

```text
HTTP proxy:  http://host.docker.internal:7897
HTTPS proxy: http://host.docker.internal:7897
```

应用设置并重启 Docker Desktop，然后先验证：

```shell
docker pull node:24-alpine
```

`blkio throttle` 和 `cgroup v1 is deprecated` 属于 Docker 环境提示，不是构建失败原因。
