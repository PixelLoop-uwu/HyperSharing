import { useQuery } from '@tanstack/react-query'
import { MotionConfig } from 'framer-motion'
import type { ReactNode } from 'react'
import { FiArrowLeft } from 'react-icons/fi'
import { Link, useParams } from 'react-router-dom'
import { getPost } from '@/features/posts/api'
import PostView from '@/features/posts/ui/PostView'
import Card from '@/shared/ui/layout/Card'
import GhostButton from '@/shared/ui/actions/GhostButton'
import StatusMark from '@/shared/ui/feedback/StatusMark'

function BackLink() {
  return (
    <Link
      to="/posts"
      className="inline-flex items-center gap-1.5 text-sm text-zinc-400 transition-colors hover:text-zinc-100 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-zinc-400"
    >
      <FiArrowLeft aria-hidden="true" className="size-4" />
      Все посты
    </Link>
  )
}

function Shell({ children }: { children: ReactNode }) {
  return (
    <MotionConfig reducedMotion="user">
      <div className="mx-auto w-full max-w-2xl pb-24 pt-10">
        <BackLink />
        <div className="mt-8">{children}</div>
      </div>
    </MotionConfig>
  )
}

export default function PostPage() {
  const { postId } = useParams<{ postId: string }>()
  const query = useQuery({
    queryKey: ['post', postId],
    queryFn: () => getPost(postId!),
    enabled: Boolean(postId),
  })

  if (query.isLoading) {
    return (
      <Shell>
        <p className="flex items-center gap-3 py-10 text-sm text-zinc-400" aria-live="polite">
          <StatusMark status="running" size={16} />
          Загружаем пост
        </p>
      </Shell>
    )
  }

  if (query.isError || !query.data) {
    return (
      <Shell>
        <Card className="p-5">
          <p role="alert" className="text-red-400">
            {query.error instanceof Error ? query.error.message : 'Не удалось загрузить пост'}
          </p>
          <GhostButton className="mt-3 -ml-2.5" onClick={() => void query.refetch()}>
            Повторить
          </GhostButton>
        </Card>
      </Shell>
    )
  }

  return (
    <Shell>
      <PostView post={query.data} />
    </Shell>
  )
}
