import { memo, useMemo } from 'react'
import { ArrowRight, Check } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { EXCLUDED_GROUPS } from '../constants'

export interface GroupShowcaseProps {
  /** From /api/pricing usable_group: { groupKey: { desc, ratio } } */
  usableGroup: Record<string, { desc?: string; ratio?: number }> | undefined
  /** Currently selected group filter (URL-driven) */
  currentGroup?: string
  /** Click handler — wired to setGroupFilter */
  onGroupSelect: (group: string) => void
  className?: string
}

/**
 * Tier classification by group_ratio:
 *
 *   ratio >= 1.0      → "premium"     blue   (官方原价 / 最稳)
 *   0.7  <= ratio < 1 → "standard"    green  (轻折扣 / 备线)
 *   ratio < 0.7       → "deep"        orange (深折扣 / 性价比, 吸睛)
 *   ratio missing     → "premium"     fallback
 *
 * Color tokens follow the project's existing dark/light Tailwind palette
 * (oklch() based) — see model-card.tsx for reference.
 */
function tierFromRatio(ratio: number | undefined) {
  const r = ratio ?? 1
  if (r < 0.7) return 'deep' as const
  if (r < 1.0) return 'standard' as const
  return 'premium' as const
}

const TIER_STYLE = {
  premium: {
    badge:
      'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
    ring: 'ring-blue-200 dark:ring-blue-800',
    accent: 'text-blue-600 dark:text-blue-400',
  },
  standard: {
    badge:
      'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
    ring: 'ring-emerald-200 dark:ring-emerald-800',
    accent: 'text-emerald-600 dark:text-emerald-400',
  },
  deep: {
    badge:
      'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300',
    ring: 'ring-orange-200 dark:ring-orange-800',
    accent: 'text-orange-600 dark:text-orange-400',
  },
}

/**
 * Format a group_ratio like 0.5 / 0.85 / 1.0 into a human discount label:
 *   1.0  → t('Standard')   ("原价" in zh)
 *   0.85 → "8.5 折"        / "15% off"
 *   0.5  → "5 折"          / "50% off"
 */
function formatDiscount(ratio: number | undefined, lang: string): string {
  const r = ratio ?? 1
  if (r >= 1) {
    return lang.startsWith('zh') ? '原价' : 'Standard'
  }
  // zh:  N 折   = N×10% of original (10折 = 100%)
  // en:  N% off = (1 - r) × 100
  if (lang.startsWith('zh')) {
    const tenths = Math.round(r * 10 * 10) / 10 // 0.85 → 8.5
    return `${tenths} 折`
  }
  const off = Math.round((1 - r) * 100)
  return `${off}% off`
}

export const GroupShowcase = memo(function GroupShowcase({
  usableGroup,
  currentGroup,
  onGroupSelect,
  className,
}: GroupShowcaseProps) {
  const { t, i18n } = useTranslation()

  const groups = useMemo(() => {
    if (!usableGroup) return []
    return Object.entries(usableGroup)
      .filter(([key]) => !EXCLUDED_GROUPS.includes(key))
      .map(([key, info]) => ({
        key,
        desc: info?.desc ?? key,
        ratio: info?.ratio ?? 1,
        tier: tierFromRatio(info?.ratio),
      }))
      // Sort: cheapest first (most attention-grabbing surfaces left)
      .sort((a, b) => a.ratio - b.ratio)
  }, [usableGroup])

  if (groups.length === 0) return null

  return (
    <section
      aria-label={t('Available pricing groups')}
      className={cn('w-full', className)}
    >
      {/* Section header */}
      <div className='mb-3 flex items-end justify-between gap-3 px-1 sm:mb-4'>
        <div>
          <h2 className='text-base font-semibold tracking-tight sm:text-lg'>
            {t('Choose your pricing group')}
          </h2>
          <p className='text-muted-foreground mt-0.5 text-xs sm:text-sm'>
            {t(
              'Same models, different upstreams — pick the channel that matches your priority.'
            )}
          </p>
        </div>
        {currentGroup ? (
          <button
            type='button'
            onClick={() => onGroupSelect('')}
            className='text-muted-foreground hover:text-foreground text-xs underline-offset-4 hover:underline'
          >
            {t('Show all groups')}
          </button>
        ) : null}
      </div>

      {/* Scroll container — horizontal on small screens, wraps on >sm */}
      <div className='-mx-1 flex snap-x snap-mandatory gap-3 overflow-x-auto px-1 pb-2 sm:flex-wrap sm:overflow-visible'>
        {groups.map((g) => {
          const isActive = currentGroup === g.key
          const style = TIER_STYLE[g.tier]
          return (
            <button
              key={g.key}
              type='button'
              onClick={() => onGroupSelect(g.key)}
              aria-pressed={isActive}
              className={cn(
                'group relative flex min-w-[260px] flex-1 snap-start flex-col rounded-2xl border bg-card p-4 text-left transition-all sm:min-w-[280px] sm:flex-none sm:basis-[calc(33.333%-0.5rem)] xl:basis-[calc(25%-0.5625rem)]',
                'hover:-translate-y-0.5 hover:shadow-lg',
                isActive
                  ? cn('ring-2 shadow-md', style.ring)
                  : 'ring-1 ring-transparent'
              )}
            >
              {/* Top: tier badge + active checkmark */}
              <div className='mb-3 flex items-start justify-between gap-2'>
                <span
                  className={cn(
                    'rounded-full px-2.5 py-1 text-xs font-semibold',
                    style.badge
                  )}
                >
                  {formatDiscount(g.ratio, i18n.language)}
                </span>
                {isActive ? (
                  <span
                    className={cn(
                      'flex h-6 w-6 items-center justify-center rounded-full',
                      style.badge
                    )}
                    aria-label={t('Selected')}
                  >
                    <Check className='h-3.5 w-3.5' />
                  </span>
                ) : null}
              </div>

              {/* Group key + description */}
              <div className='flex-1'>
                <div className='font-mono text-xs text-muted-foreground'>
                  {g.key}
                </div>
                <div className='mt-1 line-clamp-2 text-base font-semibold leading-tight sm:text-lg'>
                  {g.desc}
                </div>
              </div>

              {/* CTA */}
              <div
                className={cn(
                  'mt-4 flex items-center gap-1 text-sm font-medium opacity-0 transition-opacity group-hover:opacity-100',
                  style.accent,
                  isActive && 'opacity-100'
                )}
              >
                {isActive ? t('Currently filtering') : t('Use this group')}
                {!isActive && (
                  <ArrowRight className='h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5' />
                )}
              </div>
            </button>
          )
        })}
      </div>
    </section>
  )
})
