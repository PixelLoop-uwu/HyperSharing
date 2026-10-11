import { create } from 'zustand'
import { createDraft, publishPost } from '../api'
import { uploadFileParts } from '../lib'
import type { FileCompleteIn, PostFullRead } from '../types'
import { buildItems } from './buildItems'
import { useDigests } from './digests'
import { useDraftStore } from './draft'

export type PublishStage = 'idle' | 'drafting' | 'uploading' | 'publishing'

interface PublishState {
  stage: PublishStage
  /** Upload progress in parts, not bytes. */
  progress: { done: number; total: number } | null
  error: string | null
  /** Runs the whole flow; resolves to the created post, or null on failure or cancel. */
  start: () => Promise<PostFullRead | null>
  /** Aborts between files. Mid-file abort needs `signal` support in uploadFileParts. */
  cancel: () => void
}

class Cancelled extends Error {}

let controller: AbortController | null = null

export const usePublishStore = create<PublishState>()((set, get) => ({
  stage: 'idle',
  progress: null,
  error: null,

  async start() {
    if (get().stage !== 'idle') return null

    const { title, sections } = useDraftStore.getState()
    const built = buildItems(sections, useDigests.getState().byFileId)
    if (!built.ok) {
      const message =
        built.reason === 'empty'
          ? 'Добавьте текст или файл'
          : built.reason === 'hashing'
            ? 'Файлы ещё обрабатываются, подождите немного'
            : `Не удалось обработать «${built.fileName ?? 'файл'}». Уберите его и добавьте снова`
      set({ error: message })
      return null
    }

    const ctrl = new AbortController()
    controller = ctrl
    const checkpoint = () => {
      if (ctrl.signal.aborted) throw new Cancelled()
    }

    set({ stage: 'drafting', progress: null, error: null })
    try {
      const draft = await createDraft({ title: title.trim() || null, items: built.items })
      checkpoint()

      const complete: FileCompleteIn[] = []
      if (draft.files.length > 0) {
        const total = draft.files.reduce((n, f) => n + f.parts.length, 0)
        let doneParts = 0
        set({ stage: 'uploading', progress: { done: 0, total } })

        for (const upload of draft.files) {
          const file = built.uploads.get(upload.hash)
          if (!file) throw new Error('Сервер запросил файл, которого нет в посте')
          const parts = await uploadFileParts(file, upload, (done) =>
            set({ progress: { done: doneParts + done, total } }),
          )
          checkpoint()
          complete.push({ hash: upload.hash, parts })
          doneParts += upload.parts.length
        }
      }

      set({ stage: 'publishing' })
      const post = await publishPost({ draft_id: draft.draft_id, files: complete })
      useDraftStore.getState().reset()
      set({ stage: 'idle', progress: null })
      return post
    } catch (err) {
      set({
        stage: 'idle',
        progress: null,
        error:
          err instanceof Cancelled
            ? null
            : err instanceof Error
              ? err.message
              : 'Не удалось опубликовать пост',
      })
      return null
    } finally {
      if (controller === ctrl) controller = null
    }
  },

  cancel() {
    controller?.abort()
  },
}))

// Any edit makes the previous error stale.
useDraftStore.subscribe(() => {
  if (usePublishStore.getState().error) usePublishStore.setState({ error: null })
})
