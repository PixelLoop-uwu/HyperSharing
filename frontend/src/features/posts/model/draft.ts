import { create } from 'zustand'
import * as ops from './sections'
import type { DraftData, FileTarget } from './types'

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

// Editing state only. Persistence, hashing and publishing live in their own modules
// and watch this store from the outside.
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
