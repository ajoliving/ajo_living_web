# web-new 严格对齐 HTML 重构 Spec

## Why
现有 web-new 还原度不足，用户要求按照 HTML 设计稿 `ajo_living_desktop_20260624(3)(10).html` 严格对齐，直接重构 web-new 目录。

## What Changes
- 提取 HTML 全局 CSS 变量与共用样式到 `styles/tokens.css` 和 `styles/index.css`
- 重写导航栏 `AppHeader.vue`，严格对齐 HTML 的 `.nav`、`.nav-logo`、`.nav-links`、`.nl`、`.nav-r`、`.nav-login`、`.theme-toggle`、`.nav-pay-ctx`、`.npt` 样式
- 重写 16 个页面，严格对齐 HTML 结构与样式：
  1. page-home（首页）
  2. page-listing（樓盤租售列表）
  3. page-detail（樓盤詳情）
  4. page-service（服務式住宅列表）
  5. page-service-detail（服務式住宅詳情）
  6. page-market（家具市集列表）
  7. page-market-detail（家具市集詳情）
  8. page-offers（綜合優惠）
  9. page-payment（AJO Pay）
  10. page-affairs（我的大廈）
  11. page-profile（會員中心）
  12. page-management（管理端）
  13. page-trend（走勢）
  14. page-notif（通知中心）
  15. page-login（登入页）
  16. page-saved（收藏页）

## Impact
- Affected specs: html-fidelity-restore, html-fidelity-v2, ajo-pay-fix
- Affected code: `web-new/src/styles/`、`web-new/src/shared/components/navigation/AppHeader.vue`、`web-new/src/pages/` 下所有页面

## ADDED Requirements
### Requirement: 严格对齐 HTML
The system SHALL strictly align with HTML design draft structure and styles, including CSS variables, class names, layout, spacing, colors, fonts.

#### Scenario: 页面渲染
- **WHEN** user visits any page
- **THEN** the page structure and styles match HTML design draft exactly
