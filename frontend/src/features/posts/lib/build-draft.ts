import type {
  DraftSectionIn,
  PostDraftCreate,
  Section,
} from '../models/types'

export interface BuiltDraft {
  payload: PostDraftCreate
  files: File[]
}

export function buildDraft(
  title: string,
  sections: Section[],
): BuiltDraft {
  const payloadSections: DraftSectionIn[] = []
  const files: File[] = []

  for (const section of sections) {
    if (
      section.text.trim() === '' &&
      section.files.length === 0
    ) {
      continue
    }

    payloadSections.push({
      body: section.text.trim(),
      files: section.files.map(({ file }) => ({
        name: file.name,
        mime: file.type || 'application/octet-stream',
        size: file.size,
      })),
    })

    for (const { file } of section.files) {
      files.push(file)
    }
  }

  return {
    payload: {
      title: title.trim(),
      sections: payloadSections,
    },
    files,
  }
}