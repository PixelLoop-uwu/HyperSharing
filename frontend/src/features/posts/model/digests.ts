import { create } from 'zustand'
import { computeFileDigest, type FileDigest } from '../lib'
import { useDraftStore } from './draft'
import type { FileRef, Section } from './types'

export type DigestState =
  | { status: 'pending' }
  | { status: 'ready'; digest: FileDigest }
  | { status: 'error'; message: string }

interface DigestsState {
  byFileId: Record<string, DigestState>
}

/** Hash state per file id. Not persisted: hashes are recomputed after a reload. */
export const useDigests = create<DigestsState>()(() => ({ byFileId: {} }))

export const useFileDigest = (fileId: string): DigestState | undefined =>
  useDigests((s) => s.byFileId[fileId])

function patch(id: string, state: DigestState | null) {
  useDigests.setState((s) => {
    const next = { ...s.byFileId }
    if (state) next[id] = state
    else delete next[id]
    return { byFileId: next }
  })
}

const inDraft = (id: string) =>
  useDraftStore.getState().sections.some((s) => s.files.some((f) => f.id === id))

// Hash one file at a time so a big drop doesn't saturate the CPU.
let queue: Promise<void> = Promise.resolve()

function enqueue(ref: FileRef) {
  patch(ref.id, { status: 'pending' })
  queue = queue.then(async () => {
    if (!inDraft(ref.id)) return
    try {
      const digest = await computeFileDigest(ref.file)
      if (inDraft(ref.id)) patch(ref.id, { status: 'ready', digest })
    } catch (err) {
      if (inDraft(ref.id)) {
        patch(ref.id, {
          status: 'error',
          message: err instanceof Error ? err.message : 'Не удалось посчитать хэш',
        })
      }
    }
  })
}

let started = false

/** Hashes every file as soon as it appears in the draft and forgets removed ones. */
export function startDigestWatcher(): void {
  if (started) return
  started = true

  const seen = new Set<string>()
  const sync = (sections: Section[]) => {
    const present = new Set<string>()
    for (const s of sections) {
      for (const ref of s.files) {
        present.add(ref.id)
        if (!seen.has(ref.id)) {
          seen.add(ref.id)
          enqueue(ref)
        }
      }
    }
    for (const id of seen) {
      if (present.has(id)) continue
      seen.delete(id)
      patch(id, null)
    }
  }

  sync(useDraftStore.getState().sections)
  useDraftStore.subscribe((state, prev) => {
    if (state.sections !== prev.sections) sync(state.sections)
  })
}
