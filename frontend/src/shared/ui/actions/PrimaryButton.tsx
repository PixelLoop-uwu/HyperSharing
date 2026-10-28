import { motion, useReducedMotion } from 'framer-motion'
import { Link } from 'react-router-dom'
import type { ReactNode } from 'react'
import StatusMark from '../feedback/StatusMark'

/** Filled button with an animated diagonal shine (the hero CTA style). */
function Shine() {
  const reduce = useReducedMotion()
  if (reduce) return null
  return (
    <motion.span
      aria-hidden="true"
      className="pointer-events-none absolute inset-y-0 left-0 w-1/3 bg-linear-to-r from-transparent via-zinc-950/15 to-transparent"
      initial={{ x: '-150%', skewX: -18 }}
      animate={{ x: ['-150%', '450%'], skewX: -18 }}
      transition={{ duration: 1, repeat: Infinity, repeatDelay: 2.4, ease: 'easeInOut' }}
    />
  )
}

interface PrimaryButtonProps {
  children: ReactNode
  icon?: ReactNode
  /** Internal link. */
  to?: string
  /** External link. */
  href?: string
  /** With neither `to` nor `href` this renders a real <button>. */
  onClick?: () => void
  /** Button type; use "submit" inside a form. */
  type?: 'button' | 'submit'
  /** Swaps the icon for a spinner and ignores clicks. The label stays, so the width doesn't jump. */
  pending?: boolean
  className?: string
  shine?: boolean
  size?: 'default' | 'sm'
  disabled?: boolean
}

export default function PrimaryButton({
  children,
  icon,
  to,
  href,
  onClick,
  type = 'button',
  pending = false,
  className = '',
  shine = false,
  size = 'default',
  disabled = false,
}: PrimaryButtonProps) {
  const cls = [
    'relative inline-flex items-center gap-2 overflow-hidden rounded-lg bg-zinc-100 px-5 text-sm font-medium text-zinc-950',
    size === 'sm' ? 'py-2' : 'py-2.5',
    'transition-transform hover:bg-zinc-50 active:scale-[0.98]',
    'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-zinc-100',
    className,
  ]
    .filter(Boolean)
    .join(' ')

  const content = (
    <>
      {shine && <Shine />}
      {pending ? <StatusMark status="running" size={16} /> : icon}
      {children}
    </>
  )

  if (to) {
    return (
      <Link to={to} className={cls}>
        {content}
      </Link>
    )
  }
  if (href) {
    return (
      <a href={href} className={cls}>
        {content}
      </a>
    )
  }
  return (
    <button
      type={type}
      aria-busy={pending || undefined}
      aria-disabled={pending || undefined}
      onClick={(e) => {
        if (pending) {
          e.preventDefault()
          return
        }
        onClick?.()
      }}
      className={cls}
      disabled={disabled}
    >
      {content}
    </button>
  )
}
