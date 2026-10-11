import { motion } from 'framer-motion'
import { FiDownload } from 'react-icons/fi'
import { formatBytes, formatDate } from '@/shared/lib/format'
import Card from '@/shared/ui/layout/Card'
import CopyButton from '@/shared/ui/actions/CopyButton'
import FileTypeIcon from '@/shared/ui/content/FileTypeIcon'
import Markdown from '@/shared/ui/content/Markdown'
import type { FileItemOut, PostFullRead } from '../types'

type MediaKind = 'image' | 'video' | 'audio'

function mediaKind(contentType: string | null): MediaKind | null {
  if (!contentType) return null
  if (contentType.startsWith('image/')) return 'image'
  if (contentType.startsWith('video/')) return 'video'
  if (contentType.startsWith('audio/')) return 'audio'
  return null
}

type Block =
  | { kind: 'text'; text: string }
  | { kind: 'media'; media: MediaKind; item: FileItemOut }
  | { kind: 'files'; items: FileItemOut[] }

/** Text and media stay in order; consecutive downloadable files collapse into one list. */
function toBlocks(items: PostFullRead['items']): Block[] {
  const blocks: Block[] = []
  for (const item of items) {
    if (item.kind === 'paragraph') {
      if (item.text.trim()) blocks.push({ kind: 'text', text: item.text })
      continue
    }
    const media = mediaKind(item.content_type)
    if (media) {
      blocks.push({ kind: 'media', media, item })
      continue
    }
    const last = blocks[blocks.length - 1]
    if (last?.kind === 'files') last.items.push(item)
    else blocks.push({ kind: 'files', items: [item] })
  }
  return blocks
}

// Media renders inline (a subresource fetch, so the address bar never navigates to the
// presigned MinIO URL). Everything else downloads via `<a download>`, which never navigates.
function MediaView({ media, item }: { media: MediaKind; item: FileItemOut }) {
  if (media === 'image') {
    return (
      <img
        src={item.url}
        alt={item.filename || item.hash}
        loading="lazy"
        className="max-h-128 w-full rounded-2xl border border-zinc-800 bg-zinc-900 object-contain"
      />
    )
  }
  if (media === 'video') {
    return (
      <video
        src={item.url}
        controls
        className="w-full rounded-2xl border border-zinc-800 bg-zinc-950"
      />
    )
  }
  return <audio src={item.url} controls className="w-full" />
}

function FileList({ items }: { items: FileItemOut[] }) {
  return (
    <Card className="overflow-hidden">
      <ul className="divide-y divide-zinc-800">
        {items.map((item, i) => (
          <li key={`${item.hash}-${i}`}>
            <a
              href={item.url}
              download={item.filename || ''}
              className="flex items-center gap-3 px-4 py-3.5 transition-colors hover:bg-zinc-800/50 focus-visible:bg-zinc-800/50 focus-visible:outline-none"
            >
              <FileTypeIcon
                type={item.content_type ?? ''}
                name={item.filename ?? ''}
                className="size-5 shrink-0 text-zinc-300"
              />
              <span className="min-w-0 flex-1">
                <span className="block truncate text-sm text-zinc-100">
                  {item.filename || item.hash}
                </span>
                <span className="mt-0.5 block text-xs tabular-nums text-zinc-400">
                  {formatBytes(item.size)}
                </span>
              </span>
              <span className="inline-flex shrink-0 items-center gap-1.5 text-sm font-medium text-zinc-300">
                <FiDownload aria-hidden="true" className="size-4" />
                Скачать
              </span>
            </a>
          </li>
        ))}
      </ul>
    </Card>
  )
}

export default function PostView({ post }: { post: PostFullRead }) {
  const blocks = toBlocks(post.items)

  return (
    <motion.article
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, ease: [0.23, 1, 0.32, 1] }}
      className="space-y-10"
    >
      <header className="space-y-3">
        <h1 className="wrap-break-word` font-display text-3xl font-semibold tracking-tight text-zinc-100 sm:text-4xl">
          {post.title || 'Без названия'}
        </h1>
        <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
          <p className="text-sm text-zinc-400">{formatDate(post.created_at)}</p>
          <CopyButton text={() => window.location.href} />
        </div>
      </header>

      <div className="space-y-6">
        {blocks.map((block, i) =>
          block.kind === 'text' ? (
            <Markdown key={i}>{block.text}</Markdown>
          ) : block.kind === 'media' ? (
            <MediaView key={i} media={block.media} item={block.item} />
          ) : (
            <FileList key={i} items={block.items} />
          ),
        )}
      </div>
    </motion.article>
  )
}
