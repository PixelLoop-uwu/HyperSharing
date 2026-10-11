import { useRef, type ReactElement } from 'react'

/** Render `input` once somewhere in the component, call `open()` from a button. */
export function useFilePicker(onFiles: (files: File[]) => void): {
  open: () => void
  input: ReactElement
} {
  const ref = useRef<HTMLInputElement>(null)

  const input = (
    <input
      ref={ref}
      type="file"
      multiple
      hidden
      onChange={(e) => {
        const files = Array.from(e.target.files ?? [])
        e.target.value = '' // allow picking the same file again
        if (files.length > 0) onFiles(files)
      }}
    />
  )

  return { open: () => ref.current?.click(), input }
}
