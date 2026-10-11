import type { Variants } from 'framer-motion'

/** Shared easing and motion presets so feature pages animate consistently. */
export const EASE: [number, number, number, number] = [0.16, 1, 0.3, 1]

export const stagger: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.09, delayChildren: 0.05 } },
}

export const rise: Variants = {
  hidden: { opacity: 0, y: 18 },
  show: { opacity: 1, y: 0, transition: { duration: 0.55, ease: EASE } },
}
