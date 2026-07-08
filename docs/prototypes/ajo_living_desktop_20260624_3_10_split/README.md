# AJO Living Desktop Split

## 目錄用途

1. `index.html` 是可直接打開的整合預覽入口。
2. `routes/` 按主路由拆分原型頁面，保留原始 `page-*` id。
3. `partials/` 放全站導覽、浮動工具、燈箱、引導、頁尾與提醒彈窗。
4. `assets/` 放拆出的共用樣式、AJO Pay 樣式與互動腳本。

## 已清理內容

1. 刪除重複的 `scroll-progress`、`toast-container`。
2. 刪除第二套重複的物件比較、客服聊天與租金計算浮層 DOM。
3. 刪除 AJO Pay 隱藏 mobile view、mobile modal 與對應未使用互動。
4. 去除原型中的 emoji 字元，保留繁體中文文案。

## 主路由

- `page-listing` -> `routes/listing.html`：樓盤租售
- `page-market` -> `routes/market.html`：家具市集
- `page-login` -> `routes/login.html`：登入
- `page-home` -> `routes/home.html`：首頁
- `page-service` -> `routes/service.html`：服務式住宅
- `page-offers` -> `routes/offers.html`：綜合優惠
- `page-payment` -> `routes/payment.html`：AJO Pay
- `page-detail` -> `routes/detail.html`：物件詳情
- `page-market-detail` -> `routes/market-detail.html`：家具詳情
- `page-service-detail` -> `routes/service-detail.html`：服務式住宅詳情
- `page-notif` -> `routes/notif.html`：通知中心
- `page-affairs` -> `routes/affairs.html`：我的大廈
- `page-profile` -> `routes/profile.html`：會員中心
- `page-management` -> `routes/management.html`：管理中心
- `page-trend` -> `routes/trend.html`：租金走勢
- `page-saved` -> `routes/saved.html`：收藏清單
