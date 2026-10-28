import { FiCalendar, FiTrash2, FiUpload } from 'react-icons/fi'
import { QuotaSection } from '@/features/posts'
import Avatar from './Avatar'
import Card from '@/shared/ui/layout/Card'
import { useAvatarEditor } from '../model/useAvatarEditor'
import type { UserRead } from '../model/types'

export default function ProfileSidebar({ user }: { user: UserRead }) {
  const editor = useAvatarEditor()

  const joined = new Date(user.created_at).toLocaleDateString('ru-RU', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })

  return (
    <Card variant="solid" className="p-4">
      <div className="relative">
        <Avatar
          userId={user.id}
          username={user.username}
          version={user.avatar_version}
          className="aspect-square w-full rounded-full"
        />
        <div className="absolute bottom-3 right-3 flex items-center overflow-hidden bg-zinc-700/80 backdrop-blur-xl rounded-lg border border-zinc-500/50">
          <button
            type="button"
            aria-label={editor.hasAvatar ? 'Заменить аватар' : 'Загрузить аватар'}
            disabled={editor.pending}
            onClick={editor.picker.open}
            className="grid size-8 place-items-center text-zinc-100 transition-colors hover:bg-zinc-700/60 disabled:opacity-60"
          >
            <FiUpload className="size-4" />
          </button>
          {editor.hasAvatar && (
            <>
              <span aria-hidden="true" className="h-4 w-px shrink-0 bg-zinc-500/50" />
              <button
                type="button"
                aria-label="Удалить аватар"
                disabled={editor.pending}
                onClick={editor.removeAvatar}
                className="grid size-8 place-items-center text-zinc-100 transition-colors hover:bg-red-500 disabled:opacity-60"
              >
                <FiTrash2 className="size-4" />
              </button>
            </>
          )}
        </div>
      </div>

      <h1 className="mt-4 font-display text-2xl font-semibold tracking-tight text-zinc-100">
        {user.username}
      </h1>
      <p className="mt-1 truncate text-sm text-zinc-400" title={user.email}>
        {user.email}
      </p>
      <p className="mt-3 flex items-center gap-1.5 text-sm text-zinc-400">
        <FiCalendar aria-hidden="true" className="size-4" />
        Присоединился {joined}
      </p>
      {editor.picker.input}

      {editor.error && (
        <p role="alert" className="mt-4 text-sm text-red-400">
          {editor.error}
        </p>
      )}

      <QuotaSection />
    </Card>
  )
}
