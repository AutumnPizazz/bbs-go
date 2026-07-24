# Ubuntu 云服务器远程替换部署

本文记录 BBS-GO 从 Windows 开发机部署到 Ubuntu 云服务器的完整流程，适用于服务器上已有其他 Docker 应用、需要先备份再替换的场景。

## 端口模型

| 部分 | 端口 |
| --- | --- |
| SSH 外部转发端口 | SSH_PORT -> 服务器 22 |
| 论坛外部访问端口 | FORUM_PORT -> 服务器 80 |
| BBS-GO 容器端口 | 服务器 80 -> 容器 3000 |
| 容器内部 API | 8082，不直接暴露 |

将 SSH_HOST、SSH_PORT、FORUM_HOST、FORUM_PORT 替换为当前服务器的可达地址。密码只从本地 secret 文件或密码管理器读取，不写入本文、命令参数、ZIP 包或日志。

## 1. 安全边界

清理远端之前必须满足以下条件：

- 已用可达的 SSH 主机名和转发端口登录成功。
- 已确认远端旧服务的数据目录和容器挂载路径。
- 已将旧应用数据下载到本机，并能用 tar -tzf 读取归档。
- 已记录归档 SHA-256。
- 已明确只删除旧应用资源，不删除 Ubuntu 系统、root 用户目录或其他无关服务。

不要使用以下方式传递密码：

    sshpass -p 'password' ...
    ssh root@host password

优先使用 SSH 密钥。必须使用密码时，使用 SSH askpass、系统密码管理器或其他不会把密码写入命令行和 shell 历史的方式。

## 2. 本地检查

在仓库根目录执行：

    go test ./...
    cd web
    corepack pnpm typecheck
    corepack pnpm lint
    cd ..
    docker compose config --quiet

Windows 环境不一定安装 make，可以直接执行上面的独立命令。

## 3. 制作实例快照包

如果要把当前本机论坛的数据库、配置、搜索索引和上传文件一起部署，使用实例快照包：

    go run ./cmd/package-release -instance -version ops-YYYYMMDD -output dist/bbs-go-instance-ops-YYYYMMDD-linux-amd64.zip

实例包包含：

- 当前 BBS-GO Docker 镜像和 MySQL 8.4 镜像。
- MySQL 一致性导出文件 mysql-init/001-bbsgo.sql.gz。
- bbs-go.yaml、搜索索引、上传文件和运行目录。
- Compose 配置和部署脚本。

实例包可能包含用户信息、OAuth、SMTP、对象存储等敏感配置，只能通过可信渠道传输。生成后记录文件哈希：

    Get-FileHash dist\bbs-go-instance-ops-YYYYMMDD-linux-amd64.zip -Algorithm SHA256

## 4. 只读盘点远端

先确认 SSH 入口和远端服务，不要一登录就执行删除：

    ssh -p SSH_PORT root@SSH_HOST "docker ps -a --format 'container={{.Names}} status={{.Status}} ports={{.Ports}} image={{.Image}}"
    ssh -p SSH_PORT root@SSH_HOST "docker inspect app --format 'image={{.Config.Image}} mounts={{range .Mounts}}{{.Source}} -> {{.Destination}}; {{end}}'"
    ssh -p SSH_PORT root@SSH_HOST "du -sh /var/discourse; df -h /"

重点确认：

- 旧容器名称、镜像名称和宿主机挂载目录。
- 应用数据是否在 /var/discourse 等 bind mount 中。
- 是否存在 Docker 命名卷。
- 根分区是否有足够空间接收新镜像和 MySQL 数据。

## 5. 下载并校验旧应用备份

容器停止后再归档应用目录，避免 PostgreSQL 等运行中数据处于不一致状态。

在本机创建备份目录：

    New-Item -ItemType Directory -Path server-backup-YYYYMMDD-HHMMSS

停止旧容器：

    ssh -p SSH_PORT root@SSH_HOST "docker stop app"

通过 SSH 下载压缩归档：

    ssh -p SSH_PORT root@SSH_HOST "tar -C / -czf - var/discourse" > server-backup-YYYYMMDD-HHMMSS\var-discourse.tar.gz

使用其他 shell 时，确保 SSH 输出按二进制流写入文件，不要经过文本编码转换。下载完成后必须校验：

    tar -tzf server-backup-YYYYMMDD-HHMMSS\var-discourse.tar.gz
    Get-FileHash server-backup-YYYYMMDD-HHMMSS\var-discourse.tar.gz -Algorithm SHA256

同时保存不包含环境变量和密码的容器元数据：

    ssh -p SSH_PORT root@SSH_HOST "docker inspect app --format 'image={{.Config.Image}} restart={{.HostConfig.RestartPolicy.Name}} mounts={{range .Mounts}}{{.Source}} -> {{.Destination}}; {{end}}'"

如果归档不能列出内容、哈希未记录或远端 SSH 中断，先恢复旧容器：

    ssh -p SSH_PORT root@SSH_HOST "docker start app"

## 6. 安装 Ubuntu Compose 依赖

服务器有 Docker 不代表已经安装 Compose。Ubuntu 24.04 可执行：

    ssh -p SSH_PORT root@SSH_HOST "apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq docker-compose-v2"
    ssh -p SSH_PORT root@SSH_HOST "docker compose version"

若出现 docker: unknown command: docker compose 或 unknown flag: --env-file，说明 Compose 插件缺失或版本不正确。先安装并验证 Compose，再执行部署脚本。

## 7. 清理旧应用资源

只有本地备份已通过校验后，才执行以下清理。命令目标必须逐项确认，不要改成 docker system prune -a 或删除整个 /root、/：

    ssh -p SSH_PORT root@SSH_HOST "docker rm -f app flamboyant_payne || true"
    ssh -p SSH_PORT root@SSH_HOST "docker image rm local_discourse/app discourse/base:TAG || true"
    ssh -p SSH_PORT root@SSH_HOST "rm -rf -- /var/discourse"

其中容器名、镜像名和 TAG 只是本次服务器的示例；后续部署应以第 4 节实际盘点结果为准。

## 8. 上传和展开 BBS-GO 实例包

先上传 ZIP，再在远端校验哈希：

    scp -P SSH_PORT dist\bbs-go-instance-ops-YYYYMMDD-linux-amd64.zip root@SSH_HOST:/root/bbs-go-instance.zip
    ssh -p SSH_PORT root@SSH_HOST "sha256sum /root/bbs-go-instance.zip"

确认远端哈希与本地一致后展开。实例包内部目录名带有版本号，因此不要写死目录名：

    mkdir -p /opt/.bbs-go-stage
    unzip -q /root/bbs-go-instance.zip -d /opt/.bbs-go-stage
    release_dir="$(find /opt/.bbs-go-stage -mindepth 1 -maxdepth 1 -type d | head -n 1)"
    test -f "$release_dir/docker-compose.yml"
    test -f "$release_dir/.env"
    test -f "$release_dir/mysql-init/001-bbsgo.sql.gz"
    mv "$release_dir" /opt/bbs-go
    rm -rf -- /opt/.bbs-go-stage

外部端口映射为 FORUM_PORT -> 服务器 80 时，将包内默认的 3000 改为服务器 80：

    sed -i 's/^BBSGO_HTTP_PORT=.*/BBSGO_HTTP_PORT=80/' /opt/bbs-go/.env

导入离线镜像并启动：

    cd /opt/bbs-go
    docker load -i docker-images.tar
    rm -f -- docker-images.tar
    mkdir -p runtime/data runtime/logs runtime/backups runtime/uploads
    docker compose config -q
    docker compose up -d

在 Compose v2 中，从 /opt/bbs-go 执行时会自动读取同目录的 .env。

## 9. 健康检查和外部验收

先检查容器：

    cd /opt/bbs-go
    docker compose ps

MySQL 和 BBS-GO 都应显示 healthy。也可以直接检查健康状态：

    mysql_id="$(docker compose ps -q mysql)"
    app_id="$(docker compose ps -q bbs-go)"
    docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$mysql_id"
    docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$app_id"

服务器本机检查：

    curl -i http://127.0.0.1/api/install/status
    curl -I http://127.0.0.1/login

外部检查：

    http://FORUM_HOST:FORUM_PORT/login

本次实际验收中：

- /api/install/status 返回 200。
- /login 返回 200。
- MySQL、BBS-GO 容器均为 healthy。
- 直接访问 / 返回“请先登录”对应的 500，是当前内容访问策略或页面处理行为，不代表容器未启动。应先通过 /login 登录；若产品要求未登录用户浏览首页，需要另行调整内容访问配置或首页错误处理。

## 10. 日常管理

    cd /opt/bbs-go
    docker compose ps
    docker compose logs --tail=100 bbs-go
    docker compose logs --tail=100 mysql
    docker compose restart
    docker compose down
    docker compose up -d

运行数据的位置：

- MySQL：Docker 命名卷 bbs-go-instance_mysql-data。
- 配置、搜索索引：/opt/bbs-go/runtime/data。
- 上传文件：/opt/bbs-go/runtime/uploads。
- 应用备份：/opt/bbs-go/runtime/backups。
- 日志：/opt/bbs-go/runtime/logs。

升级前应先做数据库和 runtime 目录备份，确认新包哈希后再替换镜像。不要使用全局 Docker 清理命令，以免误删其他项目的镜像和卷。

## 11. 常见故障

### SSH 22 端口超时

云服务器可能只把公网转发端口映射到内部 22。不要直接使用不可达的内网 IP 加 22；使用云平台提供的公网主机名和 SSH 转发端口，并先用 ssh -p PORT user@host 验证。

### docker compose 不存在

安装 Ubuntu 包 docker-compose-v2，然后确认 docker compose version 有版本输出。不要把 docker-compose 和 docker compose 混用，除非两者都已安装并确认版本。

### docker compose --env-file 报未知参数

通常是 Compose 插件尚未安装，或实际调用的不是 Compose。先执行 docker compose version；在部署目录中可以直接使用 docker compose config -q 和 docker compose up -d，让 Compose 自动读取 .env。

### MySQL 一直不健康

检查以下内容：

    docker compose logs --tail=200 mysql
    docker volume inspect bbs-go-instance_mysql-data
    test -f /opt/bbs-go/mysql-init/001-bbsgo.sql.gz && echo dump-present

实例包必须包含 SQL 导出和 999-ready.sh 初始化完成标记。不要在未确认数据的情况下删除 MySQL 命名卷。

### 容器 healthy 但首页返回 500

先访问 /login 和 /api/install/status，再查看 BBS-GO 日志。若日志是“请先登录”，说明服务和数据库已工作，但当前站点配置要求登录后才能访问内容；这与容器健康状态是两个不同问题。

### 端口被占用

    ss -lntp | grep -E ':80|:3000'
    docker ps --format 'table {{.Names}}\t{{.Ports}}'

确认旧容器已经停止并删除，再启动新 Compose。若服务器必须保留其他 80 端口服务，应改用反向代理或调整 BBSGO_HTTP_PORT，同时更新云平台端口映射。

## 12. 本次部署结果

本次流程验证了以下可复用结论：

1. 先用可达 SSH 主机名和转发端口建立只读连接，再盘点旧容器和挂载目录。
2. 停止旧容器后下载应用目录归档，并在本地验证后才清理远端。
3. 使用 -instance 发布包可以完整迁移当前 BBS-GO 数据，而不是重新安装空论坛。
4. Ubuntu 服务器需要显式安装 docker-compose-v2；Docker Engine 本身不保证 Compose 存在。
5. 云平台的论坛外部端口映射到宿主机 80 时，BBS-GO 的 BBSGO_HTTP_PORT 应设置为 80，容器内部仍使用 3000。
6. 验收应同时检查容器健康、安装状态接口、登录页面和外部转发端口。
