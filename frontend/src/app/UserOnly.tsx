import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '@/features/user/store'

/**
 * Пропускает children только для залогиненного пользователя.
 * Пока статус грузится — показывает `fallback`; анонима редиректит на /login.
 */
export default function UserOnly({
  fallback = null,
  children,
}: {
  fallback?: ReactNode
  children: ReactNode
}) {
  const status = useAuth((s) => s.status)

  if (status === 'loading') return <>{fallback}</>
  if (status !== 'authenticated') return <Navigate to="/login" replace />
  return <>{children}</>
}
