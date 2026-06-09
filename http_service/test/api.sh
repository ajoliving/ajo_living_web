#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
HTTP_SERVICE_DIR="${ROOT_DIR}/http_service"
ENV_FILE="${ENV_FILE:-${HTTP_SERVICE_DIR}/.env}"
SERVER_LOG="${SERVER_LOG:-${HTTP_SERVICE_DIR}/test/api_server.log}"
COMPOSE_FILE="${COMPOSE_FILE:-${HTTP_SERVICE_DIR}/docker-compose.yml}"
POSTGRES_SERVICE="${POSTGRES_SERVICE:-postgres}"

BASE_URL="${BASE_URL:-}"
API_PREFIX="${API_PREFIX:-/api/v1}"
AUTO_START_SERVER="${AUTO_START_SERVER:-true}"
START_TIMEOUT="${START_TIMEOUT:-20}"

OWNER_PHONE=""
BUYER_PHONE=""
OUTSIDER_PHONE=""
OWNER_TOKEN=""
BUYER_TOKEN=""
OUTSIDER_TOKEN=""
COMMUNITY_A=""
COMMUNITY_B=""
PUBLIC_MEDIA_ID=""
PRIVATE_MEDIA_ID=""
PUBLIC_LISTING_ID=""
PRIVATE_LISTING_ID=""
CHAT_ID=""
SERVER_PID=""

print_line() {
  printf '\n%s\n' "================================================================"
}

print_step() {
  print_line
  printf '步驟: %s\n' "$1"
}

print_json() {
  printf '%s' "$1" | python3 -c '
import json
import sys

raw = sys.stdin.read()
try:
    print(json.dumps(json.loads(raw), ensure_ascii=False, indent=2))
except Exception:
    print(raw, end="")
'
}

extract_json() {
  local json="$1"
  local expr="$2"
  printf '%s' "$json" | python3 -c '
import json
import re
import sys

expr = sys.argv[1]
raw = sys.stdin.read()
try:
    obj = json.loads(raw)
except Exception:
    print("")
    sys.exit(0)

current = obj
parts = [p for p in expr.split(".") if p]
for part in parts:
    match = re.fullmatch(r"([^\[\]]+)(?:\[(\d+)\])?", part)
    if not match:
        print("")
        sys.exit(0)
    key = match.group(1)
    index = match.group(2)
    if not isinstance(current, dict):
        print("")
        sys.exit(0)
    current = current.get(key, "")
    if index is not None:
        if not isinstance(current, list):
            print("")
            sys.exit(0)
        idx = int(index)
        if idx >= len(current):
            print("")
            sys.exit(0)
        current = current[idx]

if current is None:
    print("")
elif isinstance(current, (dict, list)):
    print(json.dumps(current, ensure_ascii=False))
else:
    print(current)
' "$expr"
}

call_api() {
  local title="$1"
  local method="$2"
  local path="$3"
  local body="${4:-}"
  local auth="${5:-}"
  local extra_headers="${6:-}"

  local tmp response http_code
  tmp="$(mktemp)"

  local curl_args=(
    -sS
    -X "$method"
    "${BASE_URL}${API_PREFIX}${path}"
    -H "Content-Type: application/json"
    -w "\nHTTP_STATUS:%{http_code}\n"
  )

  if [[ -n "$auth" ]]; then
    curl_args+=(-H "Authorization: Bearer ${auth}")
  fi

  if [[ -n "$extra_headers" ]]; then
    while IFS= read -r header; do
      [[ -z "$header" ]] && continue
      curl_args+=(-H "$header")
    done <<< "$extra_headers"
  fi

  if [[ -n "$body" ]]; then
    curl_args+=(-d "$body")
  fi

  curl "${curl_args[@]}" > "$tmp"
  http_code="$(sed -n 's/^HTTP_STATUS://p' "$tmp" | tail -n 1)"
  response="$(sed '/^HTTP_STATUS:/d' "$tmp")"
  rm -f "$tmp"

  print_line
  printf '介面: %s\n' "$title"
  printf '方法: %s\n' "$method"
  printf '地址: %s%s\n' "${BASE_URL}${API_PREFIX}" "$path"
  if [[ -n "$auth" ]]; then
    printf '認證: Bearer <token>\n'
  else
    printf '認證: <none>\n'
  fi
  if [[ -n "$body" ]]; then
    printf '請求體:\n'
    print_json "$body"
  else
    printf '請求體: <empty>\n'
  fi
  printf 'HTTP 狀態: %s\n' "$http_code"
  printf '回應體:\n'
  print_json "$response"

  RESPONSE="$response"
  HTTP_STATUS="$http_code"
}

assert_status() {
  local expected="$1"
  if [[ "${HTTP_STATUS}" != "${expected}" ]]; then
    printf '預期 HTTP 狀態 %s，實際為 %s\n' "$expected" "$HTTP_STATUS"
    exit 1
  fi
}

require_commands() {
  local missing=()
  local required_cmd
  for required_cmd in curl docker python3; do
    if ! command -v "$required_cmd" >/dev/null 2>&1; then
      missing+=("$required_cmd")
    fi
  done

  if [[ "${#missing[@]}" -gt 0 ]]; then
    printf '缺少必要指令: %s\n' "${missing[*]}"
    exit 1
  fi
}

load_env_file() {
  if [[ -f "${ENV_FILE}" ]]; then
    set -a
    # shellcheck disable=SC1090
    source "${ENV_FILE}"
    set +a
  fi
  BASE_URL="${BASE_URL:-http://127.0.0.1:${APP_PORT:-8081}}"
}

docker_compose() {
  docker compose --env-file "${ENV_FILE}" -f "${COMPOSE_FILE}" "$@"
}

health_url() {
  printf '%s%s/health' "${BASE_URL}" "${API_PREFIX}"
}

wait_for_postgres() {
  local db_user="${POSTGRES_USER:-postgres}"
  local db_name="${POSTGRES_DB:-ajoliving}"
  local timeout="${POSTGRES_START_TIMEOUT:-30}"

  for _ in $(seq 1 "${timeout}"); do
    if docker_compose exec -T "${POSTGRES_SERVICE}" pg_isready -U "${db_user}" -d "${db_name}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  return 1
}

start_postgres_if_needed() {
  local db_host="${POSTGRES_HOST:-127.0.0.1}"
  local db_port="${POSTGRES_PORT:-5432}"
  local db_name="${POSTGRES_DB:-ajoliving}"

  if [[ "${DB_DRIVER:-postgres}" != "postgres" ]]; then
    printf '目前 ENV_FILE 設定的 DB_DRIVER=%s，與本腳本要求的 PostgreSQL 不一致。\n' "${DB_DRIVER:-}"
    exit 1
  fi

  print_step "確保本機 PostgreSQL 已啟動"
  docker_compose up -d "${POSTGRES_SERVICE}" >/dev/null

  if ! wait_for_postgres; then
    printf 'PostgreSQL 啟動失敗，請查看 compose 狀態。\n'
    docker_compose ps
    exit 1
  fi

  printf 'PostgreSQL 已就緒: %s:%s/%s\n' "${db_host}" "${db_port}" "${db_name}"
}

wait_for_server() {
  local url
  url="$(health_url)"
  for _ in $(seq 1 "${START_TIMEOUT}"); do
    if curl -sf "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  return 1
}

start_server_if_needed() {
  load_env_file
  mkdir -p "$(dirname "${SERVER_LOG}")"

  if curl -sf "$(health_url)" >/dev/null 2>&1; then
    print_step "偵測到後端已在執行"
    printf '目前使用既有服務: %s\n' "$(health_url)"
    return 0
  fi

  if [[ "${AUTO_START_SERVER}" != "true" ]]; then
    printf '後端尚未啟動，請先執行 go run ./http_service/cmd/server\n'
    exit 1
  fi

  print_step "自動啟動本機後端服務"
  printf '本腳本預設使用 PostgreSQL，會自動透過 docker compose 確保資料庫可用。\n'
  (
    cd "${ROOT_DIR}"
    set -a
    # shellcheck disable=SC1090
    source "${ENV_FILE}"
    set +a
    export GOTELEMETRY=off
    export GOPROXY=off
    export GOSUMDB=off
    export GOFLAGS="-mod=readonly -buildvcs=false"
    go run ./http_service/cmd/server > "${SERVER_LOG}" 2>&1
  ) &
  SERVER_PID="$!"

  if ! wait_for_server; then
    printf '後端啟動失敗，請查看日誌: %s\n' "${SERVER_LOG}"
    if [[ -f "${SERVER_LOG}" ]]; then
      cat "${SERVER_LOG}"
    fi
    exit 1
  fi

  printf '後端已啟動: %s\n' "$(health_url)"
}

cleanup() {
  if [[ -n "${SERVER_PID}" ]] && kill -0 "${SERVER_PID}" >/dev/null 2>&1; then
    kill "${SERVER_PID}" >/dev/null 2>&1 || true
    wait "${SERVER_PID}" 2>/dev/null || true
  fi
}

build_phone() {
  local seed="$1"
  printf '9%07d' "$((seed % 10000000))"
}

create_member() {
  local result_var="$1"
  local role="$2"
  local phone="$3"
  local display_name="$4"
  local community_id="$5"
  local country_code="+852"
  local otp_code
  local token

  call_api "${role} 申請 OTP" "POST" "/auth/otp/request" "$(cat <<EOF
{"phone_country_code":"${country_code}","phone_number":"${phone}","scene":"login"}
EOF
)"
  assert_status 200

  otp_code="$(extract_json "${RESPONSE}" "data.mock_code")"
  if [[ -z "${otp_code}" ]]; then
    otp_code="${OTP_MOCK_CODE:-123456}"
  fi

  call_api "${role} 驗證 OTP" "POST" "/auth/otp/verify" "$(cat <<EOF
{"phone_country_code":"${country_code}","phone_number":"${phone}","scene":"login","code":"${otp_code}"}
EOF
)"
  assert_status 200

  token="$(extract_json "${RESPONSE}" "data.access_token")"
  if [[ -z "${token}" ]]; then
    printf '%s 驗證 OTP 後未取得 access_token\n' "${role}"
    exit 1
  fi

  call_api "${role} 更新會員資料" "PATCH" "/me/profile" "$(cat <<EOF
{"display_name":"${display_name}","publisher_identity_type":"owner","primary_community_id":"${community_id}"}
EOF
)" "${token}"
  assert_status 200

  printf -v "${result_var}" '%s' "${token}"
}

create_media_asset() {
  local result_var="$1"
  local title="$2"
  local token="$3"
  local file_name="$4"
  local object_key media_asset_id upload_file upload_url content_type upload_status

  upload_file="$(mktemp)"
  dd if=/dev/zero of="${upload_file}" bs=2048 count=1 >/dev/null 2>&1

  call_api "${title} 申請上傳位址" "POST" "/oss/presign" "$(cat <<EOF
{"file_name":"${file_name}","mime_type":"image/webp","file_size":2048}
EOF
)" "${token}"
  assert_status 200
  object_key="$(extract_json "${RESPONSE}" "data.object_key")"
  upload_url="$(extract_json "${RESPONSE}" "data.upload_url")"
  content_type="$(extract_json "${RESPONSE}" "data.headers.Content-Type")"

  if [[ "${OSS_PROVIDER:-${STORAGE_PROVIDER:-mock}}" == "oss" ]]; then
    upload_status="$(curl -sS -o /dev/null -w '%{http_code}' -X PUT "${upload_url}" -H "Content-Type: ${content_type:-image/webp}" --data-binary @"${upload_file}")"
    if [[ "${upload_status}" != "200" ]]; then
      printf '上傳檔案到 OSS 失敗，HTTP 狀態: %s\n' "${upload_status}"
      rm -f "${upload_file}"
      exit 1
    fi
  fi

  call_api "${title} 登記媒體資產" "POST" "/oss/complete" "$(cat <<EOF
{"object_key":"${object_key}","mime_type":"image/webp","file_size":2048,"width":1200,"height":900,"checksum_sha256":"${file_name}-checksum"}
EOF
)" "${token}"
  assert_status 200
  media_asset_id="$(extract_json "${RESPONSE}" "data.media_asset_id")"
  rm -f "${upload_file}"

  printf -v "${result_var}" '%s' "${media_asset_id}"
}

create_listing() {
  local result_var="$1"
  local token="$2"
  local title="$3"
  local community_id="$4"
  local media_asset_id="$5"
  local visibility_scope="$6"
  local contact_method="$7"
  local listing_id

  call_api "建立 ${title} 草稿" "POST" "/secondhand/listings" "$(cat <<EOF
{
  "title":"${title}",
  "summary":"測試腳本建立的帖子",
  "description":"本筆資料由測試腳本自動建立。",
  "district_code":"kowloon",
  "community_id":"${community_id}",
  "publisher_identity_type":"owner",
  "category_code":"home_appliance",
  "price_mode":"fixed",
  "price_hkd":1200,
  "condition_level":"used_good",
  "dimension_text":"60 x 60 x 85 cm",
  "pickup_region_code":"kowloon",
  "pickup_location_text":"屋苑樓下自提",
  "delivery_tags":["self_pickup","elevator"],
  "visibility_scope":"${visibility_scope}",
  "contact_method":"${contact_method}",
  "images":[{"media_asset_id":"${media_asset_id}","sort_order":1,"is_cover":true}],
  "contact":{"show_phone":false,"show_whatsapp":true,"show_chat":true,"show_inquiry_form":false,"whatsapp":"+85291234567"}
}
EOF
)" "${token}"
  assert_status 200
  listing_id="$(extract_json "${RESPONSE}" "data.listing_id")"

  printf -v "${result_var}" '%s' "${listing_id}"
}

update_listing() {
  local token="$1"
  local listing_id="$2"
  local community_id="$3"
  local media_asset_id="$4"
  local title="$5"

  call_api "更新公開草稿" "PATCH" "/secondhand/listings/${listing_id}" "$(cat <<EOF
{
  "title":"${title}",
  "summary":"更新後摘要",
  "description":"更新後內容",
  "district_code":"kowloon",
  "community_id":"${community_id}",
  "publisher_identity_type":"owner",
  "category_code":"home_appliance",
  "price_mode":"fixed",
  "price_hkd":1100,
  "condition_level":"used_good",
  "dimension_text":"60 x 60 x 85 cm",
  "pickup_region_code":"kowloon",
  "pickup_location_text":"大堂交收",
  "delivery_tags":["self_pickup","elevator"],
  "visibility_scope":"public",
  "contact_method":"both",
  "images":[{"media_asset_id":"${media_asset_id}","sort_order":1,"is_cover":true}],
  "contact":{"show_phone":false,"show_whatsapp":true,"show_chat":true,"show_inquiry_form":false,"whatsapp":"+85291234567"}
}
EOF
)" "${token}"
  assert_status 200
}

publish_listing() {
  local title="$1"
  local token="$2"
  local listing_id="$3"
  call_api "${title} 發佈帖子" "POST" "/secondhand/listings/${listing_id}/publish" "" "${token}"
  assert_status 200
}

force_expire_listing() {
  local listing_id="$1"
  local db_user="${POSTGRES_USER:-postgres}"
  local db_name="${POSTGRES_DB:-ajoliving}"

  docker_compose exec -T "${POSTGRES_SERVICE}" psql -U "${db_user}" -d "${db_name}" \
    -c "UPDATE listings SET publication_status='expired', expire_at='2026-04-01T00:00:00Z', updated_at='2026-04-01T00:00:00Z' WHERE public_id='${listing_id}';" >/dev/null
}

usage() {
  cat <<'EOF'
用法:
  ./http_service/test/api.sh all

說明:
  1. 腳本會自動檢查 /api/v1/health。
  2. 腳本會先透過 docker compose 啟動本機 PostgreSQL。
  3. 若後端尚未啟動，預設會自動以本機 PostgreSQL 模式啟動服務。
  4. 測試結果會直接輸出到終端機。

可覆寫環境變數:
  BASE_URL
  API_PREFIX
  ENV_FILE
  COMPOSE_FILE
  POSTGRES_SERVICE
  AUTO_START_SERVER=true|false
  START_TIMEOUT=20
EOF
}

run_all() {
  local now_ts seed
  now_ts="$(date +%s)"
  seed="$((now_ts % 10000000))"
  OWNER_PHONE="$(build_phone "${seed}")"
  BUYER_PHONE="$(build_phone "$((seed + 1))")"
  OUTSIDER_PHONE="$(build_phone "$((seed + 2))")"

  call_api "1. 健康檢查" "GET" "/health"
  assert_status 200

  print_step "建立會員與社區資料"

  call_api "社區清單" "GET" "/meta/communities" "" "${OWNER_TOKEN:-invalid}"
  if [[ "${HTTP_STATUS}" != "200" ]]; then
    create_member OWNER_TOKEN "屋主" "${OWNER_PHONE}" "屋主測試帳號" ""
    call_api "屋主查詢社區清單" "GET" "/meta/communities" "" "${OWNER_TOKEN}"
    assert_status 200
  fi

  COMMUNITY_A="$(extract_json "${RESPONSE}" "data.items[0].public_id")"
  COMMUNITY_B="$(extract_json "${RESPONSE}" "data.items[1].public_id")"

  create_member OWNER_TOKEN "屋主" "${OWNER_PHONE}" "屋主測試帳號" "${COMMUNITY_A}"
  create_member BUYER_TOKEN "買家" "${BUYER_PHONE}" "買家測試帳號" "${COMMUNITY_A}"
  create_member OUTSIDER_TOKEN "外部會員" "${OUTSIDER_PHONE}" "外部測試帳號" "${COMMUNITY_B}"

  call_api "取得屋主會員資料" "GET" "/me" "" "${OWNER_TOKEN}"
  assert_status 200

  print_step "建立公開帖子"

  create_media_asset PUBLIC_MEDIA_ID "公開帖子" "${OWNER_TOKEN}" "public-cover.webp"
  create_listing PUBLIC_LISTING_ID "${OWNER_TOKEN}" "公開測試洗衣機" "${COMMUNITY_A}" "${PUBLIC_MEDIA_ID}" "public" "both"
  update_listing "${OWNER_TOKEN}" "${PUBLIC_LISTING_ID}" "${COMMUNITY_A}" "${PUBLIC_MEDIA_ID}" "公開測試洗衣機 - 已更新"
  publish_listing "公開帖子" "${OWNER_TOKEN}" "${PUBLIC_LISTING_ID}"

  call_api "頻道首頁摘要" "GET" "/channel-home/overview"
  assert_status 200
  call_api "公開列表查詢" "GET" "/secondhand/listings?page=1&page_size=20&sort_by=latest"
  assert_status 200
  call_api "公開詳情查詢" "GET" "/secondhand/listings/${PUBLIC_LISTING_ID}"
  assert_status 200
  call_api "屋主帖子列表" "GET" "/me/secondhand/listings?page=1&page_size=20&status=active" "" "${OWNER_TOKEN}"
  assert_status 200

  if force_expire_listing "${PUBLIC_LISTING_ID}"; then
    call_api "重新發佈過期公開帖子" "POST" "/secondhand/listings/${PUBLIC_LISTING_ID}/republish" "" "${OWNER_TOKEN}"
    assert_status 200
  else
    print_line
    printf '提示: 目前未執行 republish 自動測試，原因是 PostgreSQL 強制過期步驟執行失敗。\n'
  fi

  call_api "下架公開帖子" "POST" "/secondhand/listings/${PUBLIC_LISTING_ID}/deactivate" "" "${OWNER_TOKEN}"
  assert_status 200

  print_step "建立鄰里限定帖子與聊天流程"

  create_media_asset PRIVATE_MEDIA_ID "鄰里帖子" "${OWNER_TOKEN}" "private-cover.webp"
  create_listing PRIVATE_LISTING_ID "${OWNER_TOKEN}" "鄰里限定雪櫃" "${COMMUNITY_A}" "${PRIVATE_MEDIA_ID}" "building_only" "both"
  publish_listing "鄰里帖子" "${OWNER_TOKEN}" "${PRIVATE_LISTING_ID}"

  call_api "同社區會員查看鄰里詳情" "GET" "/secondhand/listings/${PRIVATE_LISTING_ID}" "" "${BUYER_TOKEN}"
  assert_status 200
  call_api "不同社區會員查看鄰里詳情" "GET" "/secondhand/listings/${PRIVATE_LISTING_ID}" "" "${OUTSIDER_TOKEN}"
  assert_status 403

  call_api "同社區會員取得聯絡方式" "POST" "/listings/${PRIVATE_LISTING_ID}/contact-access" "" "${BUYER_TOKEN}"
  assert_status 200

  call_api "建立或重用聊天" "POST" "/listings/${PRIVATE_LISTING_ID}/chats" "" "${BUYER_TOKEN}"
  assert_status 200
  CHAT_ID="$(extract_json "${RESPONSE}" "data.chat_id")"

  call_api "買家查詢聊天列表" "GET" "/chats?page=1&page_size=20" "" "${BUYER_TOKEN}"
  assert_status 200
  call_api "查詢聊天詳情" "GET" "/chats/${CHAT_ID}" "" "${OWNER_TOKEN}"
  assert_status 200
  call_api "查詢空訊息列表" "GET" "/chats/${CHAT_ID}/messages?page=1&page_size=20" "" "${BUYER_TOKEN}"
  assert_status 200

  call_api "買家傳送訊息" "POST" "/chats/${CHAT_ID}/messages" '{"content":"你好，請問仍可交易嗎？"}' "${BUYER_TOKEN}"
  assert_status 200
  call_api "屋主查詢聊天列表（未讀應為 1）" "GET" "/chats?page=1&page_size=20" "" "${OWNER_TOKEN}"
  assert_status 200
  call_api "屋主查詢訊息列表" "GET" "/chats/${CHAT_ID}/messages?page=1&page_size=20" "" "${OWNER_TOKEN}"
  assert_status 200
  call_api "屋主標記已讀" "POST" "/chats/${CHAT_ID}/read" "" "${OWNER_TOKEN}"
  assert_status 200
  call_api "屋主再次查詢聊天列表" "GET" "/chats?page=1&page_size=20" "" "${OWNER_TOKEN}"
  assert_status 200

  call_api "屋主標記鄰里帖子已售" "POST" "/secondhand/listings/${PRIVATE_LISTING_ID}/mark-sold" "" "${OWNER_TOKEN}"
  assert_status 200

  call_api "屋主登出" "POST" "/auth/logout" "" "${OWNER_TOKEN}"
  assert_status 200

  print_step "全部 API 測試完成"
}

main() {
  local cmd="${1:-all}"

  require_commands
  load_env_file
  trap cleanup EXIT

  case "${cmd}" in
    all)
      start_postgres_if_needed
      start_server_if_needed
      run_all
      ;;
    help|-h|--help)
      usage
      ;;
    *)
      usage
      exit 1
      ;;
  esac
}

main "$@"
