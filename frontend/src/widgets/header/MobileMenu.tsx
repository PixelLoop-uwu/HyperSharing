import { useEffect, useRef, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { FiGithub, FiLogIn, FiMenu, FiPlus, FiX } from 'react-icons/fi'
import { Link } from 'react-router-dom'
import { useLogout, type UserRead } from '@/features/user'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'

interface AuthProps {
  loading: boolean
  user: UserRead | null
}

const GITHUB_URL = 'https://github.com/PixelLoop-uwu/hyperSharing'

export default function MobileMenu({ loading, user }: AuthProps) {
  const [isOpen, setIsOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)
  const buttonRef = useRef<HTMLButtonElement>(null)

  const logout = useLogout()

  useEffect(() => {
    if (!isOpen) return

    const handlePointerDown = (event: PointerEvent) => {
      const target = event.target

      if (!(target instanceof Node)) return

      if (
        !menuRef.current?.contains(target) &&
        !buttonRef.current?.contains(target)
      ) {
        setIsOpen(false)
      }
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsOpen(false)
      }
    }

    document.addEventListener('pointerdown', handlePointerDown)
    document.addEventListener('keydown', handleKeyDown)

    return () => {
      document.removeEventListener('pointerdown', handlePointerDown)
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [isOpen])

  const closeMenu = () => setIsOpen(false)

  const handleLogout = () => {
    closeMenu()
    logout.mutate()
  }

  return (
    <>
      <button
        ref={buttonRef}
        type="button"
        onClick={() => setIsOpen((open) => !open)}
        className="text-zinc-400 transition-colors hover:text-zinc-100 md:hidden"
        aria-label={isOpen ? 'Закрыть меню' : 'Открыть меню'}
        aria-expanded={isOpen}
        aria-controls="mobile-navigation"
      >
        {isOpen
          ? <FiX className="size-6" />
          : <FiMenu className="size-6" />}
      </button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            ref={menuRef}
            id="mobile-navigation"
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
            className="absolute inset-x-(--header-gap) top-full mx-auto mt-2 max-w-7xl rounded-2xl border border-zinc-800/80 bg-zinc-950/60 p-6 shadow-xl backdrop-blur-md md:hidden"
          >
            <nav className="flex flex-col gap-4 text-sm text-zinc-400">
              <Link
                to="/new"
                onClick={closeMenu}
                className="inline-flex items-center gap-2 transition-colors hover:text-zinc-100"
              >
                <FiPlus className="size-4" />
                Новый пост
              </Link>

              {!loading && user && (
                <Link
                  to="/me"
                  onClick={closeMenu}
                  className="transition-colors hover:text-zinc-100"
                >
                  Профиль
                </Link>
              )}

              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noreferrer"
                onClick={closeMenu}
                className="inline-flex items-center gap-2 transition-colors hover:text-zinc-100"
              >
                <FiGithub className="size-4" />
                GitHub
              </a>

              {!loading && (
                user ? (
                  <button
                    type="button"
                    onClick={handleLogout}
                    className="text-left transition-colors hover:text-zinc-100"
                  >
                    Выйти
                  </button>
                ) : (
                  <div className="pt-2">
                    <PrimaryButton
                      to="/login"
                      size="sm"
                      icon={<FiLogIn className="size-4" />}
                      onClick={closeMenu}
                    >
                      Войти
                    </PrimaryButton>
                  </div>
                )
              )}
            </nav>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  )
}
