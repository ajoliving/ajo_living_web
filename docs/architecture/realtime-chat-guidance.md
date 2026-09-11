# AJO Living 即時聊天技術指導

## 1. 目標與範圍

本方案用於 AJO Living 平台級通信能力，覆蓋一對一聊天與大廈群聊。首要目標是支援數百名同時在線會員，並保留大廈、單位、身份與權限的服務端校驗。

本方案不把聊天綁定到二手交易語義，也不使用 WebRTC 傳送普通圖片、文件或視頻。WebRTC 僅在未來需要語音或視頻通話時另行評估。

本文件採生產基線，不以一次性原型為目標。後續擴容主要增加 Go 實例、Redis/PostgreSQL 資源與媒體 worker，不更換聊天核心資料模型。

## 2. 固定架構

```text
Vue 3
  │ HTTPS / WebSocket
Nginx
  │
Go Gin API + WebSocket handler
  ├── PostgreSQL：Chat、ChatParticipant、Message、附件關聯、outbox
  ├── Redis Pub/Sub：跨 Go 實例即時廣播
  └── Alibaba Cloud OSS：圖片、文件、視頻原始檔與處理後版本
```

- PostgreSQL 是消息、未讀數、成員狀態與附件關聯的唯一真相。
- Redis Pub/Sub 只負責把已提交的消息轉發到其他服務實例，不作持久化消息庫。
- 消息建立與附件處理狀態變更各自和對應 outbox 事件使用同一個資料庫交易；發布失敗由 worker 以指數退避重試。
- 客戶端斷線重連後，必須按 `after_message_id` 或等價游標從 PostgreSQL 補回消息。
- 初次打開會話與沒有游標的重連使用 `latest=true` 讀取最近一頁，服務端返回內容仍按時間正序排列。
- WebSocket 只傳送小型 JSON 事件，不傳送文件二進制。

## 3. 連線與鑑權

1. 前端先使用現有登入態請求短時 WebSocket ticket。
2. ticket 只允許建立短時連線，不把長效 access token 放入 WebSocket URL。
3. 建立連線時校驗 ticket、會員狀態及聊天訪問權限。
4. 大廈群聊每次訂閱、讀取、已讀、離開及發送前，繼續重新校驗目前大廈綁定或有效授權。
5. 每條 WebSocket 連線只訂閱一個會話，並使用 ping/pong 心跳、讀寫 deadline 與斷線退避重連。
6. access token 過期時，前端先使用 refresh token 輪換新的 access/refresh token 並重試一次原請求；輪換失敗才清除登入態。

## 4. 消息流程

```text
客戶端發送命令
  → ChatService 重新校驗權限、禁言、封禁及內容限制
  → 交易寫入 Message、更新 Chat 與未讀數、寫入 outbox
  → worker 發布 Redis
  → 所有持有該會話連線的 Go 實例推送事件
```

每條客戶端消息必須帶 `client_message_id`，服務端以聊天、發送者及客戶端鍵建立唯一約束，避免網絡重試造成重複消息。服務端返回正式 `message_id`、創建時間及狀態，排序以服務端序列或時間加 ID 為準。

## 5. 媒體上傳與展示

1. 客戶端請求聊天附件的 OSS 預簽名上傳地址。
2. 服務端校驗聊天訪問權限、MIME、大小、文件數量與會員配額。
3. 瀏覽器直接上傳 OSS；較大視頻使用分片或斷點續傳。
4. 上傳完成後回寫 `MediaAsset`，再建立帶附件引用的聊天消息。
5. 新增 `message_attachments` 關聯表，不把附件資料塞入 `Message.ContentText`。
6. 私有 Bucket 只透過服務端生成短時簽名下載地址。
7. 圖片與文件由媒體 worker 完成存在性校驗和掃描狀態更新；視頻轉碼、封面與播放版本由同一 worker 的 ffmpeg 命令適配器完成，病毒掃描由 ClamAV 命令適配器完成。媒體狀態與每條受影響消息的狀態事件在同一資料庫交易寫入 outbox，不可直接發布 Redis 後遺失。命令模式會使用短時簽名 URL 將原始檔下載到受控暫存目錄，處理完成後回寫並確認 OSS 衍生檔存在。`development`、`dev`、`test` 與 `testing` 預設使用僅供本機驗證的 `mock` 處理與掃描適配器；其他環境預設 `disabled` 並安全拒絕。`production`、`prod` 與 `staging` 必須明確配置 ffmpeg 相容處理器與 ClamAV 相容掃描器，否則服務拒絕啟動；生產環境必須驗收命令輸出。
8. 多個媒體 worker 以 `media_assets` 的短租約聲明任務，同一附件只允許一個 worker 處理；工作程序異常後由租約過期接管。未完成或未綁定消息的孤兒媒體由定時任務清理。

## 6. 安全與運維

- 服務端實際檢測文件 MIME，不信任副檔名或瀏覽器上報值。
- 為圖片、普通文件、視頻設定不同的大小、時長、數量及會員配額。
- 對文件執行病毒掃描；對公開展示內容按需要接入阿里雲內容安全或人工審核。
- WebSocket 連線不使用 sticky session；跨實例消息依靠 Redis 廣播，重連依靠資料庫補償。
- Nginx 必須支援 `Upgrade`、`Connection`、較長 read timeout，並關閉 WebSocket 路徑的代理緩衝。
- 監控活躍連線數、消息寫入延遲、Redis 發布錯誤、outbox 堆積、重連次數、附件失敗率與 OSS 簽名錯誤；指標名稱、預設門檻及容量基線以 [`realtime-chat-production-acceptance.md`](../development/realtime-chat-production-acceptance.md) 為準。
- `/api/v1/health` 僅代表程序存活，`/api/v1/health/ready` 檢查 PostgreSQL 可用；當前服務尚未內置 Prometheus `/metrics`，可先從結構化日誌及資料庫聚合查詢採集。
- 不引入 Kafka、Kubernetes 或獨立聊天微服務；當消息量和跨區域需求明顯上升時再評估 Redis Streams 或專業 IM 服務。

## 7. AJO 現有代碼的改造邊界

- 保留 `Chat`、`ChatParticipant`、`Message` 及大廈群聊的現有權限服務。
- 保留 `/chats/*/messages`、`/building-chats/*/messages` 作為歷史查詢與兼容發送入口。
- 新增平台級 `realtime` 連線入口，不放入 marketplace 私有語義。
- 擴展現有 OSS 預簽名流程，使用 `purpose=chat_attachment` 與獨立 object prefix；消息請求以 `attachment_ids` 引用已完成上傳。
- `message_attachments` 保存消息與媒體資產關聯；單次消息最多 5 個附件，單檔上限 50 MB，支持圖片、視頻、PDF 及純文字文件。
- realtime outbox 與必要索引已落地；聊天媒體狀態 worker 將每條附件更新作為獨立 outbox 事件，Redis 故障後仍可重試；24 小時孤兒清理也已接入，不改消息核心交易模型。

## 8. 驗收基線

- 兩個 Go 實例同時運行時，群聊消息可跨實例到達在線會員。
- Redis 暫時不可用時，已提交消息和附件最終狀態仍可查詢；恢復後 outbox 可重試發布。
- 客戶端斷線後重連，消息不重複且可補回離線期間內容。
- 已撤銷大廈綁定的舊群成員不能透過 WebSocket 或 REST 繼續讀取、已讀、離開或發送。
- 上傳中的大文件不阻塞文字消息，未授權會員不能取得附件下載地址。
- 同一 `client_message_id` 重試只產生一條消息；outbox worker 重啟後可繼續處理未完成事件。
