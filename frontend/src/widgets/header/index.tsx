import { useState, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { FiMenu, FiX } from 'react-icons/fi'
import Card from '@/shared/ui/layout/Card'
import Logo from '@/shared/ui/content/Logo'
import DesktopNav from './DesktopNav'
import MobileNav from './MobileNav'

export default function Header() {
  const [isOpen, setIsOpen] = useState(false)
  const buttonRef = useRef<HTMLButtonElement>(null)

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
          <DesktopNav />

          <button
            ref={buttonRef}
            type="button"
            onClick={() => setIsOpen((prev) => !prev)}
            className="md:hidden relative size-6 text-zinc-400 hover:text-zinc-100 transition-colors"
            aria-label="Открыть меню"
          >
            <AnimatePresence mode="wait" initial={false}>
              {isOpen ? (
                <motion.div
                  key="close"
                  initial={{ rotate: -90, opacity: 0 }}
                  animate={{ rotate: 0, opacity: 1 }}
                  exit={{ rotate: 90, opacity: 0 }}
                  transition={{ duration: 0.2 }}
                  className="absolute inset-0 flex items-center justify-center"
                >
                  <FiX className="size-6" />
                </motion.div>
              ) : (
                <motion.div
                  key="menu"
                  initial={{ rotate: 90, opacity: 0 }}
                  animate={{ rotate: 0, opacity: 1 }}
                  exit={{ rotate: -90, opacity: 0 }}
                  transition={{ duration: 0.2 }}
                  className="absolute inset-0 flex items-center justify-center"
                >
                  <FiMenu className="size-6" />
                </motion.div>
              )}
            </AnimatePresence>
          </button>
        </div>
      </Card>

      <MobileNav
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        toggleButtonRef={buttonRef}
      />
    </motion.header>
  )
}