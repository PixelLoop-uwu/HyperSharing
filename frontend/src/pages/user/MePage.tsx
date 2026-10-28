import { MotionConfig, motion } from 'framer-motion'
import { PostsSection } from '@/features/posts'
import { ProfileSidebar, useMe } from '@/features/user'
import { rise, stagger } from '@/shared/lib/motion'

export default function MePage() {
  const user = useMe().data
  if (!user) return null

  return (
    <MotionConfig reducedMotion="user">
      <motion.div
        variants={stagger}
        initial="hidden"
        animate="show"
        className="mx-auto flex w-full max-w-6xl flex-col gap-8 py-10 lg:flex-row lg:items-start"
      >
        <motion.aside
          variants={rise}
          className="lg:sticky lg:top-[calc(var(--header-total-h)+1.5rem)] lg:w-80 lg:shrink-0 xl:w-85"
        >
          <ProfileSidebar user={user} />
        </motion.aside>
        <motion.div variants={rise} className="min-w-0 flex-1">
          <PostsSection />
        </motion.div>
      </motion.div>
    </MotionConfig>
  )
}
