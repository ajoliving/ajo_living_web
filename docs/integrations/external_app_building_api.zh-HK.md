# iSmart 整合 API 參考

本文是 `external_app_building_api(1).md` 的繁體中文對譯，體例對齊 `ismart_integration_api_reference.zh-HK.md`。路徑、欄位名、JSON 鍵與錯誤字串維持原文。

AJO 實際接入仍以 `ajo_ismart` 已暴露路由為真值。現行上游密碼與通告電郵寫入端是 `/api/v1/integration/auth/password/` 與 `/api/v1/integration/auth/notification-settings/`，付款類型是 `/api/v1/integration/payments/fees/`。本文底稿中的 `/auth/change-password/`、`/auth/settings/`、`/payments/types/` 只反映該英文稿，接入前必須再核對上游。

## 目的

本文整理 iSmart 目前對外可用的程式化整合接口，保留舊路徑兼容說明，並按接口列出用途、請求、授權與注意事項。

涵蓋範圍：

- 新整合主路徑 `/api/v1/integration/...`
- 仍可使用的舊 `/api/v1/...`、`/api/v1/external/...` 與舊 webhook 路徑
- POS 與付款流程接口
- webhook 類型的機器端點

本文依據英文底稿所記錄、目前已註冊的 URL 整理。

## 環境

- production: `https://ismart.ajoliving.com`
- clouddev: `https://clouddev.ismart.ajoliving.com`

`clouddev` 的門禁、硬件與部分設備相關能力，可能因環境依賴而不完整。

## 路徑策略

- 新整合主路徑：`/api/v1/integration/...`
- 舊兼容路徑：`/api/v1/...`、`/api/v1/external/...`，以及少量位於 `/api/v1` 之外的舊 webhook 路徑
- 現有客戶端可先保留舊路徑，建議新接入改用整合主路徑

## 安全模型

各接口並未共用同一套認證。接入前必須按單接口處理。

### 受 IP whitelist 保護

以下接口使用 `@ip_whitelist('EXTERNAL_APP_ALLOWED_IPS')`：

- `/api/v1/external/building-info/`
- `/api/v1/external/blg-cs/submit/`
- `/api/v1/integration/buildings/service-cases/`
- `/api/v1/integration/buildings/service-cases/<case_id>/`
- `/api/v1/external/building-access/`
- `/api/v1/external/building-access/open-door/`
- `/api/v1/external/building-access/qrcode/`

注意：

- whitelist 設定在 `stc/settings.py`
- `EXTERNAL_APP_ALLOWED_IPS` 為空時，decorator 實際上放行全部 IP
- 這些接口不使用 session 或 token；身份通常由 JSON body 的 `user_id` 提供

### 目前不受 IP whitelist 保護

- `/api/v1/pos/login`
- `/api/v1/integration/auth/login/`
- `/api/v1/integration/auth/register/`
- `/api/v1/integration/auth/check-contact/`
- `/api/v1/integration/auth/client/`
- `/api/v1/integration/auth/change-password/`
- `/api/v1/integration/auth/settings/`
- `/api/v1/building-mf-table/`
- `/api/v1/building-of-list/`
- `/api/v1/building-notices/`
- `/api/v1/building-flat-unpaid-invoice-list`
- `/api/v1/integration/payments/types/`
- `/api/v1/get-building-payment-types`
- `/api/v1/pos-payment-to-ismart`
- `/api/v1/get-transactions-in-cashier`
- `/api/v1/get-transactions-by-flat-unit`
- `/api/v1/get-transactions-by-date`
- `/api/v1/update-transactions-status-in-cashier`
- `/api/v1/get-bank-in-record-list`
- `/api/v1/get-bank-in-record-details`

注意：

- `PosPaymentToIsmart` 程式中的 `POS_PAYMENT_ALLOWED_IPS` decorator 目前被註解
- `qfpayapi` 只檢查 header `X-QF-SIGN` 是否存在，不驗證簽名
- `test_webhook/` 只供 debug，不可當正式業務接口

## 共同規則

- `application/json` 為主要請求格式
- 新整合唯讀接口多用 `GET` + query string
- 舊路徑多仍接受 `POST` JSON
- 會員身份通常由 `user_id` 提供，外部安全接口再疊加 IP whitelist
- 本文件列出的 `api/v1` 接口均 CSRF-exempt
- 部分付款查詢會手動加減 8 小時以對齊香港時間
- 回應格式不完全統一，必須按單接口契約解析

## 錯誤碼

| HTTP | 說明 |
|---|---|
| `200` | 成功 |
| `201` | 已建立 |
| `400` | 錯誤 JSON、缺欄、無效值 |
| `401` | 憑證無效 |
| `403` | IP 不在 whitelist，或使用者無權存取該資源 |
| `404` | 找不到對應使用者、大廈、門、QR 或付款類型 |
| `405` | webhook 收到非 `POST` |
| `409` | 登入電郵／電話衝突，或副戶關係已存在 |
| `500` | 未處理的伺服器錯誤 |
| `502` | 上游設備或 QR 產生失敗 |

## 介面總表

| 分類 | 方法 | 主要路徑 | 舊路徑 | 說明 |
|---|---|---|---|---|
| 認證 | `POST` | `/api/v1/integration/auth/login/` | `/api/v1/pos/login` | 登入驗證與身份回傳 |
| 認證 | `POST` | `/api/v1/integration/auth/register/` | — | 直接註冊 `CustomUser` 與 `ClientTbl` |
| 認證 | `GET` / `POST` | `/api/v1/integration/auth/check-contact/` | — | 查詢 email / phone 是否已被使用 |
| 認證 | `GET` / `POST` / `PATCH` / `PUT` | `/api/v1/integration/auth/client/` | — | 讀取或更新主會員檔案及登入電郵／電話 |
| 認證 | `POST` | `/api/v1/integration/auth/change-password/` | — | 驗證目前密碼後修改 `CustomUser` 密碼 |
| 認證 | `GET` / `POST` / `PATCH` / `PUT` | `/api/v1/integration/auth/settings/` | — | 讀取或更新 `UserSettings.blg_notice_email` |
| 大廈資料 | `GET` | `/api/v1/integration/buildings/receivables/management-fees/` | `/api/v1/building-mf-table/` | 管理費應收摘要 |
| 大廈資料 | `GET` | `/api/v1/integration/buildings/receivables/other-fees/` | `/api/v1/building-of-list/` | 其他費用應收列表 |
| 大廈資料 | `GET` | `/api/v1/integration/buildings/notices/` | `/api/v1/building-notices/` | 有效大廈通告 |
| 外部大廈 | `GET` | `/api/v1/integration/buildings/info/` | `/api/v1/external/building-info/` | 大廈資料與文件中繼 |
| 綁定 | `POST` | `/api/v1/integration/buildings/building-flat-owner-binding-requests/` | — | 提交 `OwnerReg` 業主綁定申請，需職員審批 |
| 副戶 | `GET` | `/api/v1/integration/buildings/subaccounts/` | — | 列出業主控制單位目前有效授權副戶 |
| 副戶 | `POST` | `/api/v1/integration/buildings/subaccounts/grant/` | — | 為一個業主控制單位授予 `授權用戶` |
| 副戶 | `POST` | `/api/v1/integration/buildings/subaccounts/revoke/` | — | 從一個業主控制單位撤銷 `授權用戶` |
| 服務個案 | `POST` | `/api/v1/integration/buildings/comments/` | `/api/v1/external/blg-cs/submit/` | 提交維修／意見個案 |
| 服務個案 | `GET` / `POST` | `/api/v1/integration/buildings/service-cases/` | — | 查詢個案列表 |
| 服務個案 | `GET` / `POST` | `/api/v1/integration/buildings/service-cases/<case_id>/` | — | 個案詳情與訊息 |
| 門禁 | `POST` | `/api/v1/integration/access/buildings/` | `/api/v1/external/building-access/` | 門清單、密碼、QR 與近期紀錄 |
| 門禁 | `POST` | `/api/v1/integration/access/open-door/` | `/api/v1/external/building-access/open-door/` | 遙距開門 |
| 門禁 | `POST` | `/api/v1/integration/access/qrcode/` | `/api/v1/external/building-access/qrcode/` | 產生門禁 QR payload |
| 付款 | `GET` | `/api/v1/integration/payments/unpaid-invoices/` | `/api/v1/building-flat-unpaid-invoice-list` | 單位未繳發票 |
| 付款 | `GET` / `POST` | `/api/v1/integration/payments/types/` | `/api/v1/get-building-payment-types` | 列出大廈 `BuildingPaymentType` |
| 付款 | `POST` | `/api/v1/integration/payments/pos/` | `/api/v1/pos-payment-to-ismart` | POS 或外部來源建立付款 |
| 收銀台 | `GET` | `/api/v1/integration/cashier/transactions/` | `/api/v1/get-transactions-in-cashier` | 列出大廈目前 cashier 交易 |
| 付款 | `GET` | `/api/v1/integration/payments/transactions/by-unit/` | `/api/v1/get-transactions-by-flat-unit` | 單位付款歷史 |
| 付款 | `GET` | `/api/v1/integration/payments/transactions/by-date/` | `/api/v1/get-transactions-by-date` | 按日期查詢付款歷史 |
| 收銀台 | `POST` | `/api/v1/integration/cashier/bank-in/` | `/api/v1/update-transactions-status-in-cashier` | 轉入 bank-in 批次 |
| 收銀台 | `GET` | `/api/v1/integration/cashier/bank-in-records/` | `/api/v1/get-bank-in-record-list` | bank-in 批次列表 |
| 收銀台 | `GET` | `/api/v1/integration/cashier/bank-in-records/details/` | `/api/v1/get-bank-in-record-details` | bank-in 批次明細 |
| Webhook | `POST` | `/api/v1/integration/webhooks/qfpay/` | `/qfpayapi/` | QFPay 回調 |
| Webhook | `POST` | `/api/v1/integration/webhooks/test/` | `/test_webhook/` | 測試回調 |

## 登入電郵與電話格式

以下規則適用於 `CustomUser` 的**登入憑證**，影響註冊、check-contact 與會員資料更新。

| 憑證 | 模型欄位 | 是否唯一 |
|---|---|---|
| 登入電郵 | `CustomUser.email` | 是，不分大小寫 |
| 登入電話 | `CustomUser.phone` | 是，經 E.164 正規化後唯一 |

`ClientTbl` 的聯絡／賬務欄位（`cli_email`、`cli_tel`、`cli_tel2`）與登入憑證分開。註冊時會把同一組電話／電郵預填到新 `ClientTbl` 方便使用。

### 電郵

- 必須是標準電郵，例如 `user@example.com`
- 會 trim 並轉小寫後儲存
- `A@x.com` 與 `a@x.com` 視為同一帳號
- 無 `@`、截斷域名、純電話字串、全形 `＠` 等，在嚴格校驗時回 HTTP `400`
- 建議客戶端直接送出小寫 ASCII 電郵

### 電話

- 儲存一律為帶 `+` 的 E.164，例如 `+85291234567`
- 必須通過 Google libphonenumber `is_valid_number`
- 只傳本地數字、沒有國家碼時，8 位香港號碼會當成 `+852`
- `91234567`、`85291234567`、`+85291234567` 正規化後視為同一登入電話
- 長度錯誤、假前綴、自由文字回 HTTP `400`

建議一律送完整 E.164：

| 地區 | 國家碼 | 本地位數 | 例子 |
|---|---|---|---|
| 香港 | `852` | 8 | `+85291234567` |
| 中國內地 | `86` | 11（流動通常 `1…`） | `+8613901234567` |
| 新加坡 | `65` | 8 | `+6586726893` |
| 澳洲 | `61` | 流動 9 位（不要前置 `0`） | `+61420909995` |
| 澳門 | `853` | 8 | `+8536xxxxxxx` |
| 台灣 | `886` | 流動通常 9 位（不要前置 `0`） | `+8869xxxxxxxx` |

其他國家只要送完整 `+` E.164 且 libphonenumber 判定有效即可。

不要送：把 `852` 直接黏在內地流動號碼前面，例如 `85218906680192`。

### 登入欄位與聯絡欄位

| API 欄位 | 實際更新 | 說明 |
|---|---|---|
| `email` | 只改 `CustomUser.email` | 不會覆寫 `cli_email` |
| `phone` | 改 `CustomUser.phone`；同一請求未帶 `cli_tel` 時也會寫入 `cli_tel` | 登入電話必須是有效 E.164 |
| `cli_email` | 只改 `ClientTbl.cli_email` | 聯絡／賬務電郵 |
| `cli_tel` | 只改 `ClientTbl.cli_tel` | 單獨送出時不改登入電話 |
| `cli_tel2` | 緊急聯絡電話 | 自由文字，不是登入憑證 |

## 1. 認證

### 1.1 POS / External 登入

- 主要路徑：`POST /api/v1/integration/auth/login/`
- 舊路徑：`POST /api/v1/pos/login`
- 用途：驗證帳密，回傳身份、屋苑與單位權限
- 授權：不發 token，只做憑證校驗

請求欄位：

| 欄位 | 必填 | 說明 |
|---|---|---|
| `login_name` | 是 | 使用者名稱、電郵或電話 |
| `password` | 是 | 密碼 |
| `login_type` | 否 | `username` / `email` / `phone`，預設 `username` |

注意：

- `email` 比對不分大小寫
- `phone` 以 E.164 為主，香港 8 位與舊 `852...` 形式仍會盡量兼容
- 成功回應沿用舊 `{"code":"1","msg":...}` 結構
- 成功後會更新 `CustomUser.last_login`
- `msg.building` 與 `staff_building_permissions` 目前內容相同
- `msg.client` 是 `cli_id == username` 的主檔案，不是完整 M2M `user.clients`

`msg` 重點欄位：

| 欄位 | 說明 |
|---|---|
| `user_id` | `CustomUser.id` |
| `username` | 登入名稱 |
| `email` / `phone` | 登入電郵／電話 |
| `is_staff` | 職員或 superuser 為 `true` |
| `staff_building_permissions` | 職員可管大廈 |
| `client_building_permissions` | 由會員單位推導的大廈 |
| `client_building_flat_units_permissions` | 有效單位 ID |
| `client` | 主 `ClientTbl`；沒有對應列時為 `null` |

### 1.2 直接註冊帳號

- 主要路徑：`POST /api/v1/integration/auth/register/`
- 舊路徑：無
- 用途：直接建立 `CustomUser` 與對應 `ClientTbl`，無需審批
- 授權：目前無 IP whitelist，亦無登入要求

請求欄位：

| 欄位 | 必填 | 說明 |
|---|---|---|
| `phone` | 是 | 登入電話，儲存為 E.164 |
| `email` | 是 | 登入電郵，轉小寫後唯一 |
| `eng_name` | 是 | 英文名稱 |
| `chi_name` | 否 | 中文名稱 |
| `legal_entity` | 否 | `NA` 個人、`LE` 法人，預設 `NA` |
| `id_card` | 否 | 身份證明號碼 |
| `remark` | 否 | 寫入 `ClientTbl.cli_remark` |
| `gender` | 否 | `M` / `F` / 空 |
| `is_receive_email` | 否 | 是否開啟 `UserSettings.blg_notice_email`，預設 `true` |
| `password` | 否 | 不傳則由後端產生隨機密碼 |

仍接受的舊欄位別名：`memberphone`、`memberemail`、`memberengname`、`memberchiname`、`member_legalentity`、`memberid`、`regremark`、`membergender`、`is_receiveemail`。

成功時 `data` 回傳正規化後的 `phone` / `email`，以及產生的 `username`、`password`、`client_id`。帳號已存在時回 `Account already exists`。

### 1.3 檢查電郵／電話可用性

- 主要路徑：`GET /api/v1/integration/auth/check-contact/`
- 亦支援 `POST`
- 舊路徑：無
- 用途：預先檢查 `CustomUser` 的 email / phone 是否已被使用
- 授權：無 IP whitelist，亦無登入要求

常用參數：

| 欄位 | 說明 |
|---|---|
| `email` | 查詢電郵 |
| `phone` | 查詢電話，建議 E.164 |
| `exclude_user_id` | 編輯既有帳號時排除自身 |

注意：

- 至少提供 `email` 或 `phone`
- 空字串視為可用
- 電話格式無效回 HTTP `400`
- 只檢查登入憑證，不阻擋無關 `ClientTbl` 自由文字
- `available` 只有在所有被檢查欄位都可用時才是 `true`

### 1.4 讀取／更新 Client 檔案

- 主要路徑：`GET /api/v1/integration/auth/client/`
- 更新方法：`POST` / `PATCH` / `PUT`
- 舊路徑：無
- 用途：讀取或更新主會員檔案（`cli_id == username`）
- 授權：無 IP whitelist；由 `user_id` 指定目標會員

`client` 物件形狀與登入回應的 `msg.client` 相同。

可更新重點欄位：

| 欄位 | 說明 |
|---|---|
| `user_id` | 必填，目標 `CustomUser.id` |
| `email` | 更新登入電郵，不覆寫 `cli_email` |
| `phone` | 更新登入電話；未同時傳 `cli_tel` 時會同步 `cli_tel` |
| `cli_legalentity` / `cli_type` | 法人類型、客戶類型 |
| `cli_contact_person` / `cli_urgent_contact_person` | 聯絡人、緊急聯絡人 |
| `cli_name` / `cli_chi_name` / `cli_acname` | 英文名、中文名、收款名稱 |
| `cli_id_card` | 證件號碼 |
| `cli_tel` / `cli_tel2` / `cli_fax` | 聯絡電話、緊急電話、傳真 |
| `cli_addr` / `cli_chiadd` | 英文／中文郵寄地址 |
| `cli_email` | 聯絡／賬務電郵，不改登入電郵 |
| `cli_birthday` | `YYYY-MM-DD` |
| `cli_sex` | `M` / `F` / `U` |
| `cli_nationality` / `cli_bank` / `cli_bank_acc` / `cli_remark` | 國籍、銀行、帳號、備註 |

不可更新：`cli_id`、`cli_join_dt`。

注意：

- 登入電郵／電話與其他帳號衝突回 HTTP `409`
- 登入電話格式無效回 HTTP `400`
- 改登入電郵／電話前，先用 check-contact 並帶 `exclude_user_id`
- 沒有主 `ClientTbl` 時，GET 回 `"client": null`，更新會失敗

### 1.5 修改密碼

- 主要路徑：`POST /api/v1/integration/auth/change-password/`
- 舊路徑：無
- 用途：為既有 `CustomUser` 修改登入密碼
- 授權：目前無 IP whitelist、無登入要求；呼叫端提供 `user_id`，並必須證明目前密碼

這是變更密碼，不是管理員重設。`old_password` 必填。

請求欄位：

| 欄位 | 必填 | 說明 |
|---|---|---|
| `user_id` | 是 | 目標 `CustomUser.id` |
| `old_password` | 是 | 目前密碼，必須通過 `user.check_password()` |
| `new_password` | 是 | 新密碼，走 Django `AUTH_PASSWORD_VALIDATORS`（最少 8 位、不可太像使用者名稱／電郵、不可常見密碼、不可全數字） |
| `new_password_confirm` | 否 | 若提供，必須等於 `new_password` |

成功回應只回 `user_id` 與 `username`，**永不回傳新密碼**。

注意：

- `old_password` 錯誤或帳號停用回 HTTP `401`
- 缺欄、確認不一致、未知 `user_id`、密碼不符合規則回 HTTP `400`
- 成功只更新 `password` 欄，不改 `last_login`
- 不發 session / JWT / API token
- Django 網站 cookie session 會因密碼雜湊變更而在下次請求失效

### 1.6 讀取／更新使用者設定

- 主要路徑：`GET` / `POST` / `PATCH` / `PUT` `/api/v1/integration/auth/settings/`
- 舊路徑：無
- 用途：讀取或更新 `accounts.UserSettings`（目前是 `blg_notice_email`）
- 授權：目前無 IP whitelist、無登入要求；由 `user_id` 指定目標

註冊時已用別名 `is_receive_email`。此接口寫入接受兩個欄位名，讀取也同時回傳兩個欄位，且兩者永遠相等。

GET：`?user_id=123`

更新時至少提供 `blg_notice_email` 或 `is_receive_email`；兩個都傳時必須解析成同一個 boolean。

接受的 boolean 字串（不分大小寫）：

| True | False |
|---|---|
| `true`, `1`, `yes`, `y`, `on` | `false`, `0`, `no`, `n`, `off` |

JSON boolean 與整數 `1` / `0` 也可。`null` 或其他值回 HTTP `400`。

注意：

- `user_id` 只是身份，不是設定欄位
- 未知鍵會被拒絕
- 此接口不改登入電郵／電話，也不改 `ClientTbl`

## 2. 大廈資料

### 2.1 管理費應收表

- 主要路徑：`GET /api/v1/integration/buildings/receivables/management-fees/`
- 舊路徑：`POST /api/v1/building-mf-table/`
- 用途：由 `blg_mf_list(...)` 產生管理費應收資料
- 授權：無 IP whitelist，亦無登入要求

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |
| `data_structure` | `table` / `list` / `block_dict`，預設 `table` |

注意：回應形狀會隨 `data_structure` 改變；後端以 `group_estate=True` 產生資料。

### 2.2 其他費用清單

- 主要路徑：`GET /api/v1/integration/buildings/receivables/other-fees/`
- 舊路徑：`POST /api/v1/building-of-list/`
- 用途：由 `blg_mf_list(...)` 產生非管理費應收資料
- 授權：無 IP whitelist，亦無登入要求

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |
| `data_structure` | 回應佈局 |

空結果回 `[]`。

### 2.3 大廈通告

- 主要路徑：`GET /api/v1/integration/buildings/notices/`
- 舊路徑：`POST /api/v1/building-notices/`
- 用途：回傳有效大廈通告
- 授權：無 IP whitelist，亦無登入要求

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |

注意：

- 回應是原始陣列，沒有物件包一層
- 只回 `mess_date <= today < mess_down` 的通告

### 2.4 大廈資料與文件

- 主要路徑：`GET /api/v1/integration/buildings/info/`
- 舊路徑：`POST /api/v1/external/building-info/`
- 用途：回傳大廈資料、管理資料與文件中繼
- 授權：`EXTERNAL_APP_ALLOWED_IPS`

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |

文件分組會一併回傳。目前審計報告查詢使用 `btype='audition'`。

## 3. 綁定與副戶

業主角色（`登記業主`）經 `OwnerReg` 提交，仍需大廈管理審批。已核准業主授予 `授權用戶` 時不走審批。實際有效角色關係仍寫在 `i_flatcli_rel_tbl`。

### 3.1 業主綁定申請

- 主要路徑：`POST /api/v1/integration/buildings/building-flat-owner-binding-requests/`
- 舊路徑：無
- 用途：提交 `OwnerReg`，把使用者綁到一個或多個單位，供業主角色審批
- 授權：目前無 IP whitelist，亦無登入要求

關鍵欄位：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |
| `ownedflat` | 單位 ID 陣列或逗號字串，必須屬於該大廈 |
| `user` | 既有 `CustomUser.id`；提供後可自動補申請人資料 |
| `cli_role` | 預設業主角色 |
| `reg_tel` / `reg_email` | 沒有 `user` 時需要的申請人聯絡資料 |
| `cli_name` / `cli_id_card` / `cli_tel` | 申請人資料 |

注意：

- 只建立 `OwnerReg` 列，不會立即授予角色關係
- 職員其後經既有 `confirm_memreg` 流程審批
- 若 `OwnerReg.user` 已有值，審批會重用既有帳號，不另開新帳

### 3.2 副戶清單

- 主要路徑：`GET /api/v1/integration/buildings/subaccounts/`
- 舊路徑：無
- 用途：列出呼叫端已是核准 `登記業主` 的單位下，目前有效的 `授權用戶`
- 授權：必須是該單位的有效 `登記業主`；不傳 `unit_id` 時包含其全部業主單位

主要參數：

| 欄位 | 說明 |
|---|---|
| `user_id` | 業主 `CustomUser.id` |
| `unit_id` | 可選單位篩選 |

注意：

- 只回 `cli_role='授權用戶'` 且有效的 `i_flatcli_rel_tbl`
- `history_count` 是該關係的授權副戶歷史筆數
- 指定 `unit_id` 但呼叫端不是該單位業主時回 HTTP `403`

### 3.3 副戶授權

- 主要路徑：`POST /api/v1/integration/buildings/subaccounts/grant/`
- 舊路徑：無
- 用途：已核准業主把 `授權用戶` 授給另一個既有 `CustomUser`，無需職員審批
- 授權：必須已是該單位有效 `登記業主`

請求欄位：

| 欄位 | 必填 | 說明 |
|---|---|---|
| `user_id` | 是 | 業主 `CustomUser.id` |
| `unit_id` | 是 | 目標單位 |
| `target_user_id` | 是 | 要成為 `授權用戶` 的既有會員 |
| `remark` | 否 | 寫入關係與歷史的備註 |

驗證：

- 目標使用者必須已存在且有 `ClientTbl`
- 不可授權給自己
- 每單位最多 4 個有效 `授權用戶`
- 目標在該單位已有有效 `授權用戶` 關係時回 HTTP `409`

注意：若同一目標先前有已停用的 `授權用戶` 列，現有程式會重新啟用該列，而不是再插一筆。成功授權會寫入 `AuthorizedSubUserRequestHistory`。

### 3.4 副戶撤銷

- 主要路徑：`POST /api/v1/integration/buildings/subaccounts/revoke/`
- 舊路徑：無
- 用途：業主停用某單位的有效 `授權用戶`
- 授權：必須已是該單位有效 `登記業主`

主要參數：`user_id`、`unit_id`、`target_user_id`、可選 `remark`。

撤銷實作為把既有 `i_flatcli_rel_tbl` 的 `is_active` 設為 `False`。

## 4. 服務個案

### 4.1 提交服務個案

- 主要路徑：`POST /api/v1/integration/buildings/comments/`
- 舊路徑：`POST /api/v1/external/blg-cs/submit/`
- 用途：建立 `BuildingServiceCase`（維修報修／意見反映），寫入第一則訊息，並發職員通知電郵
- 授權：`EXTERNAL_APP_ALLOWED_IPS`，並校驗使用者與大廈關聯

建議新分類欄位：

| 欄位 | 說明 |
|---|---|
| `user_id` | 會員 ID |
| `building_id` | 大廈編號 |
| `request_type` | `repair` / `feedback` |
| `category` | 分類 |
| `subcategory` | 子分類 |
| `subject` | 標題 |
| `content` | 內容 |
| `location_text` | 位置描述 |
| `unit_id` | 單位編號 |
| `contact_name` | 聯絡人 |
| `contact_phone` | 聯絡電話 |
| `comment_type` | 舊欄位兼容值，由伺服器映射 |

舊 `comment_type` 仍可接受，由伺服器映射到新 taxonomy。

### 4.2 服務個案列表

- 主要路徑：`GET /api/v1/integration/buildings/service-cases/`
- 亦支援 `POST`
- 用途：列出某大廈下的個案，可按狀態／類型篩選
- 授權：`EXTERNAL_APP_ALLOWED_IPS`，並校驗使用者與大廈關聯

主要參數：

| 欄位 | 說明 |
|---|---|
| `user_id` | 會員 ID |
| `building_id` | 大廈編號 |
| `status` | 個案狀態 |
| `request_type` | `repair` / `feedback` |
| `scope` | `mine` 或 `building` |
| `limit` | 筆數上限 |

### 4.3 服務個案詳情

- 主要路徑：`GET /api/v1/integration/buildings/service-cases/<case_id>/`
- 亦支援 `POST`
- 用途：查看單一個案與訊息紀錄
- 授權：`EXTERNAL_APP_ALLOWED_IPS`，並校驗使用者與大廈關聯

主要參數：

| 欄位 | 說明 |
|---|---|
| `user_id` | 會員 ID |
| `building_id` | 大廈編號 |
| `case_id` | 個案 ID |

## 5. 門禁

### 5.1 門禁摘要

- 主要路徑：`POST /api/v1/integration/access/buildings/`
- 舊路徑：`POST /api/v1/external/building-access/`
- 用途：返回門清單、密碼、QR 與最近紀錄
- 授權：`EXTERNAL_APP_ALLOWED_IPS`，並校驗使用者與大廈關聯

主要參數：

| 欄位 | 說明 |
|---|---|
| `user_id` | 會員 ID |
| `building_id` | 大廈編號 |

語義上是唯讀，但目前依賴 body 中的 `user_id`，並回傳門禁密碼等敏感資料，因此維持 `POST`。門清單權限邏輯與遙距開門權限邏輯並不完全相同。

### 5.2 遙距開門

- 主要路徑：`POST /api/v1/integration/access/open-door/`
- 舊路徑：`POST /api/v1/external/building-access/open-door/`
- 用途：觸發遠端開門

主要參數：

| 欄位 | 說明 |
|---|---|
| `user_id` | 會員 ID |
| `building_id` | 大廈編號 |
| `door_id` | 門 ID |

上游設備失敗時可能回 HTTP `502`。

### 5.3 產生門禁 QR

- 主要路徑：`POST /api/v1/integration/access/qrcode/`
- 舊路徑：`POST /api/v1/external/building-access/qrcode/`
- 用途：產生門禁 QR payload

主要參數：

| 欄位 | 說明 |
|---|---|
| `user_id` | 會員 ID |
| `building_id` | 大廈編號 |
| `door_id` | 門 ID |

注意：接口只回 payload 文字，客戶端需自行產生 QR 圖。現有程式在產生前不會單獨拒絕已過期的 QR 紀錄。

## 6. 付款

### 6.1 未繳發票

- 主要路徑：`GET /api/v1/integration/payments/unpaid-invoices/`
- 舊路徑：`POST /api/v1/building-flat-unpaid-invoice-list`
- 用途：查詢單位目前未繳賬單，已扣減 cashier 或驗證中的待處理付款
- 授權：無 IP whitelist、無登入要求

主要參數：

| 欄位 | 說明 |
|---|---|
| `unit_id` | 單位編號 |

注意：

- 回應是原始陣列
- `net_amount` 是元，不是分。JSON number 是 IEEE 浮點，例如 `0.29` 並不精確。**不要**用 `int(net_amount * 100)` 轉分，這會把 `0.29` 截成 28 分
- `net_amount_cents` 是以 Decimal 四捨五入後的整數分。POS 應把此值作為 `BILL_OBJS.net_amount`，並加總為 `FINAL_AMOUNT`
- 狀態為 `in_cashier` 或 `pending_validation` 的待處理付款會從未繳金額扣除

### 6.2 付款類型列表

- 主要路徑：`GET /api/v1/integration/payments/types/`
- 舊路徑：`POST /api/v1/get-building-payment-types`
- 用途：回傳某大廈已設定的 `BuildingPaymentType`
- 授權：無 IP whitelist、無登入要求

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |

兩條路徑都接受 `GET` query 與 `POST` JSON。

成功時回 `payment_types` 陣列，欄位包括 `pay_type`、`pay_type_name_chi`、`markup`、`is_active`。

注意：

- 只限所請求的大廈
- 停用列也會回傳，客戶端應讀 `is_active`
- `pay_type` 就是 POS 呼叫 `/api/v1/integration/payments/pos/` 時的 `PAY_METHOD`
- `markup` 是手續費率，例如 `0.03` = 3%
- 沒有設定時回空陣列

### 6.3 POS 付款上傳

- 主要路徑：`POST /api/v1/integration/payments/pos/`
- 舊路徑：`POST /api/v1/pos-payment-to-ismart`
- 用途：建立 `Payment` 與 `Paymentdetails`；特定付款方式會立即建立 `BalCalcTbl` 與 `PrepaidBal`
- 授權：目前無有效 IP whitelist
- 內容類型：`application/json`。圖片**不是** multipart，而是 JSON 內的 base64 字串

重點請求欄位：

| 欄位 | 說明 |
|---|---|
| `BUILDING_ID` | 大廈編號 |
| `PAY_METHOD` | 必須是該大廈已設定的 `BuildingPaymentType` |
| `FINAL_AMOUNT` | 總金額，單位為分 |
| `BILL_OBJS` | 付款明細陣列 |
| `PIC_FILENAME` / `PIC_DATA` | 檔名與 base64，按同索引對應，建議用陣列 |
| `FAKE_TRANSACTION` | 為 true 時狀態變成 `pending_validation` |
| `bank_account_received` | 收款銀行帳號；未知值會被忽略而不是拒絕 |

`BILL_OBJS` 項目：

| 欄位 | 必填 | 說明 |
|---|---|---|
| `flat_code` | 是 | 單位 ID |
| `invoice_no` | 否 | 發票號；預付可空 |
| `item_id` | 是 | 費用項目名稱 |
| `trs_to` | 是 | 賬期或 `預付` |
| `net_amount` | 是 | 金額，單位為分。應使用未繳發票接口的 `net_amount_cents`，不要用 `int(dollar_float * 100)` |
| `transfer_fee` | 否 | 手續費，單位為分，預設 `0` |

注意：

- 請求金額一律用分；寫入 `Payment.txamount` 與明細時會轉成元
- `WEBPOS_*` 會立即建立付款端會計列
- `trs_to = "預付"` 會建立 `PrepaidBal`
- 手續費會另建 `平台手續費` 明細
- 有圖片時每張解碼後寫一筆 `PaymentPicture` 到 S3 `payment/`
- JSON body 超過 5 MB、檔名與資料長度不一致、或 base64 無效，回 `400`
- 非預付列的金額不可超過目前未繳金額
- 不要送 multipart 檔案欄位
- 相機原圖應先壓縮；base64 會讓 JSON 比原檔更大

### 6.4 Cashier 交易

- 主要路徑：`GET /api/v1/integration/cashier/transactions/`
- 舊路徑：`POST /api/v1/get-transactions-in-cashier`
- 用途：列出某大廈目前 cashier 交易

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |

注意：只回狀態 `in_cashier`；結果拆成 `payment_objs_cheque` 與 `payment_objs_cash`，沒有統一列表。

### 6.5 單位交易歷史

- 主要路徑：`GET /api/v1/integration/payments/transactions/by-unit/`
- 舊路徑：`POST /api/v1/get-transactions-by-flat-unit`
- 用途：列出一個或多個單位的付款歷史

主要參數：

| 欄位 | 說明 |
|---|---|
| `unit_id_list` | 單位 ID，可重複傳入 |

注意：

- 排除 `init` 與 `pending_validation`
- `validated_by_ismart`、`payment_captured`、`confirmed` 會正規化成 `confirmed`
- 預付消耗可能顯示為 `"{trs_to}(預付)"`

### 6.6 日期交易歷史

- 主要路徑：`GET /api/v1/integration/payments/transactions/by-date/`
- 舊路徑：`POST /api/v1/get-transactions-by-date`
- 用途：依日期區間查詢大廈付款歷史

主要參數：

| 欄位 | 必填 | 說明 |
|---|---|---|
| `building_id` | 是 | 大廈編號 |
| `from_date` / `to_date` | 是 | 日期區間 |
| `date_type` | 是 | `input_date` 或 `tran_date` |
| `pay_method` | 否 | `all` 或 `pos` |

注意：`pay_method = "pos"` 只保留名稱以 `POS` 開頭的付款類型；查詢會對日期範圍手動 `-8h` / `+8h`。

### 6.7 轉入 cashier 的 bank-in

- 主要路徑：`POST /api/v1/integration/cashier/bank-in/`
- 舊路徑：`POST /api/v1/update-transactions-status-in-cashier`
- 用途：把選定 cashier 付款組成新的 bank-in 批次，並改為 `pending_validation`

主要參數：

| 欄位 | 說明 |
|---|---|
| `payment_id_list` | 付款 ID 陣列 |

### 6.8 bank-in 批次列表

- 主要路徑：`GET /api/v1/integration/cashier/bank-in-records/`
- 舊路徑：`POST /api/v1/get-bank-in-record-list`
- 用途：列出某大廈最近的 bank-in 批次

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |

### 6.9 bank-in 批次明細

- 主要路徑：`GET /api/v1/integration/cashier/bank-in-records/details/`
- 舊路徑：`POST /api/v1/get-bank-in-record-details`
- 用途：列出某批次內的付款明細

主要參數：

| 欄位 | 說明 |
|---|---|
| `record_id` | 批次 ID |

## 7. Webhook

### 7.1 QFPay 回調

- 主要路徑：`POST /api/v1/integration/webhooks/qfpay/`
- 舊路徑：`POST /qfpayapi/`
- 用途：接收並保存 QFPay 付款回調，完成成功通知

注意：

- 目前僅檢查 `X-QF-SIGN` 是否存在，不驗證簽名
- 回應為純文字 `SUCCESS` / `UNSUCCESS`
- 非 `POST` 回 HTTP `405`

### 7.2 測試回調

- 主要路徑：`POST /api/v1/integration/webhooks/test/`
- 舊路徑：`POST /test_webhook/`
- 用途：測試用 payload 接收，並嘗試電郵一份診斷副本

注意：

- 非正式業務接口
- 沒有認證、簽名或 schema 校驗
- 回應純文字 `success`

## 8. 共同驗證規則與已知限制

- 大廈編號通常只允許英數字
- 錯誤 JSON、缺必填欄、無效值通常回 `400`
- 大廈、使用者、門、QR、交易批次查詢失敗通常回 `404`
- 外部安全接口通常依 body 中的 `user_id` + `EXTERNAL_APP_ALLOWED_IPS` 控制
- 回應 envelope 不統一：有的包 `status/data`，有的用舊 `code/message`，有的直接回陣列，webhook 則是純文字
- 付款接口混用字串狀態碼與 `status` / `message` 物件
- 部分付款查詢使用 raw SQL 並手動調整時區
- `PosPaymentToIsmart` 目前沒有有效 IP whitelist
- `qfpayapi` 不驗證回調簽名
- `ExternalBuildingInfoApi` 目前用 `btype='audition'` 查審計報告
- 門禁列表權限與遙距開門權限並不完全相同
- `clouddev` 與 production 的硬件能力可能不同

## 9. 方法設計建議

- 新的唯讀 `/api/v1/integration/...` 接口使用 `GET`
- 舊非整合別名仍以 `POST` 提供
- 寫入、webhook 與帶敏感 body 的門禁流程維持 `POST`

適合 `GET` 的例子：管理費、其他費用、通告、大廈資料、未繳發票、cashier 交易、單位／日期交易歷史、bank-in 列表與明細。

應維持 `POST` 的例子：登入、提交服務個案、遙距開門、產生 QR、POS 付款、bank-in、webhook。

門禁摘要語義上可讀，但目前靠 body `user_id` 且含敏感資料，在改成 token 認證前應繼續用 `POST`。

## 10. 操作檢查清單

- 生產環境接入前填好 `EXTERNAL_APP_ALLOWED_IPS`
- 決定 `PosPaymentToIsmart` 是否應啟用 `POS_PAYMENT_ALLOWED_IPS`
- 若審計檔必須出現在大廈資料接口，需對齊 `btype` 命名
- 對目標硬件環境測試開門與 QR
- 把 `test_webhook/` 當非生產接口；不需要就停用或加上保護
