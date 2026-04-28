# Email 登入與 SMTP 設定

## 1. 功能說明

後端支援 Email OTP 登入流程：

1. 前端呼叫 `POST /api/v1/auth/email/otp/request` 申請驗證碼。
2. 後端將驗證碼保存 5 分鐘。
3. `MAIL_ENABLED=true` 時，後端透過 SMTP 發送驗證碼郵件。
4. `MAIL_ENABLED=false` 時，後端回傳 `mock_code`，方便本地開發測試。
5. 前端呼叫 `POST /api/v1/auth/email/otp/verify` 驗證並登入；首次登入會自動建立會員。

## 2. 環境變數

```env
MAIL_ENABLED=true
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=no-reply@example.com
SMTP_PASSWORD=your-smtp-password
SMTP_FROM=AJO Living <no-reply@example.com>
SMTP_HELLO_NAME=example.com
```

## 3. 參數說明

- `MAIL_ENABLED`：設為 `true` 時啟用真實 SMTP 發送；本地可設為 `false`。
- `SMTP_HOST`：SMTP 伺服器主機名。
- `SMTP_PORT`：SMTP 伺服器連接埠，常見為 `25`、`465`、`587`。
- `SMTP_USERNAME`：SMTP 登入帳號；本機 Postfix 中繼可留空。
- `SMTP_PASSWORD`：SMTP 登入密碼或應用程式密碼。
- `SMTP_FROM`：寄件人地址，建議使用產品正式域名下的發信地址。
- `SMTP_HELLO_NAME`：SMTP HELO 名稱，建議填發信域名。

## 4. 域名建議

如果 AJO Living 已有正式品牌域名，建議優先使用同一品牌域名的發信地址，例如：

```text
no-reply@ajoliving.com
support@ajoliving.com
```

若擔心交易郵件影響主域名信譽，可使用子域名隔離：

```text
no-reply@mail.ajoliving.com
```

不建議為同一產品另註冊一個完全不同的新域名，除非要做品牌隔離或跨業務隔離。無論使用主域名或子域名，都需要配置 SPF、DKIM、DMARC，否則驗證碼郵件容易進垃圾箱。
