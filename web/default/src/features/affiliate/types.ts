// Types mirror the JSON returned by the OmniRouter affiliate REST API.
// See controller/affiliate.go for the canonical shape.

export interface CommissionLog {
  id: number
  earner_id: number
  source_user_id: number
  level: 1 | 2
  topup_amount_rmb: number
  commission_rmb: number
  rate_pct: number
  status: 'paid' | 'frozen' | 'reversed'
  ref_id?: number | null
  topup_channel: string
  note?: string
  created_at: number
}

export interface AffiliateSettings {
  level1_rate_pct: number
  level2_rate_pct: number
  min_withdrawal_rmb: number
  withdrawal_fee_rmb: number
  max_daily_commission_rmb: number
}

export type AffiliateOverviewDisabled = {
  enabled: false
}

export type AffiliateOverviewEnabled = {
  enabled: true
  balance_rmb: number
  total_earned_rmb: number
  total_withdrawn_rmb: number
  last_earned_at: number
  recent_commissions: CommissionLog[]
  pending_withdrawals: number
  settings: AffiliateSettings
}

export type AffiliateOverview = AffiliateOverviewDisabled | AffiliateOverviewEnabled

export interface Withdrawal {
  id: number
  user_id: number
  amount_rmb: number
  fee_rmb: number
  net_rmb: number
  method: 'alipay' | 'wechat' | 'bank' | 'balance'
  account: string
  status: 'pending' | 'approved' | 'rejected' | 'reversed'
  user_note?: string
  admin_note?: string
  processed_by?: number
  processed_at?: number
  created_at: number
  updated_at: number
}

export interface PagedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface WithdrawalRequest {
  amount_rmb: number
  method: Withdrawal['method']
  account: string
  user_note?: string
}
