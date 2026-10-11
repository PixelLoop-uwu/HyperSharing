import { useEffect, useState } from 'react'
import { motion, useReducedMotion } from 'framer-motion'

interface TypewriterProps {
  text: string
  speed?: number
  startDelay?: number
  className?: string
}

/** Types ``text`` character by character with a blinking block cursor. */
export default function Typewriter({
  text,
  speed = 45,
  startDelay = 300,
  className,
}: TypewriterProps) {
  const reduce = useReducedMotion()
  const [count, setCount] = useState(0)
  const done = count >= text.length

  useEffect(() => {
    if (reduce) {
      setCount(text.length)
      return
    }
    if (done) return
    const delay = count === 0 ? startDelay : speed
    const id = setTimeout(() => setCount((c) => Math.min(c + 1, text.length)), delay)
    return () => clearTimeout(id)
  }, [count, reduce, done, text.length, speed, startDelay])

  return (
    <span className={className}>
      {text.slice(0, count)}
      {!done && (
        <motion.span
          aria-hidden="true"
          className="ml-1 inline-block h-[0.9em] w-[0.5em] translate-y-[0.1em] bg-paper/80"
          animate={{ opacity: [1, 0, 1] }}
          transition={{ duration: 1.8, repeat: Infinity, ease: 'easeInOut' }}
        />
      )}
    </span>
  )
}
