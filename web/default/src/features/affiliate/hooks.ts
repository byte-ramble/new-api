import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import {
  fetchOverview,
  fetchCommissionLog,
  fetchWithdrawalsHistory,
  submitWithdrawal,
  adminFetchWithdrawals,
  adminApproveWithdrawal,
  adminRejectWithdrawal,
} from './api'
import type { WithdrawalRequest } from './types'

const KEYS = {
  overview: ['affiliate', 'overview'] as const,
  log: (page: number, pageSize: number) => ['affiliate', 'log', page, pageSize] as const,
  withdrawals: (page: number, pageSize: number) =>
    ['affiliate', 'withdrawals', page, pageSize] as const,
  adminWithdrawals: (status: string, page: number, pageSize: number) =>
    ['affiliate', 'admin', 'withdrawals', status, page, pageSize] as const,
}

export function useAffiliateOverview() {
  return useQuery({
    queryKey: KEYS.overview,
    queryFn: fetchOverview,
    staleTime: 30_000, // 30s — overview is read often during interactive use
  })
}

export function useCommissionLog(page: number, pageSize: number) {
  return useQuery({
    queryKey: KEYS.log(page, pageSize),
    queryFn: () => fetchCommissionLog(page, pageSize),
    placeholderData: (prev) => prev,
  })
}

export function useWithdrawalsHistory(page: number, pageSize: number) {
  return useQuery({
    queryKey: KEYS.withdrawals(page, pageSize),
    queryFn: () => fetchWithdrawalsHistory(page, pageSize),
    placeholderData: (prev) => prev,
  })
}

export function useSubmitWithdrawal() {
  const qc = useQueryClient()
  const { t } = useTranslation()
  return useMutation({
    mutationFn: (req: WithdrawalRequest) => submitWithdrawal(req),
    onSuccess: () => {
      toast.success(t('Withdrawal request submitted'))
      // Invalidate overview + history so the new request shows up immediately
      qc.invalidateQueries({ queryKey: ['affiliate'] })
    },
    onError: (err: Error) => {
      toast.error(err.message || t('Withdrawal request failed'))
    },
  })
}

export function useAdminWithdrawals(
  status: 'pending' | 'approved' | 'rejected' | 'reversed' | 'all',
  page: number,
  pageSize: number,
) {
  return useQuery({
    queryKey: KEYS.adminWithdrawals(status, page, pageSize),
    queryFn: () => adminFetchWithdrawals(status, page, pageSize),
    placeholderData: (prev) => prev,
  })
}

export function useAdminReviewWithdrawal() {
  const qc = useQueryClient()
  const { t } = useTranslation()
  return useMutation({
    mutationFn: async (args: {
      id: number
      action: 'approve' | 'reject'
      note: string
    }) => {
      if (args.action === 'approve') {
        await adminApproveWithdrawal(args.id, args.note)
      } else {
        await adminRejectWithdrawal(args.id, args.note)
      }
    },
    onSuccess: (_, vars) => {
      toast.success(
        vars.action === 'approve'
          ? t('Withdrawal approved')
          : t('Withdrawal rejected'),
      )
      qc.invalidateQueries({ queryKey: ['affiliate', 'admin'] })
    },
    onError: (err: Error) => {
      toast.error(err.message || t('Action failed'))
    },
  })
}
