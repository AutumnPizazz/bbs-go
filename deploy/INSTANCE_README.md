# BBS-GO 实例快照包

该私密包包含论坛数据库、运行配置、搜索索引、上传文件、源码和离线镜像。数据库及后台配置可能包含用户信息和第三方服务密钥，不得公开分发。

## Windows

启动 Docker Desktop，解压 ZIP 并进入最外层目录，双击 `deploy.cmd`。也可以运行 `powershell -ExecutionPolicy Bypass -File deploy.ps1`。脚本发现同名容器或 `bbs-go-instance_mysql-data` 数据卷已存在时会拒绝部署，不会覆盖现有实例。

## Linux

```shell
chmod +x deploy.sh
./deploy.sh
```

数据库导入完成且容器健康后访问 `http://127.0.0.1:3000`。容器内部 API 的 `8082` 不需要从宿主机访问。查看状态和日志：

```shell
docker compose ps
docker compose logs -f
```

搜索索引和上传文件已包含在包内。日志目录会在目标设备重新创建。
