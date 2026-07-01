# Tasks

- [x] Task 1: 修改 AppHeader 新增 AJO Pay 上下文切换
  - [x] SubTask 1.1: 在 `AppHeader.vue` 新增 nav-pay-ctx 区块，当路由匹配 `/payments` 时显示 AJO PAY 品牌 + 4 个 tab（首頁、付款、AJO Coin 回贈、返回 AJO Living），隐藏主 nav-links
  - [x] SubTask 1.2: 点击 tab 切换 AJO Pay 子页面（通过路由 query 参数 tab），点击返回 AJO Living 跳转 `/`

- [x] Task 2: 重写 AJO Pay 页面
  - [x] SubTask 2.1: 重写 `payments/Page.vue`，还原 3 个子页面（HOME/PAYMENT/COIN），通过 tab 切换
  - [x] SubTask 2.2: HOME 子页面：d-hero（渐变背景 + badge + 标题 + 描述 + 按钮）+ d-hero-stats（3 个统计卡片）+ d-stats-row（3 个统计项）+ d-features（3 个功能卡片）+ d-footer
  - [x] SubTask 2.3: PAYMENT 子页面：breadcrumb + d-payment-steps（4 步骤）+ d-pay-layout（3 个 step-section：选择付款项目、选择付款方式、确认付款）+ 付款紀錄
  - [x] SubTask 2.4: COIN 子页面：breadcrumb + d-coin-hero（渐变背景 + 标题 + Coin 結餘卡片）+ d-coin-stats（3 个统计）+ d-coin-steps（3 个步骤）+ d-redeem-grid（4 个兑换卡片）+ d-footer

- [x] Task 3: 重写樓盤租售列表增强功能
  - [x] SubTask 3.1: 在搜索框新增自动补全下拉（autocomplete-wrap + ac-item）
  - [x] SubTask 3.2: 新增地图视图（map-view + map-canvas + map-pin + map-popup），支持列表/地图切换
  - [x] SubTask 3.3: 卡片新增 listing-price-kind（售/租标识，售=橙色，租=蓝色）、g-location（位置）、listing-agent（代理公司 SVG + 文字）、garea-price（呎價）、compare-check（比较 checkbox）
  - [x] SubTask 3.4: 更新 4 个卡片的 mock 数据匹配新结构

- [x] Task 4: 验证
  - [x] SubTask 4.1: 运行 `npm run typecheck` 确保无类型错误

# Task Dependencies
- Task 2 依赖 Task 1（AJO Pay 页面需要导航栏上下文切换）
- Task 3 独立
- Task 4 依赖 Task 1-3
