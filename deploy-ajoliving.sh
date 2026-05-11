#!/bin/bash
#
# AJO Living deployment script.
# 1. Build backend and frontend assets locally.
# 2. Upload binary, env file, frontend assets, and optional database dump.
# 3. Configure remote PostgreSQL, supervisor, nginx, FRP, and optional SSL.
#
set -euo pipefail

APP_SERVER="47.239.117.108"
GATEWAY_SERVER="47.83.21.100"

PROJECT="ajoliving"
APP_DOMAIN="ajoliving.server.skylinedances.com"
WEB_DOMAIN="ajoliving.skylinedances.com"

APP_PORT="${APP_PORT:-20042}"
WEB_PORT="${WEB_PORT:-20041}"
DB_PORT="${DB_PORT:-45432}"

RESTORE_DB="${RESTORE_DB:-1}"
APPLY_NETWORK="${APPLY_NETWORK:-1}"
APPLY_SSL="${APPLY_SSL:-0}"

ROOT_DIR="$(git -C "$(dirname "$0")" rev-parse --show-toplevel 2>/dev/null || pwd)"
BACKEND_DIR="$ROOT_DIR/http_service"
FRONTEND_DIR="$ROOT_DIR/web"

ENV_FILE="${ENV_FILE:-$BACKEND_DIR/.env}"
DUMP_FILE="$ROOT_DIR/ajoliving_backup.sql"

if [ ! -f "$ENV_FILE" ]; then
  echo "ENV_FILE not found: $ENV_FILE"
  exit 1
fi

pushd "$BACKEND_DIR" >/dev/null
GO111MODULE=on go mod download
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ajoliving_server ./cmd/server
popd >/dev/null

pushd "$FRONTEND_DIR" >/dev/null
npm ci
npm run build
popd >/dev/null

if [ "$RESTORE_DB" = "1" ]; then
  if docker ps --format '{{.Names}}' | grep -q '^ajoliving_http_service_postgres$'; then
    docker exec ajoliving_http_service_postgres pg_dump -U postgres ajoliving > "$DUMP_FILE"
  else
    echo "Local postgres container not found, skip dump."
  fi
fi

scp "$BACKEND_DIR/ajoliving_server" admin@"$APP_SERVER":/tmp/
scp "$ENV_FILE" admin@"$APP_SERVER":/tmp/ajoliving.env
rsync -av --delete "$FRONTEND_DIR/dist/" admin@"$APP_SERVER":/tmp/ajoliving_web_dist/
rsync -av --delete "$FRONTEND_DIR/public/" admin@"$APP_SERVER":/tmp/ajoliving_web_public/
if [ "$RESTORE_DB" = "1" ] && [ -f "$DUMP_FILE" ]; then
  scp "$DUMP_FILE" admin@"$APP_SERVER":/tmp/ajoliving_backup.sql
fi

ssh admin@"$APP_SERVER" \
  PROJECT="$PROJECT" \
  APP_PORT="$APP_PORT" \
  WEB_PORT="$WEB_PORT" \
  DB_PORT="$DB_PORT" \
  APP_DOMAIN="$APP_DOMAIN" \
  WEB_DOMAIN="$WEB_DOMAIN" \
  RESTORE_DB="$RESTORE_DB" \
  bash -s <<'REMOTE_APP'
set -euo pipefail

APP_DIR="/home/admin/$PROJECT/server"
WEB_DIR="/home/admin/$PROJECT/web"
DB_DIR="/home/admin/$PROJECT/db"

mkdir -p "$APP_DIR/tmp" "$WEB_DIR/dist" "$WEB_DIR/public" "$DB_DIR"

if sudo supervisorctl status ajoliving_server >/dev/null 2>&1; then
  sudo supervisorctl stop ajoliving_server || true
fi

if ss -tlnp | grep -q ":$APP_PORT"; then
  echo "APP_PORT $APP_PORT already in use"; exit 1
fi
if ss -tlnp | grep -q ":$WEB_PORT"; then
  echo "WEB_PORT $WEB_PORT already in use, continue deploy."
fi

cp /tmp/ajoliving_server "$APP_DIR/ajoliving_server"
chmod +x "$APP_DIR/ajoliving_server"

cp /tmp/ajoliving.env "$APP_DIR/.env"
chmod 600 "$APP_DIR/.env"

set_env() {
  local key="$1"
  local value="$2"
  local escaped
  escaped="$(printf '%s' "$value" | sed "s/'/'\\\\''/g")"
  if grep -q "^$key=" "$APP_DIR/.env"; then
    awk -v key="$key" -v value="$escaped" '
      BEGIN { line = key "='\''" value "'\''" }
      $0 ~ "^" key "=" { $0 = line }
      { print }
    ' "$APP_DIR/.env" > "$APP_DIR/.env.tmp"
    mv "$APP_DIR/.env.tmp" "$APP_DIR/.env"
  else
    printf "%s='%s'\n" "$key" "$escaped" >> "$APP_DIR/.env"
  fi
}

set_env "APP_ENV" "production"
set_env "APP_PORT" "$APP_PORT"
set_env "APP_PUBLIC_BASE_URL" "https://$WEB_DOMAIN"
set_env "DB_DSN" "host=127.0.0.1 user=postgres password=postgres dbname=ajoliving port=$DB_PORT sslmode=disable TimeZone=Asia/Shanghai"
set_env "WEB_PUBLIC_DIR" "$WEB_DIR/public"

rsync -av --delete /tmp/ajoliving_web_dist/ "$WEB_DIR/dist/"
rsync -av --delete /tmp/ajoliving_web_public/ "$WEB_DIR/public/"

cat > "$DB_DIR/docker-compose.yml" <<YAML
services:
  postgres:
    image: postgres:16-alpine
    container_name: ajoliving_postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: ajoliving
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      TZ: Asia/Shanghai
      PGTZ: Asia/Shanghai
    ports:
      - "127.0.0.1:$DB_PORT:5432"
    volumes:
      - ajoliving_postgres_data:/var/lib/postgresql/data
volumes:
  ajoliving_postgres_data:
YAML

sudo docker compose -f "$DB_DIR/docker-compose.yml" up -d

for i in {1..30}; do
  if sudo docker exec ajoliving_postgres pg_isready -U postgres -d ajoliving >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if [ "$RESTORE_DB" = "1" ] && [ -f /tmp/ajoliving_backup.sql ]; then
  sudo docker exec -i ajoliving_postgres psql -U postgres -d ajoliving < /tmp/ajoliving_backup.sql
fi

cat > "$APP_DIR/run.sh" <<SH
#!/bin/bash
set -e
cd "$APP_DIR"
set -a; source .env; set +a
exec ./ajoliving_server
SH
chmod +x "$APP_DIR/run.sh"

sudo tee /etc/supervisor/conf.d/ajoliving_server.conf >/dev/null <<SUP
[program:ajoliving_server]
command=$APP_DIR/run.sh
directory=$APP_DIR
user=admin
autostart=true
autorestart=true
startsecs=3
startretries=10
stdout_logfile=$APP_DIR/tmp/supervisor-stdout.log
stdout_logfile_maxbytes=10MB
stdout_logfile_backups=5
stderr_logfile=$APP_DIR/tmp/supervisor-stderr.log
stderr_logfile_maxbytes=10MB
stderr_logfile_backups=5
stopwaitsecs=10
SUP

sudo supervisorctl reread
sudo supervisorctl update
sudo supervisorctl restart ajoliving_server

sudo tee /etc/nginx/sites-available/ajoliving_web >/dev/null <<NGINX
server {
    listen 127.0.0.1:$WEB_PORT;
    server_name _;
    root $WEB_DIR/dist;
    index index.html;

    location /api/ {
        proxy_pass http://127.0.0.1:$APP_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }

    location / {
        try_files \$uri \$uri/ /index.html;
    }
}
NGINX

sudo ln -sf /etc/nginx/sites-available/ajoliving_web /etc/nginx/sites-enabled/ajoliving_web
sudo nginx -t
sudo systemctl reload nginx
REMOTE_APP

if [ "$APPLY_NETWORK" = "1" ]; then
  ssh admin@"$APP_SERVER" \
    WEB_DOMAIN="$WEB_DOMAIN" \
    APP_DOMAIN="$APP_DOMAIN" \
    WEB_PORT="$WEB_PORT" \
    APP_PORT="$APP_PORT" \
    bash -s <<'REMOTE_NETWORK'
set -euo pipefail

# 1. Update one FRP proxy block.
update_proxy_block() {
  local name="$1"
  local port="$2"
  local domain="$3"
  awk -v name="$name" '
    BEGIN {skip=0}
    /^\[\[proxies\]\]/ {block=1; buf=$0; next}
    block==1 {buf=buf ORS $0}
    block==1 && /^name =/ {
      if ($0 ~ "name = \""name"\"") {skip=1}
    }
    block==1 && /^\s*$/ {
      if (skip==0) {print buf ORS}
      block=0; skip=0; buf=""
      next
    }
    END {
      if (block==1 && skip==0) {print buf}
    }
    block==0 {print}
  ' /home/admin/frp/frpc.toml > /home/admin/frp/frpc.toml.tmp
  mv /home/admin/frp/frpc.toml.tmp /home/admin/frp/frpc.toml

  cat >> /home/admin/frp/frpc.toml <<FRP

[[proxies]]
name = "$name"
type = "http"
localIP = "127.0.0.1"
localPort = $port
customDomains = ["$domain"]
FRP
}

update_proxy_block "ajoliving-web" "$WEB_PORT" "$WEB_DOMAIN"
update_proxy_block "ajoliving-api" "$APP_PORT" "$APP_DOMAIN"

sudo supervisorctl restart frpc
REMOTE_NETWORK
fi

if [ "$APPLY_SSL" = "1" ]; then
  ssh admin@"$GATEWAY_SERVER" \
    WEB_DOMAIN="$WEB_DOMAIN" \
    APP_DOMAIN="$APP_DOMAIN" \
    bash -s <<'REMOTE_SSL'
set -euo pipefail

sudo supervisorctl stop frps
sudo certbot certonly --standalone -d "$WEB_DOMAIN"
sudo certbot certonly --standalone -d "$APP_DOMAIN"
sudo supervisorctl start frps

sudo tee /etc/nginx/sites-available/ajoliving_web_ssl.conf >/dev/null <<NGINX
server {
    listen 443 ssl;
    server_name $WEB_DOMAIN;
    ssl_certificate /etc/letsencrypt/live/$WEB_DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$WEB_DOMAIN/privkey.pem;
    include snippets/proxy_params;
    location / {
        proxy_pass http://127.0.0.1:80;
    }
}
NGINX

sudo tee /etc/nginx/sites-available/ajoliving_api_ssl.conf >/dev/null <<NGINX
server {
    listen 443 ssl;
    server_name $APP_DOMAIN;
    ssl_certificate /etc/letsencrypt/live/$APP_DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$APP_DOMAIN/privkey.pem;
    include snippets/proxy_params;
    location / {
        proxy_pass http://127.0.0.1:80;
    }
}
NGINX

sudo ln -sf /etc/nginx/sites-available/ajoliving_web_ssl.conf /etc/nginx/sites-enabled/
sudo ln -sf /etc/nginx/sites-available/ajoliving_api_ssl.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
REMOTE_SSL
fi

echo "Deploy finished."
