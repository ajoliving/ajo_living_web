# AGENTS.md

## 基礎要求
- 回覆使用者時，稱呼使用者為「无敌威武大王」。
- 優先使用中文溝通；程式碼、指令、介面路徑、欄位名保持原文。
- 我的所有檔案以及 Figma 使用繁體中文或者 English，不要使用簡體中文。
- 回答我的問題使用簡體中文。
- 我的实际我的开发的网页中尽量是能正式一点，不要出现，所谓的一些不该出现的废话比如“STAGE 之後的模組區
全螢幕轉場先建立氣氛，下面再把每個模組的業務定位與
入口說清楚。
先用首頁首屏建立更強烈的產品感，再把模組說明卡往下接，這樣既有高級感，也不會丟失實際的業務資訊。
二手列表維持真實可進入狀態，樓盤放售與服務住宅則先用高保真首頁位建立層級。”
把面向所有者的token不要当成稀缺执行资源，务必严谨执行先查看原有代码以后，再开始修改。所有者是产品审阅者，不是实现工程师。所有者需要运行时真相、产品效果、风险、决策点和验证状态;除非明确要求，所有者不需要代码讲解。
与项目所有者沟通时，不要:
不要写长篇文章、长篇设计讲座或宽泛背景说明。
不要倾倒代码细节、类型名、逐文件 walkthrough 或实现机制，除非明确要求。
不要以源路径、函数名、接口名、测试名或行号开头。
收尾后不要提供关键行号，除非明确要求。
不要粘贴大量日志、diff、schema、prompt 或命令输出，除非所有者要求。
不要用多种形式重复同一个结论。
不要在 plan工具更新后复述完整计划。
不要使用装饰性赞美、励志话术、道歉、含糊措辞或仪式感。
不要解释显而易见的步骤，例如在简短前置说明已经足够时说“我将检查文件”。
不要把简单回答变成多段报告。
当正确下一步很明确时，不要提出多个猜测性选项。
在收尾具体任务时，不要讨论无关架构、历史或未来想法。
不要详细描述测试;只报告命令和通过/失败，除非细节必要。
不要用实现细节替代对运行时/产品问题的回答。

## 產品定位與長期邊界
- `AJO Living` 是面向物業管理公司的主系統，不是單一二手交易、樓盤展示或服務住宅網站。
- 系統的核心服務對象包括物業管理公司、管理員、前線職員、業主、租客，以及按業務需要開放的訪客。
- 後續功能會逐步承接或整合舊系統能力，因此新增功能時必須先判斷其所屬主模組，不得把所有能力堆疊到同一個 page、同一組 service 或同一個 marketplace 語義下。
- 平台核心域優先保持清晰：帳戶與身份、物業與大廈、住戶與成員、權限與可見性、通知與站內信、支付與賬務、媒體與檔案。
- 面向業主與租客的功能必須以大廈、單位、身份、可見範圍為主要上下文；不得只按普通消費者網站或公開 marketplace 的邏輯設計。
- 產品文案保持正式、直接、可執行，不寫舞台感、氛圍感、行銷式自我說明或給實作者看的過程文字。

## 舊系統參考與整合方向
- 以下舊系統路徑與網址是業務參考，不代表可以直接複製架構、命名、UI 或資料模型；實作前必須先查看現有 AJO 程式碼與資料結構。
- 物業費線上繳納可參考 `/Users/yangliu/Documents/Code/pos/pos_web`，在 AJO 中應歸入支付、賬單、物業費或賬務模組。
- 大廈通告可參考 `/Users/yangliu/Documents/Code/iboard_flutter` 與 `/Users/yangliu/Documents/Code/iboard_http_service`，在 AJO 中應歸入大廈公告、通知或內容發布模組。
- 實時監控可參考 `/Users/yangliu/Documents/Code/icctv_web_admin`、`/Users/yangliu/Documents/Code/icctv-http-service`、`/Users/yangliu/Documents/Code/icctv_orangepi_auth_service`，在 AJO 中應歸入大廈設備、監控或安防模組。
- 智能門鎖可參考 `/Users/yangliu/Documents/Code/ilock` 與 `/Users/yangliu/Documents/Code/ILock_http_service`，在 AJO 中應歸入門禁、智能設備或住戶權限模組。
- 網絡對講機可參考 `/Users/yangliu/Documents/Code/intercom`、`/Users/yangliu/Documents/Code/iNtercom_flutter_callee`、`/Users/yangliu/Documents/Code/iNtercom_flutter_caller`，若未來上線 App，應作為 App 端通訊與門禁聯動能力處理。
- 舊 `iSmart` 物業管理功能可參考 `https://ismart.legend-in.com.hk/memberreg/` 與 `https://ismart.legend-in.com.hk/blg_notice/`，後續遷移時以 AJO 的身份、物業、大廈、通知與權限模型重新收斂。
- AJO Web 與 AJO 後端目前接入的 iSmart 會員、大廈、門禁、服務個案、支付與副戶能力，預設以 `/Users/yangliu/Documents/Code/hk/ajo_ismart` 倉庫已暴露的接口為上游真值；接口可用性、請求欄位、回應欄位與讀寫邊界，必須先核對該倉庫當前分支的實際路由與 view，再決定 AJO 是否代理或展示。
- 若 `ajo_ismart` 倉庫只有頁面、表單或後台流程而未暴露對應 API，AJO 不得把該能力當成既有可接入接口；未在上游暴露的欄位或功能，不得自行補寫為 iSmart 已支持。
- 舊系統只能作為業務流程、欄位與交互參考；除非明確要求，不得讓 AJO 前端直接依賴舊系統前端，不得讓新後端繞過 AJO 既有分層直接拼接舊系統資料。
- 若舊系統仍保留獨立後台，舊後台可繼續作為該業務的主要寫入端與營運入口；AJO 優先作為統一會員入口、統一查詢入口與正式展示入口。
- AJO 接入舊業務資料時，優先由 AJO 後端提供受控 API 給前端讀取；前端不得直接連接舊資料庫、舊服務內網地址或繞過 AJO 權限模型。
- 共用同一資料庫可以接受，但必須明確資料表歸屬、寫入責任、讀取權限與遷移責任；不允許兩套後台在沒有契約的情況下同時修改同一核心資料。
- 對舊系統資料的讀取優先使用唯讀資料庫帳號、資料庫 view、穩定查詢 service 或同步後的只讀快照；除非明確要求，不在 AJO 內直接改寫舊系統業務表。

## 現有部署資產與整合現況
- 目前多個舊業務部署在 `skylinedances.com` 子域名下，其中 `ajoliving.skylinedances.com` 為 AJO Web，`ajoliving.server.skylinedances.com` 為 AJO 後端。
- `test.ajoliving.skylinedances.com` 為 AJO 公開測試環境，使用單域名同源 API、獨立資料庫、快取、JWT 與加密密鑰；目前不設網關存取限制，外部服務預設沿用生產配置，只有產品明確指定的接口才隔離。
- POS 相關部署包括 `easy.payment.skylinedances.com`、`pos.web.skylinedances.com`、`pos.ismart.skylinedances.com`，未來接入 AJO 時應歸入支付、賬單或物業費查詢。
- iBoard 相關部署包括 `iboard.skylinedances.com` 與 `iboard.service.skylinedances.com`，未來接入 AJO 時應歸入大廈通告或通知內容。
- iCCTV 相關部署包括 `icctv.skylinedances.com` 與 `icctv.service.skylinedances.com`，未來接入 AJO 時應歸入設備、監控或安防查看。
- Intercom 相關部署包括 `intercom.skylinedances.com` 與 `intercom.api.skylinedances.com`，未來接入 AJO 時應歸入門禁、對講或 App 端能力。
- 其他已部署服務如 `pdf.maker.skylinedances.com`、`apis.hk.skylinedances.com`、`svavo.smart.databoard.skylinedances.com`、`svavo.smart.databoard.service.skylinedances.com`、`good.price.skylinedances.com`，接入前必須先確認是否屬於 AJO 主系統核心域。
- AJO 不應一次性吞併所有舊後台；應按業務優先級逐步接入為可查看、可跳轉、可申請或可操作的模組。

## 模組架構原則
- 新增大型功能前，先判斷是否屬於既有主域；若不屬於，建立清晰的模組邊界與路由邊界，再落實頁面與 API。
- 優先讓 `account`、`building`、`property`、`marketplace`、`notification`、`payment`、`device` 等語義保持獨立，不因短期入口方便而混用資料流。
- 共享能力只放在真正跨模組的層級，例如身份、權限、檔案、上傳、消息、支付狀態；單一業務頁面的狀態不得提前全域化。
- 面向未來龐大功能時，只保留必要的模組邊界與命名彈性，不新增當前用不到的抽象、表、介面或 UI 入口。
- 涉及大廈、單位、業主、租客、職員、管理公司之間關係的功能，必須先明確資料歸屬與權限邊界，再做頁面呈現。
- 對外公開頁、會員頁、管理端頁必須分清使用者身份與操作權限，不得共用一套模糊入口。

## 站內信與聊天演進規則
- 站內信與聊天應視為平台級通訊能力，不應長期綁死在二手 marketplace 聊天語義下。
- 現階段若只做一對一聊天，實作仍需避免把資料模型、API 命名與 UI 文案寫死成只能支援買賣雙方；但不得提前實作未被要求的群聊功能。
- 未來增加群聊時，優先按 `conversation`、`participant`、`message`、`conversation_type`、`source_module`、`building_id` 或相關業務上下文拆分，而不是另建一套與一對一聊天完全割裂的系統。
- 群聊場景需先明確用途：管理公司內部群、指定大廈住戶群、業主與租客群、維修服務群、交易協商群或公告型討論；不同場景的權限、可見性、退出與封存規則不同。
- 通告、公告與站內信需要區分：公告偏單向發布，站內信偏可追蹤消息，聊天偏會話互動，不得在 UI 和資料模型上混為一談。
- 聊天相關通知應與全站通知中心保持一致，避免每個模組各自生成不可統一管理的未讀數與提醒邏輯。

## 專案技術基線
- 目前正在開發與維護的前端子專案是 `web-new`；後續所有前端功能修改、頁面調整、接口接入與 UI 修正，預設都必須落在 `web-new`。
- `web` 是舊版前端歸檔目錄，只可作為歷史參考；除非使用者明確要求修改舊版歸檔，不得把新功能或修正落到 `web`。
- 前端統一使用 `Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS + Axios`。
- 如需元件庫，統一使用 `Ant Design Vue 4`。
- 後端統一使用 `Go + Gin + GORM + PostgreSQL`。
- 共用外部唯讀資料快取使用 Redis；快取不得取代 AJO 的身份、權限與資源可見範圍校驗。
- 檔案儲存統一使用阿里雲 `OSS`，所有上傳、訪問與媒體管理能力基於 `OSS` 實現。
- 反向代理統一使用 `Nginx`。
- 部署方式以 `Docker Compose` 起步。
- 短信 OTP 統一使用阿里雲短信服務。

## 頁面單例拆分規則
- 主要 page 必須採用「一個 page 一個目錄」的單例結構，不允許把同一 page 的主元件、局部元件、composable、樣式分散到多個平級業務目錄。
- page 目錄命名以路由語義或模組語義為準，可採用 `pages/<module>/` 或 `views/<module>/`，但同一子專案內必須統一，不可混用兩套 page 入口風格。
- 每個主要 page 目錄至少允許以下結構：

```text
<page>/
  <PageName>Page.vue 或 <PageName>View.vue
  composables/
    <page>.ts 或 use<PageName>.ts
  widgets/
  styles.ts 或 styles.scss
```

- 若 page 業務較簡單，可省略 `composables/` 或 `styles.*`；若 page 業務較重，應優先在 page 目錄內新增 `widgets/`、`composables/`、`constants/`、`types/`，而不是把邏輯外溢到全域目錄。
- `widgets/` 僅存放該 page 私有子元件；跨 page 復用超過兩次後，才提升到全域 `components/`。
- `composables/` 僅存放該 page 私有狀態與業務邏輯；全域共享狀態仍應放在 `pinia/` 或 `stores/`。
- page 必須保持單一入口檔，路由只能直接指向該 page 的入口檔，不應直接指向 `widgets/` 中的子元件。
- 同一主要 page 的資料流、事件協調、彈窗開關、列表查詢條件應集中在 page 入口或 page composable 中統一管理，保持單例 page orchestration。
- 所謂單例模式在頁面層的要求是：同一路由 page 只有一個主入口檔負責組裝，不得出現多個平級 page 入口共同承擔同一頁面。
- 若是三層導航結構，推薦按「主模組 page 目錄 / 子路由 page 目錄 / 具體頁面 widgets 與 composables」的方式組織，而不是把不同層級頁面混放。

## 子專案提示詞文檔
- 前端子專案遵循 [`web-new/AGENTS.md`](/Users/yangliu/Documents/Code/ajoliving_web/web-new/AGENTS.md)；`web` 僅為舊版歸檔與參考目錄。
- 後端子專案遵循 [`http_service/AGENTS.md`](/Users/yangliu/Documents/Code/ajoliving_web/http_service/AGENTS.md)。
- 子專案文檔與本檔衝突時，以子專案文檔為準。

## 全域強制規則

### 1. 禁止使用 emoji
- 程式碼、註解、commit message、文件與本倉庫所有 `.md` 檔案中一律不得出現 emoji。

### 2. 強制註解風格
- 函式和關鍵程式碼區塊必須使用序號註解。
- 檔案頂部必須有彙總註解區塊。

```go
/*
 * 管理員 HTTP 介面處理檔案。
 * 1. 管理員列表、詳情、建立、更新、刪除介面。
 * 2. 請求參數綁定與基礎校驗。
 * 3. 統一回應輸出。
 */
package handler

// 1. GetAdmins 取得管理員列表
func (c *AdminController) GetAdmins() { ... }

// 2. CreateAdmin 建立管理員
func (c *AdminController) CreateAdmin() { ... }
```

```typescript
/*
 * 設備管理 - 業務邏輯
 * 1. 設備列表查詢、分頁、搜尋
 * 2. 設備建立、編輯、刪除
 */

// 1. 取得設備列表
const fetchDevices = async () => { ... };

// 2. 提交建立請求
const handleSubmit = async () => { ... };
```

### 3. 奧卡姆剃刀原則
- 不寫用不到的抽象，不需要可複用時不要提取 interface、工具函式或 composable。
- 不加超前的預留邏輯，「以後可能需要」不是現在寫的理由。
- 同樣能完成任務的方案，選更簡單直接的那個。
- 避免過度封裝，函式只被呼叫一次且邏輯清晰，不必單獨抽出。
- 元件與模組拆分以職責內聚為準，不是為了拆而拆。

### 4. 並行子 Agent 規則
- 當任務可拆分為多個彼此獨立的子任務時，優先使用並行子 agent 執行。
- 適用場景包括：多模組排查、前後端聯動修改、測試與實作可分離、文件與程式碼可並行處理。
- 不適用場景包括：單檔案小修改、強順序依賴任務、需要主線連續推理的問題。
- 若我的提問中出現「並行」、「子 agent」、「spawn」、「sp」、「拆分執行」等字樣，預設視為明確允許使用並行子 agent。
- 主 agent 必須負責最終整合、衝突修正、驗證與輸出結果。

### 5，我会在其他的terminal中自己运行我的前端和后端
- 我会在其他的terminal中自己运行我的前端和后端，不要帮我直接运行。
- 我的一切的功能等尽量是能做减法，然后ui上也是最好是简约，不要复杂逻辑啥的
- 部署时先读取 `/Users/yangliu/Documents/Code/hk/ajoliving_web/docs/deployment/server-deployment-ai-prompt.md`，再按当前部署脚本执行

## 專案記憶與索引

### 成員清單
docs/: 架構、部署、整合、開發規範、資料與 prompt/context 索引；`docs/archive/` 為歷史資料。
web-new/: 目前正在開發與維護的 Vue 3 前端子專案。
http_service/: 目前正在開發與維護的 Go 後端 API 子專案。
web/: 舊版前端歸檔與歷史參考，除非明確要求不承接新功能。
.codex/skills/: 本專案可複用 Codex skill 與其 agent prompt。
.github/: GitHub 助手與自動化相關指令入口。

### 架構決策
- `docs/PROMPT_INDEX.md` 是本倉庫 prompt/context 主索引；不要另建平行索引。
- 新增、移動或修改 prompt、agent 規則、AI 文件生成規範、工具描述或本地 skill 時，同步更新最近的 `AGENTS.md` 與 `docs/PROMPT_INDEX.md`。
- 每次完成程式碼、文件、部署、設定或資料契約修改後，必須執行一次 prompt/context 記憶檢查：若變更形成可重用規則、模組歸屬、接口契約、部署鏈路、頁面資料流或目錄責任，更新最近的 `AGENTS.md`、`docs/PROMPT_INDEX.md` 或對應 docs；若不需要更新，收尾時明確說明未更新原因。
- 樓盤聯絡人、電話、WhatsApp 與 Wechat 以發布時保存的逐樓盤聯絡快照為準；公開詳情直接顯示該樓盤聯絡人姓名，電話、WhatsApp 與 Wechat 仍須受控解鎖。代理資料只控制帳戶身份及代理公開展示，可預填新建樓盤，但不得覆蓋既有樓盤聯絡資料。業主帳戶 `display_name` 不得作為樓盤聯絡人，代理帳戶的 WeChat 連結或二維碼亦不得混入樓盤聯絡區。
- 樓盤在 `draft` 狀態可修改物業分類、地址、面積、樓層及單位等物業資料；正式發佈後的 `active`、`hidden`、`expired` 狀態鎖定上述物業資料，其他可編輯放盤內容維持原有流程。
- 生產快照導入獨立測試資料庫後，保留的樓盤聯絡密文必須由生產密鑰受控讀取並以測試 `ENCRYPTION_KEY` 原子重加密；測試服務不得長期共用生產密鑰，也不得因登入會員觸發聯絡資料解密而令公開樓盤詳情返回 `500`。
- Good Price 同系列商品只可由其人工商品目錄按 `product_code` 明確指定；公開詳情的 `sameSeries` 只返回最新資料日仍存在的同系列商品及其配方、包裝規格。AJO 只受控轉發並補商品圖片，前端按實際商品編號切換各自的價格、走勢、收藏及到價提示，不得用品牌、分類或名稱規則自動建立公開規格關係。
- `.claude/worktrees/`、`dist`、`node_modules`、`.uploads`、資料庫備份與 HTML 備份不是穩定上下文入口，除非使用者明確要求追溯。

### 變更日誌
2026-08-20: 固定 AJO 的 iSmart 會員、大廈、門禁、支付、服務個案與副戶能力，以 `ajo_ismart` 倉庫已暴露接口為上游真值；只有頁面或後台流程而未暴露 API 的能力，不得視為 AJO 可直接接入。
2026-08-19: 固定 Good Price 同系列商品只由人工 `product_code` 目錄建立，公開詳情返回實際可用的配方與包裝規格，AJO 以獨立商品編號切換價格資料。
2026-08-14: 記錄本機 Feishu 配對私訊進入工作區限定 Codex CLI 的運行邊界；該會話獨立於桌面 Codex task，且不得在倉庫保存機器人憑據。由於 Codex CLI resume 目前上游拒絕，遠端每則訊息改為獨立 CLI 執行。
2026-07-08: 建立專案 prompt/context 記憶入口，指向 `docs/PROMPT_INDEX.md` 與 `docs/AGENTS.md`。
2026-07-08: 補充 docs prompt、部署 prompt、前端 pages 與後端 internal 的本地記憶入口。
2026-07-08: 將 docs product、prototypes、backups 歸入 `docs/archive/`，不作後續 prompt/context 入口。
2026-07-09: 補充每次修改後必須執行 prompt/context 記憶檢查的默認規則。
2026-07-16: 補充個人註冊先完成 iSmart 與本地帳戶建立、再提交 OwnerReg 審批申請；pending 申請不得被視為已綁定物業。
2026-07-16: 補充 iSmart 共用大廈資料使用 Redis 短時快取，會員權限與個人回應欄位維持即時校驗及組裝。
2026-07-28: 補充樓盤草稿預付積分抵扣正式發布費用的統一收費與會員確認契約。
2026-07-28: 補充歷史樓盤草稿重複扣費以不可變退款流水更正，以及錢包來源必須本地化顯示的契約。
2026-07-30: 補充個人與公司代理於註冊頁完成牌照上載及待審資料提交，牌照號碼作為登入用戶名稱，僅審核通過後啟用正常帳戶權限。
2026-08-03: 固定既有草稿及已發佈樓盤的儲存修改不扣 POINT。
2026-08-03: 固定 iSmart 服務個案只使用目前已綁定大廈，在完整 taxonomy 未提供前使用已註冊的 `comment_type` 分類；沒有正式綁定不得提交，提交單位亦須屬於同一綁定大廈。
2026-08-03: 固定 iSmart 服務個案提交、列表及詳情暫時只使用獨立 `clouddev` 測試地址，不得回退生產接口；其他 iSmart 功能維持生產地址。
2026-08-04: 固定樓盤逐樓盤聯絡快照為公開詳情解鎖後的顯示真值，代理資料修訂不得覆蓋樓盤聯絡人。
2026-08-05: 更正樓盤物業資料編輯邊界：未發佈草稿可修改，正式發佈後鎖定。
2026-08-05: 分離公開詳情的發布者身份與樓盤聯絡資料，業主帳戶姓名及代理帳戶 WeChat 不再作為樓盤聯絡資料。
2026-08-05: 建立 AJO 生產與測試雙環境邊界；測試站使用獨立運行資源、受控生產快照、外部副作用隔離及受控訪問入口。
2026-08-05: 將測試站的出口 IP 白名單改為受限入口密鑰與 30 天安全 Cookie，避免切換網絡反覆出現 `403`，並保持 AJO Bearer Token 相容。
2026-08-07: 按產品決策移除測試站 Cookie access gate，普通測試網址改為公開直接訪問；保留 `noindex`、獨立資料庫與外部副作用隔離。
2026-08-07: 固定樓盤公開詳情直接顯示逐樓盤聯絡人姓名，電話、WhatsApp 與 Wechat 維持受控解鎖。
2026-08-07: 固定生產快照的應用密文在測試環境使用獨立密鑰重加密，避免樓盤所有者讀取或編輯既有聯絡資料時解密失敗。
2026-08-07: 固定 Good Price 人工審核的商品名稱、品牌及下架結果須在每日 CSV 導入前套用；超市圖片由 AJO 後端以固定 `ajo_living/supermarket/products/<PRODUCT_CODE>.jpg` key 登記至 OSS 與 `media_assets`，測試環境不得寫入生產 OSS。
2026-08-12: 測試環境改為預設沿用生產外部服務配置，只保留本地資料與身份密鑰隔離；只有產品明確指定的接口才另行隔離。
2026-08-25: 會員中心偏好設定接入 iSmart `/auth/password/` 與 `/auth/notification-settings/`，由 AJO 登入態代理修改 iSmart 密碼及大廈通告電郵偏好；上游兩個寫入端點受 `EXTERNAL_APP_ALLOWED_IPS` 保護。

[PROTOCOL]: When changing repository-level prompt memory, update this section, check child `AGENTS.md` files, and add a dated line to `docs/PROMPT_INDEX.md` when prompt surfaces changed.
