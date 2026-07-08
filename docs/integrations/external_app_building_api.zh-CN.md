# External App Building APIs（简体中文翻译）

## 概览

本文档说明为替换部分住户网页而新增的对外接口：

- `building_access`
- `blg_cs`
- 合并后的 `blg_form` + `blg_info`

这些接口面向可信任的服务端到服务端调用，或 App 后端集成场景。

## 安全

- 所有接口均受 `@ip_whitelist('EXTERNAL_APP_ALLOWED_IPS')` 保护。
- 允许访问的 IP 在 `stc/settings.py` 中配置。
- 这些接口不使用基于 session 的认证。
- 调用方必须显式提供 `user_id`、`building_id`、`door_id` 或 `qrcode_record_id` 等标识。

## 环境

- production：`https://ismart.ajoliving.com`
- clouddev：`https://clouddev.ismart.ajoliving.com`

由于门禁控制器未连接到 clouddev 环境，部分门禁功能可能无法在 clouddev 正常使用。

## 通用约定

- **基础路径**：`/api/v1/external/`
- **请求方法**：`POST`
- **Content-Type**：`application/json`
- **响应格式**：
  - 成功：`{"status": "success", "data": ...}`
  - 失败：`{"status": "error", "message": "..."}`

## 错误处理

常见 HTTP 状态码：

- `200` 请求成功。
- `201` 资源创建成功。
- `400` 请求体无效或缺少必填字段。
- `403` 调用方 IP 不在白名单内，或用户无权访问请求的资源。
- `404` 引用的记录不存在。
- `500` 服务端发生非预期错误。
- `502` 上游设备或服务失败，例如门禁控制器或二维码载荷生成服务。

## 1. 合并后的大厦资料 + 表格

返回 `showdata.BuildingInfo` 中的大厦资料，以及 `ismart.Blgfiles` 中的相关文件，范围包括：

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

- `building_id` string，必填，目标大厦 ID。

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

`Blgfiles` 模型（`ismart.Blgfiles`）上的 `btype` 字段决定文件所属分类。目前定义的类型如下：

| `btype` value | Category | Returned in |
| --- | --- | --- |
| `form` | 表格文件（Forms） | `documents.forms` |
| `blginfo` | 大厦资料文件（Building Info Files） | `documents.building_info_files` |
| `floorplan` | 平面图（Floor Plans） | `documents.floorplan` |
| `auditreport` | 审计报告（Audit Reports） | `documents.auditreport` |
| `mfinreport` | 财务报告（Financial Reports） | `documents.mfinreport` |

### Notes

- 即使没有文件，也会返回 `building_info`。
- 没有文件时，`documents.forms` 和 `documents.building_info_files` 会返回空数组。
- `floorplan`、`auditreport` 和 `mfinreport` 已在模型中定义，但当前不会由此合并接口返回；它们由独立的员工管理视图处理。

## 2. 大厦意见提交（`blg_cs`）

创建一条 `ismart.BlgComment` 记录，并发送现有通知邮件。

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

- `user_id` integer，必填，`accounts.CustomUser.id`。
- `building_id` string，必填，目标大厦 ID。
- `comment_type` string，必填，必须是当前 `BlgComment.COMCHOICES` 之一。
- `comment` string，必填，意见内容。

### Allowed `comment_type` Values

以下枚举值属于旧系统入参，调用时应保持原值：

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

- 提供的 `user_id` 必须有权访问提供的 `building_id`。
- 访问权限使用与当前页面逻辑相同的用户到大厦关系模式解析。

## 3. 大厦门禁摘要

返回某个用户在指定大厦下的门列表、密码信息、二维码可用性和近期门禁记录。

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

- `user_id` integer，必填，`accounts.CustomUser.id`。
- `building_id` string，必填，目标大厦 ID。

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

- `has_permission` 表示用户当前是否拥有该门的权限。
- 没有密码记录时，`password` 为 `null`。
- 二维码门禁未启用或未配置时，`qrcode` 为 `null`。
- 没有关联镜头时，`camera` 为 `null`。
- `requested_building_id` 和 `access_building_id` 可能不同，因为旧页面对特定大厦 ID 包含硬编码映射。

## 4. 远程开门

触发与网页相同的远程开门流程，并记录结果。

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

- `user_id` integer，必填，`accounts.CustomUser.id`。
- `door_id` integer，必填，`smartliving.Door.id`。
- `building_id` string，选填但建议提供，用于额外校验。

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

- 用户必须有权访问目标大厦。
- 用户还必须拥有相关单位或客户权限，且该权限在现有项目逻辑下允许远程开门。

## 5. 生成门禁二维码载荷

为已启用二维码的门禁凭证生成二维码载荷字符串。外部 App 可以在客户端将该载荷渲染为二维码。

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

- `user_id` integer，必填，`accounts.CustomUser.id`。
- `qrcode_record_id` integer，必填，`smartliving.DoorPasswordRecord.id`，且 `passwordorqrcode=True`。
- `term` string，选填，允许值：
  - `dynamic` 默认值，有效期 15 分钟。
  - `static` 有效期最长 360 天，并受凭证结束时间限制。

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

建议客户端使用合适的二维码库，将 `qrcode_value` 渲染为二维码图片。

## 6. POS Payment to iSmart

从 POS 终端或外部支付系统创建一笔支付。该接口调用与生产 POS 集成相同的后端逻辑。

- **IP 白名单**：`POS_PAYMENT_ALLOWED_IPS`，与 `EXTERNAL_APP_ALLOWED_IPS` 分开。
- `@ip_whitelist` 装饰器存在于 view class 上，但当前已被注释掉；除非取消注释，否则不会强制执行 IP 限制。

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
| --- | --- | --- | --- |
| `BLG_ID` | string | **required** | 大厦标识 |
| `PAY_METHOD` | string | **required** | 支持的支付方式之一，见下方说明 |
| `FINAL_AMOUNT` | integer | cash、cheque、bank、web POS 必填 | 金额，单位为**分**，例如 150000 = HKD 1500.00 |
| `AMOUNT` | integer | `allinpay` 必填 | 金额，单位为**分** |
| `ENTRY_DATETIME` | string | cash、cheque、bank、web POS 必填 | 本地录入时间，格式为 `YYYY-MM-DD HH:mm:ss` |
| `TRAN_DATETIME` | string | cash、cheque、bank、web POS 必填 | 交易时间，格式为 `YYYY-MM-DD HH:mm:ss` |
| `DATE` + `TIME` | string | `allinpay` 必填 | 日期 `YYYYMMDD` 与时间 `HHmmss` 拼接 |
| `TRAN_REF_NO` | string | optional | 外部参考号或流水号 |
| `TRANS_TRACE_NO` | string | `allinpay` 必填 | Allinpay 流水号 |
| `TRANS_TICKET_NO` | string | `allinpay` 必填 | Allinpay 票据号 |
| `USER_ID` | integer | optional，默认 `1` | 提交支付的 `CustomUser.id` |
| `FAKE_TRANSACTION` | bool | optional，默认 `false` | 为 `true` 时，状态设置为 `pending_validation`，用于银行转账 |
| `COMMENT` | string | optional | 自由文本备注 |
| `PIC_FILENAME` | array[string] | optional | 附件图片文件名列表 |
| `PIC_DATA` | array[string] | optional | base64 编码图片数据列表，与 `PIC_FILENAME` 按索引对应 |
| `bank_account_received` | string | optional | 银行账户号码，用于入账分组 |
| `PAYMENT_GATEWAY_RESPONSE` | object | optional | 需要保存的原始支付网关响应 |
| `BILL_OBJS` | array[object] | **required** | 账单项目列表，见下方说明 |

### `BILL_OBJS` Item Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `flat_code` | string | **required** | 单位 ID，例如 `0348200001` |
| `invoice_no` | string | optional | 发票号。**预付项目省略或留空**。当 `trs_to` 为 `預付` 时，`invoice_no` 会被强制置空。 |
| `item_id` | string | **required** | 费用项目类型，例如 `管理費`、`分攤費用`、`其他費用` |
| `trs_to` | string | **required** | 账期，例如 `2026/06`；预付项目使用 `預付` |
| `net_amount` | integer | **required** | 金额，单位为**分**，必须为正整数。预付项目中表示充值金额。 |
| `transfer_fee` | integer | optional，默认 `0` | 手续费或转账费用，单位为**分** |

### Supported `PAY_METHOD` Values

| Value | Status | Notes |
| --- | --- | --- |
| `POS_CASH` | `in_cashier` | POS 现金付款 |
| `POS_CHEQUE` | `in_cashier` | POS 支票付款 |
| `POS_BANK` | `pending_validation` | POS 银行转账 |
| `POS_ALIWE` | `in_cashier` | POS 支付宝或微信 |
| `WEBPOS_ALIPAY` | `payment_captured` | 在线支付宝，会立即创建 `BalCalcTbl` 和 `PrepaidBal` |
| `WEBPOS_WECHAT` | `payment_captured` | 在线微信，会立即创建 `BalCalcTbl` 和 `PrepaidBal` |
| `WEBPOS_CARD_UP` | `payment_captured` | 在线信用卡，会立即创建 `BalCalcTbl` 和 `PrepaidBal` |
| `allinpay` | `init` | Allinpay 网关，需要 `DATE`、`TIME`、`TRANS_TRACE_NO`、`TRANS_TICKET_NO` |

### Validation Rules

1. **发票金额**：非预付项目中，`net_amount` 不得超过该发票当前未付余额。超出时返回 `400`。
2. **大厦存在性**：`BLG_ID` 必须引用有效的 `IBuildingTbl`。不存在时返回 `404`。
3. **支付类型**：`PAY_METHOD` 必须是该大厦已配置的 `BuildingPaymentType`。未配置时返回 `404`。

### Status Lifecycle

完整状态图请参考项目 `AGENTS.md` 中的 “Payment Status Lifecycle”。简要说明如下：

- `in_cashier` -> staff bank-in -> `pending_validation` -> staff validate -> `validated_by_ismart`
- `pending_validation` -> staff validate -> `validated_by_ismart`
- `payment_captured` -> 立即创建 `BalCalcTbl` 条目与 `PrepaidBal` 记录；状态保持 `payment_captured`
- `init` -> 仅适用于 `allinpay`；在网关回调后转换为确认状态

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

- 请求中的所有金额均使用**分**作为单位，类型为整数。服务端内部会转换为元。
- 对于 `payment_captured` 流程：`BalCalcTbl` 行会以正数 `trs_val` 创建付款条目；当 `SUM(trs_val) >= 0` 时，发票会标记为 `fullypaid`。
- 手续费 `transfer_fee` 会生成独立的 `BalCalcTbl` 抵销对。
- 预付项目（`trs_to = 預付`）会创建与 `Paymentdetails` 行关联的 `PrepaidBal` 记录。
- 该接口已豁免 CSRF：`@method_decorator(csrf_exempt, name='dispatch')`。

## Example cURL Requests

### 获取合并后的大厦资料

```bash
curl -X POST "https://<your-domain>/api/v1/external/building-info/" \
  -H "Content-Type: application/json" \
  -d '{
    "building_id": "0348200"
  }'
```

### 提交大厦意见

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

### 获取大厦门禁数据

```bash
curl -X POST "https://<your-domain>/api/v1/external/building-access/" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "building_id": "0348200"
  }'
```

### 远程开门

```bash
curl -X POST "https://<your-domain>/api/v1/external/building-access/open-door/" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "building_id": "0348200",
    "door_id": 10
  }'
```

### 获取二维码载荷

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

- 将外部 App 服务端 IP 加入 `EXTERNAL_APP_ALLOWED_IPS`。
- 确认外部 App 使用 `application/json` 发送请求。
- 确认外部 App 可以在客户端渲染二维码载荷。
- 确认 App 对门禁密码数据进行了合适的保存和保护。
- 针对目标环境硬件测试远程开门和二维码生成。
