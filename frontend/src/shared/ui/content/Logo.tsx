import { FiHash } from 'react-icons/fi'
import { Link } from 'react-router-dom'

/** Brand mark linking home. Used by the header and the footer. */
export default function Logo({ compact = false }: { compact?: boolean }) {
  return (
    <Link
      to="/"
      className="group flex items-center gap-2.5 rounded-md focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-zinc-400"
    >
      <span className="grid size-7 place-items-center rounded-md bg-zinc-100/10 text-zinc-100 transition-colors group-hover:bg-zinc-100/20">
        <FiHash aria-hidden="true" className="size-4" />
      </span>
      <span
        className={`font-display font-semibold tracking-tight text-zinc-100 ${compact ? 'text-base' : 'text-lg'}`}
      >
        HyperSharing
      </span>
    </Link>
  )
}