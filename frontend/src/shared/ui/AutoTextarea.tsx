import { useLayoutEffect, useRef, type ComponentPropsWithoutRef } from 'react'

/** Borderless textarea that grows with its content. */
export default function AutoTextarea({
  value,
  className = '',
  ...props
}: ComponentPropsWithoutRef<'textarea'>) {
  const ref = useRef<HTMLTextAreaElement>(null)

  useLayoutEffect(() => {
    const el = ref.current
    if (!el) return
    const fit = () => {
      el.style.height = 'auto'
      el.style.height = `${el.scrollHeight}px`
    }
    fit()
    // Re-fit when the width changes (text reflows); ignore height-only changes to avoid loops.
    let width = el.clientWidth
    const observer = new ResizeObserver(() => {
      if (el.clientWidth === width) return
      width = el.clientWidth
      fit()
    })
    observer.observe(el)
    return () => observer.disconnect()
  }, [value])

  return (
    <textarea
      ref={ref}
      rows={1}
      value={value}
      className={`block w-full resize-none bg-transparent outline-none ${className}`.trim()}
      {...props}
    />
  )
}
