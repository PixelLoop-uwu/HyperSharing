import { api } from '@/shared/api/client'
import type {
  PostDraftCreate,
  PostDraftRead,
  PostFullRead,
  PostPublishIn,
  PostRead,
} from './types'

export async function createDraft(data: PostDraftCreate): Promise<PostDraftRead> {
  const { data: draft } = await api.post<PostDraftRead>('/posts/draft', data)
  return draft
}

export async function publishPost(data: PostPublishIn): Promise<PostFullRead> {
  const { data: post } = await api.post<PostFullRead>('/posts/publish', data)
  return post
}

export async function getPost(id: string): Promise<PostFullRead> {
  const { data } = await api.get<PostFullRead>(`/posts/${id}`)
  return data
}

export async function listPosts(offset = 0, limit = 100): Promise<PostRead[]> {
  const { data } = await api.get<PostRead[]>('/posts', { params: { offset, limit } })
  return data
}
