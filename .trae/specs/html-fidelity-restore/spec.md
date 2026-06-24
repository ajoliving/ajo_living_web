# HTML 高保真还原为 web-new Vue 项目 Spec

## Why
`docs/ajo_living_desktop_20260611(3).html` 是产品视觉基准（5526 行，16 个页面，含亮/暗双模式），现有 `web-new` Vue 项目虽已建立路由、主题体系与多个页面，但顶部导航、页面布局、卡片样式、配色细节与设计稿存在明显落差。需要将 HTML 的布局、UI、主题切换高保真还原到 Vue 项目，作为后续功能开发与对接后端的视觉基线。

## Scope
- 仅做视觉高保真还原：布局结构、组件样式、配色、字体、间距、交互态、亮/暗双模式。
- 不实现业务功能：API 调用、表单提交、登录鉴权、数据持久化等保留现有桩逻辑或使用静态 mock 数据。
- 保留现有路由结构与 `web-new/AGENTS.md` 的页面单例目录规范。
- 保留现有 4 套主题机制（default、html-fidelity、copper-sun、dark-neutral），但 UI 上以 HTML 设计稿的亮/暗切换按钮为还原目标，`html-fidelity` 作为亮色基准，`dark-neutral` 作为深色基准。
- 不引入新依赖，不引入 Ant Design Vue（HTML 设计稿未使用）。
- 不创建新页面目录，仅替换现有页面入口与 widgets 的模板与样式。

## What Changes

### 基础架构
- 对齐设计 token：将 HTML 中 `--brand`、`--ink`、`--sur`、`--bdr` 等变量映射到现有 `--color-primary`、`--color-text`、`--color-surface`、`--color-border` 体系；补齐 `--nav-h: 52px`、`--brand-mid`、`--brand-dark`、`--ink-2/3/4`、`--sur-2/3`、`--bdr-2`、`--success`、`--warning`、`--error` 等 HTML 独有 token。
- 重写 `AppHeader.vue`：还原 HTML 的 52px 高度、左侧 logo、中间导航按钮（首頁、樓盤租售、服務式住宅、家具、綜合優惠、AJO Pay、我的大廈、會員中心、管理、走勢、通知铃铛）、右侧登入按钮、主题切换按钮（月亮/太阳图标）。
- 重写 `AppShell.vue`：调整 `main` 的 `padding-top` 为 52px，还原 HTML 的页面切换淡入。
- 调整主题切换 UI：将现有下拉选择改为 HTML 设计稿的图标按钮（月亮/太阳），保留 `preferenceStore.setTheme` 调用，亮色对应 `html-fidelity`，深色对应 `dark-neutral`。
- 补齐深色模式样式：在 `styles/index.css` 中按 HTML 的 `body.dark` 规则补齐 `html[data-theme='dark-neutral']` 下的表面、文字、边框覆盖。

### 页面还原（16 个）
按 HTML 设计稿还原以下页面的布局与样式，使用静态 mock 数据：

1. **首页** `pages/home/HomePage.vue` ← `#page-home`
   - HERO 暗色背景 + SVG 城市剪影 + 标题/描述/双按钮
   - 探索服務 4 卡片网格
   - 精選樓盤 3 卡片网格
2. **樓盤租售列表** `pages/property/list/PropertyListPage.vue` ← `#page-listing`
   - 三栏布局：左侧筛选栏（地区、性质、售价、面积、房间、装修）+ 中间列表（排序栏 + 卡片网格 + 分页）+ 右侧广告栏
3. **樓盤詳情** `pages/property/detail/PropertyDetailPage.vue` ← `#page-detail`
   - 面包屑 + 图片画廊 + 标题/价格/标签 + 详情 tabs + 推荐卡片
4. **服務式住宅列表** 新建 `pages/serviced-residence/list/` ← `#page-service`
   - HERO + 筛选栏 + 卡片网格
5. **服務式住宅詳情** 新建 `pages/serviced-residence/detail/` ← `#page-service-detail`
   - 图片画廊 + 详情 + 推荐卡片
6. **家具市集列表** `pages/furniture/Page.vue` ← `#page-market`
   - 三栏布局：筛选栏 + 列表 + 广告栏
7. **家具市集詳情** 新建 `pages/furniture/detail/` ← `#page-market-detail`
   - 图片 + 详情 + 卖家信息 + 推荐
8. **綜合優惠** `pages/offers/Page.vue` ← `#page-offers`
   - 暗色 HERO + 控制栏 + 4 列卡片网格 + 分页
9. **AJO Pay** `pages/payments/Page.vue` ← `#page-payment`
   - 顶部子导航 + 概览/賬單/歷史/購物車/訂單 tabs + 内容区
10. **我的大廈** `pages/building/BuildingPage.vue` ← `#page-affairs`
    - 左侧大廈列表 + 右侧通告/設施/意見 tabs
11. **會員中心** `pages/account/my/AccountMyPage.vue` ← `#page-profile`
    - 左侧导航 + 右侧个人资料/錢包/物業/收藏 tabs
12. **管理端** `pages/marketplace/management/Page.vue` ← `#page-management`
    - 左侧子导航 + 右侧表格/卡片管理界面
13. **走勢** 新建 `pages/trend/` ← `#page-trend`
    - 图表 + 数据表
14. **通知中心** `pages/notifications/Page.vue` ← `#page-notif`
    - 通知列表 + 分类 tabs
15. **登入** `pages/account/login/LoginPage.vue` ← `#page-login`
    - 双栏布局：左侧品牌图 + 右侧登录表单
16. **收藏** `pages/marketplace/my/favorites/MarketplaceMyFavoritesPage.vue` ← `#page-saved`
    - 收藏卡片网格

### 共用组件
- `ListingCard.vue`：还原 HTML 的 `.gc`/`.mc` 卡片样式（图片 + 标签 + 标题 + 价格 + 悬停操作按钮）
- `AdCard.vue`：新建，还原 HTML 的 `.ad-card` 广告卡片（banner/vertical 两种）
- `PaginationBar.vue`：新建，还原 HTML 的分页栏
- `FilterTag.vue`：新建，还原 HTML 的 `.ft` 筛选标签
- `Breadcrumb.vue`：新建，还原 HTML 的 `.breadcrumb`

## Impact
- Affected specs: 无（首次 spec）
- Affected code:
  - `web-new/src/styles/tokens.css`、`web-new/src/styles/index.css`
  - `web-new/src/utils/theme.ts`、`web-new/src/stores/preferences.ts`
  - `web-new/src/shared/components/layout/AppShell.vue`
  - `web-new/src/shared/components/navigation/AppHeader.vue`
  - `web-new/src/shared/components/marketplace/ListingCard.vue`
  - `web-new/src/pages/home/HomePage.vue` 及其 widgets
  - `web-new/src/pages/property/list/PropertyListPage.vue`
  - `web-new/src/pages/property/detail/PropertyDetailPage.vue`
  - `web-new/src/pages/furniture/Page.vue`
  - `web-new/src/pages/offers/Page.vue`
  - `web-new/src/pages/payments/Page.vue`
  - `web-new/src/pages/building/BuildingPage.vue`
  - `web-new/src/pages/account/my/AccountMyPage.vue`
  - `web-new/src/pages/marketplace/management/Page.vue`
  - `web-new/src/pages/notifications/Page.vue`
  - `web-new/src/pages/account/login/LoginPage.vue`
  - `web-new/src/pages/marketplace/my/favorites/MarketplaceMyFavoritesPage.vue`
  - 新建 `web-new/src/pages/serviced-residence/list/`、`web-new/src/pages/serviced-residence/detail/`、`web-new/src/pages/furniture/detail/`、`web-new/src/pages/trend/`
  - `web-new/src/router/routes/` 下相关路由文件

## ADDED Requirements

### Requirement: 设计 Token 对齐
系统 SHALL 将 HTML 设计稿的 CSS 变量映射到 web-new 的 token 体系，确保亮/暗模式下颜色、字体、间距、圆角、阴影与设计稿一致。

#### Scenario: 亮色模式
- WHEN 用户选择亮色主题（`html-fidelity`）
- THEN `--color-primary` 为 `240 90 0`、`--color-text` 为 `26 26 26`、`--color-surface` 为 `255 255 255`、`--color-canvas` 为 `255 255 255`、`--nav-h` 为 `52px`

#### Scenario: 深色模式
- WHEN 用户选择深色主题（`dark-neutral`）
- THEN 表面色为深色（`#0B1120`/`#111827` 系）、文字色为浅色（`#F8FAFC`/`#E5E7EB` 系）、边框为 `#263244` 系、主色保持 `#F05A00`

### Requirement: 顶部导航高保真
系统 SHALL 还原 HTML 设计稿的顶部导航栏，高度 52px，包含 logo、11 个导航项、主题切换按钮、登入按钮。

#### Scenario: 桌面端
- WHEN 视口宽度 ≥ 1024px
- THEN 显示完整导航项、主题切换按钮、登入按钮

#### Scenario: 移动端
- WHEN 视口宽度 < 1024px
- THEN 隐藏中间导航与右侧按钮，显示汉堡菜单按钮与底部导航栏

### Requirement: 主题切换按钮
系统 SHALL 提供月亮/太阳图标按钮切换亮/暗主题，亮色显示月亮图标，深色显示太阳图标。

#### Scenario: 切换主题
- WHEN 用户点击主题切换按钮
- THEN 主题在 `html-fidelity`（亮）与 `dark-neutral`（暗）之间切换，按钮图标相应变化，显示 toast 提示

### Requirement: 页面布局高保真
系统 SHALL 还原 HTML 设计稿中每个页面的布局结构，包括栅格列数、间距、卡片样式、悬停态、分页、面包屑等。

#### Scenario: 列表页三栏布局
- WHEN 用户访问樓盤租售列表或家具市集列表
- THEN 页面显示左侧筛选栏（210px）+ 中间列表（自适应）+ 右侧广告栏（300px）的三栏布局

#### Scenario: 卡片悬停态
- WHEN 用户悬停在列表卡片上
- THEN 卡片显示底部操作按钮（收藏、比較、查看）并向上浮动

## MODIFIED Requirements

### Requirement: 现有页面视觉替换
现有 `web-new` 项目的首页、列表页、详情页等页面 SHALL 替换为 HTML 设计稿的视觉布局，保留现有路由路径与组件入口文件名。

## REMOVED Requirements
无。
