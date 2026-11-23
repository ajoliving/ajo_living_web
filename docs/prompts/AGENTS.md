# docs/prompts

> Level/parent: ../AGENTS.md

## 成員清單
ajo-living-public-ui-style-prompt.md: AJO Living 公開頁 UI 基線 prompt。
global-form-dialog-prompt.md: 全域表單彈窗結構、欄位、狀態與驗收 prompt。
home-page-prompt.md: 首頁封面、公開資訊、服務入口及我的大廈 UI prompt。
login-page-prompt.md: 登入頁視覺、表單與狀態 UI prompt。
marketplace-discover-page-prompt.md: 二手展示探索頁 UI prompt。
marketplace-filter-page-prompt.md: 二手篩選列表頁 UI prompt。
marketplace-publish-page-prompt.md: 二手發布頁表單、預覽與視覺 prompt。

## 架構決策
- 本目錄只保存可直接給 AI 或 agent 使用的 UI prompt，不保存產品驗收、API 契約或部署流程。
- 新增 prompt 前先確認是否能擴充既有文件；只有可獨立重用的頁面、表單或視覺規則才新增文件。
- 文件正文使用繁體中文或 English；UI 文案保持正式、直接，不使用 emoji，不加入給實作者看的過程文字。
- prompt 涉及具體頁面時，必須同步確認 `web_new/src/pages/AGENTS.md` 和 `docs/PROMPT_INDEX.md` 是否需要更新。
- prompt 描述產品邊界時，以根級 `AGENTS.md` 的 AJO Living 主系統定位為準，不把所有功能寫成 marketplace。

## 變更日誌
2026-11-23: 首頁超市優惠規格改為前四項，桌面四欄、手機兩欄及四張載入骨架。
2026-11-23: 首頁 prompt 移除整條頂部工具列，不保留屋苑選擇與通知入口，封面直接接續全站導航。
2026-11-23: 首頁 prompt 固定與超市列表共用商品卡、門店比價及收藏，手機兩列，頂部只保留屋苑與通知。
2026-11-20: 首頁交通與生活提示查詢連結固定直達香港城市儀表板對應欄目，摘要仍由 AJO 後端讀取。
2026-11-20: 首頁規格加入屋苑參考區、環境、港鐵、生活提示、物業資訊及真實大廈服務。
2026-11-19: 首頁 prompt 改為既有功能入口及真實優惠、樓盤，移除全螢幕輪播、捲動舞台與未接通功能。
2026-07-08: 建立 UI prompt 目錄記憶，限制新增 prompt 的用途與維護入口。

[PROTOCOL]: When adding, moving, renaming, or changing a reusable UI prompt here, update this map, `../PROMPT_INDEX.md`, and check related `web_new/src/pages/AGENTS.md` ownership.
