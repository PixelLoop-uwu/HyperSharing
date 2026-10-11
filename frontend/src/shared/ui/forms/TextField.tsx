import { useId, type ComponentPropsWithoutRef, type ReactNode } from 'react'

export interface TextFieldProps extends Omit<ComponentPropsWithoutRef<'input'>, 'id'> {
  label: string
  /** Quiet help text under the field, linked to it for screen readers. */
  hint?: string
  /** Content placed inside the field on the right (e.g. a show/hide button). */
  endSlot?: ReactNode
}

export default function TextField({ label, hint, endSlot, ...props }: TextFieldProps) {
  const id = useId()
  const hintId = `${id}-hint`

  return (
    <div>
      <label htmlFor={id} className="mb-1.5 block text-sm font-medium text-zinc-300">
        {label}
      </label>
      <div className="relative">
        <input
          {...props}
          id={id}
          aria-describedby={hint ? hintId : undefined}
          className={[
            'w-full rounded-xl border border-zinc-800 bg-zinc-900/20 py-2.5 text-sm text-zinc-100 transition-colors',
            endSlot ? 'pl-4 pr-11' : 'px-4',
            'placeholder:text-zinc-500 hover:border-zinc-700',
            'focus:border-zinc-400 focus:outline-none focus-visible:ring-2 focus-visible:ring-zinc-400/30',
            // Browser autofill paints a light background; keep the field dark.
            'autofill:shadow-[inset_0_0_0_1000px_#09090b] autofill:[-webkit-text-fill-color:#f4f4f5]',
          ].join(' ')}
        />
        {endSlot && <div className="absolute inset-y-0 right-1.5 flex items-center">{endSlot}</div>}
      </div>
      {hint && (
        <p id={hintId} className="mt-1.5 text-xs text-zinc-400">
          {hint}
        </p>
      )}
    </div>
  )
}
