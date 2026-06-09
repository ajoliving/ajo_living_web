# AGENTS.md

## 說明
- 本文件用於補充 `web` 子專案規範。
- 未特別說明之事項，沿用上層 [`AGENTS.md`](/Users/yangliu/Documents/Code/ajoliving_web/AGENTS.md)。
- 本文件及 `web` 目錄下所有 `.md` 檔案中均不得使用 emoji。

## 技術棧
- 前端統一使用 `Vue 3 + TypeScript + Vite + Pinia + Tailwind CSS + Axios`。
- 如需元件庫，統一使用 `Ant Design Vue 4`。
- 路徑別名統一使用 `@/` 映射到 `src/`。
- 樣式以 `Tailwind CSS` 為主，局部元件樣式可使用 `<style scoped lang="scss">`。

## 產品與資訊架構定位
- Web 前端是 `AJO Living` 物業管理主系統的入口，不是單一 marketplace 或內容展示站。
- 新增頁面前必須先判斷使用者身份與入口層級：公開訪客、業主、租客、職員、管理員或管理公司後台。
- 大廈、單位、成員身份、權限與可見範圍是物業管理功能的核心上下文；涉及住戶能力的頁面不得只按普通公開網站列表邏輯設計。
- 舊系統功能遷移到 Web 時，需先映射到 AJO 的主模組與路由樹，再設計頁面；不得直接照搬舊系統頁面結構。
- UI 保持正式、清楚、簡約，避免行銷式說明、舞台式提示、實作過程文案與對使用者無價值的佔位描述。
- 新增大型功能時，優先做清晰入口、可驗證主流程與必要狀態，不提前堆疊複雜設定頁或用不到的操作。
- 舊業務後台仍繼續使用時，AJO Web 優先承擔統一查看、狀態展示、會員入口與跨系統跳轉，不默認重做舊後台的全部管理能力。
- 接入 POS、iBoard、iCCTV、iLock、Intercom 等舊系統時，頁面文案與導航必須以 AJO 的物業管理語義呈現，不暴露舊系統工程名或內部服務名。
- 外部舊系統連結只能作為明確的跳轉入口或嵌入入口；業務資料展示應優先經由 AJO 後端 API 整理後返回前端。

## 目錄結構

```text
src/
  main.ts
  App.vue
  app/               # 應用裝配層
    router/
      index.ts
      guards.ts
      modules/
        home.ts
        properties.ts
        furniture.ts
        account.ts
        management.ts
        communications.ts
        payments.ts
        security.ts
        system.ts
    stores/
    i18n/
    styles/
  domains/           # 業務域 API、型別與常量
    account/
    building/
    communications/
    payments/
    security/
    property/
    marketplace/
  pages/
    home/
      HomePage.vue
      composables/
      widgets/
    properties/
      Page.vue
      detail/
        Page.vue
      editor/
        Page.vue
    serviced-residences/
      Page.vue
      detail/
        Page.vue
      editor/
        Page.vue
    furniture/
      Page.vue
      listing/
        Page.vue
      seller/
        Page.vue
      chat/
        Page.vue
    member/
      Page.vue
      profile/
        info/
          Page.vue
        wallet/
          Page.vue
      property/
        Page.vue
      listings/
        Page.vue
        editor/
          Page.vue
        preview/
          Page.vue
      favorites/
        Page.vue
      widgets/
    management/
      Page.vue
      users/
        Page.vue
      property-sales/
        Page.vue
      serviced-apartments/
        Page.vue
      secondhand-listings/
        Page.vue
      reward-ads/
        Page.vue
      wallet-transactions/
        Page.vue
      content/
        Page.vue
      settings/
        display-ads/
          Page.vue
        home-carousel/
          Page.vue
        home-hero-cards/
          Page.vue
        login-hero/
          Page.vue
        notice/
          Page.vue
      integrations/
        Page.vue
    communications/
      Page.vue
      messages/
        Page.vue
      notices/
        Page.vue
    payments/
      Page.vue
      bills/
        Page.vue
      records/
        Page.vue
    security/
      Page.vue
      cctv/
        Page.vue
      access-control/
        Page.vue
      intercom/
        Page.vue
    building/
      Page.vue
      detail/
        Page.vue
    account/
      login/
        Page.vue
        composables/
        widgets/
    not-found/
      Page.vue
  shared/            # 跨模組復用能力
    components/
      base/
      layout/
      navigation/
      feedback/
      data-display/
    composables/
    utils/
    assets/
  mock/
```

## Page 單例 Skill

- 以 `pages` 作為主路由模組邊界，主模組需優先貼近實際公開入口與 URL，例如 `home`、`properties`、`serviced-residences`、`furniture`、`member`、`management`；平台能力入口再按 `communications`、`payments`、`security`、`building`、`account/login` 保持清楚邊界。
- `src/pages` 只承載頁面樹，不承載模組級 `routes.ts`、`api/`、`types/`；路由放在 `src/app/router/modules`，API、型別與常量放在 `src/domains`。
- 主路由可直接以單檔 page 形式存在，例如 `home/HomePage.vue`、`properties/Page.vue`、`serviced-residences/Page.vue`、`furniture/Page.vue`、`member/Page.vue`、`management/Page.vue`。
- 若主路由本身承載子路由或 tab 外框，可在主模組根目錄放 `Page.vue`，例如 `member/Page.vue`、`management/Page.vue`。
- `pages/` 下的每個子目錄代表一個實際路由 page 或主路由下的實際 tab page，不論是主路由或子頁，都必須有自己的單獨目錄。
- 每個 page 目錄必須有且僅有一個主入口檔，命名優先使用 `Page.vue`；只有像 `HomePage.vue` 這類已穩定的明確單頁主路由可保留語義化命名。
- page 私有元件放在 `widgets/`，page 私有彈窗放在 `modals/`，page 私有狀態與資料協調放在 `composables/`。
- page 專用常量放在 `constants/`，page 專用樣式變數或樣式設定放在 `styles.ts` 或 `styles.scss`。
- 若 page 很簡單，可只保留 `<Route>Page.vue`；若 page 較重，再按需增加 `widgets/`、`modals/`、`composables/`、`constants/`、`styles.*`。
- 主模組根目錄可保留主路由入口與該主路由私有 `widgets/`、`composables/`；不得再混入其他主入口或無路由使用的歷史檔案。
- 只有跨 page 復用的元件，才可提升到 `src/shared/components/`；僅在單一路由使用的元件不可上提。
- page 的資料流、查詢條件、彈窗開關、提交流程與事件協調，必須集中在 page 入口或 page composable 中，不可分散到多個平級 page 檔案。

### 主入口範例

```text
pages/
  properties/
    Page.vue
    detail/
      Page.vue
    editor/
      Page.vue
  serviced-residences/
    Page.vue
    detail/
      Page.vue
    editor/
      Page.vue
  member/
    Page.vue
    profile/
      info/
        Page.vue
      wallet/
        Page.vue
    property/
      Page.vue
    listings/
      Page.vue
      editor/
        Page.vue
      preview/
        Page.vue
    favorites/
      Page.vue
  management/
    Page.vue
    users/
      Page.vue
    property-sales/
      Page.vue
    serviced-apartments/
      Page.vue
    secondhand-listings/
      Page.vue
    reward-ads/
      Page.vue
    wallet-transactions/
      Page.vue
```

- `account/login/` 是登入主路由專屬 page 目錄，登入頁相關邏輯、局部元件、彈窗都收在這個目錄內。
- `properties/` 對應樓盤租售主入口；`serviced-residences/` 對應服務式住宅主入口。兩者可復用同一底層物業能力，但 page 入口必須分開，避免從目錄上看不出公開主路由。
- `member/` 是會員中心主入口，內部以 tab 或子頁承載個人資料、錢包、發布、收藏與站內信入口；不得放回 `account/my/` 或 `marketplace/my/`。
- `management/` 是管理主入口，管理 tab page 放在 `management/` 下；無實際入口的歷史檔案應刪除，不保留在 page 樹中。
- `account/` 僅保留登入、註冊、驗證等帳戶入口；會員中心使用 `member/`。
- `furniture/` 對應二手傢俬公開入口；不要再用 `marketplace/` 作為公開 page 目錄名。

### 舊系統接入建議目錄

```text
pages/
  communications/      # 站內信、通告、通知中心
  payments/            # 物業費、賬單、付款紀錄
  security/            # CCTV、門禁、對講、智能設備
  building/            # 大廈與單位上下文
  member/              # 會員中心
  account/             # 登入與帳戶驗證
  management/          # AJO 自身管理端入口
```

- 若只是展示舊系統資料，頁面放在 AJO 對應主域下，例如 `security/cctv/`，不要放在 `legacy/icctv/`。
- 若需要跳轉舊後台，只在管理端或明確操作入口提供連結，不把舊後台頁面拆進 AJO page 目錄。
- 每個舊系統接入頁只保留必要的列表、詳情、狀態與操作入口；複雜設定仍回到原後台處理。

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
- Axios 實例統一放在 `src/shared/utils/http/index.ts`。
- `baseURL` 透過 Vite proxy 或環境變數配置，不在程式碼中硬編碼域名。
- 請求攔截器統一注入 `Authorization` header。
- 響應攔截器統一處理 `401`、登入態失效與跳轉登入頁。
- 每個業務域的 API 檔案放在 `src/domains/<domain>/` 下。
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

## 樣式規範
- 樣式以 Tailwind CSS utility class 為主。
- 元件局部樣式用 `<style scoped lang="scss">`。
- `CSS class` 命名使用 kebab-case。
- `Ant Design Vue` 樣式優先使用 props、slots、token 或 `:deep()` 微調。
- 不使用 `!important`。
- 不寫大面積全域樣式覆蓋。

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
- 站內信、聊天與未讀數應按平台級通訊入口設計，不應長期綁死在 `marketplace` 路由與文案下。
- 若聊天入口來自二手交易、物業服務、維修、大廈或管理端，前端需保留來源模組與上下文展示，但會話列表與通知狀態應能被統一承載。
- 未來群聊 UI 不得直接複製一對一聊天頁；需要先明確群類型、參與者身份、可見性、管理權限、退出或封存規則。

## 實作偏好
- 優先完成列表頁、詳情頁、發布頁、會員中心等主流程。
- 響應式體驗需同時覆蓋 desktop 與 mobile。
- 本子專案是 Web 網頁，不是原生 App；涉及「手機」時，預設只表示手機號碼或 mobile viewport，不應按 App 下載、App 入口、App 內操作或原生 App 語義設計。
- 不預設引入重量級抽象與狀態管理。
