import { useEffect, useMemo, useState } from 'react'
import { ExternalLink, Copy, Check, Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { tryPrettyJson } from '@/lib/utils'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { completeClaudeCodeOAuth, startClaudeCodeOAuth } from '../../api'

type ClaudeCodeOAuthDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onKeyGenerated: (key: string) => void
}

export function ClaudeCodeOAuthDialog({
  open,
  onOpenChange,
  onKeyGenerated,
}: ClaudeCodeOAuthDialogProps) {
  const { t } = useTranslation()
  const { copiedText, copyToClipboard } = useCopyToClipboard({ notify: false })

  const [state, setState] = useState({
    authorizeUrl: '',
    callbackUrl: '',
    isStarting: false,
    isCompleting: false,
  })

  useEffect(() => {
    if (!open) {
      setState({
        authorizeUrl: '',
        callbackUrl: '',
        isStarting: false,
        isCompleting: false,
      })
    }
  }, [open])

  const canCopyAuthorizeUrl = Boolean(state.authorizeUrl && !state.isStarting)
  const canComplete = useMemo(
    () => Boolean(state.callbackUrl.trim()) && !state.isCompleting,
    [state.callbackUrl, state.isCompleting]
  )

  const handleStart = async () => {
    setState((prev) => ({ ...prev, isStarting: true }))
    try {
      const res = await startClaudeCodeOAuth()
      if (!res.success) {
        throw new Error(res.message || 'Failed to start OAuth')
      }

      const url = res.data?.authorize_url || ''
      if (!url) {
        throw new Error('Missing authorize_url in response')
      }

      setState((prev) => ({ ...prev, authorizeUrl: url }))
      try {
        window.open(url, '_blank', 'noopener,noreferrer')
        toast.success(t('Opened authorization page'))
      } catch (error) {
        // eslint-disable-next-line no-console
        console.warn('Failed to open authorization page:', error)
        toast.warning(t('Please manually copy and open the authorization link'))
      }
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : t('OAuth start failed')
      )
    } finally {
      setState((prev) => ({ ...prev, isStarting: false }))
    }
  }

  const handleComplete = async () => {
    if (!state.callbackUrl.trim()) return
    setState((prev) => ({ ...prev, isCompleting: true }))
    try {
      const res = await completeClaudeCodeOAuth(state.callbackUrl.trim())
      if (!res.success) {
        throw new Error(res.message || 'OAuth failed')
      }

      const rawKey = res.data?.key || ''
      if (!rawKey) {
        throw new Error('Missing key in response')
      }

      onKeyGenerated(tryPrettyJson(rawKey))
      toast.success(t('Credential generated'))
      onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('OAuth failed'))
    } finally {
      setState((prev) => ({ ...prev, isCompleting: false }))
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>{t('Claude Code Authorization')}</DialogTitle>
          <DialogDescription>
            {t(
              'Generate a Claude Code OAuth credential and paste it into the channel key field.'
            )}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4'>
          <Alert>
            <AlertDescription>
              {t(
                '1) Click "Open authorization page" and log in to your Claude.ai account. 2) After approving, the page will display an authorization code (or redirect to the Anthropic callback page). 3) Copy that code (format like "code#state") or the full callback URL and paste it below. 4) Click "Generate credential".'
              )}
            </AlertDescription>
          </Alert>

          <div className='flex flex-wrap gap-2'>
            <Button onClick={handleStart} disabled={state.isStarting}>
              {state.isStarting ? (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              ) : (
                <ExternalLink className='mr-2 h-4 w-4' />
              )}
              {t('Open authorization page')}
            </Button>

            <Button
              type='button'
              variant='outline'
              disabled={!canCopyAuthorizeUrl}
              onClick={async () => {
                if (!state.authorizeUrl) return
                await copyToClipboard(state.authorizeUrl)
              }}
              aria-label={t('Copy authorization link')}
              title={t('Copy authorization link')}
            >
              {copiedText === state.authorizeUrl ? (
                <Check className='mr-2 h-4 w-4 text-green-600' />
              ) : (
                <Copy className='mr-2 h-4 w-4' />
              )}
              {t('Copy authorization link')}
            </Button>
          </div>

          <div className='space-y-2'>
            <div className='text-sm font-medium'>
              {t('Authorization code or callback URL')}
            </div>
            <Input
              value={state.callbackUrl}
              onChange={(e) =>
                setState((prev) => ({ ...prev, callbackUrl: e.target.value }))
              }
              placeholder={t(
                'Paste the code (format "code#state") or the full callback URL'
              )}
              autoComplete='off'
              spellCheck={false}
            />
            <div className='text-muted-foreground text-xs'>
              {t(
                'Tip: The generated key is a JSON credential including access_token / refresh_token; it will be auto-refreshed every 10 minutes if expiring within 24h.'
              )}
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button
            type='button'
            variant='outline'
            onClick={() => onOpenChange(false)}
            disabled={state.isStarting || state.isCompleting}
          >
            {t('Cancel')}
          </Button>
          <Button onClick={handleComplete} disabled={!canComplete}>
            {state.isCompleting && (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            )}
            {state.isCompleting ? t('Generating...') : t('Generate credential')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
