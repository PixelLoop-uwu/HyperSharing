import { FiPaperclip, FiPlus, FiType, FiUpload } from 'react-icons/fi'
import { useFileDropZone } from '@/shared/lib/fileDrag'
import { useFilePicker } from '@/shared/lib/useFilePicker'
import GhostButton from '@/shared/ui/actions/GhostButton'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'
import { useDraftStore } from '../models/useDraftStore'

interface NewSectionZoneProps {
  /** No sections yet: show the larger first-run invitation. */
  empty: boolean
  dragActive: boolean
}

export default function NewSectionZone({ empty, dragActive }: NewSectionZoneProps) {
  const addSection = useDraftStore((s) => s.addSection)
  const addFiles = useDraftStore((s) => s.addFiles)
  // Files dropped here become one new section at the end.
  const { over, handlers } = useFileDropZone((files) => addFiles(files))
  const picker = useFilePicker((files) => addFiles(files))

  if (empty) {
    const tone = over
      ? 'border-zinc-200 bg-zinc-100/10'
      : dragActive
        ? 'border-zinc-500'
        : 'border-zinc-700'
    return (
      <div
        {...handlers}
        className={`mt-8 flex flex-col items-center gap-5 rounded-2xl border border-dashed px-6 py-14 text-center transition-colors ${tone}`}
      >
        <FiUpload aria-hidden="true" className="size-6 text-zinc-400" />
        <div>
          <p className="font-display text-lg font-semibold text-zinc-100">
            {over ? 'Отпустите, чтобы создать секцию' : 'Начните с текста или файлов'}
          </p>
          <p className="mt-1 text-sm text-zinc-400">
            Перетащите файлы сюда, и они станут первой секцией поста.
          </p>
        </div>
        <div className="flex flex-wrap items-center justify-center gap-3">
          <PrimaryButton size="sm" icon={<FiType className="size-4" />} onClick={() => addSection()}>
            Написать текст
          </PrimaryButton>
          <GhostButton icon={<FiPaperclip className="size-4" />} onClick={picker.open}>
            Выбрать файлы
          </GhostButton>
        </div>
        {picker.input}
      </div>
    )
  }

  const tone = over
    ? 'border-zinc-200 bg-zinc-100/10 text-zinc-100'
    : dragActive
      ? 'border-zinc-500 text-zinc-300'
      : 'border-zinc-700 text-zinc-400 hover:border-zinc-500 hover:text-zinc-100'

  return (
    <button
      type="button"
      {...handlers}
      onClick={() => addSection()}
      className={`mt-3 flex w-full items-center justify-center gap-2 rounded-2xl border border-dashed py-4 text-sm font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-zinc-400 ${tone}`}
    >
      <FiPlus className="size-4" />
      {over ? 'Отпустите, чтобы создать секцию' : 'Новая секция'}
    </button>
  )
}
