import { idbDelete, idbGet, idbSet } from '@/shared/lib/idb'
import { useDraftStore } from './useDraftStore'
import type { DraftData, FileRef, Section } from '../models/types'

const META_KEY = 'hypersharing:post-draft:v2'
const SAVE_DELAY_MS = 300

const blobKey = (id: string) => `hypersharing:file:${id}`

interface StoredFile {
  id: string
  name: string
  type: string
  lastModified: number
}

interface StoredSection {
  id: string
  text: string
  files: StoredFile[]
}

interface StoredDraft {
  title: string
  sections: StoredSection[]
}

type BlobStatus = 'pending' | 'saved'

const emptyDraft = (): DraftData => ({
  title: '',
  sections: [],
})

/** Записи файлов, которые уже сохранены или сохраняются. */
const blobStatus = new Map<string, BlobStatus>()

let saveTimer: ReturnType<typeof setTimeout> | undefined
let initPromise: Promise<void> | undefined

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function isStoredDraft(value: unknown): value is StoredDraft {
  if (
    !isRecord(value) ||
    typeof value.title !== 'string' ||
    !Array.isArray(value.sections)
  ) {
    return false
  }

  return value.sections.every((section: unknown) => {
    if (
      !isRecord(section) ||
      typeof section.id !== 'string' ||
      typeof section.text !== 'string' ||
      !Array.isArray(section.files)
    ) {
      return false
    }

    return section.files.every((file: unknown) =>
      isRecord(file) &&
      typeof file.id === 'string' &&
      typeof file.name === 'string' &&
      typeof file.type === 'string' &&
      typeof file.lastModified === 'number' &&
      Number.isFinite(file.lastModified),
    )
  })
}

function readMeta(): StoredDraft | null {
  try {
    const raw = localStorage.getItem(META_KEY)

    if (!raw) return null

    const parsed: unknown = JSON.parse(raw)
    return isStoredDraft(parsed) ? parsed : null
  } catch {
    return null
  }
}

async function load(): Promise<DraftData> {
  const meta = readMeta()

  if (!meta) return emptyDraft()

  const sections: Section[] = []

  for (const stored of meta.sections) {
    const files: FileRef[] = []

    for (const fileMeta of stored.files) {
      const blob = await idbGet<Blob>(
        blobKey(fileMeta.id),
      ).catch(() => undefined)

      // Файл мог быть удалён из IndexedDB или не сохраниться.
      // Сохраняем текст секции и пропускаем только потерянное вложение.
      if (!blob) continue

      const file =
        blob instanceof File
          ? blob
          : new File([blob], fileMeta.name, {
              type: fileMeta.type,
              lastModified: fileMeta.lastModified,
            })

      files.push({ id: fileMeta.id, file })
      blobStatus.set(fileMeta.id, 'saved')
    }

    if (stored.text !== '' || files.length > 0) {
      sections.push({
        id: stored.id,
        text: stored.text,
        files,
      })
    }
  }

  return {
    title: meta.title,
    sections,
  }
}

function saveMeta(): void {
  saveTimer = undefined

  const { title, sections } = useDraftStore.getState()

  try {
    const storedSections: StoredSection[] = sections
      .map((section) => ({
        id: section.id,
        text: section.text,
        files: section.files
          .filter(({ id }) => blobStatus.get(id) === 'saved')
          .map(({ id, file }) => ({
            id,
            name: file.name,
            type: file.type,
            lastModified: file.lastModified,
          })),
      }))
      // Не храним пустые секции без текста и сохранённых файлов.
      .filter(
        (section) =>
          section.text !== '' || section.files.length > 0,
      )

    if (title === '' && storedSections.length === 0) {
      localStorage.removeItem(META_KEY)
      return
    }

    const data: StoredDraft = {
      title,
      sections: storedSections,
    }

    localStorage.setItem(META_KEY, JSON.stringify(data))
  } catch {
    // Ошибка хранилища не должна ломать редактор.
    // Текущий черновик остаётся в состоянии Zustand.
  }
}

function scheduleSave(): void {
  if (saveTimer !== undefined) {
    clearTimeout(saveTimer)
  }

  saveTimer = setTimeout(saveMeta, SAVE_DELAY_MS)
}

function syncBlobs(sections: Section[]): void {
  const current = new Map<string, FileRef>()

  for (const section of sections) {
    for (const ref of section.files) {
      current.set(ref.id, ref)
    }
  }

  // Начинаем сохранять только новые файлы.
  for (const [id, ref] of current) {
    if (blobStatus.has(id)) continue

    blobStatus.set(id, 'pending')

    void idbSet(blobKey(id), ref.file)
      .then(() => {
        // Файл могли удалить, пока его запись выполнялась.
        if (blobStatus.get(id) !== 'pending') {
          void idbDelete(blobKey(id)).catch(() => {})
          return
        }

        blobStatus.set(id, 'saved')
        scheduleSave()
      })
      .catch(() => {
        // Разрешаем следующей синхронизации попробовать запись снова.
        if (blobStatus.get(id) === 'pending') {
          blobStatus.delete(id)
        }
      })
  }

  // Удаляем больше не используемые вложения.
  for (const id of blobStatus.keys()) {
    if (current.has(id)) continue

    blobStatus.delete(id)
    void idbDelete(blobKey(id)).catch(() => {})
  }
}

function flushMeta(): void {
  if (saveTimer === undefined) return

  clearTimeout(saveTimer)
  saveTimer = undefined
  saveMeta()
}

async function initializePostEditor(): Promise<void> {
  let data = emptyDraft()

  try {
    data = await load()
  } catch {
    blobStatus.clear()
  }

  useDraftStore.getState().hydrate(data)

  useDraftStore.subscribe((state, prev) => {
    if (!state.hydrated) return

    const sectionsChanged = state.sections !== prev.sections
    const titleChanged = state.title !== prev.title

    if (sectionsChanged) {
      syncBlobs(state.sections)
    }

    if (sectionsChanged || titleChanged) {
      scheduleSave()
    }
  })

  window.addEventListener('pagehide', flushMeta)
}

/** Повторные вызовы ожидают одну и ту же инициализацию. */
export function initPostEditor(): Promise<void> {
  initPromise ??= initializePostEditor()
  return initPromise
}