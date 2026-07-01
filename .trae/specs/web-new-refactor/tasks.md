# Tasks

- [ ] Task 1: 提取 HTML 全局 CSS 变量与共用样式
  - [ ] SubTask 1.1: 重写 `web-new/src/styles/tokens.css`，严格对齐 HTML 的 `:root` 变量（`--brand`、`--ink`、`--sur`、`--bdr`、`--sp-*`、`--r-*`、`--text-*`、`--shadow-*`、`--nav-h`、`--font`、`--font-serif`）
  - [ ] SubTask 1.2: 重写 `web-new/src/styles/index.css`，添加 HTML 的 `body`、`.btn-base`、`.btn-p`、`.breadcrumb`、`.bc-link`、`.bc-sep`、`.bc-current`、深色模式 `body.dark` 等共用样式

- [ ] Task 2: 重写导航栏 AppHeader.vue
  - [ ] SubTask 2.1: 严格对齐 HTML 的 `.nav`（52px、sticky、box-shadow）、`.nav-logo`（serif 字体、橙色下划线）、`.nav-links`、`.nl`（11 个导航项）、`.nav-r`、`.nav-login`、`.theme-toggle`（月亮/太阳 SVG）
  - [ ] SubTask 2.2: 严格对齐 HTML 的 `.nav-pay-ctx`、`.nav-pay-brand`、`.nav-pay-divider`、`.nav-pay-tabs`、`.npt`、`.pay-exit`（AJO Pay 上下文栏）

- [ ] Task 3: 重写首页 HomePage.vue
  - [ ] SubTask 3.1: 严格对齐 HTML 的 `#page-home` 结构与样式（HERO + 探索服務 + 精選樓盤）

- [ ] Task 4: 重写樓盤租售列表 PropertyListPage.vue
  - [ ] SubTask 4.1: 严格对齐 HTML 的 `#page-listing` 结构与样式（三栏布局 + 筛选 + 列表 + 广告 + 地图视图）

- [ ] Task 5: 重写樓盤詳情 PropertyDetailPage.vue
  - [ ] SubTask 5.1: 严格对齐 HTML 的 `#page-detail` 结构与样式

- [ ] Task 6: 重写服務式住宅列表 ServicedResidenceListPage.vue
  - [ ] SubTask 6.1: 严格对齐 HTML 的 `#page-service` 结构与样式

- [ ] Task 7: 重写服務式住宅詳情 ServicedResidenceDetailPage.vue
  - [ ] SubTask 7.1: 严格对齐 HTML 的 `#page-service-detail` 结构与样式

- [ ] Task 8: 重写家具市集列表 furniture/Page.vue
  - [ ] SubTask 8.1: 严格对齐 HTML 的 `#page-market` 结构与样式

- [ ] Task 9: 重写家具市集詳情 FurnitureDetailPage.vue
  - [ ] SubTask 9.1: 严格对齐 HTML 的 `#page-market-detail` 结构与样式

- [ ] Task 10: 重写綜合優惠 offers/Page.vue
  - [ ] SubTask 10.1: 严格对齐 HTML 的 `#page-offers` 结构与样式

- [ ] Task 11: 重写 AJO Pay payments/Page.vue
  - [ ] SubTask 11.1: 严格对齐 HTML 的 `#page-payment` 结构与样式（3 子页面 HOME/PAYMENT/COIN）

- [ ] Task 12: 重写我的大廈 BuildingPage.vue
  - [ ] SubTask 12.1: 严格对齐 HTML 的 `#page-affairs` 结构与样式

- [ ] Task 13: 重写會員中心 AccountMyPage.vue
  - [ ] SubTask 13.1: 严格对齐 HTML 的 `#page-profile` 结构与样式

- [ ] Task 14: 重写管理端 management/Page.vue
  - [ ] SubTask 14.1: 严格对齐 HTML 的 `#page-management` 结构与样式

- [ ] Task 15: 重写走勢页 TrendPage.vue
  - [ ] SubTask 15.1: 严格对齐 HTML 的 `#page-trend` 结构与样式

- [ ] Task 16: 重写通知中心 notifications/Page.vue
  - [ ] SubTask 16.1: 严格对齐 HTML 的 `#page-notif` 结构与样式

- [ ] Task 17: 重写登入页 LoginPage.vue
  - [ ] SubTask 17.1: 严格对齐 HTML 的 `#page-login` 结构与样式

- [ ] Task 18: 重写收藏页 MarketplaceMyFavoritesPage.vue
  - [ ] SubTask 18.1: 严格对齐 HTML 的 `#page-saved` 结构与样式

- [ ] Task 19: 验证
  - [ ] SubTask 19.1: 运行 `npm run typecheck`

# Task Dependencies
- Task 2 依赖 Task 1
- Task 3-18 依赖 Task 1、2
- Task 19 依赖 Task 1-18
