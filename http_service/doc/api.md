# AJO Living HTTP Service API 文件

## 本機啟動方式

### PostgreSQL 啟動

```bash
docker compose --env-file http_service/.env -f http_service/docker-compose.yml up -d postgres
```

### 後端服務啟動

```bash
set -a
source http_service/.env
set +a
go run ./http_service/cmd/server
```

### 一鍵測試腳本

```bash
./http_service/test/api.sh all
```

- 本文件預設資料庫為 `PostgreSQL`。
- 若本機尚未啟動資料庫，測試腳本會自動透過 `docker compose` 啟動 `postgres` 服務。
- 若需要建立第一個 `staff` 帳號，可在啟動前設定 `BOOTSTRAP_STAFF_PHONES=+85291238888` 這類完整手機號碼白名單。

## API 介面列表

### System 模組

| 編號 | 介面 | 方法 | 簡介/功能 | 權限 |
| --- | --- | --- | --- | --- |
| 1 | /api/v1/health | GET | 服務健康檢查 | 無 |

### Discovery 模組

| 編號 | 介面 | 方法 | 簡介/功能 | 權限 |
| --- | --- | --- | --- | --- |
| 2 | /api/v1/channel-home/overview | GET | 取得頻道首頁摘要 | 無 / 會員 |
| 3 | /api/v1/meta/communities | GET | 取得社區清單 | 會員 |

### Account / Auth 模組

| 編號 | 介面 | 方法 | 簡介/功能 | 權限 |
| --- | --- | --- | --- | --- |
| 4 | /api/v1/auth/otp/request | POST | 申請登入驗證碼 | 無 |
| 5 | /api/v1/auth/otp/verify | POST | 驗證 OTP 並登入 | 無 |
| 34 | /api/v1/auth/email/otp/request | POST | 申請 Email 登入驗證碼 | 無 |
| 35 | /api/v1/auth/email/otp/verify | POST | 驗證 Email OTP 並登入 | 無 |
| 49 | /api/v1/auth/email/register | POST | 建立郵箱與手機密碼帳戶 | 無 |
| 50 | /api/v1/auth/email/login | POST | 使用郵箱密碼登入 | 無 |
| 51 | /api/v1/auth/phone/login | POST | 使用手機密碼登入 | 無 |
| 6 | /api/v1/auth/logout | POST | 登出目前會員 | 會員 |
| 7 | /api/v1/me | GET | 取得目前會員資料 | 會員 |
| 8 | /api/v1/me/profile | PATCH | 更新會員資料 | 會員 |
| 9 | /api/v1/me/secondhand/listings | GET | 取得我的二手帖子 | 會員 |
| 66 | /api/v1/me/secondhand/favorites | GET | 取得我的二手收藏 | 會員 |
| 52 | /api/v1/me/wallet | GET | 取得 AJO Point 錢包總覽 | 會員 |
| 53 | /api/v1/me/wallet/transactions | GET | 查詢我的積分流水 | 會員 |
| 54 | /api/v1/me/wallet/ad-tasks | GET | 查詢可領取的廣告積分任務 | 會員 |
| 55 | /api/v1/me/wallet/ad-tasks/{taskId}/start | POST | 開始廣告觀看任務 | 會員 |
| 56 | /api/v1/me/wallet/ad-tasks/{taskId}/claim | POST | 領取廣告積分 | 會員 |
| 66 | /api/v1/me/wallet/ad-tasks/{taskId}/click | POST | 記錄廣告連結點擊 | 會員 |
| 33 | /api/v1/me/orders | GET | 取得我的訂單列表 | 會員 |

### OSS 模組

| 編號 | 介面 | 方法 | 簡介/功能 | 權限 |
| --- | --- | --- | --- | --- |
| 10 | /api/v1/oss/presign | POST | 申請圖片上傳位址 | 會員 |
| 11 | /api/v1/oss/complete | POST | 登記上傳完成媒體 | 會員 |
| 30 | /api/v1/oss/assets | GET | 查詢我的媒體資產列表 | 會員 |
| 31 | /api/v1/oss/assets/{mediaAssetId} | GET | 查詢單一媒體資產詳情 | 會員 |
| 32 | /api/v1/oss/assets/{mediaAssetId} | DELETE | 刪除未被帖子引用的媒體資產 | 會員 |

### Secondhand 模組

| 編號 | 介面 | 方法 | 簡介/功能 | 權限 |
| --- | --- | --- | --- | --- |
| 12 | /api/v1/secondhand/listings | GET | 查詢二手公開列表 | 無 / 會員 |
| 43 | /api/v1/secondhand/discover | GET | 查詢二手發現頁廣告位 | 無 / 會員 |
| 13 | /api/v1/secondhand/listings/{listingId} | GET | 查詢二手帖子詳情 | 無 / 會員 |
| 14 | /api/v1/secondhand/listings | POST | 建立二手草稿帖子 | 會員 |
| 15 | /api/v1/secondhand/listings/{listingId} | PATCH | 更新二手帖子內容 | 會員 |
| 16 | /api/v1/secondhand/listings/{listingId}/publish | POST | 發佈二手帖子 | 會員 |
| 17 | /api/v1/secondhand/listings/{listingId}/republish | POST | 重新發佈過期帖子 | 會員 |
| 72 | /api/v1/secondhand/listings/{listingId}/renew | POST | 續期上架中二手帖子 | 會員 |
| 18 | /api/v1/secondhand/listings/{listingId}/mark-sold | POST | 標記帖子為已售 | 會員 |
| 19 | /api/v1/secondhand/listings/{listingId}/deactivate | POST | 下架二手帖子 | 會員 |
| 67 | /api/v1/secondhand/listings/{listingId}/favorite | POST | 收藏二手帖子 | 會員 |
| 68 | /api/v1/secondhand/listings/{listingId}/favorite | DELETE | 取消收藏二手帖子 | 會員 |
| 20 | /api/v1/listings/{listingId}/contact-access | POST | 取得可聯絡方式 | 會員 |
| 44 | /api/v1/secondhand/settings/listings | GET | 設定頁查詢全部二手帖子 | Staff |
| 45 | /api/v1/secondhand/settings/discover-placements | GET | 設定頁查詢發現頁廣告位 | Staff |
| 46 | /api/v1/secondhand/settings/discover-placements | PUT | 設定頁儲存發現頁廣告位 | Staff |
| 47 | /api/v1/secondhand/settings/listings/{listingId}/mark-sold | POST | 設定頁標記任意帖子已售 | Staff |
| 48 | /api/v1/secondhand/settings/listings/{listingId}/deactivate | POST | 設定頁下架任意帖子 | Staff |

### Chat 模組

| 編號 | 介面 | 方法 | 簡介/功能 | 權限 |
| --- | --- | --- | --- | --- |
| 21 | /api/v1/listings/{listingId}/chats | POST | 建立或重用聊天 | 會員 |
| 22 | /api/v1/chats | GET | 查詢聊天會話列表 | 會員 |
| 23 | /api/v1/chats/{chatId} | GET | 查詢聊天會話詳情 | 會員 |
| 24 | /api/v1/chats/{chatId}/messages | GET | 查詢聊天訊息列表 | 會員 |
| 25 | /api/v1/chats/{chatId}/messages | POST | 傳送聊天訊息 | 會員 |
| 26 | /api/v1/chats/{chatId}/read | POST | 標記聊天已讀 | 會員 |

### Order 模組

| 編號 | 介面 | 方法 | 簡介/功能 | 權限 |
| --- | --- | --- | --- | --- |
| 34 | /api/v1/listings/{listingId}/orders | POST | 建立訂單請求 | 會員 |
| 35 | /api/v1/orders/{orderId} | GET | 查詢訂單詳情 | 會員 |
| 36 | /api/v1/orders/{orderId}/confirm | POST | 賣家確認訂單 | 會員 |
| 37 | /api/v1/orders/{orderId}/cancel | POST | 取消訂單 | 會員 |
| 38 | /api/v1/orders/{orderId}/complete | POST | 完成訂單 | 會員 |

### Notification 模組

| 編號 | 介面 | 方法 | 簡介/功能 | 權限 |
| --- | --- | --- | --- | --- |
| 39 | /api/v1/notifications | GET | 查詢我的通知列表 | 會員 |
| 40 | /api/v1/notifications/unread-count | GET | 查詢未讀通知數 | 會員 |
| 41 | /api/v1/notifications/{notificationId}/read | POST | 標記單一通知已讀 | 會員 |
| 42 | /api/v1/notifications/read-all | POST | 標記全部通知已讀 | 會員 |

### Staff 模組

| 編號 | 介面 | 方法 | 簡介/功能 | 權限 |
| --- | --- | --- | --- | --- |
| 27 | /api/v1/staff/me | GET | 取得目前 staff 帳號摘要 | Staff |
| 28 | /api/v1/staff/users | GET | 查詢會員與 staff 清單 | Staff |
| 29 | /api/v1/staff/users | POST | 新增管理員帳戶 | Staff |
| 30 | /api/v1/staff/users/{userId}/role | PATCH | 更新目標帳號角色 | Staff |
| 42 | /api/v1/staff/roles | GET | 查詢可用角色與權限矩陣 | Staff |
| 57 | /api/v1/staff/wallet/transactions | GET | 查詢平台積分流水 | Staff |
| 58 | /api/v1/staff/wallet/grants | POST | 手動發放 AJO Point | Staff |
| 59 | /api/v1/staff/wallet/reward-ads | GET | 查詢廣告積分任務 | Staff |
| 60 | /api/v1/staff/wallet/reward-ads | POST | 建立廣告積分任務 | Staff |
| 61 | /api/v1/staff/wallet/reward-ads/{taskId} | PATCH | 更新廣告積分任務 | Staff |
| 62 | /api/v1/staff/secondhand/listings | GET | 管理頁查詢二手帖子列表 | Staff |
| 63 | /api/v1/staff/secondhand/listings/{listingId}/publish | POST | 管理頁上架二手帖子 | Staff |
| 64 | /api/v1/staff/secondhand/listings/{listingId}/deactivate | POST | 管理頁下架二手帖子 | Staff |
| 65 | /api/v1/staff/secondhand/listings/{listingId}/renew | POST | 管理頁續期二手帖子 | Staff |
| 66 | /api/v1/staff/property-sales | GET | 管理頁查詢樓盤放售列表 | Staff |
| 67 | /api/v1/staff/property-sales/{listingId}/publish | POST | 管理頁上架樓盤放售 | Staff |
| 68 | /api/v1/staff/property-sales/{listingId}/deactivate | POST | 管理頁下架樓盤放售 | Staff |
| 69 | /api/v1/staff/property-sales/{listingId}/renew | POST | 管理頁續期樓盤放售 | Staff |
| 70 | /api/v1/staff/serviced-apartments | GET | 管理頁查詢服務式住宅列表 | Staff |
| 71 | /api/v1/staff/serviced-apartments/{listingId}/renew | POST | 管理頁續期服務式住宅 | Staff |

---

## 詳細介面參數/回應/腳本

## System 模組

### 1. /api/v1/health [GET]
- **簡介**: 服務健康檢查
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "status": "ok"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:08:46Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/health"
```
- **Powershell測試**
```powershell
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/health" -Method GET
```

---

## Discovery 模組

### 2. /api/v1/channel-home/overview [GET]
- **簡介**: 取得頻道首頁摘要
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "channels": [
      {
        "code": "property_sale",
        "title": "Property Sale",
        "description": "Future module placeholder"
      }
    ],
    "featured_secondhand": [
      {
        "listing_id": "01KSECONDHAND001",
        "title": "九成新洗衣機",
        "summary": "保養良好，可即日自提",
        "district_code": "kwun_tong",
        "category_code": "home_appliance",
        "price_mode": "fixed",
        "price_hkd": 1200,
        "condition_level": "used_good",
        "visibility_scope": "public",
        "contact_method": "both",
        "is_free_giveaway": false,
        "publisher_identity_type": "owner",
        "publication_status": "active",
        "business_status": "available",
        "updated_at": "2026-04-17T05:10:00Z"
      }
    ]
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:10:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/channel-home/overview"
```
- **Powershell測試**
```powershell
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/channel-home/overview" -Method GET
```

---

### 3. /api/v1/meta/communities [GET]
- **簡介**: 取得社區清單
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "public_id": "01KCOMMUNITY001",
        "community_type": "estate",
        "name_zh": "康怡花園",
        "name_en": "Kornhill",
        "district_code": "hk_east",
        "address_text": "Quarry Bay"
      }
    ]
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:11:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/meta/communities" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/meta/communities" -Method GET -Headers $headers
```

---

## Account / Auth 模組

### 4. /api/v1/auth/otp/request [POST]
- **簡介**: 申請登入驗證碼
- **請求參數**
```json
{
  "phone_country_code": "+852", // 必填
  "phone_number": "91234567", // 必填
  "scene": "login" // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "expires_in": 300,
    "mock_code": "123456"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:12:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/otp/request" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_country_code": "+852",
    "phone_number": "91234567",
    "scene": "login"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  phone_country_code="+852"
  phone_number="91234567"
  scene="login"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/otp/request" -Method POST -Headers $headers -Body $body
```

---

### 5. /api/v1/auth/otp/verify [POST]
- **簡介**: 驗證 OTP 並登入
- **請求參數**
```json
{
  "phone_country_code": "+852", // 必填
  "phone_number": "91234567", // 必填
  "scene": "login", // 可選
  "code": "123456" // 必填
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "access_token": "jwt-access-token",
    "refresh_token": "jwt-refresh-token",
    "expires_in": 7200,
    "user": {
      "public_id": "01KUSER001",
      "member_status": "active",
      "member_type": "user",
      "is_staff": false,
      "role": "user",
      "roles": ["member"],
      "permissions": [
        "account.profile.read",
        "account.profile.write",
        "listing.own.manage",
        "chat.use",
        "order.create",
        "order.own.manage",
        "notification.read"
      ],
      "profile_completed": false
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:13:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/otp/verify" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_country_code": "+852",
    "phone_number": "91234567",
    "scene": "login",
    "code": "123456"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  phone_country_code="+852"
  phone_number="91234567"
  scene="login"
  code="123456"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/otp/verify" -Method POST -Headers $headers -Body $body
```

---

### 34. /api/v1/auth/email/otp/request [POST]
- **簡介**: 申請 Email 登入驗證碼
- **請求參數**
```json
{
  "email": "member@example.com", // 必填
  "scene": "login" // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "expires_in": 300,
    "mock_code": "123456"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:14:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/email/otp/request" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "member@example.com",
    "scene": "login"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  email="member@example.com"
  scene="login"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/email/otp/request" -Method POST -Headers $headers -Body $body
```

---

### 35. /api/v1/auth/email/otp/verify [POST]
- **簡介**: 驗證 Email OTP 並登入
- **請求參數**
```json
{
  "email": "member@example.com", // 必填
  "scene": "login", // 可選
  "code": "123456", // 必填
  "display_name": "Email Member" // 可選，首次建立會員時使用
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "access_token": "jwt-access-token",
    "refresh_token": "jwt-refresh-token",
    "expires_in": 7200,
    "user": {
      "public_id": "01KUSER001",
      "member_status": "active",
      "member_type": "user",
      "is_staff": false,
      "role": "user",
      "roles": ["member"],
      "permissions": [
        "account.profile.read",
        "account.profile.write",
        "listing.own.manage",
        "chat.use",
        "order.create",
        "order.own.manage",
        "notification.read"
      ],
      "profile_completed": false
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:15:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/email/otp/verify" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "member@example.com",
    "scene": "login",
    "code": "123456",
    "display_name": "Email Member"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  email="member@example.com"
  scene="login"
  code="123456"
  display_name="Email Member"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/email/otp/verify" -Method POST -Headers $headers -Body $body
```

---

### 49. /api/v1/auth/email/register [POST]
- **簡介**: 建立郵箱與手機密碼帳戶
- **請求參數**
```json
{
  "email": "member@example.com", // 必填
  "password": "safe-password-123", // 必填，至少 8 個字元
  "display_name": "Email Member", // 必填
  "phone_country_code": "+852", // 必填
  "phone_number": "91234567" // 必填
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "access_token": "jwt-access-token",
    "refresh_token": "jwt-refresh-token",
    "expires_in": 7200,
    "user": {
      "public_id": "01KUSER001",
      "member_status": "active",
      "member_type": "user",
      "is_staff": false,
      "role": "user",
      "roles": ["member"],
      "permissions": [
        "account.profile.read",
        "account.profile.write",
        "listing.own.manage",
        "chat.use",
        "order.create",
        "order.own.manage",
        "notification.read"
      ],
      "profile_completed": false
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:15:30Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/email/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "member@example.com",
    "password": "safe-password-123",
    "display_name": "Email Member",
    "phone_country_code": "+852",
    "phone_number": "91234567"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  email="member@example.com"
  password="safe-password-123"
  display_name="Email Member"
  phone_country_code="+852"
  phone_number="91234567"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/email/register" -Method POST -Headers $headers -Body $body
```

---

### 50. /api/v1/auth/email/login [POST]
- **簡介**: 使用郵箱密碼登入
- **請求參數**
```json
{
  "email": "member@example.com", // 必填
  "password": "safe-password-123" // 必填
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "access_token": "jwt-access-token",
    "refresh_token": "jwt-refresh-token",
    "expires_in": 7200,
    "user": {
      "public_id": "01KUSER001",
      "member_status": "active",
      "member_type": "user",
      "is_staff": false,
      "role": "user",
      "roles": ["member"],
      "permissions": [
        "account.profile.read",
        "account.profile.write",
        "listing.own.manage",
        "chat.use",
        "order.create",
        "order.own.manage",
        "notification.read"
      ],
      "profile_completed": false
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:16:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/email/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "member@example.com",
    "password": "safe-password-123"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  email="member@example.com"
  password="safe-password-123"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/email/login" -Method POST -Headers $headers -Body $body
```

---

### 51. /api/v1/auth/phone/login [POST]
- **簡介**: 使用手機密碼登入
- **請求參數**
```json
{
  "phone_country_code": "+852", // 必填
  "phone_number": "91234567", // 必填
  "password": "safe-password-123" // 必填
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "access_token": "jwt-access-token",
    "refresh_token": "jwt-refresh-token",
    "expires_in": 7200,
    "user": {
      "public_id": "01KUSER001",
      "member_status": "active",
      "member_type": "user",
      "is_staff": false,
      "role": "user",
      "roles": ["member"],
      "permissions": [
        "account.profile.read",
        "account.profile.write",
        "listing.own.manage",
        "chat.use",
        "order.create",
        "order.own.manage",
        "notification.read"
      ],
      "profile_completed": false
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:16:30Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/phone/login" \
  -H "Content-Type: application/json" \
  -d '{
    "phone_country_code": "+852",
    "phone_number": "91234567",
    "password": "safe-password-123"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  phone_country_code="+852"
  phone_number="91234567"
  password="safe-password-123"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/phone/login" -Method POST -Headers $headers -Body $body
```

---

### 6. /api/v1/auth/logout [POST]
- **簡介**: 登出目前會員
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "logged_out": true
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:14:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/logout" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/logout" -Method POST -Headers $headers
```

---

### 7. /api/v1/me [GET]
- **簡介**: 取得目前會員資料
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "public_id": "01KUSER001",
    "phone_country_code": "+852",
    "phone_number": "91234567",
    "member_status": "active",
    "member_type": "user",
    "is_staff": false,
    "role": "user",
    "roles": ["member"],
    "permissions": [
      "account.profile.read",
      "account.profile.write",
      "listing.own.manage",
      "chat.use",
      "order.create",
      "order.own.manage",
      "notification.read"
    ],
    "display_name": "Neighbour User",
    "publisher_identity_type": "owner",
    "district_code": "hk_east",
    "primary_community": {
      "public_id": "01KCOMMUNITY001",
      "community_type": "estate",
      "name_zh": "康怡花園",
      "name_en": "Kornhill",
      "district_code": "hk_east",
      "address_text": "Quarry Bay"
    },
    "profile_completed": true
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:15:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/me" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/me" -Method GET -Headers $headers
```

---

### 8. /api/v1/me/profile [PATCH]
- **簡介**: 更新會員資料
- **請求參數**
```json
{
  "display_name": "Neighbour User", // 可選
  "publisher_identity_type": "owner", // 可選
  "primary_community_id": "01KCOMMUNITY001", // 可選
  "district_code": "hk_east" // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "public_id": "01KUSER001",
    "member_type": "user",
    "is_staff": false,
    "role": "user",
    "roles": ["member"],
    "permissions": [
      "account.profile.read",
      "account.profile.write",
      "listing.own.manage",
      "chat.use",
      "order.create",
      "order.own.manage",
      "notification.read"
    ],
    "display_name": "Neighbour User",
    "publisher_identity_type": "owner",
    "district_code": "hk_east",
    "profile_completed": true
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:16:00Z"
}
```
- **Curl測試**
```bash
curl -X PATCH "http://127.0.0.1:8080/api/v1/me/profile" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Neighbour User",
    "publisher_identity_type": "owner",
    "primary_community_id": "01KCOMMUNITY001",
    "district_code": "hk_east"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token";"Content-Type"="application/json"}
$body=@{
  display_name="Neighbour User"
  publisher_identity_type="owner"
  primary_community_id="01KCOMMUNITY001"
  district_code="hk_east"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/me/profile" -Method PATCH -Headers $headers -Body $body
```

---

### 9. /api/v1/me/secondhand/listings [GET]
- **簡介**: 取得我的二手帖子
- **請求參數**
```json
{
  "page": 1, // 可選
  "page_size": 20, // 可選
  "status": "active" // 可選，draft / active / expired / sold / hidden
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "listing_id": "01KSECONDHAND001",
        "title": "九成新洗衣機",
        "publication_status": "active",
        "business_status": "available",
        "updated_at": "2026-04-17T05:17:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:17:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/me/secondhand/listings?page=1&page_size=20&status=active" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/me/secondhand/listings?page=1&page_size=20&status=active" -Method GET -Headers $headers
```

---

### 33. /api/v1/me/orders [GET]
- **簡介**: 取得我的訂單列表
- **請求參數**
```json
{
  "page": 1, // 可選
  "page_size": 20, // 可選
  "status": "confirmed", // 可選
  "role": "buyer" // 可選，buyer / seller
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "order_id": "01KORDER001",
        "listing_id": "01KLISTING001",
        "listing_title": "九成新洗衣機",
        "order_status": "confirmed",
        "role_in_order": "buyer",
        "buyer_note": "Can pick up tonight.",
        "handover_method": "lobby_pickup",
        "cancel_reason": "",
        "created_at": "2026-04-17T05:40:00Z",
        "updated_at": "2026-04-17T05:45:00Z",
        "confirmed_at": "2026-04-17T05:45:00Z",
        "peer": {
          "public_id": "01KUSERSELLER001",
          "display_name": "Owner Seller",
          "role_in_order": "seller"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1
    }
  }
}
```

---

## OSS 模組

### 10. /api/v1/oss/presign [POST]
- **簡介**: 申請圖片上傳位址
- **請求參數**
```json
{
  "file_name": "cover.webp", // 必填
  "mime_type": "image/webp", // 必填
  "file_size": 182736 // 必填
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "upload_url": "https://ajo-living.oss-cn-hongkong.aliyuncs.com/ajo_living/01kupload.webp?x-oss-signature=example",
    "object_key": "ajo_living/01kupload.webp",
    "headers": {
      "Content-Type": "image/webp"
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:18:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/oss/presign" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "cover.webp",
    "mime_type": "image/webp",
    "file_size": 182736
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token";"Content-Type"="application/json"}
$body=@{
  file_name="cover.webp"
  mime_type="image/webp"
  file_size=182736
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/oss/presign" -Method POST -Headers $headers -Body $body
```

---

### 11. /api/v1/oss/complete [POST]
- **簡介**: 登記上傳完成媒體
- **請求參數**
```json
{
  "object_key": "ajo_living/01kupload.webp", // 必填
  "mime_type": "image/webp", // 必填
  "file_size": 182736, // 必填
  "width": 1200, // 可選
  "height": 900, // 可選
  "checksum_sha256": "abc123" // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "media_asset_id": "01KMEDIA001",
    "storage_provider": "oss",
    "bucket_name": "ajo-living",
    "object_key": "ajo_living/01kupload.webp",
    "mime_type": "image/webp",
    "width": 1200,
    "height": 900,
    "file_size": 182736,
    "checksum_sha256": "abc123",
    "url": "https://ajo-living.oss-cn-hongkong.aliyuncs.com/ajo_living/01kupload.webp",
    "in_use": false,
    "created_at": "2026-04-17T05:19:00Z"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:19:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/oss/complete" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "object_key": "ajo_living/01kupload.webp",
    "mime_type": "image/webp",
    "file_size": 182736,
    "width": 1200,
    "height": 900,
    "checksum_sha256": "abc123"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token";"Content-Type"="application/json"}
$body=@{
  object_key="ajo_living/01kupload.webp"
  mime_type="image/webp"
  file_size=182736
  width=1200
  height=900
  checksum_sha256="abc123"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/oss/complete" -Method POST -Headers $headers -Body $body
```

---

### 30. /api/v1/oss/assets [GET]
- **簡介**: 查詢我的媒體資產列表
- **請求參數**
```json
{
  "page": 1, // 可選
  "page_size": 20 // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "media_asset_id": "01KMEDIA001",
        "storage_provider": "oss",
        "bucket_name": "ajo-living",
        "object_key": "ajo_living/01kupload.webp",
        "mime_type": "image/webp",
        "width": 1200,
        "height": 900,
        "file_size": 182736,
        "checksum_sha256": "abc123",
        "url": "https://ajo-living.oss-cn-hongkong.aliyuncs.com/ajo_living/01kupload.webp",
        "in_use": false,
        "created_at": "2026-04-17T05:19:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:20:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/oss/assets?page=1&page_size=20" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/oss/assets?page=1&page_size=20" -Method GET -Headers $headers
```

---

### 31. /api/v1/oss/assets/{mediaAssetId} [GET]
- **簡介**: 查詢單一媒體資產詳情
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "media_asset_id": "01KMEDIA001",
    "storage_provider": "oss",
    "bucket_name": "ajo-living",
    "object_key": "ajo_living/01kupload.webp",
    "mime_type": "image/webp",
    "width": 1200,
    "height": 900,
    "file_size": 182736,
    "checksum_sha256": "abc123",
    "url": "https://ajo-living.oss-cn-hongkong.aliyuncs.com/ajo_living/01kupload.webp",
    "in_use": false,
    "created_at": "2026-04-17T05:19:00Z"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:21:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/oss/assets/01KMEDIA001" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/oss/assets/01KMEDIA001" -Method GET -Headers $headers
```

---

### 32. /api/v1/oss/assets/{mediaAssetId} [DELETE]
- **簡介**: 刪除未被帖子引用的媒體資產
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "media_asset_id": "01KMEDIA001",
    "storage_provider": "oss",
    "bucket_name": "ajo-living",
    "object_key": "ajo_living/01kupload.webp",
    "mime_type": "image/webp",
    "width": 1200,
    "height": 900,
    "file_size": 182736,
    "checksum_sha256": "abc123",
    "url": "https://ajo-living.oss-cn-hongkong.aliyuncs.com/ajo_living/01kupload.webp",
    "in_use": false,
    "created_at": "2026-04-17T05:19:00Z"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:22:00Z"
}
```
- **Curl測試**
```bash
curl -X DELETE "http://127.0.0.1:8080/api/v1/oss/assets/01KMEDIA001" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/oss/assets/01KMEDIA001" -Method DELETE -Headers $headers
```

---

## Secondhand 模組

### 12. /api/v1/secondhand/listings [GET]
- **簡介**: 查詢二手公開列表
- **請求參數**
```json
{
  "page": 1, // 可選
  "page_size": 20, // 可選
  "keyword": "洗衣機", // 可選
  "category_code": "home_appliance", // 可選
  "region_code": "kowloon", // 可選
  "district_code": "kwun_tong", // 可選
  "condition_level": "used_good", // 可選
  "price_mode": "fixed", // 可選
  "min_price_hkd": 100, // 可選
  "max_price_hkd": 2000, // 可選
  "only_building": false, // 可選
  "only_free": false, // 可選
  "sort_by": "latest" // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "listing_id": "01KSECONDHAND001",
        "title": "九成新洗衣機",
        "summary": "保養良好，可即日自提",
        "district_code": "kwun_tong",
        "category_code": "home_appliance",
        "price_mode": "fixed",
        "price_hkd": 1200,
        "condition_level": "used_good",
        "visibility_scope": "public",
        "contact_method": "both",
        "is_free_giveaway": false,
        "publisher_identity_type": "owner",
        "publication_status": "active",
        "business_status": "available",
        "is_favorited": true,
        "updated_at": "2026-04-17T05:20:00Z",
        "owner": {
          "user_id": "1001",
          "public_id": "01KUSERSELLER001",
          "display_name": "Owner Seller",
          "avatar_url": "http://localhost:9000/local-bucket/ajo_living/avatar.webp",
          "publisher_identity_type": "owner"
        },
        "cover_image": {
          "media_asset_id": "01KMEDIA001",
          "url": "http://localhost:9000/local-bucket/ajo_living/01kupload.webp",
          "sort_order": 1,
          "is_cover": true
        }
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:20:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/secondhand/listings?page=1&page_size=20&keyword=%E6%B4%97%E8%A1%A3%E6%A9%9F&sort_by=latest"
```
- **Powershell測試**
```powershell
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/secondhand/listings?page=1&page_size=20&keyword=%E6%B4%97%E8%A1%A3%E6%A9%9F&sort_by=latest" -Method GET
```

---

### 43. /api/v1/secondhand/discover [GET]
- **簡介**: 查詢二手發現頁人工配置廣告位。只返回仍然公開、已審核、未售出、未下架的帖子。
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "hero": [
      {
        "scene": "discover_hero",
        "category_code": "",
        "slot_index": 1,
        "listing": {
          "listing_id": "01KSECONDHAND001",
          "title": "九成新洗衣機",
          "category_code": "home_appliance",
          "publication_status": "active",
          "business_status": "available"
        }
      }
    ],
    "categories": {
      "home_appliance": [
        {
          "scene": "discover_category_carousel",
          "category_code": "home_appliance",
          "slot_index": 1,
          "listing": {
            "listing_id": "01KSECONDHAND001",
            "title": "九成新洗衣機",
            "category_code": "home_appliance"
          }
        }
      ]
    }
  }
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/secondhand/discover"
```
- **Powershell測試**
```powershell
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/secondhand/discover" -Method GET
```

---

### 13. /api/v1/secondhand/listings/{listingId} [GET]
- **簡介**: 查詢二手帖子詳情
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "title": "九成新洗衣機",
    "description": "使用正常，附基本保養資訊",
    "dimension_text": "60 x 60 x 85 cm",
    "pickup_region_code": "kwun_tong",
    "pickup_location_text": "屋苑樓下自提",
    "delivery_tags": ["self_pickup", "elevator"],
    "is_favorited": true,
    "owner": {
      "user_id": "1001",
      "public_id": "01KUSERSELLER001",
      "display_name": "Owner Seller",
      "avatar_url": "http://localhost:9000/local-bucket/ajo_living/avatar.webp",
      "publisher_identity_type": "owner"
    },
    "contact_summary": {
      "show_phone": false,
      "show_whatsapp": true,
      "show_chat": true,
      "show_inquiry_form": false
    },
    "images": [
      {
        "media_asset_id": "01KMEDIA001",
        "url": "http://localhost:9000/local-bucket/ajo_living/01kupload.webp",
        "sort_order": 1,
        "is_cover": true
      }
    ]
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:21:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001"
```
- **Powershell測試**
```powershell
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001" -Method GET
```

---

### 14. /api/v1/secondhand/listings [POST]
- **簡介**: 建立二手草稿帖子
- **請求參數**
```json
{
  "charge_draft": true, // 可選 Query，true 時儲存草稿扣 50 AJO Point
  "title": "九成新洗衣機", // 必填
  "summary": "保養良好，可即日自提", // 可選
  "description": "使用正常，附基本保養資訊", // 可選
  "district_code": "kwun_tong", // 必填
  "community_id": "01KCOMMUNITY001", // building_only 建議提供
  "publisher_identity_type": "owner", // 可選
  "category_code": "home_appliance", // 必填
  "price_mode": "fixed", // 必填
  "price_hkd": 1200, // free 時可省略
  "condition_level": "used_good", // 必填
  "dimension_text": "60 x 60 x 85 cm", // 可選
  "pickup_region_code": "kwun_tong", // 必填
  "pickup_location_text": "屋苑樓下自提", // 必填
  "delivery_tags": ["self_pickup", "elevator"], // 可選
  "visibility_scope": "public", // 必填
  "contact_method": "both", // 必填
  "images": [
    {
      "media_asset_id": "01KMEDIA001",
      "sort_order": 1,
      "is_cover": true
    }
  ],
  "contact": {
    "phone": "", // 可選
    "whatsapp": "+85291234567", // 可選
    "email": "", // 可選
    "show_phone": false,
    "show_whatsapp": true,
    "show_chat": true,
    "show_inquiry_form": false
  }
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "title": "九成新洗衣機",
    "publication_status": "draft",
    "business_status": "available",
    "points_charged": 50,
    "points_balance_after": 950
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:22:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/secondhand/listings?charge_draft=true" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "九成新洗衣機",
    "summary": "保養良好，可即日自提",
    "description": "使用正常，附基本保養資訊",
    "district_code": "kwun_tong",
    "community_id": "01KCOMMUNITY001",
    "publisher_identity_type": "owner",
    "category_code": "home_appliance",
    "price_mode": "fixed",
    "price_hkd": 1200,
    "condition_level": "used_good",
    "dimension_text": "60 x 60 x 85 cm",
    "pickup_region_code": "kwun_tong",
    "pickup_location_text": "屋苑樓下自提",
    "delivery_tags": ["self_pickup", "elevator"],
    "visibility_scope": "public",
    "contact_method": "both",
    "images": [{"media_asset_id": "01KMEDIA001", "sort_order": 1, "is_cover": true}],
    "contact": {"show_phone": false, "show_whatsapp": true, "show_chat": true, "show_inquiry_form": false, "whatsapp": "+85291234567"}
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token";"Content-Type"="application/json"}
$body=@{
  title="九成新洗衣機"
  summary="保養良好，可即日自提"
  description="使用正常，附基本保養資訊"
  district_code="kwun_tong"
  community_id="01KCOMMUNITY001"
  publisher_identity_type="owner"
  category_code="home_appliance"
  price_mode="fixed"
  price_hkd=1200
  condition_level="used_good"
  dimension_text="60 x 60 x 85 cm"
  pickup_region_code="kwun_tong"
  pickup_location_text="屋苑樓下自提"
  delivery_tags=@("self_pickup","elevator")
  visibility_scope="public"
  contact_method="both"
  images=@(@{media_asset_id="01KMEDIA001";sort_order=1;is_cover=$true})
  contact=@{show_phone=$false;show_whatsapp=$true;show_chat=$true;show_inquiry_form=$false;whatsapp="+85291234567"}
}|ConvertTo-Json -Depth 6
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/secondhand/listings?charge_draft=true" -Method POST -Headers $headers -Body $body
```

---

### 15. /api/v1/secondhand/listings/{listingId} [PATCH]
- **簡介**: 更新二手帖子內容
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001", // 路徑參數
  "charge_draft": true, // 可選 Query，草稿儲存時扣 50 AJO Point；已發佈帖子編輯仍按編輯費扣除
  "title": "九成新洗衣機 - 已更新", // 必填
  "summary": "已更新描述", // 可選
  "description": "更新後說明", // 可選
  "district_code": "kwun_tong", // 必填
  "category_code": "home_appliance", // 必填
  "price_mode": "fixed", // 必填
  "price_hkd": 1100, // 可選
  "condition_level": "used_good", // 必填
  "pickup_region_code": "kwun_tong", // 必填
  "pickup_location_text": "大堂自提", // 必填
  "visibility_scope": "public", // 必填
  "contact_method": "both" // 必填
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "title": "九成新洗衣機 - 已更新",
    "publication_status": "draft",
    "points_charged": 50,
    "points_balance_after": 900
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:23:00Z"
}
```
- **Curl測試**
```bash
curl -X PATCH "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001?charge_draft=true" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "九成新洗衣機 - 已更新",
    "summary": "已更新描述",
    "description": "更新後說明",
    "district_code": "kwun_tong",
    "community_id": "01KCOMMUNITY001",
    "publisher_identity_type": "owner",
    "category_code": "home_appliance",
    "price_mode": "fixed",
    "price_hkd": 1100,
    "condition_level": "used_good",
    "dimension_text": "60 x 60 x 85 cm",
    "pickup_region_code": "kwun_tong",
    "pickup_location_text": "大堂自提",
    "delivery_tags": ["self_pickup", "elevator"],
    "visibility_scope": "public",
    "contact_method": "both",
    "images": [{"media_asset_id": "01KMEDIA001", "sort_order": 1, "is_cover": true}],
    "contact": {"show_phone": false, "show_whatsapp": true, "show_chat": true, "show_inquiry_form": false, "whatsapp": "+85291234567"}
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token";"Content-Type"="application/json"}
$body=@{
  title="九成新洗衣機 - 已更新"
  summary="已更新描述"
  description="更新後說明"
  district_code="kwun_tong"
  community_id="01KCOMMUNITY001"
  publisher_identity_type="owner"
  category_code="home_appliance"
  price_mode="fixed"
  price_hkd=1100
  condition_level="used_good"
  dimension_text="60 x 60 x 85 cm"
  pickup_region_code="kwun_tong"
  pickup_location_text="大堂自提"
  delivery_tags=@("self_pickup","elevator")
  visibility_scope="public"
  contact_method="both"
  images=@(@{media_asset_id="01KMEDIA001";sort_order=1;is_cover=$true})
  contact=@{show_phone=$false;show_whatsapp=$true;show_chat=$true;show_inquiry_form=$false;whatsapp="+85291234567"}
}|ConvertTo-Json -Depth 6
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001?charge_draft=true" -Method PATCH -Headers $headers -Body $body
```

---

### 16. /api/v1/secondhand/listings/{listingId}/publish [POST]
- **簡介**: 發佈二手帖子
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "publication_status": "active",
    "business_status": "available",
    "expire_at": "2026-05-01T05:24:00Z"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:24:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001/publish" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001/publish" -Method POST -Headers $headers
```

---

### 17. /api/v1/secondhand/listings/{listingId}/republish [POST]
- **簡介**: 重新發佈過期帖子
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "publication_status": "active",
    "business_status": "available",
    "expire_at": "2026-05-01T05:25:00Z"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:25:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001/republish" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001/republish" -Method POST -Headers $headers
```

---

### 72. /api/v1/secondhand/listings/{listingId}/renew [POST]
- **簡介**: 續期上架中的二手帖子，扣除 50 AJO Point，並以目前時間重新計算 14 天到期時間。
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "publication_status": "active",
    "business_status": "available",
    "expire_at": "2026-05-01T05:25:00Z",
    "points_charged": 50,
    "points_balance_after": 950,
    "points_transaction_id": "01KTXRENEW001"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:25:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001/renew" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001/renew" -Method POST -Headers $headers
```

---

### 18. /api/v1/secondhand/listings/{listingId}/mark-sold [POST]
- **簡介**: 標記帖子為已售
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "business_status": "sold"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:26:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001/mark-sold" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001/mark-sold" -Method POST -Headers $headers
```

---

### 19. /api/v1/secondhand/listings/{listingId}/deactivate [POST]
- **簡介**: 下架二手帖子
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "publication_status": "hidden"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:27:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001/deactivate" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/secondhand/listings/01KSECONDHAND001/deactivate" -Method POST -Headers $headers
```

---

### 66. /api/v1/me/secondhand/favorites [GET]
- **簡介**: 取得我的二手收藏列表。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20
}
```
- **回應參數**: 與 `/api/v1/secondhand/listings` 列表項相同，`is_favorited` 固定為 `true`。
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/me/secondhand/favorites?page=1&page_size=20" \
  -H "Authorization: Bearer $token"
```

---

### 67. /api/v1/secondhand/listings/{listingId}/favorite [POST]
- **簡介**: 收藏一個目前可見的二手帖子。
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "is_favorited": true
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:28:00Z"
}
```

---

### 68. /api/v1/secondhand/listings/{listingId}/favorite [DELETE]
- **簡介**: 取消收藏二手帖子。
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "is_favorited": false
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:29:00Z"
}
```

---

### 44. /api/v1/secondhand/settings/listings [GET]
- **簡介**: 設定頁查詢全部二手帖子，支援關鍵字、分類與狀態篩選。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20,
  "keyword": "洗衣機",
  "category_code": "home_appliance",
  "status": "active"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "listing_id": "01KSECONDHAND001",
        "title": "九成新洗衣機",
        "category_code": "home_appliance",
        "publication_status": "active",
        "business_status": "available"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1
    }
  }
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/secondhand/settings/listings?page=1&page_size=20&status=active" \
  -H "Authorization: Bearer $token"
```

---

### 45. /api/v1/secondhand/settings/discover-placements [GET]
- **簡介**: 設定頁查詢完整發現頁廣告位矩陣。封面大推固定 4 個位置，每個分類固定 10 個位置。
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "hero": [
      {
        "scene": "discover_hero",
        "category_code": "",
        "slot_index": 1,
        "listing": {
          "listing_id": "01KSECONDHAND001",
          "title": "九成新洗衣機"
        }
      }
    ],
    "categories": {
      "home_appliance": [
        {
          "scene": "discover_category_carousel",
          "category_code": "home_appliance",
          "slot_index": 1
        }
      ]
    }
  }
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/secondhand/settings/discover-placements" \
  -H "Authorization: Bearer $token"
```

---

### 46. /api/v1/secondhand/settings/discover-placements [PUT]
- **簡介**: 設定頁批量儲存發現頁廣告位。`listing_id` 留空會清空指定位置。
- **請求參數**
```json
{
  "placements": [
    {
      "scene": "discover_hero",
      "category_code": "",
      "slot_index": 1,
      "listing_id": "01KSECONDHAND001"
    },
    {
      "scene": "discover_category_carousel",
      "category_code": "home_appliance",
      "slot_index": 1,
      "listing_id": "01KSECONDHAND001"
    }
  ]
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "hero": [
      {
        "scene": "discover_hero",
        "category_code": "",
        "slot_index": 1,
        "listing": {
          "listing_id": "01KSECONDHAND001",
          "title": "九成新洗衣機"
        }
      }
    ],
    "categories": {
      "home_appliance": [
        {
          "scene": "discover_category_carousel",
          "category_code": "home_appliance",
          "slot_index": 1,
          "listing": {
            "listing_id": "01KSECONDHAND001",
            "title": "九成新洗衣機"
          }
        }
      ]
    }
  }
}
```
- **Curl測試**
```bash
curl -X PUT "http://127.0.0.1:8080/api/v1/secondhand/settings/discover-placements" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{"placements":[{"scene":"discover_hero","category_code":"","slot_index":1,"listing_id":"01KSECONDHAND001"}]}'
```

---

### 47. /api/v1/secondhand/settings/listings/{listingId}/mark-sold [POST]
- **簡介**: 設定頁標記任意二手帖子為已售。
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "business_status": "sold"
  }
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/secondhand/settings/listings/01KSECONDHAND001/mark-sold" \
  -H "Authorization: Bearer $token"
```

---

### 48. /api/v1/secondhand/settings/listings/{listingId}/deactivate [POST]
- **簡介**: 設定頁下架任意二手帖子。
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "publication_status": "hidden"
  }
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/secondhand/settings/listings/01KSECONDHAND001/deactivate" \
  -H "Authorization: Bearer $token"
```

---

### 20. /api/v1/listings/{listingId}/contact-access [POST]
- **簡介**: 取得可聯絡方式
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND002" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND002",
    "allowed_channels": {
      "phone": false,
      "whatsapp": true,
      "chat": true,
      "inquiry_form": false
    },
    "contact_payload": {
      "whatsapp_url": "https://wa.me/85291234567?text=I%20am%20interested..."
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:28:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/listings/01KSECONDHAND002/contact-access" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/listings/01KSECONDHAND002/contact-access" -Method POST -Headers $headers
```

---

## Chat 模組

### 21. /api/v1/listings/{listingId}/chats [POST]
- **簡介**: 建立或重用聊天
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND002" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "chat_id": "01KCHAT001",
    "is_new": true
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:29:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/listings/01KSECONDHAND002/chats" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/listings/01KSECONDHAND002/chats" -Method POST -Headers $headers
```

---

### 22. /api/v1/chats [GET]
- **簡介**: 查詢聊天會話列表
- **請求參數**
```json
{
  "page": 1, // 可選
  "page_size": 20 // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "chat_id": "01KCHAT001",
        "listing_id": "01KSECONDHAND002",
        "listing_title": "九成新洗衣機",
        "chat_type": "direct_listing_chat",
        "last_message_preview": "你好，請問仍可交易嗎？",
        "last_message_at": "2026-04-17T05:30:00Z",
        "unread_count": 1
      },
      {
        "chat_id": "01KCHATNOTICE001",
        "listing_id": "",
        "listing_title": "通知",
        "chat_type": "system_notice",
        "last_message_preview": "最高100幣，可以當錢花",
        "last_message_at": "2026-04-17T05:28:00Z",
        "unread_count": 4
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:30:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/chats?page=1&page_size=20" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/chats?page=1&page_size=20" -Method GET -Headers $headers
```

---

### 23. /api/v1/chats/{chatId} [GET]
- **簡介**: 查詢聊天會話詳情
- **請求參數**
```json
{
  "chatId": "01KCHAT001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "chat_id": "01KCHAT001",
    "listing_id": "01KSECONDHAND002",
    "listing_title": "九成新洗衣機",
    "chat_type": "direct_listing_chat",
    "created_at": "2026-04-17T05:29:00Z",
    "participants": [
      {
        "user_id": "1",
        "role_in_chat": "owner"
      },
      {
        "user_id": "2",
        "role_in_chat": "buyer"
      }
    ]
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:31:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/chats/01KCHAT001" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/chats/01KCHAT001" -Method GET -Headers $headers
```

---

### 24. /api/v1/chats/{chatId}/messages [GET]
- **簡介**: 查詢聊天訊息列表
- **請求參數**
```json
{
  "chatId": "01KCHAT001", // 路徑參數
  "page": 1, // 可選
  "page_size": 20 // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "message_id": "01KMESSAGE001",
        "sender_user_id": "2",
        "content": "你好，請問仍可交易嗎？",
        "message_type": "text",
        "status": "sent",
        "created_at": "2026-04-17T05:32:00Z"
      },
      {
        "message_id": "01KNOTICE001",
        "sender_user_id": "1",
        "content": "紅包到賬提醒\n拼手氣，瓜分 HK$35999 現金紅包",
        "message_type": "notice_card",
        "action_label": "去查看",
        "action_url": "/marketplace/discover",
        "status": "sent",
        "created_at": "2026-04-17T05:28:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:32:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/chats/01KCHAT001/messages?page=1&page_size=20" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/chats/01KCHAT001/messages?page=1&page_size=20" -Method GET -Headers $headers
```

---

### 25. /api/v1/chats/{chatId}/messages [POST]
- **簡介**: 傳送聊天訊息
- **請求參數**
```json
{
  "chatId": "01KCHAT001", // 路徑參數
  "content": "你好，請問仍可交易嗎？" // 必填
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "message_id": "01KMESSAGE001",
    "sender_user_id": "2",
    "content": "你好，請問仍可交易嗎？",
    "message_type": "text",
    "status": "sent",
    "created_at": "2026-04-17T05:33:00Z"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:33:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/chats/01KCHAT001/messages" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "你好，請問仍可交易嗎？"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token";"Content-Type"="application/json"}
$body=@{
  content="你好，請問仍可交易嗎？"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/chats/01KCHAT001/messages" -Method POST -Headers $headers -Body $body
```

---

### 26. /api/v1/chats/{chatId}/read [POST]
- **簡介**: 標記聊天已讀
- **請求參數**
```json
{
  "chatId": "01KCHAT001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "chat_id": "01KCHAT001",
    "read": true
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:34:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/chats/01KCHAT001/read" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/chats/01KCHAT001/read" -Method POST -Headers $headers
```

---

## Staff 模組

## Wallet 模組

### 52. /api/v1/me/wallet [GET]
- **簡介**: 取得目前會員 AJO Point 錢包總覽、扣費規則與最近流水。
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "account": {
      "balance": 1200,
      "total_earned": 1500,
      "total_spent": 300,
      "today_ad_reward_points": 50,
      "daily_ad_reward_limit": 1000
    },
    "recent_transactions": [
      {
        "transaction_id": "01KTX001",
        "direction": "debit",
        "amount": 100,
        "balance_before": 1300,
        "balance_after": 1200,
        "source_type": "listing_charge",
        "biz_module": "secondhand",
        "action_type": "publish",
        "note": "secondhand publish",
        "created_at": "2026-04-17T06:00:00Z"
      }
    ],
    "charge_rules": [
      { "biz_module": "secondhand", "label": "二手交易", "publish": 100, "draft_save": 50, "edit": 100, "republish": 100, "renew": 50 },
      { "biz_module": "property_sale", "label": "樓盤放售", "publish": 1000, "edit": 1000, "republish": 1000 },
      { "biz_module": "serviced_apartment", "label": "服務式住宅", "publish": 800, "edit": 800, "republish": 800 }
    ],
    "recharge_enabled": false
  }
}
```

---

### 53. /api/v1/me/wallet/transactions [GET]
- **簡介**: 分頁查詢目前會員積分流水。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "transaction_id": "01KTX001",
        "direction": "credit",
        "amount": 50,
        "balance_before": 100,
        "balance_after": 150,
        "source_type": "reward_ad",
        "biz_module": "wallet",
        "action_type": "ad_reward",
        "note": "rewarded ad completed",
        "created_at": "2026-04-17T06:10:00Z"
      }
    ],
    "pagination": { "page": 1, "page_size": 20, "total": 1 }
  }
}
```

---

### 54. /api/v1/me/wallet/ad-tasks [GET]
- **簡介**: 查詢目前可用的看廣告得積分任務。後端會返回每日積分上限、任務預算與廣告統計狀態；同一廣告可多次觀看，直到每日總積分上限或任務預算用完。
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "task_id": "01KAD001",
        "title": "AJO Living Reward",
        "summary": "Watch official video.",
        "cover_url": "https://cdn.example.com/ad.png",
        "media_url": "https://cdn.example.com/ad.mp4",
        "target_url": "https://www.ajoliving.com",
        "reward_points": 50,
        "watch_seconds": 30,
        "can_claim_today": true,
        "claimed_today": false,
        "remaining_budget": 950,
        "watch_count": 120,
        "total_watch_seconds": 3600,
        "link_click_count": 18,
        "link_click_rate": 15
      }
    ]
  }
}
```

---

### 55. /api/v1/me/wallet/ad-tasks/{taskId}/start [POST]
- **簡介**: 開始廣告觀看任務，建立服務端計時會話，並原子增加廣告觀看次數。
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "claim_id": "01KCLAIM001",
    "task_id": "01KAD001",
    "watch_seconds": 30,
    "started_at": "2026-04-17T06:15:00Z",
    "available_at": "2026-04-17T06:15:30Z",
    "reward_points": 50
  }
}
```

---

### 56. /api/v1/me/wallet/ad-tasks/{taskId}/claim [POST]
- **簡介**: 領取廣告積分。後端會校驗觀看時間、每日總積分上限與任務預算，成功後原子增加總觀看時間與已發積分。
- **請求參數**
```json
{
  "claim_id": "01KCLAIM001"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "points_charged": 50,
    "points_balance_after": 150,
    "points_transaction_id": "01KTX002"
  }
}
```

---

### 66. /api/v1/me/wallet/ad-tasks/{taskId}/click [POST]
- **簡介**: 記錄廣告跳轉連結點擊，原子增加點擊次數並返回最新點擊率。
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "task_id": "01KAD001",
    "target_url": "https://www.ajoliving.com",
    "watch_count": 120,
    "link_click_count": 19,
    "link_click_rate": 15.8333333333
  }
}
```

---

### 27. /api/v1/staff/me [GET]
- **簡介**: 取得目前 staff 帳號摘要
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "public_id": "01KUSERSTAFF001",
    "phone_country_code": "+852",
    "phone_number": "91238888",
    "member_status": "active",
    "member_type": "user",
    "is_staff": true,
    "role": "staff",
    "roles": ["super_admin", "member"],
    "permissions": [
      "account.profile.read",
      "account.profile.write",
      "listing.own.manage",
      "chat.use",
      "order.create",
      "order.own.manage",
      "notification.read",
      "staff.console.access",
      "staff.user.read",
      "staff.user.manage",
      "staff.role.read",
      "staff.role.manage",
      "staff.review.manage",
      "staff.support.manage"
    ],
    "display_name": "Staff Operator",
    "publisher_identity_type": "owner",
    "district_code": "hk_east",
    "primary_community": {
      "public_id": "01KCOMMUNITY001",
      "community_type": "estate",
      "name_zh": "康怡花園",
      "name_en": "Kornhill",
      "district_code": "hk_east",
      "address_text": "Quarry Bay"
    },
    "created_at": "2026-04-17T06:00:00Z",
    "updated_at": "2026-04-17T06:00:00Z"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T06:00:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/staff/me" \
  -H "Authorization: Bearer $staff_token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $staff_token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/staff/me" -Method GET -Headers $headers
```

---

### 28. /api/v1/staff/users [GET]
- **簡介**: 查詢會員與 staff 清單
- **請求參數**
```json
{
  "page": 1, // 可選
  "page_size": 20, // 可選
  "keyword": "9123", // 可選
  "status": "active", // 可選
  "member_type": "pro_user", // 可選，user / pro_user
  "role_code": "member", // 可選
  "is_staff": false // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "public_id": "01KUSERPRO001",
        "email": "pro@example.com",
        "phone_country_code": "+852",
        "phone_number": "91239999",
        "member_status": "active",
        "member_type": "pro_user",
        "is_staff": false,
        "role": "pro_user",
        "roles": ["member"],
        "permissions": [
          "account.profile.read",
          "account.profile.write",
          "listing.own.manage",
          "chat.use",
          "order.create",
          "order.own.manage",
          "notification.read"
        ],
        "display_name": "Pro Seller",
        "publisher_identity_type": "owner",
        "district_code": "hk_east",
        "created_at": "2026-04-17T06:10:00Z",
        "updated_at": "2026-04-17T06:10:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T06:10:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/staff/users?page=1&page_size=20&member_type=pro_user&role_code=member" \
  -H "Authorization: Bearer $staff_token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $staff_token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/staff/users?page=1&page_size=20&member_type=pro_user&role_code=member" -Method GET -Headers $headers
```

---

### 29. /api/v1/staff/users [POST]
- **簡介**: 新增可登入的管理員帳戶，並套用指定 staff 角色。
- **請求參數**
```json
{
  "email": "ops@example.com",
  "password": "staffpass123",
  "display_name": "Operations Staff",
  "phone_country_code": "+852",
  "phone_number": "91237777",
  "member_type": "user",
  "role_codes": ["staff"]
}
```
- `role_codes` 必須包含 staff 範圍角色，例如 `staff` 或 `super_admin`。
- **回應參數**: 與 `/api/v1/staff/users` 列表項相同。

---

### 30. /api/v1/staff/users/{userId}/role [PATCH]
- **簡介**: 更新目標帳號角色
- **請求參數**
```json
{
  "userId": "01KUSERPRO001", // 路徑參數
  "member_type": "pro_user", // 可選，user / pro_user
  "role_codes": ["member"], // 可選
  "is_staff": false // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "public_id": "01KUSERPRO001",
    "member_status": "active",
    "member_type": "pro_user",
    "is_staff": false,
    "role": "pro_user",
    "roles": ["member"],
    "permissions": [
      "account.profile.read",
      "account.profile.write",
      "listing.own.manage",
      "chat.use",
      "order.create",
      "order.own.manage",
      "notification.read"
    ],
    "updated_at": "2026-04-17T06:20:00Z"
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T06:20:00Z"
}
```
- **Curl測試**
```bash
curl -X PATCH "http://127.0.0.1:8080/api/v1/staff/users/01KUSERPRO001/role" \
  -H "Authorization: Bearer $staff_token" \
  -H "Content-Type: application/json" \
  -d '{
    "member_type": "pro_user",
    "role_codes": ["member"],
    "is_staff": false
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $staff_token";"Content-Type"="application/json"}
$body=@{
  member_type="pro_user"
  role_codes=@("member")
  is_staff=$false
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/staff/users/01KUSERPRO001/role" -Method PATCH -Headers $headers -Body $body
```

---

### 42. /api/v1/staff/roles [GET]
- **簡介**: 查詢可用角色與權限矩陣
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "code": "super_admin",
        "scope": "staff",
        "name": "Super Admin",
        "description": "Full back-office access.",
        "permissions": [
          "staff.console.access",
          "staff.user.read",
          "staff.user.manage",
          "staff.role.read",
          "staff.role.manage",
          "staff.review.manage",
          "staff.support.manage"
        ]
      },
      {
        "code": "staff",
        "scope": "staff",
        "name": "Staff",
        "description": "Day-to-day back-office access.",
        "permissions": [
          "staff.console.access",
          "staff.user.read",
          "staff.user.manage",
          "staff.role.read",
          "staff.review.manage",
          "staff.support.manage"
        ]
      },
      {
        "code": "member",
        "scope": "member",
        "name": "Member",
        "description": "Baseline member access.",
        "permissions": [
          "account.profile.read",
          "account.profile.write",
          "listing.own.manage",
          "chat.use",
          "order.create",
          "order.own.manage",
          "notification.read"
        ]
      }
    ]
  }
}
```

---

### 57. /api/v1/staff/wallet/transactions [GET]
- **簡介**: Staff 查詢平台積分流水，可按會員、方向、來源與業務模組篩選。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20,
  "user_id": "01KUSER001",
  "direction": "credit",
  "source_type": "operator_grant",
  "biz_module": "wallet"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "transaction_id": "01KTX003",
        "direction": "credit",
        "amount": 300,
        "balance_before": 0,
        "balance_after": 300,
        "source_type": "operator_grant",
        "biz_module": "wallet",
        "action_type": "operator_grant",
        "note": "Launch credit",
        "target_user": {
          "user_id": "01KUSER001",
          "phone_country_code": "+852",
          "phone_number": "91239999",
          "display_name": "Member"
        },
        "operator_user": {
          "user_id": "01KSTAFF001",
          "phone_country_code": "+852",
          "phone_number": "91238888",
          "display_name": "Staff Operator"
        },
        "created_at": "2026-04-17T06:30:00Z"
      }
    ],
    "pagination": { "page": 1, "page_size": 20, "total": 1 }
  }
}
```

---

### 58. /api/v1/staff/wallet/grants [POST]
- **簡介**: Staff 手動發放 AJO Point。必須提供發放原因，並寫入不可變積分流水。
- **請求參數**
```json
{
  "user_id": "01KUSER001",
  "amount": 300,
  "note": "Launch credit"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "charge": {
      "points_charged": 300,
      "points_balance_after": 300,
      "points_transaction_id": "01KTX003"
    },
    "target_user": { "user_id": "01KUSER001" },
    "operator": { "user_id": "01KSTAFF001" }
  }
}
```

---

### 59. /api/v1/staff/wallet/reward-ads [GET]
- **簡介**: Staff 查詢看廣告得積分任務。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20,
  "keyword": "AJO",
  "is_active": true
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "task_id": "01KAD001",
        "title": "AJO Living Reward",
        "summary": "Watch official video.",
        "cover_url": "https://cdn.example.com/ad.png",
        "media_url": "https://cdn.example.com/ad.mp4",
        "target_url": "https://www.ajoliving.com",
        "reward_points": 50,
        "watch_seconds": 30,
        "total_budget": 1000,
        "total_granted": 50,
        "remaining_budget": 950,
        "watch_count": 120,
        "total_watch_seconds": 3600,
        "link_click_count": 18,
        "link_click_rate": 15,
        "is_active": true,
        "created_at": "2026-04-17T06:00:00Z",
        "updated_at": "2026-04-17T06:00:00Z"
      }
    ],
    "pagination": { "page": 1, "page_size": 20, "total": 1 }
  }
}
```

---

### 60. /api/v1/staff/wallet/reward-ads [POST]
- **簡介**: Staff 建立看廣告得積分任務。觀看秒數低於 30 會被後端提升為 30。
- **請求參數**
```json
{
  "title": "AJO Living Reward",
  "summary": "Watch official video.",
  "cover_url": "https://cdn.example.com/ad.png",
  "media_url": "https://cdn.example.com/ad.mp4",
  "target_url": "https://www.ajoliving.com",
  "reward_points": 50,
  "watch_seconds": 30,
  "total_budget": 1000,
  "is_active": true,
  "starts_at": "2026-04-17T06:00:00Z",
  "ends_at": "2026-05-17T06:00:00Z"
}
```

---

### 61. /api/v1/staff/wallet/reward-ads/{taskId} [PATCH]
- **簡介**: Staff 更新看廣告得積分任務，可用於修改素材、預算或啟停任務。
- **請求參數**
```json
{
  "title": "AJO Living Reward",
  "reward_points": 80,
  "watch_seconds": 30,
  "total_budget": 2000,
  "is_active": false
}
```

---

### 62. /api/v1/staff/secondhand/listings [GET]
- **簡介**: Staff 管理頁分頁查詢二手帖子，支援 keyword 與 status 篩選。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20,
  "keyword": "sofa",
  "status": "active"
}
```
- **回應參數**: 與 `/api/v1/secondhand/settings/listings` 相同。

---

### 63-65. /api/v1/staff/secondhand/listings/{listingId}/{action} [POST]
- **簡介**: Staff 對二手帖子執行狀態操作。`action` 支援 `publish`、`deactivate`、`renew`。
- **請求參數**
```json
{
  "listingId": "01KSECONDHAND001",
  "renewal_days": 14
}
```
- `renewal_days` 僅 `renew` 使用，未傳時預設 14 天，允許 1 至 365 天。
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSECONDHAND001",
    "publication_status": "active",
    "business_status": "available"
  }
}
```

---

### 66. /api/v1/staff/property-sales [GET]
- **簡介**: Staff 管理頁分頁查詢樓盤放售，支援 keyword 與 status 篩選。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20,
  "keyword": "Central",
  "status": "active"
}
```
- **回應參數**: 與 `/api/v1/property-sales` 列表項結構一致，管理頁可見全部狀態。

---

### 67-69. /api/v1/staff/property-sales/{listingId}/{action} [POST]
- **簡介**: Staff 對樓盤放售執行狀態操作。`action` 支援 `publish`、`deactivate`、`renew`。
- **請求參數**
```json
{
  "listingId": "01KPROPERTY001"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KPROPERTY001",
    "publication_status": "active",
    "business_status": "available"
  }
}
```

---

### 70. /api/v1/staff/serviced-apartments [GET]
- **簡介**: Staff 管理頁分頁查詢服務式住宅，支援 keyword 與 status 篩選。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20,
  "keyword": "suite",
  "status": "active"
}
```
- **回應參數**: 與 `/api/v1/serviced-apartments` 列表項結構一致，管理頁可見全部狀態。

---

### 71. /api/v1/staff/serviced-apartments/{listingId}/renew [POST]
- **簡介**: Staff 續期服務式住宅。
- **請求參數**
```json
{
  "listingId": "01KSERVICED001"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01KSERVICED001",
    "publication_status": "active",
    "business_status": "available"
  }
}
```

---

## Order 模組

### 34. /api/v1/listings/{listingId}/orders [POST]
- **簡介**: 建立訂單請求
- **請求參數**
```json
{
  "listingId": "01KLISTING001", // 路徑參數
  "buyer_note": "Can pick up tonight.", // 可選
  "handover_method": "lobby_pickup" // 可選，face_to_face / lobby_pickup / courier
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "order_id": "01KORDER001",
    "listing_id": "01KLISTING001",
    "listing_title": "九成新洗衣機",
    "order_status": "pending_confirm",
    "role_in_order": "buyer",
    "buyer_note": "Can pick up tonight.",
    "handover_method": "lobby_pickup",
    "cancel_reason": "",
    "created_at": "2026-04-17T05:40:00Z",
    "updated_at": "2026-04-17T05:40:00Z",
    "cover_image": {
      "media_asset_id": "01KMEDIA001",
      "url": "http://localhost:9000/local-bucket/ajo_living/01kupload.webp",
      "sort_order": 1,
      "is_cover": true
    },
    "peer": {
      "public_id": "01KUSERSELLER001",
      "display_name": "Owner Seller",
      "role_in_order": "seller"
    },
    "logs": [
      {
        "action_type": "create_order",
        "from_status": "",
        "to_status": "pending_confirm",
        "operator_user_id": "10001",
        "note": "Can pick up tonight.",
        "created_at": "2026-04-17T05:40:00Z"
      }
    ]
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:40:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/listings/01KLISTING001/orders" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "buyer_note": "Can pick up tonight.",
    "handover_method": "lobby_pickup"
  }'
```
- **Powershell測試**
```powershell
$headers=@{
  "Authorization"="Bearer $token"
  "Content-Type"="application/json"
}
$body=@{
  buyer_note="Can pick up tonight."
  handover_method="lobby_pickup"
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/listings/01KLISTING001/orders" -Method POST -Headers $headers -Body $body
```

### 35. /api/v1/orders/{orderId} [GET]
- **簡介**: 查詢訂單詳情
- **請求參數**
```json
{
  "orderId": "01KORDER001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "order_id": "01KORDER001",
    "listing_id": "01KLISTING001",
    "listing_title": "九成新洗衣機",
    "order_status": "confirmed",
    "role_in_order": "buyer",
    "buyer_note": "Can pick up tonight.",
    "handover_method": "lobby_pickup",
    "cancel_reason": "",
    "created_at": "2026-04-17T05:40:00Z",
    "updated_at": "2026-04-17T05:45:00Z",
    "confirmed_at": "2026-04-17T05:45:00Z",
    "cover_image": {
      "media_asset_id": "01KMEDIA001",
      "url": "http://localhost:9000/local-bucket/ajo_living/01kupload.webp",
      "sort_order": 1,
      "is_cover": true
    },
    "peer": {
      "public_id": "01KUSERSELLER001",
      "display_name": "Owner Seller",
      "role_in_order": "seller"
    },
    "logs": [
      {
        "action_type": "create_order",
        "from_status": "",
        "to_status": "pending_confirm",
        "operator_user_id": "10001",
        "note": "Can pick up tonight.",
        "created_at": "2026-04-17T05:40:00Z"
      },
      {
        "action_type": "confirm_order",
        "from_status": "pending_confirm",
        "to_status": "confirmed",
        "operator_user_id": "10002",
        "note": "",
        "created_at": "2026-04-17T05:45:00Z"
      }
    ]
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:45:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/orders/01KORDER001" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/orders/01KORDER001" -Method GET -Headers $headers
```

### 36. /api/v1/orders/{orderId}/confirm [POST]
- **簡介**: 賣家確認訂單
- **請求參數**
```json
{
  "orderId": "01KORDER001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "order_id": "01KORDER001",
    "listing_id": "01KLISTING001",
    "listing_title": "九成新洗衣機",
    "order_status": "confirmed",
    "role_in_order": "seller",
    "buyer_note": "Can pick up tonight.",
    "handover_method": "lobby_pickup",
    "cancel_reason": "",
    "created_at": "2026-04-17T05:40:00Z",
    "updated_at": "2026-04-17T05:45:00Z",
    "confirmed_at": "2026-04-17T05:45:00Z",
    "cover_image": {
      "media_asset_id": "01KMEDIA001",
      "url": "http://localhost:9000/local-bucket/ajo_living/01kupload.webp",
      "sort_order": 1,
      "is_cover": true
    },
    "peer": {
      "public_id": "01KUSERBUYER001",
      "display_name": "Neighbour Buyer",
      "role_in_order": "buyer"
    },
    "logs": [
      {
        "action_type": "create_order",
        "from_status": "",
        "to_status": "pending_confirm",
        "operator_user_id": "10001",
        "note": "Can pick up tonight.",
        "created_at": "2026-04-17T05:40:00Z"
      },
      {
        "action_type": "confirm_order",
        "from_status": "pending_confirm",
        "to_status": "confirmed",
        "operator_user_id": "10002",
        "note": "",
        "created_at": "2026-04-17T05:45:00Z"
      }
    ]
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:45:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/orders/01KORDER001/confirm" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/orders/01KORDER001/confirm" -Method POST -Headers $headers
```

### 37. /api/v1/orders/{orderId}/cancel [POST]
- **簡介**: 取消訂單
- **請求參數**
```json
{
  "orderId": "01KORDER001", // 路徑參數
  "reason": "Need to reschedule." // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "order_id": "01KORDER001",
    "listing_id": "01KLISTING001",
    "listing_title": "九成新洗衣機",
    "order_status": "cancelled",
    "role_in_order": "buyer",
    "buyer_note": "Can pick up tonight.",
    "handover_method": "lobby_pickup",
    "cancel_reason": "Need to reschedule.",
    "created_at": "2026-04-17T05:40:00Z",
    "updated_at": "2026-04-17T05:50:00Z",
    "cancelled_at": "2026-04-17T05:50:00Z",
    "cover_image": {
      "media_asset_id": "01KMEDIA001",
      "url": "http://localhost:9000/local-bucket/ajo_living/01kupload.webp",
      "sort_order": 1,
      "is_cover": true
    },
    "peer": {
      "public_id": "01KUSERSELLER001",
      "display_name": "Owner Seller",
      "role_in_order": "seller"
    },
    "logs": [
      {
        "action_type": "create_order",
        "from_status": "",
        "to_status": "pending_confirm",
        "operator_user_id": "10001",
        "note": "Can pick up tonight.",
        "created_at": "2026-04-17T05:40:00Z"
      },
      {
        "action_type": "cancel_order",
        "from_status": "pending_confirm",
        "to_status": "cancelled",
        "operator_user_id": "10001",
        "note": "Need to reschedule.",
        "created_at": "2026-04-17T05:50:00Z"
      }
    ]
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:50:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/orders/01KORDER001/cancel" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Need to reschedule."
  }'
```
- **Powershell測試**
```powershell
$headers=@{
  "Authorization"="Bearer $token"
  "Content-Type"="application/json"
}
$body=@{
  reason="Need to reschedule."
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/orders/01KORDER001/cancel" -Method POST -Headers $headers -Body $body
```

### 38. /api/v1/orders/{orderId}/complete [POST]
- **簡介**: 完成訂單
- **請求參數**
```json
{
  "orderId": "01KORDER001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "order_id": "01KORDER001",
    "listing_id": "01KLISTING001",
    "listing_title": "九成新洗衣機",
    "order_status": "completed",
    "role_in_order": "buyer",
    "buyer_note": "Can pick up tonight.",
    "handover_method": "lobby_pickup",
    "cancel_reason": "",
    "created_at": "2026-04-17T05:40:00Z",
    "updated_at": "2026-04-17T06:00:00Z",
    "confirmed_at": "2026-04-17T05:45:00Z",
    "completed_at": "2026-04-17T06:00:00Z",
    "cover_image": {
      "media_asset_id": "01KMEDIA001",
      "url": "http://localhost:9000/local-bucket/ajo_living/01kupload.webp",
      "sort_order": 1,
      "is_cover": true
    },
    "peer": {
      "public_id": "01KUSERSELLER001",
      "display_name": "Owner Seller",
      "role_in_order": "seller"
    },
    "logs": [
      {
        "action_type": "create_order",
        "from_status": "",
        "to_status": "pending_confirm",
        "operator_user_id": "10001",
        "note": "Can pick up tonight.",
        "created_at": "2026-04-17T05:40:00Z"
      },
      {
        "action_type": "confirm_order",
        "from_status": "pending_confirm",
        "to_status": "confirmed",
        "operator_user_id": "10002",
        "note": "",
        "created_at": "2026-04-17T05:45:00Z"
      },
      {
        "action_type": "complete_order",
        "from_status": "confirmed",
        "to_status": "completed",
        "operator_user_id": "10001",
        "note": "",
        "created_at": "2026-04-17T06:00:00Z"
      }
    ]
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T06:00:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/orders/01KORDER001/complete" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/orders/01KORDER001/complete" -Method POST -Headers $headers
```

---

## Notification 模組

### 39. /api/v1/notifications [GET]
- **簡介**: 查詢我的通知列表
- **請求參數**
```json
{
  "page": 1, // 可選
  "page_size": 20, // 可選
  "only_unread": false // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "items": [
      {
        "notification_id": "01KNOTIFY001",
        "category": "order_confirmed",
        "title": "Order confirmed",
        "body": "The seller has confirmed your order request.",
        "related_type": "order",
        "related_public_id": "01KORDER001",
        "is_read": false,
        "created_at": "2026-04-17T05:45:00Z"
      },
      {
        "notification_id": "01KNOTIFY002",
        "category": "chat_message",
        "title": "New chat message",
        "body": "See you tonight at the lobby.",
        "related_type": "chat",
        "related_public_id": "01KCHAT001",
        "is_read": true,
        "read_at": "2026-04-17T05:52:00Z",
        "created_at": "2026-04-17T05:50:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 2
    },
    "unread_count": 1
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T06:05:00Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/notifications?page=1&page_size=20&only_unread=false" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/notifications?page=1&page_size=20&only_unread=false" -Method GET -Headers $headers
```

### 40. /api/v1/notifications/unread-count [GET]
- **簡介**: 查詢未讀通知數
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "unread_count": 1
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T06:05:30Z"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/notifications/unread-count" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/notifications/unread-count" -Method GET -Headers $headers
```

### 41. /api/v1/notifications/{notificationId}/read [POST]
- **簡介**: 標記單一通知已讀
- **請求參數**
```json
{
  "notificationId": "01KNOTIFY001" // 路徑參數
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "notification_id": "01KNOTIFY001",
    "read": true
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T06:06:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/notifications/01KNOTIFY001/read" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/notifications/01KNOTIFY001/read" -Method POST -Headers $headers
```

### 42. /api/v1/notifications/read-all [POST]
- **簡介**: 標記全部通知已讀
- **請求參數**
```json
{}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "updated": 3
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T06:07:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/notifications/read-all" \
  -H "Authorization: Bearer $token"
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/notifications/read-all" -Method POST -Headers $headers
```

---
