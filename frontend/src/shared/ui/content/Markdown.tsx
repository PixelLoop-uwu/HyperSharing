import ReactMarkdown, { type Components } from 'react-markdown'
import remarkBreaks from 'remark-breaks'
import remarkGfm from 'remark-gfm'

const heading = 'font-display font-semibold tracking-tight text-zinc-100'

const components: Components = {
  h1: ({ node: _n, ref: _r, ...p }) => <h2 className={`pt-2 text-2xl ${heading}`} {...p} />,
  h2: ({ node: _n, ref: _r, ...p }) => <h2 className={`pt-2 text-2xl ${heading}`} {...p} />,
  h3: ({ node: _n, ref: _r, ...p }) => <h3 className={`pt-1 text-xl ${heading}`} {...p} />,
  h4: ({ node: _n, ref: _r, ...p }) => <h4 className={`text-lg ${heading}`} {...p} />,
  h5: ({ node: _n, ref: _r, ...p }) => <h4 className={`text-lg ${heading}`} {...p} />,
  h6: ({ node: _n, ref: _r, ...p }) => <h4 className={`text-lg ${heading}`} {...p} />,
  p: ({ node: _n, ref: _r, ...p }) => <p className="leading-relaxed text-zinc-200" {...p} />,
  a: ({ node: _n, ref: _r, ...p }) => (
    <a
      className="text-zinc-100 underline decoration-zinc-600 underline-offset-4 transition-colors hover:decoration-zinc-300"
      target="_blank"
      rel="noopener noreferrer nofollow ugc"
      {...p}
    />
  ),
  strong: ({ node: _n, ref: _r, ...p }) => <strong className="font-semibold text-zinc-100" {...p} />,
  em: ({ node: _n, ref: _r, ...p }) => <em className="italic" {...p} />,
  del: ({ node: _n, ref: _r, ...p }) => <del className="text-zinc-400" {...p} />,
  ul: ({ node: _n, ref: _r, className, ...p }: any) => (
    <ul
      className={`list-disc space-y-1.5 pl-5 text-zinc-200 marker:text-zinc-600 [&.contains-task-list]:list-none [&.contains-task-list]:pl-0 ${className ?? ''}`}
      {...p}
    />
  ),
  ol: ({ node: _n, ref: _r, ...p }) => (
    <ol className="list-decimal space-y-1.5 pl-5 text-zinc-200 marker:text-zinc-500" {...p} />
  ),
  li: (props: any) => {
    const { node: _n, ref: _r, checked: _c, task: _t, ...p } = props
    return <li className="leading-relaxed" {...p} />
  },
  blockquote: ({ node: _n, ref: _r, ...p }) => (
    <blockquote className="border-l-2 border-zinc-700 pl-4 text-zinc-300" {...p} />
  ),
  hr: ({ node: _n, ref: _r, ...p }) => <hr className="border-zinc-800" {...p} />,
  pre: ({ node: _n, ref: _r, ...p }) => (
    <pre
      className="overflow-x-auto rounded-xl border border-zinc-800 bg-zinc-950 p-4 text-sm leading-relaxed text-zinc-200"
      {...p}
    />
  ),
  code: ({ node: _n, ref: _r, className, ...p }: any) => {
    const isInline = !className?.includes('language-')
    return (
      <code
        className={
          isInline
            ? `rounded bg-zinc-800 px-1.5 py-0.5 font-mono text-[0.9em] text-zinc-100 ${className ?? ''}`
            : `font-mono text-[0.9em] text-zinc-100 ${className ?? ''}`
        }
        {...p}
      />
    )
  },
  table: ({ node: _n, ref: _r, ...p }) => (
    <div className="overflow-x-auto">
      <table className="w-full text-sm" {...p} />
    </div>
  ),
  th: ({ node: _n, ref: _r, ...p }) => (
    <th className="border-b border-zinc-700 px-3 py-2 text-left font-medium text-zinc-100" {...p} />
  ),
  td: ({ node: _n, ref: _r, ...p }) => (
    <td className="border-b border-zinc-800 px-3 py-2 text-zinc-300" {...p} />
  ),
  img: () => null,
}

export default function Markdown({
  children,
  className = '',
}: {
  children: string
  className?: string
}) {
  return (
    <div className={`space-y-4 wrap-break-word ${className}`.trim()}>
      <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]} components={components}>
        {children}
      </ReactMarkdown>
    </div>
  )
}