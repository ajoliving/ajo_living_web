# AGENTS.md

## 說明
- 本文件用於補充 `web-new` 子專案規範。
- 未特別說明之事項，沿用上層 [`AGENTS.md`](/Users/yangliu/Documents/Code/ajoliving_web/AGENTS.md)。
- 本文件及 `web-new` 目錄下所有 `.md` 檔案中均不得使用 emoji。
- `web-new` 是目前正在開發與維護的前端子專案；後續前端功能修改、頁面調整、接口接入與 UI 修正預設都在本目錄完成。`web` 僅為舊版前端歸檔目錄，不承接新功能，除非使用者明確要求修改舊版歸檔。

## 技術棧
- 前端統一使用 `Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS + Axios`。
- 如需元件庫，統一使用 `Ant Design Vue 4`。
- 路徑別名統一使用 `@/` 映射到 `src/`。
- 樣式以 `Tailwind CSS` 為主，局部元件樣式可使用 `<style scoped lang="scss">`。

## 目錄結構

```text
src/
  httpapis/          # API 請求函式
  model/             # TypeScript 型別定義
  pinia/             # Pinia 全域狀態
  router/            # 路由定義與導航守衛
  stores/            # 額外共享狀態
  utils/             # 無業務狀態工具函式
  components/        # 跨頁面通用元件
  pages/
    404.vue
    home/
      HomePage.vue
      composables/
      widgets/
    building/
      BuildingPage.vue
    first/
      FirstPage.vue
    marketplace/
      Page.vue
      composables/
      widgets/
      <sub-route>/
        Page.vue
        composables/
        widgets/
        modals/
        constants/
        types/
        styles.ts
    account/
      login/
        Page.vue
        composables/
        widgets/
      my/
        Page.vue
        composables/
        widgets/
        <sub-route>/
          Page.vue
          composables/
          widgets/
          modals/
          constants/
          types/
          styles.ts
```

## Page 單例 Skill

- 以 `pages` 作為主路由模組邊界，主模組以業務路由劃分，例如 `account`、`home`、`building`、`first`、`marketplace`。
- `src/pages` 只承載頁面樹，不承載模組級 `routes.ts`、`api/`、`types/`；這些內容分別放在 `src/router`、`src/httpapis`、`src/model`。
- 主路由可直接以單檔 page 形式存在，例如 `home/HomePage.vue`、`building/BuildingPage.vue`、`first/FirstPage.vue`。
- 若主路由本身承載子路由外框，可在主模組根目錄放 `Page.vue`，例如 `marketplace/Page.vue`、`account/my/Page.vue`。
- `pages/` 下的每個子目錄代表一個實際路由 page，不論是主路由或子路由，都必須有自己的單獨目錄。
- 每個 page 目錄必須有且僅有一個主入口檔，命名優先使用 `Page.vue`；只有像 `HomePage.vue`、`BuildingPage.vue`、`FirstPage.vue` 這類明確單頁主路由可保留語義化命名。
- page 私有元件放在 `widgets/`，page 私有彈窗放在 `modals/`，page 私有狀態與資料協調放在 `composables/`。
- page 專用常量放在 `constants/`，page 專用樣式變數或樣式設定放在 `styles.ts` 或 `styles.scss`。
- 若 page 很簡單，可只保留 `<Route>Page.vue`；若 page 較重，再按需增加 `widgets/`、`modals/`、`composables/`、`constants/`、`styles.*`。
- 主模組根目錄可保留主路由入口與該主路由私有 `widgets/`、`composables/`；不得再混入其他子路由的私有元件或私有 composable。
- 只有跨 page 復用的元件，才可提升到模組級共享目錄或全域 `components/`；僅在單一路由使用的元件不可上提。
- page 的資料流、查詢條件、彈窗開關、提交流程與事件協調，必須集中在 page 入口或 page composable 中，不可分散到多個平級 page 檔案。

### Account 範例

```text
pages/
  account/
    login/
      Page.vue
      composables/
      widgets/
      modals/
    my/
      Page.vue
      widgets/
      composables/
      overview/
        Page.vue
      profile/
        Page.vue
      preferences/
        Page.vue
        widgets/
        modals/
        composables/
        styles.ts
```

- `account/login/` 是登入主路由專屬 page 目錄，登入頁相關邏輯、局部元件、彈窗都收在這個目錄內。
- `account/my/` 是會員中心主路由外框，負責側欄、內容區與子路由承載。
- `account/my/preferences/` 是設定頁專屬 page 目錄，可像參考專案 `pages/auth/preferences/` 一樣拆成 `widgets/`、`modals/`、`styles.ts`。
- 若未來 `account/profile/`、`account/notifications/`、`marketplace/publish/` 等頁面變重，沿用相同單例 page 結構，不得回退到模組根目錄平鋪。

## 命名規範

| 目標 | 規則 | 範例 |
| --- | --- | --- |
| 元件檔案 | PascalCase | `DeviceEditDialog.vue` |
| composable 檔案 | `use` + PascalCase | `useDevice.ts` |
| API 檔案 | 小寫 kebab-case | `device.ts` |
| 元件註冊名 | PascalCase | `DeviceEditDialog` |
| prop / emit | camelCase | `deviceId` |
| template 引用 | PascalCase | `<DeviceEditDialog />` |
| 路由 `name` | PascalCase | `Device` |
| CSS class | kebab-case | `device-list` |
| 變數 / 函式 | camelCase | `isLoading` |
| 常數 | SCREAMING_SNAKE | `MAX_RETRY_COUNT` |
| 型別 / 介面 | PascalCase | `Device` |
| model 欄位 | 與後端 JSON key 一致 | `created_at` |

## 元件規範

### 1. 單文件元件順序

```vue
<!--
 * 元件用途簡要說明
 * 1. 核心功能點
 * 2. 核心功能點
-->
<script setup lang="ts">
// imports
// props & emits
// composables
// reactive state
// computed
// methods
// lifecycle
</script>

<template>
  <!-- 結構 -->
</template>

<style scoped lang="scss">
</style>
```

### 2. 核心規則
- 統一使用 `<script setup lang="ts">`，不用 Options API。
- 元件模板中只寫簡單表達式，複雜邏輯移到 `script`。
- Props 必須用型別宣告，不用運行時宣告。
- Emit 必須使用型別宣告。
- 重複使用兩次以上的區塊再抽成元件。
- 單個元件超過 200 行且職責可拆時，再考慮拆分。
- 彈窗、表單、列表項優先作為頁面級子元件。

## TypeScript 規範
- 型別優先級：`interface > type > class`。
- 純資料結構用 `interface`。
- 需要交叉型別或聯合型別時用 `type`。
- 需要建構函式、預設值、靜態方法時才用 `class`。
- 禁止使用 `any`，特殊情況需改用 `unknown` 或具體型別並加註解。
- 禁止使用 `// @ts-ignore`。
- 禁止隱式 `any`，保持 `tsconfig.strict: true`。
- `ref`、事件參數、函式回傳值型別需明確。

## Composable 規範
- 檔案放在對應 page 目錄的 `composables/` 下，不與 page 分離。
- 命名使用 `use<Module>` 或 `use<Module>Data`。
- loading 狀態必須在 `finally` 中重置。
- 錯誤需要顯式提示，不可靜默吞掉。
- composable 不直接操作 DOM。

## API 請求規範
- Axios 實例統一放在 `src/httpapis/index.ts`。
- `baseURL` 透過 Vite proxy 或環境變數配置，不在程式碼中硬編碼域名。
- 請求攔截器統一注入 `Authorization` header。
- 響應攔截器統一處理 `401`、登入態失效與跳轉登入頁。
- 每個資源一個 API 檔案。
- API 函式不做 UI 錯誤提示，錯誤由呼叫方處理。
- API 路徑與後端保持一致，不額外封裝路徑語義。

## 路由規範
- 需登入頁面加上 `meta: { requiresAuth: true }`。
- 路由守衛集中在 `router` 目錄中管理。
- 不在元件內做路由鑑權判斷。
- 路由 `path` 使用 kebab-case，`name` 使用 PascalCase。

## Pinia 規範
- 只存全域共享狀態，例如登入態、使用者資訊、全域設定。
- 模組級狀態優先放在 composable 內。
- getter 只做派生計算，不加副作用。
- action 才允許非同步操作。
- 登入狀態還原需讓路由守衛等待同一個進行中的 hydrate promise，不可在 `isHydrating` 時提前返回造成 staff 權限誤判。

## 樣式規範
- 樣式以 Tailwind CSS utility class 為主。
- 元件局部樣式用 `<style scoped lang="scss">`。
- `CSS class` 命名使用 kebab-case。
- `Ant Design Vue` 樣式優先使用 props、slots、token 或 `:deep()` 微調。
- 不使用 `!important`。
- 不寫大面積全域樣式覆蓋。

## 行動端基線
- 新增或調整 page 時，必須維持 `iPhone 14 Pro Max` 尺寸下可用：不出現橫向頁面溢出，底部導航不得遮擋表單、按鈕、聊天輸入、付款操作或彈窗操作。
- 手機端優先沿用 `src/styles/tokens.css` 與 `src/styles/index.css` 的安全區、底部導航、單欄、表格橫向滾動與觸控尺寸基線，不在單一 page 內另寫互相衝突的全域規則。
- 管理、會員、支付、聊天、編輯器與設定頁在手機端應採用單欄內容、橫向分段導航、全寬主要操作與可滾動表格；桌面側欄不得在手機端保持 sticky 雙欄。
- 會員中心、通知中心、管理中心與設定頁在手機導航斷點內使用無外框的緊湊標題帶與橫向可滑動導航，不保留大型側欄卡片；會員資料操作採同列網格，錢包摘要採雙列，列表狀態篩選採橫向滑動控件，避免首屏被重複導航和縱向按鈕佔滿。
- 表單、彈窗與抽屜在手機端需保留 `safe-area` 底部空間，確認提交、取消、關閉與返回操作可直接觸達。
- 樓盤放售與家具公開列表在手機端不得常駐整段篩選條件；搜尋框獨立置頂，常用條件與排序拆為可橫向滑動的獨立篩選控件，結果統計置於其下。樓盤快捷欄依次保留地區、物業類型、房間、裝修與排序，家具快捷欄依次保留分類、地區、成色與排序。搜尋與快捷篩選控制帶沿用優惠頁的單層 `12px` 邊距、白色控制區與淺灰結果區分層；其餘條件以「更多」進入帶遮罩、右上角關閉與底部結果操作的緊湊型底部彈窗。彈窗必須貼齊可視區底部且在短視口與橫屏下保留可滾動內容和可見操作列，桌面端保留左側篩選欄與排序欄。
- 手機端可聚焦或可點擊控件需保留底部導航避讓距離，避免 `scrollIntoView`、輸入聚焦或表單操作把按鈕和輸入框定位到固定底欄下方。
- 使用 `Teleport to="body"` 的彈窗必須使用可在 body 層命中的手機端選擇器，不可只依賴 `#app` 內部選擇器。
- 手機端 document 層必須保留真實觸摸拖動滾動能力，不得在全域 body/root 使用 `overscroll-behavior-y: none` 或同類禁用滾動鏈的規則。
- 手機端導覽斷點（`1023px` 以下）不顯示桌面大頁腳；固定底部導航與手機抽屜承擔主要入口，客服、私隱與條款等正式二級入口應在具備真實頁面後放入會員設定或手機抽屜。
- 手機端頂部導航採三段式：左側漢堡選單、中間固定置中的 `AJO LIVING` Logo、右側主題與語系切換；未登入時於右側補充直接進入登入頁的「登入」入口，登入後隱藏該入口並保留底部「我的」會員入口；語系切換直接使用既有繁中與 English 偏好，不另設彈窗或第二層入口。
- 手機端漢堡選單使用左側抽屜，不覆蓋整個畫面；抽屜約佔 `76vw`、其餘區域使用半透明遮罩，點擊遮罩或關閉按鈕均須收起抽屜。
- 手機端高度使用 `100svh` 或既有安全區 token，不使用 `100vh`、`min-h-screen`、`h-screen`、`max-h-screen` 這類會在 iOS 造成視口高度偏差的寫法。
- 完整手機端驗收使用 `npm run audit:mobile`；該命令需要既有前端服務可訪問，預設 mock API，不會自行啟動前端或後端，並覆蓋主要路由、舊入口重定向、管理中心 tab、未登入私有入口重定向與手機導航可見性、普通會員訪問 staff 入口攔截與手機導航可見性、短視口與橫屏重點路由、觸控目標、真實觸摸拖動滾動、可聚焦控件、頁面底部可操作內容、console warning/error、失敗請求、API 4xx/5xx 與登入、聊天、付款、錢包、編輯、設定、底部導航、手機抽屜等主要操作流。
- 如需縮小定位範圍，可使用 `node scripts/audit-mobile.mjs --section=guest|staff|member-staff|viewport` 分段驗收；分段結果必須與完整手機端基線保持一致。
- 真實 API 手機端驗收可使用 `node scripts/audit-mobile.mjs --mock-api=false --section=guest`；該模式會以真實列表第一條資料生成詳情路由，預設跳過需驗證碼、付款、聊天、發布或設定寫入的互動流。staff 或普通會員真實驗收需透過環境變數提供對應 token。

## 模板規範
- 屬性值始終帶引號。
- 多屬性元件寫法需換行。
- `v-for` 必須加唯一 `:key`，不可使用 `index`。
- 不在同一元素上同時使用 `v-if` 和 `v-for`。
- 不在 template 中寫複雜 JavaScript 表達式。
- 簡單條件用 `v-if` / `v-else`，需保留節點時再用 `v-show`。

## 檔案註解規範
- `.ts` 檔案需使用頂部多行註解說明職責。
- `.vue` 檔案需使用頂部註解區塊說明元件用途。
- 主要函式與關鍵邏輯區塊需使用序號註解。

```typescript
/*
 * 模組名 - 檔案職責
 * 1. 功能點一
 * 2. 功能點二
 * 3. 功能點三
 */

// 1. 取得列表資料
const fetchList = async () => { ... };
```

```vue
<!--
 * 元件名 - 元件用途
 * 1. 功能點一
 * 2. 功能點二
-->
```

## 與後端協作邊界
- 所有業務資料必須經由後端 HTTP API 取得。
- 所有上傳流程只能調用後端簽發的上傳介面，不可在前端保存阿里雲 `OSS` 密鑰。
- 權限判斷以前端引導為輔、後端校驗為準，不可只依賴前端隱藏按鈕。
- 聯絡方式、聊天、會員身份與 `building_only` 可見性，全部以前後端介面契約為準。

## 實作偏好
- 優先完成列表頁、詳情頁、發布頁、會員中心等主流程。
- 響應式體驗需同時覆蓋 desktop 與 mobile。
- 不預設引入重量級抽象與狀態管理。
- 樓盤未完成資料可保存為草稿；正式發布在前端校驗失敗時必須顯示具體缺失欄位並切換到對應步驟。回歸測試需覆蓋未完成草稿、建立草稿、更新圖片 payload 與正式發布三段請求。

## 子目錄記憶
- `src/pages/AGENTS.md` 是前端 page 樹的本地記憶入口；新增、移動、重命名 page 或改變頁面主域歸屬時同步更新。
- UI prompt 落地前先讀 `src/pages/AGENTS.md`、相關 page 目錄與 `docs/prompts/`，不要只按 prompt 重新發明頁面結構。

## 變更日誌
2026-07-08: 補充前端 page 目錄記憶入口，連接 UI prompt 與實際頁面邊界。
2026-07-09: 補充 body 層 Teleport 彈窗的手機端基線要求。
2026-07-09: 補充登入狀態 hydrate 需支援路由守衛並發等待，避免 staff 頁面誤退回會員中心。
2026-07-09: 補充手機端控件需保留底部導航避讓距離，避免表單按鈕與輸入框被固定底欄遮擋。
2026-07-09: 補充手機端不得使用 `100vh` 與 Tailwind screen 高度工具，統一改用 `100svh` 或安全區 token。
2026-07-09: 補充 `npm run audit:mobile` 作為 iPhone 14 Pro Max 主要路由與控件可達性驗收入口。
2026-07-09: 補充 `npm run audit:mobile` 需覆蓋登入、聊天、付款、錢包、編輯與設定主要操作流。
2026-07-09: 擴充 `npm run audit:mobile` 覆蓋舊入口重定向、管理中心全部 tab、底部導航與管理 tab 點擊流。
2026-07-09: 擴充 `npm run audit:mobile` 檢查手機端觸控目標尺寸，並將頁腳、breadcrumb、分頁、收藏與 Toast 操作納入基線。
2026-07-09: 擴充 `npm run audit:mobile` 捕捉 console warning/error、失敗請求、API 4xx/5xx 與手機抽屜交互流。
2026-07-09: 擴充 `npm run audit:mobile` 覆蓋 iPhone 14 Pro Max 短視口與橫屏重點路由，並加入 1023px 以下觸控尺寸基線。
2026-07-09: 擴充 `npm run audit:mobile` 檢查頁面滾到底部後可操作內容不得被固定底部導航遮擋。
2026-07-09: 擴充 `npm run audit:mobile` 檢查未登入訪問會員、大廈、支付、通知與管理入口時必須回到登入頁且手機導航不洩漏私有入口。
2026-07-09: 擴充 `npm run audit:mobile` 檢查普通會員訪問 staff 管理與設定入口時必須回到會員中心且手機導航不顯示管理入口。
2026-07-10: 擴充手機端真實 API 驗收，使用真實列表資料生成詳情路由，並將需寫入或驗證碼的互動流留在 mock 或顯式允許模式。
2026-07-10: 調整舊公開詳情入口回到對應列表頁，避免真實資料庫不存在示例 ID 時顯示 404。
2026-07-10: 補充手機端必須保留 document 層真實觸摸拖動滾動能力，並將觸摸滑動探針納入手機端審計。
2026-07-10: 補充手機端導覽斷點隱藏桌面大頁腳，避免與固定底部導航重複並延長無效滾動。
2026-07-10: 固定手機端頂部導航為左側選單、中間 Logo、右側主題與語系切換，並納入瀏覽器版面與交互審計。
2026-07-10: 調整手機端選單為左側抽屜與可點擊遮罩，避免展開時覆蓋整個畫面。
2026-07-10: 補充未登入手機端於頂部導航提供直接登入入口，並納入單元與瀏覽器交互審計。
2026-07-10: 調整樓盤放售與家具公開列表的手機端篩選為搜尋框左側入口與緊湊型底部彈窗，桌面端保留左側篩選欄。
2026-07-10: 補充樓盤發布校驗需指出缺失欄位並跳回對應步驟，新增建立草稿、更新圖片與正式發布的前端回歸測試。
2026-07-12: 對齊未完成樓盤草稿可保存、完整資料才可正式發布的前端契約。
2026-07-12: 對齊樓盤與家具公開列表手機端的搜尋與快捷篩選控制帶至優惠頁單層邊距與白色控制區布局，排序保留在同一篩選行。
2026-07-12: 固定樓盤手機快捷欄的地區、物業類型、房間、裝修與排序，以及家具快捷欄的分類、地區、成色與排序。
2026-07-12: 統一會員、通知、管理與設定頁的手機端緊湊標題帶、橫向導航、會員操作網格、錢包雙列摘要與列表狀態篩選。

[PROTOCOL]: When frontend page ownership or prompt-facing behavior changes, check `src/pages/AGENTS.md`, parent `../AGENTS.md`, and `../docs/PROMPT_INDEX.md`.
