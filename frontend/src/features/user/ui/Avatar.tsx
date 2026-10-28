import { useState } from 'react'
import { getAvatarUrl } from '../libs/avatar'

interface AvatarProps {
  userId: string | number
  username: string
  version?: string | null
  className?: string
  fallbackClassName?: string
}

export default function Avatar({
  userId,
  username,
  version,
  className = '',
  fallbackClassName = 'text-5xl',
}: AvatarProps) {
  const src = getAvatarUrl(userId, version ?? null)

  const [failedSrc, setFailedSrc] = useState<string | null>(null)
  const failed = src !== null && failedSrc === src
  const showFallback = !src || failed

  return (
    <div className={`relative overflow-hidden bg-zinc-800 ${className}`}>
      {showFallback && (
        <span
          aria-hidden="true"
          className={`absolute inset-0 grid place-items-center font-semibold text-zinc-400 ${fallbackClassName}`}
        >
          {username.charAt(0).toUpperCase()}
        </span>
      )}

      {src && !failed && (
        <img
          src={src}
          alt=""
          className="relative size-full object-cover"
          onError={() => setFailedSrc(src)}
        />
      )}
    </div>
  )
}