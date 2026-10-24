// features/posts/ui/PostCard.tsx
import { FiFolder } from 'react-icons/fi'
import { Link } from 'react-router-dom'
import { formatBytes, formatDate } from '@/shared/lib/format'
import Card from '@/shared/ui/layout/Card'
import type { PostListItem } from '../types'

export default function PostCard({ post }: { post: PostListItem }) {
  return (
    <li>
      <Link
        to={`/posts/${post.id}`}
        className="block rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500"
      >
        <Card className="flex flex-col gap-1.5 p-5 transition-colors hover:bg-zinc-800/50 sm:flex-row sm:items-center sm:gap-4">
          <div className="flex min-w-0 flex-1 items-center gap-2">
            <FiFolder aria-hidden="true" className="size-4 shrink-0 text-zinc-400" />
            <h3 className="min-w-0 truncate font-display text-base font-semibold text-zinc-100">
              {post.title || 'Без названия'}
            </h3>
          </div>
          <p className="shrink-0 pl-6 text-sm tabular-nums text-zinc-400 sm:pl-0">
            {formatBytes(post.total_size)} · {formatDate(post.created_at)}
          </p>
        </Card>
      </Link>
    </li>
  )
}