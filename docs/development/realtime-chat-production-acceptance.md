# 即時聊天生產驗收

本文列出部署前後必須執行的驗收，不在本機單元測試中啟動服務或修改線上環境。聊天核心可先上線；媒體轉碼、內容安全尚未完成生產部署驗收前，不得宣稱為完整媒體生產能力。

## 服務與代理

- [ ] 依 [`realtime-chat-nginx.conf`](../deployment/realtime-chat-nginx.conf) 配置 WebSocket `Upgrade`、`proxy_read_timeout` 和 `proxy_buffering off`。
- [x] 發布腳本已內置上述 WebSocket 代理區塊，並在生產主機發布前檢查 `ffmpeg` 與 `clamscan` 可執行；仍需在實際伺服器執行 `nginx -t` 和 reload 驗收。
- [ ] 執行 `nginx -t`，再由 Supervisor 或 Docker 重啟 AJO 後端。
- [ ] 確認兩個 Go 實例共用同一 PostgreSQL、Redis 與 OSS 配置，且 `database.Migrate` 已建立 `realtime_outboxes`、`message_attachments`。
- [ ] `/api/v1/health` 作 liveness（程序存活）探針，`/api/v1/health/ready` 檢查 PostgreSQL 可用；Redis 為可選廣播依賴時，readiness 不因 Redis 短暫故障阻斷，但必須觸發 outbox 告警。

## 功能與故障演練

- [ ] 兩個實例分別建立 WebSocket 連線，跨實例發送文字、圖片與視頻消息。
- [ ] 暫停 Redis，確認消息仍可寫入 PostgreSQL；恢復 Redis 後確認 outbox worker 發布完成。
- [ ] 重啟其中一個實例，確認未發布 outbox 可恢復，客戶端按 `after_message_id` 補回且不重複。
- [ ] 撤銷大廈綁定、禁言、封禁及退出群組，確認 REST 與 WebSocket 均即時拒絕後續操作。

每項演練至少保存：執行時間、release id、實例名稱、請求 ID、預期結果、實際結果及回滾判定。Redis 故障演練不得以清空資料庫代替；消息必須能從 PostgreSQL 查回。

## 媒體處理前置條件

- [ ] 生產媒體 worker 安裝並固定 ffmpeg 版本，視頻附件只有在封面、播放版本生成後標記 `ready`。
- [ ] 配置病毒掃描或內容安全服務；掃描失敗或拒絕時，附件標記 `rejected` 且消息不提供下載地址。
- [ ] 配置 OSS 分片上傳策略與未完成分片清理週期，並驗證超過 multipart 閾值的視頻續傳；聊天附件總大小仍受 50 MB 限制。

## 監控基線

- [ ] 建立下列指標或等價查詢。當前服務尚未內置 `/metrics`，可先由 JSON access log、worker log 與 PostgreSQL 查詢採集；接入 Prometheus 時沿用固定名稱。

| 指標 | 建議名稱 | 告警門檻（預設） | 來源 |
| --- | --- | --- | --- |
| 活躍 WebSocket 連線 | `ajo_realtime_connections` | 單實例 > 80% 連線上限持續 5 分鐘 | hub gauge／連線日誌 |
| 消息寫入延遲 p95 | `ajo_realtime_message_write_seconds` | > 0.5 s 持續 5 分鐘；> 2 s 立即告警 | ChatService transaction log |
| outbox 待處理數 | `ajo_realtime_outbox_pending` | > 1,000 持續 5 分鐘 | `realtime_outboxes` 聚合查詢 |
| outbox 最老事件年齡 | `ajo_realtime_outbox_oldest_age_seconds` | > 60 s 告警；> 300 s 重大告警 | `MIN(created_at)` 查詢 |
| Redis 發布失敗率 | `ajo_realtime_outbox_publish_errors_total` / 發布總數 | 5 分鐘失敗率 > 1% | outbox worker log |
| 客戶端重連率 | `ajo_realtime_reconnect_total` | 10 分鐘內每 100 連線 > 20 次 | 前端 telemetry／Nginx log |
| 附件處理失敗率 | `ajo_chat_attachment_processing_total{status="rejected"}` | 15 分鐘 > 5% 或連續 10 件 | media worker log／`media_assets` |
| HTTP 5xx | `ajo_http_requests_total{status=~"5.."}` | 5 分鐘 > 1% 且請求數 > 100 | access log |

- [ ] 日誌不得記錄 access token、WebSocket ticket、OSS upload token 或私有文件完整 URL；`request_id`、`chat_id`、`message_id` 可記錄，用戶身份只記錄內部 ID。
- [ ] 監控資料保留至少 14 天；告警需包含實例、release、指標值、首次發生時間與 runbook 連結。
- [ ] 未配置監控採集或任一重大告警未收斂時，阻斷生產發布，不以「稍後補監控」替代。

無 `/metrics` 時可先使用以下只讀查詢建立 outbox 與媒體狀態告警：

```sql
SELECT
  count(*) FILTER (WHERE status IN ('pending', 'failed', 'processing')) AS pending_count,
  COALESCE(EXTRACT(EPOCH FROM (now() - min(created_at)))::bigint, 0) AS oldest_age_seconds
FROM realtime_outboxes
WHERE status <> 'published';

SELECT processing_status, scan_status, count(*) AS asset_count
FROM media_assets
GROUP BY processing_status, scan_status;
```

## 容量基線

上線前用兩個 Go 實例、共用 PostgreSQL/Redis/OSS 進行 30 分鐘穩態壓測，另做 10 分鐘突發測試。固定記錄以下結果，作為下一次擴容比較基線：

| 項目 | 必填記錄 |
| --- | --- |
| 連線 | 每實例最大穩定連線、CPU、記憶體、斷線率 |
| 消息 | 每秒寫入量、p50/p95/p99 寫入延遲、重試率 |
| PostgreSQL | `max_connections`、連線池上限/峰值、CPU、慢查詢數、`messages` 增長量 |
| Redis | 記憶體峰值、Pub/Sub client 數、發布延遲、拒絕或斷線次數 |
| OSS | 上傳吞吐、4xx/5xx、簽名錯誤、單日新增容量 |
| 媒體 | `processing` 最長等待、成功/拒絕數；若未接入 ffmpeg/掃描則明確標記為未測 |

容量判定：穩態期間 p95 消息寫入 < 0.5 s、WebSocket 斷線率 < 1%、outbox 最老事件 < 60 s、HTTP 5xx < 1%；任一不滿足需先降並發或擴容，再進行發布。依實測結果把單實例安全連線上限設為峰值的 80%，不可直接沿用開發機數值。
