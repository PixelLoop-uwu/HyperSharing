import { useState, type FormEvent } from 'react'

/**
 * Pending/error state and the submit handler for a form.
 * `action` does the work (and navigates on success); a thrown Error becomes the message.
 */
export function useSubmit(action: () => Promise<void>, fallbackError: string) {
  const [error, setError] = useState<string | null>(null)
  const [pending, setPending] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (pending) return // ignore Enter or a second click while a request is in flight
    setError(null)
    setPending(true)
    try {
      await action()
    } catch (err) {
      setError(err instanceof Error ? err.message : fallbackError)
    } finally {
      setPending(false)
    }
  }

  return { onSubmit, error, pending }
}
