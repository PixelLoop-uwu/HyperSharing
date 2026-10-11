import { Link } from 'react-router-dom'
import { ArrowLeft, House } from 'lucide-react'

export default function NotFoundPage() {
  return (
    <section className="flex flex-1 flex-col items-center justify-center py-24 text-center">
      <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-zinc-800 bg-zinc-900/50 px-3 py-1.5">
        <span className="size-1.5 rounded-full bg-zinc-500" />
        <span className="text-xs font-medium text-zinc-400">
          404 · Страница не найдена
        </span>
      </div>

      <h1 className="text-8xl font-semibold tracking-[-0.08em] text-zinc-100 sm:text-9xl">
        404
      </h1>

      <h2 className="mt-4 text-xl font-medium tracking-tight text-zinc-100 sm:text-2xl">
        Кажется, здесь ничего нет
      </h2>

      <p className="mt-3 max-w-sm text-sm leading-relaxed text-zinc-400">
        Возможно, ссылка устарела или страница была удалена.
        Проверьте адрес или вернитесь на главную.
      </p>

      <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
        <Link
          to="/"
          className="inline-flex h-10 items-center gap-2 rounded-xl bg-zinc-100 px-4 text-sm font-medium text-zinc-950 transition-colors hover:bg-zinc-300"
        >
          <House className="size-4" />
          На главную
        </Link>

        <button
          type="button"
          onClick={() => window.history.back()}
          className="inline-flex h-10 items-center gap-2 rounded-xl border border-zinc-800 bg-zinc-900/50 px-4 text-sm font-medium text-zinc-300 transition-colors hover:border-zinc-700 hover:bg-zinc-800/70 hover:text-zinc-100"
        >
          <ArrowLeft className="size-4" />
          Назад
        </button>
      </div>
    </section>
  )
}