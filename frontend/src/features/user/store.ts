import { create } from 'zustand'
import { fetchMe, login, logout, register } from './api'
import type { LoginIn, RegisterIn, UserRead } from './types'

type AuthStatus = 'loading' | 'authenticated' | 'anonymous'

interface AuthState {
  user: UserRead | null
  status: AuthStatus
  load: () => Promise<void>
  login: (input: LoginIn) => Promise<void>
  register: (input: RegisterIn) => Promise<void>
  logout: () => Promise<void>
}

export const useAuth = create<AuthState>()((set) => ({
  user: null,
  status: 'loading',
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
}))
