import { MotionConfig, motion } from 'framer-motion'
import type { ReactNode } from 'react'
import { rise, stagger } from '@/shared/lib/motion'
import Card from '@/shared/ui/layout/Card'

interface AuthShellProps {
  title: string
  lead: string
  children: ReactNode
  footer: ReactNode
}

export default function AuthShell({ title, lead, children, footer }: AuthShellProps) {
  return (
    <MotionConfig reducedMotion="user">
        <motion.div
          variants={stagger}
          initial="hidden"
          animate="show"
          className="mx-auto w-full flex-1 max-w-md flex flex-col items-stretch justify-center"
        >
          <motion.div variants={rise}>
            <Card variant="solid" className="p-6 sm:p-8">
              <h1 className="font-display text-3xl font-semibold tracking-tight text-zinc-100">
                {title}
              </h1>
              <p className="mt-2 text-sm text-zinc-400">{lead}</p>
              {children}
            </Card>
          </motion.div>
          <motion.p variants={rise} className="mt-6 text-center text-sm text-zinc-400">
            {footer}
          </motion.p>
        </motion.div>
    </MotionConfig>
  )
}
