import { api } from '@/shared/api/client'
import type { PostDraftCreate, PostDraftRead, PostListRead, PostRead, QuotaRead } from '../models/types'


export async function createDraft(data: PostDraftCreate): Promise<PostDraftRead> {
  const { data: draft } = await api.post<PostDraftRead>('/drafts', data)
  return draft
}

export async function publishDraft(draftId: string): Promise<PostRead> {
  const { data: post } = await api.post<PostRead>(`/drafts/${draftId}/publish`)
  return post
}

export async function getPost(shortKey: string): Promise<PostRead> {
  const { data } = await api.get<PostRead>(`/posts/${shortKey}`)
  return data
}

export async function deletePost(postId: string): Promise<void> {
  await api.delete(`/posts/${postId}`)
}

export async function listMyPosts(cursor?: string, limit = 20): Promise<PostListRead> {
  const { data } = await api.get<PostListRead>('/me/posts', { params: { cursor, limit } })
  return data
}

export async function getMyQuota(): Promise<QuotaRead> {
  const { data } = await api.get<QuotaRead>('/me/quota')
  return data
}
