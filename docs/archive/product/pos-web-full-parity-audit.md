# POS Web 支付中心逐頁 Parity 審計

## 結論

- 本審計以 AJO `web/src/pages/payments`、`web/src/domains/payments`、`http_service/internal/service/pos_payment*`，以及舊 POS Web `/Users/yangliu/Documents/Code/hk/pos/pos_web/src` 為依據。
- AJO 已覆蓋舊 POS Web 的物業費賬單、購物車、H5 / QR 支付、銀行轉賬、支票、現金、訂單、交易歷史與會計清機主流程。
- 登入與選單位屬於 AJO 等價改造：AJO 使用統一會員身份、ismart 綁定與會員資料保存單位；不應回退為 POS Web 的獨立登入與本地 selection store。
- 本輪已補齊兩項明確差異：H5 建單會送 `expire_seconds=180`；Staff 歷史預設查詢不再被目前選中單位限制。
- AJO `/payments` 是支付中心 shell，不直接請求支付 API；`/payments/records` 只是 redirect 到 `/payments/history`；`/payment/order`、`/payment/result`、`/payment/h5/redirect` 是舊回傳兼容路由，不作獨立支付中心 page 審計。

## 頁面對照矩陣

| AJO 頁面 | 舊 POS Web 頁面 | AJO 使用 API | 舊 POS Web 使用 API | 頁面顯示字段 | 提交 / 狀態字段 | 狀態 | 修改方案 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `/payments` shell | POS Web `AppLayout` | 無支付 API；只讀 AJO session state 判斷是否顯示 `會計` | 無支付 API；舊 layout 讀登入狀態判斷是否顯示 `/accounting` | 左側支付導航、子頁承載區、`會計` Staff 可見性 | `is_staff` / `ismart_msg.is_staff` | covered | 保持；父級 shell 不應額外請求 `overview` |
| `/payments/units` | `/unit-selecting` | `GET /pos/buildings`、`GET /pos/buildings/{buildingId}/units`、`PATCH /me/profile` | `GET /building`、`GET /building/{buildingId}/units` | 大廈、樓層、單位；來源字段包括 `building_id`、`buildname_chi`、`buildname`、`unit_id`、`floor`、`unit`、`unit_name` | AJO 保存 `primary_community_id`、`primary_community_name`、`residence_floor`、`residence_unit` | AJO 等價改造 | 保持 AJO 會員資料作支付單位真相；不恢復 POS Web local selection 作主真相 |
| `/payments/bills` | `/billing` | `GET /me/payments/pos/overview`、`GET /me/payments/pos/bills` | `GET /unit/{unitID}/bills`，帶 `pos_type=web_staff/web_client` | AJO 顯示賬單號、項目、期數、賬單日、狀態、金額；字段包括 `invoice_no`、`bill_no`、`bill_number`、`item_name`、`item_id`、`trs_to`、`bill_dt`、`status`、`net_amount` | 加入購物車使用賬單 key 與當前 `building_id`、`unit_id` 隔離 | covered | 若要完全貼近舊頁，可補顯示「單位」與「備註」；付款主鏈路不受影響 |
| `/payments/cart` | `/cart` | `GET /me/payments/pos/overview`、`GET /me/payments/pos/fees`、`GET /me/payments/pos/bank-accounts`、`GET /me/payments/pos/bills`、`POST /me/payments/pos/orders`、`POST /me/payments/pos/payments/report`、`POST /me/payments/pos/terminal/pay` | `GET /building/{buildingID}/fee`、`GET /building/{buildingID}/bank_account`、`GET /unit/{unitID}/bills`、`POST /h5/orders`、`POST /bill`、瀏覽器本機 `fetch(posAddr)` | 賬單、項目、期數、賬單日、金額、實付金額、手續費、結賬總額、付款方式、交易時間、交易參考、銀行戶口、憑證、備註、找零 | H5: `scene`、`pay_channel`、`expire_seconds`、`final_amount`、`handle_fee_amount`、`bill_objs`、`handle_fee_obj`、`return_path`、`remark`、`gateway_request_overrides`；線下: `FINAL_AMOUNT`、`ENTRY_DATETIME`、`TRAN_DATETIME`、`TRAN_REF_NO`、`COMMENT`、`PIC_FILENAME`、`PIC_DATA`、`BILL_OBJS`、`PAY_METHOD`、`BLG_ID`、`UNIT_ID`、`bank_account_received` | covered | 保留提交前重新拉賬單校驗金額；POS 機維持 AJO 後端代理；手機 / 平板優先 H5，桌面優先 QR，但所有可用入口保留 |
| `/payments/orders` | `/orders` | `GET /me/payments/pos/orders`、`GET /me/payments/pos/orders/{mchOrderNo}`、`POST /me/payments/pos/orders/query`、`POST /me/payments/pos/orders/{mchOrderNo}/close`、`POST /me/payments/pos/orders/{mchOrderNo}/cancel` | 本地 `pos_local_orders`、`GET /h5/orders/{mchOrderNo}`、`POST /h5/orders/query`、`POST /h5/orders/{mchOrderNo}/close`、`POST /h5/orders/{mchOrderNo}/cancel`、`GET /events` | 訂單號、支付單號、支付方式、支付狀態、回寫狀態、時間、金額、QR、訂單詳情、賬單明細、request / response 處理摘要 | `mch_order_no`、`pay_order_id`、`state`、`business_state`、`pay_data_type`、`pay_data`、`gateway_message`、`business_message`、`receipt_no`、`close_reason` | AJO 等價增強 | AJO 使用服務端訂單快照作真相，不遷回瀏覽器本地訂單；30 秒輪詢可替代舊 SSE，除非後續明確要求即時 SSE |
| `/payments/accounting` | `/accounting` | `GET /me/payments/pos/overview?summary=false`、`GET /me/payments/pos/accounting`、`POST /me/payments/pos/accounting/clear`、`GET /me/payments/pos/accounting/records/{recordId}` | `GET /building/{buildingID}/accounting`、`POST /accounting`、`GET /building/{buildingID}/accounting_record`、`GET /building/{buildingID}/accounting_record/{recordId}` | 現金待清機、支票待清機、交易號、參考號、時間、金額、清機歷史、清機詳情、交易明細 | `payment_id_list`；回應字段包括 `payment_objs_cash`、`payment_objs_cheque`、`record_id`、`payment_detail_objs`、`trs_val` | covered | 保持；AJO 合併待清機與清機歷史，減少頁面跳轉 |
| `/payments/history` | `/transaction-history` | `GET /me/payments/pos/overview`、`GET /me/payments/pos/history`、`GET /me/payments/pos/history/{paymentId}` | Staff: `GET /building/{buildingID}/bill_history`；會員: `POST /transactions/flat_units`；另讀 `GET /building/{id}`、`GET /building/{id}/units` 用於會員可查單位 | 日期範圍、日期類型、收款方式、單位多選、交易、收據、狀態、時間、金額、交易詳情、賬單明細 | `from_date`、`to_date`、`date_type`、`pay_method`、`unit_ids` / `unit_id_list`；詳情字段包括 `payment_id`、`receipt_id`、`pay_method`、`tran_datetime`、`input_time`、`bank_account_received` | covered | Staff 預設省略 `unit_id`，讓後端走 building 級 `bill_history`；只有手動選單位時才傳 `unit_ids` |
| `/payments/records` | 無獨立頁 | 無 | 無 | 無；路由直接 redirect 到 `/payments/history` | 無 | intentional-hidden | 保持 redirect，不作獨立 page |

## API 對照矩陣

| 能力 | 舊 POS Web API | AJO API | AJO 後端實際代理 / 處理 | 狀態 |
| --- | --- | --- | --- | --- |
| POS 登入 | `POST /login`、`POST /poslogin`、`POST /renew` | AJO 會員登入與 ismart 綁定 | 後端保存 relay token，POS 代理失效時刷新 | AJO 等價改造 |
| 樓宇列表 | `GET /building` | `GET /pos/buildings` | 代理 POS building list | covered |
| 單位列表 | `GET /building/{buildingID}/units` | `GET /pos/buildings/{buildingId}/units` | 代理 POS units list | covered |
| 賬單列表 | `GET /unit/{unitID}/bills` | `GET /me/payments/pos/bills` | 代理 `/unit/{unitID}/bills`，補 `pos_type` | covered |
| 手續費 / 支付方式 | `GET /building/{buildingID}/fee` | `GET /me/payments/pos/fees` | 代理 `/building/{buildingID}/fee`，補 `pos_type` | covered |
| 銀行戶口 | `GET /building/{buildingID}/bank_account` | `GET /me/payments/pos/bank-accounts` | 代理 `/building/{buildingID}/bank_account` | covered |
| 線下繳費 | `POST /bill` | `POST /me/payments/pos/payments/report` | 校驗當前單位可見後代理 `/bill` | covered |
| POS 機收款 | 瀏覽器直連 `posAddr` 後 `POST /bill` | `POST /me/payments/pos/terminal/pay` | Staff 限定，後端讀服務端 terminal URL，成功後代理 `/bill` | AJO 等價改造 |
| H5 建單 | `POST /h5/orders` | `POST /me/payments/pos/orders` | 後端補 `expire_seconds=180`、`report_payment_data`、`operator`、`clientIp` 後調 H5 payment service | covered |
| H5 查詢 | `GET /h5/orders/{mchOrderNo}`、`POST /h5/orders/query` | `GET /me/payments/pos/orders/{mchOrderNo}`、`POST /me/payments/pos/orders/query` | 校驗訂單單位歸屬後返回 | covered |
| H5 詳情 | `GET /h5/orders/{mchOrderNo}/detail` | `GET /me/payments/pos/orders/{mchOrderNo}?detail=1` | 校驗歸屬後返回詳細字段 | covered |
| H5 關閉 / 取消 | `POST /h5/orders/{mchOrderNo}/close`、`POST /cancel` | AJO 同名代理 | 校驗歸屬後代理 | covered |
| 訂單快照 | `GET /order-records`、`POST /order-records/bulk` | `GET /me/payments/pos/orders` | AJO 只讀服務端快照，不使用瀏覽器 bulk sync 作主真相 | AJO 等價改造 |
| 交易歷史 | `GET /building/{buildingID}/bill_history`、`POST /transactions/flat_units` | `GET /me/payments/pos/history` | Staff 預設走 building，手動選單位時走 multi-unit；會員走 visible flat units | covered |
| 清機待處理 | `GET /building/{buildingID}/accounting` | `GET /me/payments/pos/accounting` | Staff building 權限校驗後代理 | covered |
| 清機提交 | `POST /accounting` | `POST /me/payments/pos/accounting/clear` | 先確認 payment id 屬於待清機列表，再代理 | covered |
| 清機歷史 / 詳情 | `GET /accounting_record`、`GET /accounting_record/{recordId}` | `GET /me/payments/pos/accounting/records`、`GET /records/{recordId}` | Staff building 權限校驗後代理 | covered |

## 支付方式鏈路矩陣

| 支付方式 | 舊 POS Web 鏈路 | AJO 鏈路 | 業務狀態 | 修改方案 |
| --- | --- | --- | --- | --- |
| 微信 H5 | `POS_WECHAT` 按設備解析為 `WX_H5`，建立 H5 訂單，跳轉支付，返回 `/orders` 查單 | 使用者可選 `WX_H5`，AJO 建單後直接跳轉 `pay_data`，返回 `/payments/orders` 查單；手機 / 平板優先展示 H5 | covered | 保持 |
| 微信 QR | Desktop 解析為 `WX_QR`，訂單頁展示 QR | 使用者可選 `WX_QR`，訂單頁展示 QR；桌面優先展示 QR | covered | 保持 |
| 支付寶大陸 H5 | `ALI_H5` + `channelExtra={"walletType":"CN"}` | `ALI_H5` + AJO 受控 `walletType=CN` | covered | 保持 AJO 文檔契約；無需回退為前端 `channelExtra` |
| 支付寶香港 H5 | `ALI_H5` + `channelExtra={"walletType":"HK"}` | `ALI_H5` + AJO 受控 `walletType=HK` | covered | 保持 |
| 支付寶 QR | 舊頁可把桌面支付寶 H5 以 QR 方式展示，也有 `ALI_QR` 支持 | AJO 顯式提供 `ALI_QR` | covered | 保持 |
| 雲閃付 QR | `POS_YSF_QR` / `YSF_QR` 建單後展示 QR | `YSF_QR` 建單後展示 QR | covered | 保持 |
| 銀行轉賬 | 拉銀行戶口，輸入交易時間、備註，上傳憑證，`POST /bill` | 同樣拉銀行戶口、憑證、時間、備註，經 AJO `payments/report` 代理 `/bill` | covered | 保持；需用真實舊 POS 驗證 `PIC_DATA` data URL 接受情況 |
| 支票 | 支票號、是否已存入銀行、銀行戶口，`POST /bill` | 同樣校驗支票號、是否入銀行、銀行戶口，代理 `/bill` | covered | 保持 |
| 現金 | 交易時間、實收現金、找零，`POST /bill` | 同樣校驗實收現金與找零，代理 `/bill` | covered | 保持 |
| POS_CARD 銀行卡 | 舊頁存在 POS 直連能力，但 current `mapPaymentMethods` 會排除 `POS_CARD`，實際依賴本機 `posAddr` 設定 | AJO 只對 Staff 開放後端 terminal proxy，服務端配置設備地址 | AJO 等價改造 | 保持後端代理與預設關閉；不允許前端提交內網 POS 地址 |

## 字段 Parity 檢查

| 字段群 | AJO 使用位置 | 舊 POS Web 使用位置 | 狀態 | Action |
| --- | --- | --- | --- | --- |
| 單位上下文 `building_id`、`unit_id`、`floor`、`unit` | 所有 `overview`、`bills`、`cart`、`history`、`accounting` API | `selectionStore.current`、賬單與歷史查詢 | covered | 保持 AJO 後端做可見性校驗 |
| 賬單識別 `invoice_no`、`bill_no`、`bill_number`、`bill_id`、`id` | 賬單列表、購物車 key、訂單 bill objects | 賬單列表、購物車 key、local order items | covered | 保持 |
| 賬單展示 `item_name`、`item_id`、`trs_to`、`bill_dt`、`net_amount` | 賬單、購物車、訂單明細、歷史詳情 | 賬單、購物車、訂單本地快照、歷史詳情 | covered | AJO 可補 `remark` 顯示 |
| 支付金額 `paid_amount`、`final_amount`、`handle_fee_amount` | 購物車自訂實付、H5 建單、線下上報 | 購物車自訂實付、H5 建單、線下上報 | covered | 保持提交前最新賬單校驗 |
| 線下交易 `TRAN_DATETIME`、`TRAN_REF_NO`、`COMMENT`、`PIC_FILENAME`、`PIC_DATA` | AJO 線下上報 | POS Web 線下上報 | covered | 保持 |
| 銀行戶口 `bank_account_received` | 銀行轉賬、支票入銀行 | 銀行轉賬、支票入銀行 | covered | 保持 |
| H5 訂單 `mch_order_no`、`pay_order_id`、`pay_data_type`、`pay_data`、`state`、`business_state`、`receipt_no` | 訂單列表與詳情 | 本地 order store 與 H5 查單 | covered | AJO 以服務端快照為真相 |
| H5 到期 `expire_seconds` / `expire_time` | AJO 建單明確送 `expire_seconds=180`，訂單頁顯示 `expire_time` | POS Web 建單明確送 `expire_seconds=180` | covered | 保持 |
| Staff 歷史單位範圍 | AJO Staff 預設不傳 `unit_id`，手動選單位才傳 `unit_ids` | POS Web Staff 預設 building 級 `bill_history` | covered | 保持 |

## 建議修改清單

| 優先級 | 修改 | 理由 | 涉及範圍 |
| --- | --- | --- | --- |
| P0 已完成 | AJO H5 建單補 `expire_seconds=180` | 對齊舊 POS Web 的三分鐘支付窗口與本地訂單過期邏輯 | AJO 後端 H5 建單 payload，前端型別 |
| P0 已完成 | Staff 歷史預設查 building，不帶當前 `unit_id` | 對齊 POS Web Staff `bill_history` 預設查整棟；避免 Staff 以為已查全樓但實際只查當前單位 | AJO history page query build |
| P1 | AJO 賬單列表補「單位」「備註」顯示 | 視覺字段更貼近舊 POS Web；不影響支付 | AJO bills page |
| P1 | 手工驗證銀行轉賬 `PIC_DATA` 格式 | 舊 POS 是否接受同樣 data URL 需要真實服務確認 | 線下上報驗收 |
| P2 | 如現場要求秒級訂單更新，再做 SSE 後端代理 | 目前 30 秒輪詢足以等價，但不是完全相同 | AJO orders page + backend |

## 驗證狀態

- 已完成：靜態代碼審查與二次復核，覆蓋 AJO 前端頁面、AJO 後端代理與舊 POS Web 對應頁面。
- 已完成：P0 兩項修正，並補充購物車支付方式裝置偏好排序。
- 已通過：`npm run build`。
- 已通過：`go test ./internal/service -run 'TestBuildH5OrderPayloadKeepsOldReportFields|TestSanitizePOSGatewayRequestOverrides|TestNormalizePOSPaymentScene'`。
- 已通過：`go test ./...`。
- 未執行：真實支付、真實 POS relay、真實 POS terminal；未啟動前後端服務。
