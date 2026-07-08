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

## 實作偏好
- 優先完成列表頁、詳情頁、發布頁、會員中心等主流程。
- 響應式體驗需同時覆蓋 desktop 與 mobile。
- 不預設引入重量級抽象與狀態管理。

## 子目錄記憶
- `src/pages/AGENTS.md` 是前端 page 樹的本地記憶入口；新增、移動、重命名 page 或改變頁面主域歸屬時同步更新。
- UI prompt 落地前先讀 `src/pages/AGENTS.md`、相關 page 目錄與 `docs/prompts/`，不要只按 prompt 重新發明頁面結構。

## 變更日誌
2026-07-08: 補充前端 page 目錄記憶入口，連接 UI prompt 與實際頁面邊界。

[PROTOCOL]: When frontend page ownership or prompt-facing behavior changes, check `src/pages/AGENTS.md`, parent `../AGENTS.md`, and `../docs/PROMPT_INDEX.md`.
