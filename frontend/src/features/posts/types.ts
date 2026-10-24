// Wire DTOs (snake_case, matching the Go backend) + the editor's domain models.
// Kept in one file to mirror `features/user/types.ts`.

// --- Editor domain models ---

/** A file in the draft. The id keys its IndexedDB blob. */
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

// --- Draft creation (POST /drafts) ---

export interface DraftFileIn {
  name: string
  mime: string
  size: number
}

export interface DraftSectionIn {
  body: string
  files: DraftFileIn[]
}

export interface PostDraftCreate {
  title: string
  sections: DraftSectionIn[]
}

/** Response of POST /drafts: slot ids for each declared section/file. */
export interface DraftFileRead {
  id: string
}

export interface DraftSectionRead {
  id: string
  files: DraftFileRead[]
}

export interface PostDraftRead {
  id: string
  sections: DraftSectionRead[]
}

// --- Published post (GET /posts/{id}, POST /drafts/{id}/publish) ---

export interface FileRead {
  name: string
  mime: string
  size: number
  url: string
}

export interface SectionRead {
  body: string
  files: FileRead[]
}

export interface AuthorRead {
  username: string
  avatar_url: string | null
}

export interface PostRead {
  id: string
  title: string
  author: AuthorRead | null
  created_at: string
  sections: SectionRead[]
}

// --- My posts list (GET /me/posts) ---

export interface PostListItem {
  id: string
  title: string
  total_size: number
  created_at: string
}

export interface PostListRead {
  posts: PostListItem[]
  next_cursor: string | null
}

// --- Quota (GET /me/quota) ---

export interface QuotaMetric {
  used: number
  limit: number
}

export interface QuotaRead {
  storage: QuotaMetric
  daily: QuotaMetric
}
