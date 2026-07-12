# iSmart Integration API Reference

## Purpose

This document is the current integration reference for the programmatic endpoints exposed by the iSmart Django application.

It covers:

- the new primary integration endpoints under `/api/v1/integration/...`
- the existing legacy paths that remain available for backward compatibility
- POS and payment workflow APIs
- webhook-style machine endpoints

This is a code-driven document based on the endpoints currently registered in the project URL configuration.

## Environments

- production: `https://ismart.ajoliving.com`
- clouddev: `https://clouddev.ismart.ajoliving.com`

Some smartliving and door-controller features may not function in `clouddev` because hardware connectivity is environment-dependent.

## Path Strategy

Primary integration routes now use:

- `/api/v1/integration/...`

Legacy routes are still kept and continue to work:

- old `/api/v1/...` paths
- old `/api/v1/external/...` paths
- old webhook paths outside `/api/v1`

Clients should migrate to the new integration paths, but existing clients do not need to change immediately.

## Current Endpoint Inventory

| Group | Method | Primary Path | Legacy Path | Purpose |
|---|---|---|---|---|
| Auth | `POST` | `/api/v1/integration/auth/login/` | `/api/v1/pos/login` | Credential check for POS or external clients |
| Auth | `POST` | `/api/v1/integration/auth/register/` | — | Create an iSmart `CustomUser` account directly, no approval |
| Building data | `GET` | `/api/v1/integration/buildings/receivables/management-fees/` | `/api/v1/building-mf-table/` | Management-fee receivable summary |
| Building data | `GET` | `/api/v1/integration/buildings/receivables/other-fees/` | `/api/v1/building-of-list/` | Other-fee receivable list |
| Building data | `GET` | `/api/v1/integration/buildings/notices/` | `/api/v1/building-notices/` | Active building notices |
| External building | `GET` | `/api/v1/integration/buildings/info/` | `/api/v1/external/building-info/` | Building profile plus document metadata |
| Member role binding | `POST` | `/api/v1/integration/buildings/building-flat-owner-binding-requests/` | — | Submit an `OwnerReg` request for owner-role binding, staff approval required |
| Member subaccount | `GET` | `/api/v1/integration/buildings/subaccounts/` | — | List current active authorized sub users for owner-controlled unit(s) |
| Member subaccount | `POST` | `/api/v1/integration/buildings/subaccounts/grant/` | — | Grant `授權用戶` to another `CustomUser` for one owner-controlled unit |
| Member subaccount | `POST` | `/api/v1/integration/buildings/subaccounts/revoke/` | — | Revoke `授權用戶` from one owner-controlled unit |
| External building | `POST` | `/api/v1/integration/buildings/comments/` | `/api/v1/external/blg-cs/submit/` | Submit building comment / complaint |
| External door access | `POST` | `/api/v1/integration/access/buildings/` | `/api/v1/external/building-access/` | Door list, password, QR info, recent records |
| External door access | `POST` | `/api/v1/integration/access/open-door/` | `/api/v1/external/building-access/open-door/` | Remote open a door |
| External door access | `POST` | `/api/v1/integration/access/qrcode/` | `/api/v1/external/building-access/qrcode/` | Generate QR payload for door access |
| Payment | `GET` | `/api/v1/integration/payments/unpaid-invoices/` | `/api/v1/building-flat-unpaid-invoice-list` | Outstanding invoices for one unit |
| Payment | `POST` | `/api/v1/integration/payments/pos/` | `/api/v1/pos-payment-to-ismart` | Create payment from POS or external payment source |
| Cashier | `GET` | `/api/v1/integration/cashier/transactions/` | `/api/v1/get-transactions-in-cashier` | List current cashier payments for a building |
| Payment | `GET` | `/api/v1/integration/payments/transactions/by-unit/` | `/api/v1/get-transactions-by-flat-unit` | List payment history for one or more units |
| Payment | `GET` | `/api/v1/integration/payments/transactions/by-date/` | `/api/v1/get-transactions-by-date` | List building payments by date range |
| Cashier | `POST` | `/api/v1/integration/cashier/bank-in/` | `/api/v1/update-transactions-status-in-cashier` | Bank-in cashier payments and move them to pending validation |
| Cashier | `GET` | `/api/v1/integration/cashier/bank-in-records/` | `/api/v1/get-bank-in-record-list` | List recent bank-in batches |
| Cashier | `GET` | `/api/v1/integration/cashier/bank-in-records/details/` | `/api/v1/get-bank-in-record-details` | List payment rows inside one bank-in batch |
| Webhook | `POST` | `/api/v1/integration/webhooks/qfpay/` | `/qfpayapi/` | QFPay callback endpoint |
| Debug webhook | `POST` | `/api/v1/integration/webhooks/test/` | `/test_webhook/` | Test-only webhook capture endpoint |

## Security Model

These endpoints do not share a single authentication scheme. Integration callers must understand the current behavior before connecting.

### IP Whitelist Protected

The following endpoints are protected by `@ip_whitelist('EXTERNAL_APP_ALLOWED_IPS')`:

- `/api/v1/external/building-info/`
- `/api/v1/external/blg-cs/submit/`
- `/api/v1/external/building-access/`
- `/api/v1/external/building-access/open-door/`
- `/api/v1/external/building-access/qrcode/`

Important:

- the whitelist is configured in `stc/settings.py`
- if `EXTERNAL_APP_ALLOWED_IPS` is empty, the decorator effectively allows all IPs
- these endpoints do not use session auth or token auth
- user identity is supplied in the JSON body, usually as `user_id`

### Not IP Whitelisted

The following currently do not use `@ip_whitelist`:

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

Important:

- `PosPaymentToIsmart` has a commented-out `POS_PAYMENT_ALLOWED_IPS` decorator in code
- `qfpayapi` only checks whether header `X-QF-SIGN` exists; it does not validate the signature
- `test_webhook/` is a debug endpoint and should not be treated as a production integration API

## Shared Request Conventions

- method depends on endpoint
- integration read-only endpoints now use `GET` with query parameters
- legacy aliases remain available as `POST`, usually with JSON request bodies
- content type: `application/json` for `POST` APIs, except webhook senders may use their own payload shape
- CSRF: all documented `api/v1` endpoints in this file are CSRF-exempt
- timezone: several payment endpoints manually convert timestamps to Hong Kong time by adding or subtracting 8 hours

## Shared Response Conventions

Response formats are not fully standardized.

Current patterns in use:

- wrapped success: `{"status":"success","data":...}`
- wrapped error: `{"status":"error","message":"..."}`
- legacy success: `{"code":"200","message":"success",...}`
- POS login success: `{"code":"1","msg":...}`
- raw array success: `[...]`
- plain text webhook response: `SUCCESS`, `UNSUCCESS`, or `success`

Clients should parse each endpoint according to its own contract rather than assuming a global response envelope.

## Error Codes

Common HTTP status codes currently used:

- `200` successful request
- `201` resource created
- `400` bad request, invalid JSON, missing field, or invalid field value
- `401` invalid credentials
- `403` IP not whitelisted or user not authorized for the requested resource
- `404` referenced user, building, door, QR record, or payment type not found
- `405` non-POST request on webhook endpoints
- `500` unhandled server error
- `502` upstream device or QR generator failure

## 1. Authentication

### 1.1 POS / External Login

- endpoint: `POST /api/v1/integration/auth/login/`
- legacy path: `POST /api/v1/pos/login`
- purpose: verifies credentials and returns the user identity plus building and unit permission lists
- auth model: no token is issued; this is a credential-check endpoint only

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
| `login_name` | string | yes | Username, email, or phone depending on `login_type` |
| `password` | string | yes | User password |
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

- `building` and `staff_building_permissions` currently return the same data
- the endpoint does not create a session, JWT, or API token

### 1.2 Direct iSmart Account Registration

- endpoint: `POST /api/v1/integration/auth/register/`
- legacy path: none
- purpose: directly creates a new `accounts.CustomUser` and linked `ClientTbl` profile with no approval flow
- access control: currently no IP whitelist and no user authentication

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
| `phone` | string | yes | Primary phone, same as `memberphone` in the legacy member registration flow |
| `email` | string | yes | Primary email |
| `eng_name` | string | yes | English display name |
| `chi_name` | string | no | Chinese name |
| `legal_entity` | string | no | `NA` for individual, `LE` for legal entity. Default: `NA` |
| `id_card` | string | no | ID card / document number |
| `remark` | string | no | Free-text note saved to `ClientTbl.cli_remark` |
| `gender` | string | no | `M`, `F`, or empty |
| `is_receive_email` | boolean | no | Whether to enable `UserSettings.blg_notice_email`. Default: `true` |
| `password` | string | no | If omitted, the backend generates a random password |

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

- if phone or email already belongs to an existing `CustomUser`, the API returns HTTP `409`
- if a matching `ClientTbl` already exists by phone, email, or `id_card`, its `cli_id` is reused as the new username
- no `OwnerReg` row is created by this endpoint

## 2. Building Data APIs

### 2.1 Building Management Fee Table

- endpoint: `GET /api/v1/integration/buildings/receivables/management-fees/`
- legacy path: `POST /api/v1/building-mf-table/`
- purpose: returns management-fee receivable data generated from `blg_mf_list(...)`
- access control: no IP whitelist and no user authentication

#### Integration Query Parameters

```text
?building_id=0348200&data_structure=table
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | Building ID |
| `data_structure` | string | no | Response layout. Common values: `table`, `list`, `block_dict`. Default: `table` |

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

- response shape changes with `data_structure`
- values may be amounts or status labels such as `已付`, `審核中`, or `錯誤`
- the backend calls the receivable generator with `group_estate=True`

### 2.2 Building Other-Fee List

- endpoint: `GET /api/v1/integration/buildings/receivables/other-fees/`
- legacy path: `POST /api/v1/building-of-list/`
- purpose: returns non-management-fee receivable data generated from `blg_mf_list(...)`
- access control: no IP whitelist and no user authentication

#### Integration Query Parameters

```text
?building_id=0348200&data_structure=list
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | Building ID |
| `data_structure` | string | no | Common values: `list` or `block_dict`. Default: `list` |

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

- response shape changes with `data_structure`
- empty result is returned as `[]`

### 2.3 Building Notices

- endpoint: `GET /api/v1/integration/buildings/notices/`
- legacy path: `POST /api/v1/building-notices/`
- purpose: returns active notices for one building
- access control: no IP whitelist and no user authentication

#### Integration Query Parameters

```text
?building_id=0348200
```

#### Legacy Request

```json
{
  "blg_id": "0348200"
}
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

- response is a raw array, not an object wrapper
- only notices where `mess_date <= today < mess_down` are returned

## 3. External Building APIs

These endpoints are intended for trusted server-to-server or app-backend integrations.

### 3.0 Member Role Binding

This section covers member-to-unit binding flows implemented in this session.

- owner-role binding (`登記業主`) is submitted through `OwnerReg` and still requires building-management approval
- authorized sub user binding (`授權用戶`) does not use approval when initiated by an already-approved owner
- the actual active role relationship is still stored in `i_flatcli_rel_tbl`
- a separate history model now records authorized sub user grant / revoke operations

### 3.0.1 Owner Role Binding Request

- endpoint: `POST /api/v1/integration/buildings/building-flat-owner-binding-requests/`
- legacy path: none
- purpose: submits an `OwnerReg` request for binding a user to one or more units, mainly for the owner-role approval flow
- access control: currently no IP whitelist and no user authentication

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

- this endpoint creates an `OwnerReg` row only; no role relationship is granted immediately
- staff approval later runs through the existing `confirm_memreg` flow
- if `OwnerReg.user` is present, approval reuses the existing account instead of creating a new account

### 3.0.2 List Authorized Sub Users

- endpoint: `GET /api/v1/integration/buildings/subaccounts/`
- legacy path: none
- purpose: lists active `授權用戶` relationships for unit(s) where the caller is already the approved `登記業主`
- access control: caller must be an active `登記業主` of the requested unit; if `unit_id` is omitted, all owner-controlled units are included

#### Integration Query Parameters

```text
?user_id=123&unit_id=0348200001
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | `accounts.CustomUser.id` of the owner |
| `unit_id` | string | no | Optional unit filter. If omitted, returns all active subaccounts for all units owned by the user |

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

- this endpoint returns only active `i_flatcli_rel_tbl` rows where `cli_role='授權用戶'`
- `history_count` counts rows in the dedicated authorized-sub-user history model for the relation
- if `unit_id` is provided and the caller is not the approved owner, the API returns HTTP `403`

### 3.0.3 Grant Authorized Sub User

- endpoint: `POST /api/v1/integration/buildings/subaccounts/grant/`
- legacy path: none
- purpose: lets an already-approved owner grant `授權用戶` to another existing `CustomUser` for one unit, with no staff approval
- access control: caller must already be an active `登記業主` of the requested unit

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
| `user_id` | integer | yes | Owner `accounts.CustomUser.id` |
| `unit_id` | string | yes | Target unit ID |
| `target_user_id` | integer | yes | Existing `accounts.CustomUser.id` to become `授權用戶` |
| `remark` | string | no | Free-text audit note stored on both relation and history |

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

- `user_id` must currently be an active `登記業主` of the supplied `unit_id`
- `target_user_id` must already exist and have a linked `ClientTbl`
- `target_user_id` cannot equal `user_id`
- each unit can have at most 4 active `授權用戶`
- if the target already has an active `授權用戶` relation for the unit, the API returns HTTP `409`

#### Notes

- if the same target user had an old inactive `授權用戶` row, the current code reactivates that row instead of creating a second one
- each successful grant writes one row into `AuthorizedSubUserRequestHistory`

### 3.0.4 Revoke Authorized Sub User

- endpoint: `POST /api/v1/integration/buildings/subaccounts/revoke/`
- legacy path: none
- purpose: lets an owner deactivate an active `授權用戶` relationship for one unit
- access control: caller must already be an active `登記業主` of the requested unit

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

- revoke is implemented as `is_active=False` on the existing `i_flatcli_rel_tbl` row
- each successful revoke writes one row into `AuthorizedSubUserRequestHistory`
- revoking a user who is not currently an active `授權用戶` of the unit returns HTTP `404`

### 3.1 Building Info and Documents

- endpoint: `GET /api/v1/integration/buildings/info/`
- legacy path: `POST /api/v1/external/building-info/`
- purpose: returns building profile, building info fields, and building file metadata
- access control: `EXTERNAL_APP_ALLOWED_IPS`

#### Integration Query Parameters

```text
?building_id=0348200
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | Target building ID |

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
| `documents.audit_reports` | `auditreport` / `audition` |
| `documents.financial_reports` | `mfinreport` |

#### Notes

- arrays are returned even when no files exist
- `building_info` is returned with `null` or empty-string values when no `BuildingInfo` row exists
- AJO member proxy normalizes `auditreport` / `audition` / `audit_report` / `auditreports` / `auditions` into `documents.audit_reports`; the upstream endpoint must still return one of these groups

### 3.2 Submit Building Comment

- endpoint: `POST /api/v1/integration/buildings/comments/`
- legacy path: `POST /api/v1/external/blg-cs/submit/`
- purpose: creates `BlgComment` and sends the existing notification email flow
- access control: `EXTERNAL_APP_ALLOWED_IPS` plus user-to-building authorization

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
| `building_id` | string | yes | Target building ID |
| `comment_type` | string | yes | Must be one of the allowed comment types |
| `comment` | string | yes | Comment body |

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

- the supplied `user_id` must be authorized for the supplied `building_id`
- success status code is `201`
- email sending happens inside the request flow

## 4. External Door Access APIs

### 4.1 Building Access Summary

- endpoint: `POST /api/v1/integration/access/buildings/`
- legacy path: `POST /api/v1/external/building-access/`
- purpose: returns public door list, camera reference, password info, QR info, and recent access records for a user in a building
- access control: `EXTERNAL_APP_ALLOWED_IPS` plus user-to-building authorization

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

- `has_permission` reflects `DoorPermission`
- `password`, `qrcode`, and `camera` can be `null`
- two legacy building remaps are applied in current code:
  - `0348300` is remapped to `0348200`
  - `0324900` is remapped to `0325000`
- `requested_building_id` and `access_building_id` can therefore differ

### 4.2 Open Door

- endpoint: `POST /api/v1/integration/access/open-door/`
- legacy path: `POST /api/v1/external/building-access/open-door/`
- purpose: triggers the same remote-open flow used by the smartliving web pages
- access control: `EXTERNAL_APP_ALLOWED_IPS` plus user-to-building authorization

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
| `building_id` | string | no | Optional extra validation that the door belongs to the building |

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

- failure to reach the door controller returns HTTP `502`
- current open-door permission logic is not identical to the listing endpoint:
  - listing checks `DoorPermission`
  - open-door checks whether the user has any active flat-client relationship in the building

### 4.3 Generate Door QR Payload

- endpoint: `POST /api/v1/integration/access/qrcode/`
- legacy path: `POST /api/v1/external/building-access/qrcode/`
- purpose: generates a QR payload string that the client can render as a QR image
- access control: `EXTERNAL_APP_ALLOWED_IPS` plus user ownership of the QR record

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
| `qrcode_record_id` | integer | yes | `DoorPasswordRecord.id` where `passwordorqrcode=True` |
| `term` | string | no | `dynamic` or `static`. Default: `dynamic` |

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

- `dynamic`: expires at current time plus 15 minutes
- `static`: expires at current time plus 360 days, capped by the record `end_time`

#### Notes

- the API returns payload text only; the client must render the QR image
- current code does not separately reject already-expired QR records before generation

## 5. Payment APIs

### 5.1 Unit Outstanding Invoice List

- endpoint: `GET /api/v1/integration/payments/unpaid-invoices/`
- legacy path: `POST /api/v1/building-flat-unpaid-invoice-list`
- purpose: returns currently unpaid invoice rows for one unit, net of payments already pending in cashier or validation
- access control: no IP whitelist and no user authentication

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

- response is a raw array
- `net_amount` is returned in dollars, not cents
- pending payments with status `in_cashier` or `pending_validation` are deducted from the outstanding amount

### 5.2 POS Payment to iSmart

- endpoint: `POST /api/v1/integration/payments/pos/`
- legacy path: `POST /api/v1/pos-payment-to-ismart`
- purpose: creates a `Payment` and `Paymentdetails`, and for selected payment methods also creates `BalCalcTbl` and `PrepaidBal` rows immediately
- access control: no active IP whitelist in current code

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
| `PIC_FILENAME` | string array | no | Uploaded image filenames |
| `PIC_DATA` | string array | no | Base64 image payloads matching `PIC_FILENAME` |
| `bank_account_received` | string | no | Target bank account number |
| `PAYMENT_GATEWAY_RESPONSE` | object | no | Raw gateway response to store |
| `BILL_OBJS` | object array | yes | Payment line items |

#### `BILL_OBJS` Item Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `flat_code` | string | yes | Unit ID |
| `invoice_no` | string | no | Invoice number. Empty or omitted for prepaid |
| `item_id` | string | yes | Fee item name |
| `trs_to` | string | yes | Billing period or `預付` |
| `net_amount` | integer | yes | Amount in cents |
| `transfer_fee` | integer | no | Handling fee in cents. Default: `0` |

#### Supported `PAY_METHOD` Values

| Value | Initial Status | Notes |
|---|---|---|
| `POS_CASH` | `in_cashier` | Cash payment |
| `POS_CHEQUE` | `in_cashier` | Cheque payment |
| `POS_BANK` | `pending_validation` | Bank transfer |
| `POS_ALIWE` | `in_cashier` | Alipay / WeChat on POS |
| `WEBPOS_ALIPAY` | `payment_captured` | Immediate accounting writeback |
| `WEBPOS_WECHAT` | `payment_captured` | Immediate accounting writeback |
| `WEBPOS_CARD_UP` | `payment_captured` | Immediate accounting writeback |
| `allinpay` | `init` | Special legacy flow |

#### Success Response

```json
{
  "code": "200",
  "message": "success",
  "receipt_id": "1000042"
}
```

#### Validation Rules

- building must exist
- `PAY_METHOD` must be configured as a `BuildingPaymentType` for the building
- for non-prepaid lines, requested payment amount cannot exceed current unpaid amount
- unsupported pay method returns `400`

#### Important Behavior

- all request money values are in cents
- created `Payment.txamount` and line amounts are stored internally in dollars
- `WEBPOS_*` methods immediately create payment-side accounting rows
- prepaid lines with `trs_to = "預付"` create `PrepaidBal`
- handling fees create separate detail and accounting rows for `平台手續費`

#### Important Caveats

- the whitelist decorator for this endpoint is currently commented out
- unknown `bank_account_received` is silently ignored rather than rejected
- the `allinpay` datetime parsing path is legacy and should be tested carefully before use

### 5.3 Transactions in Cashier

- endpoint: `GET /api/v1/integration/cashier/transactions/`
- legacy path: `POST /api/v1/get-transactions-in-cashier`
- purpose: lists current cashier transactions for one building

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

- only status `in_cashier` is returned
- results are split into `payment_objs_cheque` and `payment_objs_cash`
- endpoint does not return a unified list

### 5.4 Transactions by Flat Unit

- endpoint: `GET /api/v1/integration/payments/transactions/by-unit/`
- legacy path: `POST /api/v1/get-transactions-by-flat-unit`
- purpose: lists payment history for one or more units

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

- `init` and `pending_validation` are excluded
- `validated_by_ismart`, `payment_captured`, and `confirmed` are normalized to `confirmed`
- prepaid consumption may appear as `"{trs_to}(預付)"`

### 5.5 Transactions by Date

- endpoint: `GET /api/v1/integration/payments/transactions/by-date/`
- legacy path: `POST /api/v1/get-transactions-by-date`
- purpose: lists building payments filtered by date range

#### Integration Query Parameters

```text
?building_id=0348200&from_date=2026-06-01&to_date=2026-06-30&date_type=input_date&pay_method=all
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | Building ID |
| `from_date` | string | yes | Start day |
| `to_date` | string | yes | End day |
| `date_type` | string | yes | `input_date` or `tran_date` |
| `pay_method` | string | no | `all` or `pos` |

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

- `date_type` controls whether filtering uses `txdtm` or `tran_date`
- `pay_method = "pos"` keeps only payment types whose name starts with `POS`
- the query applies manual `-8h` and `+8h` adjustments around the date range

## 6. Cashier and Bank-In APIs

### 6.1 Update Transactions Status in Cashier

- endpoint: `POST /api/v1/integration/cashier/bank-in/`
- legacy path: `POST /api/v1/update-transactions-status-in-cashier`
- purpose: groups selected cashier payments into a new bank-in batch and updates them to `pending_validation`

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

- all selected payments must currently be `in_cashier`
- the endpoint does not separately verify that all selected rows belong to the same building and pay type before using the first row as batch metadata

### 6.2 Bank-In Record List

- endpoint: `GET /api/v1/integration/cashier/bank-in-records/`
- legacy path: `POST /api/v1/get-bank-in-record-list`
- purpose: returns the latest bank-in batches for one building

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

- returns at most 30 records

### 6.3 Bank-In Record Details

- endpoint: `GET /api/v1/integration/cashier/bank-in-records/details/`
- legacy path: `POST /api/v1/get-bank-in-record-details`
- purpose: returns detailed payment rows inside one bank-in batch

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

- if the batch ID exists but has no matching rows, the API returns an empty list

## 7. Webhook Endpoints

These are machine endpoints but they do not follow the same contract style as the JSON APIs above.

### 7.1 QFPay Callback

- endpoint: `POST /api/v1/integration/webhooks/qfpay/`
- legacy path: `POST /qfpayapi/`
- purpose: stores callback payloads and finalizes successful QFPay payment notifications

#### Expected Inputs

- header: `X-QF-SIGN`
- body: JSON payload from QFPay

Fields referenced by current code include:

- `notify_type`
- `status`
- `out_trade_no`

#### Current Behavior

- if header `X-QF-SIGN` exists, the endpoint returns plain text `SUCCESS`
- if the body indicates a successful payment callback, the endpoint loads `Payment.id = out_trade_no` and runs the legacy success handler
- if the header is missing, the endpoint still stores the payload and returns plain text `UNSUCCESS`

#### Important Notes

- current code checks only for presence of `X-QF-SIGN`; it does not verify the signature
- response body is plain text, not JSON

### 7.2 Test Webhook

- endpoint: `POST /api/v1/integration/webhooks/test/`
- legacy path: `POST /test_webhook/`
- purpose: stores and emails incoming webhook payloads for debugging

#### Current Behavior

- accepts any `POST` body
- stores raw body and content type
- attempts to email a diagnostic copy
- returns plain text `success`
- returns HTTP `405` for non-`POST`

#### Important Notes

- this is a debug utility endpoint, not a production business API
- there is no authentication, signature validation, or schema validation

## Validation Rules Summary

Common validation patterns used across the APIs:

- building IDs are often restricted to alphanumeric values only
- malformed JSON returns `400`
- missing required fields return `400`
- building, user, door, or QR lookup failures return `404`
- authorization checks for external endpoints usually depend on `user_id` supplied in the request body

## Known Contract Caveats

Integrators should be aware of the following inconsistencies in current implementation:

- response envelopes are inconsistent across endpoints
- some endpoints return raw arrays while others return wrapped objects
- payment endpoints mix string status codes and `status` / `message` objects
- several payment queries use raw SQL and manual timezone shifting
- `PosPaymentToIsmart` is currently exposed without an active IP whitelist
- `qfpayapi` does not validate the callback signature
- `ExternalBuildingInfoApi` audit report `btype` naming must stay aligned with the data source
- door-access list permission logic and remote-open permission logic are not identical

## Method Design Recommendation

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

## Example cURL Requests

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

## Operational Checklist

- populate `EXTERNAL_APP_ALLOWED_IPS` for production integrations
- decide whether `PosPaymentToIsmart` should be protected by `POS_PAYMENT_ALLOWED_IPS`
- keep audit report `btype` naming aligned if audit files must appear in `/api/v1/integration/buildings/info/`
- test door open and QR generation against the target hardware environment
- treat `test_webhook/` as non-production and disable or protect it if not needed
