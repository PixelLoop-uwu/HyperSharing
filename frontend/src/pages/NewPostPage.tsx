import { useNavigate } from 'react-router-dom'
import PostEditor from '@/features/posts/ui/PostEditor'

export default function NewPostPage() {
  const navigate = useNavigate()
  return <PostEditor onPublished={(post) => navigate(`/posts/${post.id}`)} />
}