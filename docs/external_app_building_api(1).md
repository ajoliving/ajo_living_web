# External App Building APIs

## Overview

This document describes the external-facing APIs added to replace parts of the resident web pages:

- `building_access`
- `blg_cs`
- merged `blg_form` + `blg_info`

These APIs are intended for trusted server-to-server or app-backend integrations.

## Security

- All endpoints are protected by `@ip_whitelist('EXTERNAL_APP_ALLOWED_IPS')`
- Allowed IPs are configured in `stc/settings.py`
- There is no session-based authentication on these endpoints
- The caller must provide explicit identifiers such as `user_id`, `building_id`, `door_id`, or `qrcode_record_id`

## Environment
- production : `https://ismart.ajoliving.com`
- clouddev : `https://clouddev.ismart.ajoliving.com`

some door access features may not work in clouddev because the door controller is not connected to the clouddev environment

## Common Conventions

- **Base path**: `/api/v1/external/`
- **Method**: `POST`
- **Content-Type**: `application/json`
- **Response format**:
  - success: `{"status": "success", "data": ...}`
  - error: `{"status": "error", "message": "..."}`

## Error Handling

Common HTTP status codes: 

- `200` request succeeded
- `201` resource created successfully
- `400` invalid request body or missing required field
- `403` caller IP not whitelisted or user not allowed to access the requested resource
- `404` referenced record does not exist
- `500` unexpected server-side error
- `502` upstream device/service failed, such as door controller or QR payload generator

## 1. Merged Building Info + Forms

Returns the building profile from `showdata.BuildingInfo` together with related files from `ismart.Blgfiles` for:

- `btype = form`
- `btype = blginfo`

### Endpoint

`POST /api/v1/external/building-info/`

### Request Body

```json
{
  "building_id": "0348200"
}
```

### Request Fields

- `building_id` string, required, target building ID

### Success Response

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
      "forms": [
        {
          "id": 1,
          "title": "裝修申請表",
          "type": "form",
          "file_url": "https://...",
          "file_date": "2026-06-01",
          "file_month": null,
          "created_date": "2026-06-14"
        }
      ],
      "building_info_files": [
        {
          "id": 2,
          "title": "管理處資料",
          "type": "blginfo",
          "file_url": "https://...",
          "file_date": "2026-06-01",
          "file_month": null,
          "created_date": "2026-06-14"
        }
      ]
    }
  }
}
```

### Building File Types (`btype`)

The `btype` field on `Blgfiles` (model `ismart.Blgfiles`) determines which document category a file belongs to. The currently defined types are:

| `btype` value | Category | Returned in |
|---|---|---|
| `form` | 表單文件 (Forms) | `documents.forms` |
| `blginfo` | 大廈資料文件 (Building Info Files) | `documents.building_info_files` |
| `floorplan` | 平面圖 (Floor Plans) | `documents.floorplan` |
| `auditreport` | 審計報告 (Audit Reports) | `documents.auditreport` |
| `mfinreport` | 財務報告 (Financial Reports) | `documents.mfinreport` |

### Notes

- `building_info` is returned even if there are no documents
- `documents.forms` and `documents.building_info_files` are empty arrays when no files exist
- `floorplan`, `auditreport`, and `mfinreport` are defined on the model but **not currently returned** by this combined endpoint; they are handled by separate staff management views

## 2. Building Comment Submission (`blg_cs`)

Creates a `ismart.BlgComment` record and sends the existing notification email.

### Endpoint

`POST /api/v1/external/blg-cs/submit/`

### Request Body

```json
{
  "user_id": 123,
  "building_id": "0348200",
  "comment_type": "其他事宜",
  "comment": "Air-conditioner water leakage at the corridor outside 18/F."
}
```

### Request Fields

- `user_id` integer, required, `accounts.CustomUser.id`
- `building_id` string, required, target building ID
- `comment_type` string, required, must be one of the current `BlgComment.COMCHOICES`
- `comment` string, required, comment content

### Allowed `comment_type` Values

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

### Success Response

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

### Permission Rule

- The supplied `user_id` must be able to access the supplied `building_id`
- Access is resolved using the same user-to-building relationship pattern as the current page logic

## 3. Building Access Summary

Returns the door list, password information, QR availability, and recent access records for a user in a building.

### Endpoint

`POST /api/v1/external/building-access/`

### Request Body

```json
{
  "user_id": 1,
  "building_id": "0999900"
}
```

### Request Fields

- `user_id` integer, required, `accounts.CustomUser.id`
- `building_id` string, required, target building ID

### Success Response

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

### Notes

- `has_permission` indicates whether the user currently has door permission
- `password` is `null` when no password record exists
- `qrcode` is `null` when QR access is not enabled or not provisioned
- `camera` is `null` when no linked camera exists
- `requested_building_id` and `access_building_id` may differ because the legacy page contains a hardcoded building remap for specific building IDs

## 4. Open Door

Triggers the same remote-open flow used by the web page and records the result.

### Endpoint

`POST /api/v1/external/building-access/open-door/`

### Request Body

```json
{
  "user_id": 123,
  "building_id": "0348200",
  "door_id": 10
}
```

### Request Fields

- `user_id` integer, required, `accounts.CustomUser.id`
- `door_id` integer, required, `smartliving.Door.id`
- `building_id` string, optional but recommended, used as an extra validation check

### Success Response

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

### Upstream Failure Response Example

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

### Permission Rule

- The user must be allowed to access the target building
- The user must also have a related flat/client permission that authorizes door open under the existing project logic

## 5. Generate Door QR Payload

Generates a QR payload string for a QR-enabled door credential. The external app can render this payload as a QR code on the client side.

### Endpoint

`POST /api/v1/external/building-access/qrcode/`

### Request Body

```json
{
  "user_id": 123,
  "qrcode_record_id": 77,
  "term": "dynamic"
}
```

### Request Fields

- `user_id` integer, required, `accounts.CustomUser.id`
- `qrcode_record_id` integer, required, `smartliving.DoorPasswordRecord.id` where `passwordorqrcode=True`
- `term` string, optional, allowed values:
  - `dynamic` default, valid for 15 minutes
  - `static` valid for up to 360 days, capped by the credential end time

### Success Response

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

### Client Rendering Recommendation

Render `qrcode_value` as a QR code image on the client side using your preferred QR library.

## 6. POS Payment to iSmart

Creates a payment from a POS terminal or external payment system. This endpoint calls the same backend logic as the production POS integration.

- **IP whitelist**: `POS_PAYMENT_ALLOWED_IPS` (separate from `EXTERNAL_APP_ALLOWED_IPS`)
- The `@ip_whitelist` decorator exists on the view class but is **currently commented out** — IP restriction is not enforced unless you uncomment it.

### Endpoint

`POST /api/v1/pos-payment-to-ismart`

### Request Body

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

### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `BLG_ID` | string | **required** | Building identifier |
| `PAY_METHOD` | string | **required** | One of the supported pay methods (see below) |
| `FINAL_AMOUNT` | integer | **required** for cash/cheque/bank/web POS | Amount in **cents** (e.g. 150000 = HKD 1500.00) |
| `AMOUNT` | integer | **required** for `allinpay` | Amount in **cents** |
| `ENTRY_DATETIME` | string | **required** for cash/cheque/bank/web POS | Local entry datetime (`YYYY-MM-DD HH:mm:ss`) |
| `TRAN_DATETIME` | string | **required** for cash/cheque/bank/web POS | Transaction datetime (`YYYY-MM-DD HH:mm:ss`) |
| `DATE` + `TIME` | string | **required** for `allinpay` | Date (`YYYYMMDD`) and time (`HHmmss`) concatenated |
| `TRAN_REF_NO` | string | optional | External reference / trace number |
| `TRANS_TRACE_NO` | string | **required** for `allinpay` | Allinpay trace number |
| `TRANS_TICKET_NO` | string | **required** for `allinpay` | Allinpay ticket number |
| `USER_ID` | integer | optional, default `1` | The `CustomUser.id` submitting the payment |
| `FAKE_TRANSACTION` | bool | optional, default `false` | When `true`, status is set to `pending_validation` (for bank transfers) |
| `COMMENT` | string | optional | Free-text comment |
| `PIC_FILENAME` | array[string] | optional | List of filenames for attached images |
| `PIC_DATA` | array[string] | optional | List of base64-encoded image data (matching `PIC_FILENAME` by index) |
| `bank_account_received` | string | optional | Bank account number (for bank-in grouping) |
| `PAYMENT_GATEWAY_RESPONSE` | object | optional | Raw gateway response payload to store |
| `BILL_OBJS` | array[object] | **required** | List of billed items (see below) |

### `BILL_OBJS` Item Fields

| Field | Type | Required | Description                                                                                                  |
|---|---|---|--------------------------------------------------------------------------------------------------------------|
| `flat_code` | string | **required** | Unit ID (e.g. `0348200001`)                                                                                  |
| `invoice_no` | string | optional | Invoice number. **Omit or set empty for prepaid (預付)**. When `trs_to` is `預付`, `invoice_no` is forced empty. |
| `item_id` | string | **required** | Fee item type (e.g. `管理費`, `分攤費用`, ·其他費用·)                                                                   |
| `trs_to` | string | **required** | Period (e.g. `2026/06`) or `預付` for prepaid items                                                            |
| `net_amount` | integer | **required** | Amount in **cents** (positive integer). For prepaid items this is the deposit amount.                        |
| `transfer_fee` | integer | optional, default `0` | Handling/transfer fee in **cents**                                                                           |

### Supported `PAY_METHOD` Values

| Value | Status | Notes |
|---|---|---|
| `POS_CASH` | `in_cashier` | Cash payment at POS |
| `POS_CHEQUE` | `in_cashier` | Cheque payment at POS |
| `POS_BANK` | `pending_validation` | Bank transfer at POS |
| `POS_ALIWE` | `in_cashier` | Alipay/WeChat at POS |
| `WEBPOS_ALIPAY` | `payment_captured` | Online Alipay — auto-creates `BalCalcTbl` and `PrepaidBal` immediately |
| `WEBPOS_WECHAT` | `payment_captured` | Online WeChat — auto-creates `BalCalcTbl` and `PrepaidBal` immediately |
| `WEBPOS_CARD_UP` | `payment_captured` | Online credit card — auto-creates `BalCalcTbl` and `PrepaidBal` immediately |
| `allinpay` | `init` | Allinpay gateway — requires `DATE`, `TIME`, `TRANS_TRACE_NO`, `TRANS_TICKET_NO` |

### Validation Rules

1. **Invoice amounts**: For non-prepay items, the `net_amount` must not exceed the current outstanding (unpaid) balance for that invoice. Returns `400` if exceeded.
2. **Building existence**: `BLG_ID` must reference a valid `IBuildingTbl`. Returns `404` if not found.
3. **Payment type**: `PAY_METHOD` must be configured as a `BuildingPaymentType` for the building. Returns `404` if not configured.

### Status Lifecycle

See "Payment Status Lifecycle" in the project AGENTS.md for the full status diagram. In summary:

- `in_cashier` → staff bank-in → `pending_validation` → staff validate → `validated_by_ismart`
- `pending_validation` → staff validate → `validated_by_ismart`
- `payment_captured` → `BalCalcTbl` entries and `PrepaidBal` records created immediately; status remains `payment_captured`
- `init` → `allinpay` only; transitions to a confirmed state on gateway callback

### Success Response

```json
{
  "code": "200",
  "message": "success",
  "receipt_id": "1000042"
}
```

### Error Response Examples

```json
{
  "code": "400",
  "message": "Payment amount 1500.00 exceeds unpaid amount 1400.00 for invoice INV000123"
}
```

```json
{
  "code": "404",
  "message": "Building with ID 0348200 not found"
}
```

```json
{
  "code": "400",
  "message": "Missing required field: PAY_METHOD"
}
```

### Notes

- All monetary values use **cents** (integer) in the request. The server converts to dollars internally.
- For `payment_captured` flows: `BalCalcTbl` rows are created with positive `trs_val` for payment entries, and invoices are marked `fullypaid` when `SUM(trs_val) >= 0`.
- Handling fees (`transfer_fee`) generate separate `BalCalcTbl` offsetting pairs.
- Prepaid items (`trs_to = 預付`) create `PrepaidBal` records linked to the `Paymentdetails` row.
- The endpoint is CSRF-exempt (`@method_decorator(csrf_exempt, name='dispatch')`).

## Example cURL Requests

### Get merged building info

```bash
curl -X POST "https://<your-domain>/api/v1/external/building-info/" \
  -H "Content-Type: application/json" \
  -d '{
    "building_id": "0348200"
  }'
```

### Submit a building comment

```bash
curl -X POST "https://<your-domain>/api/v1/external/blg-cs/submit/" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "building_id": "0348200",
    "comment_type": "其他事宜",
    "comment": "Please inspect the corridor lighting on 18/F."
  }'
```

### Get building access data

```bash
curl -X POST "https://<your-domain>/api/v1/external/building-access/" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "building_id": "0348200"
  }'
```

### Open a door

```bash
curl -X POST "https://<your-domain>/api/v1/external/building-access/open-door/" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "building_id": "0348200",
    "door_id": 10
  }'
```

### Get QR payload

```bash
curl -X POST "https://<your-domain>/api/v1/external/building-access/qrcode/" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "qrcode_record_id": 77,
    "term": "dynamic"
  }'
```

### POS payment

```bash
curl -X POST "https://<your-domain>/api/v1/pos-payment-to-ismart" \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

## Deployment Checklist

- Add the external app server IPs to `EXTERNAL_APP_ALLOWED_IPS`
- Confirm the external app sends `application/json`
- Confirm the external app can render QR payloads client-side
- Verify the app stores and protects door password data appropriately
- Test door open and QR generation against the target environment hardware
