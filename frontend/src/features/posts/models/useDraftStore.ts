import { create } from 'zustand'
import * as ops from '../lib/sections'
import type { DraftData, FileTarget, Section } from './types'

interface DraftState extends DraftData {
  /** Становится true после восстановления сохранённого черновика. */
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

function updateSections(
  state: DraftState,
  update: (sections: Section[]) => Section[],
): DraftState | { sections: Section[] } {
  const sections = update(state.sections)

  return sections === state.sections
    ? state
    : { sections }
}

export const useDraftStore = create<DraftState>()((set) => ({
  title: '',
  sections: [],
  hydrated: false,

  setTitle: (title) => set({ title }),

  addSection: (at) =>
    set((state) => ({
      sections: ops.insertSection(
        state.sections,
        ops.makeSection(),
        at,
      ),
    })),

  setText: (sectionId, text) =>
    set((state) =>
      updateSections(state, (sections) =>
        ops.setText(sections, sectionId, text),
      ),
    ),

  addFiles: (files, target) =>
    set((state) => {
      if (files.length === 0) return state

      const refs = ops.makeRefs(files)

      return updateSections(state, (sections) =>
        ops.addFiles(
          sections,
          refs,
          target ?? { newAt: sections.length },
        ),
      )
    }),

  removeFile: (sectionId, fileId) =>
    set((state) =>
      updateSections(state, (sections) =>
        ops.removeFile(sections, sectionId, fileId),
      ),
    ),

  removeSection: (sectionId) =>
    set((state) =>
      updateSections(state, (sections) =>
        ops.removeSection(sections, sectionId),
      ),
    ),

  moveSection: (from, to) =>
    set((state) =>
      updateSections(state, (sections) =>
        ops.moveSection(sections, from, to),
      ),
    ),

  hydrate: (data) =>
    set({
      ...data,
      hydrated: true,
    }),

  reset: () =>
    set({
      title: '',
      sections: [],
    }),
}))