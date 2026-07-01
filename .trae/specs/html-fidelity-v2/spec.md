# HTML 设计稿 v2 高保真还原 Spec

## Why
HTML 设计稿更新至 `ajo_living_desktop_20260624(3)(10).html`，相比旧版有 3 处重大变化：导航栏新增 AJO Pay 上下文切换、AJO Pay 页面完全重写、樓盤租售列表新增多项功能。需要将这些变化同步到 Vue 项目。

## What Changes
- 修改 `AppHeader.vue`：新增 AJO Pay 上下文切换栏（nav-pay-ctx），进入 `/payments` 路由时显示 AJO PAY 品牌 + 4 个 tab（首頁、付款、AJO Coin 回贈、返回 AJO Living）
- 重写 `payments/Page.vue`：还原 3 个子页面（HOME/PAYMENT/COIN），包含 Hero、统计、功能卡片、付款流程步骤、AJO Coin 回贈计划
- 重写 `PropertyListPage.vue`：新增搜索自动补全、地图视图（map-view with pins/popup）、售/租价格标识（listing-price-kind）、代理公司信息（listing-agent）、呎價（garea-price）、比较 checkbox

## Impact
- Affected specs: html-fidelity-restore
- Affected code:
  - `web-new/src/shared/components/navigation/AppHeader.vue`
  - `web-new/src/pages/payments/Page.vue`
  - `web-new/src/pages/property/list/PropertyListPage.vue`

## ADDED Requirements
### Requirement: AJO Pay 上下文导航
The system SHALL display an AJO Pay context navigation bar when the user is on `/payments` routes, showing AJO PAY brand + 4 tabs (首頁、付款、AJO Coin 回贈、返回 AJO Living).

#### Scenario: 进入 AJO Pay 页面
- **WHEN** user navigates to `/payments`
- **THEN** the nav-pay-ctx bar appears, main nav-links hidden
- **AND** clicking 返回 AJO Living returns to home

### Requirement: AJO Pay 三子页面
The system SHALL render 3 sub-pages within AJO Pay: HOME (Hero + stats + features), PAYMENT (4-step payment flow), COIN (回贈计划 + redeem grid).

#### Scenario: 切换 AJO Pay 子页面
- **WHEN** user clicks 首頁/付款/AJO Coin 回质 tab
- **THEN** the corresponding sub-page displays with fade-in animation

### Requirement: 樓盤租售列表增强
The system SHALL display search autocomplete, map view toggle, price kind badges (售/租), agent info, per-sqft price, and compare checkboxes in the property list page.

#### Scenario: 切换列表/地图视图
- **WHEN** user clicks map view toggle
- **THEN** the map view displays with price pins and clickable popups

## MODIFIED Requirements
### Requirement: 导航栏
导航栏在 AJO Pay 路由下显示 AJO Pay 上下文栏，其他路由保持原有 11 项导航。
