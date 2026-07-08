# 會員中心與我的大廈 API 需求說明

## 目標

本文件整理目前前端已標註「需要介面」的會員中心與我的大廈功能，供後端工程師拆分實作。

前端已存在頁面與互動骨架，但以下功能仍未完整接入真實資料：

- 會員中心：授權副戶、物業綁定。
- 我的大廈：最新通告、大廈財務、業戶帳目、意見提供/維修報修、設備監測。

所有 API 預設使用 `/api/v1` 前綴，回應沿用現有格式：

```json
{
  "code": "OK",
  "message": "success",
  "data": {},
  "request_id": "01KWDXXXX",
  "timestamp": "2026-07-01T00:00:00Z"
}
```

## 已有可復用介面

| 模組 | 既有介面 | 現況 |
| --- | --- | --- |
| 目前會員 | `GET /me`、`PATCH /me/profile` | 已可取得與更新會員資料。 |
| 可見 POS 大廈與單位 | `GET /me/pos/buildings`、`GET /me/pos/buildings/{buildingId}/units` | 已可用於物業綁定表單候選資料。 |
| POS 賬務 | `GET /me/payments/pos/overview`、`GET /me/payments/pos/bills`、`GET /me/payments/pos/history`、`GET /me/payments/pos/accounting` | 可承接業戶帳目與部分大廈財務資料，需前端展示欄位對齊。 |
| iSmart 大廈資料 | `GET /me/ismart/buildings`、`GET /me/ismart/building-info` | 已可讀目前會員可見大廈與大廈基本資料、表格文件。 |
| iSmart 意見提交 | `POST /me/ismart/building-comments` | 可短期承接意見/維修提交，但只有提交能力，沒有 AJO 工單列表、狀態、附件與回覆追蹤。 |
| iSmart 門禁 | `GET /me/ismart/building-access`、`POST /me/ismart/building-access/open-door`、`POST /me/ismart/building-access/qrcode` | 已接入智能門禁，不在本輪缺口內。 |
| OSS | `POST /oss/presign`、`POST /oss/complete`、`GET /oss/assets` | 可用於物業綁定證明文件、維修/意見附件上載。 |
| 通知中心 | `GET /notifications`、`POST /staff/system-notices` | 系統通知已進通知中心；「最新通告」本輪只需要返回大廈通告頁 URL。 |

## 建議優先級

| 優先級 | 功能 | 交付目標 |
| --- | --- | --- |
| P0 | 意見提供/維修報修 | 只保留兩個介面：查詢最近記錄、提交新事項。 |
| P0 | 最新通告 | 只返回目前會員可見大廈的通告頁 URL，前端直接開啟。 |
| P1 | 業戶帳目 | 先用 POS 賬單/付款記錄接入會員可見單位。 |
| P1 | 物業綁定 | 一個介面提交綁定申請與文件，審批流程由後台既有或後續入口處理。 |
| P2 | 大廈財務 | 一個介面返回管理費總覽、財務報表、核數報告。 |
| P2 | 授權副戶 | 一個介面返回只讀列表。 |
| P2 | 設備監測 | 一個介面返回設備狀態列表。 |

## 共同權限要求

- 會員端 API 必須校驗目前登入會員是否可見該 `building_id`、`unit_id`。
- Staff 端 API 必須校驗管理員是否有對應大廈或管理公司權限。
- 前端不得提交舊系統 `user_id` 作為權限依據；後端需由 AJO token 解析目前會員，再映射 iSmart / POS 帳戶。
- 涉及維修/意見提交結果、物業綁定審批結果的提醒，應寫入通知中心，不應寫入訊息管理聊天會話。

## 1. 會員可見大廈上下文

目前多個功能都需要同一個「會員可見大廈/單位」選擇器。建議新增 AJO 穩定介面，不讓前端長期直接依賴 iSmart 或 POS 的原始結構。

### 1.1 查詢我的可見大廈

`GET /me/buildings`

Query：

| 欄位 | 類型 | 必填 | 說明 |
| --- | --- | --- | --- |
| `include_units` | boolean | 否 | 是否一併返回可見單位。 |

Response：

```json
{
  "items": [
    {
      "building_id": "0348200",
      "building_name": "協和大廈",
      "source": "ismart",
      "role": "owner",
      "is_staff": false,
      "units": [
        {
          "unit_id": "034820001A",
          "floor": "01",
          "unit": "A",
          "display_name": "01樓 A"
        }
      ]
    }
  ]
}
```

## 2. 最新通告

本輪只需要一個介面，返回目前會員可見大廈的通告頁 URL。前端不需要在 AJO 內做通告列表、詳情、已讀或發佈。

### 2.1 查詢大廈通告 URL

`GET /me/building-notice-url`

Query：

| 欄位 | 類型 | 必填 | 說明 |
| --- | --- | --- | --- |
| `building_id` | string | 否 | 指定大廈；不傳時取目前會員預設或第一個可見大廈。 |

Response：

```json
{
  "building_id": "0348200",
  "building_name": "協和大廈",
  "notice_url": "https://iboard.skylinedances.com/blg_notice/?building_id=0348200",
  "source": "iboard"
}
```

要求：

- 後端必須校驗目前會員可見該 `building_id`。
- `notice_url` 由後端組裝或由舊系統配置讀取，前端只負責打開 URL。
- 沒有可用通告 URL 時，回傳明確錯誤或 `notice_url` 為空並附 `message`，不要返回 mock URL。
- 本介面不建立通知中心記錄，不建立聊天會話。

## 3. 大廈財務

本輪只需要一個介面，一次返回前端三個子頁需要的資料：管理費總覽、財務報表、核數報告。

### 3.1 查詢大廈財務資料

`GET /me/building-finance`

Query：

| 欄位 | 類型 | 必填 | 說明 |
| --- | --- | --- | --- |
| `building_id` | string | 否 | 指定大廈；不傳時取目前會員預設或第一個可見大廈。 |
| `period_from` | string | 否 | 例如 `2025-08`。 |
| `period_to` | string | 否 | 例如 `2025-11`。 |

Response：

```json
{
  "building_id": "0348200",
  "building_name": "協和大廈",
  "management_fee_overview": {
    "periods": ["2025-11", "2025-10", "2025-09", "2025-08_before"],
    "items": [
      {
        "unit_id": "034820001A",
        "unit_display": "01樓 A",
        "period_amounts": [
          {
            "period": "2025-11",
            "amount_hkd": -476,
            "status": "due",
            "status_label": "未繳"
          }
        ]
      }
    ]
  },
  "financial_reports": [
    {
      "report_id": "report_01",
      "title": "YIG 財務報告 2023-03",
      "period": "2023-03",
      "file_url": "https://..."
    }
  ],
  "audit_reports": [
    {
      "report_id": "audit_01",
      "title": "仁英大廈核數報告 2018",
      "period": "2018",
      "file_url": "https://..."
    }
  ]
}
```

說明：

- 若 iSmart `building-info` 已可返回 `documents.mfinreport`、`documents.auditreport`，可先復用資料並轉成上述格式。
- 若舊系統仍由舊後台維護文件，AJO 只做受控讀取。

## 4. 業戶帳目

本輪只需要一個介面，返回目前會員指定大廈/單位的帳目摘要與繳款記錄。

### 4.1 查詢業戶帳目

`GET /me/building-owner-account`

Query：

| 欄位 | 類型 | 必填 | 說明 |
| --- | --- | --- | --- |
| `building_id` | string | 否 | 指定大廈；不傳時取目前會員預設或第一個可見大廈。 |
| `unit_id` | string | 否 | 不傳時取會員預設綁定單位。 |
| `page` | number | 否 | 預設 1。 |
| `page_size` | number | 否 | 預設 20。 |

Response：

```json
{
  "context": {
    "building_id": "0348200",
    "building_name": "協和大廈",
    "unit_id": "034820001A",
    "unit_display": "01樓 A"
  },
  "summary": {
    "outstanding_amount_hkd": 952,
    "last_paid_at": "2026-06-12T10:00:00+08:00"
  },
  "items": [
    {
      "account_item_id": "bill_01",
      "period": "2026-06",
      "title": "管理費",
      "amount_hkd": 476,
      "status": "paid",
      "status_label": "已付",
      "due_at": "2026-06-30T23:59:59+08:00",
      "paid_at": "2026-06-12T10:00:00+08:00",
      "receipt_url": "https://..."
    }
  ]
}
```

說明：

- 可優先由現有 POS `bills`、`history`、`overview` 聚合。
- Staff 的整棟會計資料可繼續使用現有 `GET /me/payments/pos/accounting`，會員端不得看到其他單位資料。

## 5. 意見提供/維修報修

目前前端有完整四步流程：選擇分類、填寫內容、上載圖片/影片、確認提交。本輪只需要兩個介面：

- `GET /me/building-requests`：查詢最近記錄。
- `POST /me/building-requests`：提交維修報修或意見反映。

不再拆分分類、詳情、補充回覆、Staff 狀態流轉介面。分類資料可先由前端常量維持，後端按白名單校驗。

前端目前分類：

| 類型 | 大類 | 次分類 |
| --- | --- | --- |
| 維修報修 | 電力與燈光 | 走廊照明、大堂照明、電制故障、其他燈光 |
| 維修報修 | 結構與門窗 | 門鎖、窗戶、牆身、天花板、其他結構 |
| 維修報修 | 環境與衛生 | 清潔、積水、蟲鼠、其他衛生 |
| 維修報修 | 升降機 | 升降機故障、升降機清潔、其他升降機 |
| 維修報修 | 水務 | 食水、沖廁水、漏水、其他水務 |
| 維修報修 | 其他維修 | 其他 |
| 意見反映 | 環境與衛生 | 清潔建議、環境改善、其他衛生 |
| 意見反映 | 公共設施 | 設施建議、設施損壞、其他設施 |
| 意見反映 | 管理服務 | 管理服務建議、職員表現、其他管理 |
| 意見反映 | 系統與平台功能 | 平台功能、系統問題、其他系統 |
| 意見反映 | 其他意見 | 其他 |

### 5.1 查詢最近記錄

`GET /me/building-requests`

Query：

| 欄位 | 類型 | 必填 | 說明 |
| --- | --- | --- | --- |
| `building_id` | string | 否 | 指定大廈；不傳時取目前會員預設或第一個可見大廈。 |
| `request_type` | string | 否 | `repair` 或 `feedback`。 |
| `page` | number | 否 | 預設 1。 |
| `page_size` | number | 否 | 預設 20。 |

Response item：

```json
{
  "request_id": "br_01",
  "ticket_no": "BR202607010001",
  "request_type": "repair",
  "request_type_label": "維修報修",
  "building_id": "0348200",
  "building_name": "協和大廈",
  "unit_id": "034820001A",
  "unit_display": "01樓 A",
  "subject": "公共走廊照明檢查",
  "category_label": "電力與燈光",
  "subcategory_label": "走廊照明",
  "content": "12樓公共走廊近升降機位置照明不穩。",
  "status": "processing",
  "status_label": "處理中",
  "updated_at": "2026-07-01T10:20:00+08:00",
  "attachment_count": 2
}
```

### 5.2 提交維修報修或意見反映

`POST /me/building-requests`

Payload：

```json
{
  "request_type": "repair",
  "building_id": "0348200",
  "unit_id": "034820001A",
  "category_label": "電力與燈光",
  "subcategory_label": "走廊照明",
  "location_text": "12樓公共走廊近升降機",
  "subject": "公共走廊照明檢查",
  "content": "照明不穩，晚間出現閃爍，請安排檢查燈泡及電制。",
  "contact_name": "陳先生",
  "contact_phone": "+85291234567",
  "asset_ids": ["asset_01", "asset_02"]
}
```

Response：

```json
{
  "request_id": "br_01",
  "ticket_no": "BR202607010001",
  "status": "submitted",
  "status_label": "待跟進",
  "created_at": "2026-07-01T10:20:00+08:00"
}
```

驗證規則：

- `request_type` 必須是 `repair` 或 `feedback`。
- `building_id` 必須在目前會員可見範圍內。
- `unit_id` 若有提交，必須在目前會員可見單位內。
- `category_label`、`subcategory_label` 必須在後端允許白名單內。
- `content` 必填，建議 10 至 2000 字。
- `asset_ids` 只接受目前會員已完成 OSS 上載且未被刪除的資產。

狀態建議：

| 狀態 | 顯示 | 說明 |
| --- | --- | --- |
| `submitted` | 待跟進 | 會員剛提交。 |
| `processing` | 處理中 | 管理處已受理。 |
| `waiting_member` | 等待補充 | 需要會員補資料。 |
| `completed` | 已完成 | 已處理完成。 |
| `rejected` | 不受理 | 管理處拒絕或不屬管理範圍。 |
| `cancelled` | 已取消 | 會員或管理員取消。 |

### 5.3 iSmart 同步建議

若短期仍需把維修/意見同步到 iSmart `blg_cs`，前端仍只呼叫 `POST /me/building-requests`，由後端內部決定是否轉發到既有 `POST /me/ismart/building-comments`。

- AJO 先建立本地 `building_request` 記錄。
- 後端按分類映射 iSmart `comment_type`。
- 成功同步後保存 `external_source = "ismart"`、`external_comment_id`、`external_synced_at`。
- iSmart 同步失敗時，本地記錄仍保留，狀態可標記為 `submitted`，並記錄 `external_sync_error` 供後續排查。

建議映射：

| AJO 類型/大類 | iSmart `comment_type` |
| --- | --- |
| 維修報修 / 電力與燈光 | 電力問題 |
| 維修報修 / 水務 / 食水 | 水質問題 |
| 維修報修 / 水務 / 沖廁水、漏水、其他水務 | 渠務問題 |
| 維修報修 / 環境與衛生 | 清潔衛生 |
| 維修報修 / 結構與門窗 | 其他事宜 |
| 維修報修 / 升降機 | 其他事宜 |
| 意見反映 / 環境與衛生 | 清潔衛生 |
| 意見反映 / 公共設施 | 增加服務 |
| 意見反映 / 管理服務 | 增加服務 |
| 意見反映 / 系統與平台功能 | 其他事宜 |
| 其他 | 其他事宜 |

## 6. 設備監測

本輪只需要一個介面，返回指定大廈的設備狀態列表。

### 6.1 查詢設備狀態

`GET /me/building-equipment-status`

Query：

| 欄位 | 類型 | 必填 | 說明 |
| --- | --- | --- | --- |
| `building_id` | string | 否 | 指定大廈；不傳時取目前會員預設或第一個可見大廈。 |

Response：

```json
{
  "items": [
    {
      "equipment_id": "eq_01",
      "device_name": "升降機 1 號",
      "device_type": "lift",
      "location": "大堂",
      "status": "normal",
      "status_label": "正常",
      "updated_at": "2026-07-01T10:20:00+08:00",
      "latest_event": "巡檢正常"
    }
  ]
}
```

狀態建議：

- `normal`：正常。
- `warning`：需檢查。
- `offline`：離線。
- `maintenance`：維修中。

資料來源可先使用 AJO 自有表或人工維護，後續再接 iCCTV、iLock、升降機或水泵監測來源。

## 7. 授權副戶

本輪只需要一個只讀介面，返回會員目前可見物業下的授權副戶資料。「新增授權」可後續再做，不放入本輪介面。

### 7.1 查詢授權副戶

`GET /me/subaccounts`

Response：

```json
{
  "items": [
    {
      "grant_id": "grant_01",
      "building_id": "0348200",
      "building_name": "協和大廈",
      "unit_id": "034820001A",
      "unit_display": "01樓 A",
      "target_user_id": 123,
      "target_display_name": "住客帳戶",
      "relationship": "tenant",
      "permission_scope": "readonly",
      "status": "active",
      "created_at": "2026-07-01T10:20:00+08:00"
    }
  ]
}
```

說明：

- 會員只能看到自己有權查看的單位授權資料。
- 初期只需要 `readonly` 顯示，不做新增、撤銷或審批。

## 8. 物業綁定

本輪只需要一個提交介面，用於提交物業綁定申請。既有已綁定物業可由 `GET /me/buildings` 顯示，不在本介面內重複做列表。

### 8.1 提交物業綁定申請

`POST /me/property-bindings`

Payload：

```json
{
  "platform": "pm",
  "building_id": "0348200",
  "floor": "01",
  "unit": "A",
  "applicant_name": "陳先生",
  "applicant_phone": "+85291234567",
  "applicant_email": "user@example.com",
  "identity_type": "owner",
  "asset_ids": ["asset_01", "asset_02"],
  "note": "申請綁定單位。"
}
```

Response：

```json
{
  "application_id": "pb_01",
  "building_id": "0348200",
  "building_name": "協和大廈",
  "unit_display": "01樓 A",
  "status": "submitted",
  "status_label": "已提交",
  "submitted_at": "2026-07-01T10:20:00+08:00"
}
```

狀態建議：

- `submitted`：已提交。
- `reviewing`：審批中。
- `approved`：已綁定。
- `rejected`：已拒絕。
- `cancelled`：已取消。

## 9. 建議資料表

只列當前功能必需的表，避免過度抽象。

| 表 | 用途 |
| --- | --- |
| `building_notice_links` | 可選；保存大廈對應通告 URL。如 URL 可由規則組裝，則不需要建表。 |
| `building_requests` | 維修報修與意見反映主表。 |
| `building_request_attachments` | 工單附件關聯 OSS asset。 |
| `property_binding_applications` | 物業綁定申請。 |
| `subaccount_grants` | 授權副戶關係。 |
| `building_equipment_statuses` | 設備狀態快照。 |

## 10. 驗收標準

- 會員只能看到自己可見大廈與單位資料。
- 最新通告只提供 `GET /me/building-notice-url`，前端拿到 `notice_url` 後直接打開。
- 大廈財務、業戶帳目、設備監測各只提供一個查詢介面。
- 維修報修/意見提供只提供 `GET /me/building-requests` 與 `POST /me/building-requests` 兩個介面。
- 授權副戶本輪只提供 `GET /me/subaccounts` 只讀列表。
- 物業綁定本輪只提供 `POST /me/property-bindings` 提交申請。
- 所有附件必須走 OSS asset，不接受前端直接提交任意外部 URL。
- 有列表返回的介面需支援分頁或合理限制筆數，避免一次返回整棟大廈全部資料。
