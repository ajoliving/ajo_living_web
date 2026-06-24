# Tasks

## 阶段一：基础架构

- [x] Task 1: 对齐设计 token
  - [x] SubTask 1.1: 在 `web-new/src/styles/tokens.css` 中补齐 HTML 设计稿独有 token：`--nav-h: 52px`、`--brand-mid`、`--brand-dark`、`--ink-2/3/4`、`--sur-2/3`、`--bdr-2`、`--success`、`--warning`、`--error` 及对应 bg 变体
  - [x] SubTask 1.2: 在 `web-new/src/utils/theme.ts` 的 `html-fidelity` 与 `dark-neutral` preset 中补齐对应 token，确保 `buildThemeCssVariables` 输出新 token
  - [x] SubTask 1.3: 在 `web-new/src/styles/index.css` 中按 HTML 的 `body.dark` 规则补齐 `html[data-theme='dark-neutral']` 下的表面、文字、边框、卡片、导航、页脚覆盖样式

- [x] Task 2: 重写顶部导航 AppHeader
  - [x] SubTask 2.1: 调整 `web-new/src/shared/components/navigation/AppHeader.vue` 高度为 52px，还原 HTML 的 logo（`AJO LIVING` + serif 字体）、11 个导航按钮（首頁、樓盤租售、服務式住宅、家具、綜合優惠、AJO Pay、我的大廈、會員中心、管理、走勢、通知铃铛）
  - [x] SubTask 2.2: 还原右侧登入按钮（橙色填充）与主题切换按钮（月亮/太阳 SVG 图标），点击切换 `html-fidelity` ↔ `dark-neutral`
  - [x] SubTask 2.3: 还原移动端汉堡菜单与底部导航栏样式
  - [x] SubTask 2.4: 调整 `AppShell.vue` 的 `main` padding-top 为 52px

- [x] Task 3: 新建共用组件
  - [x] SubTask 3.1: 新建 `web-new/src/shared/components/marketplace/AdCard.vue`，还原 HTML `.ad-card`（banner/vertical 两种，含 `.ad-visual` 渐变背景、`.ad-copy` 文案层）
  - [x] SubTask 3.2: 新建 `web-new/src/shared/components/navigation/PaginationBar.vue`，还原 HTML `.listing-pagination` 样式
  - [x] SubTask 3.3: 新建 `web-new/src/shared/components/base/FilterTag.vue`，还原 HTML `.ft` 筛选标签（含 `.on` 选中态）
  - [x] SubTask 3.4: 新建 `web-new/src/shared/components/navigation/AppBreadcrumb.vue`，还原 HTML `.breadcrumb` 样式

## 阶段二：核心页面还原

- [x] Task 4: 还原首页 `pages/home/HomePage.vue`
  - [x] SubTask 4.1: 还原 HERO 区：暗色渐变背景 + SVG 城市剪影 + `AJO LIVING` eyebrow + `理想生活\n由此出發` 标题 + 描述 + 双按钮（搜尋樓盤/了解服務）
  - [x] SubTask 4.2: 还原「探索服務」区：4 个 `.cat-card`（樓盤租售、服務式住宅、家具市集、綜合優惠）
  - [x] SubTask 4.3: 还原「精選樓盤」区：3 个 `.feat-card`（图片 + 标签 + 标题 + 价格 + 面积）
  - [x] SubTask 4.4: 移除现有 `HomeStageShowcase`/`HomeScrollGrid`/`HomeModuleSection` 的 API 调用，改用静态 mock 数据

- [x] Task 5: 还原樓盤租售列表 `pages/property/list/PropertyListPage.vue`
  - [x] SubTask 5.1: 还原三栏布局 `.lp`（210px + 1fr + 300px）
  - [x] SubTask 5.2: 还原左侧筛选栏 `.lf`：搜索框 + 6 组筛选标签（地区、性质、售价、面积、房间、装修）
  - [x] SubTask 5.3: 还原中间列表区 `.lr`：排序栏 + 4 个 `.gc` 卡片（图片 + 标签 + 标题 + 价格 + 悬停操作按钮）+ 分页
  - [x] SubTask 5.4: 还原右侧广告栏 `.listing-ad-aside`：2 个 `.ad-card.vertical`

- [x] Task 6: 还原樓盤詳情 `pages/property/detail/PropertyDetailPage.vue`
  - [x] SubTask 6.1: 还原面包屑 + 图片画廊（主图 + 缩略图）+ 标题/价格/标签
  - [x] SubTask 6.2: 还原详情 tabs（詳細資料、平面圖、地圖、周邊）+ 推荐卡片

- [x] Task 7: 还原服務式住宅列表（新建 `pages/serviced-residence/list/`）
  - [x] SubTask 7.1: 新建路由 `web-new/src/router/routes/serviced-residence.ts`，path `/serviced-residences`
  - [x] SubTask 7.2: 还原 `.sv-hero` 暗色 HERO + 筛选栏 + `.sv-card` 卡片网格
  - [x] SubTask 7.3: 在 `router/index.ts` 注册路由

- [x] Task 8: 还原服務式住宅詳情（新建 `pages/serviced-residence/detail/`）
  - [x] SubTask 8.1: 还原图片画廊 + 详情 + 推荐卡片

- [x] Task 9: 还原家具市集列表 `pages/furniture/Page.vue`
  - [x] SubTask 9.1: 还原三栏布局 `.mp`（210px + 1fr + 300px）
  - [x] SubTask 9.2: 还原左侧筛选栏 + 中间 `.mc` 卡片网格 + 右侧广告栏

- [x] Task 10: 还原家具市集詳情（新建 `pages/furniture/detail/`）
  - [x] SubTask 10.1: 新建路由，还原图片 + 详情 + 卖家信息 + 推荐

## 阶段三：业务页面还原

- [x] Task 11: 还原綜合優惠 `pages/offers/Page.vue`
  - [x] SubTask 11.1: 还原 `.gp-hero` 暗色 HERO（标题 + 右侧统计）+ `.gp-controls` 控制栏（搜索 + 筛选 pills + 排序）+ `.gp-grid` 4 列卡片网格 + 分页

- [x] Task 12: 还原 AJO Pay `pages/payments/Page.vue`
  - [x] SubTask 12.1: 还原顶部子导航（概覽、賬單、歷史、購物車、訂單、會計、單位）+ 内容区概覣卡片

- [x] Task 13: 还原我的大廈 `pages/building/BuildingPage.vue`
  - [x] SubTask 13.1: 还原左侧大廈列表 + 右侧通告/設施/意見 tabs + 通告列表卡片

- [x] Task 14: 还原會員中心 `pages/account/my/AccountMyPage.vue`
  - [x] SubTask 14.1: 还原左侧导航（個人資料、錢包、物業、收藏、通知）+ 右侧个人资料卡片

- [x] Task 15: 还原管理端 `pages/marketplace/management/Page.vue`
  - [x] SubTask 15.1: 还原左侧子导航 + 右侧表格管理界面（二手列表、樓盤放售、服務住宅、獎勵廣告、錢包等）

- [x] Task 16: 还原走勢页（新建 `pages/trend/`）
  - [x] SubTask 16.1: 新建路由 `web-new/src/router/routes/trend.ts`，path `/trend`
  - [x] SubTask 16.2: 还原图表区 + 数据表
  - [x] SubTask 16.3: 在 `router/index.ts` 注册路由

- [x] Task 17: 还原通知中心 `pages/notifications/Page.vue`
  - [x] SubTask 17.1: 还原分类 tabs + 通知列表卡片

- [x] Task 18: 还原登入页 `pages/account/login/LoginPage.vue`
  - [x] SubTask 18.1: 还原双栏布局：左侧品牌图（暗色渐变 + logo）+ 右侧登录表单（手机号 + OTP + 登入按钮）

- [x] Task 19: 还原收藏页 `pages/marketplace/my/favorites/MarketplaceMyFavoritesPage.vue`
  - [x] SubTask 19.1: 还原收藏卡片网格

## 阶段四：验证

- [x] Task 20: 视觉验证
  - [x] SubTask 20.1: 运行 `npm run typecheck` 确保无类型错误
  - [x] SubTask 20.2: 运行 `npm run dev` 逐页对比 HTML 设计稿与 Vue 渲染结果，确保布局、配色、字体、间距、悬停态、亮/暗模式一致

# Task Dependencies
- Task 2 依赖 Task 1（导航栏需要新 token）
- Task 3 依赖 Task 1（共用组件需要新 token）
- Task 4-19 依赖 Task 1、2、3（页面需要基础架构与共用组件）
- Task 4-19 之间可并行（页面相互独立）
- Task 20 依赖 Task 1-19 全部完成
