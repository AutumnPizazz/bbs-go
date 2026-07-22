# BBS-GO 部署包

该目录用于部署一个全新的 BBS-GO 实例，不包含开发设备上的数据库、`bbs-go.yaml`、上传文件、日志或密码。

## 在线部署

1. 将 `.env.example` 复制为 `.env`。
2. 将 `BBSGO_IMAGE` 改成 Release 页面标明的完整镜像和版本。
3. 为两个 MySQL 密码设置不同的长随机值。
4. 启动服务：

```shell
docker compose up -d
```

访问 `http://127.0.0.1:3000` 完成首次安装。部署到服务器时，应继续配置域名、HTTPS 和反向代理。

## 离线部署

在能够访问镜像仓库的同架构设备上执行：

```shell
docker pull IMAGE:VERSION
docker pull mysql:8.4
docker save -o bbs-go-images.tar IMAGE:VERSION mysql:8.4
```

将部署目录和 `bbs-go-images.tar` 复制到目标设备，然后执行：

```shell
docker load -i bbs-go-images.tar
docker compose up -d
```

## 日常管理

```shell
docker compose ps
docker compose logs -f
docker compose down
docker compose up -d
```

运行数据保存在 MySQL 命名卷以及当前目录的 `data/`、`logs/`、`uploads/` 中，不属于发行包。
