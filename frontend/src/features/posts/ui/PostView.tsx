import { MotionConfig, motion } from 'framer-motion'
import { FiDownload } from 'react-icons/fi'
import { formatBytes, formatDate } from '@/shared/lib/format'
import { rise, stagger } from '@/shared/lib/motion'
import Card from '@/shared/ui/layout/Card'
import CopyButton from '@/shared/ui/actions/CopyButton'
import FileTypeIcon from '@/shared/ui/content/FileTypeIcon'
import Markdown from '@/shared/ui/content/Markdown'
import type { FileRead, PostRead, SectionRead } from '../models/types'

type MediaKind = 'image' | 'video' | 'audio'

function mediaKind(mime: string): MediaKind | null {
  if (mime.startsWith('image/')) return 'image'
  if (mime.startsWith('video/')) return 'video'
  if (mime.startsWith('audio/')) return 'audio'
  return null
}

// Media renders inline (a subresource fetch, so the address bar never navigates to the
// file URL). Everything else downloads via `<a download>`, which never navigates.
function MediaView({ media, file }: { media: MediaKind; file: FileRead }) {
  if (media === 'image') {
    return (
      <img
        src={file.url}
        alt={file.name}
        loading="lazy"
        className="max-h-128 w-full rounded-2xl border border-zinc-800 bg-zinc-900 object-contain"
      />
    )
  }
  if (media === 'video') {
    return (
      <video
        src={file.url}
        controls
        className="w-full rounded-2xl border border-zinc-800 bg-zinc-950"
      />
    )
  }
  return <audio src={file.url} controls className="w-full" />
}

function FileList({ files }: { files: FileRead[] }) {
  return (
    <Card className="overflow-hidden">
      <ul className="divide-y divide-zinc-800">
        {files.map((file, i) => (
          <li key={`${file.url}-${i}`}>
            <a
              href={file.url}
              download={file.name}
              className="flex items-center gap-3 px-4 py-3.5 transition-colors hover:bg-zinc-800/50 focus-visible:bg-zinc-800/50 focus-visible:outline-none"
            >
              <FileTypeIcon
                type={file.mime}
                name={file.name}
                className="size-5 shrink-0 text-zinc-300"
              />
              <span className="min-w-0 flex-1">
                <span className="block truncate text-sm text-zinc-100">{file.name}</span>
                <span className="mt-0.5 block text-xs tabular-nums text-zinc-400">
                  {formatBytes(file.size)}
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

/** Body first, then media inline and downloadable files as a list. */
function SectionView({ section }: { section: SectionRead }) {
  const media = section.files.filter((f) => mediaKind(f.mime) !== null)
  const rest = section.files.filter((f) => mediaKind(f.mime) === null)
  return (
    <div className="space-y-6">
      {section.body.trim() !== '' && <Markdown>{section.body}</Markdown>}
      {media.map((file, i) => (
        <MediaView key={i} media={mediaKind(file.mime)!} file={file} />
      ))}
      {rest.length > 0 && <FileList files={rest} />}
    </div>
  )
}

export default function PostView({ post }: { post: PostRead }) {
  return (
    <MotionConfig reducedMotion="user">
      <motion.article
        variants={stagger}
        initial="hidden"
        animate="show"
        className="space-y-10"
      >
        <motion.header variants={rise} className="space-y-3">
          <h1 className="font-display text-3xl font-semibold tracking-tight text-zinc-100 sm:text-4xl">
            {post.title || 'Без названия'}
          </h1>
          <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
            <div className="flex items-center gap-2">
              {post.author?.avatar_url && (
                <img
                  src={post.author.avatar_url}
                  alt=""
                  className="size-6 shrink-0 rounded-full bg-zinc-800 object-cover"
                />
              )}
              <p className="text-sm text-zinc-400">
                {post.author ? `${post.author.username} · ` : ''}
                {formatDate(post.created_at)}
              </p>
            </div>
            <CopyButton text={() => window.location.href} />
          </div>
        </motion.header>

        <motion.div variants={rise} className="space-y-10">
          {post.sections.map((section, i) => (
            <SectionView key={i} section={section} />
          ))}
        </motion.div>
      </motion.article>
    </MotionConfig>
  )
}
