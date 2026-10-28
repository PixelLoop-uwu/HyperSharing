import { create } from 'zustand'
import { createDraft, publishDraft } from '../api/posts'
import { uploadFileViaTus } from '../api/tus-upload'
import { buildDraft } from '../lib/build-draft'
import { useDraftStore } from './useDraftStore'
import type { PostRead } from './types'

export type PublishStage =
  | 'idle'
  | 'drafting'
  | 'uploading'
  | 'publishing'

interface PublishState {
  stage: PublishStage
  progress: { done: number; total: number } | null
  error: string | null

  start: () => Promise<PostRead | null>
  cancel: () => void
}

let controller: AbortController | null = null

export const usePublishStore = create<PublishState>()(
  (set, get) => ({
    stage: 'idle',
    progress: null,
    error: null,

    async start() {
      if (get().stage !== 'idle') return null

      const { title, sections } = useDraftStore.getState()
      const built = buildDraft(title, sections)

      // buildDraft уже удалил пустые секции и обрезал лишние пробелы.
      if (built.payload.sections.length === 0) {
        set({ error: 'Добавьте хотя бы одну секцию' })
        return null
      }

      const ctrl = new AbortController()
      controller = ctrl

      const checkpoint = () => {
        if (ctrl.signal.aborted) {
          throw ctrl.signal.reason ??
            new DOMException('Cancelled', 'AbortError')
        }
      }

      set({
        stage: 'drafting',
        progress: null,
        error: null,
      })

      try {
        const draft = await createDraft(built.payload)
        checkpoint()

        if (built.files.length > 0) {
          // Предполагается, что сервер сохраняет порядок объявленных файлов.
          const fileIds = draft.sections.flatMap(
            (section) => section.files.map((file) => file.id),
          )

          if (fileIds.length !== built.files.length) {
            throw new Error(
              'Сервер вернул некорректный список файлов черновика',
            )
          }

          const totalBytes = built.files.reduce(
            (total, file) => total + file.size,
            0,
          )

          let doneBytes = 0

          set({
            stage: 'uploading',
            progress: { done: 0, total: totalBytes },
          })

          for (let i = 0; i < built.files.length; i++) {
            const file = built.files[i]
            const fileId = fileIds[i]

            if (!file || !fileId) {
              throw new Error(
                'Не удалось сопоставить файл со слотом загрузки',
              )
            }

            await uploadFileViaTus(
              file,
              {
                draftId: draft.id,
                fileId,
              },
              (uploaded) => {
                // Не создаём прогресс 0/0 для пустых файлов.
                if (totalBytes === 0) return

                set({
                  progress: {
                    done: doneBytes + uploaded,
                    total: totalBytes,
                  },
                })
              },
              ctrl.signal,
            )

            doneBytes += file.size
            checkpoint()
          }
        }

        checkpoint()
        set({ stage: 'publishing' })

        const post = await publishDraft(draft.id)

        useDraftStore.getState().reset()

        set({
          stage: 'idle',
          progress: null,
          error: null,
        })

        return post
      } catch (err) {
        const cancelled = ctrl.signal.aborted

        set({
          stage: 'idle',
          progress: null,
          error: cancelled
            ? null
            : err instanceof Error
              ? err.message
              : 'Не удалось опубликовать пост',
        })

        return null
      } finally {
        if (controller === ctrl) {
          controller = null
        }
      }
    },

    cancel() {
      const stage = get().stage

      // После начала публикации отмена ненадёжна:
      // сервер уже может обрабатывать запрос.
      if (stage === 'idle' || stage === 'publishing') return

      controller?.abort()
    },
  }),
)

// Ошибка публикации устаревает только при изменении содержимого,
// а не при любом обновлении состояния редактора.
useDraftStore.subscribe((state, prev) => {
  const draftChanged =
    state.title !== prev.title ||
    state.sections !== prev.sections

  if (!draftChanged) return

  if (usePublishStore.getState().error !== null) {
    usePublishStore.setState({ error: null })
  }
})