import type { FileRef, FileTarget, Section } from './types'

// Pure operations on the section list. No state, no side effects.

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
