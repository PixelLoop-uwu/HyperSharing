import { useEffect, useRef, useState, type DragEvent as ReactDragEvent } from 'react'

const hasFiles = (dt: DataTransfer | null): boolean => !!dt && Array.from(dt.types).includes('Files')

/** Files from a drop, skipping folders (the browser reports them as empty "files"). */
function filesFrom(dt: DataTransfer): File[] {
  const out: File[] = []
  for (const item of Array.from(dt.items)) {
    if (item.kind !== 'file') continue
    if (item.webkitGetAsEntry?.()?.isDirectory) continue
    const file = item.getAsFile()
    if (file) out.push(file)
  }
  return out
}

/**
 * True while files from the OS are being dragged over the window.
 * Internal drags (text, elements) are ignored. Also stops the browser from
 * navigating to a file that was dropped outside any zone.
 */
export function useFileDragActive(): boolean {
  const [active, setActive] = useState(false)

  useEffect(() => {
    // dragenter/dragleave fire for every child element, so count them.
    let depth = 0
    const reset = () => {
      depth = 0
      setActive(false)
    }
    const enter = (e: DragEvent) => {
      if (!hasFiles(e.dataTransfer)) return
      depth += 1
      setActive(true)
    }
    const leave = (e: DragEvent) => {
      if (!hasFiles(e.dataTransfer)) return
      depth = Math.max(0, depth - 1)
      if (depth === 0) setActive(false)
    }
    const over = (e: DragEvent) => {
      if (hasFiles(e.dataTransfer)) e.preventDefault()
    }
    const drop = (e: DragEvent) => {
      if (hasFiles(e.dataTransfer)) e.preventDefault()
      reset()
    }

    window.addEventListener('dragenter', enter)
    window.addEventListener('dragleave', leave)
    window.addEventListener('dragover', over)
    window.addEventListener('drop', drop)
    window.addEventListener('dragend', reset)
    return () => {
      window.removeEventListener('dragenter', enter)
      window.removeEventListener('dragleave', leave)
      window.removeEventListener('dragover', over)
      window.removeEventListener('drop', drop)
      window.removeEventListener('dragend', reset)
    }
  }, [])

  return active
}

/** Spread `handlers` on any element to make it a file drop zone; `over` is true while hovered. */
export function useFileDropZone(onFiles: (files: File[]) => void) {
  const [over, setOver] = useState(false)
  const depth = useRef(0)

  const handlers = {
    onDragEnter: (e: ReactDragEvent) => {
      if (!hasFiles(e.dataTransfer)) return
      depth.current += 1
      setOver(true)
    },
    onDragLeave: (e: ReactDragEvent) => {
      if (!hasFiles(e.dataTransfer)) return
      depth.current = Math.max(0, depth.current - 1)
      if (depth.current === 0) setOver(false)
    },
    onDragOver: (e: ReactDragEvent) => {
      if (!hasFiles(e.dataTransfer)) return
      e.preventDefault()
      e.dataTransfer.dropEffect = 'copy'
    },
    onDrop: (e: ReactDragEvent) => {
      if (!hasFiles(e.dataTransfer)) return
      e.preventDefault()
      depth.current = 0
      setOver(false)
      const files = filesFrom(e.dataTransfer)
      if (files.length > 0) onFiles(files)
    },
  }

  return { over, handlers }
}
