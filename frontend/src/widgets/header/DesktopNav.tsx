import { FiGithub, FiLogIn, FiPlus } from 'react-icons/fi'
import { UserMenu, type UserRead } from '@/features/user'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'
import GhostLink from '@/shared/ui/actions/GhostLink'

interface AuthProps {
  loading: boolean
  user: UserRead | null
}

const GITHUB_URL = 'https://github.com/PixelLoop-uwu/hyperSharing'

export default function DesktopNav({ loading, user }: AuthProps) {
  return (
    <nav className="hidden items-center gap-2 text-sm text-zinc-400 md:flex">
      <GhostLink
        to="/new"
        className="inline-flex items-center gap-1.5 transition-colors hover:text-zinc-100"
      >
        <FiPlus className="size-4" />
        Создать
      </GhostLink>

      <GhostLink
        to={GITHUB_URL}
        target="_blank"
        rel="noreferrer"
        className="inline-flex items-center gap-1.5 transition-colors hover:text-zinc-100"
      >
        <FiGithub className="size-4" />
        GitHub
      </GhostLink>

      {!loading && user && <UserMenu />}

      {!loading && !user && (
        <PrimaryButton
          to="/login"
          size="sm"
          icon={<FiLogIn className="size-4" />}
        >
          Войти
        </PrimaryButton>
      )}
    </nav>
  )
}