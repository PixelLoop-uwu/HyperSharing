import { create } from 'zustand'
import { fetchMe, login, logout, register } from './api'
import type { LoginIn, RegisterIn, UserRead } from './types'

type AuthStatus = 'loading' | 'authenticated' | 'anonymous'

interface AuthState {
  user: UserRead | null
  status: AuthStatus
  avatarVersion: number
  load: () => Promise<void>
  login: (input: LoginIn) => Promise<void>
  register: (input: RegisterIn) => Promise<void>
  logout: () => Promise<void>
  bumpAvatarVersion: () => void
}

export const useAuth = create<AuthState>()((set) => ({
  user: null,
  status: 'loading',
  avatarVersion: 0,
  load: async () => {
    try {
      const user = await fetchMe()
      set({ user, status: 'authenticated' })
    } catch {
      set({ user: null, status: 'anonymous' })
    }
  },
  login: async (input) => {
    const user = await login(input)
    set({ user, status: 'authenticated' })
  },
  register: async (input) => {
    const user = await register(input)
    set({ user, status: 'authenticated' })
  },
  logout: async () => {
    await logout()
    set({ user: null, status: 'anonymous' })
  },
  bumpAvatarVersion: () => set((s) => ({ avatarVersion: s.avatarVersion + 1 })),
}))
