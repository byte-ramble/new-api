import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Wallet, TrendingUp, Send, History } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { SectionPageLayout } from '@/components/layout'
import {
  useAffiliateOverview,
  useCommissionLog,
  useWithdrawalsHistory,
  useSubmitWithdrawal,
} from './hooks'
import type { CommissionLog, Withdrawal, WithdrawalRequest } from './types'

const PAGE_SIZE = 20

export function Affiliate() {
  const { t } = useTranslation()
  const overview = useAffiliateOverview()

  if (overview.isLoading) {
    return (
      <SectionPageLayout>
        <Skeleton className='h-32 w-full' />
        <Skeleton className='mt-4 h-64 w-full' />
      </SectionPageLayout>
    )
  }

  // Feature-disabled state
  if (overview.data && !overview.data.enabled) {
    return (
      <SectionPageLayout>
        <Card>
          <CardContent className='py-12 text-center'>
            <h2 className='text-lg font-semibold'>
              {t('Affiliate program is currently disabled')}
            </h2>
            <p className='text-muted-foreground mt-2 text-sm'>
              {t('Please contact the operator to enable the affiliate program.')}
            </p>
          </CardContent>
        </Card>
      </SectionPageLayout>
    )
  }

  if (!overview.data || !overview.data.enabled) return null
  const data = overview.data

  return (
    <SectionPageLayout>
      <OverviewCards
        balanceRmb={data.balance_rmb}
        totalEarnedRmb={data.total_earned_rmb}
        totalWithdrawnRmb={data.total_withdrawn_rmb}
        pendingWithdrawals={data.pending_withdrawals}
      />

      <SettingsBanner
        l1={data.settings.level1_rate_pct}
        l2={data.settings.level2_rate_pct}
        minWithdrawalRmb={data.settings.min_withdrawal_rmb}
        feeRmb={data.settings.withdrawal_fee_rmb}
      />

      <Tabs defaultValue='log' className='mt-6'>
        <TabsList>
          <TabsTrigger value='log'>
            <TrendingUp className='mr-1.5 h-4 w-4' />
            {t('Commission log')}
          </TabsTrigger>
          <TabsTrigger value='withdraw'>
            <Send className='mr-1.5 h-4 w-4' />
            {t('Request withdrawal')}
          </TabsTrigger>
          <TabsTrigger value='history'>
            <History className='mr-1.5 h-4 w-4' />
            {t('Withdrawal history')}
          </TabsTrigger>
        </TabsList>

        <TabsContent value='log' className='mt-4'>
          <CommissionLogTable />
        </TabsContent>

        <TabsContent value='withdraw' className='mt-4'>
          <WithdrawalForm
            balanceRmb={data.balance_rmb}
            minRmb={data.settings.min_withdrawal_rmb}
            feeRmb={data.settings.withdrawal_fee_rmb}
          />
        </TabsContent>

        <TabsContent value='history' className='mt-4'>
          <WithdrawalHistoryTable />
        </TabsContent>
      </Tabs>
    </SectionPageLayout>
  )
}

// ─────────────────────────────────────────────────────────────────────
// Overview cards
// ─────────────────────────────────────────────────────────────────────

interface OverviewCardsProps {
  balanceRmb: number
  totalEarnedRmb: number
  totalWithdrawnRmb: number
  pendingWithdrawals: number
}

function OverviewCards({
  balanceRmb,
  totalEarnedRmb,
  totalWithdrawnRmb,
  pendingWithdrawals,
}: OverviewCardsProps) {
  const { t } = useTranslation()
  return (
    <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
      <StatCard
        icon={<Wallet className='h-4 w-4' />}
        label={t('Available balance')}
        value={fmtRmb(balanceRmb)}
        highlight
      />
      <StatCard
        icon={<TrendingUp className='h-4 w-4' />}
        label={t('Total earned')}
        value={fmtRmb(totalEarnedRmb)}
      />
      <StatCard
        icon={<History className='h-4 w-4' />}
        label={t('Total withdrawn')}
        value={fmtRmb(totalWithdrawnRmb)}
      />
      <StatCard
        icon={<Send className='h-4 w-4' />}
        label={t('Pending withdrawals')}
        value={String(pendingWithdrawals)}
      />
    </div>
  )
}

function StatCard({
  icon,
  label,
  value,
  highlight,
}: {
  icon: React.ReactNode
  label: string
  value: string
  highlight?: boolean
}) {
  return (
    <Card className={highlight ? 'border-emerald-200 dark:border-emerald-900' : undefined}>
      <CardHeader className='pb-2'>
        <CardTitle className='text-muted-foreground flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide'>
          {icon}
          {label}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div
          className={`text-2xl font-bold tabular-nums ${
            highlight ? 'text-emerald-600 dark:text-emerald-400' : ''
          }`}
        >
          {value}
        </div>
      </CardContent>
    </Card>
  )
}

function SettingsBanner({
  l1,
  l2,
  minWithdrawalRmb,
  feeRmb,
}: {
  l1: number
  l2: number
  minWithdrawalRmb: number
  feeRmb: number
}) {
  const { t } = useTranslation()
  return (
    <Card className='mt-4 border-dashed bg-muted/30'>
      <CardContent className='py-3 text-sm'>
        <div className='flex flex-wrap items-center gap-x-6 gap-y-1'>
          <span>
            <strong>{t('Level 1')}</strong>: {l1}%
          </span>
          <span>
            <strong>{t('Level 2')}</strong>: {l2}%
          </span>
          <span className='text-muted-foreground'>
            {t('Min withdrawal')}: {fmtRmb(minWithdrawalRmb)} ·{' '}
            {t('Fee')}: {fmtRmb(feeRmb)}
          </span>
        </div>
      </CardContent>
    </Card>
  )
}

// ─────────────────────────────────────────────────────────────────────
// Commission log
// ─────────────────────────────────────────────────────────────────────

function CommissionLogTable() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const q = useCommissionLog(page, PAGE_SIZE)

  if (q.isLoading) return <Skeleton className='h-64 w-full' />
  const data = q.data
  if (!data || data.items.length === 0) {
    return (
      <Card>
        <CardContent className='py-12 text-center text-muted-foreground'>
          {t('No commission events yet. Share your referral link to start earning.')}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardContent className='p-0'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Date')}</TableHead>
              <TableHead>{t('Source user')}</TableHead>
              <TableHead>{t('Level')}</TableHead>
              <TableHead>{t('Topup')}</TableHead>
              <TableHead>{t('Rate')}</TableHead>
              <TableHead className='text-right'>{t('Commission')}</TableHead>
              <TableHead>{t('Status')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.items.map((row) => (
              <TableRow key={row.id}>
                <TableCell className='whitespace-nowrap text-xs text-muted-foreground'>
                  {fmtDate(row.created_at)}
                </TableCell>
                <TableCell className='font-mono text-xs'>#{row.source_user_id}</TableCell>
                <TableCell>L{row.level}</TableCell>
                <TableCell>{fmtRmb(row.topup_amount_rmb)}</TableCell>
                <TableCell className='text-xs'>{row.rate_pct}%</TableCell>
                <TableCell className='text-right font-semibold tabular-nums'>
                  {fmtRmb(row.commission_rmb)}
                </TableCell>
                <TableCell>
                  <StatusBadge status={row.status} />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <Pagination
          page={page}
          total={data.total}
          pageSize={PAGE_SIZE}
          onChange={setPage}
        />
      </CardContent>
    </Card>
  )
}

// ─────────────────────────────────────────────────────────────────────
// Withdrawal form
// ─────────────────────────────────────────────────────────────────────

function WithdrawalForm({
  balanceRmb,
  minRmb,
  feeRmb,
}: {
  balanceRmb: number
  minRmb: number
  feeRmb: number
}) {
  const { t } = useTranslation()
  const [amount, setAmount] = useState<number>(Math.max(minRmb, 0))
  const [method, setMethod] = useState<WithdrawalRequest['method']>('alipay')
  const [account, setAccount] = useState('')
  const [note, setNote] = useState('')
  const submit = useSubmitWithdrawal()

  const net = Math.max(0, amount - feeRmb)
  const overBalance = amount > balanceRmb
  const belowMin = minRmb > 0 && amount < minRmb
  const disabled = !account || overBalance || belowMin || amount <= 0 || submit.isPending

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Request withdrawal')}</CardTitle>
      </CardHeader>
      <CardContent className='space-y-4'>
        <div className='grid gap-4 sm:grid-cols-2'>
          <div className='space-y-1.5'>
            <Label htmlFor='wd-amount'>
              {t('Amount (RMB)')}{' '}
              <span className='text-muted-foreground text-xs'>
                · {t('Available')}: {fmtRmb(balanceRmb)}
              </span>
            </Label>
            <Input
              id='wd-amount'
              type='number'
              inputMode='decimal'
              min={0}
              max={balanceRmb}
              step={0.01}
              value={amount}
              onChange={(e) => setAmount(Number(e.target.value) || 0)}
            />
            {overBalance && (
              <p className='text-xs text-red-600'>{t('Exceeds available balance')}</p>
            )}
            {belowMin && (
              <p className='text-xs text-red-600'>
                {t('Minimum withdrawal is {{amount}}', { amount: fmtRmb(minRmb) })}
              </p>
            )}
          </div>

          <div className='space-y-1.5'>
            <Label htmlFor='wd-method'>{t('Method')}</Label>
            <Select value={method} onValueChange={(v) => setMethod(v as typeof method)}>
              <SelectTrigger id='wd-method'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='alipay'>{t('Alipay')}</SelectItem>
                <SelectItem value='wechat'>{t('WeChat Pay')}</SelectItem>
                <SelectItem value='bank'>{t('Bank transfer')}</SelectItem>
                <SelectItem value='balance'>{t('Convert to OmniRouter balance')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className='space-y-1.5 sm:col-span-2'>
            <Label htmlFor='wd-account'>{t('Receiving account')}</Label>
            <Input
              id='wd-account'
              value={account}
              onChange={(e) => setAccount(e.target.value)}
              placeholder={
                method === 'alipay'
                  ? 'xxx@example.com / phone'
                  : method === 'bank'
                    ? '6225 0888 ... — bank, branch, holder name'
                    : ''
              }
            />
          </div>

          <div className='space-y-1.5 sm:col-span-2'>
            <Label htmlFor='wd-note'>{t('Note (optional)')}</Label>
            <Input
              id='wd-note'
              value={note}
              onChange={(e) => setNote(e.target.value)}
              maxLength={200}
            />
          </div>
        </div>

        <div className='bg-muted/30 rounded-lg p-3 text-sm'>
          <div className='flex justify-between'>
            <span>{t('Amount')}</span>
            <span className='tabular-nums'>{fmtRmb(amount)}</span>
          </div>
          <div className='flex justify-between'>
            <span className='text-muted-foreground'>{t('Fee')}</span>
            <span className='text-muted-foreground tabular-nums'>-{fmtRmb(feeRmb)}</span>
          </div>
          <div className='mt-1 flex justify-between border-t pt-1 font-semibold'>
            <span>{t('Net to receive')}</span>
            <span className='tabular-nums'>{fmtRmb(net)}</span>
          </div>
        </div>

        <div className='flex justify-end'>
          <Button
            disabled={disabled}
            onClick={() =>
              submit.mutate({
                amount_rmb: amount,
                method,
                account,
                user_note: note,
              })
            }
          >
            {submit.isPending ? t('Submitting...') : t('Submit request')}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

// ─────────────────────────────────────────────────────────────────────
// Withdrawal history
// ─────────────────────────────────────────────────────────────────────

function WithdrawalHistoryTable() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const q = useWithdrawalsHistory(page, PAGE_SIZE)

  if (q.isLoading) return <Skeleton className='h-64 w-full' />
  const data = q.data
  if (!data || data.items.length === 0) {
    return (
      <Card>
        <CardContent className='py-12 text-center text-muted-foreground'>
          {t('No withdrawal requests yet.')}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardContent className='p-0'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Submitted')}</TableHead>
              <TableHead>{t('Amount')}</TableHead>
              <TableHead>{t('Method')}</TableHead>
              <TableHead>{t('Account')}</TableHead>
              <TableHead>{t('Status')}</TableHead>
              <TableHead>{t('Admin note')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.items.map((row) => (
              <TableRow key={row.id}>
                <TableCell className='whitespace-nowrap text-xs text-muted-foreground'>
                  {fmtDate(row.created_at)}
                </TableCell>
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
                <TableCell>
                  <StatusBadge status={row.status} />
                </TableCell>
                <TableCell className='max-w-[200px] truncate text-xs text-muted-foreground'>
                  {row.admin_note || '—'}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <Pagination
          page={page}
          total={data.total}
          pageSize={PAGE_SIZE}
          onChange={setPage}
        />
      </CardContent>
    </Card>
  )
}

// ─────────────────────────────────────────────────────────────────────
// Tiny shared widgets
// ─────────────────────────────────────────────────────────────────────

function Pagination({
  page,
  total,
  pageSize,
  onChange,
}: {
  page: number
  total: number
  pageSize: number
  onChange: (p: number) => void
}) {
  const { t } = useTranslation()
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  if (totalPages <= 1) return null
  return (
    <div className='flex items-center justify-between border-t p-3 text-sm'>
      <span className='text-muted-foreground text-xs'>
        {t('Page {{page}} / {{total}} · {{count}} items', {
          page,
          total: totalPages,
          count: total,
        })}
      </span>
      <div className='flex gap-2'>
        <Button size='sm' variant='outline' disabled={page <= 1} onClick={() => onChange(page - 1)}>
          {t('Prev')}
        </Button>
        <Button
          size='sm'
          variant='outline'
          disabled={page >= totalPages}
          onClick={() => onChange(page + 1)}
        >
          {t('Next')}
        </Button>
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: CommissionLog['status'] | Withdrawal['status'] }) {
  const { t } = useTranslation()
  const map: Record<string, { label: string; className: string }> = {
    paid: {
      label: t('Paid'),
      className: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
    },
    pending: {
      label: t('Pending'),
      className: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
    },
    approved: {
      label: t('Approved'),
      className: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
    },
    rejected: {
      label: t('Rejected'),
      className: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300',
    },
    frozen: {
      label: t('Frozen'),
      className: 'bg-slate-100 text-slate-700 dark:bg-slate-900/40 dark:text-slate-300',
    },
    reversed: {
      label: t('Reversed'),
      className: 'bg-slate-100 text-slate-700 dark:bg-slate-900/40 dark:text-slate-300',
    },
  }
  const m = map[status] ?? { label: status, className: '' }
  return (
    <Badge variant='outline' className={m.className}>
      {m.label}
    </Badge>
  )
}

// ─────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────

function fmtRmb(v: number): string {
  if (!Number.isFinite(v)) return '—'
  return `¥${v.toFixed(2)}`
}

function fmtDate(unix: number): string {
  if (!unix) return '—'
  const d = new Date(unix * 1000)
  return d.toLocaleString()
}
