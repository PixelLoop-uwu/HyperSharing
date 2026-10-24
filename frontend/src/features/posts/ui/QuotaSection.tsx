import Skeleton from '@/shared/ui/feedback/Skeleton'
import { useQuota } from '../hooks/useQuota'
import QuotaBar from './QuotaBar'

export default function QuotaSection() {
  const quota = useQuota()

  if (quota.isLoading) {
    return (
      <div className="mt-6 space-y-4 border-t border-zinc-800 pt-5" aria-busy="true">
        <Skeleton className="h-6 w-full" />
        <Skeleton className="h-6 w-full" />
      </div>
    )
  }
  if (!quota.data) return null
  return (
    <dl className="mt-6 space-y-4 border-t border-zinc-800 pt-5">
      <QuotaBar label="Хранилище" metric={quota.data.storage} />
      <QuotaBar label="Загружено сегодня" metric={quota.data.daily} />
    </dl>
  )
}
