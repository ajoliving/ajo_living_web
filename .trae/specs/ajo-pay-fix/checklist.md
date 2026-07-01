# Checklist

## sessionStore 未登录预览
- [x] `isAuthenticated` 始终返回 true
- [x] `loadCurrentUser` 未登录时不报错，返回 mock 用户数据
- [x] 未登录访问 `/account/profile`、`/building` 等页面正常渲染

## AJO Pay 导航栏
- [x] AJO LIVING logo 始终显示，带橙色下划线
- [x] `.ajo-nav__pay-ctx` 样式与 HTML 设计稿匹配
- [x] `.ajo-nav__pay-tab` 样式与 HTML 设计稿匹配
- [x] `.ajo-nav__pay-exit` 样式与 HTML 设计稿匹配

## AJO Pay 页面
- [x] `.d-hero` 渐变背景与 HTML 设计稿匹配
- [x] `.d-pay-page` 间距与 HTML 设计稿匹配
- [x] `.d-coin-hero` 渐变背景与 HTML 设计稿匹配

## 验证
- [x] `npm run typecheck` 通过
