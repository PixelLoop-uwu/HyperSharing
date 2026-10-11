import type { ComponentPropsWithoutRef, ReactNode } from 'react'

interface GhostButtonProps extends Omit<ComponentPropsWithoutRef<'button'>, 'type'> {
  icon?: ReactNode
  /** Square padding for buttons that hold only an icon (give it an aria-label). */
  iconOnly?: boolean
}

export default function GhostButton({
  icon,
  iconOnly = false,
  className = '',
  children,
  ...props
}: GhostButtonProps) {
  const cls = [
    'inline-flex items-center gap-1.5 rounded-lg text-sm font-medium text-zinc-400 transition-colors',
    'hover:bg-zinc-800/70 hover:text-zinc-100',
    'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-zinc-400',
    iconOnly ? 'p-1.5' : 'px-2.5 py-1.5',
    className,
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <button type="button" className={cls} {...props}>
      {icon}
      {children}
    </button>
  )
}
