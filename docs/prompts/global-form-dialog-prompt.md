# Global Form Dialog Prompt

## 可直接使用的 Prompt

請為 AJO Living 專案建立或重構一個「可填寫資訊的表單彈窗」。彈窗必須沿用目前專案主題、後台管理頁風格與既有互動規則。介面固定文案使用繁體中文或 English，不使用簡體中文，不加入教學式說明、行銷語句、裝飾性文案或多餘提示。

## 適用範圍

1. 新增、編輯、續期、設定、審核、發放、上傳等需要填寫資訊的彈窗。
2. 任何有必填欄位、可保存資料、可能產生未儲存內容的彈窗。
3. 不適用於純確認刪除、純提示訊息、toast 或無表單內容的小彈窗。

## 彈窗結構

1. 遮罩
   - 使用固定定位覆蓋全屏。
   - 背景為低透明深色遮罩，允許輕微 blur，但不要做強烈玻璃效果。
   - 桌面與手機都需要保留 `1rem` 以上安全邊距。

2. 彈窗面板
   - 圓角不超過 `8px`。
   - 使用專案 surface 顏色與 border token。
   - 寬度按內容類型選擇：
     - 小型確認：`min(100%, 30rem)`。
     - 一般表單：`min(100%, 42rem)`。
     - 複雜表單或上傳預覽：`min(100%, 56rem)`。
   - 最大高度使用 `calc(100dvh - 2rem)`。
   - 面板本身 `overflow: hidden`，滾動只放在 body 區域。

3. Header
   - 左側顯示短 kicker 與標題。
   - kicker 只用於模組名或狀態，例如「廣告任務」「會員設定」。
   - 標題短而具體，例如「發布列表廣告」「編輯會員資料」。
   - 右側固定 close icon button。
   - close button 需要有 `aria-label`，尺寸約 `2.25rem`，不得只放文字。

4. Body
   - 必須可垂直滾動。
   - 使用 `overflow-y: auto`、`overscroll-behavior: contain`。
   - 滾動條要細且低調。
   - 內容 gap 約 `0.8rem - 1rem`。
   - 不要讓 footer 或底部按鈕被視窗裁掉。

5. Footer / Actions
   - 表單提交按鈕與取消按鈕放在底部操作區。
   - 對於長表單，操作區可以 sticky 在 body 底部。
   - 主操作按鈕靠右，手機版改為全寬堆疊。
   - 主按鈕只做一件事，例如「建立」「更新」「儲存」。
   - 次按鈕用於取消、重置、留在此頁，不要和主按鈕視覺權重相同。

## 表單欄位規則

1. 必填標識
   - 所有必填欄位 label 後方顯示紅色 `*`。
   - `*` 只標真正阻止提交的欄位。
   - 可選欄位不要標 `*`，也不要寫長段說明。

2. 欄位校驗
   - 表單提交時必須先做前端校驗。
   - 不完整時禁止送出 API。
   - 每個錯誤要顯示在對應欄位下方。
   - 同時可以用 toast 顯示一句總錯誤，但不能只靠 toast。
   - 錯誤狀態需要改變欄位 border，並顯示紅色錯誤文字。

3. Label
   - label 使用 12px 左右，font-weight 700。
   - 顏色使用 text-muted token。
   - 與控制元件間距約 `0.45rem`。
   - 不要使用過長 label；必要資訊放 placeholder 或錯誤文字。

4. Input / Textarea / Select
   - 高度約 `2.75rem`。
   - border radius `8px`。
   - 使用專案 surface、border、text token。
   - focus 狀態要有清晰 border 與柔和 ring。
   - disabled 狀態要降低透明度，cursor 使用 `not-allowed`。
   - textarea 只允許 vertical resize。

5. Select
   - 若使用原生 select，至少要和 input 高度、圓角、字級一致。
   - 若做 custom select，trigger 與 menu 必須對齊，menu 最大高度約 `14rem` 並可滾動。
   - option 需要 hover、active、selected 狀態。

6. Upload
   - 上傳欄位要顯示目前狀態：未選擇、已選擇、上傳中。
   - 限制檔案類型時，選錯檔案要立刻提示。
   - 已有媒體時要顯示預覽。
   - 圖片使用 `object-fit: contain` 或按業務場景使用 `cover`，不要拉伸。

7. 數字欄位
   - 必須設定 `min`、`max`、`step`。
   - 顯示錯誤時用穩定文案，例如「請輸入有效天數」。

## 未儲存關閉

1. 有草稿內容時，點擊 close、遮罩、路由離開都要二次確認。
2. 二次確認彈窗使用全局 `AppUnsavedChangesDialog` 或同等視覺規格。
3. 二次確認提供三個動作：
   - 保存並離開。
   - 不保存離開。
   - 留在此頁。
4. 正在保存或上傳時禁止關閉主彈窗。
5. 空表單關閉不需要二次確認。

## 佈局規則

1. 桌面彈窗表單可用兩欄或三欄，但主標題、摘要、上傳、預覽等長內容應跨欄。
2. 手機版一律單欄。
3. 欄位不要過窄，避免下拉框、按鈕、錯誤文字擠壓。
4. 指標、統計資料可放在表單頂部，用小型 metric grid 呈現。
5. 不要把卡片套卡片；彈窗面板內的表單區域保持簡潔。

## 交互狀態

1. loading：表單資料載入時在 body 顯示簡短 loading 狀態。
2. saving：主按鈕 disabled，文字切換為「儲存中」或對應動詞。
3. uploading：同 saving 一樣阻止關閉與重複提交。
4. success：成功後關閉彈窗並刷新列表或更新當前資料。
5. error：API 錯誤使用 toast；欄位錯誤使用欄位下方錯誤文案。

## 實作要求

1. Vue 使用 `<script setup lang="ts">`。
2. 表單狀態與彈窗開關由 page 或 page composable 集中管理。
3. 不把表單校驗散落在 template 裡。
4. 新增欄位必須同步型別、保存 payload、校驗、i18n 文案。
5. 不使用 `any`、`// @ts-ignore`。
6. 不引入新的表單庫，除非專案已有統一表單方案。
7. 不啟動前端或後端服務，除非使用者明確要求。

## 驗收標準

1. 空表單提交時，必填欄位下方有紅色錯誤提示。
2. 必填欄位 label 有紅色 `*`。
3. 表單內容超出視窗時 body 可滾動，底部操作按鈕可點。
4. 有草稿時關閉會出現二次確認。
5. 保存中或上傳中不可重複提交、不可關閉。
6. 手機寬度下不出現橫向溢出。
7. `npm run build` 需要通過；涉及後端欄位時跑對應 Go 測試。
