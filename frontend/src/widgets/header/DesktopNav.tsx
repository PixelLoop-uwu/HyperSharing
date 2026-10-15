import { FiGithub, FiLogIn, FiPlus } from 'react-icons/fi'
import { Link } from 'react-router-dom'
import { useAuth } from '@/features/user/store'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'

export default function DesktopNav() {
  const status = useAuth((s) => s.status)
  const user = useAuth((s) => s.user)
  const logout = useAuth((s) => s.logout)

  return (
    <nav className="hidden md:flex items-center gap-6 text-sm text-zinc-400">
      <Link
        to="/posts/new"
        className="inline-flex items-center gap-1.5 transition-colors hover:text-zinc-100"
      >
        <FiPlus className="size-4" />
        Новый пост
      </Link>
      <a
        href="https://github.com/PixelLoop-uwu/hyperSharing"
        target="_blank"
        rel="noreferrer"
        className="inline-flex items-center gap-1.5 transition-colors hover:text-zinc-100"
      >
        <FiGithub className="size-4" />
        GitHub
      </a>
      {status === 'loading' ? null : user ? (
        <button
          type="button"
          onClick={() => void logout()}
          className="transition-colors hover:text-zinc-100"
        >
          Выйти
        </button>
      ) : (
        <PrimaryButton to="/login" size="sm" icon={<FiLogIn className="size-4" />}>
          Войти
        </PrimaryButton>
      )}
    </nav>
  )
}