// Pure pricing calculator math — no React, no DOM.
//
// Reuses the same per-1M-token cost model as features/pricing/lib/price.ts:
//   base_per_M_usd = model_ratio * 2 * group_ratio
//   input  = base
//   output = base * completion_ratio
//   cache  = base * cache_ratio          (only when cache_ratio set)
//
// "* 2" comes from QuotaPerUnit = 500_000 (server constant): a model_ratio of
// 1.0 corresponds to $2 per 1M tokens at default group_ratio.

import type { PricingModel } from '../types'

export interface CalculatorInput {
  model: PricingModel
  groupRatio: number
  /** Average prompt tokens per day */
  dailyPromptTokens: number
  /** Average completion tokens per day */
  dailyCompletionTokens: number
  /** Cache hit rate 0..1 — fraction of prompt tokens that hit prompt cache */
  cacheHitRate: number
  /** How many days the calculation covers (default 30 = monthly) */
  days?: number
}

export interface CalculatorResult {
  /** Total cost on OmniRouter (USD), accounting for group ratio + cache */
  omnirouterUsd: number
  /** Total cost at direct upstream pricing (USD), no group discount, no cache */
  directUsd: number
  /** Difference (direct - omnirouter), always >= 0 in normal cases */
  savingsUsd: number
  /** Savings as percentage of direct cost, 0..100 */
  savingsPct: number
  /** Whether cache pricing was actually applied (model has cache_ratio set) */
  cacheApplied: boolean
}

const TOKENS_PER_M = 1_000_000

/**
 * Compute monthly cost on OmniRouter and the direct-pricing comparison.
 *
 * Assumptions:
 *   - Model is token-based (quota_type === 0). Per-request models return
 *     all-zero result; caller should hide them from the model picker.
 *   - Cache hit rate applies only to prompt tokens (output is always
 *     billed at full output price).
 *   - When the model has no cache_ratio, cache_hit_rate is silently
 *     ignored — same as upstream behavior.
 */
export function computeMonthlyCost(input: CalculatorInput): CalculatorResult {
  const {
    model,
    groupRatio,
    dailyPromptTokens,
    dailyCompletionTokens,
    cacheHitRate,
    days = 30,
  } = input

  // Per-1M-token unit prices in USD
  const baseDirect = model.model_ratio * 2 * 1.0
  const baseDiscounted = model.model_ratio * 2 * groupRatio

  const directInputPerM = baseDirect
  const directOutputPerM = baseDirect * model.completion_ratio

  const orInputPerM = baseDiscounted
  const orOutputPerM = baseDiscounted * model.completion_ratio
  const cacheRatio =
    typeof model.cache_ratio === 'number' && model.cache_ratio !== null
      ? model.cache_ratio
      : null
  const orCachePerM =
    cacheRatio !== null ? baseDiscounted * cacheRatio : orInputPerM

  // Volumes in millions of tokens
  const promptM = (dailyPromptTokens * days) / TOKENS_PER_M
  const completionM = (dailyCompletionTokens * days) / TOKENS_PER_M

  // OmniRouter cost — split prompt by cache hit rate
  const cacheApplied = cacheRatio !== null && cacheHitRate > 0
  const cachedPromptM = cacheApplied ? promptM * cacheHitRate : 0
  const uncachedPromptM = promptM - cachedPromptM

  const orPromptCost = uncachedPromptM * orInputPerM + cachedPromptM * orCachePerM
  const orCompletionCost = completionM * orOutputPerM
  const omnirouterUsd = orPromptCost + orCompletionCost

  // Direct cost — no discount, no cache
  const directUsd = promptM * directInputPerM + completionM * directOutputPerM

  const savingsUsd = Math.max(0, directUsd - omnirouterUsd)
  const savingsPct = directUsd > 0 ? (savingsUsd / directUsd) * 100 : 0

  return {
    omnirouterUsd,
    directUsd,
    savingsUsd,
    savingsPct,
    cacheApplied,
  }
}

/**
 * Format a USD amount for display in the calculator result panel.
 * Uses 2 decimals for >= $1, 4 for smaller — keeps both "$45.20" and
 * "$0.0125" readable.
 */
export function formatUsd(value: number): string {
  if (!Number.isFinite(value)) return '—'
  if (value === 0) return '$0'
  if (value < 1) return `$${value.toFixed(4)}`
  if (value < 100) return `$${value.toFixed(2)}`
  return `$${value.toFixed(0)}`
}
