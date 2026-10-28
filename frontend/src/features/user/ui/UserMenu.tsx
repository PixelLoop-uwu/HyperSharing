import { useState } from 'react'
import { FiLogOut } from 'react-icons/fi'
import { Link } from 'react-router-dom'
import Avatar from './Avatar'
import { useLogout, useMe } from '../model/hooks'

export default function UserMenu() {
  const user = useMe().data
  const logout = useLogout()
  const [pending, setPending] = useState(false)

  if (!user) return null

  async function handleLogout() {
    if (pending) return

    setPending(true)
    try {
      await logout.mutateAsync()
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="flex min-w-0 items-center">
      <Link
        to="/me"
        className="flex min-w-0 items-center gap-2 rounded-lg px-2 py-1.5 text-zinc-300 transition-colors hover:bg-white/5 hover:text-zinc-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500"
      >
        <Avatar
          userId={user.id}
          version={user.avatar_version}
          username={user.username}
          className="size-7 shrink-0 rounded-full"
          fallbackClassName="text-xs"
        />

        <span className="max-w-40 truncate text-sm">
          {user.username}
        </span>
      </Link>

      <button
        type="button"
        aria-label="Выйти"
        title="Выйти"
        disabled={pending}
        onClick={() => void handleLogout()}
        className="grid size-8 shrink-0 place-items-center rounded-lg text-zinc-500 transition-colors hover:bg-red-500/10 hover:text-red-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500 disabled:cursor-wait disabled:opacity-50"
      >
        <FiLogOut
          aria-hidden="true"
          className={`size-4 ${pending ? 'animate-pulse' : ''}`}
        />
      </button>
    </div>
  )
}