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
  edit: number;
  republish: number;
}

// 4. 定義錢包總覽
export interface WalletOverviewResponse {
  account: WalletAccountResponse;
  recent_transactions: WalletTransactionResponse[];
  charge_rules: WalletChargeRuleResponse[];
  recharge_enabled: boolean;
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

// 7. 定義扣費或獎勵結果
export interface PointsChargeResponse {
  points_charged: number;
  points_balance_after: number;
  points_transaction_id: string;
}

// 8. 定義 Staff 錢包會員摘要
export interface StaffWalletUserResponse {
  user_id: string;
  phone_country_code: string;
  phone_number: string;
  display_name: string;
}

// 9. 定義 Staff 錢包流水
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

// 10. 定義 Staff 積分發放結果
export interface StaffWalletGrantResponse {
  charge: PointsChargeResponse;
  target_user: StaffWalletUserResponse;
  operator: StaffWalletUserResponse;
}

// 11. 定義 Staff 廣告任務
export interface StaffRewardAdResponse {
  task_id: string;
  title: string;
  summary: string;
  cover_url: string;
  media_url: string;
  media_type: 'image' | 'video';
  target_url: string;
  reward_points: number;
  watch_seconds: number;
  total_budget: number;
  total_granted: number;
  remaining_budget: number;
  is_active: boolean;
  starts_at?: string;
  ends_at?: string;
  created_at: string;
  updated_at: string;
}

// 12. 定義 Staff 廣告任務保存資料
export interface StaffRewardAdPayload {
  title?: string;
  summary?: string;
  cover_url?: string;
  media_url?: string;
  media_type?: 'image' | 'video';
  target_url?: string;
  reward_points?: number;
  watch_seconds?: number;
  total_budget?: number;
  is_active?: boolean;
  starts_at?: string;
  ends_at?: string;
  retention_days?: number;
}
