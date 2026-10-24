import { api } from '@/shared/api/client'
import type { PostDraftCreate, PostDraftRead, PostListRead, PostRead, QuotaRead } from './types'

export async function createDraft(data: PostDraftCreate): Promise<PostDraftRead> {
  const { data: draft } = await api.post<PostDraftRead>('/drafts', data)
  return draft
}

export async function publishDraft(draftId: string): Promise<PostRead> {
  const { data: post } = await api.post<PostRead>(`/drafts/${draftId}/publish`)
  return post
}

export async function getPost(id: string): Promise<PostRead> {
  const { data } = await api.get<PostRead>(`/posts/${id}`)
  return data
}

/** Cursor-paginated list of the signed-in user's posts. `cursor` is the previous page's `next_cursor`. */
export async function listMyPosts(cursor?: string, limit = 20): Promise<PostListRead> {
  const { data } = await api.get<PostListRead>('/me/posts', { params: { cursor, limit } })
  return data
}

/** Current storage + daily upload quota for the signed-in user (approximate). */
export async function getMyQuota(): Promise<QuotaRead> {
  const { data } = await api.get<QuotaRead>('/me/quota')
  return data
}
