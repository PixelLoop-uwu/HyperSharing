import { useState } from 'react'

export default function Avatar({
  userId,
  username,
  version,
  className = '',
  fallbackClassName = 'text-5xl',
  onMissing,
  onLoaded,
}: {
  userId: string | number
  username: string
  version: number
  className?: string
  fallbackClassName?: string
  onMissing?: () => void
  onLoaded?: () => void
}) {
  const [failed, setFailed] = useState(false)
  return (
    <div className={`relative overflow-hidden bg-zinc-800 ${className}`}>
      <span
        className={`absolute inset-0 grid place-items-center font-semibold text-zinc-400 ${fallbackClassName}`}
      >
        {username.charAt(0).toUpperCase()}
      </span>
      {!failed && (
        <img
          src={`/api/v1/avatars/${userId}?v=${version}`}
          alt=""
          className="relative size-full object-cover"
          onLoad={onLoaded}
          onError={() => {
            setFailed(true)
            onMissing?.()
          }}
        />
      )}
    </div>
  )
}
