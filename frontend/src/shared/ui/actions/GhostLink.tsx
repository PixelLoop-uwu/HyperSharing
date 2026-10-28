import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import type { LinkProps } from 'react-router-dom'

interface GhostLinkProps extends Omit<LinkProps, 'className'> {
  icon?: ReactNode
  iconOnly?: boolean
  className?: string
  accentTextOnly?: boolean
}

export default function GhostLink({
  icon,
  iconOnly = false,
  className = '',
  children,
  accentTextOnly,
  ...props
}: GhostLinkProps) {
  const cls = [
    'inline-flex items-center gap-1.5 min-h-10 rounded-lg text-sm font-medium text-zinc-400 transition-colors',
    accentTextOnly
      ? 'hover:text-zinc-100'
      : 'hover:bg-zinc-800/70 hover:text-zinc-100',
    'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-zinc-400',
    iconOnly ? 'p-1.5' : 'px-2.5 py-1.5',
    className,
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <Link className={cls} {...props}>
      {icon}
      {children}
    </Link>
  )
}