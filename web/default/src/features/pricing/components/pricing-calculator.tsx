import { memo, useEffect, useMemo, useState } from 'react'
import { Calculator, ChevronDown, ChevronUp, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
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
import { cn } from '@/lib/utils'
import { EXCLUDED_GROUPS, QUOTA_TYPE_VALUES } from '../constants'
import { computeMonthlyCost, formatUsd } from '../lib/calculator'
import type { PricingModel } from '../types'

export interface PricingCalculatorProps {
  models: PricingModel[] | undefined
  usableGroup: Record<string, { desc?: string; ratio?: number }> | undefined
  className?: string
}

const DEFAULT_PROMPT = 10_000
const DEFAULT_COMPLETION = 5_000
const DEFAULT_CACHE_PCT = 0

/**
 * PricingCalculator — interactive "what would I pay per month?" widget.
 *
 * Surfaces the actual savings vs direct upstream pricing using the live
 * catalog. The math comes from features/pricing/lib/calculator.ts so this
 * component is a thin shell over inputs + a result panel.
 *
 * Hidden when no token-based models exist (fresh install). Collapsed by
 * default — the page's primary job is browsing, calc is opt-in.
 */
export const PricingCalculator = memo(function PricingCalculator({
  models,
  usableGroup,
  className,
}: PricingCalculatorProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  // Filter to token-based models (per-request models don't apply)
  const tokenModels = useMemo(
    () =>
      (models || []).filter(
        (m) => m.quota_type === QUOTA_TYPE_VALUES.TOKEN
      ),
    [models]
  )

  const groups = useMemo(() => {
    if (!usableGroup) return [] as { key: string; ratio: number; desc: string }[]
    return Object.entries(usableGroup)
      .filter(([key]) => !EXCLUDED_GROUPS.includes(key))
      .map(([key, info]) => ({
        key,
        ratio: info?.ratio ?? 1,
        desc: info?.desc ?? key,
      }))
      .sort((a, b) => a.ratio - b.ratio)
  }, [usableGroup])

  // Inputs
  const [modelName, setModelName] = useState<string>('')
  const [groupKey, setGroupKey] = useState<string>('')
  const [dailyPrompt, setDailyPrompt] = useState<number>(DEFAULT_PROMPT)
  const [dailyCompletion, setDailyCompletion] = useState<number>(DEFAULT_COMPLETION)
  const [cachePct, setCachePct] = useState<number>(DEFAULT_CACHE_PCT)

  // Auto-pick defaults once data arrives
  useEffect(() => {
    if (!modelName && tokenModels.length > 0) {
      setModelName(tokenModels[0].model_name)
    }
  }, [tokenModels, modelName])
  useEffect(() => {
    if (!groupKey && groups.length > 0) {
      // Default to cheapest group — most flattering "you save N%" first impression
      setGroupKey(groups[0].key)
    }
  }, [groups, groupKey])

  const selectedModel = useMemo(
    () => tokenModels.find((m) => m.model_name === modelName),
    [tokenModels, modelName]
  )
  const selectedGroup = useMemo(
    () => groups.find((g) => g.key === groupKey),
    [groups, groupKey]
  )

  const result = useMemo(() => {
    if (!selectedModel || !selectedGroup) return null
    return computeMonthlyCost({
      model: selectedModel,
      groupRatio: selectedGroup.ratio,
      dailyPromptTokens: Math.max(0, dailyPrompt),
      dailyCompletionTokens: Math.max(0, dailyCompletion),
      cacheHitRate: Math.min(1, Math.max(0, cachePct / 100)),
    })
  }, [selectedModel, selectedGroup, dailyPrompt, dailyCompletion, cachePct])

  // Hidden state: no usable models means the calc would show nonsense
  if (tokenModels.length === 0 || groups.length === 0) return null

  return (
    <section
      aria-label={t('Pricing calculator')}
      className={cn(
        'overflow-hidden rounded-2xl border bg-card shadow-sm',
        className
      )}
    >
      {/* Header — always visible, click to expand */}
      <button
        type='button'
        onClick={() => setOpen((v) => !v)}
        className='flex w-full items-center justify-between gap-3 px-5 py-4 text-left hover:bg-muted/40'
        aria-expanded={open}
      >
        <div className='flex items-center gap-3'>
          <div className='flex h-10 w-10 items-center justify-center rounded-full bg-emerald-100 text-emerald-600 dark:bg-emerald-900/40 dark:text-emerald-300'>
            <Calculator className='h-5 w-5' />
          </div>
          <div>
            <div className='text-base font-semibold'>
              {t('How much will I pay per month?')}
            </div>
            <div className='text-muted-foreground text-xs sm:text-sm'>
              {t('Plug in your daily token usage to compare OmniRouter vs direct upstream pricing.')}
            </div>
          </div>
        </div>
        {open ? (
          <ChevronUp className='text-muted-foreground h-5 w-5 shrink-0' />
        ) : (
          <ChevronDown className='text-muted-foreground h-5 w-5 shrink-0' />
        )}
      </button>

      {open ? (
        <div className='border-t px-5 py-5'>
          <div className='grid gap-5 lg:grid-cols-[1fr_auto] lg:items-center'>
            {/* Inputs */}
            <div className='grid gap-4 sm:grid-cols-2'>
              <div className='space-y-1.5'>
                <Label htmlFor='calc-model'>{t('Model')}</Label>
                <Select value={modelName} onValueChange={setModelName}>
                  <SelectTrigger id='calc-model'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className='max-h-72'>
                    {tokenModels.map((m) => (
                      <SelectItem key={m.model_name} value={m.model_name}>
                        <span className='font-mono text-sm'>{m.model_name}</span>
                        {m.vendor_name ? (
                          <span className='text-muted-foreground ml-2 text-xs'>
                            {m.vendor_name}
                          </span>
                        ) : null}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className='space-y-1.5'>
                <Label htmlFor='calc-group'>{t('Group')}</Label>
                <Select value={groupKey} onValueChange={setGroupKey}>
                  <SelectTrigger id='calc-group'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {groups.map((g) => (
                      <SelectItem key={g.key} value={g.key}>
                        <span>{g.desc}</span>
                        <span className='text-muted-foreground ml-2 font-mono text-xs'>
                          ×{g.ratio}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className='space-y-1.5'>
                <Label htmlFor='calc-prompt'>{t('Daily prompt tokens')}</Label>
                <Input
                  id='calc-prompt'
                  type='number'
                  inputMode='numeric'
                  min={0}
                  value={dailyPrompt}
                  onChange={(e) => setDailyPrompt(Number(e.target.value) || 0)}
                />
              </div>

              <div className='space-y-1.5'>
                <Label htmlFor='calc-completion'>
                  {t('Daily completion tokens')}
                </Label>
                <Input
                  id='calc-completion'
                  type='number'
                  inputMode='numeric'
                  min={0}
                  value={dailyCompletion}
                  onChange={(e) =>
                    setDailyCompletion(Number(e.target.value) || 0)
                  }
                />
              </div>

              <div className='space-y-1.5 sm:col-span-2'>
                <div className='flex items-center justify-between'>
                  <Label htmlFor='calc-cache'>{t('Cache hit rate')}</Label>
                  <span className='text-muted-foreground text-xs font-mono'>
                    {cachePct}%
                  </span>
                </div>
                <input
                  id='calc-cache'
                  type='range'
                  min={0}
                  max={90}
                  step={5}
                  value={cachePct}
                  onChange={(e) => setCachePct(Number(e.target.value))}
                  className='w-full accent-violet-500'
                />
                {result && !result.cacheApplied ? (
                  <p className='text-muted-foreground text-xs'>
                    {t('Selected model has no cache pricing — slider is informational.')}
                  </p>
                ) : null}
              </div>
            </div>

            {/* Result panel */}
            {result ? (
              <div className='rounded-xl border bg-gradient-to-br from-emerald-50 to-cyan-50 p-5 dark:from-emerald-950/30 dark:to-cyan-950/30 lg:min-w-[260px]'>
                <div className='text-muted-foreground text-xs font-medium uppercase tracking-wide'>
                  {t('Monthly cost on OmniRouter')}
                </div>
                <div className='mt-1 text-3xl font-extrabold tracking-tight text-emerald-600 dark:text-emerald-400'>
                  {formatUsd(result.omnirouterUsd)}
                </div>
                <div className='text-muted-foreground mt-2 flex items-baseline gap-2 text-xs'>
                  <span>{t('vs direct')}</span>
                  <span className='font-mono line-through opacity-70'>
                    {formatUsd(result.directUsd)}
                  </span>
                </div>
                {result.savingsPct >= 1 ? (
                  <div className='mt-3 inline-flex items-center gap-1.5 rounded-full bg-emerald-500/15 px-3 py-1 text-xs font-semibold text-emerald-700 dark:text-emerald-300'>
                    <Sparkles className='h-3.5 w-3.5' />
                    {t('Saves {{amount}} ({{pct}}%)', {
                      amount: formatUsd(result.savingsUsd),
                      pct: result.savingsPct.toFixed(0),
                    })}
                  </div>
                ) : null}
              </div>
            ) : null}
          </div>

          {/* Reset */}
          <div className='mt-4 flex justify-end gap-2'>
            <Button
              variant='ghost'
              size='sm'
              onClick={() => {
                setDailyPrompt(DEFAULT_PROMPT)
                setDailyCompletion(DEFAULT_COMPLETION)
                setCachePct(DEFAULT_CACHE_PCT)
              }}
            >
              {t('Reset')}
            </Button>
          </div>
        </div>
      ) : null}
    </section>
  )
})
