import { AnimatePresence, motion } from 'framer-motion'

/** Form-level error that slides in and out. Announced to screen readers when it appears. */
export default function FormError({ message }: { message: string | null }) {
  return (
    <AnimatePresence initial={false}>
      {message && (
        <motion.p
          key={message}
          role="alert"
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: 'auto' }}
          exit={{ opacity: 0, height: 0 }}
          transition={{ duration: 0.2, ease: [0.23, 1, 0.32, 1] }}
          className="overflow-hidden text-sm text-red-400"
        >
          {message}
        </motion.p>
      )}
    </AnimatePresence>
  )
}
