import { Outlet } from 'react-router-dom'
import Header from '@/widgets/header'
import Footer from '@/widgets/Footer'

import { ConfirmProvider } from '@/shared/ui/confirm/confirmProvider'

export default function Layout() {
  return (
    <div className="relative flex min-h-svh flex-col overflow-x-clip font-sans text-paper">
      <Header />
      <ConfirmProvider>
        <main className="mx-auto w-full max-w-7xl flex-1 flex flex-col px-6 pt-(--header-total-h)">
          <Outlet />
        </main>
      </ConfirmProvider>
      <Footer />
    </div>
  )
}
