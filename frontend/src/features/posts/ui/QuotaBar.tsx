import { formatBytes } from '@/shared/lib/format'
import type { QuotaMetric } from '../types'

export default function QuotaBar({ label, metric }: { label: string; metric: QuotaMetric }) {
  const pct = metric.limit > 0 ? Math.min(100, (metric.used / metric.limit) * 100) : 0
  // Ненулевое использование не должно выглядеть как «пусто»
  const width = metric.used > 0 ? Math.max(pct, 1.5) : 0
  const tone = pct >= 90 ? 'bg-red-400' : pct >= 75 ? 'bg-amber-400' : 'bg-zinc-100'
  const text = `${formatBytes(metric.used)} / ${formatBytes(metric.limit)}`

  return (
    <div className="grid grid-cols-[1fr_auto] items-baseline gap-x-3 gap-y-1.5 text-xs">
      <dt className="text-zinc-400">{label}</dt>
      <dd className="tabular-nums text-zinc-300">{text}</dd>
      <dd className="col-span-2">
        <div
          role="progressbar"
          aria-label={label}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={Math.round(pct)}
          aria-valuetext={text}
          className="h-1.5 overflow-hidden rounded-full bg-zinc-700/50"
        >
          <div className={`h-full rounded-full transition-all ${tone}`} style={{ width: `${width}%` }} />
        </div>
      </dd>
    </div>
  )
}