import type { ComponentPropsWithoutRef } from 'react'

type Variant = 'glass' | 'solid'

interface CardProps extends ComponentPropsWithoutRef<'div'> {
  /**
   * glass — полупрозрачная с размытием: для мест над анимированным фоном (шапка, Hero).
   * solid — непрозрачная: для мест со статичным фоном (создание поста и т.п.).
   */
  variant?: Variant
}

const variants: Record<Variant, string> = {
  glass: 'rounded-2xl border border-zinc-800/60 bg-zinc-900/20 backdrop-blur-2xl',
  solid: 'rounded-2xl border border-zinc-800 bg-zinc-900/20',
}

export default function Card({ variant = 'solid', className = '', ...props }: CardProps) {
  return <div className={`${variants[variant]} ${className}`.trim()} {...props} />
}