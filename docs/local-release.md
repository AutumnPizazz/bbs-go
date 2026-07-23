# 本地制作发行包

发行包在当前电脑直接生成，不依赖 GitHub Actions、Git 标签或镜像仓库。

通用发行包包含：

- 当前源码构建出的 BBS-GO Docker 镜像
- 对应架构的 MySQL 8.4 镜像
- 当前工作区源码快照（包括未提交但未被 Git 忽略的文件）
- Docker Compose、环境变量模板、部署说明和清单

通用发行包不会包含 `bbs-go.db`、`bbs-go.yaml`、上传文件、搜索索引、日志、`deploy/.env` 或其他被 Git 忽略的本地数据。实例快照包是例外，见文档末尾的“制作实例快照包”。

## 制作通用发行包

启动 Docker，在项目目录执行：

```shell
go run ./cmd/package-release -version v1.0.1
```

默认生成：

```text
dist/bbs-go-v1.0.1-linux-amd64.zip
```

ARM64 设备使用：

```shell
go run ./cmd/package-release -version v1.0.1 -platform linux/arm64
```

命令会依次拉取 Node.js、Go、MySQL 基础镜像，再重新构建当前源码，因此制作阶段需要联网。输出 ZIP 包含运行所需的 BBS-GO 和 MySQL 镜像，体积会明显大于普通源码包。

每个版本对应一个固定输出文件。为避免误覆盖，输出文件已经存在时命令会失败：

```text
Error: output already exists: ...
```

这不是打包失败。重新打包时请使用新的版本号，或者显式指定一个不存在的输出路径：

```shell
go run ./cmd/package-release -version v1.0.2
go run ./cmd/package-release -version v1.0.1 -output dist/bbs-go-v1.0.1-rebuild.zip
```

不要直接删除旧包后覆盖，除非已经确认旧包不再需要。`-version` 同时写入镜像标签、ZIP 名称和包内 manifest，实例恢复时应保持三者一致。

## 部署通用发行包

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

## 制作实例快照包

当前论坛已经使用 MySQL 且安装完成后，可以制作包含数据库、配置、搜索索引和上传文件的私密实例包：

```shell
go run ./cmd/package-release -instance -version v1.0.1
```

命令会先构建镜像，再短暂停止 BBS-GO 写入，导出一致的 MySQL 快照并复制持久文件，完成后自动重新启动论坛。默认输出：

```text
dist/bbs-go-instance-v1.0.1-linux-amd64.zip
```

实例包可能包含用户信息、OAuth、SMTP、OSS 等密钥，只能通过可信渠道传输。目标设备解压 ZIP 后，进入最外层目录：

- Windows：双击 `deploy.cmd`，或运行 `powershell -ExecutionPolicy Bypass -File deploy.ps1`。
- Linux：执行 `./deploy.sh`。

部署脚本会加载包内镜像、导入 `mysql-init/001-bbsgo.sql.gz`，等待数据库恢复完成后启动 BBS-GO。发现同名容器或数据卷时会拒绝覆盖。恢复完成后访问 `http://127.0.0.1:3000`，而不是容器内部的 `8082`。

实例包同样遵守输出文件不覆盖规则。例如已经生成 `v1.0.1` 后再次执行同一命令会得到 `output already exists`，应改用 `v1.0.2` 或新的 `-output` 路径。
