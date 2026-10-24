import { useState, useEffect, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { FiGithub, FiLogIn, FiPlus, FiMenu, FiX } from 'react-icons/fi'
import { Link } from 'react-router-dom'
import { useAuth } from '@/features/user/store'
import UserMenu from '@/features/user/ui/UserMenu'
import Card from '@/shared/ui/layout/Card'
import Logo from '@/shared/ui/content/Logo'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'

export default function Header() {
  const [isOpen, setIsOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)
  const buttonRef = useRef<HTMLButtonElement>(null)

  const status = useAuth((s) => s.status)
  const user = useAuth((s) => s.user)
  const logout = useAuth((s) => s.logout)

  const handleLogout = () => {
    void logout()
    setIsOpen(false)
  }

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        isOpen &&
        menuRef.current &&
        !menuRef.current.contains(event.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen])

  return (
    <motion.header
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
      className="fixed inset-x-0 top-0 z-40 px-(--header-gap) pt-(--header-gap)"
    >
      <Card variant="glass" className="mx-auto max-w-7xl">
        <div className="flex h-(--header-h) items-center justify-between px-6">
          <Logo />

          <nav className="hidden md:flex items-center gap-6 text-sm text-zinc-400">
            <Link
              to="/posts/new"
              className="inline-flex items-center gap-1.5 transition-colors hover:text-zinc-100"
            >
              <FiPlus className="size-4" />
              Создать
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

            {status === 'loading' ? null : user ? <UserMenu /> : null}
            
            {status === 'loading' || user ? null : (
              <PrimaryButton to="/login" size="sm" icon={<FiLogIn className="size-4" />}>
                Войти
              </PrimaryButton>
            )}
          </nav>

          {/* Кнопка гамбургер-меню для мобилок */}
          <button
            ref={buttonRef}
            type="button"
            onClick={() => setIsOpen(!isOpen)}
            className="md:hidden text-zinc-400 hover:text-zinc-100 transition-colors"
            aria-label="Открыть меню"
          >
            {isOpen ? <FiX className="size-6" /> : <FiMenu className="size-6" />}
          </button>
        </div>
      </Card>

      {/* Выпадающее мобильное меню */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            ref={menuRef}
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
            className="md:hidden mx-auto mt-2 max-w-7xl rounded-2xl border border-zinc-800/80 bg-zinc-950/60 p-6 backdrop-blur-md shadow-xl"
          >
            <nav className="flex flex-col gap-4 text-sm text-zinc-400">
              <Link
                to="/posts/new"
                onClick={() => setIsOpen(false)}
                className="inline-flex items-center gap-2 transition-colors hover:text-zinc-100"
              >
                <FiPlus className="size-4" />
                Новый пост
              </Link>
              {status === 'loading' ? null : user ? (
                <Link
                  to="/me"
                  onClick={() => setIsOpen(false)}
                  className="transition-colors hover:text-zinc-100"
                >
                  Профиль
                </Link>
              ) : null}
              <a
                href="https://github.com/PixelLoop-uwu/hyperSharing"
                target="_blank"
                rel="noreferrer"
                onClick={() => setIsOpen(false)}
                className="inline-flex items-center gap-2 transition-colors hover:text-zinc-100"
              >
                <FiGithub className="size-4" />
                GitHub
              </a>
              {status === 'loading' ? null : user ? (
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
                    onClick={() => setIsOpen(false)}
                  >
                    Войти
                  </PrimaryButton>
                </div>
              )}
            </nav>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.header>
  )
}