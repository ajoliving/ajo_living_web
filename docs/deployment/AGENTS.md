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
- 部署類變更必須回報運行時真相、產品效果、風險、決策點與驗證狀態，不輸出大段命令日誌。

## 變更日誌
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

[PROTOCOL]: When changing deployment prompts or server-operation instructions, update this map, `../PROMPT_INDEX.md`, and verify whether root `../AGENTS.md` deployment references need updating.
