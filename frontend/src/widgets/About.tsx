const steps = [
  {
    title: 'Создайте пост',
    text: 'Добавьте заголовок, напишите пару строк или оставьте только файлы.',
  },
  {
    title: 'Загрузите материалы',
    text: 'Архивы, видео, документы и всё остальное, что нужно передать.',
  },
  {
    title: 'Отправьте ссылку',
    text: 'Получатель откроет пост и найдёт все материалы на одной странице.',
  },
]

export default function About() {
  return (
    <section id="about" className="border-t border-zinc-800 py-24 md:py-32">
      <div className="grid gap-12 md:grid-cols-2 md:gap-16">
        <div className="md:sticky md:top-[calc(var(--header-total-h)+2rem)] md:self-start">
          <h2 className="max-w-md font-display text-3xl font-semibold tracking-tight md:text-4xl lg:text-[42px]">
            Не нужно собирать всё по отдельности:
          </h2>
          <p className="mt-6 max-w-md text-lg leading-relaxed text-zinc-400">
            Текст в одном месте, файлы в другом, а потом ещё несколько сообщений со ссылками. Можно проще: собрать материалы в пост и поделиться им
          </p>
          <p className="mt-5 max-w-md leading-relaxed text-zinc-400">
            Если один и тот же файл встречается в нескольких постах, повторно загружать его в хранилище не нужно.
          </p>
        </div>

        <ol className="relative space-y-10 border-l border-zinc-800 pl-8">
          {steps.map((step) => (
            <li key={step.title} className="relative">
              <span
                aria-hidden="true"
                className="absolute left-[calc(-2rem-5.5px)] top-2 size-2.5 rounded-full bg-zinc-100"
              />
              <h3 className="font-display text-xl font-semibold text-zinc-100">{step.title}</h3>
              <p className="mt-2 max-w-md leading-relaxed text-zinc-400">{step.text}</p>
            </li>
          ))}
        </ol>
      </div>
    </section>
  )
}