import { useInfiniteQuery } from '@tanstack/react-query'
import { listMyPosts } from '../api/posts'

const PAGE_SIZE = 20

export function useMyPosts() {
  return useInfiniteQuery({
    queryKey: ['me', 'posts'],
    queryFn: ({ pageParam }) => listMyPosts(pageParam ?? undefined, PAGE_SIZE),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.next_cursor,
  })
}