# iSmart 接口接入 AJO 模块归属

## 结论

这批接口不应在 AJO 中做成一个统一的 `ismart` 前端入口。它们应按 AJO 主系统的业务域拆分，前端只调用 AJO 后端，AJO 后端再按权限受控访问 iSmart 或 POS 接口。

核心原则：

- AJO Web 不直接调用 `https://ismart.ajoliving.com/api/v1/external/...`。
- AJO 前端不传旧系统 `user_id`，应使用当前 AJO 登录态，由后端从 `UserIsmartAccount` 绑定关系解析。
- 旧系统写入仍由 iSmart 负责时，AJO 后端只做受控转发、权限校验、审计和响应结构收敛。
- 门禁、二维码、支付类接口不得直接暴露旧接口语义给浏览器。

## 接口归属表

| iSmart 主接口 | AJO 归属模块 | 前端入口 | AJO 会员接口 | 接入方式 |
| --- | --- | --- | --- | --- |
| `GET /api/v1/integration/buildings/info/` | 大厦与物业资料 | 大厦页：大厦资料、申请表格 | `GET /api/v1/me/ismart/building-info` | 后端先校验会员大厦可见性，再读取 5 分钟共享缓存；上游读取失败才回退旧 `POST /api/v1/external/building-info/` |
| `GET /api/v1/integration/buildings/notices/` | 通知与大厦公告 | 大厦页：最新通告 | `GET /api/v1/me/ismart/notices` | 后端使用 `building_id` 查询有效通告；读取失败才回退旧 `POST /api/v1/building-notices/` |
| `POST /api/v1/integration/buildings/building-flat-owner-binding-requests/` | 账户、住户与成员 | 会员中心：物业绑定 | `POST /api/v1/me/ismart/owner-binding-requests` | 后端注入当前绑定的 iSmart 用户；仅创建 `OwnerReg` 待审批申请，不建立有效物权绑定 |
| `GET /api/v1/integration/buildings/subaccounts/` | 账户、住户与成员 | 会员中心：授权用户 | `GET /api/v1/me/ismart/subaccounts` | 后端注入当前 iSmart 用户并校验请求单位可见性 |
| `POST /api/v1/integration/buildings/subaccounts/grant/` | 账户、住户与成员 | 会员中心：授权用户 | `POST /api/v1/me/ismart/subaccounts/grant` | 后端只转发 `unit_id`、`target_user_id` 与可选 `remark`，上游负责业主资格及名额校验 |
| `POST /api/v1/integration/buildings/subaccounts/revoke/` | 账户、住户与成员 | 会员中心：授权用户 | `POST /api/v1/me/ismart/subaccounts/revoke` | 后端只转发受控身份和单位资料，上游保留撤销历史与最终状态校验 |
| `Blgfiles.btype=form` | 大厦表格文件 | 大厦页「申请表格」 | 大厦文件读取能力 | 只读展示；文件 URL 需由后端校验可见性后返回 |
| `Blgfiles.btype=blginfo` | 大厦资料文件 | 大厦页「大厦资料」 | 大厦文件读取能力 | 只读展示；与基础资料可合并返回 |
| `Blgfiles.btype=floorplan` | 大厦图则/公共文件 | 大厦页「大厦资料」或独立「平面图」区域 | 大厦文件读取能力 | 现接口未返回，若要展示需补独立后端读取 |
| `Blgfiles.btype=auditreport` | 大厦财务/审计 | 大厦页「大厦财务」 | 支付账务或大厦财务读取能力 | 只面向有权限的业主、租客、职员展示 |
| `Blgfiles.btype=mfinreport` | 大厦财务报告 | 大厦页「大厦财务」 | 支付账务或大厦财务读取能力 | 与审计报告同权限模型 |
| `POST /api/v1/integration/buildings/comments/` | 大厦服务、报修、意见 | 大厦页「意见提供/维修报修」 | `POST /api/v1/me/ismart/building-comments` | AJO 校验会员大厦及单位可见性后创建 iSmart `BuildingServiceCase`；上游返回 404 时才使用旧路径 |
| `GET /api/v1/integration/buildings/service-cases/` | 大厦服务、报修、意见 | 大厦页「意见提供/维修报修」 | `GET /api/v1/me/ismart/service-cases` | 固定读取当前会员的个案，不缓存动态状态、消息或附件 |
| `GET /api/v1/integration/buildings/service-cases/<case_id>/` | 大厦服务、报修、意见 | 大厦页「意见提供/维修报修」 | `GET /api/v1/me/ismart/service-cases/:caseId` | 读取当前会员可见的个案详情、消息及附件，不暴露内部讯息 |
| `POST /api/v1/external/building-access/` | 设备、门禁、安防 | 大厦页「智能门禁」 | 建议新增 `building_access` 或 `device_access` service；不要混入账号权限 `access_service` | 只展示用户有权限的门、二维码状态、近期记录 |
| `POST /api/v1/external/building-access/open-door/` | 设备、门禁操作 | 大厦页「智能门禁」 | 同门禁 service | 高风险操作，必须由 AJO 后端校验登录态、大厦、单位、门权限并记录审计 |
| `POST /api/v1/external/building-access/qrcode/` | 设备、门禁二维码 | 大厦页「智能门禁」 | 同门禁 service | 后端取二维码载荷，前端只负责渲染二维码，不持久保存 |
| `POST /api/v1/pos-payment-to-ismart` | 支付、账单、物业费 | 支付页，不放在大厦资料页主流程 | 现有 POS payment 模块已有基础，应继续放在 payment 路由与 service | 这是外部 POS/支付系统写入接口；AJO 会员端应继续用现有 `/me/payments/pos/*` 能力 |

## 建议的 AJO 页面落点

### 大厦页

适合承接：

- 大厦基本资料。
- 大厦资料文件。
- 申请表格。
- 平面图。
- 意见提供/维修报修。
- 智能门禁。
- 视像监控入口或门禁关联镜头摘要。

不建议承接：

- 物业费付款完整流程。付款应留在支付页，只能在大厦页提供简洁入口。
- 外部 POS 写入接口。该接口属于支付系统集成，不属于会员网页直接功能。

### 支付页

适合承接：

- 物业费账单。
- 欠费与已缴记录。
- 支付方式。
- 支付订单。
- 入账与核销状态。

现有项目已有 `payments-pos` 与后端 POS payment 相关能力，这部分应继续沿用现有 payment 模块，不要重复在大厦页另做一套。

### 通知/公告页

这份接口文档未包含 iBoard 通告接口。大厦公告后续应归入通知或大厦公告模块，不应与 `blg_cs` 意见提交混在一起。

### 设备/安防能力

门禁、二维码、开门、摄像头关联应作为设备和门禁能力处理。AJO Web 当前可以先做最小可用：

- 门列表。
- 是否有权限。
- 动态二维码。
- 远程开门。
- 最近开门记录。

摄像头播放、设备监测可以先只保留入口或摘要，等 iCCTV 接口明确后再接入。

## 后端接入建议

### 认证与权限

- AJO 对浏览器继续使用当前 JWT 登录态。
- 后端通过当前 AJO `user_id` 查 `UserIsmartAccount`，解析旧系统 `IsmartUserID` 与可见大厦、单位权限。
- 请求 iSmart 外部接口时，由 AJO 后端填充旧系统 `user_id`，浏览器不得提交该字段。
- 每个接口都应校验大厦可见性；涉及单位、门、账单时继续校验更细权限。

### API 形态

当前前端使用 AJO 语义接口，而不是复制旧接口路径：

| 能力 | AJO 前端接口建议 |
| --- | --- |
| 大厦资料与文件 | `GET /api/v1/me/ismart/building-info?building_id=` |
| 大厦通告 | `GET /api/v1/me/ismart/notices?building_id=` |
| 业主绑定申请 | `POST /api/v1/me/ismart/owner-binding-requests` |
| 授权用户 | `GET /api/v1/me/ismart/subaccounts?unit_id=`、`POST /api/v1/me/ismart/subaccounts/grant`、`POST /api/v1/me/ismart/subaccounts/revoke` |
| 大厦意见提交 | `POST /api/v1/me/ismart/building-comments` |
| 服务个案列表 | `GET /api/v1/me/ismart/service-cases?building_id=&status=` |
| 服务个案详情 | `GET /api/v1/me/ismart/service-cases/:caseId?building_id=` |
| 门禁摘要 | `GET /api/v1/me/ismart/building-access?building_id=` |
| 远程开门 | `POST /api/v1/me/ismart/building-access/open-door` |
| 生成门禁二维码 | `POST /api/v1/me/ismart/building-access/qrcode` |
| 物业费账单与支付 | 继续使用现有 `/api/v1/me/payments/pos/*` |

### 数据所有权

- 大厦资料、文件、门禁、意见若仍由旧后台维护，AJO 先按只读或受控转发处理。
- `blg_cs` 属于写入旧系统，需明确是否由 AJO 直接调用旧接口写入，还是由 AJO 后端写自己的服务请求表后同步到 iSmart。
- 门禁开门属于即时设备操作，AJO 不应本地模拟成功；必须以后端返回的设备结果为准。
- POS 付款写入已有独立状态机，继续保留在 payment 域。

## 推荐接入顺序

1. 先接大厦资料和表格文件：风险最低，能替换当前大厦页 mock 数据。
2. 再接意见提交：需要写入旧系统，但业务边界清晰。
3. 再接门禁摘要和二维码：需要处理敏感凭证，但不一定先开放远程开门。
4. 最后开放远程开门：需要设备环境联调、审计和失败态处理。
5. POS payment 不跟随这批大厦接口重做，只在现有支付模块里补齐缺口。

## 当前项目对应判断

- 前端 `web-new` 已有大厦页及会员中心，已通过 AJO 会员接口读取大厦资料、通告、财务、门禁、业主绑定与授权用户资料。
- 后端已有 `UserIsmartAccount`，可以作为 AJO 用户与 iSmart 用户、大厦、单位权限的映射基础。
- 后端已有 POS building 与 POS payment 能力，支付类接口不需要另建 `ismart payment` 模块。
- 后端已有受控 iSmart proxy service，按大厦、通知、成员与门禁语义维持接口边界，不把旧接口直接暴露给浏览器。

## 大厦资料首期数据流

目标：`web-new` 的「我的大厦」页面只显示当前登录用户已绑定、且有权限查看的大厦资料。

1. 用户在 AJO Web 登录后，前端只携带 AJO JWT 调用 AJO 后端。
2. 前端进入「我的大厦」页面时，请求 `GET /api/v1/me/ismart/building-info?building_id=`；通告、业主绑定与授权用户继续使用同一会员态接口组。
3. AJO 后端从 JWT 解析当前 AJO `user_id`，不接受前端提交旧系统 `user_id`。
4. AJO 后端查询 `UserIsmartAccount`，取得绑定的 `IsmartUserID`、`ClientBuildingPermissions` 与 `ClientBuildingFlatUnitsPermissions`。
5. AJO 后端只允许选择绑定权限内的大厦；如果前端没有传 `buildingId`，可默认取用户绑定的第一个可见大厦。
6. AJO 后端使用 `ISMART_INTEGRATION_API_BASE_URL` 请求 `/buildings/info/`，并以目标 `building_id` 作为服务端请求上下文；仅在上游读取路径不可用时使用 `ISMART_EXTERNAL_APP_API_BASE_URL` 旧路径回退。
7. iSmart 返回大厦基础资料、表格文件与大厦资料文件。
8. AJO 后端将旧字段整理成 AJO 稳定响应结构，过滤掉当前页面暂不展示或无权限展示的内容。
9. 前端把返回结果渲染到「我的大厦」的大厦资料区域；大厦下拉或切换项只展示用户已绑定的大厦。

当前环境变量：

```env
ISMART_EXTERNAL_APP_BASE_URL=https://ismart.ajoliving.com
ISMART_EXTERNAL_APP_API_BASE_URL=https://ismart.ajoliving.com/api/v1/external
ISMART_INTEGRATION_API_BASE_URL=https://ismart.ajoliving.com/api/v1/integration
ISMART_EXTERNAL_APP_TIMEOUT=10s
```
