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
- `docs/integrations/` 內與 iSmart 相關的活躍接口文檔，預設以 `/Users/yangliu/Documents/Code/hk/ajo_ismart` 倉庫已暴露路由為真值來源；整理 AJO 可接入能力時，需先核對上游實際路由與 view，再更新本目錄文檔。
- 部署類 prompt 保持在 `docs/deployment/`；執行部署前仍需按根級規則先讀部署文件與現有伺服器配置。
- 文檔內容使用繁體中文或 English；除非維護既有文件，不新增簡體中文正文。
- `docs/archive/`、`.claude/worktrees/`、`dist`、`node_modules`、資料庫備份與 HTML 備份不列入穩定 prompt surface，除非使用者明確要求追溯歷史。

## 變更日誌
2026-09-12: 暫時停用 iSmart 個人檔案、大廈資料與通告 Redis 快取，改每次回源。
2026-09-10: 副戶接口暫時改走生產 `ISMART_SUBACCOUNT_API_BASE_URL`，服務個案維持 `clouddev`。
2026-09-10: 記錄 iSmart 副戶授權走 `ISMART_SUBACCOUNT_API_BASE_URL` 的 `clouddev` 接口；授權前分別核對手機號與電郵 `user_id`，不得回退生產。
2026-09-12: 記錄 `ajo_ismart` 對 AJO 開發為只讀參考倉庫；生產副戶授權因上游 `IFlatTbl.building` 不存在而失敗，缺陷只寫入 AJO 提示文檔，不改上游程式碼。
2026-09-10: 為 `docs/integrations/external_app_building_api(1).md` 新增對應繁體中文接口文檔 `external_app_building_api.zh-HK.md`。該對譯以該英文底稿為準；AJO 實際接入仍以 `ajo_ismart` 已暴露路由為真值，現行上游是 `/auth/password/`、`/auth/notification-settings/` 與 `/payments/fees/`，不得把底稿中的 `/auth/change-password/`、`/auth/settings/`、`/payments/types/` 當成已上線接口。
2026-08-25: 將 `docs/archive/integrations/` 的 iSmart external app API 重複版本收斂為 `external_app_building_api_v3.md` 及其繁體中文版本；中文與英文底稿同步以 `ajo_ismart` 已暴露的 `/auth/password/`、`/auth/notification-settings/` 路由為真值。
2026-08-20: 記錄 `docs/integrations` 的 iSmart 活躍接口文檔以 `ajo_ismart` 倉庫已暴露路由為真值來源，未暴露 API 的頁面功能不得寫成 AJO 可直接接入能力。
2026-08-19: 記錄 Good Price 同系列商品人工目錄、公開詳情與 AJO 商品編號切換契約，避免以文字規則錯誤建立公開規格關係。
2026-08-19: 記錄會員中心帳戶資料、獨立提示設定與 AJO 業戶資料關聯契約；iSmart 原始帳戶資料維持唯讀，解除綁定只影響 AJO 本地關聯與目前使用物業。
2026-08-14: 在 `PROMPT_INDEX.md` 記錄本機 Feishu 直連 Codex CLI 的工作區、憑據與不使用 CLI resume 的運行邊界。
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
2026-07-16: 記錄個人註冊後 iSmart OwnerReg 待審綁定與未綁定「我的大廈」導向會員中心申請頁的資料流。
2026-07-16: 記錄會員中心住戶單位以 POS 單位詳情為顯示真值、權限碼只作授權與失敗回退的資料流。
2026-07-16: 記錄會員中心不分 Staff 身份只顯示住戶所屬屋苑及本人單位，員工管理屋苑不進入個人綁定 UI。
2026-07-16: 記錄 AJO Staff 只保護管理中心、iSmart Staff 不提升本地管理權限及一般功能按資源可見範圍授權的契約。
2026-07-16: 記錄「我的大廈」刷新最新會員綁定、已保存大廈優先及正式名稱取代 ID 佔位值的資料流。
2026-07-16: 記錄 AJO Redis 與 PostgreSQL 同機 Docker 部署，以及 iSmart 大廈資料 5 分鐘快取的安全邊界。
2026-07-16: 記錄生產 Redis 已通過真實 iSmart 大廈資料請求的快取建立、命中、TTL 與會員資料隔離驗證。
2026-07-17: 記錄統一帳戶登入的無資料庫備份生產 release、桌面與手機 UI、API 路由及服務健康驗收。
2026-07-28: 記錄樓盤草稿預付與發布差額抵扣的 API 契約、會員確認資料流與回歸測試索引。
2026-07-28: 記錄歷史樓盤草稿重複扣費的退款更正命令、淨預付計算及錢包流水本地化契約。
2026-07-30: 記錄代理註冊、牌照上載與 Staff 審核啟用的 API 契約索引更新。
2026-08-04: 記錄樓盤逐樓盤聯絡快照、代理展示資料邊界及公開詳情解鎖契約。
2026-08-05: 記錄樓盤草稿可修改物業資料、正式發佈後由前後端共同鎖定的 API 契約。
2026-08-05: 記錄 AJO 生產與測試雙環境的獨立部署、資料快照、安全隔離及驗收入口。
2026-08-05: 記錄測試站以受限入口密鑰換取安全 Cookie 的穩定訪問規則，取代易受網絡切換影響的 IP 白名單。
2026-08-07: 記錄測試站按產品決策取消 Cookie access gate，改為公開直接訪問並維持資料與外部副作用隔離。

[PROTOCOL]: When changing prompt-bearing docs in this directory, update `PROMPT_INDEX.md`, add a dated change-log line, and check parent `../AGENTS.md` for project-level memory changes.
