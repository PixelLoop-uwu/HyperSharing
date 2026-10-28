import { createBrowserRouter } from 'react-router-dom'
import Layout from '@/app/Layout'
import UserOnly from '@/app/providers/UserOnly'
import { MePageSkeleton } from '@/features/user'
import LandingPage from '@/pages/LandingPage'
import LoginPage from '@/pages/user/LoginPage'
import MePage from '@/pages/user/MePage'
import NewPostPage from '@/pages/posts/NewPostPage'
import NotFoundPage from '@/pages/NotFound'
import PostPage from '@/pages/posts/PostPage'
import RegisterPage from '@/pages/user/RegisterPage'

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
      { path: '/new', element: <NewPostPage /> },
      { path: '/:shortKey', element: <PostPage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
])
