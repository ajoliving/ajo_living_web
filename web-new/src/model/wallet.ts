/*
 * AJO Point 錢包型別。
 * 1. 對齊會員錢包、積分流水與廣告任務接口。
 * 2. 為帳戶錢包頁與發布扣分提示提供穩定型別。
 */

// 1. 定義錢包帳戶摘要
export interface WalletAccountResponse {
  balance: number;
  total_earned: number;
  total_spent: number;
  today_ad_reward_points: number;
  daily_ad_reward_limit: number;
}

// 2. 定義錢包流水
export interface WalletTransactionResponse {
  transaction_id: string;
  direction: 'credit' | 'debit';
  amount: number;
  balance_before: number;
  balance_after: number;
  source_type: string;
  biz_module: string;
  action_type: string;
  note: string;
  created_at: string;
}

// 3. 定義三模組扣費規則
export interface WalletChargeRuleResponse {
  biz_module: 'secondhand' | 'property_sale' | 'serviced_apartment';
  label: string;
  publish: number;
  draft_save?: number;
  edit: number;
  republish: number;
  renew?: number;
}

// 4. 定義錢包總覽
export interface WalletOverviewResponse {
  account: WalletAccountResponse;
  recent_transactions: WalletTransactionResponse[];
  charge_rules: WalletChargeRuleResponse[];
  recharge_enabled: boolean;
  recharge_rate: number;
}

// 5. 定義廣告任務
export interface RewardAdTaskResponse {
  task_id: string;
  title: string;
  summary: string;
  cover_url: string;
  media_url: string;
  media_type: 'image' | 'video';
  target_url: string;
  reward_points: number;
  watch_seconds: number;
  can_claim_today: boolean;
  claimed_today: boolean;
  remaining_budget: number;
  watch_count: number;
  total_watch_seconds: number;
  link_click_count: number;
  link_click_rate: number;
}

// 6. 定義廣告觀看會話
export interface RewardAdSessionResponse {
  claim_id: string;
  task_id: string;
  watch_seconds: number;
  started_at: string;
  available_at: string;
  reward_points: number;
}

// 7. 定義廣告點擊統計結果
export interface RewardAdClickResponse {
  task_id: string;
  target_url: string;
  watch_count: number;
  link_click_count: number;
  link_click_rate: number;
}

// 8. 定義扣費或獎勵結果
export interface PointsChargeResponse {
  points_charged: number;
  points_balance_after: number;
  points_transaction_id: string;
}

// 9. 定義 Staff 錢包會員摘要
export interface StaffWalletUserResponse {
  user_id: string;
  phone_country_code: string;
  phone_number: string;
  display_name: string;
}

// 10. 定義 Staff 錢包流水
export interface StaffWalletTransactionResponse {
  transaction_id: string;
  direction: 'credit' | 'debit';
  amount: number;
  balance_before: number;
  balance_after: number;
  source_type: string;
  biz_module: string;
  action_type: string;
  note: string;
  target_user: StaffWalletUserResponse;
  operator_user?: StaffWalletUserResponse;
  created_at: string;
}

// 11. 定義 Staff 積分發放結果
export interface StaffWalletGrantResponse {
  charge: PointsChargeResponse;
  target_user: StaffWalletUserResponse;
  operator: StaffWalletUserResponse;
}

// 12. 定義 Staff 廣告任務
export interface StaffRewardAdResponse {
  task_id: string;
  ad_type: 'reward' | 'display';
  title: string;
  summary: string;
  cover_url: string;
  media_url: string;
  media_type: 'image' | 'video';
  target_url: string;
  display_channel: 'property_sale' | 'serviced_apartment' | 'furniture' | '';
  display_placement: 'listing_side' | '';
  display_layout: 'image_full' | 'image_text' | 'text_compact';
  slot_display_title: string;
  display_text: string;
  slot_target_url: string;
  sort_order: number;
  reward_points: number;
  watch_seconds: number;
  total_budget: number;
  total_granted: number;
  remaining_budget: number;
  watch_count: number;
  total_watch_seconds: number;
  link_click_count: number;
  link_click_rate: number;
  is_active: boolean;
  starts_at?: string;
  ends_at?: string;
  created_at: string;
  updated_at: string;
}

// 13. 定義 Staff 廣告任務保存資料
export interface StaffRewardAdPayload {
  ad_type?: 'reward' | 'display';
  title?: string;
  summary?: string;
  cover_url?: string;
  media_url?: string;
  media_type?: 'image' | 'video';
  target_url?: string;
  display_channel?: 'property_sale' | 'serviced_apartment' | 'furniture' | '';
  display_placement?: 'listing_side' | '';
  display_layout?: 'image_full' | 'image_text' | 'text_compact';
  sort_order?: number;
  reward_points?: number;
  watch_seconds?: number;
  total_budget?: number;
  is_active?: boolean;
  starts_at?: string;
  ends_at?: string;
  retention_days?: number;
}

// 14. 定義公開展示廣告
export interface PublicDisplayAdResponse {
  task_id: string;
  title: string;
  display_title: string;
  summary: string;
  display_text: string;
  cover_url: string;
  media_url: string;
  media_type: 'image' | 'video';
  target_url: string;
  display_channel: 'property_sale' | 'serviced_apartment' | 'furniture';
  display_placement: 'listing_side';
  display_layout: 'image_full' | 'image_text' | 'text_compact';
  sort_order: number;
  slot_index: number;
}

// 15. 定義展示廣告位設定
export interface DisplayAdSlotResponse {
  slot_index: number;
  layout: 'image_full' | 'image_text' | 'text_compact';
  ads: StaffRewardAdResponse[];
}

// 16. 定義展示廣告頻道設定
export interface DisplayAdChannelSettingsResponse {
  channel: 'property_sale' | 'serviced_apartment' | 'furniture';
  slots: DisplayAdSlotResponse[];
}

// 17. 定義展示廣告位保存項
export interface DisplayAdSlotSaveItem {
  slot_index: number;
  ad_task_ids: string[];
  ads: Array<{
    ad_task_id: string;
    display_title: string;
    display_text: string;
    target_url: string;
  }>;
}

// 18. 定義充值支付方式
export type WalletRechargePayMethod = 'wechat' | 'alipay' | 'unionpay';

// 19. 定義充值支付地區
export type WalletRechargePayRegion = 'HK' | 'CN';

// 20. 定義充值設備模式
export type WalletRechargeDeviceMode = 'mobile' | 'desktop';

// 18. 定義充值訂單
export interface WalletRechargeOrderResponse {
  order_id: string;
  mch_order_no: string;
  pay_order_id: string;
  pay_channel: 'WX_H5' | 'ALI_H5' | 'WX_QR' | 'ALI_QR' | 'YSF_QR';
  pay_region: WalletRechargePayRegion;
  pay_data_type: string;
  pay_data: string;
  state: 'PAYING' | 'SUCCESS' | 'FAILED' | 'CLOSED' | 'EXPIRED' | 'REVOKED' | 'REFUNDED';
  state_label: string;
  gateway_state_code: number;
  gateway_message: string;
  currency: string;
  amount_hkd: number | string;
  amount_cents: number;
  points_amount: number;
  created_at: string;
  updated_at: string;
  expire_time: string;
  paid_at: string;
  credited_at: string;
}
