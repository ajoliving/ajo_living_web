# Tasks

- [x] Task 1: 移除路由守卫登录检查
  - [x] SubTask 1.1: 修改 `web-new/src/router/index.ts` 的 `beforeEach`，移除 `requiresAuth` 与 `requiresStaff` 检查逻辑，直接返回 true，保留标题同步逻辑
  - [x] SubTask 1.2: 运行 `npm run typecheck` 确保无类型错误

# Task Dependencies
- 无依赖，单文件修改
