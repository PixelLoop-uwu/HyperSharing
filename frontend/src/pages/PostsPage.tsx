import { useInfiniteQuery } from '@tanstack/react-query'
import { listPosts } from '@/features/posts/api'
import PostCard from '@/shared/ui/layout/Card'

const PAGE_SIZE = 20

export default function PostsPage() {
  const query = useInfiniteQuery({
    queryKey: ['posts'],
    queryFn: ({ pageParam }) => listPosts(pageParam, PAGE_SIZE),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) =>
      lastPage.length === PAGE_SIZE ? allPages.length * PAGE_SIZE : undefined,
  })

  const posts = query.data?.pages.flat() ?? []

  return (
    <div className="mx-auto w-full max-w-3xl py-16">
      <h1 className="font-display text-3xl font-semibold tracking-tight">Посты</h1>

      {query.isLoading ? (
        <p className="mt-8 text-muted">Загрузка…</p>
      ) : query.isError ? (
        <p className="mt-8 text-red-400">
          {query.error instanceof Error ? query.error.message : 'Не удалось загрузить посты'}
        </p>
      ) : posts.length === 0 ? (
        <p className="mt-8 text-muted">Постов пока нет.</p>
      ) : (
        <div className="mt-8 grid gap-3">
          {posts.map((post) => (
            <PostCard key={post.id} />
          ))}
        </div>
      )}

      {query.hasNextPage && (
        <button
          type="button"
          onClick={() => void query.fetchNextPage()}
          disabled={query.isFetchingNextPage}
          className="mt-6 rounded-lg border border-line px-5 py-2.5 text-sm text-paper transition-colors hover:border-paper/40 disabled:opacity-60"
        >
          {query.isFetchingNextPage ? 'Загрузка…' : 'Показать ещё'}
        </button>
      )}
    </div>
  )
}
