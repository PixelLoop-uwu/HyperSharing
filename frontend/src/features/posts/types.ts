// Wire-format DTOs for the post API, kept snake_case to match the backend
// payloads directly (same convention as ``features/auth/types.ts``).

export interface ParagraphItemIn {
  kind: 'paragraph'
  text: string
}

export interface FileItemIn {
  kind: 'file'
  hash: string
  size: number
  content_type: string | null
  filename: string | null
  part_checksums: string[]
}

export type PostItemIn = ParagraphItemIn | FileItemIn

export interface PostDraftCreate {
  title: string | null
  items: PostItemIn[]
}

export interface PartPresignOut {
  part_number: number
  url: string
  checksum_sha256: string
}

export interface FileUploadOut {
  hash: string
  part_size: number
  parts: PartPresignOut[]
}

export interface PostDraftRead {
  draft_id: string
  expires_at: string
  files: FileUploadOut[]
}

export interface PartEtagIn {
  part_number: number
  etag: string
}

export interface FileCompleteIn {
  hash: string
  parts: PartEtagIn[]
}

export interface PostPublishIn {
  draft_id: string
  files: FileCompleteIn[]
}

export interface ParagraphItemOut {
  kind: 'paragraph'
  text: string
}

export interface FileItemOut {
  kind: 'file'
  hash: string
  size: number
  content_type: string | null
  filename: string | null
  url: string
}

export type PostItemOut = ParagraphItemOut | FileItemOut

export interface PostFullRead {
  id: string
  title: string | null
  user_id: string | null
  items: PostItemOut[]
  created_at: string
}

export interface PostRead {
  id: string
  title: string | null
  user_id: string | null
  created_at: string
  updated_at: string
}
