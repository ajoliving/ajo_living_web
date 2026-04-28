# Login Page UI Prompt

## 可直接使用的 Prompt

請建立一個 AJO Living 的登入頁，面向香港高端社區生活平台。頁面要正式、乾淨、可信任，不要做成普通 SaaS marketing 登入頁。桌面版使用左右分割版面：左側大圖主視覺，右側登入表單；手機版隱藏左側圖片，只保留表單。所有介面文案使用繁體中文，品牌名稱 AJO Living 保持英文。

## 整體結構

1. 外層頁面
   - 高度至少為 `100vh - header height`。
   - 背景使用 app surface。
   - 不需要額外 landing hero，不需要卡片堆疊。

2. 桌面 grid
   - 使用兩欄 grid。
   - `lg` 以上：左側約 `1.8fr`，右側 `1fr`。
   - `xl` 以上：左側約 `1.95fr`，右側 `1fr`。
   - 兩欄都撐滿剩餘視窗高度。

## 左側主視覺

1. 顯示規則
   - 只在 `lg` 以上顯示。
   - 手機與平板隱藏。

2. 圖片
   - 使用香港住宅、海港、城市生活相關高質感照片。
   - 圖片鋪滿左側，`object-cover`，置中。
   - 疊加黑色漸層：
     - 左上到右下方向。
     - from black 55%，via black 20%，to transparent。

3. 內容排版
   - 絕對定位覆蓋整張圖。
   - 使用 flex column，top 與 bottom 分離。
   - padding：
     - 桌面 `48px`，top 約 `128px`。
     - 大桌面 `64px`，top 約 `144px`。

4. 上方文案
   - kicker：「AJO Living」，12px，uppercase，letter spacing `0.32em`，白色 70%。
   - H1：「歡迎回家」，display font，48px 至 60px，白色，帶輕微陰影。
   - 描述：「開啟您的理想社區生活，管理發布、聯絡與鄰里交易。」，18px，行高 32px，白色 82%。

5. 下方 metadata
   - 左下顯示兩組資訊：
     - Label：「地點」，值為圖片地點。
     - Label：「圖片」，值為攝影作者。
   - 兩組資訊中間有 1px 高 `40px` 的白色透明分隔線。
   - 右下有 pill：「創立於 2024」。
   - pill 使用白色透明背景、白色透明邊框、backdrop blur。

## 右側表單面板

1. 面板外觀
   - 背景 app surface。
   - 有明顯但柔和的 shadow。
   - 高度填滿。
   - padding：
     - 左右 32px 至 56px。
     - 上方約 `4.75rem` 至 `5.5rem`，避開 header。
     - 底部 24px 至 32px。
   - 內容使用 flex column，頂部品牌、中央表單、底部輔助操作。

2. 頂部品牌
   - 顯示「AJO Living」。
   - display font，24px，uppercase，letter spacing `0.26em`。
   - 下方一條主色短線，寬 `48px`，高 `4px`，圓角。
   - 桌面左對齊，較小螢幕置中。

3. 表單區
   - 最大寬度 `28rem` 左右。
   - 在面板中垂直置中。
   - 標題：「住戶登入」，display font，30px。
   - 說明：「如短訊驗證暫時不可用，可先使用郵箱密碼建立帳戶並登入。」。
   - 說明文字 16px，行高 28px，muted。

## Email 登入模式

1. 欄位
   - 郵箱欄位。
   - 密碼欄位。
   - 註冊模式額外顯示「顯示名稱」欄位。

2. 欄位樣式
   - label 14px，bold，uppercase，letter spacing `0.16em`。
   - input 圓角 `8px`。
   - 背景為 surface-raised。
   - 無邊框，focus 時使用主色 1px ring。
   - padding 約 `16px`，右側保留 icon 空間。
   - icon 放在輸入框右側中央，muted，focus 時變主色。

3. 附加列
   - 左側 checkbox：「記住我」。
   - 右側文字按鈕：「忘記密碼？」。

## SMS 登入模式

1. 欄位
   - 手機號碼。
   - 驗證碼。
   - 驗證碼欄位右側有「取得驗證碼」按鈕。

2. OTP 欄位
   - 桌面或較寬手機可用 `grid-template-columns: 1fr auto`。
   - 按鈕為 border button，rounded 8px。

## 操作區

1. 主按鈕
   - 全寬。
   - 背景主色。
   - 高度約 `56px`。
   - 文字 uppercase，14px，bold，letter spacing `0.24em`。
   - hover 變深色，active 時 scale `0.98`。

2. 分隔線
   - 表單下方有水平線。
   - 中間文字：「其他方式」。
   - 文字 uppercase，12px，letter spacing `0.22em`。

3. 切換按鈕
   - 兩個等寬 secondary button。
   - 第一個切換「短信登入 / 郵箱登入」，帶 phone 或 user icon。
   - 第二個切換「註冊 / 登入」，帶 shield icon。
   - border `1px`，圓角 8px，hover 時背景變 surface-raised。

4. 底部文字
   - 顯示「還沒有帳戶？」或「已經有帳戶？」。
   - 後接文字按鈕「註冊」或「登入」。
   - 若已登入可顯示「登出帳戶」。

## 視覺規則

- 不使用表情符號。
- 不要放多餘功能介紹。
- 不要使用卡片包住整個表單，右側面板本身就是表單容器。
- 控制文字密度，避免輸入框與按鈕擠在一起。
- 表單在不同高度下不能溢出視窗。

## 響應式要求

- `lg` 以下：
  - 隱藏左側圖片主視覺。
  - 表單面板寬度 100%。
  - 品牌與標題可置中。
- 窄螢幕：
  - 表單左右 padding 約 `32px`。
  - 切換按鈕可以保持兩欄，但文字不能擠壓重疊。
