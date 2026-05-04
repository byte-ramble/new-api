import { Helmet } from 'react-helmet-async'
import { useEffect, useState } from 'react'
import { getStatus } from '@/lib/api'

interface SeoHeadProps {
  /** Number of models currently exposed — drives the "N models available" copy */
  modelCount?: number
  /** Optional explicit title override (defaults to "Model Square — {SystemName}") */
  title?: string
  /** Optional explicit description override */
  description?: string
}

/**
 * SeoHead — page-scoped <head> management for the public Model Square page.
 *
 * Why: the static index.html keeps the legacy "New API" title (Rule 5 — must
 * not be removed). For SEO and social sharing on this page specifically, we
 * inject runtime <title>, <meta>, and Open Graph tags that reflect the actual
 * brand (system_name from /api/status) and current model count.
 *
 * Note: this only affects browsers and crawlers that execute JavaScript.
 * For full SEO indexing (no-JS crawlers, social-link unfurling), this needs
 * to be paired with prerender/SSG — a follow-up task.
 */
export function SeoHead({ modelCount, title, description }: SeoHeadProps) {
  const [systemName, setSystemName] = useState<string>(() => {
    // Initial state from cached status (synchronous — avoids title flash)
    if (typeof window === 'undefined') return ''
    try {
      const cached = localStorage.getItem('status')
      if (cached) {
        const parsed = JSON.parse(cached)
        return parsed?.system_name ?? ''
      }
    } catch {
      /* ignore */
    }
    return ''
  })

  useEffect(() => {
    // Background refresh — keeps title in sync if admin updates SystemName
    getStatus()
      .then((s) => {
        if (s?.system_name) setSystemName(s.system_name as string)
      })
      .catch(() => {
        /* ignore */
      })
  }, [])

  const brand = systemName || 'OmniRouter'
  const finalTitle =
    title ??
    (modelCount
      ? `Model Square — ${modelCount}+ AI Models · ${brand}`
      : `Model Square · ${brand}`)
  const finalDesc =
    description ??
    `Browse and compare ${
      modelCount ? `${modelCount}+ ` : ''
    }AI models on ${brand}. Real-time pricing, group ratios, cache discounts. OpenAI / Anthropic / Gemini protocol-compatible.`

  const url = typeof window !== 'undefined' ? window.location.href : ''

  return (
    <Helmet>
      <title>{finalTitle}</title>
      <meta name='description' content={finalDesc} />
      {url ? <link rel='canonical' href={url} /> : null}

      {/* Open Graph (Facebook, LinkedIn, generic link previews) */}
      <meta property='og:type' content='website' />
      <meta property='og:title' content={finalTitle} />
      <meta property='og:description' content={finalDesc} />
      {url ? <meta property='og:url' content={url} /> : null}
      <meta property='og:site_name' content={brand} />

      {/* Twitter Card */}
      <meta name='twitter:card' content='summary_large_image' />
      <meta name='twitter:title' content={finalTitle} />
      <meta name='twitter:description' content={finalDesc} />
    </Helmet>
  )
}
