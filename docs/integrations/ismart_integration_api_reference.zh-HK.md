# iSmart 整合 API 參考

## 目的

本文整理 iSmart 目前對外可用的整合接口，保留舊路徑兼容說明，並按接口逐項列出用途、請求、回應與注意事項。

## 環境

- production: `https://ismart.ajoliving.com`
- clouddev: `https://clouddev.ismart.ajoliving.com`

`clouddev` 的門禁、硬件與部分設備相關能力，可能因環境依賴而不完整。

## 路徑策略

- 新整合主路徑：`/api/v1/integration/...`
- 舊兼容路徑：`/api/v1/...`、`/api/v1/external/...`、以及少量舊 webhook 路徑
- 現有客戶端可先保留舊路徑，建議新接入改用整合主路徑

## 共同規則

- `application/json` 為主要請求格式
- `GET` 介面多以 query string 傳參
- 會員身份通常由 `user_id` 提供，外部安全接口再疊加 IP whitelist
- 回應格式不完全統一，需按單接口契約處理

## 介面總表

| 分類 | 方法 | 主要路徑 | 舊路徑 | 說明 |
|---|---|---|---|---|
| 認證 | POST | `/api/v1/integration/auth/login/` | `/api/v1/pos/login` | 登入驗證與身份回傳 |
| 認證 | POST | `/api/v1/integration/auth/register/` | — | 直接註冊 `CustomUser` 與 `ClientTbl` |
| 認證 | GET/POST | `/api/v1/integration/auth/check-contact/` | — | 查詢 email / phone 是否已被使用 |
| 認證 | GET/POST/PATCH/PUT | `/api/v1/integration/auth/client/` | — | 讀取或更新主會員檔案 |
| 大廈資料 | GET | `/api/v1/integration/buildings/receivables/management-fees/` | `/api/v1/building-mf-table/` | 管理費應收資料 |
| 大廈資料 | GET | `/api/v1/integration/buildings/receivables/other-fees/` | `/api/v1/building-of-list/` | 其他費用清單 |
| 大廈資料 | GET | `/api/v1/integration/buildings/notices/` | `/api/v1/building-notices/` | 大廈通告 |
| 大廈資料 | GET | `/api/v1/integration/buildings/info/` | `/api/v1/external/building-info/` | 大廈資料與文件中繼 |
| 綁定 | POST | `/api/v1/integration/buildings/building-flat-owner-binding-requests/` | — | 業主綁定申請 |
| 副戶 | GET | `/api/v1/integration/buildings/subaccounts/` | — | 副戶清單 |
| 副戶 | POST | `/api/v1/integration/buildings/subaccounts/grant/` | — | 新增副戶授權 |
| 副戶 | POST | `/api/v1/integration/buildings/subaccounts/revoke/` | — | 撤銷副戶授權 |
| 服務個案 | POST | `/api/v1/integration/buildings/comments/` | `/api/v1/external/blg-cs/submit/` | 提交維修／意見個案 |
| 服務個案 | GET/POST | `/api/v1/integration/buildings/service-cases/` | — | 查詢個案列表 |
| 服務個案 | GET/POST | `/api/v1/integration/buildings/service-cases/<case_id>/` | — | 個案詳情與訊息 |
| 門禁 | POST | `/api/v1/integration/access/buildings/` | `/api/v1/external/building-access/` | 門禁摘要 |
| 門禁 | POST | `/api/v1/integration/access/open-door/` | `/api/v1/external/building-access/open-door/` | 遙距開門 |
| 門禁 | POST | `/api/v1/integration/access/qrcode/` | `/api/v1/external/building-access/qrcode/` | 產生 QR |
| 付款 | GET | `/api/v1/integration/payments/unpaid-invoices/` | `/api/v1/building-flat-unpaid-invoice-list` | 未繳發票 |
| 付款 | POST | `/api/v1/integration/payments/pos/` | `/api/v1/pos-payment-to-ismart` | POS 付款上傳 |
| 付款 | GET | `/api/v1/integration/cashier/transactions/` | `/api/v1/get-transactions-in-cashier` | Cashier 交易 |
| 付款 | GET | `/api/v1/integration/payments/transactions/by-unit/` | `/api/v1/get-transactions-by-flat-unit` | 單位交易歷史 |
| 付款 | GET | `/api/v1/integration/payments/transactions/by-date/` | `/api/v1/get-transactions-by-date` | 日期交易歷史 |
| 付款 | POST | `/api/v1/integration/cashier/bank-in/` | `/api/v1/update-transactions-status-in-cashier` | 轉入 bank-in 批次 |
| 付款 | GET | `/api/v1/integration/cashier/bank-in-records/` | `/api/v1/get-bank-in-record-list` | bank-in 批次列表 |
| 付款 | GET | `/api/v1/integration/cashier/bank-in-records/details/` | `/api/v1/get-bank-in-record-details` | bank-in 批次明細 |
| Webhook | POST | `/api/v1/integration/webhooks/qfpay/` | `/qfpayapi/` | QFPay 回調 |
| Webhook | POST | `/api/v1/integration/webhooks/test/` | `/test_webhook/` | 測試回調 |

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

### 1.2 直接註冊帳號

- 主要路徑：`POST /api/v1/integration/auth/register/`
- 舊路徑：無
- 用途：直接建立 `CustomUser` 與對應 `ClientTbl`
- 授權：目前無 IP whitelist，亦無登入要求

請求重點欄位：

| 欄位 | 必填 | 說明 |
|---|---|---|
| `phone` | 是 | 登入電話，會正規化為 E.164 |
| `email` | 是 | 登入電郵，會轉小寫 |
| `eng_name` | 是 | 英文名稱 |
| `chi_name` | 否 | 中文名稱 |
| `legal_entity` | 否 | 個人或法人類型 |
| `id_card` | 否 | 身份證明號碼 |
| `gender` | 否 | `M` / `F` |
| `remark` | 否 | 備註 |
| `is_receive_email` | 否 | 是否接收通告電郵 |

### 1.3 檢查電郵／電話可用性

- 主要路徑：`GET /api/v1/integration/auth/check-contact/`
- 舊路徑：無
- 用途：預先檢查 `CustomUser` 的 email / phone 是否已被使用
- 授權：無 IP whitelist，亦無登入要求

常用參數：

| 欄位 | 說明 |
|---|---|
| `email` | 查詢電郵 |
| `phone` | 查詢電話 |
| `exclude_user_id` | 排除自身帳號 |

### 1.4 讀取／更新 Client 檔案

- 主要路徑：`GET /api/v1/integration/auth/client/`
- 更新方法：`POST` / `PATCH` / `PUT`
- 舊路徑：無
- 用途：讀取或更新主會員檔案
- 授權：無 IP whitelist；目前由 `user_id` 指定目標會員

更新時可分為兩種模式：

- 只改登入電郵／電話
- 只改聯絡檔案，不改登入電郵／電話

## 2. 大廈資料

### 2.1 管理費應收表

- 主要路徑：`GET /api/v1/integration/buildings/receivables/management-fees/`
- 舊路徑：`POST /api/v1/building-mf-table/`
- 用途：管理費應收資料

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |
| `data_structure` | `table` / `list` / `block_dict` |

注意：回應格式會依 `data_structure` 改變。

### 2.2 其他費用清單

- 主要路徑：`GET /api/v1/integration/buildings/receivables/other-fees/`
- 舊路徑：`POST /api/v1/building-of-list/`
- 用途：其他費用應收資料

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |

### 2.3 大廈通告

- 主要路徑：`GET /api/v1/integration/buildings/notices/`
- 舊路徑：`POST /api/v1/building-notices/`
- 用途：回傳有效大廈通告

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 可選；不傳時由後端選擇可見大廈 |

### 2.4 大廈資料與文件

- 主要路徑：`GET /api/v1/integration/buildings/info/`
- 舊路徑：`POST /api/v1/external/building-info/`
- 用途：回傳大廈資料、管理資料與文件中繼
- 授權：`EXTERNAL_APP_ALLOWED_IPS`

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |

## 3. 綁定與副戶

### 3.1 業主綁定申請

- 主要路徑：`POST /api/v1/integration/buildings/building-flat-owner-binding-requests/`
- 用途：提交 `OwnerReg` 業主綁定申請
- 授權：無 IP whitelist；需提交正確大廈與單位

關鍵欄位：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |
| `ownedflat` | 單位 ID 陣列或逗號字串 |
| `user` | 已存在會員 ID |
| `cli_role` | 預設 `業主` |
| `reg_tel` / `reg_email` | 新申請者資料 |
| `cli_name` / `cli_id_card` / `cli_tel` | 申請人資料 |

### 3.2 副戶清單

- 主要路徑：`GET /api/v1/integration/buildings/subaccounts/`
- 用途：列出目前有效授權副戶

主要參數：

| 欄位 | 說明 |
|---|---|
| `unit_id` | 單位編號 |

### 3.3 副戶授權

- 主要路徑：`POST /api/v1/integration/buildings/subaccounts/grant/`
- 用途：新增授權副戶

主要參數：

| 欄位 | 說明 |
|---|---|
| `unit_id` | 單位編號 |
| `target_phone` | 目標電話 |
| `target_email` | 目標電郵 |

### 3.4 副戶撤銷

- 主要路徑：`POST /api/v1/integration/buildings/subaccounts/revoke/`
- 用途：撤銷副戶授權

主要參數：

| 欄位 | 說明 |
|---|---|
| `unit_id` | 單位編號 |
| `target_user_id` | 目標會員 ID |

## 4. 服務個案

### 4.1 提交服務個案

- 主要路徑：`POST /api/v1/integration/buildings/comments/`
- 舊路徑：`POST /api/v1/external/blg-cs/submit/`
- 用途：提交維修或意見個案
- 授權：`EXTERNAL_APP_ALLOWED_IPS`，並校驗使用者與大廈關聯

建議欄位：

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
| `comment_type` | 舊欄位兼容值 |

### 4.2 服務個案列表

- 主要路徑：`GET /api/v1/integration/buildings/service-cases/`
- 亦支援 `POST`
- 用途：列出某大廈下的個案

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

## 6. 付款

### 6.1 未繳發票

- 主要路徑：`GET /api/v1/integration/payments/unpaid-invoices/`
- 舊路徑：`POST /api/v1/building-flat-unpaid-invoice-list`
- 用途：查詢單位未繳賬單
- 授權：無 IP whitelist、無登入要求

主要參數：

| 欄位 | 說明 |
|---|---|
| `unit_id` | 單位編號 |

### 6.2 POS 付款上傳

- 主要路徑：`POST /api/v1/integration/payments/pos/`
- 舊路徑：`POST /api/v1/pos-payment-to-ismart`
- 用途：建立付款與付款明細
- 授權：目前無有效 IP whitelist

重點：

- 圖片以 JSON base64 陣列傳遞
- `PIC_FILENAME` 與 `PIC_DATA` 以同索引對應

### 6.3 Cashier 交易

- 主要路徑：`GET /api/v1/integration/cashier/transactions/`
- 舊路徑：`POST /api/v1/get-transactions-in-cashier`
- 用途：列出某大廈目前 cashier 交易

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |

### 6.4 單位交易歷史

- 主要路徑：`GET /api/v1/integration/payments/transactions/by-unit/`
- 舊路徑：`POST /api/v1/get-transactions-by-flat-unit`
- 用途：列出一個或多個單位的付款歷史

主要參數：

| 欄位 | 說明 |
|---|---|
| `unit_id_list` | 單位 ID，可重複傳入 |

### 6.5 日期交易歷史

- 主要路徑：`GET /api/v1/integration/payments/transactions/by-date/`
- 舊路徑：`POST /api/v1/get-transactions-by-date`
- 用途：依日期區間查詢付款歷史

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |
| `from_date` / `to_date` | 日期區間 |

### 6.6 轉入 cashier 的 bank-in

- 主要路徑：`POST /api/v1/integration/cashier/bank-in/`
- 舊路徑：`POST /api/v1/update-transactions-status-in-cashier`
- 用途：把選定交易轉成 bank-in 批次

主要參數：

| 欄位 | 說明 |
|---|---|
| `payment_id_list` | 付款 ID 陣列 |

### 6.7 bank-in 批次列表

- 主要路徑：`GET /api/v1/integration/cashier/bank-in-records/`
- 舊路徑：`POST /api/v1/get-bank-in-record-list`
- 用途：列出 bank-in 批次

主要參數：

| 欄位 | 說明 |
|---|---|
| `building_id` | 大廈編號 |

### 6.8 bank-in 批次明細

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
- 用途：接收 QFPay 付款回調

注意：

- 目前僅檢查 `X-QF-SIGN` 是否存在
- 回應為純文字 `SUCCESS` / `UNSUCCESS`

### 7.2 測試回調

- 主要路徑：`POST /api/v1/integration/webhooks/test/`
- 舊路徑：`POST /test_webhook/`
- 用途：測試用 payload 接收

注意：

- 非正式業務接口
- 只做 debug 和紀錄

## 8. 共同驗證規則與已知限制

- 錯誤 JSON、缺參數、無效編號通常回 `400`
- 建築、用戶、門、QR、交易批次查詢失敗通常回 `404`
- 外部安全接口通常依 `user_id` + `EXTERNAL_APP_ALLOWED_IPS` 控制
- 付款與 webhook 介面回應格式不一致，客戶端不可假設統一 envelope
- `clouddev` 與 production 的硬件能力可能不同

