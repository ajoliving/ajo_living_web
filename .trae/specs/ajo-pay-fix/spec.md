# AJO Pay 还原修复 Spec

## Why
HTML v2 还原后存在 3 个问题：AJO Pay 导航栏样式与设计稿不匹配导致用户感觉"离开了 AJO Living"；页面内部仍有登录重定向逻辑导致我的大厦、会员中心等页面无法预览；AJO Pay 页面整体样式还原度不足。

## What Changes
- 修复 `AppHeader.vue` 的 AJO Pay 上下文栏 CSS：对齐 HTML 设计稿的 `.nav-pay-ctx`、`.npt`、`.pay-exit` 样式，确保 logo 始终显示带橙色下划线
- 修复 `payments/Page.vue` 的 CSS：对齐 HTML 设计稿的 AJO Pay 专属样式（`.d-hero`、`.d-pay-page`、`.d-coin-hero` 等），确保渐变背景、间距、字号匹配
- 修改 `sessionStore` 的 `loadCurrentUser`：高保真还原阶段未登录时返回 mock 用户数据，不报错
- 修改 `sessionStore` 的 `isAuthenticated`：高保真还原阶段始终返回 true

## Impact
- Affected specs: html-fidelity-v2, remove-auth-guard
- Affected code:
  - `web-new/src/shared/components/navigation/AppHeader.vue`
  - `web-new/src/pages/payments/Page.vue`
  - `web-new/src/stores/session.ts`

## ADDED Requirements
### Requirement: 高保真还原阶段无登录预览
The system SHALL allow all pages to render without authentication by providing mock user data when not logged in.

#### Scenario: 未登录访问会员中心
- **WHEN** 未登录用户访问 `/account/profile`
- **THEN** 页面正常渲染，不报错，不跳转登录页

## MODIFIED Requirements
### Requirement: AJO Pay 导航栏
导航栏在 AJO Pay 路由下保留 AJO LIVING logo（带橙色下划线），显示 AJO PAY 品牌 + tab 切换，样式与 HTML 设计稿完全匹配。
