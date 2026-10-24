import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { deleteAvatar, uploadAvatar } from '../api'
import { useAuth } from '../store'
import { formatBytes } from '@/shared/lib/format'
import { useFilePicker } from '@/shared/lib/useFilePicker'

const AVATAR_MAX_BYTES = 5 * 1024 * 1024 // подстрой под лимит бэкенда

export function useAvatarEditor() {
  const avatarVersion = useAuth((s) => s.avatarVersion)
  const bumpAvatarVersion = useAuth((s) => s.bumpAvatarVersion)
  const [hasAvatar, setHasAvatar] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)

  const upload = useMutation({
    mutationFn: uploadAvatar,
    onSuccess: () => bumpAvatarVersion(),
  })
  const remove = useMutation({
    mutationFn: deleteAvatar,
    onSuccess: () => {
      setHasAvatar(false)
      bumpAvatarVersion()
    },
  })
  const pending = upload.isPending || remove.isPending

  const mutationError = upload.error ?? remove.error
  const error =
    validationError ??
    (mutationError instanceof Error
      ? mutationError.message
      : mutationError
        ? 'Не удалось обновить аватар'
        : null)

  function handleUpload(files: File[]) {
    const file = files[0]
    if (!file) return
    upload.reset()
    remove.reset()
    if (!file.type.startsWith('image/')) {
      setValidationError('Нужно выбрать изображение')
      return
    }
    if (file.size > AVATAR_MAX_BYTES) {
      setValidationError(`Файл слишком большой (максимум ${formatBytes(AVATAR_MAX_BYTES)})`)
      return
    }
    setValidationError(null)
    upload.mutate(file)
  }

  function handleDelete() {
    setValidationError(null)
    upload.reset()
    remove.mutate()
  }

  const picker = useFilePicker(handleUpload)

  return {
    avatarVersion,
    hasAvatar,
    pending,
    error,
    picker,
    onAvatarLoaded: () => setHasAvatar(true),
    onAvatarMissing: () => setHasAvatar(false),
    removeAvatar: handleDelete,
  }
}
