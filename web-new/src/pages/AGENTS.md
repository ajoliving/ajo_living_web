# web-new/src/pages

> Level/parent: ../../AGENTS.md

## 成員清單
account/: 帳戶登入、忘記密碼、會員中心、個人資料與通知頁面。
building/: 大廈與住戶上下文頁面。
first/: 首次進入或初始導流頁面。
furniture/: 二手物品公開列表、詳情與相關流程。
home/: AJO Living 首頁與公開入口。
marketplace/: 二手交易、會員列表、聊天、管理端與設定頁面；不得承接所有平台功能。
notifications/: 全站通知中心頁面。
offers/: 超市優惠與門店資料頁面。
payments/: 物業費、賬單、付款、支付紀錄與 POS 支付入口頁面。
property/: 樓盤放售列表、詳情、發布與會員頁面。
serviced-residence/: 服務住宅列表與詳情頁面。
trend/: 市場趨勢頁面。
not-found/: 404 頁面。

## 架構決策
- 本目錄只承載 page 樹；API、型別、路由、全域狀態仍分別放在 `src/httpapis`、`src/model`、`src/router`、`src/pinia` 或 `src/stores`。
- 新增頁面前先判斷主模組：帳戶、大廈、物業、通知、支付、設備、二手交易或其他明確業務域。
- 每個主要 page 保持單一入口；局部元件、彈窗、狀態協調優先收在該 page 目錄內的 `widgets/`、`modals/`、`composables/`。
- UI prompt 只作為設計與實作參考；落地前仍要讀現有頁面結構、資料流與後端 API。
- 頁面文案使用繁體中文或 English，保持正式、簡約、可執行，不加入舞台感或實作說明。

## 變更日誌
2026-07-08: 建立前端 page 目錄記憶，固定主模組歸屬與 prompt 落地入口。

[PROTOCOL]: When adding, moving, renaming, or changing page ownership, update this map, check parent `../../AGENTS.md`, and update `../../../docs/PROMPT_INDEX.md` only if prompt-facing behavior changed.
