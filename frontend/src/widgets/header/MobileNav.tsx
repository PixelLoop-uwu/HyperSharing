import {  useEffect, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { FiGithub, FiLogIn, FiPlus } from 'react-icons/fi'
import { Link } from 'react-router-dom'
import { useAuth } from '@/features/auth/store'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'

interface MobileNavProps {
  isOpen: boolean
  onClose: () => void
  toggleButtonRef: React.RefObject<HTMLButtonElement | null>
}

export default function MobileNav({ isOpen, onClose, toggleButtonRef }: MobileNavProps) {
  const menuRef = useRef<HTMLDivElement>(null)
  const status = useAuth((s) => s.status)
  const user = useAuth((s) => s.user)
  const logout = useAuth((s) => s.logout)

  const handleLogout = () => {
    void logout()
    onClose()
  }

  useEffect(() => {
    if (!isOpen) return

    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node
      const isInsideMenu = menuRef.current?.contains(target)
      const isInsideButton = toggleButtonRef.current?.contains(target)

      if (!isInsideMenu && !isInsideButton) {
        onClose()
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen, onClose, toggleButtonRef])

  return (
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
              onClick={onClose}
              className="inline-flex items-center gap-2 transition-colors hover:text-zinc-100"
            >
              <FiPlus className="size-4" />
              Новый пост
            </Link>
            <a
              href="https://github.com/PixelLoop-uwu/hyperSharing"
              target="_blank"
              rel="noreferrer"
              onClick={onClose}
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
                  onClick={onClose}
                >
                  Войти
                </PrimaryButton>
              </div>
            )}
          </nav>
        </motion.div>
      )}
    </AnimatePresence>
  )
}