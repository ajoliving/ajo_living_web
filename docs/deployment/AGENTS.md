# docs/deployment

> Level/parent: ../AGENTS.md

## 成員清單
server-deployment-ai-prompt.md: AJO 與共用伺服器部署、端口、frpc、nginx、Supervisor、Docker 與驗收指引。
oss-cdn-letsencrypt-renewal-prompt.md: OSS CDN Let's Encrypt 憑證續期與 HTTPS 驗證指引。

## 架構決策
- 部署文件是高風險操作入口；執行前必須先讀相關部署 prompt，再 SSH 查看現有伺服器配置。
- 不新增與當前部署無關的服務、目錄、反代、端口或自動化；維持已有服務穩定優先。
- AJO 部署以 `docs/deployment/server-deployment-ai-prompt.md` 和根目錄部署腳本為準；不得依賴舊 `doc copy` 路徑作為最新來源。
- 部署類變更必須回報運行時真相、產品效果、風險、決策點與驗證狀態，不輸出大段命令日誌。

## 變更日誌
2026-07-08: 建立部署 prompt 目錄記憶，固定部署前讀取與驗證邊界。

[PROTOCOL]: When changing deployment prompts or server-operation instructions, update this map, `../PROMPT_INDEX.md`, and verify whether root `../AGENTS.md` deployment references need updating.
