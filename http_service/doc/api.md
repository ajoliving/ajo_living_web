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

### 基礎資料模組

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
| 49 | /api/v1/auth/email/register | POST | 建立電郵與手機密碼帳戶 | 無 |
| 50 | /api/v1/auth/email/login | POST | 使用郵箱密碼登入 | 無 |
| 50.2 | /api/v1/auth/password/email/request | POST | 申請電郵重設密碼驗證碼 | 無 |
| 50.3 | /api/v1/auth/password/email/reset | POST | 使用電郵驗證碼重設密碼 | 無 |
| 51 | /api/v1/auth/phone/login | POST | 使用手機密碼登入 | 無 |
| 73 | /api/v1/auth/ismart/login | POST | 使用 ismart 帳戶登入並同步 POS 權限 | 無 |
| 73.1 | /api/v1/me/ismart/building-info | GET | 查詢目前會員可見大廈資料與文件中繼資料 | 會員 |
| 6 | /api/v1/auth/logout | POST | 登出目前會員 | 會員 |
| 7 | /api/v1/me | GET | 取得目前會員資料 | 會員 |
| 8 | /api/v1/me/profile | PATCH | 更新會員資料 | 會員 |
| 9 | /api/v1/me/secondhand/listings | GET | 取得我的二手帖子 | 會員 |
| 66 | /api/v1/me/secondhand/favorites | GET | 取得我的二手收藏 | 會員 |
| 52 | /api/v1/me/wallet | GET | 取得 AJO Point 錢包總覽 | 會員 |
| 53 | /api/v1/me/wallet/transactions | GET | 查詢我的積分流水 | 會員 |
| 74 | /api/v1/public/ads | GET | 查詢公開列表右側展示廣告 | 無 |
| 54 | /api/v1/me/wallet/ad-tasks | GET | 查詢可領取的廣告積分任務 | 會員 |
| 55 | /api/v1/me/wallet/ad-tasks/{taskId}/start | POST | 開始廣告觀看任務 | 會員 |
| 56 | /api/v1/me/wallet/ad-tasks/{taskId}/claim | POST | 領取廣告積分 | 會員 |
| 66 | /api/v1/me/wallet/ad-tasks/{taskId}/click | POST | 記錄廣告連結點擊 | 會員 |
| 80 | /api/v1/me/payments/pos/overview | GET | 取得 POS 物業繳費概覽 | 會員 |
| 81 | /api/v1/me/payments/pos/bills | GET | 查詢目前單位 POS 賬單 | 會員 |
| 82 | /api/v1/me/payments/pos/fees | GET | 查詢 POS 手續費與付款方式 | 會員 |
| 83 | /api/v1/me/payments/pos/bank-accounts | GET | 查詢 POS 銀行賬戶 | 會員 |
| 84 | /api/v1/me/payments/pos/payments/report | POST | 上報線下 POS 繳費 | 會員 |
| 85 | /api/v1/me/payments/pos/terminal/pay | POST | 目前所屬單位 POS 機收款並入賬 | 會員 |
| 86 | /api/v1/me/payments/pos/orders | GET | 查詢目前單位線上繳費訂單 | 會員 |
| 87 | /api/v1/me/payments/pos/orders | POST | 建立 POS H5 / QR 線上繳費訂單 | 會員 |
| 88 | /api/v1/me/payments/pos/orders/query | POST | 按商戶單號或支付單號查詢訂單 | 會員 |
| 89 | /api/v1/me/payments/pos/orders/{mchOrderNo} | GET | 查詢單一 POS H5 / QR 線上繳費訂單 | 會員 |
| 90 | /api/v1/me/payments/pos/orders/{mchOrderNo}/close | POST | 關閉 POS H5 / QR 訂單 | 會員 |
| 91 | /api/v1/me/payments/pos/orders/{mchOrderNo}/cancel | POST | 取消 POS H5 / QR 訂單 | 會員 |
| 92 | /api/v1/me/payments/pos/orders/{mchOrderNo}/simulate | POST | 會員在非 production 模擬可見 H5 訂單 | 會員 |
| 93 | /api/v1/me/payments/pos/history | GET | 查詢 POS 交易歷史 | 會員 |
| 94 | /api/v1/me/payments/pos/history/{paymentId} | GET | 查詢 POS 交易詳情 | 會員 |
| 95 | /api/v1/me/payments/pos/accounting | GET | 查詢所屬屋苑 POS 會計資料 | 會員 |
| 96 | /api/v1/me/payments/pos/accounting/clear | POST | 所屬屋苑多選清機 | 會員 |
| 97 | /api/v1/me/payments/pos/accounting/records | GET | 查詢所屬屋苑清機歷史 | 會員 |
| 98 | /api/v1/me/payments/pos/accounting/records/{recordId} | GET | 查詢所屬屋苑清機詳情 | 會員 |
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
| 47 | /api/v1/secondhand/settings/listings/{listingId}/mark-sold | POST | 設定頁標記任意帖子已售 | Staff |
| 48 | /api/v1/secondhand/settings/listings/{listingId}/deactivate | POST | 設定頁下架任意帖子 | Staff |

### Property Sale 模組

| 編號 | 介面 | 方法 | 簡介/功能 | 權限 |
| --- | --- | --- | --- | --- |
| 75 | /api/v1/property-sales | GET | 查詢公開樓盤列表，預設按廣告權重排序 | 無 / 會員 |
| 76 | /api/v1/property-sales/{listingId} | GET | 查詢樓盤詳情，公開回應不返回真實樓層 | 無 / 會員 |
| 77 | /api/v1/property-sales | POST | 建立樓盤草稿 | 會員 |
| 78 | /api/v1/property-sales/{listingId} | PATCH | 更新樓盤草稿或已發布樓盤 | 會員 |
| 79 | /api/v1/property-addresses/search | GET | 查詢屋苑或大廈地址聯想 | 無 |
| 79A | /api/v1/property-sales/translation | POST | 將樓盤標題及單位介紹翻譯成 English | 會員 |
| 80 | /api/v1/market-trends/rent | GET | 香港住宅租金走勢 | 無 |

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
| 99 | /api/v1/staff/system-notices | POST | 發布全站系統通知 | Staff |
| 57 | /api/v1/staff/wallet/transactions | GET | 查詢平台積分流水 | Staff |
| 58 | /api/v1/staff/wallet/grants | POST | 手動發放 AJO Point | Staff |
| 59 | /api/v1/staff/wallet/reward-ads | GET | 查詢廣告任務 | Staff |
| 60 | /api/v1/staff/wallet/reward-ads | POST | 建立廣告任務 | Staff |
| 61 | /api/v1/staff/wallet/reward-ads/{taskId} | PATCH | 更新廣告任務 | Staff |
| 62 | /api/v1/staff/secondhand/listings | GET | 管理頁查詢二手帖子列表 | Staff |
| 63 | /api/v1/staff/secondhand/listings/{listingId}/publish | POST | 管理頁上架二手帖子 | Staff |
| 64 | /api/v1/staff/secondhand/listings/{listingId}/deactivate | POST | 管理頁下架二手帖子 | Staff |
| 65 | /api/v1/staff/secondhand/listings/{listingId}/renew | POST | 管理頁續期二手帖子 | Staff |
| 66 | /api/v1/staff/property-sales | GET | 管理頁查詢樓盤租售列表 | Staff |
| 67 | /api/v1/staff/property-sales/{listingId}/publish | POST | 管理頁上架樓盤租售 | Staff |
| 68 | /api/v1/staff/property-sales/{listingId}/deactivate | POST | 管理頁下架樓盤租售 | Staff |
| 69 | /api/v1/staff/property-sales/{listingId}/renew | POST | 管理頁續期樓盤租售 | Staff |
| 70 | /api/v1/staff/serviced-apartments | GET | 管理頁查詢服務式住宅列表 | Staff |
| 71 | /api/v1/staff/serviced-apartments/{listingId}/renew | POST | 管理頁續期服務式住宅 | Staff |

---

## 代理帳戶與資料審核

- 註冊 `POST /api/v1/auth/email/register` 接受 `account_type`: `personal`、`individual_agent`、`agency_company`。代理帳戶註冊後狀態為 `pending_profile`；`pending_profile`、`pending_review`、`rejected` 均可使用電郵、手提電話或用戶名稱配合密碼取得受限 session，只可使用 `/me`、代理資料、代理專用 OSS 上傳及登出接口，以查看審核狀態、拒絕原因及重新提交。
- 會員資料接口：`GET|POST|PATCH /api/v1/me/agency-profile`、`POST /api/v1/me/agency-profile/submit`。回應同時返回 `active_profile` 與 `revision`，修訂審核期間繼續使用已批准版本。
- 個人代理必填中英文名、電話1及 WhatsApp 狀態、預設頭像；非海外代理另須牌照號碼及 EAA 圖片。公司必填中英文名及地址、電話1、牌照號碼、EAA 圖片及商業登記證。
- 代理 OSS `purpose`：`agency_individual_avatar`、`agency_individual_eaa`、`agency_individual_company_card`、`agency_individual_wechat_qr`、`agency_company_logo`、`agency_company_eaa`、`agency_company_business_registration`、`agency_company_company_card`。
- 管理審核接口：`GET /api/v1/staff/agency-profiles`、`GET /api/v1/staff/agency-profiles/{profileId}`、`POST /api/v1/staff/agency-profiles/{profileId}/review`。拒絕時 `review_note` 必填；列表支援 `status`、`profile_type`、`keyword`，`status=all` 表示全部。審核結果會寫入站內通知，註冊帳戶有電郵時另透過 `MailSender` 發送通過或拒絕郵件，拒絕郵件包含 `review_note`；郵件發送失敗不回滾已提交的審核交易。
- 已批准公司子帳戶接口：`GET|POST /api/v1/me/agency-profile/subaccounts`、`PATCH /api/v1/me/agency-profile/subaccounts/{subaccountId}/status`、`DELETE /api/v1/me/agency-profile/subaccounts/{subaccountId}`。權限只接受 `property_publish`、`property_manage`；建立、發布及重新發布使用 `property_publish`，更新、下架及標記售出使用 `property_manage`。公司主帳戶可統一管理所屬子帳戶樓盤。
- EAA 牌照及商業登記證物件強制使用 private ACL；會員本人及管理審核回應只返回有效 10 分鐘的 OSS 簽名下載地址，不使用公開 CDN 地址。Logo、頭像、公司卡片及微信 QR 等展示資產維持公開媒體 URL。
- 樓盤發布身份由 `account_type` 及已批准代理資料派生，忽略請求的 `publisher_identity_type`。海外代理缺少香港牌照及 EAA 圖片時不可建立、修改、發布或重新發布本地樓盤。
- 代理資料修訂在 `pending` 或 `rejected` 時，現有樓盤繼續使用舊已批准資料；修訂批准後，審核交易會刷新該代理及代理公司所有子帳戶的未刪除二手樓盤聯絡快照，包括草稿、已發布、已下架及已過期狀態。

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

## 基礎資料模組

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
  "phone_country_code": "+86", // 必填
  "phone_number": "13812345678", // 必填
  "scene": "login" // 可選
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "expires_in": 300
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
    "phone_country_code": "+86",
    "phone_number": "13812345678",
    "scene": "login"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  phone_country_code="+86"
  phone_number="13812345678"
  scene="login"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/otp/request" -Method POST -Headers $headers -Body $body
```

- 生產環境不會返回驗證碼；`mock_code` 僅可能出現在本機 mock 環境。
- 同一手機號碼在 `OTP_RESEND_COOLDOWN` 期間重複申請會返回 `RATE_LIMITED`。
- 已接入的阿里雲國內短信驗證碼僅接受 `+86` 及 11 位中國大陸手提電話號碼。

---

### 5. /api/v1/auth/otp/verify [POST]
- **簡介**: 驗證 OTP 並登入
- **請求參數**
```json
{
  "phone_country_code": "+86", // 必填
  "phone_number": "13812345678", // 必填
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
    "phone_country_code": "+86",
    "phone_number": "13812345678",
    "scene": "login",
    "code": "123456"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  phone_country_code="+86"
  phone_number="13812345678"
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
- **簡介**: 建立電郵與手機密碼帳戶
- **請求參數**
```json
{
  "email": "member@example.com", // 必填
  "password": "safe-password-123", // 必填，至少 8 個字元
  "eng_name": "CHAN TAI MAN", // 必填，英文姓名；舊 display_name 仍兼容
	"chi_name": "陳大文", // 選填，中文姓名，會同步至 iSmart
  "username": "email-member", // 選填，未提供時後端使用手機號碼生成本地登入名
  "phone_country_code": "+852", // 必填
  "phone_number": "91234567", // 必填
	"id_card": "A1234567", // 選填，身份證或證件號碼
	"remark": "registered from AJO", // 選填，iSmart 客戶備註
	"gender": "M", // 選填，只接受 M 或 F
	"is_receive_email": true, // 選填，預設 true
	"account_type": "personal", // personal、individual_agent 或 agency_company
  "publisher_identity_type": "owner", // 舊客戶端兼容欄位；後端按 account_type 派生並忽略此值
  "primary_community_id": "01KCOMMUNITY001", // 選填
  "residence_floor": "12", // 選填
  "residence_unit": "08" // 選填
}
```
- **iSmart 同步**: 註冊會先呼叫 iSmart `/api/v1/integration/auth/register/`，並以相同密碼建立 iSmart 帳戶；`individual_agent` 與 `agency_company` 固定傳送 `legal_entity=LE`，其餘帳戶傳送 `NA`。iSmart 建立失敗時不會建立本地 AJO 帳戶。
- **物業綁定申請**: 個人帳戶同時提交 `primary_community_id`、`residence_floor`、`residence_unit` 時，AJO 會在本地帳戶及 iSmart 關聯建立後，按 POS 單位清單解析 `unit_id` 並提交 OwnerReg 審批申請。iSmart HTTP `2xx` 即視為申請已受理，不等待審批；回應 `/me.residence_binding_status` 為 `pending`，且 `bound_building_ids`、`bound_flat_unit_ids` 保持空。申請失敗時回應 `account created but property binding request failed`，帳戶及 iSmart 關聯會保留，會員可登入後重新申請。
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
      "profile_completed": true
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
    "eng_name": "CHAN TAI MAN",
	"chi_name": "陳大文",
    "phone_country_code": "+852",
    "phone_number": "91234567",
	"account_type": "personal",
    "publisher_identity_type": "owner",
    "primary_community_id": "01KCOMMUNITY001",
    "residence_floor": "12",
    "residence_unit": "08"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  email="member@example.com"
  password="safe-password-123"
  eng_name="CHAN TAI MAN"
	chi_name="陳大文"
  phone_country_code="+852"
  phone_number="91234567"
	account_type="personal"
  publisher_identity_type="owner"
  primary_community_id="01KCOMMUNITY001"
  residence_floor="12"
  residence_unit="08"
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

### 50.1 /api/v1/auth/username/login [POST]
- **簡介**: 使用本地住戶用戶名密碼登入。用戶名由 `/api/v1/auth/email/register` 建立，後端會以小寫形式保存與查找。
- **請求參數**
```json
{
  "username": "email-member", // 必填
  "password": "safe-password-123" // 必填
}
```
- **回應參數**: 同 `/api/v1/auth/email/login`
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/username/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "email-member",
    "password": "safe-password-123"
  }'
```

---

### 50.2 /api/v1/auth/password/email/request [POST]
- **簡介**: 針對已綁定電郵的帳戶發送重設密碼驗證碼。
- **請求參數**
```json
{
  "email": "member@example.com" // 必填
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
  "timestamp": "2026-04-17T05:16:20Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/password/email/request" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "member@example.com"
  }'
```

---

### 50.3 /api/v1/auth/password/email/reset [POST]
- **簡介**: 使用電郵驗證碼重設本系統登入密碼。
- **請求參數**
```json
{
  "email": "member@example.com", // 必填
  "code": "123456", // 必填
  "password": "new-password-123" // 必填，至少 8 個字元
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "password_reset": true
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:16:40Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/password/email/reset" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "member@example.com",
    "code": "123456",
    "password": "new-password-123"
  }'
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

### 73. /api/v1/auth/ismart/login [POST]
- **簡介**: 使用 ismart 帳戶登入並同步 POS 權限。後端預設呼叫 `https://pos.ismart.skylinedances.com/api/poslogin`，也可透過 `POS_LOGIN_URL` 覆蓋；成功後以 `ismart_msg.user_id` 綁定本地會員，若本地不存在會自動建立會員並簽發本系統 token。
- **原始 iSmart 資料**: 登入回應 `data.user.ismart_raw` 及 `GET /api/v1/me` 的 `data.ismart_raw` 返回已保存的上游註冊 `data` 與 POS 登入 `msg` 業務欄位；未知欄位會保留，後續 POS 回應只補充可用欄位，不會以空值覆蓋註冊資料。`password`、`token`、`secret`、`authorization`、憑證與 session 類欄位會遞迴移除。`ismart_msg` 維持固定摘要欄位。
- **帳號判斷**: `account` 若符合香港手機格式，會先以 8 位本地手機號登入，失敗後再用原始輸入登入；若 `account` 是用戶名且有傳入 `phone`，會在用戶名失敗後繼續嘗試手機候選值。全部失敗後才回傳帳號或密碼錯誤。
- **請求參數**
```json
{
  "account": "testowner02", // 必填，ismart 用戶名或手機
  "password": "test02test02", // 必填
  "phone": "+85291234567", // 可選，當 account 是用戶名時可用於本地手機綁定；若 ismart user_id 已存在，會以本次值更新本地手機
  "email": "linked@example.com" // 可選，綁定到本地會員郵箱
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
      "profile_completed": false,
      "ismart_msg": {
        "user_id": 410,
        "username": "testowner02",
        "phone": "+85291234567",
        "is_staff": false,
        "building": [],
        "staff_building_permissions": [],
        "client_building_permissions": ["0999900"],
        "client_building_flat_units_permissions": [
          "09999000012",
          "09999000111",
          "09999000211"
        ]
      }
    }
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-04-17T05:17:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/auth/ismart/login" \
  -H "Content-Type: application/json" \
  -d '{
    "account": "testowner02",
    "password": "test02test02",
    "phone": "+85291234567",
    "email": "linked@example.com"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Content-Type"="application/json"}
$body=@{
  account="testowner02"
  password="test02test02"
  phone="+85291234567"
  email="linked@example.com"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/ismart/login" -Method POST -Headers $headers -Body $body
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
    "residence_floor": "12",
    "residence_unit": "08",
    "bound_building_ids": ["0999900"],
    "bound_flat_unit_ids": ["09999000012", "09999000111"],
    "primary_community": {
      "public_id": "01KCOMMUNITY001",
      "community_type": "estate",
      "name_zh": "康怡花園",
      "name_en": "Kornhill",
      "district_code": "hk_east",
      "address_text": "Quarry Bay"
    },
    "profile_completed": true,
    "ismart_linked": true,
    "ismart_username": "testowner02",
    "ismart_bound_phone": "+85291234567",
    "ismart_password": "test02test02",
    "local_password": "new-password-123"
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
  "email": "linked@example.com", // 可選，變更時需配合 email_otp_code
  "email_otp_code": "123456", // 可選
  "phone_country_code": "+852", // 可選
  "phone_number": "91234567", // 可選
  "password": "new-password-123", // 可選，更新本系統登入密碼
  "publisher_identity_type": "owner", // 可選，owner / agent
  "primary_community_id": "01KCOMMUNITY001", // 可選
  "bound_building_ids": ["0999900"], // 可選，多個綁定屋苑或大廈
  "bound_flat_unit_ids": ["09999000012", "09999000111"], // 可選，多個綁定單位
  "residence_floor": "12", // 可選
  "residence_unit": "08", // 可選
  "district_code": "hk_east" // 可選
}
```
- **身份規則**: `publisher_identity_type` 為樓盤發布身份，只接受 `owner` 或 `agent`；新帳戶及舊空值預設為 `owner`，會員可在帳號管理透過 `PATCH /me/profile` 修改。
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
    "publisher_identity_type": "owner",
    "district_code": "hk_east",
    "bound_building_ids": ["0999900"],
    "bound_flat_unit_ids": ["09999000012", "09999000111"],
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
    "primary_community_id": "01KCOMMUNITY001",
    "bound_building_ids": ["0999900"],
    "bound_flat_unit_ids": ["09999000012"],
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
  bound_building_ids=@("0999900")
  bound_flat_unit_ids=@("09999000012")
  district_code="hk_east"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/me/profile" -Method PATCH -Headers $headers -Body $body
```

---

### 8A. /api/v1/me/ismart/bind [POST]
- **簡介**: 已登入會員關聯 ismart 帳戶，後端使用 POS login 校驗帳密並同步屋苑、單位與 Staff 權限。
- **請求參數**
```json
{
  "account": "testowner02",
  "password": "test02test02",
  "phone": "+85291234567",
  "email": "linked@example.com"
}
```
- **回應參數**: 同 `/api/v1/me`，包含 `ismart_linked`、`ismart_msg`、`bound_building_ids` 與 `bound_flat_unit_ids`。

---

### 8B. /api/v1/me/ismart/building-info [GET]
- **簡介**: 查詢目前會員可見大廈資料與文件中繼資料，供 `我的大廈` 內 `大廈資料` 與 `大廈財務` 使用。
- **請求參數**
```json
{
  "building_id": "0348200"
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "selected_building_id": "0348200",
    "building_options": ["0348200"],
    "building": {
      "building_id": "0348200",
      "buildname_chi": "示例大廈",
      "buildname": "Example Building"
    },
    "building_info": {
      "owners_corporation_name": "示例法團",
      "management_office_phone": "21234567",
      "management_company_name": "Example PM Co."
    },
    "documents": {
      "forms": [],
      "building_info_files": [],
      "floorplans": [],
      "audit_reports": [],
      "financial_reports": []
    }
  }
}
```
- **備註**: 後端會兼容 `floorplan` / `floorplans`、`auditreport` / `audition` / `audit_report` / `auditreports` / `auditions` / `audit_reports`、`mfinreport` / `financial_reports`，並統一補齊 canonical 陣列。共享的大廈資料預設以 Redis 快取 5 分鐘；會員所屬大廈權限會在讀取快取前即時校驗，`building_options` 等會員資料不會寫入共享快取。

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
  "content": "你好，請問仍可交易嗎？" // 必填，最多 1000 字
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

### 80-98. /api/v1/me/payments/pos/* [GET/POST]
- **簡介**: AJO 會員支付中心的 POS 物業繳費接口。會員需先關聯 ismart，所有 POS 資料讀寫均由 AJO 後端代理，並統一限制於 `client_building_permissions` 與 `client_building_flat_units_permissions` 所屬範圍，不要求 Staff 身份。
- **查詢參數**: `overview`、`bills`、`fees`、`bank-accounts`、`orders`、`history`、`accounting` 支援 `building_id` 與 `unit_id`。`overview` 支援 `summary=false` 跳過賬單與訂單概覽統計。`history` 額外支援 `from_date`、`to_date`、`date_type`、`pay_method`、`unit_ids`。會計頁只需 `building_id`。
- **接口**
| 介面 | 方法 | 說明 |
| --- | --- | --- |
| `/api/v1/me/payments/pos/overview` | GET | 返回選中上下文、可選屋苑與單位、是否需補個人資料、是否需 ismart 登入、是否 Staff。 |
| `/api/v1/me/payments/pos/bills` | GET | 返回選中單位的 POS 賬單。 |
| `/api/v1/me/payments/pos/fees` | GET | 返回選中大廈的 POS 手續費與可用付款方式。 |
| `/api/v1/me/payments/pos/bank-accounts` | GET | 返回選中大廈的 POS 銀行賬戶。 |
| `/api/v1/me/payments/pos/payments/report` | POST | 上報現金、支票、銀行轉賬等線下繳費。 |
| `/api/v1/me/payments/pos/terminal/pay` | POST | 會員於所屬單位使用服務端配置的 POS 機地址收款，成功後代理 `/bill` 入賬；不接受前端傳入設備地址。 |
| `/api/v1/me/payments/pos/orders` | GET | 返回選中單位的 H5 / QR 線上繳費訂單。 |
| `/api/v1/me/payments/pos/orders` | POST | 建立選中單位賬單 H5 / QR 線上繳費訂單。 |
| `/api/v1/me/payments/pos/orders/query` | POST | 按 `mch_order_no` 或 `pay_order_id` 查詢訂單。 |
| `/api/v1/me/payments/pos/orders/{mchOrderNo}` | GET | 查詢單一 H5 / QR 線上繳費訂單。 |
| `/api/v1/me/payments/pos/orders/{mchOrderNo}/close` | POST | 關閉單一 H5 / QR 線上繳費訂單。 |
| `/api/v1/me/payments/pos/orders/{mchOrderNo}/cancel` | POST | 取消單一 H5 / QR 線上繳費訂單。 |
| `/api/v1/me/payments/pos/orders/{mchOrderNo}/simulate` | POST | 會員在非 production 模擬所屬單位的 H5 訂單狀態；正式環境關閉。 |
| `/api/v1/me/payments/pos/history` | GET | 返回所屬單位的 POS 交易歷史，可按所屬單位批量查詢。 |
| `/api/v1/me/payments/pos/history/{paymentId}` | GET | 返回單一 POS 交易詳情。 |
| `/api/v1/me/payments/pos/accounting` | GET | 返回會員所屬屋苑的現金、支票待清機分組與清機歷史摘要。 |
| `/api/v1/me/payments/pos/accounting/clear` | POST | 會員於所屬屋苑多選現金或支票交易清機。 |
| `/api/v1/me/payments/pos/accounting/records` | GET | 返回會員所屬屋苑清機歷史。 |
| `/api/v1/me/payments/pos/accounting/records/{recordId}` | GET | 返回會員所屬屋苑清機詳情。 |
- **建立訂單請求**
```json
{
  "scene": "cart",
  "building_id": "0999900",
  "unit_id": "09999000012",
  "pay_channel": "WX_H5",
  "final_amount": 12345,
  "handle_fee_amount": 0,
  "bill_objs": [],
  "handle_fee_obj": [],
  "return_path": "https://ajoliving.skylinedances.com/payments/orders",
  "remark": "物業費",
  "gateway_request_overrides": {
    "walletType": "HK"
  }
}
```
- **H5 建單限制**: `gateway_request_overrides` 目前只接受 `ALI_H5` 的 `walletType=CN/HK`，`clientIp` 由 AJO 後端自行寫入。
- **POS 機收款請求**
```json
{
  "building_id": "0999900",
  "unit_id": "09999000012",
  "pay_type": "POS_CARD",
  "final_amount": 12345,
  "bill_objs": [],
  "handle_fee_obj": [],
  "remark": "物業費"
}
```
- **POS 機收款配置**: 預設關閉；啟用時需配置 `POS_TERMINAL_PROXY_ENABLED=true`、`POS_TERMINAL_TYPE=allinpay`、`POS_TERMINAL_URL`，可用 `POS_TERMINAL_ALLOWED_BUILDINGS` 限定大廈。
- **H5 模擬請求**
```json
{
  "building_id": "0999900",
  "unit_id": "09999000012",
  "state": "SUCCESS",
  "gateway_message": "manual test"
}
```
- **H5 模擬限制**: 只允許操作會員所屬單位，且 `APP_ENV=production` 時固定拒絕。

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
      { "biz_module": "property_sale", "label": "樓盤租售", "publish": 1000, "draft_save": 1000, "edit": 1000, "republish": 1000 },
      { "biz_module": "serviced_apartment", "label": "服務式住宅", "publish": 800, "edit": 800, "republish": 800 }
    ],
    "recharge_enabled": true,
    "recharge_rate": 100
  }
}
```

---

### 52.1 /api/v1/me/wallet/recharge-orders [POST]
- **簡介**: 建立 AJO Point 線上充值支付訂單。支付成功後按 `1 HKD = 100 AJO Point` 自動入賬。
- **請求參數**
```json
{
  "amount_hkd": 0.01,
  "pay_method": "wechat",
  "pay_region": "HK",
  "device_mode": "desktop",
  "return_path": "https://ajoliving.skylinedances.com/account/profile/wallet"
}
```
- **欄位說明**
| 欄位 | 說明 |
| --- | --- |
| `pay_method` | `wechat`、`alipay`、`unionpay` |
| `pay_region` | `HK`、`CN`；目前主要用於支付寶錢包區分 |
| `device_mode` | `mobile` 生成 H5 支付，`desktop` 生成掃碼支付 |
| `amount_hkd` | 最小 `0.01`，按分計算；`0.01 HKD = 1 AJO Point` |
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "order_id": "01KRECHARGE001",
    "mch_order_no": "AJO_1774330341639_123456",
    "pay_order_id": "P2036315163422986241",
    "pay_channel": "WX_QR",
    "pay_region": "HK",
    "pay_data_type": "codeUrl",
    "pay_data": "https://example.com/pay",
    "state": "PAYING",
    "state_label": "支付確認中",
    "gateway_state_code": 1,
    "gateway_message": "SUCCESS",
    "currency": "HKD",
    "amount_hkd": "0.01",
    "amount_cents": 1,
    "points_amount": 1,
    "created_at": "2026-04-17T06:30:00Z",
    "updated_at": "2026-04-17T06:30:00Z",
    "expire_time": "2026-04-17T06:33:00Z",
    "paid_at": "",
    "credited_at": ""
  }
}
```

---

### 52.2 /api/v1/me/wallet/recharge-orders/{orderId} [GET]
- **簡介**: 查詢目前會員充值訂單；`orderId` 可使用 `order_id` 或支付回跳中的 `mch_order_no`。
- **請求參數**
```json
{
  "refresh": true
}
```
- **說明**: `refresh=true` 時，若訂單仍在支付中，後端會向 EasyLink 主動查單並在成功後入賬。
- **回應參數**: 同建立充值訂單回應。

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

### 74. /api/v1/public/ads [GET]
- **簡介**: 查詢公開列表右側展示廣告。只返回 active、display 類型、未過期且符合頻道與位置的廣告，最多 10 條。
- **請求參數**
```json
{
  "channel": "property_sale",
  "placement": "listing_side",
  "limit": 10
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
        "title": "AJO Living Promotion",
        "summary": "Featured listing-side display ad.",
        "media_url": "https://cdn.example.com/ad.png",
        "media_type": "image",
        "target_url": "https://www.ajoliving.com",
        "display_channel": "property_sale",
        "display_placement": "listing_side",
        "display_layout": "image_text",
        "sort_order": 1
      }
    ]
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
- **簡介**: Staff 查詢廣告任務，支援 reward 看廣告得積分與 display 列表展示廣告。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20,
  "keyword": "AJO",
  "is_active": true,
  "ad_type": "display",
  "display_channel": "property_sale"
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
        "ad_type": "display",
        "title": "AJO Living Reward",
        "summary": "Watch official video.",
        "cover_url": "https://cdn.example.com/ad.png",
        "media_url": "https://cdn.example.com/ad.mp4",
        "media_type": "video",
        "target_url": "https://www.ajoliving.com",
        "display_channel": "property_sale",
        "display_placement": "listing_side",
        "display_layout": "image_text",
        "sort_order": 1,
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
- **簡介**: Staff 建立廣告任務。`ad_type=reward` 用於看廣告得積分，`ad_type=display` 用於列表右側展示廣告。
- **請求參數**
```json
{
  "ad_type": "display",
  "title": "AJO Living Reward",
  "summary": "Watch official video.",
  "cover_url": "https://cdn.example.com/ad.png",
  "media_url": "https://cdn.example.com/ad.mp4",
  "media_type": "video",
  "target_url": "https://www.ajoliving.com",
  "display_channel": "property_sale",
  "display_placement": "listing_side",
  "display_layout": "image_text",
  "sort_order": 1,
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
- **簡介**: Staff 更新廣告任務，可用於修改素材、展示設定、預算或啟停任務。
- **請求參數**
```json
{
  "ad_type": "display",
  "title": "AJO Living Reward",
  "display_channel": "property_sale",
  "display_placement": "listing_side",
  "display_layout": "image_full",
  "sort_order": 1,
  "reward_points": 80,
  "watch_seconds": 30,
  "total_budget": 2000,
  "is_active": false
}
```

---

### 62. /api/v1/staff/secondhand/listings [GET]
- **簡介**: Staff 管理頁分頁查詢二手帖子，支援 keyword、category_code 與 status 篩選。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20,
  "keyword": "sofa",
  "category_code": "home_furniture",
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
- **簡介**: Staff 管理頁分頁查詢樓盤租售，支援 keyword、status、租售類型、物業類型與標籤篩選。
- **請求參數**
```json
{
  "page": 1,
  "page_size": 20,
  "keyword": "Central",
  "status": "active",
  "transaction_type": "rent",
  "property_type": "residential",
  "rental_type": "short_term",
  "feature_tags": "brand_new,view",
  "publisher_identity_type": "agent",
  "is_new": true
}
```
- **回應參數**: 與 `/api/v1/property-sales` 列表項結構一致，管理頁可見全部狀態。

---

### 75-79A. /api/v1/property-sales 與 /api/v1/property-addresses/search
- **簡介**: 樓盤發布支援物業編號、雙語標題與介紹、地址聯想、真實樓層保密、廣告套餐與權重排序。公開列表預設以 `ad_weight desc` 排序，再按刷新時間排序。
- **建立 / 更新樓盤主要請求欄位**
```json
{
  "property_no": "AJO-PS-0001",
  "title": "康怡花園高層兩房",
  "title_en": "High floor two-bedroom in Kornhill",
  "description": "實用兩房，交通方便。",
  "description_en": "Practical two-bedroom unit with convenient transport.",
  "transaction_type": "sale",
  "location_scope": "local",
  "listing_category": "standard",
  "multi_unit_project": false,
  "estate_name": "康怡花園",
  "address_text": "鰂魚涌康山道",
  "address_text_en": "Kornhill Road, Quarry Bay",
  "property_attributes": {
    "prn": "PRN-PRIVATE-001",
    "rent_included_items": "rates_government_rent,management_fee"
  },
  "block_name": "3座",
  "floor_raw": "25/F",
  "floor_zone": "high",
  "total_floors": 40,
  "unit_name": "A",
  "show_unit": true,
  "asking_price_hkd": 8000000,
  "monthly_rent_hkd": 0,
  "price_reference_only": true,
  "price_negotiable": false,
  "annual_prepay_discount": true,
  "annual_prepay_option": "provided",
  "usable_area_sqft": 620,
  "ad_package_code": "featured",
  "contact": {
    "phone": "61234567",
    "whatsapp": "61234567",
    "contact_attributes": {
      "phone_country_code": "+852",
      "phone_whatsapp_enabled": "yes",
      "default_avatar_gender": "male"
    }
  }
}
```
- **草稿規則**: `POST` 與草稿狀態的 `PATCH` 可保存未完成資料，讓會員稍後繼續填寫；草稿不會出現在公開列表。會員明確儲存草稿時，前端以 `charge_draft=true` 查詢參數提交，建立或更新成功會在同一交易扣除 1,000 AJO Points；餘額不足時資料與扣款均不生效。正式發布流程內部的暫存請求不帶此參數，避免重複扣款。
- **發布者身份規則**: 會員建立、更新、發布或重新發布樓盤時，後端按 `user_profiles.account_type`、已批准代理資料及公司子帳戶歸屬派生身份並忽略請求中的 `publisher_identity_type`；發布及重新發布會重新寫入當時有效的代理公開快照。
- **樓層規則**: `floor_raw` 保存會員輸入的實際樓層，`floor_zone` 保存對外顯示的 `low`、`middle` 或 `high`；公開回應不洩露 `floor_raw`。
- **發布規則**: 正式發布時必須提交完整欄位，包括 `title_en`、`description_en`、`address_text_en`，其長度分別不超過 100、2000、既有地址欄位限制；`title` 不超過 40，`description` 不超過 1000。`estate_name` 適用於住宅、車位、工商與店鋪，土地 `property_type=land` 可留空，但仍需 `property_no`、`address_text`、放售或放租價格、建築面積、土地分類標籤、有效聯絡資料及至少一張已完成登記的圖片，並按 `ad_package_code` 扣除相應 AJO Points。建立草稿、圖片登記與 `publish` 為連續但獨立的請求；直接儲存並發布只收取發布套餐費用，不另收草稿費用。
- **價格顯示規則**: `price_reference_only=true` 時前端會在顯示價格後加 `起`；`price_negotiable=true` 時前端不顯示實際金額，只顯示 `面議`。放售使用 `asking_price_hkd`，放租使用 `monthly_rent_hkd`，服務式住宅使用最低周租或月租。
- **聯絡資料規則**: 業主可在 `contact.contact_attributes` 以 `phone_whatsapp_enabled=yes` 與 `phone_2_whatsapp_enabled=yes` 分別標記電話1及電話2可使用 WhatsApp，並以 `phone_country_code` 與 `phone_2_country_code` 分別保存兩個電話區號。聯絡方式解鎖成功後，`contact_payload.phone_whatsapp_url` 與 `contact_payload.phone_2_whatsapp_url` 分別返回兩個號碼的 WhatsApp 入口，`whatsapp_url` 保留為舊版單一入口兼容欄位；舊資料缺少 `phone_2_country_code` 時，電話2會兼容使用 `phone_country_code`。`wechat` 欄位在前端顯示為 `Wechat ID`。
- **公開回應規則**: `floor_raw` 只在業主本人查看自己的樓盤詳情時返回；訪客與非業主只會看到 `floor_zone`、`floor_level`、`floor_display_range`、`public_location_text`。
- **廣告套餐**: `basic` 權重 0 / 600 / 30 天；`featured` 權重 1 / 800 / 30 天；`premium` 權重 2 / 1500 / 30 天。
- **會員狀態操作**: `publish` 僅支援草稿；`republish` 支援過期或已下架樓盤重新上架；`deactivate` 會將樓盤設為已下架。
- **內容翻譯**: `POST /api/v1/property-sales/translation` 接受繁體中文 `title` 與 `description`，返回 `title_en` 與 `description_en`。接口需要會員登入，並按會員限制為每 10 分鐘 20 次。後端以 `TRANSLATION_PROVIDER=deepl`、`DEEPL_API_BASE_URL`、`DEEPL_AUTH_KEY`、`TRANSLATION_REQUEST_TIMEOUT` 配置 DeepL；未配置時返回統一服務錯誤，不會把供應商錯誤內容寫入樓盤欄位。
```json
{
  "title": "康怡花園高層兩房",
  "description": "實用兩房，交通方便。"
}
```
```json
{
  "title_en": "High-floor two-bedroom unit in Kornhill",
  "description_en": "Practical two-bedroom unit with convenient transport."
}
```
- **地址聯想**
```json
{
  "keyword": "Kornhill",
  "district_code": "hk_east",
  "limit": 8
}
```

---

### /api/v1/serviced-apartments [POST/PATCH]
- **簡介**: 建立或更新服務式住宅草稿。前端按 xlsx 暫不開放本地/海外、放盤類別、多於一伙或發展商項目、置頂、黃金置頂與即走盤；API 仍保留兼容欄位。
- **主要請求欄位**
```json
{
  "title": "AJO Residence",
  "summary": "鄰近港鐵的服務式住宅",
  "description": "鄰近港鐵的服務式住宅",
  "description_en": "Serviced residence near MTR",
  "district_code": "hk_east",
  "publisher_identity_type": "operator",
  "project_name": "AJO Residence",
  "project_name_en": "AJO Residence",
  "project_attributes": {
    "address_street": "英皇道",
    "address_doorplate": "100號",
    "summary_en": "Serviced residence near MTR",
    "facility_custom_text": "24小時自助洗衣",
    "service_custom_text": "每週清潔"
  },
  "address_text": "英皇道 100號",
  "website_url": "https://example.com",
	  "whatsapp": "85261234567",
	  "fax": "30000000",
	  "lowest_monthly_rent_hkd": 18000,
	  "highest_monthly_rent_hkd": 22000,
	  "min_stay_value": 1,
  "min_stay_unit": "month",
  "facility_tags": ["gym", "laundry"],
  "service_tags": ["housekeeping", "wifi"],
  "room_types": [
    {
      "name": "開放式",
      "name_en": "Studio",
      "room_category": "studio",
      "usable_area_min_sqft": 180,
      "usable_area_max_sqft": 260,
      "monthly_rent_min_hkd": 18000,
      "monthly_rent_max_hkd": 22000,
      "rent_unit": "month",
      "rent_suffix_plus": true,
      "min_stay_value": 1,
      "min_stay_unit": "month",
      "included_fee_items": ["appliance_tv", "appliance_microwave"]
    }
  ],
  "contact": {
    "phone": "30000000",
    "whatsapp": "85261234567",
    "wechat": "ajo-residence",
    "email": "info@example.com",
    "show_phone": true,
    "show_whatsapp": true,
    "show_chat": true
  }
}
```
- **房型租金規則**: `rent_unit` 目前前端只開放 `week` 或 `month`；`rent_suffix_plus=true` 時前台可顯示 `XXX+`。
- **圖片欄位**: xlsx 內的房型圖片與相片標籤暫不納入本輪欄位契約。

---

### 67-69. /api/v1/staff/property-sales/{listingId}/{action} [POST]
- **簡介**: Staff 對樓盤租售執行狀態操作。`action` 支援 `publish`、`deactivate`、`renew`。
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
        "title": "新訊息",
        "body": "See you tonight at the lobby.",
        "related_type": "chat",
        "related_public_id": "01KCHAT001",
        "is_read": true,
        "read_at": "2026-04-17T05:52:00Z",
        "created_at": "2026-04-17T05:50:00Z"
      },
      {
        "notification_id": "01KNOTIFY003",
        "category": "system_notice",
        "title": "系統維護通知",
        "body": "今晚 10 時系統會進行例行維護。",
        "related_type": "notification",
        "related_public_id": "",
        "is_read": false,
        "created_at": "2026-06-29T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 3
    },
    "unread_count": 2
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

### 99. /api/v1/staff/system-notices [POST]
- **簡介**: 發布全站系統通知。系統會為每位 active 會員建立一則通知中心記錄，不會寫入訊息管理會話。
- **請求參數**
```json
{
  "title": "系統維護通知", // 必填
  "body": "今晚 10 時系統會進行例行維護。" // 必填
}
```
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "delivered_count": 120
  },
  "request_id": "01KPCXEXAMPLE",
  "timestamp": "2026-06-29T10:00:00Z"
}
```
- **Curl測試**
```bash
curl -X POST "http://127.0.0.1:8080/api/v1/staff/system-notices" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "系統維護通知",
    "body": "今晚 10 時系統會進行例行維護。"
  }'
```
- **Powershell測試**
```powershell
$headers=@{"Authorization"="Bearer $token";"Content-Type"="application/json"}
$body=@{
  title="系統維護通知"
  body="今晚 10 時系統會進行例行維護。"
}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/staff/system-notices" -Method POST -Headers $headers -Body $body
```

### 80. /api/v1/market-trends/rent [GET]
- **簡介**: 公開香港住宅租金走勢，資料來自香港 DATA.GOV.HK / 差餉物業估價署「Private Domestic - Average Rents by Class - Monthly」CSV。
- **請求參數**: 無
- **回應參數**
```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "source": "Rating and Valuation Department, DATA.GOV.HK",
    "source_url": "https://www.rvd.gov.hk/datagovhk/1.1M.csv",
    "dataset": "Private Domestic - Average Rents by Class - Monthly",
    "unit": "HKD per square metre per month",
    "display_unit": "HKD per square foot per month",
    "updated_month": "2026-05",
    "method": "A 至 C 類私人住宅各區平均租金簡單平均，並由每平方米月租換算為每平方呎月租。",
    "months": ["2025-12", "2026-01", "2026-02", "2026-03", "2026-04", "2026-05"],
    "regions": [
      {
        "key": "hk",
        "label": "香港島",
        "latest_hkd_per_sqft": 44.9,
        "monthly_change_percent": 3.1,
        "direction": "up",
        "points": [
          {
            "month": "2026-05",
            "label": "2026年5月",
            "value_hkd_per_sqft": 44.9,
            "source_value_hkd_per_sqm": 483
          }
        ]
      }
    ]
  }
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/market-trends/rent"
```

### 43. /api/v1/supermarket-offers/summary [GET]
- **簡介**: 公開超市優惠摘要，資料來自已部署 good-price 服務
- **請求參數**: 無
- **回應參數**: good-price summary 原始資料，包裝於 AJO 統一回應 `data`；當 good-price 暫時不可達時，回傳可渲染的空摘要結構與 `upstreamAvailable=false`
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/supermarket-offers/summary"
```

### 44. /api/v1/supermarket-offers/search [GET]
- **簡介**: 公開搜尋超市商品與優惠
- **請求參數**
```json
{
  "q": "商品、品牌或分類",
  "category": "分類",
  "brand": "品牌",
  "store": "商店代碼",
  "offerOnly": true,
  "sort": "discount",
  "page": 1,
  "pageSize": 20
}
```
- **回應參數**: good-price search 原始資料，包裝於 AJO 統一回應 `data`；當 good-price 暫時不可達時，回傳可渲染的空搜尋結構與 `upstreamAvailable=false`
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/supermarket-offers/search?offerOnly=true&page=1&pageSize=20"
```

### 45. /api/v1/supermarket-offers/products/{code} [GET]
- **簡介**: 公開超市商品詳情，登入時附帶 AJO 收藏與價格提示狀態
- **請求參數**
```json
{
  "code": "P000000001",
  "days": 90
}
```
- **回應參數**: good-price product detail 原始資料，並附加 `isFavorite`、`alertRule`
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/supermarket-offers/products/P000000001?days=90"
```

### 46. /api/v1/me/supermarket-offers/favorites [GET/POST]
- **簡介**: 取得或新增目前會員的超市商品收藏
- **POST 請求參數**
```json
{
  "productCode": "P000000001"
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/me/supermarket-offers/favorites" \
  -H "Authorization: Bearer $token"
curl -X POST "http://127.0.0.1:8080/api/v1/me/supermarket-offers/favorites" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{"productCode":"P000000001"}'
```

### 47. /api/v1/me/supermarket-offers/favorites/{code} [DELETE]
- **簡介**: 移除目前會員的超市商品收藏
- **Curl測試**
```bash
curl -X DELETE "http://127.0.0.1:8080/api/v1/me/supermarket-offers/favorites/P000000001" \
  -H "Authorization: Bearer $token"
```

### 48. /api/v1/me/supermarket-offers/price-alerts [GET/POST]
- **簡介**: 取得或建立目前會員的超市價格提示
- **POST 請求參數**
```json
{
  "productCode": "P000000001",
  "targetPrice": 20.5,
  "priceMode": "effective",
  "offerRequired": true,
  "enabled": true
}
```
- **Curl測試**
```bash
curl -X GET "http://127.0.0.1:8080/api/v1/me/supermarket-offers/price-alerts" \
  -H "Authorization: Bearer $token"
curl -X POST "http://127.0.0.1:8080/api/v1/me/supermarket-offers/price-alerts" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{"productCode":"P000000001","targetPrice":20.5,"priceMode":"effective","offerRequired":true,"enabled":true}'
```

### 49. /api/v1/me/supermarket-offers/price-alerts/{id} [PATCH/DELETE]
- **簡介**: 更新或刪除目前會員的超市價格提示
- **PATCH 請求參數**
```json
{
  "targetPrice": 18.8,
  "priceMode": "effective",
  "offerRequired": false,
  "enabled": true
}
```
- **Curl測試**
```bash
curl -X PATCH "http://127.0.0.1:8080/api/v1/me/supermarket-offers/price-alerts/1" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{"targetPrice":18.8,"priceMode":"effective","offerRequired":false,"enabled":true}'
curl -X DELETE "http://127.0.0.1:8080/api/v1/me/supermarket-offers/price-alerts/1" \
  -H "Authorization: Bearer $token"
```

---
