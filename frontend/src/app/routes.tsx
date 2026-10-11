import { createBrowserRouter} from 'react-router-dom'
import Layout from '@/app/Layout'
import LandingPage from '@/pages/LandingPage'
import LoginPage from '@/pages/LoginPage'
import PostPage from '@/pages/PostPage'
// import PostsPage from '@/pages/PostsPage'
import RegisterPage from '@/pages/RegisterPage'
import NewPostPage from '@/pages/NewPostPage'
import NotFoundPage from '@/pages/NotFound'



export const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      { path: '/', element: <LandingPage /> },
      { path: '/login', element: <LoginPage /> },
      { path: '/register', element: <RegisterPage /> },
      // { path: '/posts', element: <PostsPage /> },
      { path: '/posts/new', element: <NewPostPage /> },
      { path: '/posts/:postId', element: <PostPage /> },
      { path: '*', element: <NotFoundPage />}
    ], 
  },
])
