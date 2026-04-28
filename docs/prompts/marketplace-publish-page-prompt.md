# Marketplace Publish Page UI Prompt

## 可直接使用的 Prompt

請建立一個「香港高端社區二手交易平台」的新增帖子頁，用於發布、編輯與預覽二手商品。頁面要正式、乾淨、克制，整體風格要與精品電商的商品管理後台一致。所有固定 UI 文案使用繁體中文，商品標題、描述、價格、地點屬於資料內容，可使用假資料。不要加入教學式說明、行銷語句或多餘提示。

## 整體版面

1. 外層頁面
   - 內容容器最大寬度 `1280px`，水平置中。
   - 桌面 padding 約 `16px 32px 80px`，手機 padding 約 `40px 20px`。
   - 背景使用 app surface，例如 `#F9F9F7`。
   - 主要文字色 `#1A1C1B`，主色 `#002727`。

2. 頁面標題
   - 上方小標籤：「發布流程」或相近文案。
   - 小標籤 12px，font-weight 700，letter-spacing `0.1em`，uppercase，顏色 `#717878`。
   - H1 使用 display font，字級 `clamp(2rem, 4vw, 3rem)`，顏色 `#002727`，font-weight 500。

3. 主工作區
   - 桌面版左右兩欄 grid：
     - 左側表單 `minmax(0, 2fr)`。
     - 右側 sticky preview `minmax(20rem, 1fr)`。
   - 欄距 `32px`。
   - 小於 `1024px` 時改為單欄，preview 不 sticky。

## 表單區塊

1. 區塊卡片
   - 每個步驟是一張白色卡片。
   - 邊框 `1px solid #E2E3E1`。
   - 圓角 `12px`。
   - 背景 `#FFFFFF`。
   - padding `32px`。
   - 陰影 `0 10px 30px -5px rgba(0, 39, 39, 0.05)`。
   - 區塊之間 gap `32px`。

2. 區塊標題列
   - 左側為圓形步驟數字，尺寸 `32px`，背景 `#002727`，白字，font-weight 800。
   - 右側上方為 kicker，12px uppercase，顏色 `#717878`。
   - 右側標題使用 display font，24px，line-height 1.4，顏色 `#002727`。

3. 欄位 grid
   - 基礎 grid gap `24px`。
   - 桌面兩欄，重要長欄位可跨兩欄。
   - 手機一欄。
   - 尺寸欄位可使用內層兩欄 grid，gap `14px`。

## 可複用表單控制元件主題

此區塊可直接作為後續新增頁、編輯頁、後台表單的通用控制元件規格。

1. Label
   - 欄位 label 使用 12px。
   - font-weight 700。
   - letter-spacing `0.1em`。
   - text-transform uppercase。
   - 顏色 `#717878`。
   - label 與控制元件間距約 `8px`。

2. Text input / Number input / Tel input
   - 高度 `48px`。
   - 寬度 100%。
   - 邊框 `1px solid #E2E3E1`。
   - 圓角 `8px`。
   - 背景 `#FFFFFF`。
   - 文字色 `#1A1C1B`。
   - 字級 `15px`，line-height 1.6。
   - 左右 padding `16px`。
   - placeholder 顏色使用 `#717878` 並降低透明度至約 72%。
   - focus 狀態：
     - 邊框改為 `#002727`。
     - 外圈使用 `0 0 0 3px rgba(0, 39, 39, 0.16)`。
   - disabled 狀態：
     - 背景使用白色混合頁面淺色，例如 `#FFFFFF` 與 `#F9F9F7`。
     - 文字 `#717878`。
     - cursor `not-allowed`。

3. Textarea
   - 寬度 100%。
   - 最小高度 `160px`。
   - 邊框、圓角、背景、字級、focus ring 與 input 一致。
   - padding `14px 16px`。
   - resize 只允許 vertical。

4. Custom select trigger
   - 不使用瀏覽器原生 select 外觀，使用 button 模擬。
   - 高度或最小高度 `48px`。
   - display flex，左右分佈。
   - gap `12px`。
   - 邊框 `1px solid #E2E3E1`。
   - 圓角 `8px`。
   - 背景 `#FFFFFF`。
   - padding `0 16px`。
   - 選中文字 15px，font-weight 600，顏色 `#1A1C1B`。
   - 選中文字需單行省略，避免與 chevron icon 重疊。
   - chevron icon 固定在右側，尺寸 16px，不要與文字距離過大。
   - hover 與 expanded 狀態：
     - 邊框 `#002727`。
     - 外圈 `0 0 0 3px rgba(0, 39, 39, 0.16)`。

5. Custom select menu
   - menu 絕對定位在 trigger 下方，`top: calc(100% - 1px)`。
   - left/right 為 0，寬度與 trigger 對齊。
   - z-index 至少 10。
   - 背景 `#FFFFFF`。
   - 邊框 `1px solid #E2E3E1`。
   - 圓角 `0 0 10px 10px`。
   - padding `8px`。
   - gap `4px`。
   - 最大高度 `224px`，內容過多時垂直滾動。
   - 陰影 `0 16px 36px rgba(0, 0, 0, 0.12)`。
   - scrollbar 要細且低調。

6. Custom select option
   - 每個 option 高度自然，padding `10px 11px`。
   - display flex，align center，gap `10px`。
   - 圓角 `8px`。
   - 背景預設 `#FFFFFF`。
   - 文字 14px，font-weight 700，line-height 1.3。
   - hover 背景 `#F4F4F2`。
   - active 狀態文字色 `#002727`，背景同 hover。

7. Multi-select
   - trigger 與 custom select trigger 完全一致。
   - 已選多個選項時，用 ` / ` 串接。
   - 選中文字仍保持單行省略。
   - menu 內 option 使用 checkbox + label。
   - checkbox 尺寸 `16px`，accent color `#002727`。

8. Checkbox row
   - 用於「可捐贈」這類獨立勾選項。
   - 高度或最小高度 `48px`。
   - display flex，align center。
   - gap `10px`。
   - 邊框 `1px solid #E2E3E1`。
   - 圓角 `8px`。
   - 背景 `#FFFFFF`。
   - padding `12px 16px`。
   - 文字 15px，font-weight 700，顏色 `#1A1C1B`。
   - checkbox 尺寸 `16px`，圓角 `4px`，選中顏色 `#002727`。

9. Toggle switch
   - 外層尺寸 `44px x 24px`。
   - rail 背景未選 `#E2E3E1`，選中 `#002727`。
   - thumb 尺寸 `18px x 18px`。
   - thumb 白色，陰影 `0 2px 6px rgba(0, 0, 0, 0.16)`。
   - thumb 左右位移使用 `translateX(20px)`。
   - 動畫時間 `0.2s`。

10. Inline toggle field
   - 用於「允許站內聊天」等欄位。
   - 高度固定 `48px`，與 input、select 完全等高。
   - display flex，左右分佈。
   - 邊框、圓角、背景與 input 一致。
   - padding `0 16px`。
   - 文字 14px，font-weight 700，顏色 `#717878`。

11. Visibility pill option
   - border `1px solid #E2E3E1`。
   - 圓角 `9999px`。
   - padding `9px 15px`。
   - 文字 14px，font-weight 700。
   - 未選文字 `#717878`，背景 `#FFFFFF`。
   - 選中背景 `#002727`，文字 `#FFFFFF`，邊框 `#002727`。

12. Primary action button
   - 高度或最小高度 `52px`。
   - 圓角 `9999px`。
   - 背景 `#002727`。
   - 邊框 `1px solid #002727`。
   - 文字白色，14px，font-weight 800。
   - icon 與文字 gap `8px`。
   - hover 背景可變為 `#1A1C1B`。
   - disabled opacity 0.45，cursor `not-allowed`。

13. Secondary action button
   - 高度與 primary button 一致。
   - 圓角 `9999px`。
   - 背景 transparent。
   - 邊框 `1px solid #002727`。
   - 文字 `#002727`。
   - hover 背景 `#F4F4F2`。

## 新增帖子欄位內容

1. 基本資料
   - 標題 input。
   - 分類 custom select。
   - 價格模式 custom select，選項包含「固定價格」「可議價」。
   - 價格 number input。
   - 成色 custom select。
   - 地區 custom select。
   - 地區右側或同一 grid 末尾放「可捐贈」checkbox row。

2. 商品詳情
   - 摘要 input。
   - 描述 textarea。
   - 尺寸輸入四欄或兩欄：
     - 長度。
     - 寬度。
     - 高度。
     - 重量。

3. 圖片媒體
   - 使用大面積 dropzone。
   - 選擇或拖拽圖片後先在前端暫存 File，建立本地預覽。
   - 不要在選取圖片時直接表達為已上傳到 OSS；只有發布帖子時才上傳。
   - dropzone 最小高度 `208px`。
   - 邊框 `2px dashed #E2E3E1`。
   - 圓角 `12px`。
   - 背景 `#FFFFFF`。
   - hover 或拖拽 active 時邊框變 `#002727`，icon 變 `#002727`。
   - 圖片預覽 grid 桌面兩欄，gap `16px`。
   - 圖片槽高度 `160px`，圖片 `object-cover`。
   - 封面圖邊框使用 `#002727`，左上角 badge 使用深墨綠背景白字。
   - 刪除按鈕為右上角白底圓形，紅色文字，尺寸約 `28px`。

4. 可見範圍與聯絡方式
   - 頂部放一張 visibility card：
     - 左側標題與一句輔助文字。
     - 右側 toggle switch。
     - 卡片邊框、圓角、背景與 input 一致。
   - 下方放 visibility pill options。
   - 聯絡電話 input。
   - 交收地點 input。
   - 交收方式 multi-select。
   - 允許站內聊天 inline toggle field，高度必須與 input/select 一致。

## 右側 Preview Panel

1. Preview 容器
   - 桌面 sticky，top 為 header offset 加 `32px`。
   - gap `16px`。
   - 小於 `1024px` 改為 static。

2. 預覽卡
   - 邊框 `#E2E3E1`。
   - 圓角 `12px`。
   - 背景 `#FFFFFF`。
   - 陰影與表單卡片一致。
   - 圖片區比例 `4 / 3`，背景 `#F4F4F2`。
   - 有圖時 object-cover，hover 輕微 scale。
   - 無圖時顯示 picture icon，顏色 `#717878`。
   - 卡片內容 padding `18px`。
   - 商品標題 18px display font，單行省略。
   - meta 13px，font-weight 700，顏色 `#414848`。
   - 價格 19px，font-weight 800，顏色 `#002727`。

3. 發布檢查清單
   - 使用白底卡片。
   - padding `20px`。
   - 每列 icon + 文字。
   - 未完成顏色 `#717878`。
   - 完成顏色 `#002727`。
   - 字級 14px，font-weight 600。

4. 操作按鈕
   - 上方主按鈕「發布」。
   - 下方次按鈕「儲存草稿」。
   - 兩個按鈕寬度 100%，高度 `52px`。

## 視覺規則

- 主色：深墨綠 `#002727`。
- 文字深色：`#1A1C1B`。
- 次文字：`#414848`。
- 輔助文字：`#717878`。
- 邊框：`#E2E3E1`。
- 淺背景：`#F4F4F2`、`#F9F9F7`。
- 表單控制元件統一使用 48px 高度與 8px 圓角。
- 頁面要像可正式上線的發布管理頁，不要像概念稿。
- 不使用表情符號。
- 不使用大段說明性空話。
- 所有文字不得與 icon 重疊，select chevron 必須靠近欄位右側且與文字保持合理距離。

## 響應式要求

- 小於 `1024px`：
  - 主工作區改為單欄。
  - 右側 preview 取消 sticky。
- 小於 `768px`：
  - 表單卡片 padding 改為 `24px`。
  - 欄位 grid 全部單欄。
  - 圖片預覽 grid 單欄。
  - 頁面左右 padding 約 `20px`。
