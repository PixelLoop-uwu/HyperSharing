import { useMutation, useQueryClient } from '@tanstack/react-query'
import { deletePost } from '../api/posts'

export function useDeletePost() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (postId: string) => deletePost(postId),

    onSuccess: (_, postId) => {
      queryClient.removeQueries({ queryKey: ['post', postId] })
      queryClient.invalidateQueries({ queryKey: ['me', 'posts'] })
    },
  })
}