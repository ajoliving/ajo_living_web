# docs/deployment

> Level/parent: ../AGENTS.md

## 成員清單
server-deployment-ai-prompt.md: AJO 與共用伺服器部署、端口、frpc、nginx、Supervisor、Docker 與驗收指引。
oss-cdn-letsencrypt-renewal-prompt.md: OSS CDN Let's Encrypt 憑證續期與 HTTPS 驗證指引。

## 架構決策
- 部署文件是高風險操作入口；執行前必須先讀相關部署 prompt，再 SSH 查看現有伺服器配置。
- 不新增與當前部署無關的服務、目錄、反代、端口或自動化；維持已有服務穩定優先。
- AJO 部署以 `docs/deployment/server-deployment-ai-prompt.md` 和根目錄部署腳本為準；不得依賴舊 `doc copy` 路徑作為最新來源。
- AJO Redis 與 PostgreSQL 由同一份伺服器 Docker Compose 管理，只綁定應用伺服器 `127.0.0.1`；Redis 只保存可重建快取，不配置持久化 volume。
- POS 大廈與單位共用目錄使用 Redis 快取 5 分鐘；部署環境以 `POS_DIRECTORY_CACHE_TTL` 控制，會員權限、綁定、目前物業及 relay token 不得進入共用快取。
- iSmart 意見提供及維修個案的提交、列表與詳情使用獨立 `ISMART_SERVICE_CASE_API_BASE_URL`，目前指向 `clouddev`。iSmart 副戶列表、授權與撤銷使用獨立 `ISMART_SUBACCOUNT_API_BASE_URL`，目前暫時指向生產 iSmart。
- AJO 測試環境使用 `test.ajoliving.skylinedances.com` 單一同源入口、獨立目錄、Supervisor、Docker Compose project、PostgreSQL volume、Redis 與密鑰；測試庫只可透過顯式確認由生產邏輯快照重建。
- 測試環境目前公開直接訪問並保持 `noindex`；`noindex` 不等同存取控制。外部服務預設沿用生產配置，只有產品明確指定的接口才隔離；測試資料庫、Redis、JWT 與加密密鑰維持獨立。
- 生產快照恢復至測試庫後，保留的應用密文必須使用測試獨立 `ENCRYPTION_KEY` 重加密；工具不得輸出明文或密鑰，且不得修改生產資料庫。
- 測試部署每次從生產 `.env` 重新繼承外部服務配置，並只覆蓋測試站本地運行配置及回調域名，避免殘留歷史 mock 或不可用地址。
- OSS CDN 媒體域名使用網關 `acme.sh` 的 `dns_ali` 自動簽發及 `ali_cdn` 自動部署；域名續期不得退回無法自動執行的 certbot manual DNS 流程。
- 部署類變更必須回報運行時真相、產品效果、風險、決策點與驗證狀態，不輸出大段命令日誌。

## 變更日誌
2026-09-12: 暫時停用 iSmart 個人檔案、大廈資料與通告 Redis 快取，改每次回源；POS 目錄快取維持不變。
2026-09-10: 副戶列表、授權與撤銷改為暫時走生產 `ISMART_SUBACCOUNT_API_BASE_URL`，方便本機驗證住戶授權。
2026-09-10: 記錄 iSmart 副戶列表、授權與撤銷使用獨立 `ISMART_SUBACCOUNT_API_BASE_URL` 指向 `clouddev`，授權前分別核對手機號與電郵的 `user_id`，且不得回退生產接口。
2026-07-08: 建立部署 prompt 目錄記憶，固定部署前讀取與驗證邊界。
2026-07-08: 補充 POS iSmart relay v2 的實際部署鏈路、Supervisor 名稱、端口與驗證規則。
2026-07-08: 補充 AJO 後端 iSmart integration 與 legacy external app API 並存的生產環境變數。
2026-07-16: 補充 AJO Redis 容器、`6382` 本機端口、5 分鐘 iSmart 大廈快取配置與部署驗收命令。
2026-07-16: 記錄生產 Redis 已通過真實會員大廈請求的 miss、寫入、hit、TTL 倒數及會員資料隔離驗證。
2026-07-17: 補充 POS 大廈與單位共用目錄的 5 分鐘 Redis 快取配置及會員資料排除邊界。
2026-07-17: 記錄 release `20260717094836-3321` 已通過 POS 大廈與單位目錄的快取建立、命中、TTL 及回應驗證。
2026-07-17: 明確 `icctv_orangepi_auth_service` 為已部署 OrangePi 設備端保護項目，未獲明確授權不得修改或重新部署。
2026-07-17: 記錄 release `20260717115200-3302` 以 `BACKUP_DB=0` 發布統一帳戶登入，並通過桌面、手機、路由、服務及 HTTPS 驗收。
2026-07-21: 記錄 Client Ticket Board 的 Docker Compose、SQLite volume、AJO Living OSS `tickets/` 前綴、frpc `:20046`、ACME DNS HTTPS 與續期重載設定。
2026-08-03: 記錄 release `20260803075905-18657` 已發布樓盤卡片展示、草稿物業資料唯讀及現有樓盤免費儲存修改；發布前建立資料庫備份，未修改 FRP、SSL、OSS CORS 或生產環境變數，並通過前後端 HTTPS 驗收。
2026-08-03: 記錄 iSmart 服務個案獨立 `clouddev` 測試地址的部署注入規則，其他 iSmart 接口維持生產配置。
2026-08-03: 記錄 release `20260803085144-47125` 已發布 iSmart 服務個案測試環境隔離；專用地址為 `clouddev`，其他 iSmart 地址維持生產環境，並通過前後端 HTTPS、Supervisor、PostgreSQL 及 Redis 驗收。
2026-08-04: 記錄 release `20260804085952-deploy` 已發布樓盤聯絡人資料一致性及當前工作區版本；發布前建立 PostgreSQL 備份，未同步生產 `.env`、FRP 或 OSS CORS，並將 AJO HTTPS 證書續期至 2026-11-02。
2026-08-05: 記錄 release `20260804171519-contact-final` 已分離發布者身份與逐樓盤聯絡資料；發布前建立 PostgreSQL 備份，未同步生產 `.env`、FRP、SSL 或 OSS CORS，並通過真實公開樓盤頁及服務健康驗收。
2026-08-05: 建立 AJO 獨立測試環境、單域名同源 API、受控生產快照、外部副作用隔離、AliDNS ACME 證書及測試發布驗收規則。
2026-08-05: 以 30 天安全 Cookie access gate 取代不穩定的出口 IP 白名單，保持測試站私有訪問且不佔用 AJO Bearer `Authorization` header。
2026-08-07: 按產品決策移除測試站 Cookie access gate，普通測試網址改為公開直接訪問，並保留 `noindex`、獨立運行資源及外部副作用隔離。
2026-08-07: 測試 release `20260807081910-test-92331` 與生產 release `20260807082205-94112` 已同步發布目前工作區；生產先建立並校驗 PostgreSQL 備份，保留環境與網絡配置，兩套前後端及資料服務均通過驗收。
2026-08-07: 測試 release `20260807083354-test-1100` 與生產 release `20260807083705-3047` 已同步發布樓盤詳情容錯修正；相似樓盤失敗不再遮蔽主詳情，生產備份、兩套 HTTPS、服務及瀏覽器驗收均通過。
2026-08-07: 修復測試快照樓盤聯絡密文與獨立測試密鑰不相容；先備份測試庫，再將目標樓盤密文受控重加密，原 Bearer Token 詳情請求已由 `500` 恢復為 `200`，並將流程固化至測試部署腳本。
2026-08-07: 完成 Good Price 商品審核同步：生產庫下架 28 個商品、修正 6 個名稱或品牌，部署每日 CSV 覆蓋規則；AJO 生產 OSS 導入 335 張商品圖片並登記 `media_assets`，測試環境未寫入生產 OSS。
2026-08-10: 發布 Good Price 商品名稱與副標題資料回填，以及 AJO 測試 release `20260810042405-test-48684` 和生產 release `20260810043333-53744`；兩套同源 API 與真實卡片均驗證為「品牌 + 商品名稱」及獨立副標題，生產 PostgreSQL 備份、Supervisor、Redis 與 HTTPS 均正常，未改寫 FRP、SSL 或 OSS CORS。
2026-08-12: 測試環境改為預設沿用生產外部服務配置，只保留資料、快取與身份密鑰隔離；只有產品明確指定的接口才另行隔離。
2026-08-12: 將 `ajoliving.oss.skylinedances.com` 由已過期的 certbot manual DNS 憑證遷移至 `acme.sh + AliDNS + ali_cdn` 自動續期及 CDN 部署；新憑證有效至 2026-11-10，真實圖片及瀏覽器驗收正常。
2026-08-17: 發布 Good Price 尾端規格自動拆分；每日 CSV 保留人工副標題並自動分離重量、容量、尺寸、數量及包裝規格，歷史回填限定最新兩個資料日。發布前 PostgreSQL 與舊二進制均已備份，Supervisor、公開 Good Price API 及 AJO 測試代理驗證正常。
2026-08-18: 測試 release `20260818111641-test-79673` 發布住戶 iSmart 副戶授權更新；保留測試資料庫，測試前後端、同源 HTTPS、Supervisor、PostgreSQL、Redis、FRP 與 `noindex` 驗證正常，生產健康接口維持 `200`。

[PROTOCOL]: When changing deployment prompts or server-operation instructions, update this map, `../PROMPT_INDEX.md`, and verify whether root `../AGENTS.md` deployment references need updating.
