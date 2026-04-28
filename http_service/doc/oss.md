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
