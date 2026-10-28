import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchMe, login, logout, register } from '../api/user'

export const meKey = ['auth', 'me'] as const

export function useMe() {
  return useQuery({
    queryKey: meKey,
    queryFn: fetchMe,
    staleTime: Infinity,
    retry: false,
  })
}

export function useRefreshMe() {
  const qc = useQueryClient()
  return () => qc.invalidateQueries({ queryKey: meKey })
}

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: login,
    onSuccess: (user) => qc.setQueryData(meKey, user),
  })
}

export function useRegister() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: register,
    onSuccess: (user) => qc.setQueryData(meKey, user),
  })
}

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: logout,
    onSuccess: () => {
      qc.setQueryData(meKey, null)
    },
  })
}