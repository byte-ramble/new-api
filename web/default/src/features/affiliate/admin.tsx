import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, X, Clock } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { SectionPageLayout } from '@/components/layout'
import { useAdminWithdrawals, useAdminReviewWithdrawal } from './hooks'
import type { Withdrawal } from './types'

const PAGE_SIZE = 50

type StatusFilter = 'pending' | 'approved' | 'rejected' | 'reversed' | 'all'

export function AffiliateAdmin() {
  const { t } = useTranslation()
  const [status, setStatus] = useState<StatusFilter>('pending')
  const [page, setPage] = useState(1)
  const q = useAdminWithdrawals(status, page, PAGE_SIZE)
  const review = useAdminReviewWithdrawal()

  const [activeRow, setActiveRow] = useState<Withdrawal | null>(null)
  const [activeAction, setActiveAction] = useState<'approve' | 'reject'>('approve')
  const [note, setNote] = useState('')

  const openReview = (row: Withdrawal, action: 'approve' | 'reject') => {
    setActiveRow(row)
    setActiveAction(action)
    setNote('')
  }
  const closeReview = () => {
    setActiveRow(null)
    setNote('')
  }
  const confirmReview = () => {
    if (!activeRow) return
    review.mutate(
      { id: activeRow.id, action: activeAction, note },
      {
        onSettled: () => closeReview(),
      },
    )
  }

  return (
    <SectionPageLayout>
      <Card>
        <CardHeader className='flex-row items-center justify-between gap-3 space-y-0'>
          <CardTitle>{t('Withdrawal review queue')}</CardTitle>
          <Select
            value={status}
            onValueChange={(v) => {
              setStatus(v as StatusFilter)
              setPage(1)
            }}
          >
            <SelectTrigger className='w-[180px]'>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='pending'>
                <Clock className='mr-1.5 inline h-3.5 w-3.5' />
                {t('Pending')}
              </SelectItem>
              <SelectItem value='approved'>{t('Approved')}</SelectItem>
              <SelectItem value='rejected'>{t('Rejected')}</SelectItem>
              <SelectItem value='reversed'>{t('Reversed')}</SelectItem>
              <SelectItem value='all'>{t('All')}</SelectItem>
            </SelectContent>
          </Select>
        </CardHeader>
        <CardContent className='p-0'>
          {q.isLoading ? (
            <Skeleton className='h-64 w-full' />
          ) : !q.data || q.data.items.length === 0 ? (
            <div className='py-16 text-center text-muted-foreground'>
              {t('No withdrawals in this state')}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Submitted')}</TableHead>
                  <TableHead>{t('User')}</TableHead>
                  <TableHead>{t('Amount')}</TableHead>
                  <TableHead>{t('Method')}</TableHead>
                  <TableHead>{t('Account')}</TableHead>
                  <TableHead>{t('User note')}</TableHead>
                  <TableHead>{t('Status')}</TableHead>
                  <TableHead className='text-right'>{t('Action')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {q.data.items.map((row) => (
                  <TableRow key={row.id}>
                    <TableCell className='whitespace-nowrap text-xs text-muted-foreground'>
                      {fmtDate(row.created_at)}
                    </TableCell>
                    <TableCell className='font-mono text-xs'>#{row.user_id}</TableCell>
                    <TableCell className='font-semibold tabular-nums'>
                      {fmtRmb(row.amount_rmb)}
                      {row.fee_rmb > 0 && (
                        <span className='ml-1 text-xs text-muted-foreground'>
                          (-{fmtRmb(row.fee_rmb)})
                        </span>
                      )}
                    </TableCell>
                    <TableCell>{row.method}</TableCell>
                    <TableCell className='max-w-[200px] truncate font-mono text-xs'>
                      {row.account}
                    </TableCell>
                    <TableCell className='max-w-[180px] truncate text-xs text-muted-foreground'>
                      {row.user_note || '—'}
                    </TableCell>
                    <TableCell>
                      <Badge variant='outline'>{row.status}</Badge>
                    </TableCell>
                    <TableCell className='text-right'>
                      {row.status === 'pending' ? (
                        <div className='flex justify-end gap-1.5'>
                          <Button
                            size='sm'
                            variant='outline'
                            onClick={() => openReview(row, 'approve')}
                          >
                            <Check className='mr-1 h-3.5 w-3.5' />
                            {t('Approve')}
                          </Button>
                          <Button
                            size='sm'
                            variant='outline'
                            onClick={() => openReview(row, 'reject')}
                          >
                            <X className='mr-1 h-3.5 w-3.5' />
                            {t('Reject')}
                          </Button>
                        </div>
                      ) : (
                        <span className='text-xs text-muted-foreground'>
                          {row.processed_by ? `by #${row.processed_by}` : ''}
                        </span>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={!!activeRow} onOpenChange={(open) => !open && closeReview()}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {activeAction === 'approve'
                ? t('Approve withdrawal')
                : t('Reject withdrawal')}
            </DialogTitle>
          </DialogHeader>
          {activeRow && (
            <div className='space-y-3 text-sm'>
              <div className='grid grid-cols-2 gap-2 rounded-lg bg-muted/30 p-3'>
                <div>
                  <div className='text-xs text-muted-foreground'>{t('User')}</div>
                  <div className='font-mono'>#{activeRow.user_id}</div>
                </div>
                <div>
                  <div className='text-xs text-muted-foreground'>{t('Amount')}</div>
                  <div className='font-semibold'>{fmtRmb(activeRow.amount_rmb)}</div>
                </div>
                <div>
                  <div className='text-xs text-muted-foreground'>{t('Method')}</div>
                  <div>{activeRow.method}</div>
                </div>
                <div>
                  <div className='text-xs text-muted-foreground'>{t('Account')}</div>
                  <div className='truncate font-mono text-xs'>{activeRow.account}</div>
                </div>
              </div>
              <div className='space-y-1.5'>
                <Label htmlFor='admin-note'>
                  {activeAction === 'approve'
                    ? t('Note (e.g. payout transaction id)')
                    : t('Reason (visible to user)')}
                </Label>
                <Input
                  id='admin-note'
                  value={note}
                  onChange={(e) => setNote(e.target.value)}
                  maxLength={200}
                />
              </div>
              {activeAction === 'reject' && (
                <p className='text-xs text-amber-600'>
                  {t(
                    "Rejecting will refund the locked amount back to the user's affiliate balance.",
                  )}
                </p>
              )}
            </div>
          )}
          <DialogFooter>
            <Button variant='outline' onClick={closeReview} disabled={review.isPending}>
              {t('Cancel')}
            </Button>
            <Button onClick={confirmReview} disabled={review.isPending}>
              {review.isPending
                ? t('Processing...')
                : activeAction === 'approve'
                  ? t('Confirm approve')
                  : t('Confirm reject')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SectionPageLayout>
  )
}

function fmtRmb(v: number): string {
  return Number.isFinite(v) ? `¥${v.toFixed(2)}` : '—'
}
function fmtDate(unix: number): string {
  if (!unix) return '—'
  return new Date(unix * 1000).toLocaleString()
}
