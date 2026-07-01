# Tasks

- [x] Task 1: 修复 sessionStore 未登录预览
  - [x] SubTask 1.1: 修改 `web-new/src/stores/session.ts` 的 `isAuthenticated` getter，高保真还原阶段始终返回 true
  - [x] SubTask 1.2: 修改 `loadCurrentUser` 方法，未登录时返回 mock 用户数据，不报错

- [x] Task 2: 修复 AJO Pay 导航栏样式
  - [x] SubTask 2.1: 修复 `AppHeader.vue` 的 `.ajo-nav__logo` 样式，添加橙色下划线（border-bottom: 2px solid var(--color-primary)）
  - [x] SubTask 2.2: 修复 `.ajo-nav__pay-ctx`、`.ajo-nav__pay-tab`、`.ajo-nav__pay-exit` 样式，对齐 HTML 设计稿

- [x] Task 3: 修复 AJO Pay 页面样式
  - [x] SubTask 3.1: 修复 `payments/Page.vue` 的 CSS，对齐 HTML 设计稿的 AJO Pay 专属样式

- [x] Task 4: 验证
  - [x] SubTask 4.1: 运行 `npm run typecheck`

# Task Dependencies
- Task 1 独立
- Task 2 独立
- Task 3 独立
- Task 4 依赖 Task 1-3
