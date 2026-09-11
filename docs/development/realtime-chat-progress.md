# AJO Living 即時聊天實施進度

最後更新：2026-08-27

## 當前狀態

狀態：生產基線已完成核心連線、消息幂等、資料庫 outbox、可恢復的附件狀態事件、孤兒清理及 OSS multipart 續傳；命令模式媒體 worker 會下載原始檔至受控暫存目錄、掃描與轉碼，再回寫及驗證 OSS 衍生檔。實際 ffmpeg/ClamAV 部署、監控接入和容量演練仍待驗收。現有 REST 聊天與 OSS 上傳保持兼容。

## 執行清單

### 第一階段：實時文字消息

- [x] 新增短時 WebSocket ticket，避免在 URL 傳遞長效 access token。
- [x] 新增 WebSocket 連線入口與心跳、讀寫 deadline、斷線重連約定。
- [x] 重用 `ChatService` 和 `BuildingChatService` 的現時權限校驗。
- [x] 新增資料庫 outbox，保證消息提交與附件最終狀態事件一致，並由 worker 重試發布。
- [x] 接入 Redis Pub/Sub，支持多個 Go 實例互相廣播。
- [x] 增加按消息游標補回接口，處理斷線期間消息。
- [x] 前端接入連線、重連、去重及補消息。
- [x] REST/WebSocket 發送統一帶 `client_message_id`，服務端按發送者和會話做幂等去重。
- [x] 通過現有後端服務測試與前端聊天測試。

### 第二階段：聊天附件

- [x] 新增 `message_attachments` 模型、遷移及索引。
- [x] 擴展既有 OSS 預簽名與完成接口，使用 `chat_attachment` 目錄並校驗附件所有權與類型。
- [x] 前端支援圖片、視頻、PDF、純文字附件上傳與基本展示；單檔上限 50 MB，單次最多 5 個文件。
- [x] 建立孤兒媒體清理流程，按 24 小時保留期清理未綁定聊天上傳。

### 第三階段：媒體 worker 與審核

- [x] 支援大視頻分片或斷點續傳，包含已完成分片查詢、單片重試及取消清理。
- [x] 建立異步視頻處理鏈路；命令模式會下載原始檔、生成封面與播放檔、回寫並驗證 OSS。多個 worker 使用資料庫短租約避免重複處理，租約過期可接管；開發與測試預設 `mock`，其他環境未配置時安全拒絕。
- [x] 建立病毒掃描適配器邊界；開發與測試預設 `mock`，生產與預發布強制使用 ClamAV 相容掃描器。
- [x] 部署腳本與配置校驗已固定生產／預發布必須使用可執行的 ffmpeg 與 ClamAV；實際伺服器二進制安裝與輸出驗證仍需在部署驗收時完成。
- [x] 在消息層展示 `processing`、`ready`、`rejected` 及掃描狀態，媒體完成後以獨立 outbox 事件更新在線會話。

### 第四階段：生產驗收

- [x] 部署腳本已生成獨立 WebSocket upgrade、長 timeout 與關閉緩衝配置；實際 Nginx reload、部署與回滾仍需線上驗收。
- [ ] 兩實例跨節點消息壓測。
- [ ] Redis 故障、服務重啟、客戶端斷線及 outbox 重試演練。
- [ ] 大廈綁定撤銷、禁言、封禁、離開群組與附件權限回歸測試。
- [ ] 監控告警、日誌脫敏及容量基線記錄。

監控與容量的具體指標、門檻、採集來源及壓測記錄格式以 [`realtime-chat-production-acceptance.md`](realtime-chat-production-acceptance.md) 為唯一驗收口徑。當前 Go 服務提供結構化 access log、worker log、liveness `/api/v1/health` 及資料庫 readiness `/api/v1/health/ready`，尚未提供 Prometheus `/metrics`；在補齊前不得將監控告警標記為完成。

## 本次已開始的工作

- 已核對現有 `Chat`、`ChatParticipant`、`Message`、大廈群聊服務、OSS 預簽名上傳與 Redis 配置。
- 已固定資料庫真相、Redis 廣播、outbox 補償、WebSocket 不傳文件二進制的架構規則。
- 已完成短時 ticket、WebSocket 連線入口、文字消息事件、Redis 廣播、游標補回及前端重連接入。
- 已完成聊天附件鏈路、multipart 續傳、可恢復附件狀態 outbox、附件狀態 worker 與孤兒媒體清理；真正的 ffmpeg 與 ClamAV 二進制仍需在生產 worker 配置並驗收。

## 暫不納入

- 不新增 WebRTC、語音視頻通話。
- 不把二手交易聊天室拆成獨立通信資料模型。
- 不在第一階段引入 Kafka、Kubernetes 或外部專業 IM 供應商。
