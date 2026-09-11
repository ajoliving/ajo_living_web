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
| Auth | `GET` / `POST` | `/api/v1/integration/auth/check-contact/` | — | Check whether email/phone are already used on `CustomUser` |
| Auth | `GET` / `POST` / `PATCH` / `PUT` | `/api/v1/integration/auth/client/` | — | Get or update primary `ClientTbl` profile (and optional login email/phone) |
| Auth | `POST` | `/api/v1/integration/auth/change-password/` | — | Change `CustomUser` password (requires current password) |
| Auth | `GET` / `POST` / `PATCH` / `PUT` | `/api/v1/integration/auth/settings/` | — | Get or update `UserSettings` (`blg_notice_email`) |
| Building data | `GET` | `/api/v1/integration/buildings/receivables/management-fees/` | `/api/v1/building-mf-table/` | Management-fee receivable summary |
| Building data | `GET` | `/api/v1/integration/buildings/receivables/other-fees/` | `/api/v1/building-of-list/` | Other-fee receivable list |
| Building data | `GET` | `/api/v1/integration/buildings/notices/` | `/api/v1/building-notices/` | Active building notices |
| External building | `GET` | `/api/v1/integration/buildings/info/` | `/api/v1/external/building-info/` | Building profile plus document metadata |
| Member role binding | `POST` | `/api/v1/integration/buildings/building-flat-owner-binding-requests/` | — | Submit an `OwnerReg` request for owner-role binding, staff approval required |
| Member subaccount | `GET` | `/api/v1/integration/buildings/subaccounts/` | — | List current active authorized sub users for owner-controlled unit(s) |
| Member subaccount | `POST` | `/api/v1/integration/buildings/subaccounts/grant/` | — | Grant `授權用戶` to another `CustomUser` for one owner-controlled unit |
| Member subaccount | `POST` | `/api/v1/integration/buildings/subaccounts/revoke/` | — | Revoke `授權用戶` from one owner-controlled unit |
| External building | `POST` | `/api/v1/integration/buildings/comments/` | `/api/v1/external/blg-cs/submit/` | Submit building service case (維修/意見) |
| External building | `GET` / `POST` | `/api/v1/integration/buildings/service-cases/` | — | List service cases for a building (filter by status) |
| External building | `GET` / `POST` | `/api/v1/integration/buildings/service-cases/<case_id>/` | — | Service case detail + message thread |
| External door access | `POST` | `/api/v1/integration/access/buildings/` | `/api/v1/external/building-access/` | Door list, password, QR info, recent records |
| External door access | `POST` | `/api/v1/integration/access/open-door/` | `/api/v1/external/building-access/open-door/` | Remote open a door |
| External door access | `POST` | `/api/v1/integration/access/qrcode/` | `/api/v1/external/building-access/qrcode/` | Generate QR payload for door access |
| Payment | `GET` | `/api/v1/integration/payments/unpaid-invoices/` | `/api/v1/building-flat-unpaid-invoice-list` | Outstanding invoices for one unit |
| Payment | `GET` / `POST` | `/api/v1/integration/payments/types/` | `/api/v1/get-building-payment-types` | List `BuildingPaymentType` rows for one building |
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
- `/api/v1/integration/buildings/service-cases/`
- `/api/v1/integration/buildings/service-cases/<case_id>/`
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

### Login email & phone format (applies to register / check-contact / client update)

These rules apply to **login credentials** stored on `CustomUser`:

| Credential | Model field | Unique? |
|---|---|---|
| Login email | `CustomUser.email` | Yes, **case-insensitive** |
| Login phone | `CustomUser.phone` | Yes, after **E.164 normalization** |

Contact / billing fields on `ClientTbl` (`cli_email`, `cli_tel`, `cli_tel2`) are **separate** from login credentials (except on **register**, where the new client row is seeded with the same phone/email for convenience).

#### Email format

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

#### Phone format

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

#### Login vs contact fields (client profile API)

| API field | Updates | Notes |
|---|---|---|
| `email` | `CustomUser.email` only | **Does not** overwrite `cli_email` |
| `phone` | `CustomUser.phone`; also sets `cli_tel` unless `cli_tel` is sent in the same request | Login phone must be valid E.164 |
| `cli_email` | `ClientTbl.cli_email` only | Contact / billing email; free-form for display but prefer a real email |
| `cli_tel` | `ClientTbl.cli_tel` only | Contact phone; when sent alone does not change login phone |
| `cli_tel2` | Emergency contact | Free-form (not login credential) |

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

For `login_type=email`, matching is **case-insensitive**.  
For `login_type=phone`, prefer E.164 (`+852…`); national HK 8-digit and legacy `852…` forms are also matched when possible.

#### Success Response

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

- on successful login, the server updates `CustomUser.last_login` to the current time
- `building` and `staff_building_permissions` currently return the same data
- the endpoint does not create a session, JWT, or API token
- **backward compatible**: existing keys under `msg` are not renamed or removed; clients that ignore unknown keys continue to work
- `client` is the primary profile for the login account (`cli_id == username`), not the full M2M `user.clients` list

### 1.2 Direct iSmart Account Registration

- endpoint: `POST /api/v1/integration/auth/register/`
- legacy path: none
- purpose: directly creates a new `accounts.CustomUser` and linked `ClientTbl` profile with no approval flow
- access control: currently no IP whitelist and no user authentication

#### Request

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

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `phone` | string | yes | **Login phone**. Must be valid; stored as **E.164** (see [Login email & phone format](#login-email--phone-format-applies-to-register--check-contact--client-update)). Same as legacy `memberphone`. |
| `email` | string | yes | **Login email**. Valid address; stored **lowercase**; unique case-insensitively. |
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
    "phone": "+85291234567",
    "email": "user@example.com",
    "client_id": "200123"
  }
}
```

`data.phone` / `data.email` are the **normalized stored** login values (E.164 phone, lowercase email).

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

#### Format / validation errors

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

### 1.3 Check Email / Phone Availability

- endpoint: `GET` or `POST` `/api/v1/integration/auth/check-contact/`
- legacy path: none
- purpose: check whether an email and/or phone is already used as a unique login credential on `CustomUser`
- access control: currently no IP whitelist and no user authentication

Use this before registration, or before changing an existing user's login email/phone. Uniqueness is enforced on `CustomUser.email` (case-insensitive) and `CustomUser.phone` (E.164 + legacy format candidates). **Phone must be a valid number** (same rules as register); invalid phone → HTTP `400`.

#### GET Query Parameters

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

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `email` | string | conditional | Login email to check (any case). Provide `email` and/or `phone`. |
| `phone` | string | conditional | Login phone to check. Prefer **E.164** (`+852…` / `+86…`); HK 8-digit national also accepted. Provide `email` and/or `phone`. |
| `exclude_user_id` | integer | no | When editing an existing account, pass that user's id so their current email/phone are not reported as taken |

#### Success Response

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

#### Error Responses

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

### 1.4 Get / Update Client Profile

- endpoint: `GET` / `POST` / `PATCH` / `PUT` `/api/v1/integration/auth/client/`
- legacy path: none
- purpose: read or update the primary member profile used by login (`client_tbl` where `cli_id == username`)
- access control: currently no IP whitelist and no user authentication; caller supplies `user_id`

The `client` object shape matches the `msg.client` payload from login (`_serialize_client_tbl_for_login`).

#### GET Query Parameters

```text
?user_id=123
```

#### GET Success Response

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

#### Update Request (`POST` / `PATCH` / `PUT`)

Change **login** credentials (formats: see [Login email & phone format](#login-email--phone-format-applies-to-register--check-contact--client-update)):

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

#### Updatable Fields

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

#### Update Success Response

Same shape as GET success (`status` + `data` with refreshed profile).  
Top-level `data.phone` / `data.email` are login credentials (normalized).  
`data.client.cli_tel` / `data.client.cli_email` are contact fields.

#### Error Responses

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

### 1.5 Change Password

- endpoint: `POST /api/v1/integration/auth/change-password/`
- legacy path: none
- purpose: change the login password for an existing `CustomUser`
- access control: currently no IP whitelist and no user authentication; caller supplies `user_id` and must prove the current password

This is a **change**, not an admin reset. `old_password` is required.

#### Request

```json
{
  "user_id": 123,
  "old_password": "current-secret",
  "new_password": "NewPass123!",
  "new_password_confirm": "NewPass123!"
}
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `user_id` | integer | yes | Target `CustomUser.id` |
| `old_password` | string | yes | Current password. Must match `user.check_password()` |
| `new_password` | string | yes | New password. Validated with Django `AUTH_PASSWORD_VALIDATORS` (min length 8, not too similar to username/email, not a common password, not all-numeric) |
| `new_password_confirm` | string | no | If present, must equal `new_password` |

#### Success Response

```json
{
  "status": "success",
  "data": {
    "user_id": 123,
    "username": "200123"
  }
}
```

The new password is **never** returned.

#### Error Responses

```json
{
  "status": "error",
  "message": "Current password is incorrect"
}
```

HTTP `401` when `old_password` is wrong.

```json
{
  "status": "error",
  "message": "Invalid login credentials"
}
```

HTTP `401` when the user is inactive (`is_active=False`).

```json
{
  "status": "error",
  "message": "old_password is required"
}
```

```json
{
  "status": "error",
  "message": "new_password is required"
}
```

```json
{
  "status": "error",
  "message": "new_password and new_password_confirm do not match"
}
```

```json
{
  "status": "error",
  "message": "This password is too short. It must contain at least 8 characters."
}
```

HTTP `400` for missing fields, confirm mismatch, unknown `user_id`, invalid JSON, or Django password-validator failures.

#### Notes

- success calls `user.set_password()` and saves only the `password` field
- `last_login` is not updated
- this endpoint does not issue a session, JWT, or API token (integration login remains a credential check)
- Django website cookie sessions for this user stop validating on the next request, because `set_password()` changes the session auth hash

### 1.6 Get / Update User Settings

- endpoint: `GET` / `POST` / `PATCH` / `PUT` `/api/v1/integration/auth/settings/`
- legacy path: none
- purpose: read or update `accounts.UserSettings` for a user (currently `blg_notice_email`)
- access control: currently no IP whitelist and no user authentication; caller supplies `user_id`

`UserSettings` is 1:1 with `CustomUser` and is created automatically when the user is created. If the row is missing, GET/update will `get_or_create` it (`blg_notice_email=false`).

Register already uses the alias `is_receive_email` for this column. This endpoint accepts **both** names on write and returns **both** on read.

#### GET Query Parameters

```text
?user_id=123
```

#### GET / Update Success Response

```json
{
  "status": "success",
  "data": {
    "user_id": 123,
    "username": "200123",
    "blg_notice_email": true,
    "is_receive_email": true
  }
}
```

`data.is_receive_email` always equals `data.blg_notice_email`.

#### Update Request (`POST` / `PATCH` / `PUT`)

```json
{
  "user_id": 123,
  "blg_notice_email": false
}
```

or the register alias:

```json
{
  "user_id": 123,
  "is_receive_email": false
}
```

Accepted boolean values (strings are case-insensitive):

| True | False |
|---|---|
| `true`, `1`, `yes`, `y`, `on` | `false`, `0`, `no`, `n`, `off` |

JSON booleans `true` / `false` and integers `1` / `0` are also accepted. `null` and any other token return HTTP `400`.

#### Updatable Fields

| Field | Type | Description |
|---|---|---|
| `user_id` | integer | Required. Target `CustomUser.id` |
| `blg_notice_email` | boolean | Building-notice email opt-in (`UserSettings.blg_notice_email`) |
| `is_receive_email` | boolean | Alias of `blg_notice_email` (same column) |

At least one of `blg_notice_email` / `is_receive_email` is required on update. If both are sent they must parse to the same boolean.

#### Error Responses

```json
{
  "status": "error",
  "message": "user_id is required"
}
```

```json
{
  "status": "error",
  "message": "No fields to update"
}
```

```json
{
  "status": "error",
  "message": "Unknown or non-editable fields: not_a_setting"
}
```

```json
{
  "status": "error",
  "message": "blg_notice_email and is_receive_email must match"
}
```

```json
{
  "status": "error",
  "message": "blg_notice_email must be a boolean (true/false, 1/0, yes/no, on/off)"
}
```

HTTP `400` for missing `user_id`, unknown user, empty update, unknown fields, alias mismatch, invalid boolean, or invalid JSON.

#### Notes

- `user_id` is identity, not a settings field
- unknown keys are rejected (same as the client profile API)
- this does **not** update `CustomUser` login email/phone or `ClientTbl` profile fields

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
| `documents.audit_reports` | queried as `audition` in current code |
| `documents.financial_reports` | `mfinreport` |

#### Notes

- arrays are returned even when no files exist
- `building_info` is returned with `null` or empty-string values when no `BuildingInfo` row exists
- current code queries audit reports using `btype='audition'`; if your data uses `auditreport`, those files will not appear in this endpoint until code is aligned

### 3.2 Submit Building Service Case (formerly Building Comment)

- endpoint: `POST /api/v1/integration/buildings/comments/`
- legacy path: `POST /api/v1/external/blg-cs/submit/`
- purpose: creates a `BuildingServiceCase` (維修報修 / 意見反映) with the first thread message, and sends staff notification email
- access control: `EXTERNAL_APP_ALLOWED_IPS` plus user-to-building authorization

#### Request (preferred — new taxonomy)

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

#### Request (legacy bridge — still accepted)

```json
{
  "user_id": 123,
  "building_id": "0348200",
  "comment_type": "其他事宜",
  "comment": "Air-conditioner water leakage at the corridor outside 18/F."
}
```

When only `comment_type` / `comment` are sent, the API maps `comment_type` to `(request_type, category, subcategory)` and uses `comment` as `content`.

#### Request Fields

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

#### Legacy `comment_type` values (mapped server-side)

- `門卡報失`, `冷氣滴水`, `嘈音滋擾`, `樓梯雜物`, `水質問題`, `渠務問題`, `保安事宜`, `清潔衛生`, `增加服務`, `電力問題`, `其他事宜`

#### Success Response

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

### 3.3 List Building Service Cases

- endpoint: `GET` or `POST` `/api/v1/integration/buildings/service-cases/`
- purpose: list service cases for a building, optional status / type filters
- access control: `EXTERNAL_APP_ALLOWED_IPS` plus user-to-building authorization

#### Request (GET query or POST JSON)

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

#### Success Response

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

### 3.4 Building Service Case Detail

- endpoint: `GET` or `POST` `/api/v1/integration/buildings/service-cases/<case_id>/`
- purpose: case header + full message thread (attachments included)
- access control: `EXTERNAL_APP_ALLOWED_IPS`; caller must be the case creator **or** staff with access to the case’s building

#### Request

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

#### Success Response

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
    "net_amount_cents": 150000,
    "remark": ""
  }
]
```

#### Notes

- response is a raw array
- `net_amount` is returned in dollars, not cents. JSON numbers are IEEE floats, so values such as `0.29` are not exact. **Do not** convert with `int(net_amount * 100)` — that truncates `0.29` to 28 cents
- `net_amount_cents` is the leftover in integer cents, rounded with Decimal. POS clients should send this value as `BILL_OBJS.net_amount` and sum it for `FINAL_AMOUNT`
- pending payments with status `in_cashier` or `pending_validation` are deducted from the outstanding amount

### 5.2 Building Payment Type List

- endpoint: `GET /api/v1/integration/payments/types/`
- legacy path: `POST /api/v1/get-building-payment-types`
- purpose: returns the `BuildingPaymentType` rows configured for one building
- access control: no IP whitelist and no user authentication

#### Integration Query Parameters

```text
?building_id=0348200
```

#### Legacy Request

```json
{
  "building_id": "0348200"
}
```

#### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `building_id` | string | yes | Building ID |

#### Success Response

```json
{
  "building_id": "0348200",
  "payment_types": [
    {
      "pay_type": "POS_CASH",
      "pay_type_name_chi": "現金",
      "markup": 0.0,
      "is_active": true
    },
    {
      "pay_type": "ALI_WECHAT",
      "pay_type_name_chi": "支付寶/微信",
      "markup": 0.03,
      "is_active": true
    }
  ]
}
```

#### Notes

- both paths accept `GET` (query string) and `POST` (JSON body)
- the list is scoped to the requested building only
- inactive rows are included; clients should read `is_active`
- `pay_type` is the code POS sends as `PAY_METHOD` to `/api/v1/integration/payments/pos/`
- `markup` is the handling-fee rate (for example `0.03` = 3%)
- a building with no configured types returns an empty `payment_types` array

### 5.3 POS Payment to iSmart

- endpoint: `POST /api/v1/integration/payments/pos/`
- legacy path: `POST /api/v1/pos-payment-to-ismart`
- purpose: creates a `Payment` and `Paymentdetails`, and for selected payment methods also creates `BalCalcTbl` and `PrepaidBal` rows immediately
- access control: no active IP whitelist in current code
- content type: `application/json` (pictures are **not** multipart form uploads; they are base64 strings inside the JSON body)

#### Request Example (without pictures)

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

#### Request Example (with one payment receipt picture)

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

#### Request Example (multiple pictures + data URL form)

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

#### Picture upload notes

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

#### Client-side encoding example (pseudo-code)

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
| `PIC_FILENAME` | string or string array | no | Image filename(s). Prefer array. Pair by index with `PIC_DATA`. |
| `PIC_DATA` | string or string array | no | Base64 image payload(s), raw or `data:<mime>;base64,...`. Prefer array. |
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
| `net_amount` | integer | yes | Amount in cents. Use `net_amount_cents` from the unpaid-invoice API. Do not compute cents with `int(dollar_float * 100)` |
| `transfer_fee` | integer | no | Handling fee in cents. Default: `0` |

#### Supported `PAY_METHOD` Values

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

#### Success Response

```json
{
  "code": "200",
  "message": "success",
  "receipt_id": "1000042"
}
```

#### Error responses related to pictures / body size

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

#### Validation Rules

- building must exist
- `PAY_METHOD` must be configured as a `BuildingPaymentType` for the building
- for non-prepaid lines, requested payment amount cannot exceed current unpaid amount
- unsupported pay method returns `400`
- `PIC_FILENAME` / `PIC_DATA` length mismatch or invalid base64 returns `400`
- total JSON body over 5 MB returns `400` (compress images client-side; do not send full-resolution camera files)

#### Important Behavior

- all request money values are in cents
- created `Payment.txamount` and line amounts are stored internally in dollars
- `WEBPOS_*` methods immediately create payment-side accounting rows
- prepaid lines with `trs_to = "預付"` create `PrepaidBal`
- handling fees create separate detail and accounting rows for `平台手續費`
- when pictures are provided, one `PaymentPicture` row is created per decoded image and stored on S3 under `payment/`

#### Important Caveats

- the whitelist decorator for this endpoint is currently commented out
- unknown `bank_account_received` is silently ignored rather than rejected
- the `allinpay` datetime parsing path is legacy and should be tested carefully before use
- camera photos should be compressed client-side; base64 makes the JSON body larger than the original file
- do not send multipart file fields; the endpoint only reads `request.body` as JSON
### 5.4 Transactions in Cashier

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

### 5.5 Transactions by Flat Unit

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

### 5.6 Transactions by Date

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
- `ExternalBuildingInfoApi` currently queries audit reports using `btype='audition'`
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
- align audit report `btype` naming if audit files must appear in `/api/v1/integration/buildings/info/`
- test door open and QR generation against the target hardware environment
- treat `test_webhook/` as non-production and disable or protect it if not needed
