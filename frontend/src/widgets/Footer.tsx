import { FiGithub } from 'react-icons/fi'
import Logo from '@/shared/ui/content/Logo'

export default function Footer() {
  return (
    <footer className="mt-auto border-t border-zinc-800/60 bg-zinc-950/20 backdrop-blur-sm">
      <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 px-6 py-8 text-sm text-zinc-400 sm:flex-row">
        <Logo compact />
        <p className="text-center">Архивы, видео, документы и другие файлы.</p>
        <a
          href="https://github.com/PixelLoop-uwu/hyperSharing"
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1.5 rounded-md transition-colors hover:text-zinc-100 focus-visible:text-zinc-100 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-zinc-400"
        >
          <FiGithub aria-hidden="true" className="size-4" />
          GitHub
        </a>
      </div>
    </footer>
  )
}