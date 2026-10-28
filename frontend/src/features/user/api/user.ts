import { api } from '@/shared/api/client'
import { ApiError } from '@/shared/api/errors'
import type { LoginIn, RegisterIn, UserRead } from '../model/types'


export async function fetchMe(): Promise<UserRead | null> {
  try {
    const { data } = await api.get<UserRead>('/auth/me')
    return data
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) return null
    throw error
  }
}

export async function login(input: LoginIn): Promise<UserRead> {
  const { data } = await api.post<UserRead>('/auth/login', input)
  return data
}

export async function register(input: RegisterIn): Promise<UserRead> {
  const { data } = await api.post<UserRead>('/auth/register', input)
  return data
}

export async function logout(): Promise<void> {
  await api.post('/auth/logout')
}

export async function uploadAvatar(file: File): Promise<void> {
  await api.put('/me/avatar', file, {
    headers: { 'Content-Type': file.type || 'application/octet-stream' },
  })
}

export async function checkAvatar(userId: string): Promise<boolean> {
  try {
    await api.head(`/avatars/${userId}`)
    return true
  } catch (error) {
    const status = (
      error as {
        status?: number
        response?: { status?: number }
      }
    )?.status ?? (
      error as {
        response?: { status?: number }
      }
    )?.response?.status

    if (status === 404) return false

    throw error
  }
}

export async function deleteAvatar(): Promise<void> {
  await api.delete('/me/avatar')
}
