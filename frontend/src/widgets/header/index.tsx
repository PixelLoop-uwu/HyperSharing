
import { motion } from 'framer-motion'
import { useMe } from '@/features/user'
import Card from '@/shared/ui/layout/Card'
import Logo from '@/shared/ui/content/Logo'
import DesktopNav from './DesktopNav'
import MobileMenu from './MobileMenu'

export default function Header() {
  const me = useMe()
  const loading = me.isLoading
  const user = me.data ?? null

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

          <DesktopNav loading={loading} user={user} />
          <MobileMenu loading={loading} user={user} />
        </div>
      </Card>
    </motion.header>
  )
}
