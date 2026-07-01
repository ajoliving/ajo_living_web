# 移除登录守卫 Spec

## Why
HTML 高保真还原阶段需要逐页对比视觉效果，现有路由守卫会拦截未登录用户访问 payments、account、marketplace management 等页面，导致无法直接预览。需要临时移除登录限制，让所有页面不管是否登录都能访问。

## What Changes
- 修改 `web-new/src/router/index.ts` 的 `beforeEach` 守卫，移除 `requiresAuth` 与 `requiresStaff` 检查逻辑，直接放行所有路由
- 保留守卫函数骨架与注释，便于后续恢复登录逻辑

## Impact
- Affected specs: html-fidelity-restore
- Affected code: `web-new/src/router/index.ts`

## ADDED Requirements
### Requirement: 无登录预览模式
The system SHALL allow all routes to be accessible without authentication during the fidelity restore phase.

#### Scenario: 未登录访问受保护页面
- **WHEN** 未登录用户访问 `/payments`、`/account/profile`、`/marketplace/management` 等页面
- **THEN** 页面正常渲染，不跳转到登录页

## MODIFIED Requirements
### Requirement: 路由守卫
路由守卫暂时禁用 `requiresAuth` 与 `requiresStaff` 检查，仅保留标题同步逻辑。守卫函数骨架保留，注释标记为临时禁用。
