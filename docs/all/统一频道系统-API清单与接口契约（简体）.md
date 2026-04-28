# 统一频道系统 API 清单与接口契约

## 1. 文档信息

| 项目 | 内容 |
| --- | --- |
| 文档名称 | 统一频道系统 API 清单与接口契约 |
| 适用范围 | 统一入口页 + 楼盘放售频道 + 服务式住宅频道 + 二手邻里交易频道 |
| 当前版本 | `v0.1` |
| 文档状态 | Draft |
| 上游输入文档 | 三份需求文档、三份 MVP 功能清单、`统一频道系统-页面清单与页面流转图（简体）.md`、`统一频道系统-数据模型与ERD（简体）.md` |
| 文档目的 | 定义统一频道系统一期 MVP 的 API 边界、资源命名、鉴权规则、关键请求与响应结构，作为前后端联调、测试设计和权限校验的基线 |

## 2. 统一接口规范

### 2.1 基础约定
- Base URL：`/api/v1`
- 鉴权方式：`Authorization: Bearer <token>`
- 返回格式：统一 JSON
- 资源标识：前端和外部接口一律使用 `public_id`
- 时间格式：统一 `ISO 8601`
- 分页参数：`page`、`page_size`
- 写接口幂等：支付、发布、续期、重发、询盘、发消息建议支持 `Idempotency-Key`

### 2.2 统一响应结构

```json
{
  "code": "OK",
  "message": "success",
  "data": {},
  "request_id": "req_01HXXX",
  "timestamp": "2026-04-15T10:00:00Z"
}
```

### 2.3 统一错误结构

```json
{
  "code": "AUTH_REQUIRED",
  "message": "login required",
  "errors": [
    {
      "field": "token",
      "reason": "missing"
    }
  ],
  "request_id": "req_01HXXX",
  "timestamp": "2026-04-15T10:00:00Z"
}
```

### 2.4 核心错误码建议

| 错误码 | 场景 |
| --- | --- |
| `AUTH_REQUIRED` | 未登录 |
| `AUTH_FORBIDDEN` | 权限不足 |
| `VALIDATION_ERROR` | 参数校验失败 |
| `RESOURCE_NOT_FOUND` | 资源不存在 |
| `RESOURCE_EXPIRED` | 资源已过期 |
| `RESOURCE_HIDDEN` | 资源被隐藏或下架 |
| `POINTS_INSUFFICIENT` | 积分不足 |
| `PAYMENT_STATUS_INVALID` | 支付状态不允许当前操作 |
| `VISIBILITY_FORBIDDEN` | 邻里可见范围无权限 |
| `RATE_LIMITED` | 请求频率受限 |
| `IDEMPOTENCY_CONFLICT` | 幂等冲突 |
| `INTERNAL_ERROR` | 服务器异常 |

## 3. 鉴权与权限口径

### 3.1 用户分层
- `guest`：未登录访客
- `member`：已登录普通会员
- `publisher`：可发布内容的会员
- `operator`：运营后台用户

### 3.2 权限规则
- 公开列表页和公开详情页允许访客访问。
- 联系方式明文、站内消息、询盘表单、发布与编辑动作必须登录。
- `building_only` 的二手详情页必须校验用户所属社区。
- 运营接口不得暴露给普通前台用户。

## 4. 公共接口

### 4.1 认证与会员资料

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/auth/otp/request` | 否 | 请求短信验证码 |
| `POST` | `/auth/otp/verify` | 否 | 验证验证码并登录 |
| `POST` | `/auth/logout` | 是 | 登出 |
| `GET` | `/me` | 是 | 获取当前用户信息 |
| `PATCH` | `/me/profile` | 是 | 更新会员资料，例如昵称、身份、所属社区 |
| `GET` | `/meta/communities` | 是 | 获取大厦/屋苑列表，用于资料完善和筛选 |

#### `POST /auth/otp/request`

请求体：

```json
{
  "phone_country_code": "+852",
  "phone_number": "91234567",
  "scene": "login"
}
```

#### `POST /auth/otp/verify`

响应体 `data` 建议：

```json
{
  "access_token": "jwt_token",
  "refresh_token": "refresh_token",
  "expires_in": 7200,
  "user": {
    "public_id": "01HRUSERXXX",
    "member_status": "active",
    "profile_completed": true
  }
}
```

### 4.2 统一入口页与公共内容

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/channel-home/overview` | 否 | 获取统一入口页所需频道卡片、精选内容和推荐区块 |

### 4.3 上传与媒体

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/uploads/presign` | 是 | 申请上传凭证 |
| `POST` | `/uploads/complete` | 是 | 上传完成后登记媒体元数据 |

#### `POST /uploads/presign`

请求体：

```json
{
  "file_name": "cover.webp",
  "mime_type": "image/webp",
  "file_size": 182736
}
```

响应体 `data` 建议：

```json
{
  "upload_url": "https://storage.example.com/...",
  "object_key": "listings/2026/04/cover.webp",
  "headers": {
    "Content-Type": "image/webp"
  }
}
```

### 4.4 统一联系方式与聊天

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/listings/{listingId}/contact-access` | 是 | 校验权限并返回可放行的联系方式或渠道摘要 |
| `POST` | `/listings/{listingId}/chats` | 是 | 以帖子为上下文创建或复用站内会话 |
| `GET` | `/chats` | 是 | 获取当前用户会话列表 |
| `GET` | `/chats/{chatId}` | 是 | 获取会话信息 |
| `GET` | `/chats/{chatId}/messages` | 是 | 获取消息列表 |
| `POST` | `/chats/{chatId}/messages` | 是 | 发送消息 |
| `POST` | `/chats/{chatId}/read` | 是 | 标记会话已读 |

#### `POST /listings/{listingId}/contact-access`

说明：
- 该接口只在登录校验通过后放行联系方式。
- 二手 `building_only` 内容需先校验用户社区。
- 邮箱不应直接下发给前端，服务式住宅仅返回“可提交询盘表单”。

示例响应：

```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "listing_id": "01HRLISTINGXXX",
    "allowed_channels": {
      "phone": true,
      "whatsapp": true,
      "chat": true,
      "inquiry_form": false
    },
    "contact_payload": {
      "phone": "+852 9123 4567",
      "whatsapp_url": "https://wa.me/85291234567?text=..."
    }
  }
}
```

#### `POST /listings/{listingId}/chats`

示例响应：

```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "chat_id": "01HRCHATXXX",
    "is_new": false
  }
}
```

## 5. 楼盘放售模块接口

### 5.1 前台浏览接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/property-sale/listings` | 否 | 楼盘列表页 |
| `GET` | `/property-sale/listings/{listingId}` | 否 | 楼盘详情页 |

列表筛选参数建议：
- `keyword`
- `district_code`
- `estate_name`
- `min_price_hkd`
- `max_price_hkd`
- `min_saleable_area_sqft`
- `max_saleable_area_sqft`
- `bedroom_count`
- `sort_by=latest|price_asc|price_desc|area_desc`

### 5.2 卖家发布与管理接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/property-sale/listings` | 是 | 创建楼盘草稿 |
| `PATCH` | `/property-sale/listings/{listingId}` | 是 | 更新楼盘草稿或已发布房源 |
| `POST` | `/property-sale/listings/{listingId}/publish` | 是 | 消耗积分发布房源 |
| `POST` | `/property-sale/listings/{listingId}/renew` | 是 | 消耗积分续期房源 |
| `POST` | `/property-sale/listings/{listingId}/mark-sold` | 是 | 标记已售 |
| `GET` | `/me/property-sale/listings` | 是 | 我的楼盘列表 |

#### `POST /property-sale/listings`

请求体核心字段：

```json
{
  "title": "康怡花园三房放售",
  "district_code": "hk_east",
  "community_id": "01HRCOMMUNITYXXX",
  "publisher_identity_type": "owner",
  "property_type": "residential",
  "estate_name": "康怡花园",
  "asking_price_hkd": 7800000,
  "saleable_area_sqft": 620,
  "bedroom_count": 3,
  "livingroom_count": 2,
  "bathroom_count": 2,
  "feature_tags": ["owner_direct", "renovated"],
  "images": [
    { "media_asset_id": "01HRMEDIAXXX", "sort_order": 1, "is_cover": true }
  ],
  "contact": {
    "phone": "+85291234567",
    "show_phone": true,
    "show_whatsapp": true,
    "show_chat": true
  }
}
```

#### `POST /property-sale/listings/{listingId}/publish`

说明：
- 后端必须重新校验积分余额。
- 积分扣减与帖子状态变更必须同事务完成。

成功响应 `data` 建议：

```json
{
  "listing_id": "01HRLISTINGXXX",
  "publication_status": "active",
  "expire_at": "2026-05-15T10:00:00Z",
  "points_balance_after": 2500,
  "points_transaction_id": "01HRPOINTSXXX"
}
```

错误示例：

```json
{
  "code": "POINTS_INSUFFICIENT",
  "message": "insufficient AJO points"
}
```

### 5.3 积分与支付接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/me/points/account` | 是 | 获取 AJO Point 账户余额 |
| `GET` | `/me/points/transactions` | 是 | 获取积分流水 |
| `POST` | `/points/orders` | 是 | 创建积分充值订单 |
| `GET` | `/points/orders/{orderId}` | 是 | 获取充值订单详情 |
| `POST` | `/payments/callbacks/{provider}` | 否，内部签名校验 | 支付回调 |

#### `POST /points/orders`

请求体：

```json
{
  "hkd_amount": 100.00,
  "payment_provider": "stripe"
}
```

响应体 `data` 建议：

```json
{
  "order_id": "01HRORDERXXX",
  "order_no": "AJO202604150001",
  "payment_status": "pending",
  "points_to_credit": 1000,
  "checkout_payload": {
    "provider": "stripe",
    "checkout_url": "https://checkout.stripe.com/..."
  }
}
```

## 6. 服务式住宅模块接口

### 6.1 前台浏览接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/serviced-apartments/projects` | 否 | 项目列表 |
| `GET` | `/serviced-apartments/projects/{listingId}` | 否 | 项目详情，包含房型摘要 |

列表筛选参数建议：
- `keyword`
- `district_code`
- `min_monthly_rent_hkd`
- `max_monthly_rent_hkd`
- `min_lease_months`
- `facility_tags[]`
- `sort_by=latest|price_asc|price_desc`

### 6.2 发布与管理接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/serviced-apartments/projects` | 是 | 创建项目草稿 |
| `PATCH` | `/serviced-apartments/projects/{listingId}` | 是 | 更新项目 |
| `POST` | `/serviced-apartments/projects/{listingId}/publish` | 是 | 发布项目 |
| `POST` | `/serviced-apartments/projects/{listingId}/renew` | 是 | 续期项目 |
| `GET` | `/me/serviced-apartments/projects` | 是 | 我的项目列表 |
| `POST` | `/serviced-apartments/projects/{listingId}/room-types` | 是 | 新增房型 |
| `PATCH` | `/serviced-apartments/projects/{listingId}/room-types/{roomTypeId}` | 是 | 更新房型 |
| `DELETE` | `/serviced-apartments/projects/{listingId}/room-types/{roomTypeId}` | 是 | 删除房型 |

#### `POST /serviced-apartments/projects/{listingId}/room-types`

请求体：

```json
{
  "room_type_name": "一房单位",
  "saleable_area_sqft": 420,
  "monthly_rent_hkd": 23800,
  "utilities_included": true,
  "min_lease_months": 3,
  "room_facility_tags": ["wifi", "washing_machine"],
  "images": [
    { "media_asset_id": "01HRMEDIAXXX", "sort_order": 1, "is_cover": true }
  ]
}
```

### 6.3 询盘与联系接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/serviced-apartments/projects/{listingId}/inquiries` | 是 | 提交询盘表单 |

#### `POST /serviced-apartments/projects/{listingId}/inquiries`

请求体：

```json
{
  "room_type_id": "01HRROOMXXX",
  "move_in_date": "2026-05-01",
  "lease_months": 6,
  "occupant_count": 2,
  "remark_text": "希望了解是否可养宠物"
}
```

成功响应 `data` 建议：

```json
{
  "inquiry_id": "01HRINQUIRYXXX",
  "inquiry_status": "submitted",
  "backup_chat_id": "01HRCHATXXX",
  "delivery_status": "pending"
}
```

说明：
- 创建询盘后，应生成一条站内备份记录。
- 邮件发送可以异步，但 `delivery_status` 必须可追踪。

## 7. 二手邻里交易模块接口

### 7.1 前台浏览接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/secondhand/listings` | 否 | 二手列表 |
| `GET` | `/secondhand/listings/{listingId}` | 否/条件校验 | 二手详情；若为 `building_only` 则需进一步校验权限 |

列表筛选参数建议：
- `keyword`
- `category_code`
- `district_code`
- `price_mode`
- `min_price_hkd`
- `max_price_hkd`
- `only_building=true`
- `only_free=true`
- `sort_by=latest|price_asc|price_desc`

说明：
- 对于 `building_only` 帖子，未登录或非同社区用户可返回 `404` 或 `VISIBILITY_FORBIDDEN`，具体以安全策略为准。

### 7.2 卖家发布与管理接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/secondhand/listings` | 是 | 创建二手草稿 |
| `PATCH` | `/secondhand/listings/{listingId}` | 是 | 更新二手帖子 |
| `POST` | `/secondhand/listings/{listingId}/publish` | 是 | 发布二手帖子 |
| `POST` | `/secondhand/listings/{listingId}/republish` | 是 | 重发过期帖子 |
| `POST` | `/secondhand/listings/{listingId}/mark-sold` | 是 | 标记已售 |
| `POST` | `/secondhand/listings/{listingId}/deactivate` | 是 | 下架 |
| `GET` | `/me/secondhand/listings` | 是 | 我的二手帖子列表 |

#### `POST /secondhand/listings`

请求体核心字段：

```json
{
  "title": "九成新洗衣机",
  "district_code": "kwun_tong",
  "community_id": "01HRCOMMUNITYXXX",
  "category_code": "home_appliance",
  "price_mode": "fixed",
  "price_hkd": 1200,
  "condition_level": "used_good",
  "dimension_text": "60 x 60 x 85 cm",
  "pickup_region_code": "kwun_tong",
  "pickup_location_text": "屋苑楼下自提",
  "delivery_tags": ["self_pickup", "elevator"],
  "visibility_scope": "building_only",
  "contact_method": "both",
  "images": [
    { "media_asset_id": "01HRMEDIAXXX", "sort_order": 1, "is_cover": true }
  ],
  "contact": {
    "show_whatsapp": true,
    "show_chat": true,
    "whatsapp": "+85291234567"
  }
}
```

#### `POST /secondhand/listings/{listingId}/republish`

说明：
- 仅允许帖子归属人操作。
- 仅允许 `expired` 或 `inactive` 状态进入重发流程。
- 支持同步更新价格、图片或描述。

成功响应 `data` 建议：

```json
{
  "listing_id": "01HRLISTINGXXX",
  "publication_status": "active",
  "expire_at": "2026-04-29T10:00:00Z",
  "sort_refreshed_at": "2026-04-15T10:00:00Z"
}
```

## 8. 运营后台接口

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET` | `/ops/listings` | 是，`operator` | 按模块和状态搜索帖子 |
| `POST` | `/ops/listings/{listingId}/approve` | 是，`operator` | 审核通过 |
| `POST` | `/ops/listings/{listingId}/reject` | 是，`operator` | 审核拒绝 |
| `POST` | `/ops/listings/{listingId}/hide` | 是，`operator` | 隐藏帖子 |
| `POST` | `/ops/listings/{listingId}/restore` | 是，`operator` | 恢复帖子 |
| `GET` | `/ops/payments/orders` | 是，`operator` | 查看充值订单 |
| `GET` | `/ops/points/transactions` | 是，`operator` | 查看积分流水 |
| `GET` | `/ops/inquiries` | 是，`operator` | 查看询盘记录与投递状态 |
| `GET` | `/ops/notifications/deliveries` | 是，`operator` | 查看邮件投递异常 |

## 9. 统一列表返回结构建议

### 9.1 列表接口 `data` 建议

```json
{
  "items": [
    {
      "listing_id": "01HRLISTINGXXX",
      "module": "secondhand",
      "title": "九成新洗衣机",
      "cover_image_url": "https://cdn.example.com/cover.webp",
      "district_code": "kwun_tong",
      "publication_status": "active",
      "business_status": "available",
      "sort_refreshed_at": "2026-04-15T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 126
  }
}
```

### 9.2 详情接口设计口径
- 返回公开字段。
- 返回 `contact_summary`，只说明有哪些联络渠道可用。
- 不直接返回电话、WhatsApp、邮箱明文。
- 若需要真实联系方式，必须单独调用 `contact-access` 接口。

## 10. 幂等、回调与安全要求

### 10.1 幂等要求
- 以下接口必须支持幂等：
  - `POST /points/orders`
  - `POST /property-sale/listings/{id}/publish`
  - `POST /property-sale/listings/{id}/renew`
  - `POST /serviced-apartments/projects/{id}/inquiries`
  - `POST /secondhand/listings/{id}/republish`
  - `POST /chats/{id}/messages`

### 10.2 支付回调要求
- 支付回调必须校验三方签名。
- 回调处理必须幂等。
- 订单状态更新、积分到账和流水写入必须具备可追踪链路。

### 10.3 联系方式与隐私要求
- 真实电话和 WhatsApp 只通过受控接口返回。
- 服务式住宅邮箱不得返回给前端。
- 所有联系方式查看动作必须记录日志并施加限流。

## 11. 当前结论

- 当前 API 设计应采用“公共接口层 + 模块业务接口层 + 运营接口层”的三层结构。
- 联系方式必须与详情数据解耦，单独通过授权接口放行，这是整个系统的关键安全约束。
- 该接口契约已经足够支撑下一步前后端任务拆分、Mock 数据定义、联调用例编写和测试用例设计。
