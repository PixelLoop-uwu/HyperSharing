import type { PostItemIn } from '../types'
import type { DigestState } from './digests'
import type { Section } from './types'

export type BuildResult =
  | {
      ok: true
      items: PostItemIn[]
      /** One File per unique hash: what actually has to be uploaded. */
      uploads: Map<string, File>
    }
  | { ok: false; reason: 'empty' | 'hashing' | 'hash-failed'; fileName?: string }

/**
 * Flattens sections into the current wire format: for each section, its text as a
 * paragraph item, then its files. Order is preserved. Once the Go backend takes
 * sections natively, only this function changes.
 */
export function buildItems(
  sections: Section[],
  digests: Record<string, DigestState>,
): BuildResult {
  const items: PostItemIn[] = []
  const uploads = new Map<string, File>()

  for (const section of sections) {
    const text = section.text.trim()
    if (text) items.push({ kind: 'paragraph', text })

    for (const { id, file } of section.files) {
      const state = digests[id]
      if (!state || state.status === 'pending') return { ok: false, reason: 'hashing' }
      if (state.status === 'error') return { ok: false, reason: 'hash-failed', fileName: file.name }

      const d = state.digest
      items.push({
        kind: 'file',
        hash: d.hash,
        size: d.size,
        content_type: d.content_type,
        filename: file.name,
        part_checksums: d.part_checksums,
      })
      if (!uploads.has(d.hash)) uploads.set(d.hash, file)
    }
  }

  if (items.length === 0) return { ok: false, reason: 'empty' }
  return { ok: true, items, uploads }
}
