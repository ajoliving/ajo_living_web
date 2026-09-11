#!/bin/bash
#
# AJO Living production deployment script.
# 1. Build frontend and backend locally.
# 2. Publish each deployment into a timestamped remote release directory.
# 3. Switch backend and frontend with symlinks so later code updates do not touch data.
# 4. Keep database restore disabled by default; production data is preserved on normal deploys.
# 5. Create a production database backup before backend restart so additive migrations can be rolled back manually.
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
REDIS_PORT="${REDIS_PORT:-6382}"
APP_PUBLIC_BASE_URL="${APP_PUBLIC_BASE_URL:-https://$WEB_DOMAIN}"
APP_API_PUBLIC_BASE_URL="${APP_API_PUBLIC_BASE_URL:-https://$APP_DOMAIN}"
OSS_CALLBACK_ENABLED="${OSS_CALLBACK_ENABLED:-1}"
OSS_CALLBACK_URL="${OSS_CALLBACK_URL:-$APP_API_PUBLIC_BASE_URL/api/v1/oss/callback}"
OSS_ALLOWED_ORIGINS="${OSS_ALLOWED_ORIGINS:-$APP_PUBLIC_BASE_URL}"

BACKUP_DB="${BACKUP_DB:-1}"
DB_BACKUP_KEEP="${DB_BACKUP_KEEP:-10}"
RESTORE_DB="${RESTORE_DB:-0}"
RESTORE_DB_CONFIRM="${RESTORE_DB_CONFIRM:-}"
APPLY_NETWORK="${APPLY_NETWORK:-1}"
APPLY_SSL="${APPLY_SSL:-0}"
APPLY_OSS_CORS="${APPLY_OSS_CORS:-1}"
VERIFY_DEPLOY="${VERIFY_DEPLOY:-1}"
VERIFY_HTTPS="${VERIFY_HTTPS:-1}"
SYNC_ENV="${SYNC_ENV:-0}"
KEEP_RELEASES="${KEEP_RELEASES:-5}"
CERTBOT_EMAIL="${CERTBOT_EMAIL:-}"
RELEASE_ID="${RELEASE_ID:-$(date -u +%Y%m%d%H%M%S)-$$}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel 2>/dev/null || printf '%s' "$SCRIPT_DIR")"
BACKEND_DIR="$ROOT_DIR/http_service"
FRONTEND_DIR="${FRONTEND_DIR:-$ROOT_DIR/web-new}"

ENV_FILE="${ENV_FILE:-$BACKEND_DIR/.env}"
DUMP_FILE="$ROOT_DIR/ajoliving_backup.sql"
DUMP_READY="0"
REMOTE_TMP="/tmp/${PROJECT}_release_${RELEASE_ID}"

# 1. Validate local deployment inputs.
if [ ! -f "$ENV_FILE" ]; then
  echo "ENV_FILE not found: $ENV_FILE"
  exit 1
fi

# 2. Build backend binary.
pushd "$BACKEND_DIR" >/dev/null
GO111MODULE=on go mod download
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ajoliving_server ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o oss-cors ./cmd/oss-cors
popd >/dev/null

# 3. Build frontend assets.
pushd "$FRONTEND_DIR" >/dev/null
npm ci
VITE_API_BASE_URL="$APP_API_PUBLIC_BASE_URL/api/v1" npm run build
popd >/dev/null

# 4. Create a local database dump only when an explicit restore is confirmed.
if [ "$RESTORE_DB" = "1" ]; then
  if [ "$RESTORE_DB_CONFIRM" != "RESTORE_PRODUCTION_AJOLIVING" ]; then
    echo "Refusing database restore. Set RESTORE_DB_CONFIRM=RESTORE_PRODUCTION_AJOLIVING to confirm."
    exit 1
  fi
  rm -f "$DUMP_FILE"
  if docker ps --format '{{.Names}}' | grep -q '^ajoliving_http_service_postgres$'; then
    docker exec ajoliving_http_service_postgres pg_dump -U postgres --clean --if-exists --no-owner --no-privileges ajoliving > "$DUMP_FILE"
    DUMP_READY="1"
  else
    echo "Local postgres container not found, skip dump."
  fi
fi

# 5. Upload build artifacts into a release-scoped temporary directory.
ssh admin@"$APP_SERVER" "rm -rf '$REMOTE_TMP' && mkdir -p '$REMOTE_TMP/web_dist' '$REMOTE_TMP/web_public'"
scp "$BACKEND_DIR/ajoliving_server" admin@"$APP_SERVER":"$REMOTE_TMP/ajoliving_server"
scp "$BACKEND_DIR/oss-cors" admin@"$APP_SERVER":"$REMOTE_TMP/oss-cors"
scp "$ENV_FILE" admin@"$APP_SERVER":"$REMOTE_TMP/ajoliving.env"
rsync -az --delete "$FRONTEND_DIR/dist/" admin@"$APP_SERVER":"$REMOTE_TMP/web_dist/"
rsync -az --delete "$FRONTEND_DIR/public/" admin@"$APP_SERVER":"$REMOTE_TMP/web_public/"
if [ "$RESTORE_DB" = "1" ] && [ "$DUMP_READY" = "1" ] && [ -f "$DUMP_FILE" ]; then
  scp "$DUMP_FILE" admin@"$APP_SERVER":"$REMOTE_TMP/ajoliving_backup.sql"
fi

# 6. Install the release on the application server.
ssh admin@"$APP_SERVER" \
  PROJECT="$PROJECT" \
  RELEASE_ID="$RELEASE_ID" \
  REMOTE_TMP="$REMOTE_TMP" \
  APP_PORT="$APP_PORT" \
  WEB_PORT="$WEB_PORT" \
  DB_PORT="$DB_PORT" \
  REDIS_PORT="$REDIS_PORT" \
  APP_PUBLIC_BASE_URL="$APP_PUBLIC_BASE_URL" \
  APP_API_PUBLIC_BASE_URL="$APP_API_PUBLIC_BASE_URL" \
  OSS_CALLBACK_ENABLED="$OSS_CALLBACK_ENABLED" \
  OSS_CALLBACK_URL="$OSS_CALLBACK_URL" \
  OSS_ALLOWED_ORIGINS="$OSS_ALLOWED_ORIGINS" \
  BACKUP_DB="$BACKUP_DB" \
  DB_BACKUP_KEEP="$DB_BACKUP_KEEP" \
  RESTORE_DB="$RESTORE_DB" \
  RESTORE_DB_CONFIRM="$RESTORE_DB_CONFIRM" \
  DUMP_READY="$DUMP_READY" \
  SYNC_ENV="$SYNC_ENV" \
  KEEP_RELEASES="$KEEP_RELEASES" \
  bash -s <<'REMOTE_APP'
set -euo pipefail

BASE_DIR="/home/admin/$PROJECT"
APP_DIR="$BASE_DIR/server"
WEB_DIR="$BASE_DIR/web"
DB_DIR="$BASE_DIR/db"
APP_RELEASES_DIR="$APP_DIR/releases"
WEB_RELEASES_DIR="$WEB_DIR/releases"
APP_RELEASE_DIR="$APP_RELEASES_DIR/$RELEASE_ID"
WEB_RELEASE_DIR="$WEB_RELEASES_DIR/$RELEASE_ID"

mkdir -p "$APP_DIR/tmp" "$APP_RELEASES_DIR" "$WEB_RELEASES_DIR" "$DB_DIR"
mkdir -p "$APP_RELEASE_DIR" "$WEB_RELEASE_DIR/dist" "$WEB_RELEASE_DIR/public"

# 1. Remove stale database dump unless this run intentionally uploads one.
if [ "${DUMP_READY:-0}" != "1" ]; then
  rm -f "$REMOTE_TMP/ajoliving_backup.sql" /tmp/ajoliving_backup.sql
fi

# 2. Stage the new backend and frontend release files.
cp "$REMOTE_TMP/ajoliving_server" "$APP_RELEASE_DIR/ajoliving_server"
chmod +x "$APP_RELEASE_DIR/ajoliving_server"
cp "$REMOTE_TMP/oss-cors" "$APP_RELEASE_DIR/oss-cors"
chmod +x "$APP_RELEASE_DIR/oss-cors"
rsync -a --delete "$REMOTE_TMP/web_dist/" "$WEB_RELEASE_DIR/dist/"
rsync -a --delete "$REMOTE_TMP/web_public/" "$WEB_RELEASE_DIR/public/"

# 3. Keep the old single-binary layout available as a rollback target.
if [ ! -e "$APP_DIR/current" ] && [ -f "$APP_DIR/ajoliving_server" ]; then
  LEGACY_RELEASE_DIR="$APP_RELEASES_DIR/legacy-before-$RELEASE_ID"
  mkdir -p "$LEGACY_RELEASE_DIR"
  cp "$APP_DIR/ajoliving_server" "$LEGACY_RELEASE_DIR/ajoliving_server"
  chmod +x "$LEGACY_RELEASE_DIR/ajoliving_server"
  ln -sfn "$LEGACY_RELEASE_DIR" "$APP_DIR/current"
fi

# 4. Preserve production .env by default; use SYNC_ENV=1 to replace it from local .env.
if [ -f "$APP_DIR/.env" ]; then
  cp "$APP_DIR/.env" "$APP_DIR/.env.backup.$RELEASE_ID"
  if [ "$SYNC_ENV" = "1" ]; then
    cp "$REMOTE_TMP/ajoliving.env" "$APP_DIR/.env"
  fi
else
  cp "$REMOTE_TMP/ajoliving.env" "$APP_DIR/.env"
fi
chmod 600 "$APP_DIR/.env"

# 5. Fill missing environment keys without overwriting production secrets.
if [ "$SYNC_ENV" != "1" ] && [ -f "$REMOTE_TMP/ajoliving.env" ]; then
  while IFS= read -r line || [ -n "$line" ]; do
    case "$line" in
      ''|\#*) continue ;;
      *=*)
        key="${line%%=*}"
        if [ -n "$key" ] && ! grep -q "^$key=" "$APP_DIR/.env"; then
          printf '%s\n' "$line" >> "$APP_DIR/.env"
        fi
        ;;
    esac
  done < "$REMOTE_TMP/ajoliving.env"
fi

# 6. Set required production runtime values without breaking other existing secrets.
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
set_env "APP_PUBLIC_BASE_URL" "$APP_PUBLIC_BASE_URL"
set_env "APP_API_PUBLIC_BASE_URL" "$APP_API_PUBLIC_BASE_URL"
set_env "DB_DRIVER" "postgres"
set_env "DB_DSN" "host=127.0.0.1 user=postgres password=postgres dbname=ajoliving port=$DB_PORT sslmode=disable TimeZone=Asia/Shanghai"
set_env "REDIS_ENABLED" "true"
set_env "REDIS_ADDR" "127.0.0.1:$REDIS_PORT"
set_env "REDIS_PASSWORD" ""
set_env "REDIS_DB" "0"
set_env "ISMART_BUILDING_CACHE_TTL" "5m"
set_env "POS_DIRECTORY_CACHE_TTL" "5m"
set_env "WEB_PUBLIC_DIR" "$WEB_DIR/current_public"
set_env "OSS_CALLBACK_ENABLED" "$OSS_CALLBACK_ENABLED"
set_env "OSS_CALLBACK_URL" "$OSS_CALLBACK_URL"
set_env "OSS_ALLOWED_ORIGINS" "$OSS_ALLOWED_ORIGINS"
set_env "MEDIA_PROCESSING_PROVIDER" "ffmpeg"
set_env "MEDIA_FFMPEG_BINARY" "ffmpeg"
set_env "MEDIA_SCAN_PROVIDER" "clamav"
set_env "MEDIA_CLAMAV_BINARY" "clamscan"

if ! command -v ffmpeg >/dev/null 2>&1 || ! command -v clamscan >/dev/null 2>&1; then
  echo "Production chat media requires ffmpeg and clamscan on the application host."
  exit 1
fi

# 7. Keep the existing PostgreSQL volume intact.
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
  redis:
    image: redis:7-alpine
    container_name: ajoliving_redis
    restart: unless-stopped
    command: ["redis-server", "--save", "", "--appendonly", "no"]
    ports:
      - "127.0.0.1:$REDIS_PORT:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 20
      start_period: 5s
volumes:
  ajoliving_postgres_data:
YAML

sudo docker compose -f "$DB_DIR/docker-compose.yml" up -d

db_ready="0"
for i in {1..30}; do
  if sudo docker exec ajoliving_postgres pg_isready -U postgres -d ajoliving >/dev/null 2>&1; then
    db_ready="1"
    break
  fi
  sleep 2
done

if [ "$db_ready" != "1" ]; then
  echo "PostgreSQL is not ready"; exit 1
fi

redis_ready="0"
for i in {1..30}; do
  if sudo docker exec ajoliving_redis redis-cli ping 2>/dev/null | grep -q '^PONG$'; then
    redis_ready="1"
    break
  fi
  sleep 2
done

if [ "$redis_ready" != "1" ]; then
  echo "Redis is not ready"; exit 1
fi

# 8. Take a server-side production backup before backend startup runs migrations.
if [ "$BACKUP_DB" = "1" ]; then
  DB_BACKUP_DIR="$DB_DIR/backups"
  mkdir -p "$DB_BACKUP_DIR"
  DB_BACKUP_PATH="$DB_BACKUP_DIR/ajoliving_${RELEASE_ID}.dump"
  sudo docker exec ajoliving_postgres pg_dump -U postgres -d ajoliving -Fc --no-owner --no-privileges > "$DB_BACKUP_PATH"
  chmod 600 "$DB_BACKUP_PATH"

  if [ "$DB_BACKUP_KEEP" -gt 0 ] 2>/dev/null; then
    ls -1t "$DB_BACKUP_DIR"/ajoliving_*.dump 2>/dev/null | tail -n +"$((DB_BACKUP_KEEP + 1))" | xargs -r rm -f
  fi
fi

if [ "$RESTORE_DB" = "1" ] && [ "${DUMP_READY:-0}" = "1" ] && [ -f "$REMOTE_TMP/ajoliving_backup.sql" ]; then
  if [ "${RESTORE_DB_CONFIRM:-}" != "RESTORE_PRODUCTION_AJOLIVING" ]; then
    echo "Refusing database restore on server. Missing RESTORE_DB_CONFIRM=RESTORE_PRODUCTION_AJOLIVING."
    exit 1
  fi
  sudo docker exec -i ajoliving_postgres psql -v ON_ERROR_STOP=1 -U postgres -d ajoliving < "$REMOTE_TMP/ajoliving_backup.sql"
fi

# 9. Write a stable supervisor runner that follows the current release symlink.
cat > "$APP_DIR/run.sh" <<SH
#!/bin/bash
set -e
cd "$APP_DIR"
set -a; source .env; set +a
exec "$APP_DIR/current/ajoliving_server"
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

# 10. Switch public assets before backend boot so startup seed logic can read them.
previous_public_target="$(readlink -f "$WEB_DIR/current_public" 2>/dev/null || true)"
ln -sfn "$WEB_RELEASE_DIR/public" "$WEB_DIR/current_public"

# 11. Switch backend release and rollback automatically if health check fails.
previous_app_target="$(readlink -f "$APP_DIR/current" 2>/dev/null || true)"
rollback_backend() {
  if [ -n "$previous_app_target" ] && [ -d "$previous_app_target" ]; then
    ln -sfn "$previous_app_target" "$APP_DIR/current"
    sudo supervisorctl restart ajoliving_server >/dev/null 2>&1 || true
  fi
  if [ -n "$previous_public_target" ] && [ -d "$previous_public_target" ]; then
    ln -sfn "$previous_public_target" "$WEB_DIR/current_public"
  fi
}

ln -sfn "$APP_RELEASE_DIR" "$APP_DIR/current"
if sudo supervisorctl status ajoliving_server >/dev/null 2>&1; then
  if ! sudo supervisorctl restart ajoliving_server; then
    rollback_backend
    exit 1
  fi
else
  if ! sudo supervisorctl start ajoliving_server; then
    rollback_backend
    exit 1
  fi
fi

backend_ready="0"
for i in {1..30}; do
  if curl -fsS --max-time 5 "http://127.0.0.1:$APP_PORT/api/v1/health" >/dev/null; then
    backend_ready="1"
    break
  fi
  sleep 2
done

if [ "$backend_ready" != "1" ]; then
  rollback_backend
  echo "New backend release failed health check and was rolled back."
  exit 1
fi

# 12. Switch frontend release and reload nginx.
previous_dist_target="$(readlink -f "$WEB_DIR/current_dist" 2>/dev/null || true)"
rollback_frontend() {
  if [ -n "$previous_dist_target" ] && [ -d "$previous_dist_target" ]; then
    ln -sfn "$previous_dist_target" "$WEB_DIR/current_dist"
  fi
}

ln -sfn "$WEB_RELEASE_DIR/dist" "$WEB_DIR/current_dist"

sudo tee /etc/nginx/sites-available/ajoliving_web >/dev/null <<NGINX
server {
    listen 127.0.0.1:$WEB_PORT;
    server_name _;
    root $WEB_DIR/current_dist;
    index index.html;

    if (\$http_x_ajo_forwarded_proto != "https") {
        return 301 https://\$host\$request_uri;
    }

    location ^~ /api/v1/realtime/ws {
        proxy_pass http://127.0.0.1:$APP_PORT;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$http_x_forwarded_proto;
        proxy_set_header X-AJO-Forwarded-Proto \$http_x_ajo_forwarded_proto;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        proxy_buffering off;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:$APP_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$http_x_forwarded_proto;
        proxy_set_header X-AJO-Forwarded-Proto \$http_x_ajo_forwarded_proto;
    }

    location / {
        try_files \$uri \$uri/ /index.html;
    }
}
NGINX

sudo ln -sf /etc/nginx/sites-available/ajoliving_web /etc/nginx/sites-enabled/ajoliving_web
if ! sudo nginx -t; then
  rollback_frontend
  sudo nginx -t
  exit 1
fi
sudo systemctl reload nginx

# 13. Keep only the latest release directories after a successful switch.
if [ "$KEEP_RELEASES" -gt 0 ] 2>/dev/null; then
  ls -1dt "$APP_RELEASES_DIR"/* 2>/dev/null | tail -n +"$((KEEP_RELEASES + 1))" | xargs -r rm -rf
  ls -1dt "$WEB_RELEASES_DIR"/* 2>/dev/null | tail -n +"$((KEEP_RELEASES + 1))" | xargs -r rm -rf
fi

rm -rf "$REMOTE_TMP"
REMOTE_APP

# 7. Keep FRP routes aligned with the current ports and domains.
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

# 8. Optionally request or refresh the gateway certificate and HTTPS nginx blocks.
if [ "$APPLY_SSL" = "1" ]; then
  ssh admin@"$GATEWAY_SERVER" \
    WEB_DOMAIN="$WEB_DOMAIN" \
    APP_DOMAIN="$APP_DOMAIN" \
    CERTBOT_EMAIL="$CERTBOT_EMAIL" \
    bash -s <<'REMOTE_SSL'
set -euo pipefail

certbot_email_args=(--register-unsafely-without-email)
if [ -n "${CERTBOT_EMAIL:-}" ]; then
  certbot_email_args=(--email "$CERTBOT_EMAIL")
fi

sudo supervisorctl stop frps || true
cleanup() {
  sudo supervisorctl start frps >/dev/null 2>&1 || true
}
trap cleanup EXIT

sudo certbot certonly --standalone --non-interactive --agree-tos --expand "${certbot_email_args[@]}" -d "$WEB_DOMAIN" -d "$APP_DOMAIN"
sudo supervisorctl start frps
trap - EXIT

sudo tee /etc/nginx/sites-available/ajoliving_ssl.conf >/dev/null <<NGINX
server {
    listen 443 ssl;
    server_name $WEB_DOMAIN;

    ssl_certificate /etc/letsencrypt/live/$WEB_DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$WEB_DOMAIN/privkey.pem;

    include snippets/proxy_params;
    proxy_set_header X-Forwarded-Proto https;
    proxy_set_header X-AJO-Forwarded-Proto https;

    location / {
        proxy_pass http://127.0.0.1:80;
    }
}

server {
    listen 443 ssl;
    server_name $APP_DOMAIN;

    ssl_certificate /etc/letsencrypt/live/$WEB_DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$WEB_DOMAIN/privkey.pem;

    include snippets/proxy_params;
    proxy_set_header X-Forwarded-Proto https;
    proxy_set_header X-AJO-Forwarded-Proto https;

    location / {
        proxy_pass http://127.0.0.1:80;
    }
}
NGINX

sudo rm -f /etc/nginx/sites-enabled/ajoliving_web_ssl.conf /etc/nginx/sites-enabled/ajoliving_api_ssl.conf
sudo ln -sf /etc/nginx/sites-available/ajoliving_ssl.conf /etc/nginx/sites-enabled/ajoliving_ssl.conf
sudo nginx -t
sudo systemctl reload nginx
REMOTE_SSL
fi

# 9. Optionally align OSS CORS with the production frontend HTTPS origin.
if [ "$APPLY_OSS_CORS" = "1" ]; then
  ssh admin@"$APP_SERVER" \
    PROJECT="$PROJECT" \
    bash -s <<'REMOTE_OSS_CORS'
set -euo pipefail

APP_DIR="/home/admin/$PROJECT/server"
cd "$APP_DIR"
set -a; source .env; set +a
if [ -f "$APP_DIR/current/oss-cors" ]; then
  "$APP_DIR/current/oss-cors"
else
  echo "oss-cors helper not found; skip OSS CORS update."
fi
REMOTE_OSS_CORS
fi

# 10. Verify the public routes after deployment.
if [ "$VERIFY_DEPLOY" = "1" ]; then
  for i in {1..10}; do
    if curl -fsS --max-time 10 "http://$WEB_DOMAIN/api/v1/health" >/dev/null && curl -fsS --max-time 10 "http://$APP_DOMAIN/api/v1/health" >/dev/null; then
      break
    fi
    if [ "$i" = "10" ]; then
      echo "Public HTTP health check failed"; exit 1
    fi
    sleep 2
  done
  if [ "$VERIFY_HTTPS" = "1" ]; then
    for i in {1..10}; do
      if curl -fsS --max-time 10 "https://$WEB_DOMAIN/api/v1/health" >/dev/null && curl -fsS --max-time 10 "https://$APP_DOMAIN/api/v1/health" >/dev/null; then
        break
      fi
      if [ "$i" = "10" ]; then
        echo "Public HTTPS health check failed"; exit 1
      fi
      sleep 2
    done
  fi
fi

echo "Deploy finished: $RELEASE_ID"
