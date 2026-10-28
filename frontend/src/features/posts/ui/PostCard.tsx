// features/posts/ui/PostCard.tsx
import { motion } from 'framer-motion'
import { FaRegFolder } from "react-icons/fa";
import { Link } from 'react-router-dom'
import { formatBytes, formatDate } from '@/shared/lib/format'
import { rise } from '@/shared/lib/motion'
import Card from '@/shared/ui/layout/Card'
import type { PostListItem } from '../models/types'
import DeletePostButton from './DeletePostButton'


export default function PostCard({ post }: { post: PostListItem }) {
  return (
    <motion.li variants={rise}>
      <Card className="relative flex items-center gap-4 p-4 transition-colors hover:bg-zinc-800/50">
        <Link
          to={`/${post.short_key}`}
          className="flex min-w-0 flex-1 flex-col gap-1.5 focus-visible:outline-none sm:flex-row sm:items-center sm:gap-4
                     after:absolute after:inset-0 after:rounded-xl after:content-['']
                     focus-visible:after:ring-2 focus-visible:after:ring-zinc-500"
        >
          <span className="flex min-w-0 flex-1 items-center gap-2">
            <FaRegFolder aria-hidden="true" className="size-5 shrink-0 text-zinc-400 -mb-px" />
            <h3 className="min-w-0 truncate font-display text-base font-semibold text-zinc-100">
              {post.title || 'Без названия'}
            </h3>
          </span>
          <span className="shrink-0 pl-6 text-sm tabular-nums text-zinc-400 sm:pl-0">
            {formatBytes(post.total_size)} · {formatDate(post.created_at)}
          </span>
        </Link>

        <div className="relative z-10 shrink-0">
          <DeletePostButton postId={post.id} />
        </div>
      </Card>
    </motion.li>
  )
}