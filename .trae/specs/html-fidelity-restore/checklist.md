# Checklist

## 基础架构
- [x] `web-new/src/styles/tokens.css` 已补齐 `--nav-h`、`--brand-mid`、`--brand-dark`、`--ink-2/3/4`、`--sur-2/3`、`--bdr-2`、`--success`、`--warning`、`--error` 等 token
- [x] `web-new/src/utils/theme.ts` 的 `html-fidelity` 与 `dark-neutral` preset 已补齐新 token
- [x] `web-new/src/styles/index.css` 已补齐 `html[data-theme='dark-neutral']` 下的深色模式覆盖样式
- [x] `web-new/src/shared/components/navigation/AppHeader.vue` 高度为 52px，包含 11 个导航项、主题切换按钮、登入按钮
- [x] 主题切换按钮显示月亮/太阳 SVG 图标，点击切换 `html-fidelity` ↔ `dark-neutral`
- [x] `web-new/src/shared/components/layout/AppShell.vue` 的 `main` padding-top 为 52px
- [x] 移动端汉堡菜单与底部导航栏样式已还原

## 共用组件
- [x] `AdCard.vue` 支持 banner 与 vertical 两种变体，含渐变背景与文案层
- [x] `PaginationBar.vue` 还原 HTML `.listing-pagination` 样式
- [x] `FilterTag.vue` 还原 HTML `.ft` 样式与 `.on` 选中态
- [x] `AppBreadcrumb.vue` 还原 HTML `.breadcrumb` 样式

## 页面还原
- [x] 首页 HERO 暗色背景 + SVG 城市剪影 + 标题/描述/双按钮已还原
- [x] 首页「探索服務」4 卡片网格已还原
- [x] 首页「精選樓盤」3 卡片网格已还原
- [x] 樓盤租售列表三栏布局（210px + 1fr + 300px）已还原
- [x] 樓盤租售列表左侧筛选栏 6 组筛选标签已还原
- [x] 樓盤租售列表中间卡片网格含悬停操作按钮已还原
- [x] 樓盤租售列表右侧广告栏已还原
- [x] 樓盤詳情面包屑 + 图片画廊 + tabs 已还原
- [x] 服務式住宅列表 HERO + 筛选 + 卡片网格已还原
- [x] 服務式住宅詳情图片画廊 + 详情已还原
- [x] 家具市集列表三栏布局已还原
- [x] 家具市集詳情图片 + 详情 + 卖家信息已还原
- [x] 綜合優惠暗色 HERO + 控制栏 + 4 列卡片网格已还原
- [x] AJO Pay 顶部子导航 + 概覽内容已还原
- [x] 我的大廈左侧大廈列表 + 右侧 tabs 已还原
- [x] 會員中心左侧导航 + 右侧个人资料已还原
- [x] 管理端左侧子导航 + 右侧表格已还原
- [x] 走勢页图表 + 数据表已还原
- [x] 通知中心分类 tabs + 列表已还原
- [x] 登入页双栏布局 + 登录表单已还原
- [x] 收藏页卡片网格已还原

## 验证
- [x] `npm run typecheck` 通过
- [x] 逐页对比 HTML 设计稿与 Vue 渲染结果，布局、配色、字体、间距一致
- [x] 亮色模式（`html-fidelity`）下所有页面视觉与 HTML 亮色一致
- [x] 深色模式（`dark-neutral`）下所有页面视觉与 HTML `body.dark` 一致
- [x] 主题切换按钮在亮/暗模式下图标正确切换
- [x] 移动端导航与底部导航栏样式正确
