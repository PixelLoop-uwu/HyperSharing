import type { FileRef, FileTarget, Section } from '../models/types'

const clamp = (n: number, min: number, max: number) =>
  Math.min(Math.max(n, min), max)

export const makeSection = (text = '', files: FileRef[] = []): Section => ({
  id: crypto.randomUUID(),
  text,
  files,
})

export const makeRefs = (files: File[]): FileRef[] =>
  files.map((file) => ({ id: crypto.randomUUID(), file }))

export const isBlank = (section: Section): boolean =>
  section.text.trim() === '' && section.files.length === 0

export function insertSection(
  sections: Section[],
  section: Section,
  at = sections.length,
): Section[] {
  const index = Math.trunc(clamp(at, 0, sections.length))

  return [
    ...sections.slice(0, index),
    section,
    ...sections.slice(index),
  ]
}

export function moveSection(
  sections: Section[],
  from: number,
  to: number,
): Section[] {
  if (
    !Number.isInteger(from) ||
    !Number.isInteger(to) ||
    from < 0 ||
    from >= sections.length
  ) {
    return sections
  }

  const next = sections.slice()
  const [item] = next.splice(from, 1)

  if (!item) return sections

  const destination = Math.trunc(clamp(to, 0, next.length))

  // После нормализации индексов перемещение может оказаться пустой операцией.
  if (destination === from) return sections

  next.splice(destination, 0, item)
  return next
}

export function setText(
  sections: Section[],
  id: string,
  text: string,
): Section[] {
  const section = sections.find((item) => item.id === id)

  if (!section || section.text === text) return sections

  return sections.map((item) =>
    item.id === id ? { ...item, text } : item,
  )
}

export function removeSection(
  sections: Section[],
  id: string,
): Section[] {
  if (!sections.some((section) => section.id === id)) {
    return sections
  }

  return sections.filter((section) => section.id !== id)
}

export function addFiles(
  sections: Section[],
  refs: FileRef[],
  target: FileTarget,
): Section[] {
  if (refs.length === 0) return sections

  if ('sectionId' in target) {
    const exists = sections.some(
      (section) => section.id === target.sectionId,
    )

    if (exists) {
      return sections.map((section) =>
        section.id === target.sectionId
          ? { ...section, files: [...section.files, ...refs] }
          : section,
      )
    }
  }

  const at = 'newAt' in target ? target.newAt : sections.length

  return insertSection(sections, makeSection('', refs), at)
}

export function removeFile(
  sections: Section[],
  sectionId: string,
  fileId: string,
): Section[] {
  const section = sections.find((item) => item.id === sectionId)

  // Несуществующая секция или файл: состояние не меняется.
  if (!section || !section.files.some((file) => file.id === fileId)) {
    return sections
  }

  const files = section.files.filter((file) => file.id !== fileId)
  const updated = { ...section, files }

  // Пустая секция не нужна после удаления последнего файла.
  if (isBlank(updated)) {
    return removeSection(sections, sectionId)
  }

  return sections.map((item) =>
    item.id === sectionId ? updated : item,
  )
}