# docs

> Level/parent: ../AGENTS.md

## 成員清單
README.md: Docs 目錄索引與歸檔規則。
PROMPT_INDEX.md: AI prompt、agent 規則、工具描述與 prompt 合約測試索引。
architecture/: 系統架構、模組邊界與長期設計決策文件。
development/: Vue、React、Go 與 API 文件生成等開發規範。
deployment/: 部署、運維、SSL、伺服器操作與部署 prompt。
archive/: 已歸檔產品、原型與備份資料，不作 prompt/context 入口。
integrations/: 外部 App、舊系統與 AJO 模組映射文件。
data/: 地址表、參考資料與配置資料。
prompts/: 可直接給 AI 使用的頁面、表單與 UI 風格 prompt。
deployment/AGENTS.md: 部署 prompt 與伺服器操作文檔的本地維護協議。
prompts/AGENTS.md: UI prompt 文件的本地維護協議。

## 架構決策
- `docs/PROMPT_INDEX.md` 是本目錄的 prompt/context 主索引；新增或移動 prompt、agent 規則、AI 文件生成規範或工具描述時必須同步更新。
- `docs/prompts/` 只保存可直接複用的 UI 與頁面 prompt；活躍產品決策、驗收口徑與資料流優先放在 `docs/architecture/` 或 `docs/integrations/`，歷史產品資料只作 `docs/archive/product/` 參考。
- 部署類 prompt 保持在 `docs/deployment/`；執行部署前仍需按根級規則先讀部署文件與現有伺服器配置。
- 文檔內容使用繁體中文或 English；除非維護既有文件，不新增簡體中文正文。
- `docs/archive/`、`.claude/worktrees/`、`dist`、`node_modules`、資料庫備份與 HTML 備份不列入穩定 prompt surface，除非使用者明確要求追溯歷史。

## 變更日誌
2026-07-08: 建立 docs prompt/context 記憶邊界與索引維護規則。
2026-07-08: 補充 `docs/prompts` 與 `docs/deployment` 的本地 prompt 維護入口。
2026-07-08: 將 `product`、`prototypes`、`backups` 移入 `archive`，並從活躍 prompt/context 入口排除。
2026-07-08: 補充 iSmart integration API 在 POS relay 與 AJO 後端會員態接口的集成邊界。
2026-07-09: 補充 `我的大廈` 的大廈財務與業戶帳目會員態接口使用約定。
2026-07-09: 同步根級每次修改後執行 prompt/context 記憶檢查的默認規則。
2026-07-09: 在 prompt index 記錄 `web-new` 手機端瀏覽器審計命令與驗收覆蓋範圍。
2026-07-09: 補充手機端審計需覆蓋登入、聊天、付款、錢包、編輯與設定主要操作流。
2026-07-09: 補充手機端審計需覆蓋舊入口重定向、管理中心全部 tab、底部導航與管理 tab 點擊流。
2026-07-09: 補充手機端審計需檢查可見觸控目標尺寸，避免手機點擊入口過窄。
2026-07-09: 補充手機端審計需捕捉 console warning/error、失敗請求、API 4xx/5xx 與手機抽屜交互流。
2026-07-09: 補充手機端審計需覆蓋 iPhone 14 Pro Max 短視口與橫屏重點路由。
2026-07-09: 補充手機端審計需檢查頁面底部可操作內容不得被固定底部導航遮擋。
2026-07-09: 補充手機端審計需檢查未登入私有入口重定向與手機導航可見性。
2026-07-09: 補充手機端審計需檢查普通會員訪問 staff 管理與設定入口時回到會員中心且手機導航不顯示管理入口。
2026-07-10: 補充手機端真實 API 審計需以真實列表資料生成詳情路由，並將需寫入或驗證碼的互動流留在 mock 或顯式允許模式。
2026-07-10: 補充舊公開詳情入口在真實資料庫無示例 ID 時應回到對應列表頁，避免手機用戶進入 404 詳情頁。
2026-07-10: 補充手機端審計需檢查 document 層真實觸摸拖動滾動，避免全站手機頁面無法手指滑動。
2026-07-10: 補充手機端導覽斷點不顯示桌面大頁腳，避免固定底部導航與頁腳入口重複。
2026-07-10: 補充手機端審計需檢查頂部選單、置中 Logo、主題與語系切換及未登入登入入口的可見性與相對位置。
2026-07-10: 補充手機端審計需檢查左側抽屜保留可點擊遮罩，並支援遮罩關閉。
2026-07-10: 補充未登入手機端頂部登入入口必須可直接前往登入頁的交互審計。
2026-07-10: 補充樓盤放售與家具公開列表的手機端篩選需使用搜尋框左側入口與緊湊型底部彈窗。
2026-07-10: 補充樓盤與家具手機端篩選彈窗需貼齊視口底部，並覆蓋短視口與橫屏操作列可見性審計。
2026-07-10: 補充樓盤發布契約測試需覆蓋前端三段請求、具體缺失欄位提示、相片要求與土地類型後端校驗。
2026-07-12: 補充樓盤草稿可保存未完成資料，而完整欄位校驗只在正式發布時執行。

[PROTOCOL]: When changing prompt-bearing docs in this directory, update `PROMPT_INDEX.md`, add a dated change-log line, and check parent `../AGENTS.md` for project-level memory changes.
