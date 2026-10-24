import { useQuery } from '@tanstack/react-query'
import { getMyQuota } from '../api'

export function useQuota() {
  return useQuery({
    queryKey: ['me', 'quota'],
    queryFn: getMyQuota,
  })
}
