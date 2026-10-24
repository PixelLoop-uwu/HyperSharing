import { useState } from 'react'
import { FiLogOut } from 'react-icons/fi'
import { Link } from 'react-router-dom'
import Avatar from '@/shared/ui/content/Avatar'
import { useAuth } from '../store'

// Tailwind ищет классы по литералам, поэтому строка написана целиком, без сборки из частей
const NICK_FADE =
  'group-hover:[mask-image:linear-gradient(to_left,transparent_1rem,#000_2rem)] group-has-[:focus-visible]:[mask-image:linear-gradient(to_left,transparent_1rem,#000_2rem)]'

export default function UserMenu() {
  const user = useAuth((s) => s.user)
  const avatarVersion = useAuth((s) => s.avatarVersion)
  const logout = useAuth((s) => s.logout)
  const [pending, setPending] = useState(false)

  if (!user) return null

  async function handleLogout() {
    if (pending) return
    setPending(true)
    try {
      await logout()
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="group relative flex items-center">
      <Link
        to="/me"
        className="flex min-w-0 items-center gap-2 rounded-md transition-colors hover:text-zinc-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500"
      >
        <Avatar
          userId={user.id}
          username={user.username}
          version={avatarVersion}
          className="size-7 rounded-full"
          fallbackClassName="text-xs"
        />
        <span className={`max-w-[10rem] truncate ${NICK_FADE}`}>{user.username}</span>
      </Link>

      <button
        type="button"
        aria-label="Выйти"
        title="Выйти"
        disabled={pending}
        onClick={() => void handleLogout()}
        className="absolute inset-y-0 right-0 grid place-items-center rounded-md px-1.5 text-zinc-400 opacity-0 transition hover:text-zinc-100 focus-visible:opacity-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500 group-hover:opacity-100 group-has-[:focus-visible]:opacity-100 disabled:cursor-wait [@media(hover:none)]:static [@media(hover:none)]:opacity-100"
      >
        <FiLogOut aria-hidden="true" className={`size-4 ${pending ? 'animate-pulse' : ''}`} />
      </button>
    </div>
  )
}