# docs

> Level/parent: ../AGENTS.md

## 成員清單
README.md: Docs 目錄索引與歸檔規則。
PROMPT_INDEX.md: AI prompt、agent 規則、工具描述與 prompt 合約測試索引。
architecture/: 系統架構、模組邊界與長期設計決策文件。
development/: Vue、React、Go 與 API 文件生成等開發規範。
deployment/: 部署、運維、SSL、伺服器操作與部署 prompt。
product/: 產品需求、驗收、稽核與頁面規格。
integrations/: 外部 App、舊系統與 AJO 模組映射文件。
data/: 地址表、參考資料與配置資料。
prototypes/: HTML 原型與可預覽靜態資產。
backups/: 歷史備份檔案，不作為優先上下文入口。
prompts/: 可直接給 AI 使用的頁面、表單與 UI 風格 prompt。

## 架構決策
- `docs/PROMPT_INDEX.md` 是本目錄的 prompt/context 主索引；新增或移動 prompt、agent 規則、AI 文件生成規範或工具描述時必須同步更新。
- `docs/prompts/` 只保存可直接複用的 UI 與頁面 prompt；具體產品決策、驗收口徑與資料流仍放在 `docs/product/`、`docs/architecture/` 或 `docs/integrations/`。
- 部署類 prompt 保持在 `docs/deployment/`；執行部署前仍需按根級規則先讀部署文件與現有伺服器配置。
- 文檔內容使用繁體中文或 English；除非維護既有文件，不新增簡體中文正文。
- `.claude/worktrees/`、`dist`、`node_modules`、資料庫備份與 HTML 備份不列入穩定 prompt surface，除非使用者明確要求追溯歷史。

## 變更日誌
2026-07-08: 建立 docs prompt/context 記憶邊界與索引維護規則。

[PROTOCOL]: When changing prompt-bearing docs in this directory, update `PROMPT_INDEX.md`, add a dated change-log line, and check parent `../AGENTS.md` for project-level memory changes.
