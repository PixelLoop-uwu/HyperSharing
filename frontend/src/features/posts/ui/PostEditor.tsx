import { AnimatePresence, motion, MotionConfig } from 'framer-motion'
import { useEffect, useRef } from 'react'
import { useFileDragActive } from '@/shared/lib/fileDrag'
import { rise, stagger } from '@/shared/lib/motion'
import StatusMark from '@/shared/ui/feedback/StatusMark'
import { initPostEditor, useDraftStore, usePublishStore } from '../model'
import type { PostFullRead } from '../types'
import NewSectionZone from './NewSectionZone'
import PublishPanel from './PublishPanel'
import SectionCard from './SectionCard'

interface PostEditorProps {
  /** Called with the created post after a successful publish. */
  onPublished: (post: PostFullRead) => void
}

export default function PostEditor({ onPublished }: PostEditorProps) {
  const hydrated = useDraftStore((s) => s.hydrated)
  const title = useDraftStore((s) => s.title)
  const setTitle = useDraftStore((s) => s.setTitle)
  const sections = useDraftStore((s) => s.sections)
  const busy = usePublishStore((s) => s.stage !== 'idle')
  const dragActive = useFileDragActive() && !busy

  useEffect(() => {
    void initPostEditor()
  }, [])

  // While publishing, the draft is frozen: edits would be lost when it is reset on success.
  const editorRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (editorRef.current) editorRef.current.inert = busy
  }, [busy, hydrated])

  async function handlePublish() {
    const post = await usePublishStore.getState().start()
    if (post) onPublished(post)
  }

  if (!hydrated) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
        className="flex min-h-[40svh] items-center justify-center gap-3 text-sm text-zinc-400"
      >
        <StatusMark status="running" size={16} />
        Загружаем черновик
      </motion.div>
    )
  }

  return (
    <MotionConfig reducedMotion="user">
      <motion.div
        variants={stagger}
        initial="hidden"
        animate="show"
        className="grid gap-8 pb-24 pt-10 lg:grid-cols-[minmax(0,42rem)_17rem] lg:items-start lg:justify-center lg:gap-12"
      >
        <motion.div variants={rise} className="min-w-0">
        <div
          ref={editorRef}
          aria-busy={busy}
          className={`transition-opacity duration-200 ${busy ? 'opacity-60' : ''}`}
        >
          <h1 className="sr-only">Новый пост</h1>
          <motion.input
            variants={rise}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Заголовок поста"
            aria-label="Заголовок поста"
            className="w-full border-b border-zinc-800 bg-transparent pb-3 font-display text-3xl font-semibold tracking-tight text-zinc-100 outline-none transition-colors placeholder:text-zinc-600 focus:border-zinc-500"
          />

          <ul className={sections.length > 0 ? 'mt-8 space-y-3' : undefined}>
            <AnimatePresence initial={false}>
              {sections.map((section) => (
                <SectionCard key={section.id} section={section} dragActive={dragActive} />
              ))}
            </AnimatePresence>
          </ul>

          <NewSectionZone empty={sections.length === 0} dragActive={dragActive} />
        </div>
      </motion.div>

        <motion.div variants={rise} className="mt-10">
          <PublishPanel onPublish={handlePublish} />
        </motion.div>
      </motion.div>
    </MotionConfig>
  )
}
