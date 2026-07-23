#!/bin/sh
set -eu

cd "$(dirname "$0")"

docker info >/dev/null
if docker volume inspect bbs-go-instance_mysql-data >/dev/null 2>&1; then
	echo "Refusing to overwrite existing volume: bbs-go-instance_mysql-data" >&2
	exit 1
fi
if [ -n "$(docker ps -a --filter label=com.docker.compose.project=bbs-go-instance -q)" ]; then
	echo "Refusing to replace existing bbs-go-instance containers." >&2
	exit 1
fi

docker load -i docker-images.tar
mkdir -p runtime/logs
docker compose --env-file .env up -d

echo "BBS-GO restore started. MySQL will become healthy after the database import."
port=$(sed -n 's/^BBSGO_HTTP_PORT=//p' .env)
echo "Open http://127.0.0.1:${port:-3000} after 'docker compose ps' reports healthy."
