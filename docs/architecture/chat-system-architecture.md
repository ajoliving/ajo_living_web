# AJO Living 即時聊天系統架構說明

## 1. 文件定位

本文件說明目前實際落地的聊天系統架構，涵蓋一對一會話與大廈群聊。內容以現行程式碼為準，描述已經實現的能力、請求與資料邊界，以及現行限制。

本文件不是目標方案。目標方案、驗收基線與容量口徑見：

- [`realtime-chat-guidance.md`](./realtime-chat-guidance.md)：目標架構與改造邊界。
- [`realtime-chat-production-acceptance.md`](../development/realtime-chat-production-acceptance.md)：生產驗收、監控門檻與容量基線。
- [`realtime-chat-progress.md`](../development/realtime-chat-progress.md)：實施進度。

現行實現與目標方案之間的落差集中在第 9 節。

## 2. 整體分層

```text
Vue 3 前端（web-new）
   │  REST /api/v1            WebSocket /api/v1/realtime/ws        OSS 直傳
Nginx
   │
Go Gin 服務
   ├── ChatService            一對一會話與消息
   ├── BuildingChatService    大廈群聊、成員治理與加入審核
   ├── RealtimeHub            WebSocket 連線管理與事件推送
   ├── RealtimeOutboxWorker   事件投遞與指數退避重試
   └── MediaProcessingWorker  媒體轉碼、封面與病毒掃描
   │
   ├── PostgreSQL            消息、成員、未讀數、outbox、媒體資產的唯一真相
   ├── Redis Pub/Sub         跨實例即時廣播，不作持久化消息庫
   └── 阿里雲 OSS             聊天媒體原始檔、封面與播放衍生檔
```

分層原則：

- 一對一與大廈群聊共用同一套會話、成員與消息模型，只以 `chat_type` 區分，不建兩套資料流。
- WebSocket 只傳送小型 JSON 事件，不傳送二進制檔。媒體一律由瀏覽器直傳 OSS。
- Redis 只承擔跨實例廣播。Redis 不可用時消息已落庫，REST 仍可查回。

## 3. 資料模型

| 表 | 職責 | 關鍵欄位與枚舉 |
| --- | --- | --- |
| `chats` | 會話主體 | `public_id`（ULID，對外 ID）、`chat_type`、`biz_module`、`building_id`、`listing_id`、`last_message_preview`、`last_message_at` |
| `chat_participants` | 會話成員與未讀 | `role_in_chat`、`membership_status`、`muted_until`、`banned_until`、`last_read_message_id`、`unread_count` |
| `chat_join_requests` | 群聊加入申請 | `status`、`reason`、`reviewed_by`、`reviewed_at` |
| `messages` | 消息真相 | `public_id`、`sender_user_id`、`client_message_id`、`reply_to_message_id`、`message_type`、`content_text`、`message_status` |
| `message_attachments` | 消息與媒體關聯 | `message_id`、`media_asset_id`、`sort_order` |
| `media_assets` | 媒體資產與處理狀態 | `bucket_name`、`object_key`、`mime_type`、`file_size`、`processing_status`、`scan_status`、`thumbnail_object_key`、`playback_object_key`、`worker_claim_token` |
| `realtime_outboxes` | 事件投遞佇列 | `event_type`、`aggregate_id`、`payload`、`status`、`attempts`、`next_attempt_at` |
| `moderation_actions` | 群成員處置審計 | `target_type`、`target_id`、`action_type`、`reason_text`、`operator_user_id` |

枚舉值：

- `chat_type`：`direct_listing_chat`、`building_group`、`system_notice`。
- `biz_module`：`secondhand`、`property_sale`、`serviced_apartment`、`building`。
- `membership_status`：`active`、`left`、`kicked`、`banned`。
- `role_in_chat`：`inquirer`、`publisher`、`member`、`admin`、`system`。
- `message_type`：`text`、`attachment`、`mixed`；`message_status`：`sent`、`recalled`。
- `processing_status`：`ready`、`processing`、`rejected`；`scan_status`：`pending`、`passed`、`rejected`。
- `realtime_outboxes.status`：`pending`、`processing`、`failed`、`published`。

結構約束：

- `messages` 在 `chat_id + sender_user_id + client_message_id` 上唯一，作為發送冪等鍵。
- `chats` 在 `building_id` 上對 `chat_type = 'building_group'` 建部分唯一索引，保證一棟大廈只有一個群。
- `chat_join_requests` 在 `chat_id + user_id` 上唯一。

## 4. 會話類型與權限

### 4.1 一對一會話

由帖子或物業詳情頁發起，同一對象複用同一會話，不重複建立。入口依業務渠道分成三條：

- `POST /listings/{listingId}/chats`：二手交易。
- `POST /property-sales/{listingId}/chats`：樓盤放售。
- `POST /serviced-apartments/{listingId}/chats`：服務式住宅。

會話以 `biz_module` 標記業務來源，避免把物業渠道的會話混進二手交易語義。

### 4.2 大廈群聊

一棟大廈一個群，由 `POST /buildings/{buildingId}/chat` 保證建立並寫入首位成員。會員端群聊與一對一會話混排在同一會話列表中，不設獨立入口。

服務端以三個來源判定身份與可見範圍：

1. `users.is_staff`：AJO Staff，可管理所有大廈群聊。
2. `building_authorizations` 中 `owner_user_id` 且狀態為 `active`：該大廈的授權管理者，可在該大廈範圍內管理群聊。
3. `visibleBuildingIDs`：一般住戶的大廈可見範圍，由 `user_profiles.bound_building_ids`、`communities.public_id`（綁定狀態非 `pending`）與 `building_authorizations` 的受讓者記錄合成。

### 4.3 校驗點

- 所有聊天 REST 介面要求會員狀態為 active。
- 大廈群聊的每次讀取、已讀、離開、發送、加入審核與成員處置，均在服務端重新計算大廈範圍，不信任已存在的成員記錄。
- WebSocket 廣播前對每條連線重新校驗會話訪問權，綁定失效即斷開連線。
- 群聊管理只限 AJO Staff 或該大廈有效授權管理者；首位加入者不因建立群組而取得管理權。
- 禁言、踢出、封禁與解除封禁寫入 `moderation_actions` 作為審計。

## 5. 消息流程

### 5.1 發送

```text
客戶端發送（REST 或 WebSocket 命令）
  → 服務端校驗會話訪問權、成員狀態、禁言與封禁、內容與附件限制
  → 同一資料庫交易寫入 Message、更新 Chat 預覽、遞增其他成員未讀數、寫入 realtime outbox
  → outbox worker 發布至 Redis
  → 各 Go 實例的 Hub 推送事件給持有該會話連線的客戶端
```

REST 入口為 `POST /chats/{chatId}/messages` 與 `POST /building-chats/{chatId}/messages`；WebSocket 入口為命令型訊息 `{ type: "message", chat_id, content, attachment_ids, client_message_id, reply_to_message_id }`。兩條入口共用同一服務層校驗與交易。

### 5.2 事件格式

```json
{ "type": "message", "chat_id": "...", "message": {}, "client_message_id": "...", "origin": "<實例ID>" }
```

事件類型只有兩種：

- `message`：新消息。附件處理狀態變更（`processing_status`、`scan_status`）也復用此類型重推整條消息。
- `message_recalled`：撤回。事件與回應均不再攜帶原文與附件下載地址。

`origin` 為發送實例標識，用於過濾自身回環。Redis 頻道為 `ajo:realtime:chat-events`。

### 5.3 撤回

發送者可撤回自己兩分鐘內的消息。撤回與發送共用同一交易與 outbox 機制；撤回後消息狀態置為 `recalled`，引用它的消息預覽在服務端與前端同步顯示為已撤回，不另建引用表。

### 5.4 連線與鑑權

1. 前端以登入態請求 `GET /api/v1/realtime/ticket`，取得 90 秒有效的短時票據。
2. 以 `GET /api/v1/realtime/ws?ticket=...&chat_id=...` 建立連線，不把長效 access token 放入 URL。
3. 服務端校驗票據類型、會員狀態、會話訪問權與 Origin 白名單。
4. 一條連線只訂閱一個會話，切換會話時關閉舊連線並重新建立。

## 6. 可靠性設計

- **發送冪等**：`client_message_id` 三欄聯合唯一。網絡重試只會產生一條消息。
- **交易性投遞**：消息寫入與 outbox 事件在同一交易提交，不出現「消息已入庫但事件遺失」。
- **重試**：worker 每秒取一批，以資料庫短租約聲明任務，成功標記 `published`，失敗按指數退避，上限五分鐘。
- **跨實例**：多實例只透過 Redis 廣播通信，不使用 sticky session。
- **斷線補償**：無自增序號，游標為 `after_message_id`（消息 `public_id`）。無游標的首次載入與重連使用 `latest=true` 取最近一頁，服務端按時間正序返回。
- **啟動邊界**：outbox worker 只在 Redis broker 可用時啟動；媒體處理 worker 無條件啟動。

## 7. 媒體鏈路

### 7.1 上傳

1. 前端以 `purpose=chat_attachment` 請求 OSS 預簽名地址，服務端強制對象前綴為 `ajo_living/chat/`。
2. 瀏覽器直傳 OSS；超過 10 MB 走 multipart 分片，分片預設 10 MiB。
3. 上傳完成回寫 `media_assets`，再以 `attachment_ids` 建立帶附件的消息。

限制：

- 單檔上限 50 MB，單條消息最多 5 個附件。
- 允許類型：`image/*`、`video/*`、`application/pdf`、`text/plain`。
- 上傳票據以 HMAC 簽名，綁定用戶、對象鍵、MIME、檔案大小與分片上傳 ID，有效期 15 分鐘。
- 未完成的分片可查詢與取消；超過 24 小時未被消息引用的孤兒媒體由定時任務清理。

### 7.2 處理與掃描

- `video/*` 且位於聊天前綴的資產進入 `processing`，由 ffmpeg 產生封面與播放版本，回寫並複核 OSS 衍生檔存在。
- 掃描由 ClamAV 完成，通過為 `passed`，不通過則整體標記 `rejected` 並記錄原因。
- 任何處理或掃描失敗一律 fail-closed，不把未處理或未掃描對象暴露給客戶端。
- 多個處理程序以 `media_assets` 短租約聲明任務，同一資產只允許一個程序處理，異常後由租約過期接管。
- 媒體狀態變更與受影響消息的事件在同一交易寫入 outbox。
- 生產類環境（`production`、`prod`、`staging`）必須配置可執行的 ffmpeg 與 ClamAV，否則服務拒絕啟動。

### 7.3 展示門禁

附件下載地址只在 `processing_status = ready` 且 `scan_status = passed` 時下發；視頻優先使用播放衍生檔。資產在聊天前綴下而處理或掃描狀態為空的歷史資料不下發地址。

## 8. 通知邊界

聊天消息不寫入通知中心。未讀只由 `chat_participants.unread_count` 與 realtime 事件承載，避免同一則訊息在會話與系統通知兩處重複出現。通知中心只保留單向系統公告與可追蹤的業務提醒。

公告、站內信與聊天分別代表單向發布、可追蹤消息與會話互動，在資料模型與介面上不混用。

## 9. 現行限制與待決事項

| 項目 | 現行狀態 | 影響 |
| --- | --- | --- |
| 聊天附件下載地址 | 由媒體基礎位址與對象鍵拼接，靠處理與掃描狀態門禁 | 未使用短時簽名下載，與目標方案的私有桶簽名口徑不一致 |
| 會員端進群入口 | 首次進群與申請入群介面已實現，前端未接線 | 會員目前無法主動進入或申請某棟大廈群聊 |
| 前端連線心跳 | 未實現，依賴瀏覽器關閉事件觸發退避重連 | 弱網下重連時機不可控，後端 60 秒無回應即判死 |
| 前端狀態承載 | 無全域狀態層，會話、消息與連線生命週期綁定聊天頁 | 切換頁面即斷連，未讀數不跨頁保有 |
| 分片續傳 | 單次發送內可複用已傳分片，發送失敗即取消整個分片上傳 | 跨次發送的續傳能力不可用，大檔弱網場景需重傳 |
| 加入申請重提 | 申請表唯一鍵為 `chat_id + user_id`，不含狀態 | 申請被審核後無法再次提交 |
| 封禁範圍 | 只有會話級 `banned_until`，無用戶級全局封禁 | 同一用戶在其他大廈群聊不受影響 |
| 消息互動 | 有回覆引用、複製與撤回；無提及、輸入中提示、消息級已讀 | 會話級未讀之外缺少更細的已讀回執 |
| 發送限流 | 進程內記憶體計數，每分鐘 30 次 | 多實例部署時限流不共享 |
| 後端測試 | 主分支無聊天服務單元測試 | 聊天交易、權限與媒體狀態機缺少自動化防護 |
| outbox 啟動條件 | 僅在 Redis broker 可用時啟動 worker | 關閉 Redis 時事件持續堆積為 `pending`，需人工處理 |

## 10. 主要介面索引

| 方法 | 路徑 | 用途 |
| --- | --- | --- |
| POST | `/listings/{listingId}/chats` | 建立或複用二手會話 |
| POST | `/property-sales/{listingId}/chats` | 建立或複用樓盤會話 |
| POST | `/serviced-apartments/{listingId}/chats` | 建立或複用服務式住宅會話 |
| GET | `/chats` | 會話列表 |
| GET | `/chats/{chatId}/messages` | 一對一消息查詢，支援游標 |
| POST | `/chats/{chatId}/messages` | 發送一對一消息 |
| POST | `/chats/{chatId}/messages/{messageId}/recall` | 撤回消息 |
| POST | `/chats/{chatId}/read` | 標記已讀 |
| GET | `/building-chats` | 可見大廈群聊列表 |
| POST | `/buildings/{buildingId}/chat` | 建立並加入大廈群聊 |
| GET | `/building-chats/{chatId}/messages` | 群聊消息查詢 |
| POST | `/building-chats/{chatId}/messages` | 發送群聊消息 |
| POST | `/building-chats/{chatId}/read` | 群聊標記已讀 |
| POST | `/building-chats/{chatId}/leave` | 離開群聊 |
| GET | `/building-chats/{chatId}/members` | 群成員名單 |
| POST | `/building-chats/{chatId}/join-requests` | 提交加入申請 |
| POST | `/building-chat-join-requests/{requestId}/review` | 審核加入申請 |
| POST | `/building-chats/{chatId}/members/moderate` | 禁言、解除禁言、踢出、封禁、解除封禁 |
| GET | `/staff/building-chats` | Staff 查看全部群聊 |
| POST | `/staff/building-chats/{chatId}/members` | Staff 加入成員 |
| GET | `/realtime/ticket` | 取得短時連線票據 |
| GET | `/realtime/ws` | 建立 WebSocket 連線 |

## 11. 相關文件

- [`realtime-chat-guidance.md`](./realtime-chat-guidance.md)
- [`realtime-chat-production-acceptance.md`](../development/realtime-chat-production-acceptance.md)
- [`realtime-chat-progress.md`](../development/realtime-chat-progress.md)
- [`realtime-chat-nginx.conf`](../deployment/realtime-chat-nginx.conf)
- [`oss.md`](../../http_service/doc/oss.md)
