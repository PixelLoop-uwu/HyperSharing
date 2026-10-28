import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { formatBytes } from '@/shared/lib/format'
import { useFilePicker } from '@/shared/lib/useFilePicker'
import { deleteAvatar, uploadAvatar } from '../api/user'
import { useMe, useRefreshMe } from './hooks'

const AVATAR_MAX_BYTES = 5 * 1024 * 1024


export function useAvatarEditor() {
  const user = useMe().data
  const refreshMe = useRefreshMe()

  const [validationError, setValidationError] = useState<string | null>(null)

  const upload = useMutation({
    mutationFn: uploadAvatar,
    onSuccess: refreshMe,
  })

  const remove = useMutation({
    mutationFn: deleteAvatar,
    onSuccess: refreshMe,
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
    if (!file || pending) return

    upload.reset()
    remove.reset()
    setValidationError(null)

    if (!file.type.startsWith('image/')) {
      setValidationError('Нужно выбрать изображение')
      return
    }

    if (file.size > AVATAR_MAX_BYTES) {
      setValidationError(
        `Файл слишком большой (максимум ${formatBytes(AVATAR_MAX_BYTES)})`,
      )
      return
    }

    upload.mutate(file)
  }

  function handleDelete() {
    if (pending) return

    upload.reset()
    remove.reset()
    setValidationError(null)
    remove.mutate()
  }

  return {
    hasAvatar: Boolean(user?.avatar_version),
    pending,
    error,
    picker: useFilePicker(handleUpload),
    removeAvatar: handleDelete,
  }
}