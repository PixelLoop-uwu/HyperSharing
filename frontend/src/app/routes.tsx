import { createBrowserRouter } from 'react-router-dom'
import Layout from '@/app/Layout'
import UserOnly from '@/app/UserOnly'
import MePageSkeleton from '@/features/user/ui/MePageSkeleton'
import LandingPage from '@/pages/LandingPage'
import LoginPage from '@/pages/LoginPage'
import MePage from '@/pages/MePage'
import NewPostPage from '@/pages/NewPostPage'
import NotFoundPage from '@/pages/NotFound'
import PostPage from '@/pages/PostPage'
import RegisterPage from '@/pages/RegisterPage'

export const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      { path: '/', element: <LandingPage /> },
      { path: '/login', element: <LoginPage /> },
      { path: '/register', element: <RegisterPage /> },
      {
        path: '/me',
        element: (
          <UserOnly fallback={<MePageSkeleton />}>
            <MePage />
          </UserOnly>
        ),
      },
      { path: '/posts/new', element: <NewPostPage /> },
      { path: '/posts/:postId', element: <PostPage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
])
