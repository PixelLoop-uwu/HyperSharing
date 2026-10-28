import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MotionConfig } from 'framer-motion'
import { RouterProvider } from 'react-router-dom'
import { router } from '@/app/routes'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      staleTime: 30_000,
    },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <MotionConfig reducedMotion="user">
        <RouterProvider router={router} />
      </MotionConfig>
    </QueryClientProvider>
  )
}
