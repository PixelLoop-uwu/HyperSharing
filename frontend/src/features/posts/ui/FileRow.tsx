import { motion } from 'framer-motion'
import { memo } from 'react'
import { FiX } from 'react-icons/fi'
import { formatBytes } from '@/shared/lib/format'
import FileTypeIcon from '@/shared/ui/content/FileTypeIcon'
import GhostButton from '@/shared/ui/actions/GhostButton'
import StatusMark, { type StatusMarkStatus } from '@/shared/ui/feedback/StatusMark'
import { useFileDigest, type FileRef } from '../model'

interface FileRowProps {
  file: FileRef
  onRemove: (fileId: string) => void
}

function FileRow({ file, onRemove }: FileRowProps) {
  const digest = useFileDigest(file.id)
  const status: StatusMarkStatus =
    !digest || digest.status === 'pending' ? 'running' : digest.status === 'ready' ? 'done' : 'failed'

  return (
    <motion.li
      initial={{ opacity: 0, height: 0 }}
      animate={{ opacity: 1, height: 'auto' }}
      exit={{ opacity: 0, height: 0 }}
      transition={{ duration: 0.2, ease: [0.23, 1, 0.32, 1] }}
      className="overflow-hidden"
    >
      <div className="flex items-center gap-3 px-3.5 py-2.5">
        <FileTypeIcon
          type={file.file.type}
          name={file.file.name}
          className="size-4 shrink-0 text-zinc-300"
        />
        <span className="min-w-0 flex-1 truncate text-sm text-zinc-100" title={file.file.name}>
          {file.file.name}
        </span>
        <span className="shrink-0 text-xs tabular-nums text-zinc-400">
          {formatBytes(file.file.size)}
        </span>
        <StatusMark status={status} size={16} />
        <GhostButton
          iconOnly
          className="-mr-1.5"
          aria-label={`Убрать ${file.file.name}`}
          onClick={() => onRemove(file.id)}
          icon={<FiX className="size-4" />}
        />
      </div>
      {digest?.status === 'error' && (
        <p className="px-3.5 pb-2.5 text-xs text-red-400">
          Не удалось обработать файл. Уберите его и добавьте снова.
        </p>
      )}
    </motion.li>
  )
}

export default memo(FileRow)
