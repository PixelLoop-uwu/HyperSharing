import { Outlet } from 'react-router-dom'
import Header from '@/widgets/Header'
import Footer from '@/widgets/Footer'

export default function Layout() {
  return (
    <div className="relative flex min-h-svh flex-col overflow-x-clip font-sans text-paper">
      <Header />
      <main className="mx-auto w-full max-w-7xl flex-1 flex flex-col px-6 pt-(--header-total-h)">
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}
