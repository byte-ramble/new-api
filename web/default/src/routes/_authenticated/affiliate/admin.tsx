import { createFileRoute } from '@tanstack/react-router'
import { AffiliateAdmin } from '@/features/affiliate/admin'

export const Route = createFileRoute('/_authenticated/affiliate/admin')({
  component: AffiliateAdmin,
})
