import { FiArchive, FiFileText, FiFilm, FiLink } from 'react-icons/fi'
import Card from '@/shared/ui/layout/Card'

// Пример поста для лендинга. Замени домен на свой.
const LINK = 'hypersharing.app/p/k7x2m'

const files = [
  { icon: FiArchive, name: 'build-1.2.zip', size: '67 МБ' },
  { icon: FiFilm, name: 'баг-на-втором-уровне.mp4', size: '18 МБ' },
  { icon: FiFileText, name: 'lasted.log', size: '120 КБ' },
]

export default function PostPreview() {
  return (
    <Card
      variant="glass"
      role="img"
      aria-label="Пример поста: заголовок, текст, три вложения и ссылка"
      className="p-5 sm:p-6"
    >
      <h3 className="font-display text-lg font-semibold text-zinc-100">Сборка для тестировщиков</h3>
      <p className="mt-2 text-sm leading-relaxed text-zinc-400">
        Здесь собрал свежий билд, запись бага и список изменений.
      </p>

      <ul className="mt-5 divide-y divide-zinc-800 rounded-xl border border-zinc-800">
        {files.map(({ icon: Icon, name, size }) => (
          <li key={name} className="flex items-center gap-3 px-3.5 py-3">
            <Icon className="size-4 shrink-0 text-zinc-300" />
            <span className="min-w-0 flex-1 truncate text-sm text-zinc-100">{name}</span>
            <span className="shrink-0 text-xs tabular-nums text-zinc-400">{size}</span>
          </li>
        ))}
      </ul>

      <div className="mt-5 flex items-center gap-2.5 rounded-xl bg-zinc-950/50 px-3.5 py-2.5 text-sm">
        <FiLink className="size-4 shrink-0 text-zinc-300" />
        <span className="truncate text-zinc-100">{LINK}</span>
      </div>
    </Card>
  )
}