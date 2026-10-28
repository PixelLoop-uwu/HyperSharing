import { MotionConfig, motion } from 'framer-motion'
import type { ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { SearchX, TriangleAlert, WifiOff } from 'lucide-react'
import { PostView, useViewPost } from '@/features/posts'
import { getStatus } from '@/shared/api/errors'
import { EASE } from '@/shared/lib/motion'
import Card from '@/shared/ui/layout/Card'
import GhostButton from '@/shared/ui/actions/GhostButton'
import Skeleton from '@/shared/ui/feedback/Skeleton'
import StatusMark from '@/shared/ui/feedback/StatusMark'
import { BackButton } from '@/shared/ui/actions/BackButton'

function Shell({ children }: { children: ReactNode }) {
  return (
    <MotionConfig reducedMotion="user">
      <div className="mx-auto w-full max-w-2xl pb-24 pt-10">
        <BackButton />
        <div className="mt-5">{children}</div>
      </div>
    </MotionConfig>
  )
}

function FadeIn({ children }: { children: ReactNode }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, ease: EASE }}
    >
      {children}
    </motion.div>
  )
}

/* ---------- загрузка ---------- */

function PostSkeleton() {
  return (
    <div role="status" aria-busy="true">
      <span className="sr-only">Загружаем пост</span>
      <Card className="p-6">
        <Skeleton className="h-6 w-2/3" />
        <div className="mt-3 flex items-center gap-3">
          <Skeleton className="size-6 rounded-full" />
          <Skeleton className="h-3.5 w-32" />
        </div>
        <div className="mt-6 space-y-2.5">
          <Skeleton className="h-3.5 w-full" />
          <Skeleton className="h-3.5 w-11/12" />
          <Skeleton className="h-3.5 w-4/6" />
        </div>
        <Skeleton className="mt-6 h-16 w-full rounded-xl" />
      </Card>
    </div>
  )
}

/* ---------- ошибки ---------- */

type ErrorKind = 'not-found' | 'offline' | 'generic'

function getErrorKind(error: unknown): ErrorKind {
  const status = getStatus(error)
  if (status === 404) return 'not-found'
  if (status === undefined && !navigator.onLine) return 'offline'
  return 'generic'
}

const ERROR_VIEW = {
  'not-found': {
    icon: SearchX,
    title: 'Пост не найден',
    text: 'Возможно, ссылка неверна или пост был удалён.',
  },
  offline: {
    icon: WifiOff,
    title: 'Нет соединения',
    text: 'Проверьте интернет и попробуйте ещё раз.',
  },
  generic: {
    icon: TriangleAlert,
    title: 'Не удалось загрузить пост',
    text: 'Что-то пошло не так. Попробуйте ещё раз чуть позже.',
  },
} as const

function PostError({
  error,
  retrying,
  onRetry,
}: {
  error: unknown
  retrying: boolean
  onRetry: () => void
}) {
  const kind = getErrorKind(error)
  const { icon: Icon, title, text } = ERROR_VIEW[kind]

  return (
    <Card className="p-6">
      <div role="alert" className="flex flex-col items-center text-center">
        <div className="mb-5 flex size-12 items-center justify-center rounded-2xl border border-zinc-800 bg-zinc-900/80 text-zinc-300">
          <Icon className="size-5" strokeWidth={1.7} />
        </div>
        <h2 className="text-xl font-medium tracking-tight text-zinc-100">{title}</h2>
        <p className="mt-2 text-sm leading-relaxed text-zinc-400">{text}</p>

        {kind !== 'not-found' && (
          <GhostButton className="mt-5" onClick={onRetry} disabled={retrying}>
            {retrying && <StatusMark status="running" size={14} />}
            {retrying ? 'Пробуем снова' : 'Повторить'}
          </GhostButton>
        )}
      </div>
    </Card>
  )
}

/* ---------- страница ---------- */

export default function PostPage() {
  const { shortKey } = useParams<{ shortKey: string }>()
  const query = useViewPost(shortKey)

  if (query.isLoading) {
    return (
      <Shell>
        <PostSkeleton />
      </Shell>
    )
  }

  if (query.isError || !query.data) {
    return (
      <Shell>
        <FadeIn>
          <PostError
            error={query.error}
            retrying={query.isFetching}
            onRetry={() => void query.refetch()}
          />
        </FadeIn>
      </Shell>
    )
  }

  return (
    <Shell>
      <FadeIn>
        <PostView post={query.data} />
      </FadeIn>
    </Shell>
  )
}