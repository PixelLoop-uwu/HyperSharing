import { motion } from 'framer-motion'
import { memo } from 'react'
import { FiX } from 'react-icons/fi'
import { formatBytes } from '@/shared/lib/format'
import FileTypeIcon from '@/shared/ui/content/FileTypeIcon'
import GhostButton from '@/shared/ui/actions/GhostButton'
import type { FileRef } from '../models/types'

interface FileRowProps {
  file: FileRef
  onRemove: (fileId: string) => void
}

function FileRow({ file, onRemove }: FileRowProps) {
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
        <GhostButton
          iconOnly
          className="-mr-1.5"
          aria-label={`Убрать ${file.file.name}`}
          onClick={() => onRemove(file.id)}
          icon={<FiX className="size-4" />}
        />
      </div>
    </motion.li>
  )
}

export default memo(FileRow)
