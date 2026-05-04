# OmniRouter 反向代理配置

> 把 OmniRouter 部署到 `https://omnirouter.org` 需要一个反代做 HTTPS 终端 + 路由 + 安全策略。
>
> **推荐 Caddy**（自动 HTTPS，单文件配置，0 配置坑）。给出 Caddy 主推方案 + nginx 备选 + Cloudflare 前置建议。

---

## 推荐架构

```
                                                  ┌──────────────────┐
                                                  │   Cloudflare     │  ← (推荐) DDoS 防护 + CDN + WAF
                                                  │   omnirouter.org │
                                                  └────────┬─────────┘
                                                           │ HTTPS (Origin Cert)
                                                  ┌────────▼─────────┐
                                                  │   Caddy / nginx  │  ← HTTPS 终端 + 路由 + IP 限制
                                                  │   :443           │
                                                  └────────┬─────────┘
                                                           │ HTTP
                                                  ┌────────▼─────────┐
                                                  │  OmniRouter      │
                                                  │  :3000           │
                                                  └──────────────────┘
```

---

## 方案 A · Caddy（推荐）

### 安装

```bash
# Ubuntu/Debian
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install -y caddy

# 或者直接 Docker
# docker run -d --name caddy -p 80:80 -p 443:443 \
#   -v ./Caddyfile:/etc/caddy/Caddyfile \
#   -v caddy_data:/data caddy:2
```

### `/etc/caddy/Caddyfile`

```caddy
# OmniRouter — production reverse proxy
# 自动 HTTPS via Let's Encrypt（无需手动配证书）

omnirouter.org, www.omnirouter.org {
    # ---- Security headers ----
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
        X-Content-Type-Options    "nosniff"
        X-Frame-Options           "SAMEORIGIN"
        Referrer-Policy           "strict-origin-when-cross-origin"
        Permissions-Policy        "camera=(), microphone=(), geolocation=()"
        # 不暴露反代信息
        -Server
        -X-Powered-By
    }

    # ---- /metrics 锁源 IP（仅允许 Prometheus 抓取）----
    @metrics_allowed {
        path /metrics
        remote_ip 1.2.3.4 5.6.7.0/24       # ← 改成你的 Prometheus IP / VPN 网段
    }
    @metrics_blocked {
        path /metrics
        not remote_ip 1.2.3.4 5.6.7.0/24    # ← 同上
    }
    handle @metrics_blocked {
        respond "Forbidden" 403
    }

    # ---- 速率限制（防滥用）----
    # 需 caddy-rate-limit 插件，编译 Caddy 时加：
    #   xcaddy build --with github.com/mholt/caddy-ratelimit
    #
    # 配置示例（按 IP 限制）：
    # rate_limit {
    #     zone api {
    #         key {client_ip}
    #         events 600
    #         window 1m
    #     }
    # }

    # ---- WebSocket 支持（OpenAI Realtime API 用）----
    @websockets {
        header Connection *Upgrade*
        header Upgrade    websocket
    }
    handle @websockets {
        reverse_proxy localhost:3000
    }

    # ---- 主反代 ----
    reverse_proxy localhost:3000 {
        # 保留客户端真实 IP（OmniRouter 用得到）
        header_up X-Real-IP        {remote_host}
        header_up X-Forwarded-For  {remote_host}
        header_up X-Forwarded-Proto https

        # 流式响应（SSE）必须这两行 —— 否则 Claude Code 打字机效果会卡
        flush_interval -1
        transport http {
            read_timeout  600s
            write_timeout 600s
            response_header_timeout 30s
        }

        # 健康检查走我们的 /health
        health_uri      /health
        health_interval 30s
        health_timeout  5s
    }

    # ---- 日志（可选）----
    log {
        output file /var/log/caddy/omnirouter.log {
            roll_size 100mb
            roll_keep 14
        }
        format json
        # 不记探针请求避免日志噪声
        # exclude path /health /healthz /ready /readyz
    }
}

# 重定向 www → 主域（如果你只想用裸域）
# www.omnirouter.org {
#     redir https://omnirouter.org{uri} permanent
# }

# Grafana / 内部工具放子域 + 内网限制
# grafana.omnirouter.org {
#     @internal remote_ip 10.0.0.0/8
#     handle @internal {
#         reverse_proxy localhost:3001
#     }
#     handle {
#         respond "Forbidden" 403
#     }
# }
```

### 启用 + 验证

```bash
# 检查配置语法
caddy validate --config /etc/caddy/Caddyfile

# 重启
sudo systemctl reload caddy

# 验证
curl -I https://omnirouter.org/health        # 200
curl -I https://omnirouter.org/metrics       # 403（如果你不在 allow-list）
curl https://omnirouter.org/metrics --resolve "omnirouter.org:443:127.0.0.1"  # 测端到端
```

---

## 方案 B · nginx（如果你已经有 nginx 栈）

### `/etc/nginx/sites-available/omnirouter.conf`

```nginx
# Upstream
upstream omnirouter_backend {
    server 127.0.0.1:3000 max_fails=3 fail_timeout=30s;
    keepalive 64;
}

# Rate limit zone
limit_req_zone $binary_remote_addr zone=api_per_ip:10m rate=10r/s;

# 重定向 HTTP → HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name omnirouter.org www.omnirouter.org;
    return 301 https://omnirouter.org$request_uri;
}

# 主 server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name omnirouter.org;

    # ---- SSL ----
    ssl_certificate     /etc/letsencrypt/live/omnirouter.org/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/omnirouter.org/privkey.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache   shared:SSL:50m;
    ssl_session_timeout 1d;

    # ---- Security headers ----
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header X-Content-Type-Options    "nosniff" always;
    add_header X-Frame-Options           "SAMEORIGIN" always;
    add_header Referrer-Policy           "strict-origin-when-cross-origin" always;
    add_header Permissions-Policy        "camera=(), microphone=(), geolocation=()" always;
    server_tokens off;

    # ---- /metrics IP allow-list ----
    location /metrics {
        # ← 改成你的 Prometheus IP
        allow 1.2.3.4;
        allow 5.6.7.0/24;
        deny  all;
        proxy_pass http://omnirouter_backend;
    }

    # ---- WebSocket（OpenAI Realtime）----
    location /v1/realtime {
        proxy_pass http://omnirouter_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade    $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host       $host;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    # ---- 主反代 ----
    location / {
        # 速率限制
        limit_req zone=api_per_ip burst=20 nodelay;

        proxy_pass http://omnirouter_backend;
        proxy_http_version 1.1;
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header Connection        "";

        # 流式响应（SSE 必须）
        proxy_buffering        off;
        proxy_cache            off;
        proxy_request_buffering off;

        # 长超时（Claude Code 长 prompt 可能跑几分钟）
        proxy_read_timeout 600s;
        proxy_send_timeout 600s;
        proxy_connect_timeout 30s;
    }

    # ---- Logs ----
    access_log /var/log/nginx/omnirouter.access.log;
    error_log  /var/log/nginx/omnirouter.error.log warn;
}
```

### 启用

```bash
# 申请证书（用 certbot）
sudo certbot certonly --nginx -d omnirouter.org -d www.omnirouter.org

# 启用站点
sudo ln -s /etc/nginx/sites-available/omnirouter.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# 自动续签
sudo systemctl enable certbot.timer
```

---

## 方案 C · Cloudflare 前置（强烈推荐叠加 A 或 B）

### 优势
- 免费 DDoS 防护（无限带宽）
- 全球 CDN（中国海外用户加速）
- 免费 SSL（Origin Cert 用于 Cloudflare ↔ 你的服务器）
- WAF 规则（防恶意请求）
- Bot 管理（防爬虫薅羊毛）
- Analytics 看流量来源

### 操作

1. 注册 [cloudflare.com](https://cloudflare.com)，添加站点 `omnirouter.org`
2. 把域名 NS 记录改成 Cloudflare 给的（`xxx.ns.cloudflare.com`）
3. SSL/TLS → 选 **Full (strict)**（要求 Cloudflare 到你服务器也走 HTTPS）
4. SSL/TLS → Origin Server → 创建 Origin Certificate（15 年有效），下载后装到你的 Caddy/nginx
5. DNS → 添加 A 记录 `omnirouter.org → 你服务器IP`，**云朵图标点橙色**（启用代理）
6. Speed → Tiered Cache: ON
7. Caching → Configuration → Browser Cache TTL: Respect Existing Headers
8. Security → Bots → 启用 Super Bot Fight Mode
9. Security → WAF → 加规则：屏蔽常见攻击模式（SQL 注入、XSS）
10. Rules → Page Rules → 给 `/metrics` 加 "Block" 规则（Cloudflare 层就拦截）

### Cloudflare 反代你服务器后的 IP 真实性

启用 Cloudflare 代理后，所有请求源 IP 都是 Cloudflare 边缘节点。要让 OmniRouter 看到用户真实 IP：

**Caddy** 已自动处理（用 `{remote_host}` 取 `CF-Connecting-IP` header）。

**nginx** 加：
```nginx
# 信任 Cloudflare IP（自动同步用 cloudflare-ip-update 脚本）
include /etc/nginx/cloudflare-ips.conf;
real_ip_header CF-Connecting-IP;
real_ip_recursive on;
```

---

## 完整生产部署 checklist

- [ ] 域名 DNS A 记录指向服务器 IP（如启用 Cloudflare，云朵橙色）
- [ ] 服务器装好 Caddy 或 nginx + Docker + Compose
- [ ] OmniRouter 跑在 :3000（`docker compose -f docker-compose.omnirouter.yml up -d`）
- [ ] 反代配置文件（Caddyfile 或 nginx conf）就位
- [ ] 反代健康检查：`curl -I http://localhost:3000/health` → 200
- [ ] HTTPS 证书签发：`curl -I https://omnirouter.org` → 200 + valid cert
- [ ] `/metrics` 锁源 IP：`curl https://omnirouter.org/metrics` → 403（外部）/ 200（你的 Prometheus）
- [ ] WebSocket 通：测一次 OpenAI Realtime API
- [ ] 速率限制生效：`for i in {1..50}; do curl https://omnirouter.org/api/status; done` 看是否 limited
- [ ] 日志在写：`tail -f /var/log/caddy/omnirouter.log`
- [ ] 自动续签证书（Caddy 自动 / nginx 用 certbot.timer）
- [ ] HSTS 生效（curl -I 看到 `Strict-Transport-Security`）
- [ ] CSP / X-Frame-Options 生效
- [ ] Cloudflare WAF 启用（如用 CF）

---

## 故障排查速查

| 症状 | 可能原因 | 排查命令 |
|---|---|---|
| HTTPS 502 | OmniRouter 没在 :3000 / Docker 没起 | `curl -I http://localhost:3000/health`、`docker ps` |
| HTTPS 自签证书 | Let's Encrypt 限速 / DNS 没传播 | `caddy validate`、`dig omnirouter.org`、`certbot certificates` |
| SSE 流式卡顿 | 反代没 disable buffering | 检查 `proxy_buffering off` (nginx) / `flush_interval -1` (Caddy) |
| 看不到用户真实 IP | 反代没传 X-Forwarded-For | 检查 nginx `real_ip_header` / Cloudflare CF-Connecting-IP |
| 上游 API 调用 timeout 30s | 反代默认超时太短 | 改 `proxy_read_timeout 600s` |
| /metrics 公开可访问 | IP allow-list 没生效 | 检查 nginx `allow/deny` 顺序、Caddy `@metrics_blocked` |

---

## 相关文件

- [`launch-checklist.md`](../launch-checklist.md) §4 域名/DNS/HTTPS 完整流程
- [`../observability/prometheus-metrics.md`](../../observability/prometheus-metrics.md) — `/metrics` 端点设计
- [`../../docker-compose.omnirouter.yml`](../../../docker-compose.omnirouter.yml) — OmniRouter 容器栈
