import { useEffect, useRef, useState, type ReactNode } from 'react'
import { FiAlertCircle, FiCheck, FiLink } from 'react-icons/fi'
import GhostButton from '../actions/GhostButton'

interface CopyButtonProps {
  /** The text to copy, or a function evaluated at click time (e.g. the current URL). */
  text: string | (() => string)
  children?: ReactNode
}

export default function CopyButton({ text, children = 'Скопировать ссылку' }: CopyButtonProps) {
  const [state, setState] = useState<'idle' | 'copied' | 'failed'>('idle')
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  useEffect(() => () => clearTimeout(timer.current), [])

  async function copy() {
    try {
      await navigator.clipboard.writeText(typeof text === 'function' ? text() : text)
      setState('copied')
    } catch {
      setState('failed') // clipboard API needs a secure context and permission
    }
    clearTimeout(timer.current)
    timer.current = setTimeout(() => setState('idle'), 2000)
  }

  return (
    <GhostButton
      onClick={copy}
      aria-live="polite"
      icon={
        state === 'copied' ? (
          <FiCheck className="size-4" />
        ) : state === 'failed' ? (
          <FiAlertCircle className="size-4" />
        ) : (
          <FiLink className="size-4" />
        )
      }
    >
      {state === 'copied' ? 'Скопировано' : state === 'failed' ? 'Не удалось скопировать' : children}
    </GhostButton>
  )
}
