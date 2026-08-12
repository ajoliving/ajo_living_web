# http_service/internal

> Level/parent: ../AGENTS.md

## 成員清單
config/: 環境配置讀取與配置物件管理。
database/: 資料庫連線、初始化、遷移輔助與 seed。
errcode/: 統一回應格式、錯誤碼與錯誤訊息。
handler/: HTTP 參數解析、基礎校驗、service 呼叫與回應輸出。
logger/: 專案統一日誌入口。
middleware/: 認證、限流、CORS、恢復、請求 ID 與訪問日誌。
model/: GORM 模型、分頁結構與基礎資料結構。
router/: 路由註冊、中間件掛載與資源路由分組。
service/: 業務邏輯、資料庫讀寫、交易處理與第三方服務編排。
utils/: 無業務狀態的通用工具。

## 架構決策
- 依賴方向保持 `router -> handler -> service -> model/database`；不得讓 handler 直接操作資料庫或讓 service 輸出 HTTP 回應。
- 新增能力前先判斷主域，避免把支付、通知、設備、物業、交易、帳戶等能力混到單一 service。
- 舊系統接入優先做受控讀取、身份映射與權限校驗；未確認寫入責任前按唯讀處理。
- `middleware.RequireStaff` 只可掛載於管理中心、管理設定及審核接口；一般會員 handler 與 service 不得以 AJO 或 iSmart Staff 狀態作功能准入條件。
- `users.is_staff` 是 AJO 管理身份的唯一運行時來源；`user_ismart_accounts.is_staff` 只作上游業務狀態及請求元數據，不得同步提升 AJO 管理身份。一般 iSmart/POS 功能統一按 `client_building_permissions` 與 `client_building_flat_units_permissions` 控制所屬範圍。
- 資料模型、API 回應與錯誤語義變更時，同步檢查 `http_service/doc/api.md`、相關測試與 `docs/PROMPT_INDEX.md` 是否受影響。
- 帳戶身份與樓盤放售跨域時，以 `user_profiles.account_type`、已批准 `agency_profiles` 及公司子帳戶歸屬為權威來源；`publisher_identity_type` 只作兼容派生，不得由會員修改。
- 代理帳戶在首次資料批准前使用受限 session；除 `/me`、代理資料、同類型代理 OSS 上傳及登出外，會員業務統一要求 `member_status=active`。
- `pending_profile`、`pending_review`、`rejected` 代理帳戶必須仍可使用本地電郵、手提電話或用戶名稱密碼登入，以查看進度、拒絕原因及重新提交；不得在發出 token 前按非 active 狀態拒絕登入。
- 統一帳戶登入使用 `POST /auth/login` 接收 `identifier` 與 `password`，自動識別電郵、香港或中國內地手提電話及 username；本地帳戶優先，只有明確憑證拒絕才回退 iSmart。資料庫、網絡或上游服務錯誤不得觸發回退或偽裝成密碼錯誤，舊分類登入接口保持兼容。
- 代理資料保留一個已批准 active profile 及一個 revision；修訂待審或被拒時沿用舊 active profile，修訂批准後在同一交易刷新該代理及公司子帳戶全部未刪除樓盤的公開聯絡快照。
- 代理審核結果先提交資料庫交易與站內通知，再以 `MailSender` 發送電郵；郵件失敗只記錄日誌，不得回滾或改寫審核狀態，拒絕郵件必須包含審核原因。
- 公司子帳戶建立、發布及重新發布樓盤要求 `property_publish`，更新、下架及標記售出要求 `property_manage`；公司主帳戶可統一管理所屬子帳戶樓盤。
- 代理 EAA 牌照及商業登記證以 private ACL 上傳，會員本人與管理審核只能取得短時 OSS 簽名下載地址；不得以公開 CDN URL 返回證件。
- 樓盤標題及單位介紹翻譯歸入 property service；handler 只綁定輸入，service 負責長度校驗、DeepL 請求與上游錯誤隔離，router 負責會員鑑權及限流。
- 樓盤聯絡解鎖由 property service 按 `contact_attributes` 的電話1與電話2區號及 WhatsApp 狀態分別生成 URL；電話2缺少獨立區號時兼容使用電話1區號。不得向公開列表暴露原始電話，只有受控解鎖回應可返回電話或聯絡入口。
- 樓盤公開摘要由 property service 同時映射 `floor_raw` 實際樓層及 `floor_zone` 公開樓層，公開列表與詳情共用此契約；聯絡資料與 `private_note` 不得因此放寬。
- 樓盤草稿收費由 handler 映射 `charge_draft=true`，property service 在同一交易內首次保存資料並預扣 600 AJO Points；發布總費 1,000，由 service 查詢同一 `listing_id` 的草稿扣款流水後只收取未付差額。舊草稿超額預付不再扣款，未帶標記的發布前暫存及已預付草稿的後續保存不得收取草稿費。
- 歷史草稿重複扣費只可由受限維護流程對指定、未發布的 `property_sale` 以 `listing_refund` / `draft_charge_correction` 新增退款流水；不得改寫原扣款。後續草稿抵扣查詢必須扣除該專用退款，維持淨預付 600 與發布待付 400。
- AJO 註冊建立 iSmart 帳戶時，註冊值只可補足 iSmart 回應缺失的受控會員資料快照；快照不得保存密碼，後續 POS 基礎登入回應不得覆蓋既有完整資料。若需返回上游原始業務欄位，使用獨立 `ismart_raw`，遞迴排除密碼、token、secret、authorization、憑證及 session 欄位，不與固定 `ismart_msg` 摘要混用。
- 個人帳戶註冊若同時提交大廈、樓層及單位，必須在 iSmart 建戶與本地關聯提交後才發送 OwnerReg 審批申請；僅 HTTP `2xx` 代表申請已受理，`residence_binding_status=pending` 不得寫入 `bound_building_ids` 或 `bound_flat_unit_ids`，審批前不得視為已綁定。
- 會員從已授權 iSmart 單位保存本地 `bound_building_ids` 後，大廈服務必須優先使用該明確選擇，即使帳戶仍保留舊 pending 記錄；只有未寫入本地綁定的 OwnerReg 待審申請繼續受 pending 限制。社區名稱若仍等於 `public_id` 佔位值，收到正式大廈名稱時必須補正。
- iSmart 大廈資料與文件中繼資料可按 `building_id` 使用 Redis 共用快取 5 分鐘；`resolveBuildingAccess` 必須先於快取讀取執行，`building_options`、Staff 狀態及其他會員欄位只可在回應階段即時組裝，不得寫入共用快取。Redis 故障時直接回源 iSmart。
- 會員 POS 大廈列表必須為每個可見 `building_id` 返回一條記錄；會員 relay 只返回部分大廈時，以 AJO 公共 POS 目錄補齊正式名稱，公共目錄仍缺失時才保留 ID 回退記錄。
- POS 大廈與單位目錄可使用 Redis 共用快取 5 分鐘；會員接口必須先即時讀取本地 profile 與 iSmart 權限並完成大廈可見性校驗，再讀取共用目錄及過濾單位。快取不得保存 relay token、會員權限、綁定狀態或目前物業，Redis 或公共目錄失敗時可回退會員 relay。
- 大廈住戶授權以本地 `building_authorizations` 保存電話、電郵、受邀帳戶及功能範圍；未註冊受邀者使用一次性電郵連結設定帳戶，已註冊帳戶即時生效。遙距開門及通告讀取必須在實際操作 API 按大廈及功能範圍校驗。
- 部署與環境操作不在 `internal/` 內新增 prompt，統一回到 `docs/deployment/`。

## 變更日誌
2026-07-08: 建立後端 internal 目錄記憶，固定分層、主域與 prompt 影響檢查入口。
2026-07-14: 記錄帳戶身份對樓盤放售建立、編輯與發布流程的強制約束，業主與地產代理均可發布。
2026-07-14: 記錄個人帳戶發布身份的預設值、會員自助更新與舊值兼容契約。
2026-07-14: 記錄樓盤內容翻譯的 handler、service、router 與供應商安全邊界。
2026-07-14: 記錄樓盤雙電話 WhatsApp 解鎖與原始電話保密邊界。
2026-07-14: 記錄樓盤電話2獨立區號及舊資料回退契約。
2026-07-15: 記錄樓盤明確儲存草稿的查詢參數、原子扣款與發布前暫存免收草稿費契約。
2026-07-15: 記錄個人代理、代理公司、受限登入、版本化資料審核、公司子帳戶及帳戶派生樓盤身份契約。
2026-07-15: 收緊公司子帳戶樓盤發布與管理權限，並固定代理證件 private ACL 及短時簽名下載契約。
2026-07-15: 固定代理修訂待審期間沿用舊快照，批准後原子刷新代理及公司子帳戶未刪除樓盤聯絡快照。
2026-07-15: 固定待審與被拒代理的受限登入進度查詢，以及審核結果站內信與非阻斷電郵通知契約。
2026-07-16: 固定 iSmart 註冊資料快照的回應優先、註冊補足、密碼排除及後續 POS 登入保留規則。
2026-07-16: 固定 iSmart 原始業務資料以獨立 `ismart_raw` 返回、遞迴脫敏及與摘要分離的 API 契約。
2026-07-16: 固定個人註冊的 iSmart OwnerReg 申請順序、HTTP 受理條件、pending 狀態與未審批不建立有效物業綁定的契約。
2026-07-16: 固定 AJO 與 iSmart Staff 身份分離、管理中心專屬 Staff 守衛及一般會員接口按所屬資源可見範圍授權的契約。
2026-07-16: 固定已授權本地大廈選擇優先於舊 pending 狀態，並在會員保存正式名稱時修正社區 ID 佔位名稱。
2026-07-16: 固定 iSmart 大廈資料 Redis 快取的共享資料邊界、5 分鐘 TTL、權限先行及故障回源規則。
2026-07-16: 固定會員 POS 大廈列表的完整性，部分 relay 結果須以公共 POS 目錄補齊缺失大廈及正式名稱。
2026-07-17: 固定 POS 大廈與單位共用目錄 Redis 快取、會員權限先行、個人資料排除及 relay 回退契約。
2026-07-17: 固定統一帳戶登入的自動分類、本地優先、受控 iSmart 回退、故障保真及舊接口兼容契約。
2026-07-28: 固定樓盤草稿預付流水作為發布差額抵扣依據，發布交易鎖定樓盤並兼容歷史超額草稿扣款。
2026-07-28: 記錄舊版草稿重複扣費的不可變退款流水、淨預付查詢與受限維護命令邊界。
2026-07-30: 記錄代理註冊即上載牌照、建立待審版本及 Staff 批准後啟用帳戶的 API 契約；個人代理可不提供電郵。
2026-07-31: 記錄樓盤公開摘要同時返回實際樓層與公開樓層，並保留聯絡資料及內部備註的私密邊界。
2026-08-12: 固定大廈住戶授權的本地帳戶邀請、功能範圍及實際操作 API 鑑權契約。

[PROTOCOL]: When adding, moving, renaming, or changing backend internal ownership, update this map, check parent `../AGENTS.md`, and update `../../docs/PROMPT_INDEX.md` when API documentation prompts or model-facing contracts changed.
