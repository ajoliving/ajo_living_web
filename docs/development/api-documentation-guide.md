# AI 接口文档生成规范

> 本文档及本仓库所有 .md 文件中均不得使用 emoji。

## 1. 用途

本规范用于指导 AI 生成与 `/Users/yangliu/Documents/Code/intercom/ilock/tem.md` 风格一致的 HTTP 接口文档。

目标不是生成 Swagger/OpenAPI，而是生成一份给人直接阅读的 Markdown 接口文档，且文档结构、章节顺序、接口编号方式、示例代码块风格都尽量保持一致。

---

## 2. 总体要求

- 输出格式必须是 Markdown。
- 文档必须先输出“完整 API 接口列表”，再输出“详细接口参数/响应/脚本”。
- 接口总表与详细接口章节的编号、顺序、路径、方法必须完全一致。
- 每个接口都必须包含：
  - 接口标题
  - 简介
  - 请求参数
  - 响应参数
  - `curl` 测试示例
  - `Powershell` 测试示例
- 除非接口确实无鉴权，否则测试示例必须带 `Authorization: Bearer $token`。
- 请求与响应示例必须使用 `json` 代码块。
- `curl` 必须使用 `bash` 代码块。
- `Powershell` 必须使用 `powershell` 代码块。
- JSON 示例允许写注释，例如 `// 必填`、`// 可选`，以增强可读性。

---

## 3. 固定文档结构

AI 生成文档时，必须严格使用下面这套结构，不要随意改标题名、章节名、顺序名。

```md
# {系统名} - HTTP Service API 文档

## API 接口列表

| 编号 | 接口 | 方法 | 简介/功能 | 权限 |
| --- | --- | --- | --- | --- |
| 1 | /api/example | GET | 查询示例列表 | 管理员 |
| 2 | /api/example | POST | 创建示例 | 管理员 |

---

## 详细接口参数/响应/脚本

### 1. /api/example [GET]
- **简介**: 查询示例列表
- **请求参数**
```json
{
  "id": 1
}
```
- **响应参数**
```json
{
  "success": true,
  "data": []
}
```
- **Curl测试**
```bash
curl -X GET "http://127.0.0.1:8080/api/example?id=1" \
  -H "Authorization: Bearer $token"
```
- **Powershell测试**
```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/example?id=1" -Headers $headers
```

---

### 2. /api/example [POST]
- **简介**: 创建示例
- **请求参数**
```json
{
  "name": "string", // 必填
  "remark": "string" // 可选
}
```
- **响应参数**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "string",
    "remark": "string"
  }
}
```
- **Curl测试**
```bash
curl -X POST "http://127.0.0.1:8080/api/example" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "demo",
    "remark": "test"
  }'
```
- **Powershell测试**
```powershell
$headers = @{"Authorization"="Bearer $token"; "Content-Type"="application/json"}
$body = @{name='demo';remark='test'}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/example" -Method POST -Headers $headers -Body $body
```

---
```

---

## 4. 接口总表生成规则

### 4.1 表头固定

接口总表表头必须固定为：

```md
| 编号 | 接口 | 方法 | 简介/功能 | 权限 |
| --- | --- | --- | --- | --- |
```

### 4.2 表格内容规则

- `编号`：从 `1` 开始递增，不允许跳号。
- `接口`：只写路径，不带域名，例如 `/api/admin`。
- `方法`：必须全大写，例如 `GET`、`POST`、`PUT`、`DELETE`。
- `简介/功能`：一句话描述接口用途，长度控制在 8 到 20 个中文字符左右。
- `权限`：使用明确权限描述，例如：
  - `无`
  - `管理员`
  - `管理员 / 第三方应用`
  - `设备`

### 4.3 排序规则

- 总表顺序必须与详细接口章节顺序完全一致。
- 同一路径下存在多个方法时，按实际接口定义顺序输出，不要擅自重排。

---

## 5. 详细接口章节生成规则

### 5.1 标题格式固定

每个接口标题必须使用下面格式：

```md
### {编号}. {接口路径} [{HTTP方法}]
```

例如：

```md
### 7. /api/admin [DELETE]
```

### 5.2 固定字段顺序

每个接口的详细说明，字段顺序必须固定为：

1. `- **简介**: ...`
2. `- **请求参数**`
3. 请求 JSON 示例
4. `- **响应参数**`
5. 响应 JSON 示例
6. `- **Curl测试**`
7. curl 示例
8. `- **Powershell测试**`
9. Powershell 示例
10. `---`

### 5.3 简介规则

- 每个接口都应有简介。
- 简介必须是一句话，直接说明接口用途。
- 不要写空泛描述，例如“这是一个接口”“用于处理数据”。

### 5.4 请求参数规则

- 统一使用标题 `- **请求参数**`，不要混用 `请求`、`入参`、`Body`。
- 示例使用 `json` 代码块。
- 如果是 `GET` 接口：
  - 仍然使用 JSON 对象表达查询参数示例。
  - 如果没有查询参数，可写 `{}` 或注明无请求体。
- 如果是 `POST` / `PUT` / `DELETE`：
  - 默认给出 JSON Body 示例。
- 如果接口存在路径参数，例如 `/api/building/{building_id}`：
  - 标题中保留 `{building_id}`。
  - `curl` 与 `Powershell` 示例中使用具体值，例如 `1`。
- 字段说明写在 JSON 行尾注释中，例如：

```json
{
  "building_id": 1, // 必填
  "name": "string", // 必填
  "remark": "string" // 可选
}
```

### 5.5 响应参数规则

- 统一使用标题 `- **响应参数**`，不要混用 `响应`、`返回值`。
- 响应示例默认使用统一包裹格式：

```json
{
  "success": true,
  "data": {}
}
```

- 如果接口返回数组，则 `data` 应写为数组形式。
- 如果接口返回对象，则尽量给出完整字段示例。
- 删除接口可返回：

```json
{
  "success": true,
  "data": { "deleted": true }
}
```

### 5.6 分隔规则

- 每个接口章节结束后都要加一条：

```md
---
```

- 最后一个接口后也允许保留 `---`，风格上优先与前文一致。

---

## 6. 示例值生成规则

AI 生成示例值时，必须遵守以下约束：

- `id`、`building_id`、`device_id`、`nvr_id` 等 ID 字段默认使用整数，例如 `1`。
- 名称类字段默认使用 `"string"` 或带语义的示例值，例如 `"admin"`、`"一号楼"`。
- 布尔值必须使用 `true` / `false`。
- 时间字段使用字符串，例如 `"2025-11-24T10:00:00Z"`。
- URL 字段使用看起来真实的值，例如 `"http://example.com"`、`"rtsp://192.168.1.100:554/stream1"`。
- 数组字段至少提供一个示例元素。
- 对象数组要写完整结构，不要只写 `[...]`，除非确实无法确定结构。
- 未知字段不要乱编业务含义，宁可写成 `"string"`、`1`、`true` 这类中性示例值。

---

## 7. Curl 测试生成规则

### 7.1 固定要求

- 每个接口都必须有 `- **Curl测试**`。
- 代码块语言必须为 `bash`。
- 默认基地址使用：

```text
http://127.0.0.1:8080
```

### 7.2 鉴权规则

- 无权限接口：不加 `Authorization` 头。
- 需要鉴权的接口：必须加

```bash
-H "Authorization: Bearer $token"
```

### 7.3 JSON Body 接口

对于 `POST`、`PUT`、`DELETE` 等带 JSON Body 的接口，推荐固定格式：

```bash
curl -X POST "http://127.0.0.1:8080/api/example" \
  -H "Authorization: Bearer $token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "demo"
  }'
```

### 7.4 GET 接口

对于 `GET` 接口：

- 查询参数直接拼到 URL 上。
- 不要用 `-d` 传 GET 参数。

示例：

```bash
curl -X GET "http://127.0.0.1:8080/api/admin?id=1" \
  -H "Authorization: Bearer $token"
```

### 7.5 路径参数接口

例如：

```text
/api/bind/building-orangepi/{building_id}
```

生成 `curl` 时必须替换为具体值：

```bash
curl -X GET "http://127.0.0.1:8080/api/bind/building-orangepi/1" \
  -H "Authorization: Bearer $token"
```

---

## 8. Powershell 测试生成规则

### 8.1 固定要求

- 每个接口都必须有 `- **Powershell测试**`。
- 代码块语言必须为 `powershell`。

### 8.2 Header 写法

无 Body 的鉴权请求：

```powershell
$headers=@{"Authorization"="Bearer $token"}
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/example?id=1" -Headers $headers
```

带 Body 的鉴权请求：

```powershell
$headers = @{"Authorization"="Bearer $token"; "Content-Type"="application/json"}
$body = @{name='demo'}|ConvertTo-Json
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/example" -Method POST -Headers $headers -Body $body
```

### 8.3 嵌套对象规则

如果请求体中有数组或嵌套对象，必须使用：

```powershell
ConvertTo-Json -Depth 10
```

避免嵌套数据被截断。

---

## 9. AI 禁止事项

AI 生成文档时，禁止出现以下问题：

- 不要把 Markdown 文档改成 OpenAPI、YAML、HTML。
- 不要省略接口总表。
- 不要省略接口编号。
- 不要把详细章节改成二级标题或四级标题。
- 不要把 `curl` 和 `Powershell` 只保留一种。
- 不要把请求参数写成表格，本规范要求用 JSON 示例。
- 不要把响应参数写成“字段说明表”为主，本规范要求以 JSON 示例为主。
- 不要省略 `---` 分隔线。
- 不要随意改动表头名，例如把 `简介/功能` 改成 `说明`。
- 不要在标题中带域名。
- 不要在无权限接口里强行添加 `Authorization`。
- 不要在 GET 示例里使用 `-d` 传参。

---

## 10. 可直接给 AI 的执行指令

下面这段可以直接作为提示词给 AI 使用：

```md
你现在要生成一份 HTTP Service API Markdown 文档。

必须严格遵守以下规则：

1. 输出结构必须与现有模板文档一致：
   - 文档标题
   - `## API 接口列表`
   - Markdown 总表
   - `---`
   - `## 详细接口参数/响应/脚本`
   - 每个接口的详细章节

2. 每个接口详细章节必须严格包含：
   - `### {编号}. {路径} [{METHOD}]`
   - `- **简介**: ...`
   - `- **请求参数**`
   - JSON 请求示例
   - `- **响应参数**`
   - JSON 响应示例
   - `- **Curl测试**`
   - bash curl 示例
   - `- **Powershell测试**`
   - powershell 示例
   - `---`

3. 总表字段固定为：
   - `编号`
   - `接口`
   - `方法`
   - `简介/功能`
   - `权限`

4. 所有接口都必须与总表顺序一致。

5. 请求与响应都优先用 JSON 示例表达，不使用字段表格替代。

6. 需要鉴权的接口：
   - curl 必须带 `-H "Authorization: Bearer $token"`
   - powershell 必须带 `$headers=@{"Authorization"="Bearer $token"}`

7. 带 JSON Body 的接口：
   - curl 必须带 `Content-Type: application/json`
   - powershell 必须使用 `ConvertTo-Json`

8. GET 接口参数放到 URL query 中，不要在 curl 中使用 `-d`。

9. 路径参数在标题中保留 `{id}`，但在测试示例里替换成具体值，例如 `1`。

10. 输出风格必须接近传统后端接口文档，不要自由发挥 UI 文案，不要改结构，不要加额外章节。
```

---

## 11. 建议用法

当你把接口信息交给 AI 时，建议至少提供以下内容：

- 接口路径
- HTTP 方法
- 接口简介
- 权限要求
- 请求字段
- 响应字段
- 是否需要 Bearer Token

只要输入信息完整，AI 就应当按本规范产出一份与 `tem.md` 风格高度一致、并且同时包含 `curl` 与 `Powershell` 示例的 Markdown 接口文档。
