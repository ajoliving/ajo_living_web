# 阿里雲國內短信驗證碼

## 1. 範圍

目前手機 OTP 只接入阿里雲國內短信，僅傳送至 `+86` 中國大陸 11 位手提電話號碼。既有 `POST /api/v1/auth/otp/request` 與 `POST /api/v1/auth/otp/verify` 路徑不變。

## 2. 生產環境設定

```env
OTP_PROVIDER=aliyun_sms
OTP_RESEND_COOLDOWN=1m
ALIYUN_SMS_ACCESS_KEY_ID=
ALIYUN_SMS_ACCESS_KEY_SECRET=
ALIYUN_SMS_REGION_ID=cn-hangzhou
ALIYUN_SMS_SIGN_NAME=
ALIYUN_SMS_TEMPLATE_CODE_CN=
ALIYUN_SMS_TEMPLATE_PARAM_NAME=code
ALIYUN_SMS_REQUEST_TIMEOUT=10s
```

- `ALIYUN_SMS_SIGN_NAME` 必須是已審核通過的國內短信簽名。
- `ALIYUN_SMS_TEMPLATE_CODE_CN` 必須是已審核通過的驗證碼模板 Code。
- 目前模板參數使用 `code`，即短信模板中的 `${code}`。若阿里雲模板使用其他參數名，將 `ALIYUN_SMS_TEMPLATE_PARAM_NAME` 設為該名稱。
- AccessKey 只可放在部署環境的 `.env` 或密鑰管理服務，禁止寫入 Git、前端或 API 回應。

## 3. 行為與限制

- 驗證碼有效期為 5 分鐘，驗證成功後立即失效。
- 同一手機號碼 60 秒內不可重複申請；阿里雲拒絕傳送時不會保存驗證碼。
- 目前 OTP 暫存於單一後端進程記憶體。重啟後未驗證碼會失效；將來擴展多個 API 實例時，必須遷移至 Redis 或資料庫。
