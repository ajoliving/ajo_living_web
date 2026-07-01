# Checklist

## 路由守卫
- [x] `web-new/src/router/index.ts` 的 `beforeEach` 不再检查 `requiresAuth` 与 `requiresStaff`
- [x] 未登录访问 `/payments`、`/account/profile`、`/marketplace/management` 等页面不跳转登录
- [x] `npm run typecheck` 通过
