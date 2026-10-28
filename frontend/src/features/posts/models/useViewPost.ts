import { useQuery } from '@tanstack/react-query'
import { getPost } from '../api/posts'
import { getStatus } from '@/shared/api/errors'

export function useViewPost(shortKey?: string) {
  return useQuery({
    queryKey: ['post', shortKey],
    retry: (count, error) => getStatus(error) !== 404 && count < 2,
    queryFn: () => {
      if (!shortKey) {
        throw new Error('Post key is required')
      }

      return getPost(shortKey)
    },
    enabled: Boolean(shortKey),
  })
}