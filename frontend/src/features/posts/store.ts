import { create } from 'zustand'
import { createDraft, publishDraft } from './api'
import * as ops from './lib'
import type { DraftData, FileTarget, PostRead } from './types'

// --- Draft editing state ---

interface DraftState extends DraftData {
  /** False until the saved draft has been loaded; the editor should wait for it. */
  hydrated: boolean

  setTitle: (title: string) => void
  addSection: (at?: number) => void
  setText: (sectionId: string, text: string) => void
  addFiles: (files: File[], target?: FileTarget) => void
  removeFile: (sectionId: string, fileId: string) => void
  removeSection: (sectionId: string) => void
  moveSection: (from: number, to: number) => void
  hydrate: (data: DraftData) => void
  reset: () => void
}

export const useDraftStore = create<DraftState>()((set) => ({
  title: '',
  sections: [],
  hydrated: false,

  setTitle: (title) => set({ title }),

  addSection: (at) =>
    set((s) => ({ sections: ops.insertSection(s.sections, ops.makeSection(), at) })),

  setText: (sectionId, text) => set((s) => ({ sections: ops.setText(s.sections, sectionId, text) })),

  addFiles: (files, target = { newAt: Number.POSITIVE_INFINITY }) =>
    set((s) => ({ sections: ops.addFiles(s.sections, ops.makeRefs(files), target) })),

  removeFile: (sectionId, fileId) =>
    set((s) => ({ sections: ops.removeFile(s.sections, sectionId, fileId) })),

  removeSection: (sectionId) => set((s) => ({ sections: ops.removeSection(s.sections, sectionId) })),

  moveSection: (from, to) => set((s) => ({ sections: ops.moveSection(s.sections, from, to) })),

  hydrate: (data) => set({ ...data, hydrated: true }),

  reset: () => set({ title: '', sections: [] }),
}))

// --- Publish flow ---

export type PublishStage = 'idle' | 'drafting' | 'uploading' | 'publishing'

interface PublishState {
  stage: PublishStage
  /** Upload progress in bytes. */
  progress: { done: number; total: number } | null
  error: string | null
  /** Runs the whole flow; resolves to the created post, or null on failure or cancel. */
  start: () => Promise<PostRead | null>
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
    const built = ops.buildDraft(title, sections)
    if (built.files.length === 0 && built.payload.title === '') {
      set({ error: 'Добавьте текст или файл' })
      return null
    }

    const ctrl = new AbortController()
    controller = ctrl
    const checkpoint = () => {
      if (ctrl.signal.aborted) throw new Cancelled()
    }

    set({ stage: 'drafting', progress: null, error: null })
    try {
      const draft = await createDraft(built.payload)
      checkpoint()

      if (built.files.length > 0) {
        // The response lists slot ids in the same order as the declared files.
        const fileIds = draft.sections.flatMap((s) => s.files.map((f) => f.id))
        const totalBytes = built.files.reduce((n, f) => n + f.size, 0)
        let doneBytes = 0
        set({ stage: 'uploading', progress: { done: 0, total: totalBytes } })

        for (let i = 0; i < built.files.length; i++) {
          const file = built.files[i]
          await ops.uploadFileViaTus(
            file,
            { draftId: draft.id, fileId: fileIds[i] },
            (uploaded) => set({ progress: { done: doneBytes + uploaded, total: totalBytes } }),
            ctrl.signal,
          )
          doneBytes += file.size
          checkpoint()
        }
      }

      set({ stage: 'publishing' })
      const post = await publishDraft(draft.id)
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
