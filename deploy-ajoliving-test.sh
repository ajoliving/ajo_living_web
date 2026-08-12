#!/bin/bash
#
# AJO Living testing environment deployment script.
# 1. Build the current backend and frontend locally.
# 2. Publish isolated testing releases without changing production runtime files.
# 3. Create dedicated PostgreSQL and Redis containers on loopback-only ports.
# 4. Restore a production snapshot only after explicit testing confirmation.
# 5. Inherit external service configuration and switch releases after health checks.
#
set -euo pipefail

APP_SERVER="47.239.117.108"
PROJECT="ajoliving-test"
WEB_DOMAIN="test.ajoliving.skylinedances.com"

APP_PORT="${APP_PORT:-20048}"
WEB_PORT="${WEB_PORT:-20047}"
DB_PORT="${DB_PORT:-45433}"
REDIS_PORT="${REDIS_PORT:-6383}"
APP_PUBLIC_BASE_URL="https://$WEB_DOMAIN"
APP_API_PUBLIC_BASE_URL="https://$WEB_DOMAIN"

INITIALIZE_DB="${INITIALIZE_DB:-0}"
INITIALIZE_DB_CONFIRM="${INITIALIZE_DB_CONFIRM:-}"
SKIP_BUILD="${SKIP_BUILD:-0}"
KEEP_RELEASES="${KEEP_RELEASES:-5}"
RELEASE_ID="${RELEASE_ID:-$(date -u +%Y%m%d%H%M%S)-test-$$}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel 2>/dev/null || printf '%s' "$SCRIPT_DIR")"
BACKEND_DIR="$ROOT_DIR/http_service"
FRONTEND_DIR="$ROOT_DIR/web-new"
REMOTE_TMP="/tmp/${PROJECT}_release_${RELEASE_ID}"
BACKEND_ARCHIVE="$(mktemp -t ajoliving-test-server.XXXXXX.gz)"
REKEY_ARCHIVE="$(mktemp -t ajoliving-test-rekey.XXXXXX.gz)"
REKEY_BINARY="$(mktemp -t ajoliving-test-rekey.XXXXXX)"
trap 'rm -f "$BACKEND_ARCHIVE" "$REKEY_ARCHIVE" "$REKEY_BINARY"' EXIT

# 1. Require an explicit confirmation before replacing the testing database.
if [ "$INITIALIZE_DB" = "1" ] && [ "$INITIALIZE_DB_CONFIRM" != "RESTORE_AJOLIVING_TEST" ]; then
  echo "Refusing testing database restore. Set INITIALIZE_DB_CONFIRM=RESTORE_AJOLIVING_TEST."
  exit 1
fi

# 2. Build release artifacts unless a verified retry reuses the latest build.
if [ "$SKIP_BUILD" != "1" ]; then
  pushd "$BACKEND_DIR" >/dev/null
  GO111MODULE=on go mod download
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ajoliving_server ./cmd/server
  popd >/dev/null

  pushd "$FRONTEND_DIR" >/dev/null
  npm ci
  VITE_API_BASE_URL="/api/v1" npm run build
  popd >/dev/null
fi

if [ "$INITIALIZE_DB" = "1" ]; then
  pushd "$BACKEND_DIR" >/dev/null
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o "$REKEY_BINARY" ./cmd/rekey-encrypted-data
  gzip -c "$REKEY_BINARY" > "$REKEY_ARCHIVE"
  popd >/dev/null
fi

if [ ! -x "$BACKEND_DIR/ajoliving_server" ] || [ ! -f "$FRONTEND_DIR/dist/index.html" ]; then
  echo "Built backend or frontend artifacts are missing"
  exit 1
fi

# 3. Compress the backend binary for reliable remote transfer.
gzip -c "$BACKEND_DIR/ajoliving_server" > "$BACKEND_ARCHIVE"

# 4. Upload release artifacts to an isolated temporary directory.
ssh -o IPQoS=none admin@"$APP_SERVER" "rm -rf '$REMOTE_TMP' && mkdir -p '$REMOTE_TMP/web_dist' '$REMOTE_TMP/web_public'"
scp -O -o IPQoS=none "$BACKEND_ARCHIVE" admin@"$APP_SERVER":"$REMOTE_TMP/ajoliving_server.gz"
if [ "$INITIALIZE_DB" = "1" ]; then
  scp -O -o IPQoS=none "$REKEY_ARCHIVE" admin@"$APP_SERVER":"$REMOTE_TMP/rekey_encrypted_data.gz"
fi
rsync -az --delete -e "ssh -o IPQoS=none" "$FRONTEND_DIR/dist/" admin@"$APP_SERVER":"$REMOTE_TMP/web_dist/"
rsync -az --delete -e "ssh -o IPQoS=none" "$FRONTEND_DIR/public/" admin@"$APP_SERVER":"$REMOTE_TMP/web_public/"

# 5. Install and activate the testing release on the application server.
ssh -o IPQoS=none admin@"$APP_SERVER" \
  PROJECT="$PROJECT" \
  RELEASE_ID="$RELEASE_ID" \
  REMOTE_TMP="$REMOTE_TMP" \
  APP_PORT="$APP_PORT" \
  WEB_PORT="$WEB_PORT" \
  DB_PORT="$DB_PORT" \
  REDIS_PORT="$REDIS_PORT" \
  APP_PUBLIC_BASE_URL="$APP_PUBLIC_BASE_URL" \
  APP_API_PUBLIC_BASE_URL="$APP_API_PUBLIC_BASE_URL" \
  WEB_DOMAIN="$WEB_DOMAIN" \
  INITIALIZE_DB="$INITIALIZE_DB" \
  INITIALIZE_DB_CONFIRM="$INITIALIZE_DB_CONFIRM" \
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
PRODUCTION_ENV="/home/admin/ajoliving/server/.env"

mkdir -p "$APP_DIR/tmp" "$APP_RELEASE_DIR" "$WEB_RELEASE_DIR/dist" "$WEB_RELEASE_DIR/public" "$DB_DIR/backups"

# 1. Stage backend and frontend files under the release identifier.
gzip -dc "$REMOTE_TMP/ajoliving_server.gz" > "$APP_RELEASE_DIR/ajoliving_server"
chmod +x "$APP_RELEASE_DIR/ajoliving_server"
rsync -a --delete "$REMOTE_TMP/web_dist/" "$WEB_RELEASE_DIR/dist/"
rsync -a --delete "$REMOTE_TMP/web_public/" "$WEB_RELEASE_DIR/public/"

# 2. Create dedicated PostgreSQL and Redis services.
cat > "$DB_DIR/docker-compose.yml" <<YAML
name: ajoliving_test
services:
  postgres:
    image: postgres:16-alpine
    container_name: ajoliving_test_postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: ajoliving_test
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      TZ: Asia/Shanghai
      PGTZ: Asia/Shanghai
    ports:
      - "127.0.0.1:$DB_PORT:5432"
    volumes:
      - ajoliving_test_postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d ajoliving_test"]
      interval: 5s
      timeout: 5s
      retries: 20
      start_period: 10s
  redis:
    image: redis:7-alpine
    container_name: ajoliving_test_redis
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
  ajoliving_test_postgres_data:
YAML

sudo docker compose -f "$DB_DIR/docker-compose.yml" up -d

for i in {1..30}; do
  if sudo docker exec ajoliving_test_postgres pg_isready -U postgres -d ajoliving_test >/dev/null 2>&1; then
    break
  fi
  if [ "$i" = "30" ]; then
    echo "Testing PostgreSQL is not ready"
    exit 1
  fi
  sleep 2
done

for i in {1..30}; do
  if sudo docker exec ajoliving_test_redis redis-cli ping 2>/dev/null | grep -q '^PONG$'; then
    break
  fi
  if [ "$i" = "30" ]; then
    echo "Testing Redis is not ready"
    exit 1
  fi
  sleep 2
done

# 3. Restore one logical production snapshot into testing only when confirmed.
if [ "$INITIALIZE_DB" = "1" ]; then
  if [ "$INITIALIZE_DB_CONFIRM" != "RESTORE_AJOLIVING_TEST" ]; then
    echo "Refusing unconfirmed testing database restore"
    exit 1
  fi

  SNAPSHOT_PATH="$DB_DIR/backups/ajoliving_production_${RELEASE_ID}.dump"
  sudo docker exec ajoliving_postgres pg_dump -U postgres -d ajoliving -Fc --no-owner --no-privileges > "$SNAPSHOT_PATH"
  chmod 600 "$SNAPSHOT_PATH"
  sudo docker exec -i ajoliving_test_postgres pg_restore \
    -U postgres \
    -d ajoliving_test \
    --clean \
    --if-exists \
    --no-owner \
    --no-privileges < "$SNAPSHOT_PATH"

fi

# 4. Refresh external service settings while preserving testing-only secrets.
new_environment="0"
if [ ! -f "$APP_DIR/.env" ]; then
	if [ ! -f "$PRODUCTION_ENV" ]; then
		echo "Production environment file is missing"
		exit 1
	fi
	new_environment="1"
else
	cp "$APP_DIR/.env" "$APP_DIR/.env.backup.$RELEASE_ID"
fi

if [ ! -f "$PRODUCTION_ENV" ]; then
	echo "Production environment file is missing"
	exit 1
fi

testing_jwt_secret=""
testing_encryption_key=""
if [ "$new_environment" != "1" ]; then
	testing_jwt_secret="$(bash -c 'set -a; source "$1"; set +a; printf %s "${JWT_SECRET:-}"' _ "$APP_DIR/.env")"
	testing_encryption_key="$(bash -c 'set -a; source "$1"; set +a; printf %s "${ENCRYPTION_KEY:-}"' _ "$APP_DIR/.env")"
fi
cp "$PRODUCTION_ENV" "$APP_DIR/.env"

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

if [ "$new_environment" = "1" ]; then
	set_env "JWT_SECRET" "$(openssl rand -hex 32)"
	set_env "ENCRYPTION_KEY" "$(openssl rand -hex 32)"
else
	set_env "JWT_SECRET" "$testing_jwt_secret"
	set_env "ENCRYPTION_KEY" "$testing_encryption_key"
fi

set_env "APP_ENV" "testing"
set_env "APP_PORT" "$APP_PORT"
set_env "APP_PUBLIC_BASE_URL" "$APP_PUBLIC_BASE_URL"
set_env "APP_API_PUBLIC_BASE_URL" "$APP_API_PUBLIC_BASE_URL"
set_env "CORS_ALLOWED_ORIGINS" "$APP_PUBLIC_BASE_URL"
set_env "STRICT_PRODUCTION_CONFIG" "false"
set_env "DB_DRIVER" "postgres"
set_env "DB_DSN" "host=127.0.0.1 user=postgres password=postgres dbname=ajoliving_test port=$DB_PORT sslmode=disable TimeZone=Asia/Shanghai"
set_env "REDIS_ENABLED" "true"
set_env "REDIS_ADDR" "127.0.0.1:$REDIS_PORT"
set_env "REDIS_PASSWORD" ""
set_env "REDIS_DB" "0"
set_env "OSS_ALLOWED_ORIGINS" "$APP_PUBLIC_BASE_URL"
set_env "OSS_CALLBACK_URL" "$APP_API_PUBLIC_BASE_URL/api/v1/oss/callback"
set_env "WALLET_RECHARGE_RETURN_BASE_URL" "$APP_PUBLIC_BASE_URL"
set_env "WALLET_RECHARGE_NOTIFY_URL" "$APP_API_PUBLIC_BASE_URL/api/v1/payments/easylink/notify"
set_env "WEB_PUBLIC_DIR" "$WEB_DIR/current_public"
chmod 600 "$APP_DIR/.env"

# 4.1 Re-encrypt retained snapshot values with the independent testing key.
if [ "$INITIALIZE_DB" = "1" ]; then
  REKEY_BINARY="$REMOTE_TMP/rekey_encrypted_data"
  gzip -dc "$REMOTE_TMP/rekey_encrypted_data.gz" > "$REKEY_BINARY"
  chmod 700 "$REKEY_BINARY"
  (
    set -a
    source "$PRODUCTION_ENV"
    set +a
    PRODUCTION_ENCRYPTION_KEY="$ENCRYPTION_KEY"
    set -a
    source "$APP_DIR/.env"
    set +a
    REKEY_DSN="host=127.0.0.1 user=postgres password=postgres dbname=ajoliving_test port=$DB_PORT sslmode=disable TimeZone=Asia/Shanghai" \
      REKEY_FROM_KEY="$PRODUCTION_ENCRYPTION_KEY" \
      REKEY_TO_KEY="$ENCRYPTION_KEY" \
      "$REKEY_BINARY"
  )
  rm -f "$REKEY_BINARY"
fi

# 5. Configure an isolated Supervisor process.
cat > "$APP_DIR/run.sh" <<SH
#!/bin/bash
set -e
cd "$APP_DIR"
set -a; source .env; set +a
exec "$APP_DIR/current/ajoliving_server"
SH
chmod +x "$APP_DIR/run.sh"

previous_app_target="$(readlink -f "$APP_DIR/current" 2>/dev/null || true)"
previous_public_target="$(readlink -f "$WEB_DIR/current_public" 2>/dev/null || true)"
ln -sfn "$WEB_RELEASE_DIR/public" "$WEB_DIR/current_public"
ln -sfn "$APP_RELEASE_DIR" "$APP_DIR/current"

sudo tee /etc/supervisor/conf.d/ajoliving_test_server.conf >/dev/null <<SUP
[program:ajoliving_test_server]
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

# 6. Activate the backend and roll back the symlink if health checks fail.
if sudo supervisorctl status ajoliving_test_server >/dev/null 2>&1; then
  sudo supervisorctl restart ajoliving_test_server
else
  sudo supervisorctl start ajoliving_test_server
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
  if [ -n "$previous_app_target" ] && [ -d "$previous_app_target" ]; then
    ln -sfn "$previous_app_target" "$APP_DIR/current"
    sudo supervisorctl restart ajoliving_test_server || true
  fi
  if [ -n "$previous_public_target" ] && [ -d "$previous_public_target" ]; then
    ln -sfn "$previous_public_target" "$WEB_DIR/current_public"
  fi
  echo "Testing backend failed health checks"
  exit 1
fi

# 7. Activate the frontend through a same-origin API proxy.
previous_dist_target="$(readlink -f "$WEB_DIR/current_dist" 2>/dev/null || true)"
ln -sfn "$WEB_RELEASE_DIR/dist" "$WEB_DIR/current_dist"

sudo tee /etc/nginx/sites-available/ajoliving_test_web >/dev/null <<NGINX
server {
    listen 127.0.0.1:$WEB_PORT;
    server_name _;
    root $WEB_DIR/current_dist;
    index index.html;

    if (\$http_x_ajo_forwarded_proto != "https") {
        return 301 https://\$host\$request_uri;
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

sudo ln -sf /etc/nginx/sites-available/ajoliving_test_web /etc/nginx/sites-enabled/ajoliving_test_web
if ! sudo nginx -t; then
  if [ -n "$previous_dist_target" ] && [ -d "$previous_dist_target" ]; then
    ln -sfn "$previous_dist_target" "$WEB_DIR/current_dist"
  fi
  exit 1
fi
sudo systemctl reload nginx

# 8. Keep only recent testing releases after successful activation.
if [ "$KEEP_RELEASES" -gt 0 ] 2>/dev/null; then
  ls -1dt "$APP_RELEASES_DIR"/* 2>/dev/null | tail -n +"$((KEEP_RELEASES + 1))" | xargs -r rm -rf
  ls -1dt "$WEB_RELEASES_DIR"/* 2>/dev/null | tail -n +"$((KEEP_RELEASES + 1))" | xargs -r rm -rf
fi

rm -rf "$REMOTE_TMP"
REMOTE_APP

# 6. Add the dedicated FRP route without changing production proxy blocks.
ssh -o IPQoS=none admin@"$APP_SERVER" \
  WEB_DOMAIN="$WEB_DOMAIN" \
  WEB_PORT="$WEB_PORT" \
  bash -s <<'REMOTE_NETWORK'
set -euo pipefail

if ! grep -q '^name = "ajoliving-test-web"$' /home/admin/frp/frpc.toml; then
  cat >> /home/admin/frp/frpc.toml <<FRP

[[proxies]]
name = "ajoliving-test-web"
type = "http"
localIP = "127.0.0.1"
localPort = $WEB_PORT
customDomains = ["$WEB_DOMAIN"]
FRP
fi

sudo supervisorctl restart frpc
REMOTE_NETWORK

echo "Testing deploy finished: $RELEASE_ID"
