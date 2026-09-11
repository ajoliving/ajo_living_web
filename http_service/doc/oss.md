# OSS 整合說明

## 1. 上傳流程

後端目前使用不依賴 OSS 服務端 callback 的兩段式流程：

1. 前端呼叫 `POST /api/v1/oss/presign` 取得 `upload_url`、`object_key` 與必要 `headers`。
2. 前端依照回傳的 `upload_url` 與 `headers`，直接以 `PUT` 上傳檔案到 OSS。
3. 上傳成功後，前端呼叫 `POST /api/v1/oss/complete` 登記媒體資產資料。
4. 如需管理已上傳媒體，可使用：
   - `GET /api/v1/oss/assets`
   - `GET /api/v1/oss/assets/{mediaAssetId}`
   - `DELETE /api/v1/oss/assets/{mediaAssetId}`

目前 `/api/v1/uploads/presign` 與 `/api/v1/uploads/complete` 仍保留為相容路徑，但新的標準入口統一使用 `/api/v1/oss/*`。

大檔案可使用可續傳 multipart 流程：

1. `POST /api/v1/oss/multipart/initiate` 建立 upload session，返回 `upload_id`、`object_key`、`upload_token`、`part_size` 及 `part_count`。
2. 每個分片呼叫 `POST /api/v1/oss/multipart/part` 取得該 `part_number` 的預簽名 PUT URL，瀏覽器直接上傳到 OSS。
3. 斷線重連或頁面恢復時呼叫 `POST /api/v1/oss/multipart/parts` 取得已上傳分片，再只補傳缺少的編號。
4. 所有分片完成後呼叫 `POST /api/v1/oss/multipart/complete`，提交 `{part_number, etag}` 陣列；服務端完成 OSS multipart 並登記 `MediaAsset`。
5. 用戶取消或放棄上傳時呼叫 `POST /api/v1/oss/multipart/abort` 清理 OSS 未完成分片。

`upload_token` 綁定用戶、物件、MIME、檔案大小與 `upload_id`，分片 URL 有效期沿用 `OSS_PRESIGN_EXPIRES`。服務端不轉發檔案二進制；前端應在失敗後以相同分片編號重試，並在斷線重連後重新取得未完成分片的簽名 URL。

## 2. 環境變數

```env
OSS_PROVIDER=oss
OSS_BUCKET=your-bucket
OSS_REGION=cn-hongkong
OSS_ENDPOINT=oss-cn-hongkong.aliyuncs.com
OSS_ACCESS_KEY_ID=your-access-key-id
OSS_ACCESS_KEY_SECRET=your-access-key-secret
OSS_SESSION_TOKEN=
OSS_USE_CNAME=false
OSS_DISABLE_SSL=false
OSS_PRESIGN_EXPIRES=15m
OSS_MULTIPART_PART_SIZE=10485760
OSS_MEDIA_BASE_URL=https://your-bucket.oss-cn-hongkong.aliyuncs.com
```

## 3. 參數說明

- `OSS_PROVIDER`：設為 `oss` 時，後端會改用阿里雲 OSS SDK v2 產生預簽名上傳 URL。
- `OSS_BUCKET`：OSS bucket 名稱。
- `OSS_REGION`：bucket 所在 region，例如 `cn-hongkong`。
- `OSS_ENDPOINT`：請填 region endpoint，不要帶 bucket，例如 `oss-cn-hongkong.aliyuncs.com`。
- `OSS_ACCESS_KEY_ID` / `OSS_ACCESS_KEY_SECRET`：阿里雲存取憑證。
- `OSS_SESSION_TOKEN`：若使用 STS 臨時憑證才需要填寫。
- `OSS_USE_CNAME`：若 `OSS_ENDPOINT` 是自定義網域，設為 `true`。
- `OSS_DISABLE_SSL`：只在本地或特殊網路環境需要 HTTP 時設為 `true`。
- `OSS_PRESIGN_EXPIRES`：預簽名有效期，OSS v4 簽名最長為 `7d`。
- `OSS_MULTIPART_PART_SIZE`：multipart 分片大小（bytes），預設 `10485760`（10 MiB）；單一 upload 最多 10,000 個分片。
- `OSS_MEDIA_BASE_URL`：媒體公開訪問基底 URL。若未顯式設定，系統會依 `bucket + endpoint` 自動推導。

舊版 `STORAGE_*` 與 `MEDIA_BASE_URL` 仍可相容讀取，但新環境建議統一使用 `OSS_*` 命名。

## 4. 自定義網域

若你有 CDN 或自定義網域，例如 `cdn.example.com`：

```env
OSS_ENDPOINT=cdn.example.com
OSS_USE_CNAME=true
OSS_MEDIA_BASE_URL=https://cdn.example.com
```

此模式下，SDK 會以 CNAME 方式簽名；`/complete` 回傳的媒體 URL 也應對應同一個公開網域。

## 5. 聊天媒體處理與掃描

聊天視頻由 `MediaProcessingWorker` 異步處理，聊天附件只有在 `processing_status=ready` 且 `scan_status=passed` 時才返回下載 URL。`development` 與 `test` 預設使用 `mock` 適配器，以便驗證完整附件流程；這些適配器不執行真實掃描，不得用於對外環境。`production` 與 `staging` 必須配置 ClamAV 相容掃描器，否則服務拒絕啟動；掃描執行失敗時，附件仍保持安全拒絕。

```env
MEDIA_PROCESSING_PROVIDER=ffmpeg
MEDIA_PROCESSING_TIMEOUT=10m
MEDIA_FFMPEG_BINARY=/usr/bin/ffmpeg
MEDIA_FFMPEG_ARGS=-y -i {source_path} -ss 00:00:01 -frames:v 1 {thumbnail_path} -c:v libx264 -movflags +faststart {playback_path}
MEDIA_SCAN_PROVIDER=clamav
MEDIA_SCAN_TIMEOUT=2m
MEDIA_CLAMAV_BINARY=/usr/bin/clamscan
MEDIA_CLAMAV_ARGS=--no-summary {source_path}
```

命令以無 shell 模式執行，支持 `{object_key}`、`{mime_type}`、`{thumbnail_key}`、`{playback_key}`、`{source_path}`、`{thumbnail_path}` 和 `{playback_path}` 佔位符。命令模式會把原始檔下載至權限受控的暫存目錄，ffmpeg 以本地來源生成封面和播放檔，worker 回寫 OSS 並驗證衍生物件存在後才標記可用；ClamAV 同樣掃描本地原始檔。`mock` 只可用於開發與測試。
