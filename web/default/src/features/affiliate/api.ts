// API client for the OmniRouter affiliate program.
//
// Uses the project's existing axios instance (from @/lib/api). All endpoints
// already wrap responses in {success, message, data}, so we unwrap into the
// expected payload here.

import { api } from '@/lib/api'
import type {
  AffiliateOverview,
  CommissionLog,
  PagedResponse,
  Withdrawal,
  WithdrawalRequest,
} from './types'

interface ApiEnvelope<T> {
  success: boolean
  message?: string
  data?: T
}

function unwrap<T>(env: ApiEnvelope<T>): T {
  if (!env.success) {
    throw new Error(env.message || 'request failed')
  }
  return env.data as T
}

// ---- User endpoints ----

export async function fetchOverview(): Promise<AffiliateOverview> {
  const r = await api.get<ApiEnvelope<AffiliateOverview>>('/api/affiliate/overview')
  return unwrap(r.data)
}

export async function fetchCommissionLog(
  page = 1,
  pageSize = 20
): Promise<PagedResponse<CommissionLog>> {
  const r = await api.get<ApiEnvelope<PagedResponse<CommissionLog>>>(
    '/api/affiliate/log',
    { params: { page, page_size: pageSize } },
  )
  return unwrap(r.data)
}

export async function fetchWithdrawalsHistory(
  page = 1,
  pageSize = 20
): Promise<PagedResponse<Withdrawal>> {
  const r = await api.get<ApiEnvelope<PagedResponse<Withdrawal>>>(
    '/api/affiliate/withdrawals',
    { params: { page, page_size: pageSize } },
  )
  return unwrap(r.data)
}

export async function submitWithdrawal(req: WithdrawalRequest): Promise<Withdrawal> {
  const r = await api.post<ApiEnvelope<Withdrawal>>('/api/affiliate/withdrawal', req)
  return unwrap(r.data)
}

// ---- Admin endpoints ----

export async function adminFetchWithdrawals(
  status: 'pending' | 'approved' | 'rejected' | 'reversed' | 'all' = 'pending',
  page = 1,
  pageSize = 50
): Promise<PagedResponse<Withdrawal> & { status: string }> {
  const r = await api.get<
    ApiEnvelope<PagedResponse<Withdrawal> & { status: string }>
  >('/api/affiliate/admin/withdrawals', {
    params: { status, page, page_size: pageSize },
  })
  return unwrap(r.data)
}

export async function adminApproveWithdrawal(id: number, note: string): Promise<void> {
  const r = await api.post<ApiEnvelope<void>>(
    `/api/affiliate/admin/withdrawals/${id}/approve`,
    { note },
  )
  unwrap(r.data)
}

export async function adminRejectWithdrawal(id: number, note: string): Promise<void> {
  const r = await api.post<ApiEnvelope<void>>(
    `/api/affiliate/admin/withdrawals/${id}/reject`,
    { note },
  )
  unwrap(r.data)
}
