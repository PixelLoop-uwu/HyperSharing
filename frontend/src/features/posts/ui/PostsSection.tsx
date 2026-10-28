import { motion } from 'framer-motion'
import { Link } from 'react-router-dom'
import GhostButton from '@/shared/ui/actions/GhostButton'
import Skeleton from '@/shared/ui/feedback/Skeleton'
import Card from '@/shared/ui/layout/Card'
import { stagger } from '@/shared/lib/motion'
import { useMyPosts } from '../models/useMyPosts'
import PostCard from './PostCard'

const linkClass =
  'text-zinc-300 underline underline-offset-4 transition-colors hover:text-zinc-100'

export default function PostsSection() {
  const query = useMyPosts()
  const posts = query.data?.pages.flatMap((p) => p.posts) ?? []

  return (
    <section>
      <h2 className="font-display text-xl font-semibold tracking-tight text-zinc-100">Мои посты</h2>

      {query.isLoading ? (
        <ul className="mt-6 grid gap-3" aria-busy="true">
          {Array.from({ length: 4 }, (_, i) => (
            <li key={i}>
              <Skeleton className="h-22 w-full rounded-xl" />
            </li>
          ))}
        </ul>
      ) : query.isError ? (
        <Card className="mt-6 flex items-center justify-between gap-4 p-5">
          <p role="alert" className="text-sm text-red-400">
            {query.error instanceof Error ? query.error.message : 'Не удалось загрузить посты'}
          </p>
          <GhostButton onClick={() => void query.refetch()}>Повторить</GhostButton>
        </Card>
      ) : posts.length === 0 ? (
        <Card className="mt-6 p-8 text-center text-sm text-zinc-400">
          Постов пока нет.{' '}
          <Link to="/posts/new" className={linkClass}>
            Создать первый пост
          </Link>
        </Card>
      ) : (
        <motion.ul
          variants={stagger}
          initial="hidden"
          animate="show"
          className="mt-6 grid gap-3"
        >
          {posts.map((post) => (
            <PostCard key={post.id} post={post} />
          ))}
        </motion.ul>
      )}

      {query.hasNextPage && (
        <div className="mt-6">
          <GhostButton
            disabled={query.isFetchingNextPage}
            onClick={() => void query.fetchNextPage()}
          >
            {query.isFetchingNextPage ? 'Загрузка…' : 'Показать ещё'}
          </GhostButton>
        </div>
      )}
    </section>
  )
}
