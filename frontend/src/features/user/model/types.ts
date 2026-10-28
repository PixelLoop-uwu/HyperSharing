export interface UserRead {
  id: string
  email: string
  username: string
  created_at: string
  email_verified: boolean
  is_admin: boolean
  avatar_version: string | null
}

export interface RegisterIn {
  email: string
  username: string
  password: string
}

export interface LoginIn {
  email: string
  password: string
}
