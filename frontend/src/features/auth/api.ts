import { api } from '@/shared/api/client'
import type { LoginIn, RegisterIn, UserRead } from './types'

export async function fetchMe(): Promise<UserRead> {
  const { data } = await api.get<UserRead>('/auth/me')
  return data
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
