# AJO Living OSS CDN Let's Encrypt SSL Renewal Prompt

## 可直接使用的 Prompt

請協助我為 AJO Living 的 OSS CDN 媒體域名重新申請 Let's Encrypt SSL 憑證，並把憑證 PEM 檔下載到我的本機，供我手動上傳到阿里雲 CDN HTTPS 設定。

## 目標域名

```text
ajoliving.oss.skylinedances.com
```

## 目前 CDN 與用途

```text
CDN 加速域名: ajoliving.oss.skylinedances.com
CDN CNAME: ajoliving.oss.skylinedances.com.w.kunlunaq.com
用途: AJO Living 全站圖片與影片公開訪問域名
後端公開媒體基底 URL: OSS_MEDIA_BASE_URL=https://ajoliving.oss.skylinedances.com
```

## 伺服器與本機路徑

```text
網關伺服器: ssh admin@47.83.21.100
伺服器憑證目錄: /etc/letsencrypt/live/ajoliving.oss.skylinedances.com/
本機下載目錄: /Users/yangliu/Downloads/ajoliving-oss-cert/
本機證書檔: /Users/yangliu/Downloads/ajoliving-oss-cert/fullchain.pem
本機私鑰檔: /Users/yangliu/Downloads/ajoliving-oss-cert/privkey.pem
```

## 執行要求

1. 使用 DNS-01 manual challenge 申請或續期憑證，因為 CDN 域名不指向網關伺服器。
2. 不要啟動或重啟前端、後端服務。
3. 不要在對話中輸出 `privkey.pem` 私鑰內容。
4. DNS TXT 記錄新增後，必須先用 `dig` 確認解析值已生效，再繼續 certbot。
5. 申請成功後，把 `fullchain.pem` 和 `privkey.pem` 下載到本機下載目錄。
6. 下載後刪除伺服器 `/tmp` 中轉檔案。
7. 檢查 CDN HTTPS 憑證是否已配置成功，並用真實圖片或影片 URL 驗證返回 `200`。

## 建議指令流程

### 1. 確認目前 CNAME

```bash
dig +short CNAME ajoliving.oss.skylinedances.com
```

期望包含：

```text
ajoliving.oss.skylinedances.com.w.kunlunaq.com.
```

### 2. 在網關伺服器執行 certbot

```bash
ssh admin@47.83.21.100 'sudo certbot certonly --manual --preferred-challenges dns -d ajoliving.oss.skylinedances.com --agree-tos --manual-public-ip-logging-ok --email admin@skylinedances.com'
```

certbot 會要求新增 DNS TXT 記錄，記錄格式如下：

```text
主機記錄: _acme-challenge.ajoliving.oss
記錄類型: TXT
記錄值: 以 certbot 當次輸出的值為準
TTL: 預設或 10 分鐘
```

完整記錄名：

```text
_acme-challenge.ajoliving.oss.skylinedances.com
```

### 3. 確認 TXT 記錄生效

```bash
dig +short TXT _acme-challenge.ajoliving.oss.skylinedances.com
ssh admin@47.83.21.100 'dig +short TXT _acme-challenge.ajoliving.oss.skylinedances.com'
```

兩邊都看到 certbot 要求的 TXT 值後，才回到 certbot 會話按 Enter 繼續。

### 4. 下載 PEM 到本機

```bash
mkdir -p /Users/yangliu/Downloads/ajoliving-oss-cert

ssh admin@47.83.21.100 'sudo cp /etc/letsencrypt/live/ajoliving.oss.skylinedances.com/fullchain.pem /tmp/ajoliving-oss-fullchain.pem && sudo cp /etc/letsencrypt/live/ajoliving.oss.skylinedances.com/privkey.pem /tmp/ajoliving-oss-privkey.pem && sudo chown admin:admin /tmp/ajoliving-oss-fullchain.pem /tmp/ajoliving-oss-privkey.pem && sudo chmod 600 /tmp/ajoliving-oss-fullchain.pem /tmp/ajoliving-oss-privkey.pem'

scp admin@47.83.21.100:/tmp/ajoliving-oss-fullchain.pem /Users/yangliu/Downloads/ajoliving-oss-cert/fullchain.pem
scp admin@47.83.21.100:/tmp/ajoliving-oss-privkey.pem /Users/yangliu/Downloads/ajoliving-oss-cert/privkey.pem

ssh admin@47.83.21.100 'rm -f /tmp/ajoliving-oss-fullchain.pem /tmp/ajoliving-oss-privkey.pem'
chmod 600 /Users/yangliu/Downloads/ajoliving-oss-cert/fullchain.pem /Users/yangliu/Downloads/ajoliving-oss-cert/privkey.pem
```

### 5. 阿里雲 CDN 手動上傳

在阿里雲 CDN 控制台進入：

```text
CDN -> 域名管理 -> ajoliving.oss.skylinedances.com -> HTTPS 設定
```

選擇：

```text
證書來源: 自定義上傳
證書（公鑰）: fullchain.pem
私鑰: privkey.pem
```

### 6. 驗證 CDN HTTPS 生效

```bash
echo | openssl s_client -servername ajoliving.oss.skylinedances.com -connect ajoliving.oss.skylinedances.com:443 2>/dev/null | openssl x509 -noout -subject -issuer -dates
```

確認：

```text
subject=CN=ajoliving.oss.skylinedances.com
issuer 包含 Let's Encrypt
notAfter 為新的到期時間
```

再用真實物件測試：

```bash
curl -I --max-time 15 https://ajoliving.oss.skylinedances.com/ajo_living/eng/home-carousel/secondhand.webp
```

期望：

```text
HTTP/1.1 200 OK
Content-Type: image/webp
```

## 注意事項

1. Let's Encrypt 憑證有效期約 90 天。
2. manual DNS challenge 不會自動續期；到期前需要重跑本流程。
3. 舊 TXT 記錄可在新憑證簽發成功後刪除，避免 DNS 管理混亂。
4. 不要把 `privkey.pem` 放進專案目錄或 Git 倉庫。
5. 根路徑 `https://ajoliving.oss.skylinedances.com/` 返回 OSS `403` 屬於正常現象；應用真實 object key 測試圖片或影片。
