import { idbDelete, idbGet, idbSet } from '@/shared/lib/idb'
import { useDraftStore } from './store'
import type { DraftData, FileRef, Section } from './types'

// Two stores, on purpose:
//  - structure + text  -> localStorage (small, saved debounced on every edit)
//  - File blobs        -> IndexedDB, written once when a file is added, never rewritten
// so typing in a paragraph no longer rewrites every attachment.

const META_KEY = 'hypersharing:post-draft:v2'
const SAVE_DELAY_MS = 300
const blobKey = (id: string) => `hypersharing:file:${id}`

interface StoredFile {
  id: string
  name: string
  size: number
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

const emptyDraft = (): DraftData => ({ title: '', sections: [] })

/** File ids whose blob is confirmed in IndexedDB. Only these are listed in the saved meta. */
const saved = new Set<string>()
/** File ids this module has already seen in the draft. */
let known = new Set<string>()
let timer: ReturnType<typeof setTimeout> | undefined
let started = false

function readMeta(): StoredDraft | null {
  try {
    const raw = localStorage.getItem(META_KEY)
    if (!raw) return null
    const parsed: unknown = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object' || !Array.isArray((parsed as StoredDraft).sections)) {
      return null
    }
    return parsed as StoredDraft
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
    for (const f of stored.files) {
      const blob = await idbGet<Blob>(blobKey(f.id)).catch(() => undefined)
      if (!blob) continue // blob lost (quota, cleared storage): drop the file, keep the text
      const file =
        blob instanceof File
          ? blob
          : new File([blob], f.name, { type: f.type, lastModified: f.lastModified })
      files.push({ id: f.id, file })
      saved.add(f.id)
    }
    if (stored.text || files.length > 0) sections.push({ id: stored.id, text: stored.text, files })
  }
  return { title: meta.title, sections }
}

function saveMeta() {
  timer = undefined
  const { title, sections } = useDraftStore.getState()
  try {
    if (!title && sections.length === 0) {
      localStorage.removeItem(META_KEY)
      return
    }
    const data: StoredDraft = {
      title,
      sections: sections.map((s) => ({
        id: s.id,
        text: s.text,
        files: s.files
          .filter((f) => saved.has(f.id))
          .map(({ id, file }) => ({
            id,
            name: file.name,
            size: file.size,
            type: file.type,
            lastModified: file.lastModified,
          })),
      })),
    }
    localStorage.setItem(META_KEY, JSON.stringify(data))
  } catch {
    // Storage full or unavailable: the draft just stays in memory.
  }
}

function scheduleSave() {
  clearTimeout(timer)
  timer = setTimeout(saveMeta, SAVE_DELAY_MS)
}

function syncBlobs(sections: Section[]) {
  const current = new Map<string, FileRef>()
  for (const s of sections) for (const ref of s.files) current.set(ref.id, ref)

  for (const [id, ref] of current) {
    if (known.has(id)) continue
    known.add(id)
    idbSet(blobKey(id), ref.file)
      .then(() => {
        if (known.has(id)) {
          saved.add(id)
          scheduleSave()
        } else {
          // The file was removed while it was still being written.
          void idbDelete(blobKey(id)).catch(() => {})
        }
      })
      .catch(() => {})
  }

  for (const id of known) {
    if (current.has(id)) continue
    known.delete(id)
    saved.delete(id)
    void idbDelete(blobKey(id)).catch(() => {})
  }
}

/**
 * Loads the saved draft into the store and starts saving changes.
 * Safe to call more than once. The editor should wait for `hydrated`.
 */
export async function initPostEditor(): Promise<void> {
  if (started) return
  started = true

  let data = emptyDraft()
  try {
    data = await load()
  } catch {
    saved.clear()
  }
  known = new Set(data.sections.flatMap((s) => s.files.map((f) => f.id)))
  useDraftStore.getState().hydrate(data)

  useDraftStore.subscribe((state, prev) => {
    if (!state.hydrated) return
    const sectionsChanged = state.sections !== prev.sections
    if (sectionsChanged) syncBlobs(state.sections)
    if (sectionsChanged || state.title !== prev.title) scheduleSave()
  })

  window.addEventListener('pagehide', () => {
    if (timer !== undefined) {
      clearTimeout(timer)
      saveMeta()
    }
  })
}
