import { motion } from 'framer-motion'
import { useMemo } from 'react'
import { FiSend } from 'react-icons/fi'
import { formatBytes } from '@/shared/lib/format'
import Card from '@/shared/ui/layout/Card'
import GhostButton from '@/shared/ui/actions/GhostButton'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'
import StatusMark from '@/shared/ui/feedback/StatusMark'
import { isBlank } from '../lib/sections'
import { useDraftStore } from '../models/useDraftStore'
import { usePublishStore } from '../models/usePublishStore'

function Row({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex items-baseline justify-between gap-3">
      <dt className="text-zinc-400">{label}</dt>
      <dd className="tabular-nums text-zinc-100">{value}</dd>
    </div>
  )
}

export default function PublishPanel({ onPublish }: { onPublish: () => void }) {
  const sections = useDraftStore((s) => s.sections)
  const stage = usePublishStore((s) => s.stage)
  const progress = usePublishStore((s) => s.progress)
  const error = usePublishStore((s) => s.error)
  const cancel = usePublishStore((s) => s.cancel)

  const stats = useMemo(() => {
    const files = sections.flatMap((s) => s.files)
    return {
      sections: sections.filter((s) => !isBlank(s)).length,
      files: files.length,
      size: files.reduce((n, f) => n + f.file.size, 0),
    }
  }, [sections])

  const busy = stage !== 'idle'
  const percent = progress && progress.total > 0 ? (progress.done / progress.total) * 100 : 0
  const stageLabel =
    stage === 'drafting'
      ? 'Готовим пост'
      : stage === 'uploading'
        ? `Загружаем файлы: ${formatBytes(progress?.done ?? 0)} из ${formatBytes(progress?.total ?? 0)}`
        : 'Публикуем'

  return (
    <Card variant="solid" className="p-5 lg:sticky lg:top-[calc(var(--header-total-h)+1.5rem)]">
      <h2 className="font-display text-base font-semibold text-zinc-100">Публикация</h2>

      <dl className="mt-4 space-y-2 text-sm">
        <Row label="Секции" value={stats.sections} />
        <Row label="Файлы" value={stats.files} />
        <Row label="Размер" value={formatBytes(stats.size)} />
      </dl>

      {error && (
        <p role="alert" className="mt-4 text-sm text-red-400">
          {error}
        </p>
      )}

      <div className="mt-5">
        {busy ? (
          <div aria-live="polite">
            <p className="flex items-center gap-2 text-sm text-zinc-300">
              <StatusMark status="running" size={16} />
              {stageLabel}
            </p>
            {stage === 'uploading' && (
              <div
                role="progressbar"
                aria-label="Загрузка файлов"
                aria-valuemin={0}
                aria-valuemax={100}
                aria-valuenow={Math.round(percent)}
                className="mt-3 h-1 overflow-hidden rounded-full bg-zinc-800"
              >
                <motion.div
                  className="h-full rounded-full bg-zinc-100"
                  animate={{ width: `${percent}%` }}
                  transition={{ duration: 0.3, ease: [0.23, 1, 0.32, 1] }}
                />
              </div>
            )}
            {stage !== 'publishing' && (
              <GhostButton onClick={cancel} className="mt-3 w-full justify-center">
                Отменить
              </GhostButton>
            )}
          </div>
        ) : (
          <PrimaryButton
            className="w-full justify-center"
            icon={<FiSend className="size-4" />}
            onClick={onPublish}
          >
            Опубликовать
          </PrimaryButton>
        )}
      </div>
    </Card>
  )
}
