/** A file in the draft. The id keys its IndexedDB blob and its digest. */
export interface FileRef {
  id: string
  file: File
}

/** One unit of the post: optional text plus any number of files. */
export interface Section {
  id: string
  text: string
  files: FileRef[]
}

export interface DraftData {
  title: string
  sections: Section[]
}

/** Where dropped files go: into an existing section, or into a new one at an index. */
export type FileTarget = { sectionId: string } | { newAt: number }
