# 服务器部署指引

> 本文档及本仓库所有 .md 文件中均不得使用 emoji。

---

## 一、服务器架构

| 服务器 | IP | 角色 | SSH |
| --- | --- | --- | --- |
| 网关 | `47.83.21.100` | frps + nginx(SSL) + certbot | `ssh admin@47.83.21.100` |
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
| `pos.ismart.skylinedances.com` | 47.83.21.100 | → 门店 | Go relay | 独立 |
| `iboard.skylinedances.com` | 47.239.117.108 | :9002 | Docker(前端) | 独立 |
| `iboard.service.skylinedances.com` | 47.239.117.108 | :10031 | Docker(Go) | 独立 |
| `icctv.skylinedances.com` | 47.239.117.108 | :32002 | Docker(前端) | SAN(含 icctv.service) |
| `icctv.service.skylinedances.com` | 47.239.117.108 | :32001 | Docker(Go) | 同上 |
| `pdf.maker.skylinedances.com` | 47.239.117.108 | :12101 | Docker(Node) | 独立 |
| `apis.hk.skylinedances.com` | 47.239.117.108 | :20036 | nginx 静态 | 独立 |
| `svavo.smart.databoard.skylinedances.com` | 47.239.117.108 | :20038 | nginx 静态 | 独立 |
| `svavo.smart.databoard.service.skylinedances.com` | 47.239.117.108 | :20037 | Go + Supervisor | 独立 |
| `ajoliving.skylinedances.com` | 47.239.117.108 | :20041 | nginx 静态 | 待申请 |
| `ajoliving.server.skylinedances.com` | 47.239.117.108 | :20042 | Go + Supervisor | 待申请 |


---

## 二、操作规则

1. **先读后改**：改配置前先 SSH 读原文件。
2. **端口冲突检查**：`ss -tlnp | grep <port>`。
3. **数据库端口仅绑 127.0.0.1**，严禁 0.0.0.0。
4. **改完先验证**：`nginx -t` 或 `supervisorctl reread`，确认再 reload。
5. **SSL 证书操作需停 frps**（frps 占用 80 端口，certbot standalone 也需要 80）。

### 关键文件路径

**网关 47.83.21.100：**

| 路径 | 用途 |
| --- | --- |
| `/home/admin/frp/frps.toml` | frps 配置 |
| `/etc/supervisor/conf.d/*.conf` | Supervisor 服务配置 |
| `/etc/nginx/sites-enabled/*.conf` | HTTPS 反代站点 |
| `/etc/nginx/snippets/proxy_params` | 通用反代头（Host, X-Real-IP, X-Forwarded-For, X-Forwarded-Proto） |
| `/etc/letsencrypt/live/*/fullchain.pem` | SSL 证书 |

**应用 47.239.117.108：**

| 路径 | 用途 |
| --- | --- |
| `/home/admin/frp/frpc.toml` | frpc 配置（所有 HTTP 代理定义在这里） |
| `/home/admin/frp/frpc` | frpc 二进制 |
| `/etc/supervisor/conf.d/*.conf` | frpc + Go 业务服务 |
| `/etc/nginx/sites-enabled/*` | nginx 本地站点（前端 + API 反代） |
| `/home/admin/<project>/` | 业务服务目录 |
| `/home/admin/svavo_smart_databoard/` | svavo 项目目录（后端、前端 dist、.env、compose） |

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
| 20041 | 47.239.117.108 | ajoliving 前端 nginx | 127.0.0.1 |
| 20042 | 47.239.117.108 | ajoliving 后端 API | 127.0.0.1 |
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
| 45432 | 47.239.117.108 | ajoliving PostgreSQL(Docker) | 127.0.0.1 |

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

## 九、AJOLIVING 项目固定配置（已落地）

### 1. 应用服务器 47.239.117.108

- 项目目录：`/home/admin/ajoliving`
- 后端 Supervisor：`/etc/supervisor/conf.d/ajoliving_server.conf`
- 后端本地地址：`127.0.0.1:20042`
- 前端 nginx 站点：`/etc/nginx/sites-available/ajoliving_web`
- 前端本地地址：`127.0.0.1:20041`
- 数据库容器：`ajoliving_postgres`
- 数据库端口：`127.0.0.1:45432`

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
└── db/                        # 数据库容器配置
    ├── docker-compose.yml     # PostgreSQL 16
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

# 最近生产库备份
ssh admin@47.239.117.108 "ls -lt /home/admin/ajoliving/db/backups | head"

# frpc 代理确认
ssh admin@47.239.117.108 "cat /home/admin/frp/frpc.toml | grep -A 3 'name = \"ajoliving'"

# SSL 申请完成后验证 HTTPS
curl -I https://ajoliving.skylinedances.com/
curl -I https://ajoliving.server.skylinedances.com/api/v1/health
```

### 6. 后续 SSL 配置

待网关 47.83.21.100 申请证书后，新增两个 nginx HTTPS 反代站点：

```bash
# 申请证书（需停 frps）
ssh admin@47.83.21.100 "sudo supervisorctl stop frps"
ssh admin@47.83.21.100 "sudo certbot certonly --standalone -d ajoliving.skylinedances.com -d ajoliving.server.skylinedances.com"
ssh admin@47.83.21.100 "sudo supervisorctl start frps"

# 创建 nginx HTTPS 反代（参考 SVAVO 或通用场景 V-5）
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
