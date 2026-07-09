# iSmart 整合 API 參考

## 目的

本文檔是 iSmart Django 應用目前對外提供的程式化端點整合參考。

本文涵蓋：

- 位於 `/api/v1/integration/...` 下的新主要整合端點
- 仍保留以維持向後相容的既有舊路徑
- POS 與支付流程 API
- webhook 類型的機器端點

本文檔依據專案 URL 設定中目前已註冊的端點整理而成。

## 環境

- production: `https://ismart.ajoliving.com`
- clouddev: `https://clouddev.ismart.ajoliving.com`

部分 smartliving 與門禁控制器功能可能無法在 `clouddev` 正常運作，因為硬體連線取決於環境。

## 路徑策略

主要整合路由現在使用：

- `/api/v1/integration/...`

舊路由仍保留並可繼續使用：

- 舊 `/api/v1/...` 路徑
- 舊 `/api/v1/external/...` 路徑
- 位於 `/api/v1` 之外的舊 webhook 路徑

客戶端應逐步遷移到新的整合路徑，但既有客戶端不需要立即變更。

## 目前端點清單

| 分組 | 方法 | 主要路徑 | 舊版路徑 | 用途 |
|---|---|---|---|---|
| Auth | `POST` | `/api/v1/integration/auth/login/` | `/api/v1/pos/login` | POS 或外部客戶端憑證檢查 |
| Auth | `POST` | `/api/v1/integration/auth/register/` | — | 直接建立 iSmart `CustomUser` 帳戶，無需審批 |
| Building data | `GET` | `/api/v1/integration/buildings/receivables/management-fees/` | `/api/v1/building-mf-table/` | 管理費應收摘要 |
| Building data | `GET` | `/api/v1/integration/buildings/receivables/other-fees/` | `/api/v1/building-of-list/` | 其他費用應收列表 |
| Building data | `GET` | `/api/v1/integration/buildings/notices/` | `/api/v1/building-notices/` | 有效大廈通告 |
| External building | `GET` | `/api/v1/integration/buildings/info/` | `/api/v1/external/building-info/` | 大廈資料與文件中繼資料 |
| Member role binding | `POST` | `/api/v1/integration/buildings/building-flat-owner-binding-requests/` | — | 提交 `OwnerReg` 業主角色綁定申請，需要職員審批 |
| Member subaccount | `GET` | `/api/v1/integration/buildings/subaccounts/` | — | 列出業主控制單位目前有效的授權子用戶 |
| Member subaccount | `POST` | `/api/v1/integration/buildings/subaccounts/grant/` | — | 將某個業主控制單位的 `授權用戶` 授權給另一個 `CustomUser` |
| Member subaccount | `POST` | `/api/v1/integration/buildings/subaccounts/revoke/` | — | 從某個業主控制單位撤銷 `授權用戶` |
| External building | `POST` | `/api/v1/integration/buildings/comments/` | `/api/v1/external/blg-cs/submit/` | 提交大廈意見或投訴 |
| External door access | `POST` | `/api/v1/integration/access/buildings/` | `/api/v1/external/building-access/` | 門禁列表、密碼、QR 資訊與近期紀錄 |
| External door access | `POST` | `/api/v1/integration/access/open-door/` | `/api/v1/external/building-access/open-door/` | 遠端開門 |
| External door access | `POST` | `/api/v1/integration/access/qrcode/` | `/api/v1/external/building-access/qrcode/` | 生成門禁 QR payload |
| Payment | `GET` | `/api/v1/integration/payments/unpaid-invoices/` | `/api/v1/building-flat-unpaid-invoice-list` | 查詢單位未繳賬單 |
| Payment | `POST` | `/api/v1/integration/payments/pos/` | `/api/v1/pos-payment-to-ismart` | 從 POS 或外部支付來源建立付款 |
| Cashier | `GET` | `/api/v1/integration/cashier/transactions/` | `/api/v1/get-transactions-in-cashier` | 查詢某大廈目前在收銀台的付款 |
| Payment | `GET` | `/api/v1/integration/payments/transactions/by-unit/` | `/api/v1/get-transactions-by-flat-unit` | 查詢一個或多個單位的付款歷史 |
| Payment | `GET` | `/api/v1/integration/payments/transactions/by-date/` | `/api/v1/get-transactions-by-date` | 按日期範圍查詢大廈付款 |
| Cashier | `POST` | `/api/v1/integration/cashier/bank-in/` | `/api/v1/update-transactions-status-in-cashier` | 將收銀台付款入賬成 bank-in 批次，並移至待驗證 |
| Cashier | `GET` | `/api/v1/integration/cashier/bank-in-records/` | `/api/v1/get-bank-in-record-list` | 查詢近期 bank-in 批次 |
| Cashier | `GET` | `/api/v1/integration/cashier/bank-in-records/details/` | `/api/v1/get-bank-in-record-details` | 查詢某 bank-in 批次內的付款明細 |
| Webhook | `POST` | `/api/v1/integration/webhooks/qfpay/` | `/qfpayapi/` | QFPay 回調端點 |
| Debug webhook | `POST` | `/api/v1/integration/webhooks/test/` | `/test_webhook/` | 測試用 webhook 接收端點 |

## 安全模型

這些端點並不共用同一套認證方案。整合調用方接入前必須理解目前行為。

### IP 白名單保護

以下端點受 `@ip_whitelist('EXTERNAL_APP_ALLOWED_IPS')` 保護：

- `/api/v1/external/building-info/`
- `/api/v1/external/blg-cs/submit/`
- `/api/v1/external/building-access/`
- `/api/v1/external/building-access/open-door/`
- `/api/v1/external/building-access/qrcode/`

重要事項：

- 白名單在 `stc/settings.py` 中設定
- 若 `EXTERNAL_APP_ALLOWED_IPS` 為空，decorator 實際上會允許所有 IP
- 這些端點不使用 session auth 或 token auth
- 使用者身份由 JSON body 提供，通常是 `user_id`

### 未使用 IP 白名單

以下端點目前未使用 `@ip_whitelist`：

- `/api/v1/pos/login`
- `/api/v1/building-mf-table/`
- `/api/v1/building-of-list/`
- `/api/v1/building-notices/`
- `/api/v1/building-flat-unpaid-invoice-list`
- `/api/v1/pos-payment-to-ismart`
- `/api/v1/get-transactions-in-cashier`
- `/api/v1/get-transactions-by-flat-unit`
- `/api/v1/get-transactions-by-date`
- `/api/v1/update-transactions-status-in-cashier`
- `/api/v1/get-bank-in-record-list`
- `/api/v1/get-bank-in-record-details`

重要事項：

- `PosPaymentToIsmart` 在程式碼中有一個已註解的 `POS_PAYMENT_ALLOWED_IPS` decorator
- `qfpayapi` 只檢查 header `X-QF-SIGN` 是否存在，並不驗證簽名
- `test_webhook/` 是除錯端點，不應視為生產整合 API

## 共用請求約定

- 方法依端點而定
- 整合讀取型端點現在使用 `GET` 搭配 query parameters
- 舊別名仍以 `POST` 保留，通常使用 JSON request body
- content type: `POST` API 使用 `application/json`，但 webhook 發送方可能使用自己的 payload 形態
- CSRF: 本文件記錄的所有 `api/v1` 端點均豁免 CSRF
- timezone: 多個支付端點會手動加減 8 小時以轉換香港時間

## 共用回應約定

回應格式尚未完全標準化。

目前使用中的格式包括：

- wrapped success: `{"status":"success","data":...}`
- wrapped error: `{"status":"error","message":"..."}`
- legacy success: `{"code":"200","message":"success",...}`
- POS login success: `{"code":"1","msg":...}`
- raw array success: `[...]`
- plain text webhook response: `SUCCESS`, `UNSUCCESS`, or `success`

客戶端應按各端點自身契約解析，不應假設存在全域統一的 response envelope。

## 錯誤碼

目前常用 HTTP status codes：

- `200` 請求成功
- `201` 資源已建立
- `400` bad request、JSON 無效、缺少欄位或欄位值無效
- `401` 憑證無效
- `403` IP 不在白名單，或使用者無權存取請求資源
- `404` 找不到引用的使用者、大廈、門禁、QR 紀錄或支付類型
- `405` webhook 端點收到非 POST 請求
- `500` 未處理的伺服器錯誤
- `502` 上游設備或 QR 生成器失敗

## 1. 認證

### 1.1 POS / External Login

- endpoint: `POST /api/v1/integration/auth/login/`
- legacy path: `POST /api/v1/pos/login`
- purpose: 驗證憑證，並返回使用者身份、大廈權限與單位權限列表
- auth model: 不發出 token；此端點只做憑證檢查

#### Request

```json
{
  "login_name": "demo_user",
  "password": "secret",
  "login_type": "username"
}
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `login_name` | string | yes | 依 `login_type` 而定，可為 username、email 或 phone |
| `password` | string | yes | 使用者密碼 |
| `login_type` | string | no | `username`, `email`, or `phone`. Default: `username` |

#### Success Response

```json
{
  "code": "1",
  "msg": {
    "user_id": 123,
    "username": "demo_user",
    "is_staff": false,
    "building": ["0348200"],
    "staff_building_permissions": ["0348200"],
    "client_building_permissions": ["0348200"],
    "client_building_flat_units_permissions": ["0348200001"]
  }
}
```

#### Error Responses

```json
{
  "status": "error",
  "message": "Invalid login type"
}
```

```json
{
  "status": "error",
  "message": "Invalid login credentials"
}
```

#### Notes

- `building` 與 `staff_building_permissions` 目前返回相同資料
- 此端點不建立 session、JWT 或 API token

### 1.2 直接註冊 iSmart 帳戶

- endpoint: `POST /api/v1/integration/auth/register/`
- legacy path: none
- purpose: 直接建立新的 `accounts.CustomUser` 及關聯 `ClientTbl` profile，不經審批流程
- access control: 目前沒有 IP 白名單，也沒有使用者認證

#### Request

```json
{
  "phone": "91234567",
  "email": "user@example.com",
  "eng_name": "CHAN TAI MAN",
  "chi_name": "陳大文",
  "id_card": "A1234567",
  "gender": "M",
  "legal_entity": "NA",
  "remark": "registered from external app",
  "is_receive_email": true
}
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `phone` | string | yes | 主要電話，等同舊會員註冊流程中的 `memberphone` |
| `email` | string | yes | 主要電郵 |
| `eng_name` | string | yes | 英文顯示名稱 |
| `chi_name` | string | no | 中文姓名 |
| `legal_entity` | string | no | `NA` 代表個人，`LE` 代表法人。Default: `NA` |
| `id_card` | string | no | 身份證或證件號碼 |
| `remark` | string | no | 寫入 `ClientTbl.cli_remark` 的自由文字備註 |
| `gender` | string | no | `M`, `F`, or empty |
| `is_receive_email` | boolean | no | 是否啟用 `UserSettings.blg_notice_email`. Default: `true` |
| `password` | string | no | 若省略，後端會生成隨機密碼 |

#### Legacy-Compatible Aliases Also Accepted

| Alias | Normalized Field |
|---|---|
| `memberphone` | `phone` |
| `memberemail` | `email` |
| `memberengname` | `eng_name` |
| `memberchiname` | `chi_name` |
| `member_legalentity` | `legal_entity` |
| `memberid` | `id_card` |
| `regremark` | `remark` |
| `membergender` | `gender` |
| `is_receiveemail` | `is_receive_email` |

#### Success Response

```json
{
  "status": "success",
  "data": {
    "user_id": 123,
    "username": "200123",
    "password": "generated-password",
    "phone": "91234567",
    "email": "user@example.com",
    "client_id": "200123"
  }
}
```

#### Duplicate Account Response

```json
{
  "status": "error",
  "message": "Account already exists",
  "data": {
    "user_id": 123,
    "username": "200123"
  }
}
```

#### Notes

- 若 phone 或 email 已屬於既有 `CustomUser`，API 返回 HTTP `409`
- 若按 phone、email 或 `id_card` 找到匹配 `ClientTbl`，會重用其 `cli_id` 作為新 username
- 此端點不建立 `OwnerReg` row

## 2. 大廈資料 API

### 2.1 大廈管理費表

- endpoint: `GET /api/v1/integration/buildings/receivables/management-fees/`
- legacy path: `POST /api/v1/building-mf-table/`
- purpose: 返回由 `blg_mf_list(...)` 生成的管理費應收資料
- access control: 沒有 IP 白名單，也沒有使用者認證

#### Integration Query Parameters

```text
?building_id=0348200&data_structure=table
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | 大廈 ID |
| `data_structure` | string | no | 回應結構。常見值：`table`, `list`, `block_dict`. Default: `table` |

#### Success Response: `table`

```json
[
  {
    "單位": "G樓  01",
    "2026/06": "已付",
    "2026/05": "已付",
    "2026/04": "審核中",
    "2026/03月前": "已付"
  }
]
```

#### Success Response: `block_dict`

```json
{
  "blocks": [
    {
      "name": "A",
      "floors": [
        {
          "name": "3",
          "units": [
            {
              "name": "01",
              "bills": [
                {
                  "2026/06": 1415.0
                },
                {
                  "Total": 2830.0
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

#### Notes

- 回應形態會隨 `data_structure` 改變
- 值可能是金額，也可能是 `已付`, `審核中`, `錯誤` 等狀態標籤
- 後端以 `group_estate=True` 調用應收資料生成器

### 2.2 大廈其他費用列表

- endpoint: `GET /api/v1/integration/buildings/receivables/other-fees/`
- legacy path: `POST /api/v1/building-of-list/`
- purpose: 返回由 `blg_mf_list(...)` 生成的非管理費應收資料
- access control: 沒有 IP 白名單，也沒有使用者認證

#### Integration Query Parameters

```text
?building_id=0348200&data_structure=list
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | 大廈 ID |
| `data_structure` | string | no | 常見值：`list` 或 `block_dict`. Default: `list` |

#### Success Response: `list`

```json
[
  {
    "invoice_no": "INV000123",
    "flat_code": "0348200001",
    "item_id": "分攤費用",
    "trs_to": "2026/06",
    "trs_val": -350.0,
    "remark": "冷氣維修"
  }
]
```

#### Success Response: `block_dict`

```json
{
  "blocks": [
    {
      "name": "A",
      "floors": [
        {
          "name": "3",
          "units": [
            {
              "name": "01",
              "bills": [
                {
                  "unit_id": "0348200001",
                  "trs_to": "2026/06",
                  "trs_val": -350.0,
                  "item_id": "分攤費用",
                  "remark": "冷氣維修"
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

#### Notes

- 回應形態會隨 `data_structure` 改變
- 空結果返回 `[]`

### 2.3 大廈通告

- endpoint: `GET /api/v1/integration/buildings/notices/`
- legacy path: `POST /api/v1/building-notices/`
- purpose: 返回某大廈的有效通告
- access control: 沒有 IP 白名單，也沒有使用者認證

#### Integration Query Parameters

```text
?building_id=0348200
```

#### Success Response

```json
[
  {
    "id": 18,
    "mess_code": "N20260614",
    "mess_title": "Water Suspension Notice",
    "mess_type": "公告",
    "mess_date": "2026-06-14",
    "mess_down": "2026-06-21",
    "mess_file": "https://..."
  }
]
```

#### Notes

- 回應是 raw array，不是 object wrapper
- 只返回 `mess_date <= today < mess_down` 的通告

## 3. 外部大廈 API

這些端點供可信的 server-to-server 或 app-backend 整合使用。

### 3.0 會員角色綁定

本節涵蓋本次實作的會員與單位綁定流程。

- 業主角色綁定（`登記業主`）透過 `OwnerReg` 提交，仍需要物業管理審批
- 由已審批業主發起的授權子用戶綁定（`授權用戶`）不需要審批
- 實際有效角色關係仍儲存在 `i_flatcli_rel_tbl`
- 另有獨立歷史模型記錄授權子用戶的授權與撤銷操作

### 3.0.1 業主角色綁定申請

- endpoint: `POST /api/v1/integration/buildings/building-flat-owner-binding-requests/`
- legacy path: none
- purpose: 提交 `OwnerReg` 申請，將使用者綁定到一個或多個單位，主要用於業主角色審批流程
- access control: 目前沒有 IP 白名單，也沒有使用者認證

#### Request Example: Existing User Requests Owner Binding

```json
{
  "building_id": "0348200",
  "ownedflat": ["0348200001"],
  "user": 123,
  "cli_role": "業主",
  "ownernote": "request from external app",
  "is_receive_email": true
}
```

#### Request Example: New Applicant Without Existing User

```json
{
  "building_id": "0348200",
  "ownedflat": ["0348200001"],
  "cli_role": "業主",
  "reg_tel": "91234567",
  "reg_email": "user@example.com",
  "cli_name": "CHAN TAI MAN",
  "cli_id_card": "A1234567",
  "cli_tel": "91234567"
}
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | 大廈 ID |
| `ownedflat` | string array or comma-separated string | yes | 一個或多個 `unit_id`，且必須屬於指定大廈 |
| `user` | integer | conditional | 既有 `accounts.CustomUser.id`。若提供，缺失的申請人資料會從關聯 user/client profile 自動補全 |
| `cli_role` | string | no | 預設為 `業主` |
| `ownernote` | string | no | 自由文字備註 |
| `is_receive_email` | boolean | no | 審批通過後是否啟用通告電郵偏好。Default: `false` |
| `reg_tel` | string | conditional | 未提供 `user` 時必填 |
| `reg_email` | string | no | 申請人電郵 |
| `cli_name` | string | conditional | 未提供 `user` 時必填 |
| `cli_id_card` | string | no | 申請人身份證件號碼 |
| `cli_tel` | string | no | 申請人聯絡電話 |

#### Success Response

```json
{
  "status": "success",
  "data": {
    "owner_reg_id": 45,
    "building_id": "0348200",
    "ownedflat": ["0348200001"],
    "user": 123,
    "status": "pending_approval"
  }
}
```

#### Notes

- 此端點只建立 `OwnerReg` row；不會立即授予角色關係
- 後續職員審批仍走既有 `confirm_memreg` 流程
- 若 `OwnerReg.user` 存在，審批時會重用既有帳戶，而不是建立新帳戶

### 3.0.2 列出授權子用戶

- endpoint: `GET /api/v1/integration/buildings/subaccounts/`
- legacy path: none
- purpose: 列出調用者已是已審批 `登記業主` 的單位下的有效 `授權用戶` 關係
- access control: 調用者必須是所請求單位的有效 `登記業主`；若省略 `unit_id`，則包含所有業主控制單位

#### Integration Query Parameters

```text
?user_id=123&unit_id=0348200001
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | 業主的 `accounts.CustomUser.id` |
| `unit_id` | string | no | 可選單位篩選。若省略，返回該使用者所有業主控制單位下的有效子帳戶 |

#### Success Response

```json
{
  "status": "success",
  "data": {
    "items": [
      {
        "relation_info_id": 1001,
        "building_id": "0348200",
        "building_name": "Example Building",
        "unit_id": "0348200001",
        "unit_display": "0348200001",
        "target_user_id": 456,
        "target_username": "200123",
        "target_client_id": "200123",
        "target_name": "CHAN SIU MING",
        "target_phone": "91230000",
        "target_email": "sub@example.com",
        "role": "授權用戶",
        "history_count": 2
      }
    ],
    "count": 1
  }
}
```

#### Notes

- 此端點只返回有效的 `i_flatcli_rel_tbl` rows，且 `cli_role='授權用戶'`
- `history_count` 計算此關係在授權子用戶歷史模型中的記錄數
- 若提供 `unit_id`，但調用者不是已審批業主，API 返回 HTTP `403`

### 3.0.3 授權子用戶

- endpoint: `POST /api/v1/integration/buildings/subaccounts/grant/`
- legacy path: none
- purpose: 讓已審批業主將某單位的 `授權用戶` 授予另一個既有 `CustomUser`，無需職員審批
- access control: 調用者必須已是該單位的有效 `登記業主`

#### Request

```json
{
  "user_id": 123,
  "unit_id": "0348200001",
  "target_user_id": 456,
  "remark": "family member"
}
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | 業主 `accounts.CustomUser.id` |
| `unit_id` | string | yes | 目標單位 ID |
| `target_user_id` | integer | yes | 要成為 `授權用戶` 的既有 `accounts.CustomUser.id` |
| `remark` | string | no | 自由文字稽核備註，會同時儲存在 relation 與 history |

#### Success Response

```json
{
  "status": "success",
  "data": {
    "history_id": 12,
    "relation_info_id": 1001,
    "building_id": "0348200",
    "unit_id": "0348200001",
    "target_user_id": 456,
    "target_username": "200123",
    "role": "授權用戶",
    "active_subaccount_count": 2
  }
}
```

#### Validation Rules

- `user_id` 必須目前是所提供 `unit_id` 的有效 `登記業主`
- `target_user_id` 必須已存在，且有關聯的 `ClientTbl`
- `target_user_id` 不可等於 `user_id`
- 每個單位最多可有 4 個有效 `授權用戶`
- 若目標使用者已對該單位有有效 `授權用戶` 關係，API 返回 HTTP `409`

#### Notes

- 若同一目標使用者曾有舊的非有效 `授權用戶` row，目前程式碼會重新啟用該 row，而不是建立第二個 row
- 每次成功授權都會寫入一筆 `AuthorizedSubUserRequestHistory`

### 3.0.4 撤銷授權子用戶

- endpoint: `POST /api/v1/integration/buildings/subaccounts/revoke/`
- legacy path: none
- purpose: 讓業主停用某單位的一個有效 `授權用戶` 關係
- access control: 調用者必須已是該單位的有效 `登記業主`

#### Request

```json
{
  "user_id": 123,
  "unit_id": "0348200001",
  "target_user_id": 456,
  "remark": "moved out"
}
```

#### Success Response

```json
{
  "status": "success",
  "data": {
    "history_id": 13,
    "relation_info_id": 1001,
    "building_id": "0348200",
    "unit_id": "0348200001",
    "target_user_id": 456,
    "target_username": "200123",
    "role": "授權用戶",
    "active_subaccount_count": 1
  }
}
```

#### Notes

- revoke 透過將既有 `i_flatcli_rel_tbl` row 設為 `is_active=False` 實作
- 每次成功撤銷都會寫入一筆 `AuthorizedSubUserRequestHistory`
- 若要撤銷的使用者目前不是該單位的有效 `授權用戶`，API 返回 HTTP `404`

### 3.1 大廈資料與文件

- endpoint: `GET /api/v1/integration/buildings/info/`
- legacy path: `POST /api/v1/external/building-info/`
- purpose: 返回大廈 profile、大廈資料欄位與大廈文件中繼資料
- access control: `EXTERNAL_APP_ALLOWED_IPS`

#### Integration Query Parameters

```text
?building_id=0348200
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | 目標大廈 ID |

#### Success Response

```json
{
  "status": "success",
  "data": {
    "building": {
      "building_id": "0348200",
      "buildname_chi": "示例大廈",
      "buildname": "Example Building",
      "building_type": "res",
      "area": "",
      "district": "",
      "street": "",
      "street_no": "",
      "court": "",
      "block": ""
    },
    "building_info": {
      "year_built": 1998,
      "total_floor": 32,
      "total_unit": 256,
      "total_carpark": 120,
      "google_map_url": "https://www.google.com/maps/...",
      "owners_corporation_name": "示例法團",
      "management_office_phone": "21234567",
      "management_company_name": "Example PM Co.",
      "management_company_phone": "21234568",
      "management_company_email": "pm@example.com",
      "management_company_fax": "21234569",
      "home_affairs_department_phone": "28350000"
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

#### Document Groups Returned

| Response Field | Source `btype` |
|---|---|
| `documents.forms` | `form` |
| `documents.building_info_files` | `blginfo` |
| `documents.floorplans` | `floorplan` |
| `documents.audit_reports` | queried as `audition` in current code |
| `documents.financial_reports` | `mfinreport` |

#### Notes

- 即使沒有文件，也會返回陣列
- 若沒有 `BuildingInfo` row，`building_info` 會返回 `null` 或空字串值
- 目前程式碼使用 `btype='audition'` 查詢審計報告；若資料使用 `auditreport`，在程式碼對齊前，這些文件不會出現在此端點
- AJO Web `大廈財務` 使用本接口顯示 `管理處總覽`、`財務報表` 與 `核數報表`
- AJO 後端會員態代理會兼容 `floorplan` / `floorplans`、`auditreport` / `audition` / `audit_reports`、`mfinreport` / `financial_reports`

### 3.2 提交大廈意見

- endpoint: `POST /api/v1/integration/buildings/comments/`
- legacy path: `POST /api/v1/external/blg-cs/submit/`
- purpose: 建立 `BlgComment`，並發送既有通知電郵流程
- access control: `EXTERNAL_APP_ALLOWED_IPS` 加使用者對大廈的授權檢查

#### Request

```json
{
  "user_id": 123,
  "building_id": "0348200",
  "comment_type": "其他事宜",
  "comment": "Air-conditioner water leakage at the corridor outside 18/F."
}
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | `accounts.CustomUser.id` |
| `building_id` | string | yes | 目標大廈 ID |
| `comment_type` | string | yes | 必須是允許的 comment type 之一 |
| `comment` | string | yes | 意見內容 |

#### Allowed `comment_type` Values

- `門卡報失`
- `冷氣滴水`
- `嘈音滋擾`
- `樓梯雜物`
- `水質問題`
- `渠務問題`
- `保安事宜`
- `清潔衛生`
- `增加服務`
- `電力問題`
- `其他事宜`

#### Success Response

```json
{
  "status": "success",
  "message": "Comment submitted successfully",
  "data": {
    "comment_id": 98,
    "building_id": "0348200",
    "comment_type": "其他事宜",
    "comment": "Air-conditioner water leakage at the corridor outside 18/F.",
    "created_at": "2026-06-14T10:20:30.000000+08:00"
  }
}
```

#### Notes

- 所提供的 `user_id` 必須有權存取所提供的 `building_id`
- 成功狀態碼是 `201`
- 電郵發送會在請求流程內執行

## 4. 外部門禁 API

### 4.1 大廈門禁摘要

- endpoint: `POST /api/v1/integration/access/buildings/`
- legacy path: `POST /api/v1/external/building-access/`
- purpose: 返回使用者在某大廈的公共門禁列表、鏡頭參考、密碼資訊、QR 資訊與近期門禁紀錄
- access control: `EXTERNAL_APP_ALLOWED_IPS` 加使用者對大廈的授權檢查

#### Request

```json
{
  "user_id": 123,
  "building_id": "0348200"
}
```

#### Success Response

```json
{
  "status": "success",
  "data": {
    "building": {
      "requested_building_id": "0348200",
      "access_building_id": "0348200",
      "buildname_chi": "示例大廈",
      "buildname": "Example Building"
    },
    "doors": [
      {
        "door_id": 10,
        "title": "地下大門",
        "serial": "A123456789",
        "door_no": 1,
        "building_id": "0348200",
        "is_public": true,
        "is_qrcode_enabled": true,
        "has_permission": true,
        "camera": {
          "id": 8,
          "title": "地下大堂鏡頭",
          "url": "rtsp://...",
          "source": "building_cam"
        },
        "password": {
          "record_id": 55,
          "value": "123456",
          "start_time": "2026-06-01T00:00:00+08:00",
          "end_time": "2050-01-01T00:00:00+08:00"
        },
        "qrcode": {
          "record_id": 77,
          "start_time": "2026-06-01T00:00:00+08:00",
          "end_time": "2050-01-01T00:00:00+08:00"
        }
      }
    ],
    "recent_records": [
      {
        "door": {
          "id": 10,
          "title": "地下大門"
        },
        "records": [
          {
            "open_time": "2026-06-14 09:18:00",
            "open_type": "remote",
            "is_success": 1
          }
        ]
      }
    ]
  }
}
```

#### Notes

- `has_permission` 反映 `DoorPermission`
- `password`, `qrcode`, `camera` 可為 `null`
- 目前程式碼會套用兩個舊版大廈映射：
  - `0348300` 映射為 `0348200`
  - `0324900` 映射為 `0325000`
- 因此 `requested_building_id` 和 `access_building_id` 可能不同

### 4.2 遠端開門

- endpoint: `POST /api/v1/integration/access/open-door/`
- legacy path: `POST /api/v1/external/building-access/open-door/`
- purpose: 觸發與 smartliving 網頁相同的遠端開門流程
- access control: `EXTERNAL_APP_ALLOWED_IPS` 加使用者對大廈的授權檢查

#### Request

```json
{
  "user_id": 1,
  "building_id": "0999900",
  "door_id": 1
}
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | `accounts.CustomUser.id` |
| `door_id` | integer | yes | `smartliving.Door.id` |
| `building_id` | string | no | 額外驗證門禁是否屬於該大廈，可選 |

#### Success Response

```json
{
  "status": "success",
  "message": "Door opened successfully",
  "data": {
    "door_id": 10,
    "building_id": "0348200",
    "is_success": true
  }
}
```

#### Upstream Failure Response

```json
{
  "status": "error",
  "message": "Unable to connect to door controller",
  "data": {
    "door_id": 10,
    "building_id": "0348200",
    "is_success": false
  }
}
```

#### Notes

- 無法連接門禁控制器時返回 HTTP `502`
- 目前開門權限邏輯與列表端點不完全一致：
  - 列表檢查 `DoorPermission`
  - 開門檢查使用者在該大廈是否有任意有效 flat-client relationship

### 4.3 生成門禁 QR Payload

- endpoint: `POST /api/v1/integration/access/qrcode/`
- legacy path: `POST /api/v1/external/building-access/qrcode/`
- purpose: 生成 QR payload 字串，由客戶端渲染為 QR 圖片
- access control: `EXTERNAL_APP_ALLOWED_IPS` 加使用者對 QR record 的所有權檢查

#### Request

```json
{
  "user_id": 123,
  "qrcode_record_id": 77,
  "term": "dynamic"
}
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | `accounts.CustomUser.id` |
| `qrcode_record_id` | integer | yes | `DoorPasswordRecord.id`，其中 `passwordorqrcode=True` |
| `term` | string | no | `dynamic` 或 `static`. Default: `dynamic` |

#### Success Response

```json
{
  "status": "success",
  "data": {
    "qrcode_record_id": 77,
    "door_id": 10,
    "term": "dynamic",
    "qrcode_value": "encrypted-qr-payload-from-door-controller",
    "expires_at": "2026-06-14T10:35:30.000000+08:00"
  }
}
```

#### QR Term Rules

- `dynamic`: 到期時間為目前時間加 15 分鐘
- `static`: 到期時間為目前時間加 360 天，但不超過 record `end_time`

#### Notes

- API 只返回 payload 文字；客戶端必須自行渲染 QR 圖片
- 目前程式碼在生成前不會單獨拒絕已過期的 QR record

## 5. 支付 API

### 5.1 單位未繳賬單列表

- endpoint: `GET /api/v1/integration/payments/unpaid-invoices/`
- legacy path: `POST /api/v1/building-flat-unpaid-invoice-list`
- purpose: 返回某單位目前未繳賬單列，已扣除目前在收銀台或待驗證中的付款
- access control: 沒有 IP 白名單，也沒有使用者認證

#### Integration Query Parameters

```text
?unit_id=09999000111
```

#### Success Response

```json
[
  {
    "invoice_no": "INV000123",
    "flat_code": "09999000111",
    "item_id": "管理費",
    "trs_to": "2026/06",
    "bill_dt": "2026-06-01",
    "net_amount": 1500.0,
    "remark": ""
  }
]
```

#### Notes

- 回應是 raw array
- `net_amount` 以元為單位，不是 cents
- 狀態為 `in_cashier` 或 `pending_validation` 的待處理付款會從未繳金額中扣除

### 5.2 POS Payment to iSmart

- endpoint: `POST /api/v1/integration/payments/pos/`
- legacy path: `POST /api/v1/pos-payment-to-ismart`
- purpose: 建立 `Payment` 與 `Paymentdetails`；對部分支付方式，也會立即建立 `BalCalcTbl` 與 `PrepaidBal` rows
- access control: 目前程式碼中沒有啟用 IP 白名單

#### Request Example

```json
{
  "BLG_ID": "0348200",
  "PAY_METHOD": "POS_CASH",
  "FINAL_AMOUNT": 150000,
  "ENTRY_DATETIME": "2026-06-14 10:30:00",
  "TRAN_DATETIME": "2026-06-14 10:30:00",
  "TRAN_REF_NO": "TXN123456789",
  "USER_ID": 1,
  "FAKE_TRANSACTION": false,
  "COMMENT": "",
  "PIC_FILENAME": [],
  "PIC_DATA": [],
  "bank_account_received": null,
  "PAYMENT_GATEWAY_RESPONSE": {},
  "BILL_OBJS": [
    {
      "flat_code": "0348200001",
      "invoice_no": "INV000123",
      "item_id": "管理費",
      "trs_to": "2026/06",
      "net_amount": 150000,
      "transfer_fee": 0
    }
  ]
}
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `BLG_ID` | string | yes | 大廈 ID |
| `PAY_METHOD` | string | yes | 支援值見下表 |
| `FINAL_AMOUNT` | integer | conditional | POS 與 web POS 方式必填，單位為 cents |
| `AMOUNT` | integer | conditional | `allinpay` 必填，單位為 cents |
| `ENTRY_DATETIME` | string | conditional | POS 與 web POS 方式必填，格式 `YYYY-MM-DD HH:mm:ss` |
| `TRAN_DATETIME` | string | conditional | POS 與 web POS 方式必填，格式 `YYYY-MM-DD HH:mm:ss` |
| `DATE` | string | conditional | `allinpay` 必填，格式 `YYYYMMDD` |
| `TIME` | string | conditional | `allinpay` 必填，格式 `HHmmss` |
| `TRAN_REF_NO` | string | no | 外部參考號 |
| `TRANS_TRACE_NO` | string | conditional | `allinpay` 必填 |
| `TRANS_TICKET_NO` | string | conditional | `allinpay` 必填 |
| `USER_ID` | integer | no | 預設為 `1` |
| `FAKE_TRANSACTION` | boolean | no | 若為 true，狀態變為 `pending_validation` |
| `COMMENT` | string | no | 自由文字備註 |
| `PIC_FILENAME` | string array | no | 已上傳圖片檔名 |
| `PIC_DATA` | string array | no | 與 `PIC_FILENAME` 對應的 Base64 圖片內容 |
| `bank_account_received` | string | no | 目標收款銀行帳號 |
| `PAYMENT_GATEWAY_RESPONSE` | object | no | 要保存的原始支付閘道回應 |
| `BILL_OBJS` | object array | yes | 付款明細項目 |

#### `BILL_OBJS` Item Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `flat_code` | string | yes | 單位 ID |
| `invoice_no` | string | no | 賬單號。預付可為空或省略 |
| `item_id` | string | yes | 費用項目名稱 |
| `trs_to` | string | yes | 賬期或 `預付` |
| `net_amount` | integer | yes | 金額，單位為 cents |
| `transfer_fee` | integer | no | 手續費，單位為 cents。Default: `0` |

#### Supported `PAY_METHOD` Values

| Value | Initial Status | Notes |
|---|---|---|
| `POS_CASH` | `in_cashier` | 現金付款 |
| `POS_CHEQUE` | `in_cashier` | 支票付款 |
| `POS_BANK` | `pending_validation` | 銀行轉賬 |
| `POS_ALIWE` | `in_cashier` | POS 支付寶 / 微信 |
| `WEBPOS_ALIPAY` | `payment_captured` | 立即寫入會計記錄 |
| `WEBPOS_WECHAT` | `payment_captured` | 立即寫入會計記錄 |
| `WEBPOS_CARD_UP` | `payment_captured` | 立即寫入會計記錄 |
| `allinpay` | `init` | 特殊舊流程 |

#### Success Response

```json
{
  "code": "200",
  "message": "success",
  "receipt_id": "1000042"
}
```

#### Validation Rules

- 大廈必須存在
- `PAY_METHOD` 必須是該大廈已配置的 `BuildingPaymentType`
- 對非預付明細，請求付款金額不可超過目前未繳金額
- 不支援的 pay method 返回 `400`

#### Important Behavior

- 所有請求金額值均以 cents 為單位
- 建立的 `Payment.txamount` 與明細金額在內部以元儲存
- `WEBPOS_*` 方法會立即建立 payment-side accounting rows
- `trs_to = "預付"` 的預付明細會建立 `PrepaidBal`
- 手續費會為 `平台手續費` 建立獨立明細和會計 rows

#### Important Caveats

- 此端點的 whitelist decorator 目前被註解
- 未知的 `bank_account_received` 會被靜默忽略，而不是拒絕
- `allinpay` 的日期時間解析路徑屬於舊流程，使用前應仔細測試

### 5.3 收銀台交易

- endpoint: `GET /api/v1/integration/cashier/transactions/`
- legacy path: `POST /api/v1/get-transactions-in-cashier`
- purpose: 列出某大廈目前在收銀台中的交易

#### Integration Query Parameters

```text
?building_id=0348200
```

#### Success Response

```json
{
  "payment_objs_cheque": [
    {
      "payment_id": "123",
      "input_time": "2026-06-14 10:30:00",
      "trs_val": 1500.0,
      "receipt_id": "1000042",
      "ref_no": "TXN123456789",
      "payment_detail_objs": [
        {
          "floor": "18",
          "unit": "A",
          "item_id": "管理費",
          "term": "2026/06",
          "trs_val": 1500.0,
          "remark": ""
        }
      ]
    }
  ],
  "payment_objs_cash": []
}
```

#### Notes

- 只返回狀態為 `in_cashier` 的付款
- 結果拆分為 `payment_objs_cheque` 與 `payment_objs_cash`
- 端點不返回統一列表

### 5.4 按單位查詢交易

- endpoint: `GET /api/v1/integration/payments/transactions/by-unit/`
- legacy path: `POST /api/v1/get-transactions-by-flat-unit`
- purpose: 列出一個或多個單位的付款歷史

#### Integration Query Parameters

```text
?unit_id_list=0348200001&unit_id_list=0348200002
```

#### Success Response

```json
{
  "payment_objs": [
    {
      "payment_id": "123",
      "input_time": "2026-06-14 10:30:00",
      "tran_time": "2026-06-14 10:30:00",
      "trs_val": 1500.0,
      "receipt_id": "1000042",
      "ref_no": "TXN123456789",
      "pay_type": "POS_CASH",
      "status": "confirmed",
      "payment_detail_objs": [
        {
          "floor": "18",
          "unit": "A",
          "item_id": "管理費",
          "term": "2026/06",
          "trs_val": 1500.0,
          "remark": ""
        }
      ]
    }
  ]
}
```

#### Notes

- `init` 與 `pending_validation` 會被排除
- `validated_by_ismart`, `payment_captured`, `confirmed` 會被標準化為 `confirmed`
- 預付扣款可能顯示為 `"{trs_to}(預付)"`

### 5.5 按日期查詢交易

- endpoint: `GET /api/v1/integration/payments/transactions/by-date/`
- legacy path: `POST /api/v1/get-transactions-by-date`
- purpose: 按日期範圍篩選大廈付款

#### Integration Query Parameters

```text
?building_id=0348200&from_date=2026-06-01&to_date=2026-06-30&date_type=input_date&pay_method=all
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | 大廈 ID |
| `from_date` | string | yes | 起始日期 |
| `to_date` | string | yes | 結束日期 |
| `date_type` | string | yes | `input_date` 或 `tran_date` |
| `pay_method` | string | no | `all` 或 `pos` |

#### Success Response

```json
{
  "payment_objs": [
    {
      "payment_id": "123",
      "input_time": "2026-06-14 10:30:00",
      "tran_time": "2026-06-14 10:30:00",
      "trs_val": 1500.0,
      "receipt_id": "1000042",
      "ref_no": "TXN123456789",
      "pay_type": "POS_CASH",
      "status": "in_cashier",
      "payment_detail_objs": [
        {
          "floor": "18",
          "unit": "A",
          "item_id": "管理費",
          "term": "2026/06",
          "trs_val": 1500.0,
          "remark": ""
        }
      ]
    }
  ]
}
```

#### Notes

- `date_type` 控制使用 `txdtm` 或 `tran_date` 篩選
- `pay_method = "pos"` 只保留名稱以 `POS` 開頭的 payment types
- 查詢會在日期範圍前後套用手動 `-8h` 與 `+8h` 調整

## 6. 收銀台與 Bank-In API

### 6.1 更新收銀台交易狀態

- endpoint: `POST /api/v1/integration/cashier/bank-in/`
- legacy path: `POST /api/v1/update-transactions-status-in-cashier`
- purpose: 將選中的收銀台付款分組為新的 bank-in 批次，並更新為 `pending_validation`

#### Request

```json
{
  "payment_id_list": [101, 102, 103]
}
```

#### Success Response

```json
{
  "code": "200",
  "message": "success",
  "bank_in_record_id": "88"
}
```

#### Notes

- 所有選中付款目前必須是 `in_cashier`
- 此端點在使用第一筆 row 作為批次中繼資料前，沒有額外驗證所有選中 rows 是否屬於同一大廈與同一 pay type

### 6.2 Bank-In 紀錄列表

- endpoint: `GET /api/v1/integration/cashier/bank-in-records/`
- legacy path: `POST /api/v1/get-bank-in-record-list`
- purpose: 返回某大廈最新 bank-in 批次

#### Integration Query Parameters

```text
?building_id=0348200
```

#### Success Response

```json
{
  "bank_in_record_list": [
    {
      "record_id": 88,
      "building_id": "0348200",
      "bankin_time": "2026-06-14 17:20:00",
      "pay_type": "POS_CASH",
      "trs_val": 3500.0
    }
  ]
}
```

#### Notes

- 最多返回 30 筆紀錄

### 6.3 Bank-In 紀錄詳情

- endpoint: `GET /api/v1/integration/cashier/bank-in-records/details/`
- legacy path: `POST /api/v1/get-bank-in-record-details`
- purpose: 返回某 bank-in 批次內的詳細付款 rows

#### Integration Query Parameters

```text
?building_id=0348200&record_id=88
```

#### Success Response

```json
{
  "bank_in_record_details": [
    {
      "id": 123,
      "payment_id": 123,
      "input_time": "2026-06-14 10:30:00",
      "txamount": 1500.0,
      "receipt_id": "1000042",
      "ref_no": "TXN123456789",
      "floor": "18",
      "unit": "A",
      "invoice_no": "INV000123",
      "trs_val": 1500.0,
      "item_id": "管理費",
      "term": "2026/06",
      "remark": "",
      "pay_type": "POS_CASH"
    }
  ]
}
```

#### Notes

- 若批次 ID 存在但沒有匹配 rows，API 返回空列表

## 7. Webhook 端點

這些是機器端點，但不遵循上述 JSON API 的同一套契約格式。

### 7.1 QFPay 回調

- endpoint: `POST /api/v1/integration/webhooks/qfpay/`
- legacy path: `POST /qfpayapi/`
- purpose: 儲存回調 payload，並完成成功 QFPay 付款通知

#### Expected Inputs

- header: `X-QF-SIGN`
- body: JSON payload from QFPay

目前程式碼引用的欄位包括：

- `notify_type`
- `status`
- `out_trade_no`

#### Current Behavior

- 若 header `X-QF-SIGN` 存在，端點返回 plain text `SUCCESS`
- 若 body 表示成功付款回調，端點載入 `Payment.id = out_trade_no` 並執行舊成功處理器
- 若缺少 header，端點仍會儲存 payload，並返回 plain text `UNSUCCESS`

#### Important Notes

- 目前程式碼只檢查 `X-QF-SIGN` 是否存在；不驗證簽名
- response body 是 plain text，不是 JSON

### 7.2 測試 Webhook

- endpoint: `POST /api/v1/integration/webhooks/test/`
- legacy path: `POST /test_webhook/`
- purpose: 儲存並電郵傳入 webhook payload，以供除錯

#### Current Behavior

- 接受任意 `POST` body
- 儲存 raw body 與 content type
- 嘗試電郵發送診斷副本
- 返回 plain text `success`
- 對非 `POST` 返回 HTTP `405`

#### Important Notes

- 這是除錯工具端點，不是生產業務 API
- 沒有 authentication、signature validation 或 schema validation

## 驗證規則摘要

各 API 常見驗證模式：

- 大廈 ID 通常限制為只允許英數字
- 格式錯誤 JSON 返回 `400`
- 缺少必填欄位返回 `400`
- 大廈、使用者、門禁或 QR 查找失敗返回 `404`
- 外部端點的授權檢查通常依賴 request body 中提供的 `user_id`

## 已知契約注意事項

整合方應注意目前實作中的以下不一致：

- 各端點 response envelope 不一致
- 有些端點返回 raw arrays，有些返回 wrapped objects
- 支付端點混用 string status codes 與 `status` / `message` objects
- 多個支付查詢使用 raw SQL 與手動 timezone shifting
- `PosPaymentToIsmart` 目前未啟用 IP 白名單保護
- `qfpayapi` 不驗證回調簽名
- `ExternalBuildingInfoApi` 目前使用 `btype='audition'` 查詢審計報告
- 門禁列表權限邏輯與遠端開門權限邏輯不完全一致

## 方法設計建議

在目前相容策略下：

- 新的 read-only `/api/v1/integration/...` 端點現在使用 `GET`
- 舊非 integration aliases 仍以 `POST` 保留
- 寫入操作、webhooks 與敏感的 body-driven access flows 仍維持 `POST`

以下整合端點已使用 `GET`：

- `/api/v1/integration/buildings/receivables/management-fees/`
- `/api/v1/integration/buildings/receivables/other-fees/`
- `/api/v1/integration/buildings/notices/`
- `/api/v1/integration/buildings/info/`
- `/api/v1/integration/payments/unpaid-invoices/`
- `/api/v1/integration/cashier/transactions/`
- `/api/v1/integration/payments/transactions/by-unit/`
- `/api/v1/integration/payments/transactions/by-date/`
- `/api/v1/integration/cashier/bank-in-records/`
- `/api/v1/integration/cashier/bank-in-records/details/`

以下仍使用 `POST`：

- `/api/v1/integration/auth/login/`
- `/api/v1/integration/buildings/comments/`
- `/api/v1/integration/access/open-door/`
- `/api/v1/integration/access/qrcode/`
- `/api/v1/integration/payments/pos/`
- `/api/v1/integration/cashier/bank-in/`
- `/api/v1/integration/webhooks/qfpay/`
- `/api/v1/integration/webhooks/test/`

`/api/v1/integration/access/buildings/` 屬於混合情況：

- 語義上是 read-only，因此可使用 `GET`
- 實作上目前依賴 body 提供的 `user_id`，並返回敏感的門禁與密碼相關資料
- 在認證模型從 body identifier 升級到 token-based authentication 前，保留 `POST` 會更安全

## cURL 請求範例

### Building Info

```bash
curl -G "https://<your-domain>/api/v1/integration/buildings/info/" \
  --data-urlencode "building_id=0348200"
```

### Building Access Summary

```bash
curl -X POST "https://<your-domain>/api/v1/integration/access/buildings/" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "building_id": "0348200"
  }'
```

### POS Payment

```bash
curl -X POST "https://<your-domain>/api/v1/integration/payments/pos/" \
  -H "Content-Type: application/json" \
  -d '{
    "BLG_ID": "0348200",
    "PAY_METHOD": "POS_CASH",
    "FINAL_AMOUNT": 150000,
    "ENTRY_DATETIME": "2026-06-14 10:30:00",
    "TRAN_DATETIME": "2026-06-14 10:30:00",
    "TRAN_REF_NO": "TXN123456789",
    "USER_ID": 1,
    "BILL_OBJS": [
      {
        "flat_code": "0348200001",
        "invoice_no": "INV000123",
        "item_id": "管理費",
        "trs_to": "2026/06",
        "net_amount": 150000,
        "transfer_fee": 0
      }
    ]
  }'
```

## AJO 後端集成設計

AJO Web 前端不直接調用 iSmart 原始整合端點；前端統一調用 AJO 後端會員態接口，由 AJO 後端使用目前登入使用者的 iSmart 綁定資料補齊 `user_id` 並執行可見範圍校驗。

### Payment 類接口

POS payment、收銀台、bank-in 與支付歷史兼容優先落在 `ismart_pos_relay`：

- 舊 `/api/...` route 保持兼容，舊 upstream 不可用時可 fallback 到新 `/api/v1/integration/...`
- 新 `/api/v1/integration/...` route 也在 relay 中提供，並可在新 upstream 不可用時按保守規則 fallback 到舊 upstream
- 寫入類接口不對普通業務錯誤自動重送，避免重複付款或重複 bank-in
- base URL 由 `ISMART_OLD_API_BASE_URL` 與 `ISMART_INTEGRATION_BASE_URL` 控制

### AJO 會員態 iSmart 接口

AJO 後端提供以下會員態入口：

| AJO Path | Method | Upstream |
|---|---|---|
| `/api/v1/me/ismart/account-registration` | `POST` | `/api/v1/integration/auth/register/` |
| `/api/v1/me/ismart/buildings` | `GET` | AJO 本地 iSmart 綁定快照 |
| `/api/v1/me/ismart/building-info` | `GET` | `/api/v1/integration/buildings/info/`，必要時 fallback 舊 external path |
| `/api/v1/me/ismart/management-fees` | `GET` | `/api/v1/integration/buildings/receivables/management-fees/` |
| `/api/v1/me/ismart/other-fees` | `GET` | `/api/v1/integration/buildings/receivables/other-fees/` |
| `/api/v1/me/ismart/notices` | `GET` | `/api/v1/integration/buildings/notices/` |
| `/api/v1/me/ismart/building-comments` | `POST` | `/api/v1/integration/buildings/comments/` |
| `/api/v1/me/ismart/owner-binding-requests` | `POST` | `/api/v1/integration/buildings/building-flat-owner-binding-requests/` |
| `/api/v1/me/ismart/subaccounts` | `GET` | `/api/v1/integration/buildings/subaccounts/` |
| `/api/v1/me/ismart/subaccounts/grant` | `POST` | `/api/v1/integration/buildings/subaccounts/grant/` |
| `/api/v1/me/ismart/subaccounts/revoke` | `POST` | `/api/v1/integration/buildings/subaccounts/revoke/` |
| `/api/v1/me/ismart/building-access` | `GET` | `/api/v1/integration/access/buildings/` |
| `/api/v1/me/ismart/building-access/open-door` | `POST` | `/api/v1/integration/access/open-door/` |
| `/api/v1/me/ismart/building-access/qrcode` | `POST` | `/api/v1/integration/access/qrcode/` |

### AJO Web 頁面使用約定

| AJO Web Page | 前端區塊 | AJO 會員態接口 | 說明 |
|---|---|---|---|
| `我的大廈` | `大廈財務 > 管理處總覽` | `/api/v1/me/ismart/building-info`, `/api/v1/me/ismart/management-fees`, `/api/v1/me/ismart/other-fees` | 顯示管理處資料、管理費應收摘要與其他費用應收摘要。 |
| `我的大廈` | `大廈財務 > 財務報表` | `/api/v1/me/ismart/building-info` | 使用 `documents.financial_reports`，兼容舊 `mfinreport`。 |
| `我的大廈` | `大廈財務 > 核數報表` | `/api/v1/me/ismart/building-info` | 使用 `documents.audit_reports`，兼容舊 `auditreport` 與 `audition`。 |
| `我的大廈` | `業戶帳目 > 未繳費賬單列表` | `/api/v1/me/payments/pos/bills` | 按目前登入會員可見單位讀取未繳賬單；前端不可直接調用 `/api/v1/integration/payments/unpaid-invoices/`。 |
| `我的大廈` | `業戶帳目 > 繳費記錄` | `/api/v1/me/payments/pos/history` | 按目前登入會員可見單位讀取付款歷史；前端不可直接調用 `/api/v1/integration/payments/transactions/by-unit/` 或 `/api/v1/integration/payments/transactions/by-date/`。 |

AJO 後端使用以下環境變數區分新舊 upstream：

`/api/v1/me/ismart/building-info` 會保留並補齊 `documents.forms`、`documents.building_info_files`、`documents.floorplans`、`documents.audit_reports`、`documents.financial_reports` 陣列；AJO Web `大廈財務` 直接使用其中 `building_info`、`documents.financial_reports` 與 `documents.audit_reports`。

```bash
ISMART_EXTERNAL_APP_BASE_URL=https://ismart.ajoliving.com
ISMART_EXTERNAL_APP_API_BASE_URL=https://ismart.ajoliving.com/api/v1/external
ISMART_INTEGRATION_API_BASE_URL=https://ismart.ajoliving.com/api/v1/integration
```

## 操作檢查清單

- 為生產整合配置 `EXTERNAL_APP_ALLOWED_IPS`
- 決定 `PosPaymentToIsmart` 是否應受 `POS_PAYMENT_ALLOWED_IPS` 保護
- 若審計文件必須出現在 `/api/v1/integration/buildings/info/`，需對齊 audit report 的 `btype` 命名
- 針對目標硬體環境測試遠端開門與 QR 生成
- 將 `test_webhook/` 視為非生產端點；如不需要，應停用或加保護
