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
把面向所有者的token当成稀缺执行资源。所有者是产品审阅者，不是实现工程师。所有者需要运行时真相、产品效果、风险、决策点和验证状态;除非明确要求，所有者不需要代码讲解。
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
## 專案技術基線
- 前端統一使用 `Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS + Axios`。
- 如需元件庫，統一使用 `Ant Design Vue 4`。
- 後端統一使用 `Go + Gin + GORM + PostgreSQL`。
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
- 前端子專案遵循 [`web/AGENTS.md`](/Users/yangliu/Documents/Code/ajoliving_web/web/AGENTS.md)。
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
