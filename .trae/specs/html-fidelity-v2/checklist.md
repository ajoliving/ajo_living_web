# Checklist

## 导航栏
- [x] AppHeader 在 `/payments` 路由下显示 nav-pay-ctx（AJO PAY 品牌 + 4 tab）
- [x] 点击返回 AJO Living 跳转首页
- [x] 非 payments 路由保持原有 11 项导航

## AJO Pay 页面
- [x] HOME 子页面：d-hero 渐变背景 + 统计卡片 + 功能卡片 + footer
- [x] PAYMENT 子页面：4 步骤指示器 + 付款项目选择 + 付款方式选择 + 确认付款 + 付款紀錄
- [x] COIN 子页面：d-coin-hero 渐变背景 + Coin 結餘 + 统计 + 步骤 + 兑换卡片
- [x] tab 切换有 fade-in 动画

## 樓盤租售列表
- [x] 搜索框有自动补全下拉
- [x] 支持列表/地图视图切换
- [x] 地图视图有 price pins 和 popup
- [x] 卡片有售/租价格标识（橙色售、蓝色租）
- [x] 卡片有位置（g-location）
- [x] 代理盘卡片有代理公司信息（listing-agent）
- [x] 卡片有呎價（garea-price）
- [x] 卡片有比较 checkbox

## 验证
- [x] `npm run typecheck` 通过
