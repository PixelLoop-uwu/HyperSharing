import GhostButton from '@/shared/ui/actions/GhostButton'
import { useConfirm } from '@/shared/ui/confirm/useConfirm'
import { useDeletePost } from '../models/useDeletePost'

export default function DeletePostButton({ postId }: { postId: string }) {
  const confirm = useConfirm()
  const { mutate: remove, isPending } = useDeletePost()

  const handleClick = async () => {
    const ok = await confirm({
      title: 'Удалить пост?',
      text: 'Это действие нельзя отменить.',
      confirmLabel: 'Удалить',
      danger: true,
    })
    if (ok) remove(postId)
  }

  return (
    <GhostButton onClick={handleClick} disabled={isPending} className="text-red-400 min-h-0 py-0">
      {isPending ? 'Удаление…' : 'Удалить'}
    </GhostButton>
  )
}