# iSmart 集成 API 参考（简体中文）

## 文档范围

本文档是 iSmart Django 应用当前已注册接口的中文集成参考，覆盖会员、楼宇、门禁、付款、出纳和 Webhook。

- 生产环境：`https://ismart.ajoliving.com`
- 测试环境：`https://clouddev.ismart.ajoliving.com`
- 门禁硬件依赖实际控制器连接，`clouddev` 中部分门禁能力可能不可用。
- 原始英文参考：`external_app_building_api(1) copy.md`；其中保留较长的完整响应示例与图片 Base64 示例。

## 路径与兼容策略

新的主集成路径使用 `/api/v1/integration/...`。旧路径继续保留，现有调用方不必立即迁移；新接入应使用主路径。

| 类别 | 主路径方法 | 旧路径兼容方式 |
| --- | --- | --- |
| 只读集成接口 | `GET` + query 参数 | 大多保留 `POST` + JSON body |
| 写入、敏感门禁、Webhook | `POST` | 保留现有 `POST` 路径 |
| 认证资料接口 | 以接口定义为准 | 登录旧路径继续可用 |

## 鉴权与响应

### IP 白名单

以下能力受 `EXTERNAL_APP_ALLOWED_IPS` 白名单保护：楼宇资料、服务个案、门禁摘要、远程开门和二维码。白名单为空时，当前装饰器会允许所有 IP；生产环境必须显式配置。

其余会员、应收、付款、出纳接口当前没有统一的 token 或 session 鉴权。调用方不得将 `user_id` 视为可信认证凭证，应在调用方服务端实施身份控制。

### 响应格式

接口未完全统一，调用方必须按接口处理：

- 成功包装：`{"status":"success","data":...}`
- 失败包装：`{"status":"error","message":"..."}`
- 旧付款成功：`{"code":"200","message":"success",...}`
- 登录成功：`{"code":"1","msg":...}`
- 部分查询成功直接返回数组。

常见 HTTP 状态：`200` 成功、`201` 已建立、`400` 请求或字段错误、`401` 凭证错误、`403` 无访问权或 IP 未许可、`404` 资源不存在、`409` 账号资料冲突、`502` 上游设备或二维码服务失败。

## 接口总览

| 模块 | 方法 | 当前主路径 | 旧路径 | 用途 |
| --- | --- | --- | --- | --- |
| 登录 | `POST` | `/api/v1/integration/auth/login/` | `/api/v1/pos/login` | 校验登录凭证并返回用户、楼宇及单位权限 |
| 注册 | `POST` | `/api/v1/integration/auth/register/` | 无 | 直接建立 `CustomUser` 与 `ClientTbl` |
| 联系方式查重 | `GET` / `POST` | `/api/v1/integration/auth/check-contact/` | 无 | 检查登录邮箱或手机号是否已使用 |
| 会员资料 | `GET` / `POST` / `PATCH` / `PUT` | `/api/v1/integration/auth/client/` | 无 | 查询或更新主 `ClientTbl` 资料 |
| 管理费应收 | `GET` | `/api/v1/integration/buildings/receivables/management-fees/` | `/api/v1/building-mf-table/` | 管理费应收摘要 |
| 其他费用应收 | `GET` | `/api/v1/integration/buildings/receivables/other-fees/` | `/api/v1/building-of-list/` | 非管理费应收列表 |
| 大厦通告 | `GET` | `/api/v1/integration/buildings/notices/` | `/api/v1/building-notices/` | 有效通告列表 |
| 业主绑定申请 | `POST` | `/api/v1/integration/buildings/building-flat-owner-binding-requests/` | 无 | 提交 `OwnerReg`，仍须管理处审批 |
| 授权用户列表 | `GET` | `/api/v1/integration/buildings/subaccounts/` | 无 | 查询业主控制单位下的授权用户 |
| 授权用户新增 | `POST` | `/api/v1/integration/buildings/subaccounts/grant/` | 无 | 业主授予单位授权用户 |
| 授权用户撤销 | `POST` | `/api/v1/integration/buildings/subaccounts/revoke/` | 无 | 业主撤销单位授权用户 |
| 大厦资料与文件 | `GET` | `/api/v1/integration/buildings/info/` | `/api/v1/external/building-info/` | 大厦资料、基本信息及文件元数据 |
| 提交服务个案 | `POST` | `/api/v1/integration/buildings/comments/` | `/api/v1/external/blg-cs/submit/` | 维修报修或意见反映 |
| 服务个案列表 | `GET` / `POST` | `/api/v1/integration/buildings/service-cases/` | 无 | 查询楼宇服务个案 |
| 服务个案详情 | `GET` / `POST` | `/api/v1/integration/buildings/service-cases/<case_id>/` | 无 | 查询个案及消息线程 |
| 门禁摘要 | `POST` | `/api/v1/integration/access/buildings/` | `/api/v1/external/building-access/` | 门、密码、二维码和近期记录 |
| 远程开门 | `POST` | `/api/v1/integration/access/open-door/` | `/api/v1/external/building-access/open-door/` | 请求控制器开门 |
| 门禁二维码 | `POST` | `/api/v1/integration/access/qrcode/` | `/api/v1/external/building-access/qrcode/` | 生成二维码载荷 |
| 未付账单 | `GET` | `/api/v1/integration/payments/unpaid-invoices/` | `/api/v1/building-flat-unpaid-invoice-list` | 单位未付账单 |
| POS 付款 | `POST` | `/api/v1/integration/payments/pos/` | `/api/v1/pos-payment-to-ismart` | 写入付款及付款明细 |
| 收银交易 | `GET` | `/api/v1/integration/cashier/transactions/` | `/api/v1/get-transactions-in-cashier` | 查询当前收银付款 |
| 单位交易历史 | `GET` | `/api/v1/integration/payments/transactions/by-unit/` | `/api/v1/get-transactions-by-flat-unit` | 查询一个或多个单位的缴费记录 |
| 日期交易历史 | `GET` | `/api/v1/integration/payments/transactions/by-date/` | `/api/v1/get-transactions-by-date` | 按日期范围查询大厦缴费 |
| 收银存行 | `POST` | `/api/v1/integration/cashier/bank-in/` | `/api/v1/update-transactions-status-in-cashier` | 将收银款转为待核实 |
| 存行批次 | `GET` | `/api/v1/integration/cashier/bank-in-records/` | `/api/v1/get-bank-in-record-list` | 查询近期存行批次 |
| 存行批次明细 | `GET` | `/api/v1/integration/cashier/bank-in-records/details/` | `/api/v1/get-bank-in-record-details` | 查询批次内付款行 |
| QFPay 回调 | `POST` | `/api/v1/integration/webhooks/qfpay/` | `/qfpayapi/` | 处理 QFPay 机器回调 |
| 测试 Webhook | `POST` | `/api/v1/integration/webhooks/test/` | `/test_webhook/` | 调试用回调捕获 |

## 1. 会员与认证

### 1.1 登录

`POST /api/v1/integration/auth/login/`

```json
{
  "login_name": "demo_user",
  "password": "secret",
  "login_type": "username"
}
```

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `login_name` | 是 | 用户名、邮箱或手机号 |
| `password` | 是 | 用户密码 |
| `login_type` | 否 | `username`、`email` 或 `phone`；默认 `username` |

成功时返回 `code="1"` 与 `msg`。既有字段不会重命名或删除；新增 `email`、`phone`、`last_login`、`client`、`client_building_permissions` 与 `client_building_flat_units_permissions`。该接口只校验凭证，不建立 session、JWT 或 API token。

### 1.2 注册、联系方式查重与会员资料

| 接口 | 关键请求字段 | 关键规则 |
| --- | --- | --- |
| `POST /auth/register/` | `phone`、`email`、`eng_name`；可选 `chi_name`、`id_card`、`gender`、`legal_entity`、`remark`、`password` | 手机号存为 E.164，邮箱转小写；重复返回 `409`；未提供密码时后端生成并在成功响应返回 |
| `GET` / `POST /auth/check-contact/` | `email`、`phone`、可选 `exclude_user_id` | 至少提供邮箱或手机号；返回各字段可用状态及整体 `available` |
| `GET /auth/client/?user_id=<id>` | `user_id` | 返回用户及主要 `ClientTbl`；没有主资料时 `client` 为 `null` |
| `POST` / `PATCH` / `PUT /auth/client/` | `user_id` 加可更新字段 | 更新资料；登录凭证与联络资料字段独立 |

手机号必须是有效号码。香港 8 位号码可自动识别为 `+852`，新调用建议统一传入完整 E.164，例如 `+85291234567`。登录邮箱大小写不敏感且后端以小写存储。

更新会员资料时，`email` / `phone` 更新 `CustomUser` 登录凭证；`cli_email` / `cli_tel` 更新联络或账单资料。仅提供 `phone` 时会同步更新 `cli_tel`，如两者需要不同，必须同时传入 `phone` 与 `cli_tel`。不可更新 `cli_id`、`cli_join_dt`。

## 2. 楼宇数据

### 2.1 应收与通告

| 接口 | Query 参数 | 返回 |
| --- | --- | --- |
| `GET /buildings/receivables/management-fees/` | `building_id`，可选 `data_structure=table|list|block_dict` | 原始数组或按座、楼层、单位组织的 `blocks`；值可能为金额或 `已付`、`审核中` 等状态 |
| `GET /buildings/receivables/other-fees/` | `building_id`，可选 `data_structure=list|block_dict` | 其他费用，应收行含 `invoice_no`、`flat_code`、`item_id`、`trs_to`、`trs_val`、`remark` |
| `GET /buildings/notices/` | `building_id` | 原始数组，通告含 `id`、`mess_code`、`mess_title`、`mess_date`、`mess_down`、`mess_file` |

通告仅返回 `mess_date <= 当日 < mess_down` 的有效记录。应收和通告查询当前没有 IP 白名单或用户鉴权。

### 2.2 业主申请与授权用户

#### 业主绑定申请

`POST /api/v1/integration/buildings/building-flat-owner-binding-requests/`

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `building_id` | 是 | 大厦 ID |
| `ownedflat` | 是 | 单位 ID 数组或逗号分隔字符串，必须属于该大厦 |
| `user` | 条件 | 已有 `CustomUser.id`；未提供时需提交申请人资料 |
| `reg_tel`、`cli_name` | 条件 | 未提供 `user` 时必填 |
| `cli_role` | 否 | 默认 `業主` |
| `ownernote`、`is_receive_email` | 否 | 申请备注与邮件偏好 |

成功仅代表建立 `OwnerReg`，状态为 `pending_approval`；不会立即建立有效业主关系，仍须管理处按既有流程审批。

#### 授权用户

| 接口 | 必填字段 | 规则 |
| --- | --- | --- |
| `GET /buildings/subaccounts/` | `user_id`；可选 `unit_id` | 只返回请求人作为已批准业主的单位下有效 `授權用戶` |
| `POST /buildings/subaccounts/grant/` | `user_id`、`unit_id`、`target_user_id`；可选 `remark` | 目标用户须存在并有关联 `ClientTbl`；每个单位最多 4 名；重复有效关系返回 `409` |
| `POST /buildings/subaccounts/revoke/` | `user_id`、`unit_id`、`target_user_id`；可选 `remark` | 逻辑上将关系设为无效，并写入历史记录；无有效关系返回 `404` |

### 2.3 大厦资料、服务个案

#### 大厦资料与文件

`GET /api/v1/integration/buildings/info/?building_id=<id>`

返回 `building`、`building_info` 与 `documents`。文件组为 `forms`、`building_info_files`、`floorplans`、`audit_reports`、`financial_reports`，无文件时也返回空数组。现行代码使用 `btype='audition'` 查询审计报告；若资料实际使用 `auditreport`，该组不会返回。

#### 提交服务个案

`POST /api/v1/integration/buildings/comments/`

推荐请求：

```json
{
  "user_id": 123,
  "building_id": "0348200",
  "request_type": "repair",
  "category": "drainage",
  "subcategory": "blocked_drain",
  "subject": "走廊冷气滴水",
  "content": "18 楼走廊冷气滴水",
  "location_text": "18 楼走廊",
  "unit_id": "03482001801",
  "contact_name": "CHAN TAI MAN",
  "contact_phone": "+85291234567"
}
```

旧版 `comment_type` 与 `comment` 仍可传入，服务端会映射至新分类。`user_id` 必须有该大厦访问权。成功返回 `case_id`、`case_no`、分类、`submitted` 状态与建立时间。

#### 服务个案查询

| 接口 | 参数 | 说明 |
| --- | --- | --- |
| `GET` / `POST /buildings/service-cases/` | `user_id`、`building_id`；可选 `status`、分页参数 | 返回该用户可见的个案列表 |
| `GET` / `POST /buildings/service-cases/<case_id>/` | `user_id`、`building_id` | 返回个案详情及消息线程 |

## 3. 门禁

所有门禁接口均要求 IP 白名单，并按 `user_id` 与大厦或门的权限校验。

| 接口 | 请求字段 | 返回或行为 |
| --- | --- | --- |
| `POST /access/buildings/` | `user_id`、`building_id` | 门列表、门密码、二维码可用性、摄像头资料和近期记录；接口保持 `POST`，因为数据敏感 |
| `POST /access/open-door/` | `user_id`、`door_id` | 调用上游控制器开门；上游失败会返回 `502` |
| `POST /access/qrcode/` | `user_id`、`qrcode_record_id` | 返回二维码载荷与有效期；客户端只应渲染载荷，不应自行生成或篡改 |

门禁摘要的列表权限与远程开门权限在现行代码中并非完全一致，接入前应以实际用户和控制器环境验收。

## 4. 付款与出纳

### 4.1 未付账单与 POS 付款

`GET /api/v1/integration/payments/unpaid-invoices/?unit_id=<id>` 返回原始数组。每行包含 `invoice_no`、`flat_code`、`item_id`、`trs_to`、`bill_dt`、`net_amount`、`remark`；金额单位为元，已在收银或待核实中的款项会先行抵扣。

`POST /api/v1/integration/payments/pos/` 的关键字段：

| 字段 | 说明 |
| --- | --- |
| `BLG_ID`、`PAY_METHOD`、`BILL_OBJS` | 必填；大厦、付款方式和付款项目 |
| `FINAL_AMOUNT` | POS 或 Web POS 必填，单位为分 |
| `ENTRY_DATETIME`、`TRAN_DATETIME` | POS 或 Web POS 必填，格式 `YYYY-MM-DD HH:mm:ss` |
| `TRAN_REF_NO`、`USER_ID`、`COMMENT` | 可选的外部编号、用户与备注 |
| `PIC_FILENAME`、`PIC_DATA` | 可选的并行数组；图片为 JSON 内 Base64，不可使用 `multipart/form-data` |
| `bank_account_received`、`PAYMENT_GATEWAY_RESPONSE` | 可选 |

`BILL_OBJS` 每项要求 `flat_code`、`item_id`、`trs_to`、`net_amount`；可选 `invoice_no` 与 `transfer_fee`。所有请求金额均以分传入，内部付款金额以元存储。整个 JSON 请求体上限为 5 MB，图片应压缩后编码，文件名和数据数组长度必须一致。

支持的付款方式与初始状态：`POS_CASH` / `POS_CHEQUE` 为 `in_cashier`，`POS_BANK` 为 `pending_validation`，`POS_ALIWE` 为 `in_cashier`，`WEBPOS_ALIPAY`、`WEBPOS_WECHAT`、`WEBPOS_CARD_UP` 为 `payment_captured`，`allinpay` 为 `init`。成功返回 `{"code":"200","message":"success","receipt_id":"..."}`。

### 4.2 查询与存行

| 接口 | Query 或 body | 返回 |
| --- | --- | --- |
| `GET /cashier/transactions/` | `building_id` | 当前收银付款 |
| `GET /payments/transactions/by-unit/` | 单位 ID 或单位 ID 列表 | 单位缴费历史 |
| `GET /payments/transactions/by-date/` | `building_id`、日期范围 | 按日期的大厦缴费记录 |
| `POST /cashier/bank-in/` | 大厦、收银交易或批次所需字段 | 把收银付款转为 `pending_validation` |
| `GET /cashier/bank-in-records/` | `building_id` | 近期存行批次 |
| `GET /cashier/bank-in-records/details/` | `building_id`、`record_id` | 批次内付款明细 |

付款查询存在旧代码的时区手动换算，接入方应以实际返回时间测试香港时区展示。

## 5. Webhook

| 接口 | 行为与风险 |
| --- | --- |
| `POST /api/v1/integration/webhooks/qfpay/` | 读取 `X-QF-SIGN` 与 QFPay JSON；当前仅检查该 header 是否存在，不验证签名；返回纯文本 `SUCCESS` 或 `UNSUCCESS` |
| `POST /api/v1/integration/webhooks/test/` | 保存原始请求并尝试发送诊断邮件，返回纯文本 `success`；仅供调试，不可视为生产业务接口 |

## 实施检查

- 生产环境先配置并验证 `EXTERNAL_APP_ALLOWED_IPS`。
- 新接入使用 `/api/v1/integration/...`，旧调用保留旧路径直到完成迁移。
- 付款图片在客户端压缩，整个 Base64 JSON 请求体不得超过 5 MB。
- 对 `POS` 付款、远程开门、二维码和 Webhook 在目标环境进行真实联调。
- QFPay 回调和测试 Webhook 当前缺少完整签名或认证校验，不可暴露给非受控网络。

## 变更记录

2026-07-30：根据 `external_app_building_api(1) copy.md` 更新为当前完整集成接口范围，补充认证、楼宇、授权用户、服务个案、付款、出纳和 Webhook。
