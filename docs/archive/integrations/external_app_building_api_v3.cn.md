# iSmart 整合 API 參考

## 目的

本文是 iSmart Django 應用目前已暴露程式化端點的整合參考。

本文涵蓋：

- `/api/v1/integration/...` 下的新主要整合端點
- 為保持向後相容而保留的舊路徑
- POS 與支付流程 API
- webhook 類型的機器端點

本文依據專案 URL 設定中目前已註冊的端點整理而成。

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
| 認證 | `POST` | `/api/v1/integration/auth/login/` | `/api/v1/pos/login` | POS 或外部客戶端憑證檢查 |
| 認證 | `POST` | `/api/v1/integration/auth/register/` | — | 直接建立 iSmart `CustomUser` 帳戶，無需審批 |
| 認證 | `GET` / `POST` | `/api/v1/integration/auth/check-contact/` | — | 檢查電郵／電話是否已被 `CustomUser` 使用 |
| 認證 | `GET` / `POST` / `PATCH` / `PUT` | `/api/v1/integration/auth/client/` | — | 取得或更新主要 `ClientTbl` 會員資料及登入電郵／電話 |
| 認證 | `POST` | `/api/v1/integration/auth/password/` | — | 驗證目前密碼後修改會員密碼 |
| 通知設定 | `GET` / `PATCH` | `/api/v1/integration/auth/notification-settings/` | — | 讀取或更新大廈通告電郵偏好 |
| 大廈資料 | `GET` | `/api/v1/integration/buildings/receivables/management-fees/` | `/api/v1/building-mf-table/` | 管理費應收摘要 |
| 大廈資料 | `GET` | `/api/v1/integration/buildings/receivables/other-fees/` | `/api/v1/building-of-list/` | 其他費用應收列表 |
| 大廈資料 | `GET` | `/api/v1/integration/buildings/notices/` | `/api/v1/building-notices/` | 有效大廈通告 |
| 外部大廈 | `GET` | `/api/v1/integration/buildings/info/` | `/api/v1/external/building-info/` | 大廈資料與文件中繼資料 |
| 會員角色綁定 | `POST` | `/api/v1/integration/buildings/building-flat-owner-binding-requests/` | — | 提交 `OwnerReg` 業主角色綁定申請，需要職員審批 |
| 會員副戶 | `GET` | `/api/v1/integration/buildings/subaccounts/` | — | 列出業主控制單位目前有效的授權子用戶 |
| 會員副戶 | `POST` | `/api/v1/integration/buildings/subaccounts/grant/` | — | 將 `授權用戶` 授權給某個業主控制單位 |
| 會員副戶 | `POST` | `/api/v1/integration/buildings/subaccounts/revoke/` | — | 從某個業主控制單位撤銷 `授權用戶` |
| 外部大廈 | `POST` | `/api/v1/integration/buildings/comments/` | `/api/v1/external/blg-cs/submit/` | 提交大廈服務個案（維修／意見） |
| 外部大廈 | `GET` / `POST` | `/api/v1/integration/buildings/service-cases/` | — | 按狀態篩選大廈服務個案列表 |
| 外部大廈 | `GET` / `POST` | `/api/v1/integration/buildings/service-cases/<case_id>/` | — | 服務個案詳情及訊息串 |
| 外部門禁 | `POST` | `/api/v1/integration/access/buildings/` | `/api/v1/external/building-access/` | 門禁列表、密碼、QR 資訊與近期紀錄 |
| 外部門禁 | `POST` | `/api/v1/integration/access/open-door/` | `/api/v1/external/building-access/open-door/` | 遠端開門 |
| 外部門禁 | `POST` | `/api/v1/integration/access/qrcode/` | `/api/v1/external/building-access/qrcode/` | 生成門禁 QR payload |
| 支付 | `GET` | `/api/v1/integration/payments/unpaid-invoices/` | `/api/v1/building-flat-unpaid-invoice-list` | 查詢單位未繳賬單 |
| 支付 | `POST` | `/api/v1/integration/payments/pos/` | `/api/v1/pos-payment-to-ismart` | 從 POS 或外部支付來源建立付款 |
| 收銀台 | `GET` | `/api/v1/integration/cashier/transactions/` | `/api/v1/get-transactions-in-cashier` | 查詢某大廈目前在收銀台的付款 |
| 支付 | `GET` | `/api/v1/integration/payments/transactions/by-unit/` | `/api/v1/get-transactions-by-flat-unit` | 查詢單位付款歷史 |
| 支付 | `GET` | `/api/v1/integration/payments/transactions/by-date/` | `/api/v1/get-transactions-by-date` | 按日期範圍查詢大廈付款 |
| 收銀台 | `POST` | `/api/v1/integration/cashier/bank-in/` | `/api/v1/update-transactions-status-in-cashier` | 將收銀台付款寫入 bank-in 批次並移至待驗證 |
| 收銀台 | `GET` | `/api/v1/integration/cashier/bank-in-records/` | `/api/v1/get-bank-in-record-list` | 查詢近期 bank-in 批次 |
| 收銀台 | `GET` | `/api/v1/integration/cashier/bank-in-records/details/` | `/api/v1/get-bank-in-record-details` | 查詢 bank-in 批次內的付款明細 |
| Webhook | `POST` | `/api/v1/integration/webhooks/qfpay/` | `/qfpayapi/` | QFPay 回調端點 |
| 除錯 Webhook | `POST` | `/api/v1/integration/webhooks/test/` | `/test_webhook/` | 僅供測試的 webhook 接收端點 |

## 安全模型

這些端點並不共用同一套認證方案。整合調用方接入前必須理解目前行為。

### IP 白名單保護

以下端點受 `@ip_whitelist('EXTERNAL_APP_ALLOWED_IPS')` 保護：

- `/api/v1/integration/auth/password/`
- `/api/v1/integration/auth/notification-settings/`
- `/api/v1/external/building-info/`
- `/api/v1/external/blg-cs/submit/`
- `/api/v1/integration/buildings/service-cases/`
- `/api/v1/integration/buildings/service-cases/<case_id>/`
- `/api/v1/external/building-access/`
- `/api/v1/external/building-access/open-door/`
- `/api/v1/external/building-access/qrcode/`

重要事項：

- 白名單在 `stc/settings.py` 中設定
- 若 `EXTERNAL_APP_ALLOWED_IPS` 為空，decorator 實際上會允許所有 IP
- 這些端點不使用 session auth 或 token auth
- 使用者身份由 JSON body 提供，通常是 `user_id`

### 未使用 IP 白名單

The following currently do not use `@ip_whitelist`:

- `/api/v1/pos/login`
- `/api/v1/integration/auth/login/`
- `/api/v1/integration/auth/register/`
- `/api/v1/integration/auth/check-contact/`
- `/api/v1/integration/auth/client/`
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

- `PosPaymentToIsmart` has a commented-out `POS_PAYMENT_ALLOWED_IPS` decorator in code
- `qfpayapi` only checks whether header `X-QF-SIGN` exists; it does not validate the signature
- `test_webhook/` is a debug endpoint and should not be treated as a production integration API

## 共用請求約定

- method depends on endpoint
- integration read-only endpoints now use `GET` with query parameters
- legacy aliases remain available as `POST`, usually with JSON request bodies
- content type: `application/json` for `POST` APIs, except webhook senders may use their own payload shape
- CSRF: all documented `api/v1` endpoints in this file are CSRF-exempt
- timezone: several payment endpoints manually convert timestamps to Hong Kong time by adding or subtracting 8 hours

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

目前常用 HTTP status code：

- `200` successful request
- `201` resource created
- `400` bad request, invalid JSON, missing field, or invalid field value
- `401` invalid credentials
- `403` IP not whitelisted or user not authorized for the requested resource
- `404` referenced user, building, door, QR record, or payment type not found
- `405` non-POST request on webhook endpoints
- `500` unhandled server error
- `502` upstream device or QR generator failure

## 1. 認證

### 登入電郵與電話格式（適用於註冊、check-contact、會員資料更新）

以下規則適用 to **login credentials** stored on `CustomUser`:

| Credential | Model field | Unique? |
|---|---|---|
| Login email | `CustomUser.email` | Yes, **case-insensitive** |
| Login phone | `CustomUser.phone` | Yes, after **E.164 normalization** |

Contact / billing fields on `ClientTbl` (`cli_email`, `cli_tel`, `cli_tel2`) are **separate** from login credentials (except on **register**, where the new client row is seeded with the same phone/email for convenience).

#### 電郵格式

| Rule | Detail |
|---|---|
| Required shape | Standard email, e.g. `user@example.com` |
| Normalization | Trimmed and stored **lowercase** (e.g. `User@Example.com` → `user@example.com`) |
| Uniqueness | Case-insensitive: `A@x.com` and `a@x.com` are the **same** account identity |
| Invalid | Names without `@`, truncated domains, phone-only strings, full-width `＠`, etc. → HTTP `400` on register / update when strict validation applies |
| Recommended input | Already-lowercase ASCII email from the client app |

**Examples (accepted):**

```text
user@example.com
User@Example.COM          → stored as user@example.com
yf.huang@nus.edu.sg
```

**Examples (rejected as login email):**

```text
CHU KIT MING
69960833@
youweida@72
94150526＠gmail.com       (full-width ＠)
```

#### 電話格式

| Rule | Detail |
|---|---|
| Storage | Always **E.164** with leading `+`, e.g. `+85291234567` |
| Validation | Must be a real number accepted by Google libphonenumber (`is_valid_number`) |
| Default region | If only national digits are sent (no country code), the server assumes **Hong Kong (`+852`)** for 8-digit numbers |
| Uniqueness | Compared after normalization; legacy variants (`91234567`, `85291234567`, `+85291234567`) match the same login phone |
| Invalid | Wrong length, fake prefixes, free-text → HTTP `400` with a Chinese/English format error message |

**Preferred input (recommended for all external apps): full E.164**

```text
+85291234567      Hong Kong mobile
+8613901234567    Mainland China mobile (11 digits, typically starts with 1)
+6586726893       Singapore mobile
+61420909995      Australia mobile (national mobile without leading 0)
```

**Also accepted (normalized by server):**

| Input | Interpreted as |
|---|---|
| `91234567` | `+85291234567` (HK 8-digit national) |
| `85291234567` | `+85291234567` |
| `8613901234567` | `+8613901234567` |
| `+86 139 0123 4567` | `+8613901234567` (spaces stripped) |

**Do not send:**

| Bad input | Why |
|---|---|
| `85218906680192` | Wrong: `852` glued onto a CN mobile; use `+8618906680192` or country `86` + `18906680192` |
| `12345` | Too short / invalid |
| Landline free text / names | Not a phone number |

**Country-specific notes:**

| Country | Country code | National number | E.164 example |
|---|---|---|---|
| Hong Kong | `852` | 8 digits | `+85291234567` |
| Mainland China | `86` | 11 digits (mobile usually `1…`) | `+8613901234567` |
| Singapore | `65` | 8 digits | `+6586726893` |
| Australia | `61` | 9 digits mobile (without leading `0`) | `+61420909995` |
| Macau | `853` | 8 digits | `+8536xxxxxxx` |
| Taiwan | `886` | typically 9 digits mobile without leading `0` | `+8869xxxxxxxx` |

Other countries are accepted when sent as full **`+` E.164** and libphonenumber considers them valid.

#### 登入欄位與聯絡欄位（會員資料 API）

| API field | Updates | Notes |
|---|---|---|
| `email` | `CustomUser.email` only | **Does not** overwrite `cli_email` |
| `phone` | `CustomUser.phone`; also sets `cli_tel` unless `cli_tel` is sent in the same request | Login phone must be valid E.164 |
| `cli_email` | `ClientTbl.cli_email` only | Contact / billing email; free-form for display but prefer a real email |
| `cli_tel` | `ClientTbl.cli_tel` only | Contact phone; when sent alone does not change login phone |
| `cli_tel2` | Emergency contact | Free-form (not login credential) |

### 1.1 POS / External 登入

- 端點： `POST /api/v1/integration/auth/login/`
- 舊版路徑： `POST /api/v1/pos/login`
- 用途： verifies credentials and returns the user identity plus building and unit permission lists
- 認證模式： no token is issued; this is a credential-check endpoint only

#### 請求

```json
{
  "login_name": "demo_user",
  "password": "secret",
  "login_type": "username"
}
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `login_name` | string | yes | Username, email, or phone depending on `login_type` |
| `password` | string | yes | User password |
| `login_type` | string | no | `username`, `email`, or `phone`. Default: `username` |

For `login_type=email`, matching is **case-insensitive**.  
For `login_type=phone`, prefer E.164 (`+852…`); national HK 8-digit and legacy `852…` forms are also matched when possible.

#### 成功回應

Top-level shape is **unchanged for legacy clients**: `{ "code": "1", "msg": { ... } }`.

Existing `msg` keys keep the same names and types. New fields are **appended only**.

```json
{
  "code": "1",
  "msg": {
    "user_id": 123,
    "username": "demo_user",
    "email": "demo@example.com",
    "phone": "91234567",
    "is_staff": false,
    "last_login": "2026-07-16T10:30:00+08:00",
    "building": ["0348200"],
    "staff_building_permissions": ["0348200"],
    "client_building_permissions": ["0348200"],
    "client_building_flat_units_permissions": ["0348200001"],
    "client": {
      "cli_id": "demo_user",
      "cli_legalentity": "NA",
      "cli_type": null,
      "cli_contact_person": "",
      "cli_urgent_contact_person": "",
      "cli_name": "CHAN TAI MAN",
      "cli_chi_name": "陳大文",
      "cli_acname": "CHAN TAI MAN",
      "cli_join_dt": "2020-01-15",
      "cli_id_card": "A1234567",
      "cli_tel": "91234567",
      "cli_tel2": "",
      "cli_fax": null,
      "cli_addr": "",
      "cli_chiadd": "",
      "cli_email": "demo@example.com",
      "cli_birthday": "1990-01-01",
      "cli_sex": "M",
      "cli_nationality": null,
      "cli_bank": null,
      "cli_bank_acc": null,
      "cli_remark": null
    }
  }
}
```

#### `msg` fields

| Field | Type | Legacy | Description |
|---|---|---|---|
| `user_id` | integer | yes | `CustomUser.id` |
| `username` | string | yes | `CustomUser.username` |
| `email` | string or null | **new** | `CustomUser.email` |
| `phone` | string or null | **new** | `CustomUser.phone` |
| `is_staff` | boolean | yes | `true` if staff or superuser |
| `last_login` | string or null | **new** | ISO-8601 datetime after this successful login |
| `building` | string array | yes | Staff building IDs (same as `staff_building_permissions`) |
| `staff_building_permissions` | string array | yes | Building IDs from `UserBuildingManagementPermission` |
| `client_building_permissions` | string array | yes | Building IDs derived from linked client units |
| `client_building_flat_units_permissions` | string array | yes | Unit IDs from active client–unit relations |
| `client` | object or null | **new** | `client_tbl` row where `cli_id = username`; `null` if no matching row |

#### 錯誤回應

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

- on successful login, the server updates `CustomUser.last_login` to the current time
- `building` and `staff_building_permissions` currently return the same data
- the endpoint does not create a session, JWT, or API token
- **backward compatible**: existing keys under `msg` are not renamed or removed; clients that ignore unknown keys continue to work
- `client` is the primary profile for the login account (`cli_id == username`), not the full M2M `user.clients` list

### 1.2 直接註冊 iSmart 帳戶

- 端點： `POST /api/v1/integration/auth/register/`
- 舊版路徑： none
- 用途： directly creates a new `accounts.CustomUser` and linked `ClientTbl` profile with no approval flow
- 存取控制： currently no IP whitelist and no user authentication

#### 請求

```json
{
  "phone": "+85291234567",
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

HK national-only phone is still accepted and stored as E.164:

```json
{
  "phone": "91234567",
  "email": "user@example.com",
  "eng_name": "CHAN TAI MAN"
}
```

→ stored login phone `+85291234567`, login email `user@example.com` (lowercase).

Mainland China example:

```json
{
  "phone": "+8613901234567",
  "email": "user@example.com",
  "eng_name": "ZHANG SAN"
}
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `phone` | string | yes | **登入電話**。必須有效，並以 **E.164** 儲存（參見[登入電郵與電話格式](#登入電郵與電話格式適用於註冊check-contact會員資料更新)）。等同舊欄位 `memberphone`。 |
| `email` | string | yes | **Login email**. Valid address; stored **lowercase**; unique case-insensitively. |
| `eng_name` | string | yes | English display name |
| `chi_name` | string | no | Chinese name |
| `legal_entity` | string | no | `NA` for individual, `LE` for legal entity. Default: `NA` |
| `id_card` | string | no | ID card / document number |
| `remark` | string | no | Free-text note saved to `ClientTbl.cli_remark` |
| `gender` | string | no | `M`, `F`, or empty |
| `is_receive_email` | boolean | no | Whether to enable `UserSettings.blg_notice_email`. Default: `true` |
| `password` | string | no | If omitted, the backend generates a random password |

#### 仍接受的舊版欄位別名

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

#### 成功回應

```json
{
  "status": "success",
  "data": {
    "user_id": 123,
    "username": "200123",
    "password": "generated-password",
    "phone": "+85291234567",
    "email": "user@example.com",
    "client_id": "200123"
  }
}
```

`data.phone` / `data.email` are the **normalized stored** login values (E.164 phone, lowercase email).

#### 重複帳戶回應

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

#### 格式／驗證錯誤

```json
{
  "status": "error",
  "message": "電話號碼格式無效或未能識別國家碼: …"
}
```

```json
{
  "status": "error",
  "message": "電郵格式無效: …"
}
```

HTTP `400` when phone/email fail format validation.

#### Notes

- if phone or email already belongs to an existing `CustomUser` (after normalization / case-insensitive email match), the API returns HTTP `409`
- if a matching `ClientTbl` already exists by phone candidates, email, or `id_card`, its `cli_id` is reused as the new username
- on create, primary `ClientTbl.cli_tel` / `cli_email` are seeded from the same login phone/email; later updates can diverge (see client profile API)
- no `OwnerReg` row is created by this endpoint
- before register, callers may use `GET/POST /api/v1/integration/auth/check-contact/` to probe uniqueness without creating an account

### 1.3 檢查電郵／電話可用性

- 端點： `GET` or `POST` `/api/v1/integration/auth/check-contact/`
- 舊版路徑： none
- 用途： check whether an email and/or phone is already used as a unique login credential on `CustomUser`
- 存取控制： currently no IP whitelist and no user authentication

Use this before registration, or before changing an existing user's login email/phone. Uniqueness is enforced on `CustomUser.email` (case-insensitive) and `CustomUser.phone` (E.164 + legacy format candidates). **Phone must be a valid number** (same rules as register); invalid phone → HTTP `400`.

#### GET 查詢參數

```text
?email=user@example.com&phone=%2B85291234567&exclude_user_id=123
```

#### POST Request

```json
{
  "email": "user@example.com",
  "phone": "+85291234567",
  "exclude_user_id": 123
}
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `email` | string | conditional | Login email to check (any case). Provide `email` and/or `phone`. |
| `phone` | string | conditional | Login phone to check. Prefer **E.164** (`+852…` / `+86…`); HK 8-digit national also accepted. Provide `email` and/or `phone`. |
| `exclude_user_id` | integer | no | When editing an existing account, pass that user's id so their current email/phone are not reported as taken |

#### 成功回應

```json
{
  "status": "success",
  "data": {
    "email": "user@example.com",
    "email_available": false,
    "email_user_id": 456,
    "phone": "+85291234567",
    "phone_normalized": "+85291234567",
    "phone_available": true,
    "phone_user_id": null,
    "user_id": 456,
    "exclude_user_id": 123,
    "available": false
  }
}
```

| Field | Type | Description |
|---|---|---|
| `email` | string or null | Input email (stripped), or `null` if not checked |
| `email_available` | boolean or null | `true` if free; `null` if email was not checked |
| `email_user_id` | integer or null | Matching user's id when the email is taken; otherwise `null` |
| `phone` | string or null | Raw phone input, or `null` if not checked |
| `phone_normalized` | string or null | **E.164** form used for matching when phone was checked and non-empty |
| `phone_available` | boolean or null | `true` if free; `null` if phone was not checked |
| `phone_user_id` | integer or null | Matching user's id when the phone is taken; otherwise `null` |
| `user_id` | integer or null | Matching user id when all matched contacts identify one user. `null` when no user matches or the email and phone belong to different users |
| `exclude_user_id` | integer or null | Echo of exclude id |
| `available` | boolean | `true` only when every checked field is free |

#### 錯誤回應

```json
{
  "status": "error",
  "message": "Provide email and/or phone to check"
}
```

```json
{
  "status": "error",
  "message": "電話號碼格式無效或未能識別國家碼: …"
}
```

#### Notes

- empty email/phone inputs are treated as available (`true`) when that key is present
- invalid phone format returns HTTP `400` (same phone normalization as registration)
- this checks **login credentials** on `CustomUser` only; it does not block matching free-text values on unrelated `ClientTbl` rows

### 1.4 取得／更新會員資料

- 端點： `GET` / `POST` / `PATCH` / `PUT` `/api/v1/integration/auth/client/`
- 舊版路徑： none
- 用途： read or update the primary member profile used by login (`client_tbl` where `cli_id == username`)
- 存取控制： currently no IP whitelist and no user authentication; caller supplies `user_id`

The `client` object shape matches the `msg.client` payload from login (`_serialize_client_tbl_for_login`).

#### GET 查詢參數

```text
?user_id=123
```

#### GET 成功回應

```json
{
  "status": "success",
  "data": {
    "user_id": 123,
    "username": "200123",
    "email": "demo@example.com",
    "phone": "+85291234567",
    "is_active": true,
    "client": {
      "cli_id": "200123",
      "cli_legalentity": "NA",
      "cli_type": null,
      "cli_contact_person": "",
      "cli_urgent_contact_person": "",
      "cli_name": "CHAN TAI MAN",
      "cli_chi_name": "陳大文",
      "cli_acname": "CHAN TAI MAN",
      "cli_join_dt": "2020-01-15",
      "cli_id_card": "A1234567",
      "cli_tel": "+85291234567",
      "cli_tel2": "",
      "cli_fax": null,
      "cli_addr": "",
      "cli_chiadd": "",
      "cli_email": "demo@example.com",
      "cli_birthday": "1990-01-01",
      "cli_sex": "M",
      "cli_nationality": null,
      "cli_bank": null,
      "cli_bank_acc": null,
      "cli_remark": null
    }
  }
}
```

#### 更新請求 (`POST` / `PATCH` / `PUT`)

變更**登入**憑證（格式參見[登入電郵與電話格式](#登入電郵與電話格式適用於註冊check-contact會員資料更新)）：

```json
{
  "user_id": 123,
  "email": "new@example.com",
  "phone": "+85291234568"
}
```

Change **contact** profile only (login email/phone unchanged):

```json
{
  "user_id": 123,
  "cli_email": "billing@example.com",
  "cli_tel": "+85291234568",
  "cli_chi_name": "陳大文",
  "cli_addr": "1 Example Road",
  "cli_chiadd": "示例路1號",
  "cli_birthday": "1990-01-01",
  "cli_sex": "M",
  "cli_remark": "updated from external app"
}
```

Set login phone and a **different** contact phone in one request:

```json
{
  "user_id": 123,
  "phone": "+8613901234567",
  "cli_tel": "+85291234567"
}
```

#### 可更新欄位

| Field | Type | Description |
|---|---|---|
| `user_id` | integer | Required. Target `CustomUser.id` |
| `email` | string or null | Updates unique **login** `CustomUser.email` (lowercase, case-insensitive unique). **Does not** overwrite `cli_email`. Empty/null clears login email if allowed by validation path. |
| `phone` | string | Updates unique **login** `CustomUser.phone` (**must be valid**; stored E.164). Also sets `cli_tel` to the same value **unless** `cli_tel` is present in the same payload. |
| `cli_legalentity` | string | e.g. `NA`, `LE` |
| `cli_type` | string | Client type |
| `cli_contact_person` | string | Contact person |
| `cli_urgent_contact_person` | string | Emergency contact |
| `cli_name` | string | English name |
| `cli_chi_name` | string | Chinese name |
| `cli_acname` | string | Account / payee name |
| `cli_id_card` | string | ID document number |
| `cli_tel` | string | **Contact** phone only (billing/contact). Prefer E.164. Does **not** change login phone when sent alone. |
| `cli_tel2` | string | Emergency contact phone (free-form) |
| `cli_fax` | string | Fax |
| `cli_addr` | string | English mailing address |
| `cli_chiadd` | string | Chinese mailing address |
| `cli_email` | string | **Contact / billing** email only. Does **not** change login `CustomUser.email`. Prefer a normal lowercase email. |
| `cli_birthday` | string | `YYYY-MM-DD` |
| `cli_sex` | string | `M`, `F`, or `U` |
| `cli_nationality` | string | Nationality |
| `cli_bank` | string | Bank name |
| `cli_bank_acc` | string | Bank account number |
| `cli_remark` | string | Remark |

Not updatable: `cli_id`, `cli_join_dt`.

There is no `ClientTbl` column for an emergency-contact email. External clients must not send or display a writable `emergency_contact_email` field until iSmart has a separately approved data model for it.

#### 更新成功回應

Same shape as GET success (`status` + `data` with refreshed profile).  
Top-level `data.phone` / `data.email` are login credentials (normalized).  
`data.client.cli_tel` / `data.client.cli_email` are contact fields.

### 1.5 修改密碼

- 端點： `POST /api/v1/integration/auth/password/`
- 用途：修改一名會員的 iSmart 登入密碼
- 存取控制：受 `EXTERNAL_APP_ALLOWED_IPS` 保護，只供受信任的 AJO server-to-server 呼叫

請求：

```json
{
  "user_id": 123,
  "current_password": "CurrentPass123!",
  "new_password": "NewPass123!",
  "confirm_password": "NewPass123!"
}
```

目前密碼必須匹配。新密碼會按 Django 已配置的密碼驗證器檢查。API 永不返回密碼值。

成功回應：

```json
{
  "status": "success",
  "data": {
    "user_id": 123,
    "password_updated": true
  }
}
```

AJO 代理此操作時，只能在此端點成功後更新本地加密的 iSmart 關聯憑證。不支援由瀏覽器直接呼叫此端點。

### 1.6 大廈通告電郵設定

- 端點：`GET` / `PATCH` `/api/v1/integration/auth/notification-settings/`
- 用途：讀取或更新 `UserSettings.blg_notice_email`（接收大廈通告電郵提示）
- 存取控制：受 `EXTERNAL_APP_ALLOWED_IPS` 保護，只供受信任的 AJO server-to-server 呼叫

讀取請求：

```text
GET /api/v1/integration/auth/notification-settings/?user_id=123
```

更新請求：

```json
{
  "user_id": 123,
  "receive_building_notice_email": true
}
```

`receive_building_notice_email` 接受 JSON boolean、`0`／`1` 或標準 boolean 字串。回應會使用相同欄位名返回已儲存的 boolean 值。

#### 錯誤回應

```json
{
  "status": "error",
  "message": "user_id is required"
}
```

```json
{
  "status": "error",
  "message": "email is already in use"
}
```

```json
{
  "status": "error",
  "message": "phone is already in use"
}
```

```json
{
  "status": "error",
  "message": "電話號碼格式無效或未能識別國家碼: …"
}
```

```json
{
  "status": "error",
  "message": "Unknown or non-editable fields: cli_id"
}
```

#### Notes

- HTTP `409` is returned when login `email` / `phone` conflict with another user
- HTTP `400` is returned when login `phone` fails E.164 / validity checks
- before changing login email/phone, call check-contact with `exclude_user_id=<this user_id>`
- **login email ≠ contact email**: `email` never overwrites `cli_email`; set `cli_email` explicitly for billing/contact
- **login phone** updates `cli_tel` by default; send both `phone` and `cli_tel` when they must differ
- `cli_email` / `cli_tel` alone do **not** change unique login credentials
- if the user has no primary client row, GET returns `"client": null`; update returns an error

## 2. 大廈資料 API

### 2.1 大廈管理費表

- 端點： `GET /api/v1/integration/buildings/receivables/management-fees/`
- 舊版路徑： `POST /api/v1/building-mf-table/`
- 用途： returns management-fee receivable data generated from `blg_mf_list(...)`
- 存取控制： no IP whitelist and no user authentication

#### 整合查詢參數

```text
?building_id=0348200&data_structure=table
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | Building ID |
| `data_structure` | string | no | Response layout. Common values: `table`, `list`, `block_dict`. Default: `table` |

#### 成功回應

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

#### 成功回應

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

- response shape changes with `data_structure`
- values may be amounts or status labels such as `已付`, `審核中`, or `錯誤`
- the backend calls the receivable generator with `group_estate=True`

### 2.2 大廈其他費用列表

- 端點： `GET /api/v1/integration/buildings/receivables/other-fees/`
- 舊版路徑： `POST /api/v1/building-of-list/`
- 用途： returns non-management-fee receivable data generated from `blg_mf_list(...)`
- 存取控制： no IP whitelist and no user authentication

#### 整合查詢參數

```text
?building_id=0348200&data_structure=list
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | Building ID |
| `data_structure` | string | no | Common values: `list` or `block_dict`. Default: `list` |

#### 成功回應

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

#### 成功回應

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

- response shape changes with `data_structure`
- empty result is returned as `[]`

### 2.3 大廈通告

- 端點： `GET /api/v1/integration/buildings/notices/`
- 舊版路徑： `POST /api/v1/building-notices/`
- 用途： returns active notices for one building
- 存取控制： no IP whitelist and no user authentication

#### 整合查詢參數

```text
?building_id=0348200
```

#### 成功回應

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

- response is a raw array, not an object wrapper
- only notices where `mess_date <= today < mess_down` are returned

## 3. 外部大廈 API

These endpoints are intended for trusted server-to-server or app-backend integrations.

### 3.0 會員角色綁定

This section covers member-to-unit binding flows implemented in this session.

- owner-role binding (`登記業主`) is submitted through `OwnerReg` and still requires building-management approval
- authorized sub user binding (`授權用戶`) does not use approval when initiated by an already-approved owner
- the actual active role relationship is still stored in `i_flatcli_rel_tbl`
- a separate history model now records authorized sub user grant / revoke operations

### 3.0.1 業主角色綁定申請

- 端點： `POST /api/v1/integration/buildings/building-flat-owner-binding-requests/`
- 舊版路徑： none
- 用途： submits an `OwnerReg` request for binding a user to one or more units, mainly for the owner-role approval flow
- 存取控制： currently no IP whitelist and no user authentication

#### 請求範例： Existing User Requests Owner Binding

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

#### 請求範例： New Applicant Without Existing User

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

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | Building ID |
| `ownedflat` | string array or comma-separated string | yes | One or more `unit_id` values that must belong to the supplied building |
| `user` | integer | conditional | Existing `accounts.CustomUser.id`. If provided, missing applicant info is auto-filled from the linked user/client profile |
| `cli_role` | string | no | Defaults to `業主` |
| `ownernote` | string | no | Free-text note |
| `is_receive_email` | boolean | no | Whether to enable notice email preference on approval. Default: `false` |
| `reg_tel` | string | conditional | Required when `user` is not provided |
| `reg_email` | string | no | Applicant email |
| `cli_name` | string | conditional | Required when `user` is not provided |
| `cli_id_card` | string | no | Applicant identity document number |
| `cli_tel` | string | no | Applicant contact phone |

#### 成功回應

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

- this endpoint creates an `OwnerReg` row only; no role relationship is granted immediately
- staff approval later runs through the existing `confirm_memreg` flow
- if `OwnerReg.user` is present, approval reuses the existing account instead of creating a new account

### 3.0.2 列出授權子用戶

- 端點： `GET /api/v1/integration/buildings/subaccounts/`
- 舊版路徑： none
- 用途： lists active `授權用戶` relationships for unit(s) where the caller is already the approved `登記業主`
- 存取控制： caller must be an active `登記業主` of the requested unit; if `unit_id` is omitted, all owner-controlled units are included

#### 整合查詢參數

```text
?user_id=123&unit_id=0348200001
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | `accounts.CustomUser.id` of the owner |
| `unit_id` | string | no | Optional unit filter. If omitted, returns all active subaccounts for all units owned by the user |

#### 成功回應

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

- this endpoint returns only active `i_flatcli_rel_tbl` rows where `cli_role='授權用戶'`
- `history_count` counts rows in the dedicated authorized-sub-user history model for the relation
- if `unit_id` is provided and the caller is not the approved owner, the API returns HTTP `403`

### 3.0.3 授權子用戶

- 端點： `POST /api/v1/integration/buildings/subaccounts/grant/`
- 舊版路徑： none
- 用途： lets an already-approved owner grant `授權用戶` to another existing `CustomUser` for one unit, with no staff approval
- 存取控制： caller must already be an active `登記業主` of the requested unit

#### 請求

```json
{
  "user_id": 123,
  "unit_id": "0348200001",
  "target_user_id": 456,
  "remark": "family member"
}
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | Owner `accounts.CustomUser.id` |
| `unit_id` | string | yes | Target unit ID |
| `target_user_id` | integer | yes | Existing `accounts.CustomUser.id` to become `授權用戶` |
| `remark` | string | no | Free-text audit note stored on both relation and history |

#### 成功回應

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

#### 驗證規則

- `user_id` must currently be an active `登記業主` of the supplied `unit_id`
- `target_user_id` must already exist and have a linked `ClientTbl`
- `target_user_id` cannot equal `user_id`
- each unit can have at most 4 active `授權用戶`
- if the target already has an active `授權用戶` relation for the unit, the API returns HTTP `409`

#### Notes

- if the same target user had an old inactive `授權用戶` row, the current code reactivates that row instead of creating a second one
- each successful grant writes one row into `AuthorizedSubUserRequestHistory`

### 3.0.4 撤銷授權子用戶

- 端點： `POST /api/v1/integration/buildings/subaccounts/revoke/`
- 舊版路徑： none
- 用途： lets an owner deactivate an active `授權用戶` relationship for one unit
- 存取控制： caller must already be an active `登記業主` of the requested unit

#### 請求

```json
{
  "user_id": 123,
  "unit_id": "0348200001",
  "target_user_id": 456,
  "remark": "moved out"
}
```

#### 成功回應

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

- revoke is implemented as `is_active=False` on the existing `i_flatcli_rel_tbl` row
- each successful revoke writes one row into `AuthorizedSubUserRequestHistory`
- revoking a user who is not currently an active `授權用戶` of the unit returns HTTP `404`

### 3.1 大廈資料與文件

- 端點： `GET /api/v1/integration/buildings/info/`
- 舊版路徑： `POST /api/v1/external/building-info/`
- 用途： returns building profile, building info fields, and building file metadata
- 存取控制： `EXTERNAL_APP_ALLOWED_IPS`

#### 整合查詢參數

```text
?building_id=0348200
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | Target building ID |

#### 成功回應

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

#### 返回的文件分組

| Response Field | Source `btype` |
|---|---|
| `documents.forms` | `form` |
| `documents.building_info_files` | `blginfo` |
| `documents.floorplans` | `floorplan` |
| `documents.audit_reports` | queried as `audition` in current code |
| `documents.financial_reports` | `mfinreport` |

#### Notes

- arrays are returned even when no files exist
- `building_info` is returned with `null` or empty-string values when no `BuildingInfo` row exists
- current code queries audit reports using `btype='audition'`; if your data uses `auditreport`, those files will not appear in this endpoint until code is aligned

### 3.2 提交大廈服務個案（原大廈意見）

- 端點： `POST /api/v1/integration/buildings/comments/`
- 舊版路徑： `POST /api/v1/external/blg-cs/submit/`
- 用途： creates a `BuildingServiceCase` (維修報修 / 意見反映) with the first thread message, and sends staff notification email
- 存取控制： `EXTERNAL_APP_ALLOWED_IPS` plus user-to-building authorization

#### 請求（建議使用的新分類）

```json
{
  "user_id": 123,
  "building_id": "0348200",
  "request_type": "repair",
  "category": "drainage",
  "subcategory": "blocked_drain",
  "subject": "走廊冷氣滴水",
  "content": "Air-conditioner water leakage at the corridor outside 18/F.",
  "location_text": "18樓走廊",
  "unit_id": "03482001801",
  "contact_name": "Chan Tai Man",
  "contact_phone": "+85291234567"
}
```

#### 請求（仍接受的舊版兼容格式）

```json
{
  "user_id": 123,
  "building_id": "0348200",
  "comment_type": "其他事宜",
  "comment": "Air-conditioner water leakage at the corridor outside 18/F."
}
```

When only `comment_type` / `comment` are sent, the API maps `comment_type` to `(request_type, category, subcategory)` and uses `comment` as `content`.

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | `accounts.CustomUser.id` |
| `building_id` | string | yes | Target building ID |
| `content` | string | yes* | Case body (*or legacy `comment`) |
| `comment` | string | yes* | Legacy alias for `content` |
| `request_type` | string | recommended | `repair` or `feedback` |
| `category` | string | with type | Taxonomy category code |
| `subcategory` | string | with type | Taxonomy subcategory code |
| `subject` | string | no | Subject line (defaults from content / comment_type) |
| `location_text` | string | no | Free-text location |
| `unit_id` | string | no | `IFlatTbl.unit_id` in this building |
| `contact_name` | string | no | Defaults to user display name |
| `contact_phone` | string | no | Defaults to user phone |
| `comment_type` | string | legacy | Old free-text type; mapped if taxonomy fields omitted |

If taxonomy fields are omitted and `comment_type` is missing, defaults to `feedback` / `other_feedback` / `other`.

#### 舊版 `comment_type` 值（由伺服器映射）

- `門卡報失`, `冷氣滴水`, `嘈音滋擾`, `樓梯雜物`, `水質問題`, `渠務問題`, `保安事宜`, `清潔衛生`, `增加服務`, `電力問題`, `其他事宜`

#### 成功回應

```json
{
  "status": "success",
  "message": "Case submitted successfully",
  "data": {
    "case_id": "ibscase_…",
    "case_no": "CS20260728abcdefgh",
    "building_id": "0348200",
    "request_type": "repair",
    "category": "drainage",
    "subcategory": "blocked_drain",
    "status": "submitted",
    "status_label": "待跟進",
    "created_at": "2026-07-28T10:20:30.000000+08:00"
  }
}
```

#### Notes

- the supplied `user_id` must be authorized for the supplied `building_id`
- success status code is `201`
- creates `BuildingServiceCase` + initial `BuildingServiceCaseMessage` (not the removed `BlgComment` model)
- email notification to management is sent inside the request flow (`fail_silently`)
- resident UI: `/blg_cs/`; staff UI: `/building_service_case_list/` (permission `building_service_case_list`)

### 3.3 列出大廈服務個案

- 端點： `GET` or `POST` `/api/v1/integration/buildings/service-cases/`
- 用途： list service cases for a building, optional status / type filters
- 存取控制： `EXTERNAL_APP_ALLOWED_IPS` plus user-to-building authorization

#### 請求（GET query 或 POST JSON）

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | `accounts.CustomUser.id` |
| `building_id` | string | yes | Target building ID |
| `status` | string | no | One of: `submitted`, `processing`, `pending_information`, `completed`, `rejected`, `cancelled` |
| `request_type` | string | no | `repair` or `feedback` |
| `scope` | string | no | `mine` (default) = only cases created by this user; `building` = all cases (staff only) |
| `limit` | integer | no | Max rows to return (default 100, max 500) |

```http
GET /api/v1/integration/buildings/service-cases/?user_id=123&building_id=0348200&status=submitted
```

```json
{
  "user_id": 123,
  "building_id": "0348200",
  "status": "processing",
  "scope": "mine",
  "limit": 50
}
```

#### 成功回應

```json
{
  "status": "success",
  "data": {
    "building_id": "0348200",
    "scope": "mine",
    "filters": { "status": "submitted", "request_type": null },
    "total": 2,
    "returned": 2,
    "limit": 100,
    "status_choices": [
      { "code": "submitted", "label": "待跟進" },
      { "code": "processing", "label": "處理中" }
    ],
    "cases": [
      {
        "case_id": "ibscase_…",
        "case_no": "CS20260728abcdefgh",
        "building_id": "0348200",
        "unit_id": null,
        "request_type": "repair",
        "request_type_label": "維修報修",
        "category": "waterworks",
        "category_label": "水務",
        "subcategory": "leakage",
        "subcategory_label": "漏水",
        "subject": "走廊漏水",
        "status": "submitted",
        "status_label": "待跟進",
        "location_text": "18樓",
        "contact_name": "Chan",
        "contact_phone": "+85291234567",
        "created_by_id": 123,
        "assigned_staff_id": null,
        "closed_at": null,
        "created_at": "2026-07-28T10:20:30+08:00",
        "updated_at": "2026-07-28T10:20:30+08:00"
      }
    ]
  }
}
```

#### Notes

- non-staff users always get `scope=mine` behavior (their own cases only)
- `scope=building` returns 403 for non-staff
- `total` is the full filtered count; `returned` is after `limit`

### 3.4 大廈服務個案詳情

- 端點： `GET` or `POST` `/api/v1/integration/buildings/service-cases/<case_id>/`
- 用途： case header + full message thread (attachments included)
- 存取控制： `EXTERNAL_APP_ALLOWED_IPS`; caller must be the case creator **or** staff with access to the case’s building

#### 請求

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | Viewer user id |
| `case_id` | string | yes* | Case PK (`ibscase_…`); may also be in the URL path |

```http
GET /api/v1/integration/buildings/service-cases/ibscase_xxx/?user_id=123
```

```json
{
  "user_id": 123,
  "case_id": "ibscase_xxx"
}
```

#### 成功回應

```json
{
  "status": "success",
  "data": {
    "case_id": "ibscase_…",
    "case_no": "CS20260728abcdefgh",
    "building_id": "0348200",
    "subject": "走廊漏水",
    "content": "Full original body…",
    "status": "processing",
    "status_label": "處理中",
    "messages": [
      {
        "message_id": "ibscmsg_…",
        "author_id": 123,
        "author_role": "resident",
        "author_username": "member1",
        "body": "…",
        "is_internal": false,
        "created_at": "2026-07-28T10:20:30+08:00",
        "attachments": [
          { "id": 1, "file_url": "https://…", "message_id": "ibscmsg_…", "uploaded_by_id": 123, "created_at": "…" }
        ]
      }
    ],
    "attachments": []
  }
}
```

#### Notes

- residents never receive messages with `is_internal=true`
- staff with building access receive the full thread including internal notes
- 403 if the user is neither the creator nor authorized staff for that building

## 4. 外部門禁 API

### 4.1 大廈門禁摘要

- 端點： `POST /api/v1/integration/access/buildings/`
- 舊版路徑： `POST /api/v1/external/building-access/`
- 用途： returns public door list, camera reference, password info, QR info, and recent access records for a user in a building
- 存取控制： `EXTERNAL_APP_ALLOWED_IPS` plus user-to-building authorization

#### 請求

```json
{
  "user_id": 123,
  "building_id": "0348200"
}
```

#### 成功回應

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

- `has_permission` reflects `DoorPermission`
- `password`, `qrcode`, and `camera` can be `null`
- two legacy building remaps are applied in current code:
  - `0348300` is remapped to `0348200`
  - `0324900` is remapped to `0325000`
- `requested_building_id` and `access_building_id` can therefore differ

### 4.2 遠端開門

- 端點： `POST /api/v1/integration/access/open-door/`
- 舊版路徑： `POST /api/v1/external/building-access/open-door/`
- 用途： triggers the same remote-open flow used by the smartliving web pages
- 存取控制： `EXTERNAL_APP_ALLOWED_IPS` plus user-to-building authorization

#### 請求

```json
{
  "user_id": 1,
  "building_id": "0999900",
  "door_id": 1
}
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | `accounts.CustomUser.id` |
| `door_id` | integer | yes | `smartliving.Door.id` |
| `building_id` | string | no | Optional extra validation that the door belongs to the building |

#### 成功回應

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

#### 上游失敗回應

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

- failure to reach the door controller returns HTTP `502`
- current open-door permission logic is not identical to the listing endpoint:
  - listing checks `DoorPermission`
  - open-door checks whether the user has any active flat-client relationship in the building

### 4.3 生成門禁 QR Payload

- 端點： `POST /api/v1/integration/access/qrcode/`
- 舊版路徑： `POST /api/v1/external/building-access/qrcode/`
- 用途： generates a QR payload string that the client can render as a QR image
- 存取控制： `EXTERNAL_APP_ALLOWED_IPS` plus user ownership of the QR record

#### 請求

```json
{
  "user_id": 123,
  "qrcode_record_id": 77,
  "term": "dynamic"
}
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | `accounts.CustomUser.id` |
| `qrcode_record_id` | integer | yes | `DoorPasswordRecord.id` where `passwordorqrcode=True` |
| `term` | string | no | `dynamic` or `static`. Default: `dynamic` |

#### 成功回應

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

#### QR 詞彙規則

- `dynamic`: expires at current time plus 15 minutes
- `static`: expires at current time plus 360 days, capped by the record `end_time`

#### Notes

- the API returns payload text only; the client must render the QR image
- current code does not separately reject already-expired QR records before generation

## 5. 支付 API

### 5.1 單位未繳賬單列表

- 端點： `GET /api/v1/integration/payments/unpaid-invoices/`
- 舊版路徑： `POST /api/v1/building-flat-unpaid-invoice-list`
- 用途： returns currently unpaid invoice rows for one unit, net of payments already pending in cashier or validation
- 存取控制： no IP whitelist and no user authentication

#### 整合查詢參數

```text
?unit_id=09999000111
```

#### 成功回應

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

- response is a raw array
- `net_amount` is returned in dollars, not cents
- pending payments with status `in_cashier` or `pending_validation` are deducted from the outstanding amount

### 5.2 POS Payment to iSmart

- 端點： `POST /api/v1/integration/payments/pos/`
- 舊版路徑： `POST /api/v1/pos-payment-to-ismart`
- 用途： creates a `Payment` and `Paymentdetails`, and for selected payment methods also creates `BalCalcTbl` and `PrepaidBal` rows immediately
- 存取控制： no active IP whitelist in current code
- content type: `application/json` (pictures are **not** multipart form uploads; they are base64 strings inside the JSON body)

#### 請求 Example (without pictures)

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

#### 請求 Example (with one payment receipt picture)

Pictures are uploaded as **parallel arrays**: each index in `PIC_FILENAME` matches the same index in `PIC_DATA`.

`PIC_DATA` accepts either:

1. raw base64 (recommended), or
2. a data URL such as `data:image/jpeg;base64,<payload>`

```json
{
  "BLG_ID": "0348200",
  "PAY_METHOD": "POS_BANK",
  "FINAL_AMOUNT": 150000,
  "ENTRY_DATETIME": "2026-06-14 10:30:00",
  "TRAN_DATETIME": "2026-06-14 10:30:00",
  "TRAN_REF_NO": "BANK-REF-7788",
  "USER_ID": 1,
  "FAKE_TRANSACTION": false,
  "COMMENT": "bank transfer receipt attached",
  "PIC_FILENAME": ["bank_receipt.jpg"],
  "PIC_DATA": [
    "/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAAKAAoDASIAAhEBAxEB/8QAHwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQRBRIhMUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMiMoEIFEKRobHBCSMzUvAVYnLRChYkNOEl8RcYGRomJygpKjU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6goOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4uPk5ebn6Onq8vP09fb3+Pn6/9oADAMBAAIRAxEAPwDi6KKK+ZP3E//Z"
  ],
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

#### 請求 Example (multiple pictures + data URL form)

```json
{
  "BLG_ID": "0348200",
  "PAY_METHOD": "POS_CHEQUE",
  "FINAL_AMOUNT": 88000,
  "ENTRY_DATETIME": "2026-06-14 11:05:00",
  "TRAN_DATETIME": "2026-06-14 11:05:00",
  "TRAN_REF_NO": "CHQ-00912",
  "USER_ID": 1,
  "COMMENT": "cheque front and back",
  "PIC_FILENAME": [
    "cheque_front.jpg",
    "cheque_back.jpg"
  ],
  "PIC_DATA": [
    "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAAKAAoDASIAAhEBAxEB/8QAHwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQRBRIhMUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMiMoEIFEKRobHBCSMzUvAVYnLRChYkNOEl8RcYGRomJygpKjU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6goOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4uPk5ebn6Onq8vP09fb3+Pn6/9oADAMBAAIRAxEAPwDi6KKK+ZP3E//Z",
    "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
  ],
  "BILL_OBJS": [
    {
      "flat_code": "0348200001",
      "invoice_no": "",
      "item_id": "管理費",
      "trs_to": "預付",
      "net_amount": 88000,
      "transfer_fee": 0
    }
  ]
}
```

#### 圖片上載說明

| Rule | Detail |
|---|---|
| Transport | JSON body only (`Content-Type: application/json`). Do **not** use `multipart/form-data`. |
| Field pairing | `PIC_FILENAME[i]` is the display/source filename for `PIC_DATA[i]`. Lengths must match. |
| Accepted shapes | Prefer **arrays**. A single string for each field is also accepted and treated as a 1-item list. |
| Filename | Send a simple name such as `receipt.jpg`. Full device paths (e.g. `/storage/emulated/0/...`) are accepted; only the basename is stored. |
| Encoding | Standard base64. Data URLs (`data:image/jpeg;base64,...`) are supported. Whitespace/newlines inside base64 are OK. |
| Storage | Files are saved to object storage under `payment/<uuid>_<filename>` and linked as `PaymentPicture` rows. |
| Size limit | Whole JSON request body max **5 MB** (`DATA_UPLOAD_MAX_MEMORY_SIZE`) — this is a **total payload** limit, not per image. Base64 expands binary size by ~33% (1 MB file ≈ 1.33 MB in JSON). |
| When applied | Picture fields are processed for POS / web POS methods (`POS_CASH`, `POS_CHEQUE`, `POS_BANK`, `POS_ALIWE`, `WEBPOS_*`). Empty arrays / omitted fields mean no pictures. |

> **Developer warning — check size and compress before send**
>
> - Do **not** upload full-resolution camera originals. Receipt / cheque photos only need to be readable.
> - **Recommended:** resize longest side to ~1280–1920 px and JPEG quality ~60–80 so **each image is ≤ 500 KB–1 MB decoded** before base64.
> - **Hard ceiling:** entire JSON body must be ≤ **5 MB**. With base64 overhead, total decoded image bytes should stay roughly under **~3.5 MB** for all pictures combined.
> - Measure payload size client-side (`JSON.stringify(payload).length` / UTF-8 byte length) before `POST`. If it is close to 5 MB, compress further or send fewer pictures.
> - Oversized requests return `400` with message `Request body too large...` — fix compression rather than raising the server limit.

#### 客戶端編碼範例（偽代碼）

```text
# Python
import base64, json, requests

with open("bank_receipt.jpg", "rb") as f:
    b64 = base64.b64encode(f.read()).decode("ascii")

payload = {
    "BLG_ID": "0348200",
    "PAY_METHOD": "POS_BANK",
    "FINAL_AMOUNT": 150000,
    "ENTRY_DATETIME": "2026-06-14 10:30:00",
    "TRAN_DATETIME": "2026-06-14 10:30:00",
    "TRAN_REF_NO": "BANK-REF-7788",
    "PIC_FILENAME": ["bank_receipt.jpg"],
    "PIC_DATA": [b64],   # or f"data:image/jpeg;base64,{b64}"
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

requests.post(
    "https://<host>/api/v1/pos-payment-to-ismart",
    headers={"Content-Type": "application/json"},
    data=json.dumps(payload),
    timeout=60,
)
```

```text
// JavaScript / React Native style
const fileBase64 = await readFileAsBase64(fileUri); // without data: prefix preferred
const payload = {
  BLG_ID: "0348200",
  PAY_METHOD: "POS_BANK",
  FINAL_AMOUNT: 150000,
  ENTRY_DATETIME: "2026-06-14 10:30:00",
  TRAN_DATETIME: "2026-06-14 10:30:00",
  TRAN_REF_NO: "BANK-REF-7788",
  PIC_FILENAME: ["bank_receipt.jpg"],
  PIC_DATA: [fileBase64],
  BILL_OBJS: [/* ... */],
};
await fetch("https://<host>/api/v1/pos-payment-to-ismart", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(payload),
});
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `BLG_ID` | string | yes | Building ID |
| `PAY_METHOD` | string | yes | Supported values listed below |
| `FINAL_AMOUNT` | integer | conditional | Required for POS and web POS methods, in cents |
| `AMOUNT` | integer | conditional | Required for `allinpay`, in cents |
| `ENTRY_DATETIME` | string | conditional | Required for POS and web POS methods, `YYYY-MM-DD HH:mm:ss` |
| `TRAN_DATETIME` | string | conditional | Required for POS and web POS methods, `YYYY-MM-DD HH:mm:ss` |
| `DATE` | string | conditional | Required for `allinpay`, format `YYYYMMDD` |
| `TIME` | string | conditional | Required for `allinpay`, format `HHmmss` |
| `TRAN_REF_NO` | string | no | External reference number |
| `TRANS_TRACE_NO` | string | conditional | Required for `allinpay` |
| `TRANS_TICKET_NO` | string | conditional | Required for `allinpay` |
| `USER_ID` | integer | no | Defaults to `1` |
| `FAKE_TRANSACTION` | boolean | no | If true, status becomes `pending_validation` |
| `COMMENT` | string | no | Free-text comment |
| `PIC_FILENAME` | string or string array | no | Image filename(s). Prefer array. Pair by index with `PIC_DATA`. |
| `PIC_DATA` | string or string array | no | Base64 image payload(s), raw or `data:<mime>;base64,...`. Prefer array. |
| `bank_account_received` | string | no | Target bank account number |
| `PAYMENT_GATEWAY_RESPONSE` | object | no | Raw gateway response to store |
| `BILL_OBJS` | object array | yes | Payment line items |

#### `BILL_OBJS` 項目欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `flat_code` | string | yes | Unit ID |
| `invoice_no` | string | no | Invoice number. Empty or omitted for prepaid |
| `item_id` | string | yes | Fee item name |
| `trs_to` | string | yes | Billing period or `預付` |
| `net_amount` | integer | yes | Amount in cents |
| `transfer_fee` | integer | no | Handling fee in cents. Default: `0` |

#### 支援的 `PAY_METHOD` 值

| Value | Initial Status | Notes |
|---|---|---|
| `POS_CASH` | `in_cashier` | Cash payment |
| `POS_CHEQUE` | `in_cashier` | Cheque payment |
| `POS_BANK` | `pending_validation` | Bank transfer (commonly used with receipt photos) |
| `POS_ALIWE` | `in_cashier` | Alipay / WeChat on POS |
| `WEBPOS_ALIPAY` | `payment_captured` | Immediate accounting writeback |
| `WEBPOS_WECHAT` | `payment_captured` | Immediate accounting writeback |
| `WEBPOS_CARD_UP` | `payment_captured` | Immediate accounting writeback |
| `allinpay` | `init` | Special legacy flow (does not process `PIC_*` fields) |

#### 成功回應

```json
{
  "code": "200",
  "message": "success",
  "receipt_id": "1000042"
}
```

#### 圖片／body 大小相關錯誤回應

```json
{
  "code": "400",
  "message": "Request body too large. Maximum allowed is 5 MB (including base64-encoded PIC_DATA). Compress images or send fewer pictures."
}
```

```json
{
  "code": "400",
  "message": "PIC_FILENAME and PIC_DATA length mismatch (2 filenames vs 1 data items)"
}
```

```json
{
  "code": "400",
  "message": "Invalid base64 in PIC_DATA for file 'receipt.jpg': Incorrect padding"
}
```

#### 驗證規則

- building must exist
- `PAY_METHOD` must be configured as a `BuildingPaymentType` for the building
- for non-prepaid lines, requested payment amount cannot exceed current unpaid amount
- unsupported pay method returns `400`
- `PIC_FILENAME` / `PIC_DATA` length mismatch or invalid base64 returns `400`
- total JSON body over 5 MB returns `400` (compress images client-side; do not send full-resolution camera files)

#### 重要行為

- all request money values are in cents
- created `Payment.txamount` and line amounts are stored internally in dollars
- `WEBPOS_*` methods immediately create payment-side accounting rows
- prepaid lines with `trs_to = "預付"` create `PrepaidBal`
- handling fees create separate detail and accounting rows for `平台手續費`
- when pictures are provided, one `PaymentPicture` row is created per decoded image and stored on S3 under `payment/`

#### 重要注意事項

- the whitelist decorator for this endpoint is currently commented out
- unknown `bank_account_received` is silently ignored rather than rejected
- the `allinpay` datetime parsing path is legacy and should be tested carefully before use
- camera photos should be compressed client-side; base64 makes the JSON body larger than the original file
- do not send multipart file fields; the endpoint only reads `request.body` as JSON
### 5.3 收銀台交易

- 端點： `GET /api/v1/integration/cashier/transactions/`
- 舊版路徑： `POST /api/v1/get-transactions-in-cashier`
- 用途： lists current cashier transactions for one building

#### 整合查詢參數

```text
?building_id=0348200
```

#### 成功回應

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

- only status `in_cashier` is returned
- results are split into `payment_objs_cheque` and `payment_objs_cash`
- endpoint does not return a unified list

### 5.4 按單位查詢交易

- 端點： `GET /api/v1/integration/payments/transactions/by-unit/`
- 舊版路徑： `POST /api/v1/get-transactions-by-flat-unit`
- 用途： lists payment history for one or more units

#### 整合查詢參數

```text
?unit_id_list=0348200001&unit_id_list=0348200002
```

#### 成功回應

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

- `init` and `pending_validation` are excluded
- `validated_by_ismart`, `payment_captured`, and `confirmed` are normalized to `confirmed`
- prepaid consumption may appear as `"{trs_to}(預付)"`

### 5.5 按日期查詢交易

- 端點： `GET /api/v1/integration/payments/transactions/by-date/`
- 舊版路徑： `POST /api/v1/get-transactions-by-date`
- 用途： lists building payments filtered by date range

#### 整合查詢參數

```text
?building_id=0348200&from_date=2026-06-01&to_date=2026-06-30&date_type=input_date&pay_method=all
```

#### 請求欄位

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | Building ID |
| `from_date` | string | yes | Start day |
| `to_date` | string | yes | End day |
| `date_type` | string | yes | `input_date` or `tran_date` |
| `pay_method` | string | no | `all` or `pos` |

#### 成功回應

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

- `date_type` controls whether filtering uses `txdtm` or `tran_date`
- `pay_method = "pos"` keeps only payment types whose name starts with `POS`
- the query applies manual `-8h` and `+8h` adjustments around the date range

## 6. 收銀台與 Bank-In API

### 6.1 更新收銀台交易狀態

- 端點： `POST /api/v1/integration/cashier/bank-in/`
- 舊版路徑： `POST /api/v1/update-transactions-status-in-cashier`
- 用途： groups selected cashier payments into a new bank-in batch and updates them to `pending_validation`

#### 請求

```json
{
  "payment_id_list": [101, 102, 103]
}
```

#### 成功回應

```json
{
  "code": "200",
  "message": "success",
  "bank_in_record_id": "88"
}
```

#### Notes

- all selected payments must currently be `in_cashier`
- the endpoint does not separately verify that all selected rows belong to the same building and pay type before using the first row as batch metadata

### 6.2 Bank-In 紀錄列表

- 端點： `GET /api/v1/integration/cashier/bank-in-records/`
- 舊版路徑： `POST /api/v1/get-bank-in-record-list`
- 用途： returns the latest bank-in batches for one building

#### 整合查詢參數

```text
?building_id=0348200
```

#### 成功回應

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

- returns at most 30 records

### 6.3 Bank-In 紀錄詳情

- 端點： `GET /api/v1/integration/cashier/bank-in-records/details/`
- 舊版路徑： `POST /api/v1/get-bank-in-record-details`
- 用途： returns detailed payment rows inside one bank-in batch

#### 整合查詢參數

```text
?building_id=0348200&record_id=88
```

#### 成功回應

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

- if the batch ID exists but has no matching rows, the API returns an empty list

## 7. Webhook 端點

These are machine endpoints but they do not follow the same contract style as the JSON APIs above.

### 7.1 QFPay 回調

- 端點： `POST /api/v1/integration/webhooks/qfpay/`
- 舊版路徑： `POST /qfpayapi/`
- 用途： stores callback payloads and finalizes successful QFPay payment notifications

#### 預期輸入

- header: `X-QF-SIGN`
- body: JSON payload from QFPay

Fields referenced by current code include:

- `notify_type`
- `status`
- `out_trade_no`

#### 目前行為

- if header `X-QF-SIGN` exists, the endpoint returns plain text `SUCCESS`
- if the body indicates a successful payment callback, the endpoint loads `Payment.id = out_trade_no` and runs the legacy success handler
- if the header is missing, the endpoint still stores the payload and returns plain text `UNSUCCESS`

#### 重要事項

- current code checks only for presence of `X-QF-SIGN`; it does not verify the signature
- response body is plain text, not JSON

### 7.2 測試 Webhook

- 端點： `POST /api/v1/integration/webhooks/test/`
- 舊版路徑： `POST /test_webhook/`
- 用途： stores and emails incoming webhook payloads for debugging

#### 目前行為

- accepts any `POST` body
- stores raw body and content type
- attempts to email a diagnostic copy
- returns plain text `success`
- returns HTTP `405` for non-`POST`

#### 重要事項

- this is a debug utility endpoint, not a production business API
- there is no authentication, signature validation, or schema validation

## 驗證規則摘要

Common validation patterns used across the APIs:

- building IDs are often restricted to alphanumeric values only
- malformed JSON returns `400`
- missing required fields return `400`
- building, user, door, or QR lookup failures return `404`
- authorization checks for external endpoints usually depend on `user_id` supplied in the request body

## 已知契約注意事項

Integrators should be aware of the following inconsistencies in current implementation:

- response envelopes are inconsistent across endpoints
- some endpoints return raw arrays while others return wrapped objects
- payment endpoints mix string status codes and `status` / `message` objects
- several payment queries use raw SQL and manual timezone shifting
- `PosPaymentToIsmart` is currently exposed without an active IP whitelist
- `qfpayapi` does not validate the callback signature
- `ExternalBuildingInfoApi` currently queries audit reports using `btype='audition'`
- door-access list permission logic and remote-open permission logic are not identical

## 方法設計建議

Under the current compatibility strategy:

- the new read-only `/api/v1/integration/...` endpoints now use `GET`
- the old non-integration aliases remain available as `POST`
- write operations, webhooks, and sensitive body-driven access flows remain `POST`

The following integration endpoints already use `GET`:

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

These continue to use `POST`:

- `/api/v1/integration/auth/login/`
- `/api/v1/integration/buildings/comments/`
- `/api/v1/integration/access/open-door/`
- `/api/v1/integration/access/qrcode/`
- `/api/v1/integration/payments/pos/`
- `/api/v1/integration/cashier/bank-in/`
- `/api/v1/integration/webhooks/qfpay/`
- `/api/v1/integration/webhooks/test/`

`/api/v1/integration/access/buildings/` is a mixed case:

- semantically it is read-only, so `GET` is possible
- practically it currently depends on body-supplied `user_id` and returns sensitive door/password-related data
- it is safer to keep `POST` until the auth model is upgraded from body identifiers to token-based authentication

## cURL 請求範例

### 大廈資料

```bash
curl -G "https://<your-domain>/api/v1/integration/buildings/info/" \
  --data-urlencode "building_id=0348200"
```

### 大廈門禁摘要

```bash
curl -X POST "https://<your-domain>/api/v1/integration/access/buildings/" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "building_id": "0348200"
  }'
```

### POS 支付

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

## 操作檢查清單

- populate `EXTERNAL_APP_ALLOWED_IPS` for production integrations
- decide whether `PosPaymentToIsmart` should be protected by `POS_PAYMENT_ALLOWED_IPS`
- align audit report `btype` naming if audit files must appear in `/api/v1/integration/buildings/info/`
- test door open and QR generation against the target hardware environment
- treat `test_webhook/` as non-production and disable or protect it if not needed
