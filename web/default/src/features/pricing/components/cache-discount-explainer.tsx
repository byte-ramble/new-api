import { memo, useMemo } from 'react'
import { Sparkles, ZapOff, TrendingDown } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import type { PricingModel } from '../types'

export interface CacheDiscountExplainerProps {
  models: PricingModel[] | undefined
  className?: string
}

/**
 * CacheDiscountExplainer — surfaces the prompt-cache savings that PackyAPI's
 * documentation talks about but never visualizes for users.
 *
 * For the public model square, this is the single most differentiating data
 * point: "your real cost is 50-75% less than direct because we forward upstream
 * cache reads at a fraction of the price." This component computes the actual
 * numbers from the live model catalog (rather than copy-pasting marketing
 * claims) so what you see matches what you get.
 *
 * Hidden when no models in the catalog support caching (e.g. fresh install).
 */
export const CacheDiscountExplainer = memo(function CacheDiscountExplainer({
  models,
  className,
}: CacheDiscountExplainerProps) {
  const { t } = useTranslation()

  const stats = useMemo(() => {
    if (!models || models.length === 0) return null
    const cached = models.filter(
      (m) => typeof m.cache_ratio === 'number' && m.cache_ratio !== null && m.cache_ratio > 0 && m.cache_ratio < 1
    )
    if (cached.length === 0) return null
    const discounts = cached.map((m) => 1 - (m.cache_ratio as number))
    const max = Math.max(...discounts)
    const avg = discounts.reduce((a, b) => a + b, 0) / discounts.length
    return {
      modelCount: cached.length,
      maxPct: Math.round(max * 100), // best case savings, e.g. 90 for cache_ratio=0.10
      avgPct: Math.round(avg * 100), // average savings across cache-enabled models
    }
  }, [models])

  if (!stats) return null

  return (
    <section
      aria-label={t('Prompt cache savings')}
      className={cn(
        'relative overflow-hidden rounded-2xl border bg-gradient-to-br from-violet-50 via-white to-cyan-50 p-5 shadow-sm dark:from-violet-950/30 dark:via-background dark:to-cyan-950/30',
        className
      )}
    >
      {/* Decorative gradient orbs */}
      <div
        aria-hidden
        className='pointer-events-none absolute -right-20 -top-20 h-56 w-56 rounded-full bg-violet-300/40 blur-3xl dark:bg-violet-600/20'
      />
      <div
        aria-hidden
        className='pointer-events-none absolute -bottom-20 -left-20 h-56 w-56 rounded-full bg-cyan-300/40 blur-3xl dark:bg-cyan-600/20'
      />

      <div className='relative grid gap-5 sm:grid-cols-[1fr_auto] sm:items-center'>
        {/* Left: pitch */}
        <div className='min-w-0'>
          <div className='inline-flex items-center gap-1.5 rounded-full bg-violet-500/10 px-2.5 py-1 text-xs font-medium text-violet-700 dark:bg-violet-500/20 dark:text-violet-300'>
            <Sparkles className='h-3.5 w-3.5' />
            {t('Prompt cache savings')}
          </div>
          <h3 className='mt-3 text-lg font-bold leading-tight sm:text-xl'>
            {t('Repeat prompts cost a fraction of the original')}
          </h3>
          <p className='text-muted-foreground mt-1.5 text-sm leading-relaxed sm:mt-2'>
            {t(
              'For long-context tools (Claude Code / Codex / Cursor) where your project context repeats every call, cache hits are charged at the upstream cache-read rate — typically a small fraction of the standard price.'
            )}
          </p>
        </div>

        {/* Right: stat trio */}
        <div className='flex flex-col gap-3 sm:flex-row sm:gap-4'>
          <Stat
            icon={<TrendingDown className='h-4 w-4' />}
            label={t('Up to')}
            value={`${stats.maxPct}%`}
            valueClass='text-violet-600 dark:text-violet-400'
            sub={t('off on cache hit')}
          />
          <Stat
            icon={<ZapOff className='h-4 w-4' />}
            label={t('Average')}
            value={`${stats.avgPct}%`}
            valueClass='text-cyan-600 dark:text-cyan-400'
            sub={t('across {{count}} models', { count: stats.modelCount })}
          />
        </div>
      </div>
    </section>
  )
})

interface StatProps {
  icon: React.ReactNode
  label: string
  value: string
  sub: string
  valueClass?: string
}
function Stat({ icon, label, value, sub, valueClass }: StatProps) {
  return (
    <div className='min-w-[140px] rounded-xl border bg-background/80 px-4 py-3 backdrop-blur-sm'>
      <div className='text-muted-foreground flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide'>
        {icon}
        {label}
      </div>
      <div className={cn('mt-1 text-2xl font-extrabold leading-none', valueClass)}>
        {value}
      </div>
      <div className='text-muted-foreground mt-0.5 text-xs'>{sub}</div>
    </div>
  )
}
