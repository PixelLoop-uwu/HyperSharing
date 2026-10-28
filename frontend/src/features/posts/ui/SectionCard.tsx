import { AnimatePresence, motion } from 'framer-motion'
import { memo, useCallback, useState } from 'react'
import { FiPaperclip, FiPlus, FiTrash2, FiType } from 'react-icons/fi'
import { useFileDropZone } from '@/shared/lib/fileDrag'
import { useFilePicker } from '@/shared/lib/useFilePicker'
import AutoTextarea from '@/shared/ui/AutoTextarea'
import Card from '@/shared/ui/layout/Card'
import GhostButton from '@/shared/ui/actions/GhostButton'
import { useDraftStore } from '../models/useDraftStore'
import type { Section } from '../models/types'
import FileRow from './FileRow'

interface SectionCardProps {
  section: Section
  /** True while files from the OS are dragged anywhere over the window. */
  dragActive: boolean
}

function SectionCard({ section, dragActive }: SectionCardProps) {
  const setText = useDraftStore((s) => s.setText)
  const addFiles = useDraftStore((s) => s.addFiles)
  const removeFile = useDraftStore((s) => s.removeFile)
  const removeSection = useDraftStore((s) => s.removeSection)

  const hasFiles = section.files.length > 0
  // A files-only section shows a quiet "add text" button until the user asks for the field.
  const [textOpen, setTextOpen] = useState(false)
  const showText = section.text !== '' || !hasFiles || textOpen

  const attach = useCallback(
    (files: File[]) => addFiles(files, { sectionId: section.id }),
    [addFiles, section.id],
  )
  const { over, handlers } = useFileDropZone(attach)
  const picker = useFilePicker(attach)
  const handleRemoveFile = useCallback(
    (fileId: string) => removeFile(section.id, fileId),
    [removeFile, section.id],
  )

  return (
    <motion.li
      layout="position"
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.98 }}
      transition={{ duration: 0.22, ease: [0.23, 1, 0.32, 1] }}
    >
      <Card
        variant="solid"
        {...handlers}
        className={`relative p-4 transition-shadow ${dragActive && !over ? 'ring-1 ring-zinc-600' : ''}`}
      >
        {showText && (
          <AutoTextarea
            value={section.text}
            onChange={(e) => setText(section.id, e.target.value)}
            onBlur={() => setTextOpen(false)}
            autoFocus={textOpen || (section.text === '' && !hasFiles)}
            placeholder="Напишите текст"
            aria-label="Текст секции"
            className="text-base leading-relaxed text-zinc-100 placeholder:text-zinc-600"
          />
        )}

        {hasFiles && (
          <ul
            className={`${showText ? 'mt-3 ' : ''}divide-y divide-zinc-800 overflow-hidden rounded-xl border border-zinc-800`}
          >
            <AnimatePresence initial={false}>
              {section.files.map((file) => (
                <FileRow key={file.id} file={file} onRemove={handleRemoveFile} />
              ))}
            </AnimatePresence>
          </ul>
        )}

        <div className="mt-3 flex items-center gap-1">
          <GhostButton icon={<FiPaperclip className="size-4" />} onClick={picker.open}>
            Добавить файл
          </GhostButton>
          {!showText && (
            <GhostButton icon={<FiType className="size-4" />} onClick={() => setTextOpen(true)}>
              Добавить текст
            </GhostButton>
          )}
          <GhostButton
            iconOnly
            className="ml-auto"
            aria-label="Удалить секцию"
            onClick={() => removeSection(section.id)}
            icon={<FiTrash2 className="size-4" />}
          />
        </div>
        {picker.input}

        <AnimatePresence>
          {over && (
            <motion.div
              aria-hidden="true"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.12 }}
              className="pointer-events-none absolute inset-0 flex items-center justify-center gap-2 rounded-2xl border border-zinc-300 bg-zinc-950/80 text-sm font-medium text-zinc-100 backdrop-blur-sm"
            >
              <FiPlus className="size-4" />
              Добавить в секцию
            </motion.div>
          )}
        </AnimatePresence>
      </Card>
    </motion.li>
  )
}

export default memo(SectionCard)
