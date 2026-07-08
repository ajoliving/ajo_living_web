# AGENTS.md

## 說明
- 本文件用於補充 `http_service` 子專案規範。
- 未特別說明之事項，沿用上層 [`AGENTS.md`](/Users/yangliu/Documents/Code/ajoliving_web/AGENTS.md)。
- 本文件及 `http_service` 目錄下所有 `.md` 檔案中均不得使用 emoji。
- 本規範適用於 `http_service` 專案下全部 Go 後端程式碼。

## 適用範圍
- `cmd/server`
- `internal/router`
- `internal/handler`
- `internal/service`
- `internal/model`
- `internal/middleware`
- `internal/config`
- `internal/database`
- `internal/errcode`
- `internal/logger`
- `internal/utils`

## 技術基線
- 後端統一使用 `Go + Gin + GORM + PostgreSQL`。
- 檔案儲存統一接入阿里雲 `OSS`。
- HTTP 服務負責認證、會員資料、列表發佈、聊天、上傳、生命周期與權限校驗。
- API 路徑統一使用 `/api/` 前綴。

## 業務域定位
- 後端是 `AJO Living` 物業管理主系統的 API，不是單一 marketplace 後端。
- 新增資料模型與 service 前，必須先判斷業務域：帳戶身份、物業大廈、單位與成員、通知站內信、支付賬務、設備安防、交易列表或管理端配置。
- 涉及業主、租客、職員、管理公司、大廈、單位關係的功能，必須在 service 層明確資料歸屬、可見性與權限校驗。
- 舊系統遷移只能作為流程與欄位參考；不得讓新 service 直接形成對舊前端或舊資料語義的長期耦合。
- 面向未來 POS、iBoard、iCCTV、iLock、Intercom、iSmart 等能力時，只建立當前功能必需的模型與 API，不提前鋪設未使用的抽象。
- 舊業務後台仍繼續使用時，AJO 後端優先提供受控讀取、身份映射、權限校驗與資料聚合，不默認接管舊後台的寫入流程。
- 共用資料庫時，必須明確 AJO 對每張舊表是唯讀、可寫還是同步後讀取；未確認前按唯讀處理。
- AJO service 不得直接把舊系統表結構原樣暴露給前端，必須轉成 AJO 穩定回應結構與權限語義。
- 舊系統資料寫入仍由舊後台負責時，AJO 只做讀取與展示；若確需寫入，必須先建立明確的所有權、交易邊界與回滾策略。

## 舊系統接入分層
- POS 接入歸入支付或物業費服務，優先提供賬單列表、付款狀態、付款紀錄與跳轉付款入口。
- iBoard 接入歸入通知或大廈公告服務，優先提供按大廈、身份、時間範圍過濾的公告讀取。
- iCCTV 接入歸入安防或設備服務，優先提供可見攝像頭列表、播放授權、狀態展示與權限校驗。
- iLock 接入歸入門禁或智能設備服務，優先提供門鎖狀態、權限展示與必要操作入口。
- Intercom 接入歸入對講或門禁聯動服務，Web 端只做適合網頁的展示或跳轉，App 端能力不得強塞到 Web 頁面。
- 對舊系統資料的存取優先使用獨立 service 檔案承載，例如 `payment_*`、`notice_*`、`security_*`、`access_*`，不要把所有整合邏輯塞進單一 `legacy_service.go`。

## 站內信與聊天模型原則
- 聊天與站內信屬於平台級通訊能力，不應永久綁定在二手交易或單一 marketplace 模型下。
- 一對一聊天可先按最小需求實作，但資料命名應保留會話、參與者、消息、來源模組與業務上下文的演進空間。
- 未來群聊應優先基於同一套 conversation / participant / message 概念擴展，避免另建一套割裂的 group chat 表與 API。
- 群聊功能落地前必須先定義 `conversation_type`、參與者身份、管理權限、邀請/退出/封存、未讀數與消息可見性。
- 公告與聊天必須分清：公告是單向發布，站內信是可追蹤消息，聊天是互動會話；API、模型與通知策略不得混淆。
- 未讀數、提醒與消息狀態應可被全站通知中心統一統計，不得在各業務模組中各自孤立計算。

## 專案架構規範

### 1. 分層結構

| 層級 | 目錄 | 職責 |
| --- | --- | --- |
| 入口 | `cmd/server` | 服務啟動、初始化、遷移、系統級裝配 |
| 路由 | `internal/router` | 路由註冊、中間件掛載、快取與限流策略編排 |
| 控制器 | `internal/handler` | HTTP 請求解析、基礎參數校驗、回應輸出 |
| 服務 | `internal/service` | 業務邏輯、資料庫讀寫、交易處理、第三方服務編排 |
| 模型 | `internal/model` | GORM 模型、分頁結構、基礎資料結構 |
| 中間件 | `internal/middleware` | 認證、限流、快取、上下文注入等橫切能力 |
| 配置 | `internal/config` | 配置讀取與配置物件管理 |
| 資料庫 | `internal/database` | 資料庫連線池與資料庫初始化 |
| 錯誤碼 | `internal/errcode` | 統一回應格式、錯誤碼、錯誤訊息 |
| 日誌 | `internal/logger` | 專案統一日誌入口 |
| 工具 | `internal/utils` | 通用且無業務狀態的工具函式 |

### 2. 依賴方向

```text
router -> handler -> service -> model/database
```

約束如下：
- handler 不得直接操作資料庫。
- handler 不得承載完整業務流程。
- service 不得直接輸出 HTTP 回應。
- service 不得依賴 `gin.Context`。
- utils 不得承載業務規則。
- middleware 不得寫資源級業務邏輯。

### 3. 實作原則
- 滿足業務目標的前提下，優先採用依賴更少、抽象更少、呼叫鏈更短的實作方式。
- 無明確複用價值時，不新增介面層、不新增包裝層、不做預設型通用設計。
- 新程式碼必須與現有專案分層和目錄結構保持一致。

## 檔案組織與命名規範

### 1. 包命名
- 包名使用全小寫英文單詞。
- 包名與目錄名保持一致。
- 包名不得使用底線、混合大小寫或無意義縮寫。

### 2. 檔案命名
- 檔名統一使用小寫字母與底線。
- 檔名應體現資源或職責。
- 同一資源的 handler、service、model 建議使用一致命名。

### 3. 型別命名
- 結構體使用大駝峰，例如 `AdminController`。
- 請求體使用 `XxxRequest`。
- 回應體使用 `XxxResponse`。
- 領域模型名稱與業務實體一致。

### 4. 介面命名
- 僅在確有抽象邊界、替換需求或測試隔離需求時定義介面。
- 新程式碼中的介面命名不得使用 `InterfaceXxx` 前綴，應直接表達職責。

## 註解規範

### 1. 檔案頂部多行註解
- `cmd/server`、`internal/router`、`internal/handler`、`internal/service`、`internal/middleware`、`internal/model`、`internal/config`、`internal/database` 下的檔案必須在 `package` 前加入多行註解。
- 使用 `/* ... */` 格式。
- 內容需說明檔案職責、主要內容與邊界。

### 2. 函式編號註解
- 所有導出函式、路由註冊函式、handler 主處理函式、service 主業務函式，以及邏輯較長或承擔關鍵路徑的私有函式，必須在定義上方加入編號註解。
- 格式統一採用 `// 1. 函式說明`。
- 同一檔案內按出現順序連續編號。

## Router 規範
- router 層僅負責 HTTP 路由組織。
- 只註冊路由、路由群組、中間件、快取與限流。
- 不在 router 中寫資料庫查詢與具體業務判斷。
- 路由路徑使用資源名詞，保持 REST 風格。

## Handler 規範
- handler 層只負責參數提取、參數綁定、基礎格式處理、呼叫 service、回應輸出。
- 基礎格式處理包含 `TrimSpace`、預設值修正、分頁參數邊界修正。
- 業務規則校驗放在 service，不得在 handler 中擴展為完整業務流程。
- 回應統一透過 `internal/errcode` 輸出。
- 新程式碼不得直接散落 `c.JSON(...)`。

## Service 規範
- 所有資料庫讀寫均在 service 層完成。
- 跨表寫入、狀態流轉、補償操作必須顯式控制交易。
- service 不得依賴 `gin.Context`、HTTP 狀態碼、`gin.H`。
- service 回傳業務結果與 `error`，由上層負責轉換回應。
- 原生 SQL 僅用於遷移場景或明確的效能優化場景，並應寫明原因。
- 禁止以 panic 處理運行期業務錯誤。

## Middleware 規範
- middleware 僅處理認證、鑑權、上下文透傳、限流、快取、審計等通用能力。
- 統一從請求中提取認證資訊並注入上下文。
- 中間件返回錯誤時，回應格式需盡量與 `errcode` 保持一致。
- 禁止在中間件中編排完整業務流程。

## Model 與資料庫規範
- model 主要承載 GORM 模型與基礎資料結構。
- 欄位名必須與業務語義一致。
- `json` tag、`gorm` tag 必須明確。
- 模型中不放置複雜業務流程。
- 所有資料庫訪問必須透過 service 層。
- 查詢、更新、刪除必須控制篩選條件，禁止不帶條件的風險操作。

## 回應與錯誤處理規範

### 1. 回應格式

```json
{
  "code": 100000,
  "message": "成功",
  "data": {}
}
```

### 2. 回應規則
- 成功統一使用 `errcode.Success`。
- 參數錯誤統一使用 `errcode.ParamError` 或 `errcode.FailWithMessage(...)`。
- 資源不存在統一使用 `errcode.NotFound`。
- 權限錯誤、限流錯誤、資料庫錯誤統一走 `errcode` 定義。
- 錯誤訊息必須可讀、穩定、可定位，不得暴露內部堆疊、SQL 細節或密鑰資訊。

### 3. 錯誤處理分工
- handler 負責錯誤到回應的轉換。
- service 負責回傳業務錯誤。
- middleware 負責處理鑑權與通用入口錯誤。
- config 與啟動階段可採用 fail-fast，但運行期業務程式碼不得濫用 panic。

## 日誌規範
- 專案統一使用 `internal/logger` 輸出日誌。
- 新增程式碼不得繼續擴散 `log.Printf` 風格。
- 日誌必須記錄關鍵上下文。
- 禁止打印密碼、JWT、密鑰、資料庫密碼與完整認證資訊。
- 同一錯誤只在最合適的一層記錄一次。

## 配置規範
- 所有環境變數統一在 `internal/config/config.go` 中讀取。
- 業務程式碼、中間件、handler、service 中不得直接使用 `os.Getenv(...)`。
- 必填配置使用顯式必填策略。
- 預設值僅用於本地可控場景。

## 檔案規模與複雜度
- 單個函式上限 `80` 行。
- 單個檔案上限 `400` 行。
- 同一檔案中若職責明顯分叉，必須拆分。
- 對於 `logger` 的包和 `errcode` 的包的封裝，按照業務判斷是否需要封裝即可。

## 測試與文件規範
- 新增業務邏輯時，至少補充正常路徑測試。
- 修復缺陷時，至少補充回歸測試。
- 修改 API 請求或回應結構時，必須同步更新介面文件。
- 修改錯誤碼、錯誤語義與錯誤回應方式時，必須同步更新錯誤碼文件。

## 提交前最低檢查

```bash
gofmt -w ./...
go test ./...
```

## 開發執行清單
- 路由是否僅承擔註冊職責。
- handler 是否僅承擔接入與回應職責。
- service 是否完整承載業務與資料庫邏輯。
- 回應是否統一走 errcode。
- 是否遵守檔案頂部多行註解規範。
- 是否遵守主要函式編號註解規範。
- 是否避免引入無明確收益的新抽象層。
- 是否同步補充測試與文件。
- 对于我的部署的话，需要参考我的/Users/yangliu/Documents/Code/ajoliving_web/doc copy/server-deployment-ai-prompt.md 文档，以及我的/Users/yangliu/Documents/Code/ajoliving_web/deploy-ajoliving.sh 脚本
- 部署不要破坏我的已经部署好的项目，可以先ssh链接上服务器后直接查看相关的服务器的配置文件等等，然后再部署，不要新加多余的垃圾

## 子目錄記憶
- `internal/AGENTS.md` 是後端內部分層與目錄責任的本地記憶入口；新增、移動、重命名 internal 能力或改變主域歸屬時同步更新。
- API 文檔、錯誤語義、模型契約或 AI 文件生成規範受影響時，同步檢查 `doc/api.md` 與 `../docs/PROMPT_INDEX.md`。

## 變更日誌
2026-07-08: 補充後端 internal 目錄記憶入口，連接 API 契約與 prompt/context 索引。

[PROTOCOL]: When backend internal ownership or prompt-facing API contract changes, check `internal/AGENTS.md`, parent `../AGENTS.md`, and `../docs/PROMPT_INDEX.md`.
