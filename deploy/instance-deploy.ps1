$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

docker info *> $null
if ($LASTEXITCODE -ne 0) {
    throw "Docker is not running."
}

docker volume inspect bbs-go-instance_mysql-data *> $null
if ($LASTEXITCODE -eq 0) {
    throw "Refusing to overwrite existing volume: bbs-go-instance_mysql-data"
}
$existingContainer = docker ps -a --filter "label=com.docker.compose.project=bbs-go-instance" -q
if ($existingContainer) {
    throw "Refusing to replace existing bbs-go-instance containers."
}

docker load -i docker-images.tar
if ($LASTEXITCODE -ne 0) {
    throw "Failed to load Docker images."
}

New-Item -ItemType Directory -Force runtime/logs | Out-Null
docker compose --env-file .env up -d
if ($LASTEXITCODE -ne 0) {
    throw "Failed to start BBS-GO."
}

Write-Host "BBS-GO restore started. MySQL will become healthy after the database import."
$port = ((Get-Content .env | Where-Object { $_ -match '^BBSGO_HTTP_PORT=' }) -split '=', 2)[1]
Write-Host "Open http://127.0.0.1:$port after 'docker compose ps' reports healthy."
