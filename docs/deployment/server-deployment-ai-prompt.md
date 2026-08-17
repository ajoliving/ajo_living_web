# 服务器部署指引

> 本文档及本仓库所有 .md 文件中均不得使用 emoji。

---

## 一、服务器架构

| 服务器 | IP | 角色 | SSH |
| --- | --- | --- | --- |
| 网关 | `47.83.21.100` | frps + nginx(SSL) + acme.sh | `ssh admin@47.83.21.100` |
| 应用 | `47.239.117.108` | 业务服务 + 数据库 + frpc | `ssh admin@47.239.117.108` |

### 请求链路

```
用户 → DNS(指向 47.83.21.100) → nginx(:443 SSL 终止) → frps(:80) → frpc → 应用服务器本地服务
```

所有子域名 DNS A 记录指向 `47.83.21.100`，由 frps 按 Host 分发。

### 已部署服务

| 域名 | 服务器 | 端口 | 方式 | SSL |
| --- | --- | --- | --- | --- |
| `easy.payment.skylinedances.com` | 47.239.117.108 | :20034 | Go + Supervisor | SAN(含 pos.web) |
| `pos.web.skylinedances.com` | 47.239.117.108 | :20035 | nginx 静态 | 同上 |
| `pos.ismart.skylinedances.com` | 47.83.21.100 | :28889 | Go relay + Supervisor `ismart_pos_http_relay_v2` | 独立 |
| `iboard.skylinedances.com` | 47.239.117.108 | :9002 | Docker(前端) | 独立 |
| `iboard.service.skylinedances.com` | 47.239.117.108 | :10031 | Docker(Go) | 独立 |
| `icctv.skylinedances.com` | 47.239.117.108 | :32002 | Docker(前端) | SAN(含 icctv.service) |
| `icctv.service.skylinedances.com` | 47.239.117.108 | :32001 | Docker(Go) | 同上 |
| `pdf.maker.skylinedances.com` | 47.239.117.108 | :12101 | Docker(Node) | 独立 |
| `apis.hk.skylinedances.com` | 47.239.117.108 | :20036 | nginx 静态 | 独立 |
| `svavo.smart.databoard.skylinedances.com` | 47.239.117.108 | :20038 | nginx 静态 | 独立 |
| `svavo.smart.databoard.service.skylinedances.com` | 47.239.117.108 | :20037 | Go + Supervisor | 独立 |
| `intercom.skylinedances.com` | 47.239.117.108 | :20040 | nginx 靜態 + `/api/` 反代 | 獨立 |
| `intercom.api.skylinedances.com` | 47.83.21.100 | 待補齊 frpc | gateway nginx SSL 已配置 | 獨立 |
| `ajoliving.skylinedances.com` | 47.239.117.108 | :20041 | nginx 静态 | SAN(含 ajoliving.server) |
| `ajoliving.server.skylinedances.com` | 47.239.117.108 | :20042 | Go + Supervisor | 同上 |
| `test.ajoliving.skylinedances.com` | 47.239.117.108 | :20047 → `/api/` :20048 | nginx 靜態 + Go + Supervisor | 獨立（ACME DNS） |
| `good.price.skylinedances.com` | 47.239.117.108 | :20045 → `/api/` :20044 | nginx 静态 + `/api/` 反代 | 独立 |
| `ticket.skylinedances.com` | 47.239.117.108 | :20046 | Docker Compose（前端 + API + SQLite） | 独立（ACME DNS） |


---

## 二、操作规则

1. **先读后改**：改配置前先 SSH 读原文件。
2. **端口冲突检查**：`ss -tlnp | grep <port>`。
3. **数据库端口仅绑 127.0.0.1**，严禁 0.0.0.0。
4. **改完先验证**：`nginx -t` 或 `supervisorctl reread`，确认再 reload。
5. **SSL 证书操作按验证方式执行**：`certbot --standalone` 需先停 `frps` 释放 80 端口；ACME DNS 验证不需停止 `frps`，应优先采用以避免中断现有 HTTP 服务。

### OrangePi 项目保护规则

`icctv_orangepi_auth_service` 已部署在各台 OrangePi 上，属于设备端运行项目。除非用户明确要求并确认设备端发布方案，不得修改其代码、配置、Docker 文件、镜像或运行中的服务，也不得因为服务端需求自动重新部署 OrangePi。涉及 ICCTV 功能时，优先只修改 ICCTV 服务端、管理端或公开接口；任何设备端变更必须先单独确认影响范围和发布步骤。

### 关键文件路径

**网关 47.83.21.100：**

| 路径 | 用途 |
| --- | --- |
| `/home/admin/frp/frps.toml` | frps 配置 |
| `/etc/supervisor/conf.d/*.conf` | Supervisor 服务配置 |
| `/etc/nginx/sites-enabled/*.conf` | HTTPS 反代站点 |
| `/etc/nginx/snippets/proxy_params` | 通用反代头（Host, X-Real-IP, X-Forwarded-For, X-Forwarded-Proto） |
| `/etc/letsencrypt/live/*/fullchain.pem` | SSL 证书 |
| `/home/admin/.acme.sh/ticket.skylinedances.com_ecc/` | `ticket.skylinedances.com` ACME DNS 证书及私钥 |

**应用 47.239.117.108：**

| 路径 | 用途 |
| --- | --- |
| `/home/admin/frp/frpc.toml` | frpc 配置（所有 HTTP 代理定义在这里） |
| `/home/admin/frp/frpc` | frpc 二进制 |
| `/etc/supervisor/conf.d/*.conf` | frpc + Go 业务服务 |
| `/etc/nginx/sites-enabled/*` | nginx 本地站点（前端 + API 反代） |
| `/home/admin/<project>/` | 业务服务目录 |
| `/home/admin/svavo_smart_databoard/` | svavo 项目目录（后端、前端 dist、.env、compose） |
| `/home/admin/good_price_databoard/` | good-price 项目目录（后端、前端 dist、.env、PostgreSQL compose） |
| `/home/admin/client-ticket-board/` | Client Ticket Board 项目目录（Docker Compose、`.env`、SQLite volume） |

### POS iSmart Relay 現況與部署

`pos.ismart.skylinedances.com` 目前部署在網關伺服器 `47.83.21.100`，不要部署到應用伺服器。

目前請求鏈路：

```text
nginx `/etc/nginx/sites-enabled/pos_ismart_ssl.conf`
→ 127.0.0.1:80 frps
→ `/home/admin/ismart_pos/frp/frpc.ini`
→ localPort 28889
→ `/home/admin/ismart_pos/ismart_pos_relay/ismart_pos_http_relay`
```

目前生效服務：

| 項目 | 值 |
| --- | --- |
| Supervisor | `ismart_pos_http_relay_v2` |
| 服務目錄 | `/home/admin/ismart_pos/ismart_pos_relay` |
| 二進制 | `/home/admin/ismart_pos/ismart_pos_relay/ismart_pos_http_relay` |
| 設定檔 | `/home/admin/ismart_pos/ismart_pos_relay/.conf` |
| 監聽端口 | `0.0.0.0:28889` |
| 日誌 | `/home/admin/ismart_pos/ismart_pos_relay/http_service.log`、`/home/admin/ismart_pos/ismart_pos_relay/supervisor.log` |

歷史服務 `ismart_pos_http_relay` 仍在 `28888`，但目前 public domain 不指向它。除非先確認 frpc/nginx 映射已變更，不要把新版本部署到 v1 目錄。

部署方式：

```bash
# 1. 本地構建
cd /Users/yangliu/Documents/Code/hk/pos/ismart_pos_relay
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/ismart_pos_http_relay

# 2. 先備份線上 v2 二進制
ssh admin@47.83.21.100 'cd /home/admin/ismart_pos/ismart_pos_relay && cp ismart_pos_http_relay "ismart_pos_http_relay.bak_$(date +%Y%m%d_%H%M%S)"'

# 3. 上傳、替換、重啟
scp /tmp/ismart_pos_http_relay admin@47.83.21.100:/home/admin/ismart_pos/ismart_pos_relay/ismart_pos_http_relay.__new
ssh admin@47.83.21.100 'cd /home/admin/ismart_pos/ismart_pos_relay && chmod +x ismart_pos_http_relay.__new && mv ismart_pos_http_relay.__new ismart_pos_http_relay && sudo supervisorctl restart ismart_pos_http_relay_v2'
```

驗證方式：

```bash
ssh admin@47.83.21.100 "sudo supervisorctl status ismart_pos_http_relay_v2"
curl -i https://pos.ismart.skylinedances.com/api/poslogin
curl -i https://pos.ismart.skylinedances.com/api/v1/integration/auth/login/
```

POS relay 支援以下可選設定；可放在 Supervisor 環境變數，也可追加到 `/home/admin/ismart_pos/ismart_pos_relay/.conf`，未設定時使用生產預設值：

| 變數 | 用途 | 預設 |
| --- | --- | --- |
| `ISMART_OLD_API_BASE_URL` | 舊 AWS iSmart POS API base URL | `https://uqf0jqfm77.execute-api.ap-east-1.amazonaws.com/prod/v1` |
| `ISMART_INTEGRATION_BASE_URL` | 新 iSmart integration API host | `https://ismart.ajoliving.com` |

---

## 三、部署场景

### 场景 A：Go 后端 + Supervisor

> 范例：`easy_payment_http_service`

```bash
# 1. 本地交叉编译
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o my_service

# 2. 上传到应用服务器
ssh admin@47.239.117.108 "mkdir -p /home/admin/my_project/tmp"
scp my_service admin@47.239.117.108:/home/admin/my_project/

# 3. 创建配置（二选一）
#    .env 方式：
scp .env admin@47.239.117.108:/home/admin/my_project/
#    config.yaml 方式（如 ilock）：
scp config.yaml admin@47.239.117.108:/home/admin/my_project/

# 4. 创建 run.sh
ssh admin@47.239.117.108 'cat > /home/admin/my_project/run.sh << '\''EOF'\''
#!/bin/bash
cd /home/admin/my_project
set -a; source .env; set +a    # .env 方式用这行，config.yaml 方式删掉
exec ./my_service               # exec 让 Supervisor 直接管理进程
EOF
chmod +x /home/admin/my_project/run.sh'

# 5. 创建 Supervisor 配置
ssh admin@47.239.117.108 'sudo tee /etc/supervisor/conf.d/my_project.conf << EOF
[program:my_project]
command=/home/admin/my_project/run.sh
directory=/home/admin/my_project
user=admin
autostart=true
autorestart=true
startsecs=3
startretries=10
stdout_logfile=/home/admin/my_project/tmp/supervisor-stdout.log
stdout_logfile_maxbytes=10MB
stdout_logfile_backups=5
stderr_logfile=/home/admin/my_project/tmp/supervisor-stderr.log
stderr_logfile_maxbytes=10MB
stderr_logfile_backups=5
stopwaitsecs=10
EOF'

# 6. 启动
ssh admin@47.239.117.108 "sudo supervisorctl reread && sudo supervisorctl update"
```

### 场景 B：Docker Compose 全栈

> 范例：`icctv`、`pdf-maker`

```bash
# 1. 上传配置
ssh admin@47.239.117.108 "mkdir -p /home/admin/my_project"
scp docker-compose.yml .env admin@47.239.117.108:/home/admin/my_project/

# 2. docker-compose.yml 规范（注意: 数据库端口绑 127.0.0.1）
# 顶层必须设置唯一 name，避免多个项目目录同名 db 时互相接管容器：
# name: my_project
#
# services:
#   mysql:
#     ports:
#       - "127.0.0.1:<port>:3306"    # 严禁 0.0.0.0
#   app:
#     ports:
#       - "0.0.0.0:<port>:8080"      # 业务端口按需

# 3. 启动
ssh admin@47.239.117.108 "cd /home/admin/my_project && sudo docker compose up -d"

# 4. 验证
ssh admin@47.239.117.108 "sudo docker compose -f /home/admin/my_project/docker-compose.yml ps"
```

### 场景 C：Docker 前端（多阶段构建）

> 范例：`iboard_web_admin`、`icctv_web_admin`

```bash
# 1. Dockerfile（多阶段：node 构建 → nginx 托管）
# FROM node:lts-alpine AS build-stage
# WORKDIR /app
# COPY . .
# RUN yarn && yarn build
#
# FROM nginx:stable-alpine AS production-stage
# COPY --from=build-stage /app/dist /usr/share/nginx/html
# COPY nginx/default.conf /etc/nginx/conf.d/default.conf
# EXPOSE 80

# 2. 内置 nginx/default.conf
# server {
#     listen 80;
#     root /usr/share/nginx/html;
#     location /api/ {
#         proxy_pass http://host.docker.internal:<backend_port>/api/;
#         proxy_set_header Host $host;
#         proxy_set_header X-Real-IP $remote_addr;
#         proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
#     }
#     location / { try_files $uri $uri/ /index.html; }
# }

# 3. 构建、导出、上传
docker buildx build --platform linux/amd64 -t my_project_web .
docker save my_project_web | gzip > my_project_web.tar.gz
scp my_project_web.tar.gz admin@47.239.117.108:/tmp/
ssh admin@47.239.117.108 "sudo docker load < /tmp/my_project_web.tar.gz"

# 4. 运行（前端端口通常绑 127.0.0.1，外部走 frpc）
ssh admin@47.239.117.108 "sudo docker run -d \
  --name my_project_web \
  --restart unless-stopped \
  -p 127.0.0.1:<port>:80 \
  --add-host=host.docker.internal:host-gateway \
  my_project_web"
```

### 场景 D：静态前端 + nginx（无 Docker）

> 范例：`pos_web`、`hk-dashboard`

```bash
# 1. 本地构建
cd my-frontend && yarn build

# 2. 上传（注意: 用 dist/* 避免嵌套）
ssh admin@47.239.117.108 "mkdir -p /home/admin/my_frontend/dist"
scp -r dist/* admin@47.239.117.108:/home/admin/my_frontend/dist/

# 3. 创建 nginx 站点
ssh admin@47.239.117.108 'sudo tee /etc/nginx/sites-available/my_frontend << '\''EOF'\''
server {
    listen <port>;
    server_name _;
    root /home/admin/my_frontend/dist;
    index index.html;

    location /api/ {
        proxy_pass http://127.0.0.1:<backend_port>/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
    location / {
        try_files $uri $uri/ /index.html;
    }
}
EOF'

# 4. 启用并 reload
ssh admin@47.239.117.108 "sudo ln -sf /etc/nginx/sites-available/my_frontend /etc/nginx/sites-enabled/ && sudo nginx -t && sudo systemctl reload nginx"
```

### 场景 E：旧地址迁移反代

> 已迁移到新域名，但旧 IP:Port 还有客户端在用。

```bash
# 例：旧 http://39.108.49.167:10031 → 新 https://iboard.service.skylinedances.com
ssh root@39.108.49.167 'sudo tee /etc/nginx/sites-available/proxy.conf << '\''EOF'\''
server {
    listen <旧端口>;
    server_name _;
    location / {
        proxy_pass https://<新域名>;
        proxy_set_header Host <新域名>;
        proxy_ssl_server_name on;
    }
}
EOF'

ssh root@39.108.49.167 "sudo ln -sf /etc/nginx/sites-available/proxy.conf /etc/nginx/sites-enabled/ && sudo nginx -t && sudo systemctl reload nginx"
```

---

## 四、通用上线步骤

> 业务服务部署完成后，按顺序执行。

### 1. frpc 添加代理（应用服务器）

```bash
# 先读取当前配置
ssh admin@47.239.117.108 "cat /home/admin/frp/frpc.toml"

# 追加新代理（不覆盖原有内容）
ssh admin@47.239.117.108 'cat >> /home/admin/frp/frpc.toml << EOF

[[proxies]]
name = "my-project"
type = "http"
localIP = "127.0.0.1"
localPort = <local_port>
customDomains = ["my-project.skylinedances.com"]
EOF'

# 重启 frpc
ssh admin@47.239.117.108 "sudo supervisorctl restart frpc"
```

### 2. DNS 解析（需人工操作）

阿里云控制台添加 A 记录：`my-project.skylinedances.com` → `47.83.21.100`

### 3. UFW 防火墙（按需）

仅当端口绑定 `0.0.0.0` 时才需要开放。绑 `127.0.0.1` 的服务走 frpc 穿透，不需要开。

```bash
ssh admin@47.239.117.108 "sudo ufw allow <port>/tcp"
```

### 4. SSL 证书（网关服务器）

```bash
# 注意: 必须先停 frps 释放 80 端口！
ssh admin@47.83.21.100 "sudo supervisorctl stop frps"

# 申请证书（SAN 多域名用多个 -d）
ssh admin@47.83.21.100 "sudo certbot certonly --standalone -d my-project.skylinedances.com"

# 恢复服务
ssh admin@47.83.21.100 "sudo supervisorctl start frps"
```

如遇超长子域名导致 nginx 报错 `could not build server_names_hash`，在 `/etc/nginx/nginx.conf` 的 `http {}` 中显式设置：

```nginx
server_names_hash_bucket_size 128;
```

### 5. nginx HTTPS 反代（网关服务器）

```bash
ssh admin@47.83.21.100 'sudo tee /etc/nginx/sites-available/my_project_ssl.conf << '\''EOF'\''
server {
    listen 443 ssl;
    server_name my-project.skylinedances.com;
    ssl_certificate /etc/letsencrypt/live/my-project.skylinedances.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/my-project.skylinedances.com/privkey.pem;
    include snippets/proxy_params;
    location / {
        proxy_pass http://127.0.0.1:80;
    }
}
EOF'

ssh admin@47.83.21.100 "sudo ln -sf /etc/nginx/sites-available/my_project_ssl.conf /etc/nginx/sites-enabled/ && sudo nginx -t && sudo systemctl reload nginx"
```

### 6. 验证

```bash
curl -s -o /dev/null -w "%{http_code}" https://my-project.skylinedances.com/
```

---

## 五、端口分配表

> 新服务分配端口前先查阅，避免冲突。建议范围：20037-20099（Go）、32100-32999（Docker）

| 端口 | 服务器 | 服务 | 绑定 |
| --- | --- | --- | --- |
| 22 | 两台 | SSH | 0.0.0.0 |
| 80 | 47.83.21.100 | frps vhostHTTP | 0.0.0.0 |
| 443 | 47.83.21.100 | nginx SSL | 0.0.0.0 |
| 7000 | 47.83.21.100 | frps bindPort | 0.0.0.0 |
| 8443 | 47.83.21.100 | frps vhostHTTPS | 0.0.0.0 |
| 13000 | 47.83.21.100 | headcounter | 0.0.0.0 |
| 13001 | 47.83.21.100 | headcounter 前端 | 127.0.0.1 |
| 13002 | 47.83.21.100 | headcounter API | 127.0.0.1 |
| 20034 | 47.239.117.108 | Go 支付 API | 0.0.0.0 |
| 20035 | 47.239.117.108 | pos_web 前端 | 0.0.0.0 |
| 20036 | 47.239.117.108 | hk-dashboard | 0.0.0.0 |
| 20037 | 47.239.117.108 | svavo 后端 API | 127.0.0.1 |
| 20038 | 47.239.117.108 | svavo 前端 nginx | 127.0.0.1 |
| 20039 | 47.239.117.108 | intercom 後端 API | 0.0.0.0 |
| 20040 | 47.239.117.108 | intercom 前端 nginx | 0.0.0.0 |
| 20041 | 47.239.117.108 | ajoliving 前端 nginx | 127.0.0.1 |
| 20042 | 47.239.117.108 | ajoliving 后端 API | 127.0.0.1 |
| 20043 | 47.239.117.108 | hk-dashboard border API | 127.0.0.1 |
| 20046 | 47.239.117.108 | Client Ticket Board 前端 + `/api/` | 127.0.0.1 |
| 20044 | 47.239.117.108 | good-price 后端 API | 127.0.0.1 |
| 20045 | 47.239.117.108 | good-price 前端 nginx | 127.0.0.1 |
| 20047 | 47.239.117.108 | ajoliving 測試前端 nginx + `/api/` | 127.0.0.1 |
| 20048 | 47.239.117.108 | ajoliving 測試後端 API | 不經 frpc 或 UFW 公開 |
| 9002 | 47.239.117.108 | iboard 前端 | 0.0.0.0 |
| 10031 | 47.239.117.108 | iboard 后端 | 0.0.0.0 |
| 32001 | 47.239.117.108 | icctv 后端 | 0.0.0.0 |
| 32002 | 47.239.117.108 | icctv 前端 | 127.0.0.1 |
| 12101 | 47.239.117.108 | pdf-maker | 127.0.0.1 |
| 35432 | 47.239.117.108 | svavo PostgreSQL(Docker) | 127.0.0.1 |
| 3309 | 47.239.117.108 | MySQL(iboard) | 127.0.0.1 |
| 3311 | 47.239.117.108 | MySQL(pos_web) | 127.0.0.1 |
| 3312 | 47.239.117.108 | MySQL(icctv) | 127.0.0.1 |
| 6381 | 47.239.117.108 | Redis(iboard) | 127.0.0.1 |
| 6382 | 47.239.117.108 | ajoliving Redis(Docker) | 127.0.0.1 |
| 6383 | 47.239.117.108 | ajoliving 測試 Redis(Docker) | 127.0.0.1 |
| 45432 | 47.239.117.108 | ajoliving PostgreSQL(Docker) | 127.0.0.1 |
| 45433 | 47.239.117.108 | ajoliving 測試 PostgreSQL(Docker) | 127.0.0.1 |
| 55432 | 47.239.117.108 | good-price PostgreSQL(Docker) | 127.0.0.1 |

---

## 八、SVAVO 项目固定配置（已落地）

### 1. 应用服务器 47.239.117.108

- 项目目录：`/home/admin/svavo_smart_databoard`
- 后端 Supervisor：`/etc/supervisor/conf.d/svavo_smart_databoard.conf`
- 后端本地地址：`127.0.0.1:20037`
- 前端 nginx 站点：`/etc/nginx/sites-available/svavo_frontend`
- 前端本地地址：`127.0.0.1:20038`
- 数据库容器：`svavo-postgres`
- 数据库端口：`127.0.0.1:35432`

### 2. frpc 代理（应用服务器）

`/home/admin/frp/frpc.toml` 已追加：

- `svavo.smart.databoard.skylinedances.com -> 127.0.0.1:20038`
- `svavo.smart.databoard.service.skylinedances.com -> 127.0.0.1:20037`

### 3. 网关服务器 47.83.21.100

- SSL 站点文件：
    - `/etc/nginx/sites-available/svavo.web.ssl.nginx.conf`
    - `/etc/nginx/sites-available/svavo.api.ssl.nginx.conf`
- 证书路径：
    - `/etc/letsencrypt/live/svavo.smart.databoard.skylinedances.com/`
    - `/etc/letsencrypt/live/svavo.smart.databoard.service.skylinedances.com/`

### 4. 验收命令

```bash
curl -I https://svavo.smart.databoard.skylinedances.com/
curl -I https://svavo.smart.databoard.service.skylinedances.com/api/health
ssh admin@47.239.117.108 "sudo supervisorctl status svavo_smart_databoard frpc"
ssh admin@47.239.117.108 "sudo docker ps --filter name=svavo-postgres"
```

---

## 九、INTERCOM 專案固定配置（部分已落地）

### 1. 應用伺服器 47.239.117.108

- 前端目錄：`/home/admin/intercom_web_admin/dist`
- 前端 nginx 站點：`/etc/nginx/sites-available/intercom_web`
- 前端本地地址：`127.0.0.1:20040`
- 後端目錄：`/home/admin/intercom_http_service`
- 後端 Supervisor：`/etc/supervisor/conf.d/intercom_http_service.conf`
- 後端監聽地址：`:20039`

### 2. frpc 代理（應用伺服器）

`/home/admin/frp/frpc.toml` 已追加：

- `intercom.skylinedances.com -> 127.0.0.1:20040`

目前未在應用伺服器 `frpc.toml` 看到 `intercom.api.skylinedances.com` 對應代理。API 訪問目前應優先走 `https://intercom.skylinedances.com/api/`，若要啟用獨立 API 域名，需要先確認後端路由前綴與網關轉發方式，再補齊 frpc 或 nginx 配置。

### 3. 網關伺服器 47.83.21.100

- SSL 站點檔案：`/etc/nginx/sites-available/intercom_ssl.conf`
- 已配置 HTTPS 域名：
    - `intercom.skylinedances.com`
    - `intercom.api.skylinedances.com`
- 證書路徑：
    - `/etc/letsencrypt/live/intercom.skylinedances.com/`
    - `/etc/letsencrypt/live/intercom.api.skylinedances.com/`
- 證書有效期：至 `2026-07-30`

### 4. 驗收命令

```bash
curl -I https://intercom.skylinedances.com/
curl -I https://intercom.skylinedances.com/api/
curl -I https://intercom.api.skylinedances.com/
ssh admin@47.239.117.108 "sudo supervisorctl status intercom_http_service frpc"
ssh admin@47.239.117.108 "cat /home/admin/frp/frpc.toml | grep -A 5 'name = \"intercom_web\""
```

---

## 十、AJOLIVING 项目固定配置（已落地）

### 1. 应用服务器 47.239.117.108

- 项目目录：`/home/admin/ajoliving`
- 后端 Supervisor：`/etc/supervisor/conf.d/ajoliving_server.conf`
- 后端本地地址：`127.0.0.1:20042`
- 前端 nginx 站点：`/etc/nginx/sites-available/ajoliving_web`
- 前端本地地址：`127.0.0.1:20041`
- 数据库容器：`ajoliving_postgres`
- 数据库端口：`127.0.0.1:45432`
- Redis 容器：`ajoliving_redis`
- Redis 端口：`127.0.0.1:6382`

### 2. frpc 代理（应用服务器）

`/home/admin/frp/frpc.toml` 已追加：

- `ajoliving.skylinedances.com -> 127.0.0.1:20041`（前端）
- `ajoliving.server.skylinedances.com -> 127.0.0.1:20042`（API）

### 3. 目录结构（应用服务器）

```
/home/admin/ajoliving/
├── server/                    # Go 后端
│   ├── ajoliving_server       # 可执行文件
│   ├── .env                   # 环境变量（生产配置）
│   ├── run.sh                 # Supervisor 启动脚本
│   └── tmp/                   # 日志目录
├── web/                       # Vue 3 前端
│   ├── dist/                  # 编译产物（nginx 托管）
│   └── public/                # 静态资源
└── db/                        # 資料庫與快取容器配置
    ├── docker-compose.yml     # PostgreSQL 16 + Redis 7
    └── backups/               # 部署前生产库备份
```

### 4. 本地更新部署命令

正常发布前端和后端：

```bash
cd /Users/yangliu/Documents/Code/ajoliving_web
./deploy-ajoliving.sh
```

默认行为：

- 会本地构建 Go 后端与 Vue 前端，并发布到服务器新的 release 目录。
- 会保留服务器 PostgreSQL volume，不会删除、重建或覆盖生产数据。
- 会在后端重启前备份服务器生产库到 `/home/admin/ajoliving/db/backups/ajoliving_<release_id>.dump`。
- 后端启动时会执行 GORM `AutoMigrate`，新增字段会自动补齐；正常情况下不会删除已有表或已有字段。
- `.env` 默认保留服务器现有版本；只有显式设置 `SYNC_ENV=1` 才会用本地 `.env` 覆盖服务器 `.env`。
- 會以 Docker Compose 啟動 `ajoliving_redis`，只綁定 `127.0.0.1:6382`；Redis 僅保存可重建快取，不配置持久化 volume。
- 每次部署的话，我想你能先ssh然后能将服务的配置搞懂后再部署，部署的话尽量奥卡姆剃刀原理，不要添加到了无关的服务或者文件啥的，要简单些尽量

2026-08-05 已部署 release `20260804171519-contact-final`：

- 公開樓盤詳情已分離發布者身份與逐樓盤聯絡資料，業主帳戶姓名及代理帳戶 WeChat 不再作為樓盤聯絡資料；站內訊息仍連接發布者帳戶。
- 發布前建立 PostgreSQL 備份，保留生產 `.env`，未修改 FRP、SSL 或 OSS CORS。
- 前端、API HTTPS、Supervisor、PostgreSQL、Redis 及真實公開樓盤頁均通過驗收；公開業主盤未解鎖時不再顯示帳戶姓名。

2026-08-07 已部署 release `20260807082205-94112`：

- 測試與生產已同步發布目前工作區版本，會員從「我的樓盤」編輯既有樓盤時沿用原有放盤類別、廣告等級、樓盤、圖片及逐樓盤聯絡資料。
- 發布前建立並校驗 PostgreSQL 備份 `/home/admin/ajoliving/db/backups/ajoliving_20260807082205-94112.dump`；保留生產 `.env`，未恢復資料庫，未修改 FRP、SSL 或 OSS CORS。
- 生產前端、API、Supervisor、PostgreSQL、Redis 及真實公開樓盤列表均通過驗收；測試站與隔離服務在生產發布後維持正常。

2026-08-07 已部署 release `20260807083705-3047`：

- 樓盤主詳情與相似樓盤推薦改為獨立失敗邊界，相似樓盤讀取失敗不再遮蔽已成功讀取的主樓盤資料。
- 發布前建立並校驗 PostgreSQL 備份 `/home/admin/ajoliving/db/backups/ajoliving_20260807083705-3047.dump`；保留生產 `.env`，未恢復資料庫，未修改網絡、SSL 或 OSS CORS。
- 生產與測試的樓盤頁及同源 API 均返回 `200`；Supervisor、PostgreSQL、Redis 與真實瀏覽器樓盤詳情驗收通過。

2026-08-10 已部署測試 release `20260810042405-test-48684` 與生產 release `20260810043333-53744`：

- Good Price 已回填獨立商品副標題，AJO 商品卡片統一顯示「品牌 + 商品名稱」及獨立副標題。
- 測試與生產同源 API、桌面及移動端卡片、Supervisor、PostgreSQL、Redis 與 HTTPS 均通過驗收。
- 生產發布前建立並校驗 PostgreSQL 備份 `/home/admin/ajoliving/db/backups/ajoliving_20260810043333-53744.dump`；保留生產 `.env`，未恢復資料庫，未修改 FRP、SSL 或 OSS CORS。

### 4.1 iSmart integration 環境變數

AJO 後端同時保留舊 external app API 與新 integration API，生產 `.env` 至少保持以下設定：

```bash
ISMART_EXTERNAL_APP_BASE_URL=https://ismart.ajoliving.com
ISMART_EXTERNAL_APP_API_BASE_URL=https://ismart.ajoliving.com/api/v1/external
ISMART_INTEGRATION_API_BASE_URL=https://ismart.ajoliving.com/api/v1/integration
ISMART_SERVICE_CASE_API_BASE_URL=https://clouddev.ismart.ajoliving.com/api/v1/integration
REDIS_ENABLED=true
REDIS_ADDR=127.0.0.1:6382
REDIS_PASSWORD=
REDIS_DB=0
ISMART_BUILDING_CACHE_TTL=5m
POS_DIRECTORY_CACHE_TTL=5m
```

`/api/v1/me/ismart/building-comments` 與 `/api/v1/me/ismart/service-cases...` 暫時只使用 `ISMART_SERVICE_CASE_API_BASE_URL` 測試環境，失敗時不得回退生產接口。其餘 `/api/v1/me/ismart/...` 會員態接口繼續使用 `ISMART_INTEGRATION_API_BASE_URL`；當新路徑缺失時，已記錄的讀取類接口可回退到舊路徑。不要把 iSmart 原始無認證寫入口直接暴露給前端。

POS 大廈與單位目錄只快取共用唯讀資料，分別使用 `ajo:pos:buildings:v1` 與 `ajo:pos:building-units:v1:<building_id>`；會員權限、綁定狀態、目前物業及 relay token 不得寫入 Redis。會員接口必須先即時校驗本地可見範圍，再讀取共用目錄並過濾；Redis 故障時回源 POS。

### 4.2 Redis 生產快取狀態（已驗證）

2026-07-16 已在應用伺服器完成實際請求驗證：

- `ajoliving_redis` 狀態為 `healthy`，`redis-cli ping` 返回 `PONG`。
- Redis 只綁定 `127.0.0.1:6382`，不經 frpc、nginx 或公網提供服務。
- iSmart 大廈資料使用 `ajo:ismart:building-info:v1:<building_id>` key，TTL 為 300 秒。
- 首次會員大廈資料請求會在 Redis 未命中後回源 iSmart，成功後建立快取。
- 相同大廈的第二次請求會增加 `keyspace_hits`，不增加 `keyspace_misses`；TTL 繼續倒數而不重設，表示實際讀取既有快取。
- 兩次請求的業務 `data` 一致；共享快取不保存 `building_options`、`selected_building_id` 或其他會員個人欄位。
- Redis 故障或快取過期時，後端會直接回源 iSmart；Redis 重新啟動後快取遺失屬正常行為，不影響 PostgreSQL 或會員主資料。

2026-07-17 已隨 release `20260717094836-3321` 驗證 POS 共用目錄快取：

- 生產 `.env` 已設定 `POS_DIRECTORY_CACHE_TTL=5m`。
- 公共大廈目錄建立 `ajo:pos:buildings:v1`，大廈 `0999900` 單位目錄建立 `ajo:pos:building-units:v1:0999900`。
- 首次回源請求約為 `1.80s` 與 `1.78s`；相同請求命中 Redis 後約為 `0.54ms` 與 `0.39ms`。
- 第二次請求後 `keyspace_hits` 由 `41` 增至 `45`，`keyspace_misses` 保持 `59`；兩個 key 的 TTL 同步由約 `289s` 繼續倒數至 `258s`，沒有被命中請求重設。
- 生產回應確認大廈目錄為 47 項，`0999900` 單位目錄為 36 項；快取只保存共用目錄，會員接口仍先即時校驗本地權限。

2026-07-17 已部署 release `20260717115200-3302` 並驗證統一帳戶登入：

- 本次使用 `BACKUP_DB=0`，未建立 `/home/admin/ajoliving/db/backups/ajoliving_20260717115200-3302.dump`。
- `POST /api/v1/auth/login` 已在線上提供服務，缺少必要欄位時返回 `VALIDATION_ERROR`，證明新路由已生效。
- 生產登入頁在桌面及 `393 x 852` 手機視口均只顯示一個「手提電話 / 電郵 / iSmart username」輸入框，沒有登入方式分頁或獨立電話輸入，手機頁面沒有水平溢出。
- `ajoliving_server`、PostgreSQL、Redis、前端及 API HTTPS 健康檢查均通過；本次未重配 frpc、SSL 或 OSS CORS。

生產快取驗收：

```bash
# 容器健康與連線
ssh admin@47.239.117.108 "sudo docker inspect --format '{{.State.Health.Status}}' ajoliving_redis"
ssh admin@47.239.117.108 "sudo docker exec ajoliving_redis redis-cli ping"

# 查詢快取統計與現有 iSmart 大廈 key
ssh admin@47.239.117.108 "sudo docker exec ajoliving_redis redis-cli INFO stats | grep -E '^(keyspace_hits|keyspace_misses):'"
ssh admin@47.239.117.108 "sudo docker exec ajoliving_redis redis-cli --scan --pattern 'ajo:ismart:building-info:v1:*'"
ssh admin@47.239.117.108 "sudo docker exec ajoliving_redis redis-cli --scan --pattern 'ajo:pos:*'"

# 使用有效會員 access token 請求一次後，確認 TTL 接近 300 秒
curl -H "Authorization: Bearer <access_token>" \
  "https://ajoliving.server.skylinedances.com/api/v1/me/ismart/building-info?building_id=<building_id>"
ssh admin@47.239.117.108 \
  "sudo docker exec ajoliving_redis redis-cli TTL ajo:ismart:building-info:v1:<building_id>"
```

危险操作：

```bash
RESTORE_DB=1 RESTORE_DB_CONFIRM=RESTORE_PRODUCTION_AJOLIVING ./deploy-ajoliving.sh
```

只有同时设置以上两个变量，脚本才允许把本地 dump 恢复到服务器。正常更新前后端不要设置 `RESTORE_DB=1`。

### 5. 验收命令

```bash
# 前端可访问
curl -I http://127.0.0.1:20041/

# 后端 API 可访问
curl -I http://127.0.0.1:20042/api/v1/health

# 服务运行状态
ssh admin@47.239.117.108 "sudo supervisorctl status ajoliving_server frpc"

# 数据库容器
ssh admin@47.239.117.108 "sudo docker ps --filter name=ajoliving_postgres"

# Redis 快取容器
ssh admin@47.239.117.108 "sudo docker exec ajoliving_redis redis-cli ping"

# 最近生产库备份
ssh admin@47.239.117.108 "ls -lt /home/admin/ajoliving/db/backups | head"

# frpc 代理确认
ssh admin@47.239.117.108 "cat /home/admin/frp/frpc.toml | grep -A 3 'name = \"ajoliving'"

# SSL 申请完成后验证 HTTPS
curl -I https://ajoliving.skylinedances.com/
curl -I https://ajoliving.server.skylinedances.com/api/v1/health
```

### 6. SSL 現況

- 生產前端及 API 使用 `/etc/letsencrypt/live/ajoliving.skylinedances.com/` 的同一張 SAN 證書，涵蓋 `ajoliving.skylinedances.com` 與 `ajoliving.server.skylinedances.com`。
- 2026-08-05 驗證證書有效期至 2026-11-02，兩個生產 HTTPS 入口均正常。
- 測試域名使用 AliDNS DNS 驗證簽發的獨立 ECC 證書，不需要停止 `frps`。

### 7. 測試環境固定配置（已落地）

測試環境只使用 `test.ajoliving.skylinedances.com`。瀏覽器以同源 `/api/v1` 訪問測試後端，不建立獨立測試 API 域名，不把 `20048` 加入 frpc 或 UFW 公開入口。

應用伺服器配置：

- 專案目錄：`/home/admin/ajoliving-test`
- 後端 Supervisor：`ajoliving_test_server`
- 後端端口：`20048`
- 前端 nginx 站點：`/etc/nginx/sites-available/ajoliving_test_web`
- 前端端口：`127.0.0.1:20047`
- PostgreSQL 容器：`ajoliving_test_postgres`
- PostgreSQL 資料庫及端口：`ajoliving_test`、`127.0.0.1:45433`
- PostgreSQL volume：`ajoliving_test_ajoliving_test_postgres_data`
- Redis 容器及端口：`ajoliving_test_redis`、`127.0.0.1:6383`
- Compose project：`ajoliving_test`
- frpc proxy：`ajoliving-test-web`

網關配置：

- DNS：`test.ajoliving.skylinedances.com -> 47.83.21.100`
- HTTPS nginx：`/etc/nginx/sites-available/ajoliving_test_ssl.conf`
- ACME 證書來源：`/home/admin/.acme.sh/test.ajoliving.skylinedances.com_ecc/`
- nginx 安裝證書：`/home/admin/ajoliving-test/cert/`
- 證書有效期：2026-08-05 至 2026-11-03，續期後由 acme.sh reload nginx
- 公網訪問不設 IP 白名單、Cookie gate 或 Basic Auth，普通測試網址可直接訪問。
- 網關持續回應 `X-Robots-Tag: noindex, nofollow`；該 header 只阻止搜尋引擎收錄，不構成存取控制。
- 測試站如需重新改為私有訪問，必須同步更新本節、網關 nginx、根級與部署級 `AGENTS.md` 及 `docs/PROMPT_INDEX.md`。

資料隔離與外部服務配置：

- 首次資料由生產 PostgreSQL 邏輯快照以 `pg_dump -Fc` 建立，再 restore 至獨立測試資料庫；不得複製或共用生產 Docker volume。
- restore 後的本地 password hash 保留供測試登入；外部服務憑據由測試發布時從生產 `.env` 重新繼承。
- 測試環境使用獨立 `JWT_SECRET` 與 `ENCRYPTION_KEY`。
- restore 後保留的樓盤聯絡密文必須由 `rekey-encrypted-data` 在同一交易內以生產密鑰讀取並改用測試 `ENCRYPTION_KEY` 保存；工具不輸出密鑰或明文，已使用測試密鑰的值保持不變，無法識別的歷史測試密文保持原值。
- `APP_ENV=testing`；測試站目前公開訪問，測試操作會按生產外部服務配置執行真實 OTP、郵件、OSS、支付、POS、iSmart 及其他上游請求。
- 測試部署預設不按服務或讀寫類型隔離外部接口；只有產品明確指定的接口才可改用 mock、不可用地址或獨立測試地址。
- 每次部署先保留測試 `JWT_SECRET` 與 `ENCRYPTION_KEY`，再從生產 `.env` 重新載入外部服務配置，最後覆蓋測試資料庫、Redis、公開域名及回調 URL。

部署命令：

```bash
# 正常更新測試前後端，保留測試資料庫
./deploy-ajoliving-test.sh

# 首次建立或明確重建測試資料庫
INITIALIZE_DB=1 \
INITIALIZE_DB_CONFIRM=RESTORE_AJOLIVING_TEST \
./deploy-ajoliving-test.sh
```

`SKIP_BUILD=1` 只可用於同一次已驗證建置後的傳輸重試，不得作為一般發布方式。後端二進制使用 legacy SCP 傳輸，避免目前網絡環境的 SFTP 固定窗口停滯。

測試網址可在任何瀏覽器直接開啟，不需要入口密鑰或 Cookie。網關不再保留 `/__test_access` 路由及 `gateway-access.env`。

2026-08-05 已部署 release `20260805042109-test-87911`：

- 生產快照已恢復至測試庫，使用者與樓盤記錄數和快照一致，外部登入密文清理結果為零筆殘留。
- 前端、同源 API、Supervisor、PostgreSQL、Redis、FRP、AliDNS、HTTPS 及 `noindex` 均通過驗收。
- 真實瀏覽器已確認 Vue 應用正常掛載、`zh-HK` 語系生效且沒有 console warning 或 error。
- 生產前端、API、Supervisor、PostgreSQL 與 Redis 在測試部署後仍維持正常。

2026-08-07 已部署 release `20260807081910-test-92331`：

- 使用目前工作區重新建置測試前後端，保留既有測試資料庫，未重新匯入生產快照。
- 測試前端、同源 API、Supervisor、PostgreSQL、Redis、HTTPS 與 `noindex` 均通過驗收。
- 支付、POS 外部寫入、郵件、OSS 寫入、到期 ticker 及 Good Price 通知維持關閉，生產服務在測試發布後正常。

2026-08-07 已部署 release `20260807083354-test-1100`：

- 發布樓盤詳情容錯修正，保留既有測試資料庫及外部副作用隔離設定。
- 測試樓盤頁、同源 API、Supervisor、PostgreSQL、Redis 與真實瀏覽器樓盤詳情驗收通過。
- 生產 release `20260807083705-3047` 隨後同步發布相同修正。

2026-08-07 已修復測試快照密文不相容：

- 修復前先建立 `/home/admin/ajoliving-test/db/backups/ajoliving_test_before_rekey_202608070918.dump`，未修改生產資料庫。
- 將目標樓盤的有效生產聯絡密文只讀同步至測試庫，再使用測試獨立密鑰重加密；帶原測試 Bearer Token 的樓盤詳情請求由 `500` 恢復為 `200` 並返回 `editable_contact`。
- `deploy-ajoliving-test.sh` 在日後顯式重建測試資料庫時自動執行同一重加密流程，正常保留資料的測試發布不改動現有密文。

驗收命令：

```bash
# 不帶 Cookie 直接訪問時應返回 200
curl -I https://test.ajoliving.skylinedances.com/

# 驗證同源 API
curl -I https://test.ajoliving.skylinedances.com/api/v1/health

ssh admin@47.239.117.108 \
  "sudo supervisorctl status ajoliving_test_server frpc"
ssh admin@47.239.117.108 \
  "sudo docker ps --filter name=ajoliving_test"
ssh admin@47.239.117.108 \
  "sudo docker exec ajoliving_test_redis redis-cli ping"
```

2026-08-05 已以 Cookie access gate 取代 IP 白名單：未授權與錯誤 token 均返回 `403`；正確入口設置 30 天安全 Cookie 後，首頁與 `/api/v1/health` 均返回 `200`，並已在真實瀏覽器驗證網絡切換不再依賴固定出口 IP。

2026-08-07 已按產品決策移除 Cookie access gate 及網關入口密鑰：普通首頁與 `/api/v1/health` 在不帶 Cookie 時均返回 `200`；`noindex`、獨立資料庫、外部憑據清理及生產副作用隔離維持不變。

---

## 十一、GOOD PRICE 项目固定配置（已落地）

### 1. 应用服务器 47.239.117.108

- 项目目录：`/home/admin/good_price_databoard`
- 后端 Supervisor：`/etc/supervisor/conf.d/good_price_databoard.conf`
- 后端本地地址：`127.0.0.1:20044`
- 前端 nginx 站点：`/etc/nginx/sites-available/good_price_databoard`
- 前端本地地址：`127.0.0.1:20045`
- 前端静态目录：`/home/admin/good_price_databoard/frontend-dist`
- 数据库容器：`good-price-postgres`
- 数据库端口：`127.0.0.1:55432`
- Docker Compose 项目名：`good_price_databoard`

### 2. frpc 代理（应用服务器）

`/home/admin/frp/frpc.toml` 已追加：

- `good.price.skylinedances.com -> 127.0.0.1:20045`

该项目使用同一域名承载前后端：前端 nginx 的 `/api/` 反代到 `127.0.0.1:20044`，不需要单独 API 子域名。

### 3. 网关服务器 47.83.21.100

- SSL 站点文件：`/etc/nginx/sites-available/good_price_ssl.conf`
- 证书路径：`/etc/letsencrypt/live/good.price.skylinedances.com/`
- 证书有效期：至 `2026-08-26`

### 4. 数据更新与数据卷

- 后端 `.env` 使用 `PRICEWATCH_DAILY_IMPORT_HOUR=9`，按服务器本地时间每天 09:00 自动下载当天繁体中文 CSV、导入 PostgreSQL、刷新缓存并评估邮件提醒。
- `PRICEWATCH_HISTORY_DAYS=90`，历史价格记录保留 90 天，超过窗口的数据由导入流程清理。
- 前端生产构建默认使用同源 API，即请求 `/api/...`；不要把生产包构建成 `VITE_API_BASE=http://localhost:8080`，否则线上浏览器会请求用户本机 8080 导致无数据。
- PostgreSQL 数据通过 Docker volume 持久化。正常发布前后端不要删除 volume，不要执行 `docker compose down -v`。

### 5. 目录结构（应用服务器）

```
/home/admin/good_price_databoard/
├── backend/                   # Go 后端、.env、run.sh、导入 dump
├── frontend-dist/             # React/Vite 编译产物（nginx 托管）
├── db/                        # PostgreSQL compose
│   └── docker-compose.yml     # 顶层 name: good_price_databoard
└── tmp/                       # Supervisor 日志目录
```

### 6. 验收命令

```bash
# 前端可访问
curl -I https://good.price.skylinedances.com/

# 后端 API 可访问
curl -I https://good.price.skylinedances.com/api/health

# 数据确认
curl -s https://good.price.skylinedances.com/api/summary

# 服务运行状态
ssh admin@47.239.117.108 "sudo supervisorctl status good_price_databoard frpc"

# 数据库容器
ssh admin@47.239.117.108 "sudo docker ps --filter name=good-price-postgres"

# Compose 项目名隔离确认
ssh admin@47.239.117.108 "sudo docker inspect -f '{{.Name}} {{ index .Config.Labels \"com.docker.compose.project\" }}' good-price-postgres ajoliving_postgres svavo-postgres"

# 前端生产包不能包含本地 API 地址
ssh admin@47.239.117.108 "grep -R 'localhost:8080' -n /home/admin/good_price_databoard/frontend-dist/assets || true"

# frpc 代理确认
ssh admin@47.239.117.108 "cat /home/admin/frp/frpc.toml | grep -A 5 'name = \"good-price\"'"
```

### 7. 部署注意事项

- 改动前先读 `/etc/nginx/sites-available/good_price_databoard`、`/etc/supervisor/conf.d/good_price_databoard.conf`、`/home/admin/frp/frpc.toml` 和 `/home/admin/good_price_databoard/db/docker-compose.yml`。
- 数据库端口必须保持 `127.0.0.1:55432`，不要绑定 `0.0.0.0`。
- `db/docker-compose.yml` 必须保留顶层 `name: good_price_databoard`。如果缺失，`docker compose` 会用目录名作为 project，容易与其他 `/db` 目录项目冲突。
- 静态前端更新只需同步 `frontend/dist/` 到 `/home/admin/good_price_databoard/frontend-dist/`，通常不需要重启后端或 nginx。
- 后端更新需重新上传二进制并 `sudo supervisorctl restart good_price_databoard`。

### 8. 商品目录与 AJO 图片同步（2026-08-07）

- 本次以已审核的 Good Price 工作簿和三个图片目录为来源；生产 PostgreSQL 已下架 28 个 `product_code`，并修正 6 个商品名称或品牌。
- `backend/product_catalog_overrides.go` 的覆盖规则会在每日 CSV 写入前再次执行，避免下架商品或人工修正被日常导入恢复。
- 335 张可用图片已由 `/home/admin/ajoliving/server/import-supermarket-images` 导入 AJO 生产 OSS，固定物件键前缀为 `ajo_living/supermarket/products/`，并同步建立 `media_assets` 记录；导入命令可重复执行，已有记录默认跳过。
- 三条工作簿中标记有图但本地没有文件的商品 `P000002837`、`P000002463`、`P000003230` 保持待补，不以占位图写入 OSS。
- 本次备份：`/home/admin/good_price_databoard/db/good_price_before_catalog_20260807_181842.dump`、`/home/admin/ajoliving/db/backups/ajoliving_before_supermarket_images_20260807_181842.dump`。Good Price 旧二进制保留为 `/home/admin/good_price_databoard/backend/good_price_backend.bak.20260807_catalog`。
- 测试环境的 `OSS_PROVIDER=mock` 与外部副作用隔离规则保持不变，本批图片只写入生产 OSS；需要测试图片时应使用独立测试对象存储并另行确认。
- 2026-08-10：新增 `product_subtitle` 持久化與人工商品名稱/副標題回填；`P000005327` 統一為「抗菌全效洗衣膠囊 (海洋) 30個」及副標題 `30個`。每日 CSV 導入及服務啟動回填均會保留該欄位。最終部署前備份為 `/home/admin/good_price_databoard/db/good_price_before_name_subtitle_20260810_1240.dump`，舊二進制保留為 `/home/admin/good_price_databoard/backend/good_price_backend.bak.20260810_1240`。
- 2026-08-17：每日 CSV 導入新增尾端規格自動拆分，人工審核副標題維持優先；生產回填只處理最新兩個資料日。發布前備份為 `/home/admin/good_price_databoard/db/good_price_before_units_20260817.dump`，舊二進制保留為 `/home/admin/good_price_databoard/backend/good_price_backend.bak.20260817_units`，未修改 FRP、nginx、SSL 或資料庫連接配置。

---

## 十二、Client Ticket Board 固定配置（已落地）

### 1. 应用服务器 47.239.117.108

- 项目目录：`/home/admin/client-ticket-board`
- Docker Compose 文件：`docker-compose.prod.yml`
- Compose 项目名：`client_ticket_board`
- 前端与 API 本地入口：`127.0.0.1:20046`
- API 健康检查：`http://127.0.0.1:20046/api/health`
- SQLite 数据卷：`client_ticket_board_data`
- 附件储存：复用 AJO Living OSS，物件键前缀为 `tickets/`；凭证仅保存在服务器 `.env`，不得写入前端或文档。

服务以同一域名承载前端和 API，浏览器请求 `/api/`，不设置独立 API 子域名。服务当前为公开访问，不配置登录鉴权。

### 2. frpc 代理（应用服务器）

`/home/admin/frp/frpc.toml` 已追加：

- `client-ticket-board`: `ticket.skylinedances.com -> 127.0.0.1:20046`

更新工单服务后，只重建该项目容器；不要删除 `client_ticket_board_data` volume，也不要改动其他 `frpc` 代理。

```bash
ssh admin@47.239.117.108 "cd /home/admin/client-ticket-board && sudo docker compose -f docker-compose.prod.yml up -d --build"
ssh admin@47.239.117.108 "curl -fsS http://127.0.0.1:20046/api/health"
```

### 3. 网关服务器 47.83.21.100

- HTTPS 站点文件：`/etc/nginx/sites-available/ticket.skylinedances.com`
- 启用链接：`/etc/nginx/sites-enabled/ticket.skylinedances.com`
- 反向代理目标：`http://127.0.0.1:80`（frps HTTP 虚拟主机）
- 证书：ACME DNS 签发的 ECDSA 证书
- 证书链：`/home/admin/.acme.sh/ticket.skylinedances.com_ecc/fullchain.cer`
- 私钥：`/home/admin/.acme.sh/ticket.skylinedances.com_ecc/ticket.skylinedances.com.key`
- 续期行为：`acme.sh` 续期后执行 `nginx -t && systemctl reload nginx`
- 当前证书有效期：至 `2026-10-18`

该证书使用 AliDNS 验证，申请或续期时不得停止 `frps`。不要将此域名的 HTTP 入口改为 Nginx 80 端口，因为该端口由 `frps` 持有。

### 4. 验收命令

```bash
curl -fsS https://ticket.skylinedances.com/api/health
curl -fsSI https://ticket.skylinedances.com/
ssh admin@47.239.117.108 "cd /home/admin/client-ticket-board && sudo docker compose -f docker-compose.prod.yml ps"
ssh admin@47.83.21.100 "sudo nginx -t"
```

---

## 六、常用命令

```bash
# Supervisor
sudo supervisorctl status                              # 查看状态
sudo supervisorctl reread && sudo supervisorctl update  # 加载新配置
sudo supervisorctl restart <name>                       # 重启

# Docker
sudo docker ps -a                                      # 查看容器
sudo docker compose -f /path/docker-compose.yml up -d  # 启动
sudo docker logs -f <container>                        # 日志

# Nginx
sudo nginx -t && sudo systemctl reload nginx           # 检查并重载

# SSL
sudo certbot certificates                              # 查看证书

# 端口
ss -tlnp | grep <port>                                 # 端口占用
sudo ufw allow <port>/tcp                              # 开放端口

# 日志排查
tail -f /home/admin/<project>/tmp/supervisor-stderr.log   # 应用日志
tail -f /home/admin/frp/supervisor.log                     # frpc 日志
sudo tail -f /var/log/nginx/error.log                      # nginx 日志
```

---

## 七、部署 Checklist

**应用服务器：**
- [ ] 服务本地可访问 `curl http://127.0.0.1:<port>/`
- [ ] 数据库端口绑 `127.0.0.1`
- [ ] `run.sh` 使用 `exec`
- [ ] Supervisor `autostart + autorestart`
- [ ] `frpc.toml` 已追加代理，frpc 已重启

**网关服务器：**
- [ ] DNS A 记录已添加（阿里云） `→ 47.83.21.100`
- [ ] SSL 证书已申请
- [ ] nginx HTTPS block 已配置，`nginx -t` 通过

**验证：**
- [ ] `https://<domain>/` 正常访问
- [ ] `sudo supervisorctl status` 全部 RUNNING
