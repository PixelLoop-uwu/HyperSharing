import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useMe } from '@/features/user'

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
  const me = useMe()

  if (me.isLoading) return <>{fallback}</>
  if (!me.data) return <Navigate to="/login" replace />
  return <>{children}</>
}
