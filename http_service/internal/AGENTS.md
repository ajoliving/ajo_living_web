# http_service/internal

> Level/parent: ../AGENTS.md

## 成員清單
config/: 環境配置讀取與配置物件管理。
database/: 資料庫連線、初始化、遷移輔助與 seed。
errcode/: 統一回應格式、錯誤碼與錯誤訊息。
handler/: HTTP 參數解析、基礎校驗、service 呼叫與回應輸出。
logger/: 專案統一日誌入口。
middleware/: 認證、限流、CORS、恢復、請求 ID 與訪問日誌。
model/: GORM 模型、分頁結構與基礎資料結構。
router/: 路由註冊、中間件掛載與資源路由分組。
service/: 業務邏輯、資料庫讀寫、交易處理與第三方服務編排。
utils/: 無業務狀態的通用工具。

## 架構決策
- 依賴方向保持 `router -> handler -> service -> model/database`；不得讓 handler 直接操作資料庫或讓 service 輸出 HTTP 回應。
- 新增能力前先判斷主域，避免把支付、通知、設備、物業、交易、帳戶等能力混到單一 service。
- 舊系統接入優先做受控讀取、身份映射與權限校驗；未確認寫入責任前按唯讀處理。
- 資料模型、API 回應與錯誤語義變更時，同步檢查 `http_service/doc/api.md`、相關測試與 `docs/PROMPT_INDEX.md` 是否受影響。
- 部署與環境操作不在 `internal/` 內新增 prompt，統一回到 `docs/deployment/`。

## 變更日誌
2026-07-08: 建立後端 internal 目錄記憶，固定分層、主域與 prompt 影響檢查入口。

[PROTOCOL]: When adding, moving, renaming, or changing backend internal ownership, update this map, check parent `../AGENTS.md`, and update `../../docs/PROMPT_INDEX.md` when API documentation prompts or model-facing contracts changed.
