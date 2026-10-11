import { Link, type LinkProps } from 'react-router-dom'

/** Inline text link in the site style. */
export default function TextLink({ className = '', ...props }: LinkProps) {
  return (
    <Link
      className={`text-zinc-100 underline decoration-zinc-600 underline-offset-4 transition-colors hover:decoration-zinc-300 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-zinc-400 ${className}`.trim()}
      {...props}
    />
  )
}
