import type {
  DraftSectionIn,
  FileRef,
  FileTarget,
  PostDraftCreate,
  Section,
} from './types'

const TUS_VERSION = '1.0.0'
const CHUNK_SIZE = 8 * 1024 * 1024

// --- Pure operations on the section list. No state, no side effects. ---

const clamp = (n: number, min: number, max: number) => Math.min(Math.max(n, min), max)

export const makeSection = (text = '', files: FileRef[] = []): Section => ({
  id: crypto.randomUUID(),
  text,
  files,
})

export const makeRefs = (files: File[]): FileRef[] =>
  files.map((file) => ({ id: crypto.randomUUID(), file }))

/** A section with no text and no files carries nothing; it is skipped on publish. */
export const isBlank = (s: Section): boolean => s.text.trim() === '' && s.files.length === 0

export function insertSection(
  sections: Section[],
  section: Section,
  at = sections.length,
): Section[] {
  const index = clamp(at, 0, sections.length)
  return [...sections.slice(0, index), section, ...sections.slice(index)]
}

export function moveSection(sections: Section[], from: number, to: number): Section[] {
  if (from === to || from < 0 || from >= sections.length) return sections
  const next = sections.slice()
  const [item] = next.splice(from, 1)
  if (!item) return sections
  next.splice(clamp(to, 0, next.length), 0, item)
  return next
}

export const setText = (sections: Section[], id: string, text: string): Section[] =>
  sections.map((s) => (s.id === id ? { ...s, text } : s))

export const removeSection = (sections: Section[], id: string): Section[] =>
  sections.filter((s) => s.id !== id)

/**
 * Drop onto a section appends to it; drop onto the "new section" zone creates one
 * holding all the dropped files. An unknown sectionId falls back to a new section at the end.
 */
export function addFiles(sections: Section[], refs: FileRef[], target: FileTarget): Section[] {
  if (refs.length === 0) return sections
  if ('sectionId' in target && sections.some((s) => s.id === target.sectionId)) {
    return sections.map((s) =>
      s.id === target.sectionId ? { ...s, files: [...s.files, ...refs] } : s,
    )
  }
  const at = 'newAt' in target ? target.newAt : sections.length
  return insertSection(sections, makeSection('', refs), at)
}

/** Removing the last file of a section with no text removes the section too. */
export function removeFile(sections: Section[], sectionId: string, fileId: string): Section[] {
  return sections.flatMap((s) => {
    if (s.id !== sectionId) return [s]
    const next = { ...s, files: s.files.filter((f) => f.id !== fileId) }
    return isBlank(next) ? [] : [next]
  })
}

// --- Draft payload ---

export interface BuiltDraft {
  payload: PostDraftCreate
  /** Files in declaration order, for the tus upload right after the draft is created. */
  files: File[]
}

/** Turns the editor state into the POST /drafts body, keeping the files in order. */
export function buildDraft(title: string, sections: Section[]): BuiltDraft {
  const nonBlank = sections.filter((s) => !isBlank(s))
  const out: DraftSectionIn[] = []
  const files: File[] = []
  for (const s of nonBlank) {
    out.push({
      body: s.text.trim(),
      files: s.files.map(({ file }) => ({
        name: file.name,
        mime: file.type || 'application/octet-stream',
        size: file.size,
      })),
    })
    for (const { file } of s.files) files.push(file)
  }
  return { payload: { title: title.trim(), sections: out }, files }
}

// --- tus upload ---

export interface TusMeta {
  draftId: string
  fileId: string
}

/** base64 for Upload-Metadata values (tus requires base64-encoded values). */
function b64(value: string): string {
  return btoa(value)
}

/**
 * Uploads `file` through the tus endpoint (`POST /uploads`, then PATCH chunks).
 * The upload id is the file's slot id, so no ETag/checksum bookkeeping is needed:
 * the backend already ties every byte to the declared slot.
 */
export async function uploadFileViaTus(
  file: File,
  meta: TusMeta,
  onProgress?: (uploaded: number, total: number) => void,
  signal?: AbortSignal,
): Promise<void> {
  const createRes = await fetch('/api/v1/uploads', {
    method: 'POST',
    headers: {
      'Tus-Resumable': TUS_VERSION,
      'Upload-Length': String(file.size),
      'Upload-Metadata': `draftId ${b64(meta.draftId)},fileId ${b64(meta.fileId)}`,
    },
    credentials: 'same-origin',
    signal,
  })
  if (!createRes.ok) {
    throw new Error(`не удалось начать загрузку «${file.name}» (${createRes.status})`)
  }
  const location = createRes.headers.get('Location')
  if (!location) throw new Error('сервер не вернул адрес загрузки')

  let offset = 0
  while (offset < file.size) {
    const end = Math.min(file.size, offset + CHUNK_SIZE)
    const res = await fetch(location, {
      method: 'PATCH',
      headers: {
        'Tus-Resumable': TUS_VERSION,
        'Upload-Offset': String(offset),
        'Content-Type': 'application/offset+octet-stream',
      },
      credentials: 'same-origin',
      body: file.slice(offset, end),
      signal,
    })
    if (res.status !== 204) {
      throw new Error(`загрузка «${file.name}» не удалась (${res.status})`)
    }
    offset = end
    onProgress?.(offset, file.size)
  }
}
