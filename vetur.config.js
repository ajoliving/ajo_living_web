/*
 * Vetur 工作區設定檔。
 * 1. 將 Vue 前端子專案根目錄指定為 web。
 * 2. 讓 Vetur 從 web/package.json 讀取依賴資訊。
 * 3. 讓 Vetur 使用 web/jsconfig.json 解析路徑別名。
 */

module.exports = {
  projects: [
    {
      root: './web',
      package: './package.json',
      tsconfig: './jsconfig.json'
    }
  ]
};
