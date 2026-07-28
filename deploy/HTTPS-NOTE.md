# HTTPS 配置 — forum.tsntec.com

> 部署于 2026-07-28。本文件为云端 `/www/server/panel/vhost/nginx/forum.tsntec.com.conf` 的本地镜像。

## Nginx 配置

```nginx
server
{
    listen 80;
    server_name forum.tsntec.com;
    client_max_body_size 100M;

    # ACME HTTP 验证（证书续期时需要）
    location ^~ /.well-known/acme-challenge/ {
        root /www/wwwroot/tsntec.com;
    }

    # 其余请求强制跳转 HTTPS
    location /
    {
        return 301 https://$host$request_uri;
    }
}

server
{
    listen 443 ssl http2;
    server_name forum.tsntec.com;
    client_max_body_size 100M;

    ssl_certificate /www/server/panel/vhost/cert/tsntec.com/fullchain.pem;
    ssl_certificate_key /www/server/panel/vhost/cert/tsntec.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers EECDH+CHACHA20:EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    add_header Strict-Transport-Security "max-age=31536000";
    error_page 497 https://$host$request_uri;

    location /
    {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_read_timeout 120s;
    }

    access_log /www/wwwlogs/forum.tsntec.com.log;
    error_log /www/wwwlogs/forum.tsntec.com.error.log;
}
```

## 证书管理

- **证书提供商**：Let's Encrypt（免费）
- **管理工具**：certbot（宝塔面板 cron 调度）
- **证书位置**：`/etc/letsencrypt/live/tsntec.com/` → 同步到 `/www/server/panel/vhost/cert/tsntec.com/`
- **覆盖域名**：`tsntec.com`, `www.tsntec.com`, `forum.tsntec.com`
- **自动续期**：系统 cron 每周一 03:30 执行 certbot renew

## 数据库 baseURL

`t_sys_config.baseURL` = `https://forum.tsntec.com`

该值在数据库持久化，Docker 重建不影响。若重置数据库，需在管理后台重新设置。

## 注意事项

- 本 Nginx 配置由宝塔面板管理，文件位于 `/www/server/panel/vhost/nginx/forum.tsntec.com.conf`
- 重新部署 Docker 容器 **不会** 影响 Nginx 配置和证书
- 证书续期依赖 certbot，不可卸载 `/www/server/panel/pyenv/bin/certbot`
- `.well-known/acme-challenge/` 路径在 HTTP 端豁免了 301 重定向，用于证书续期验证
