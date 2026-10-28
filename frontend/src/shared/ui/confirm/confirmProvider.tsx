import { useCallback, useEffect, useRef, useState } from "react";
import { TriangleAlert, CircleHelp } from "lucide-react";
import { ConfirmContext, type ConfirmOptions } from "./useConfirm";

const DURATION = 300; // держи в синхроне с duration-300 ниже

export function ConfirmProvider({ children }: { children: React.ReactNode }) {
  const [opts, setOpts] = useState<ConfirmOptions | null>(null);
  const [open, setOpen] = useState(false);   // dialog в top layer
  const [shown, setShown] = useState(false); // для анимации

  const resolver = useRef<((v: boolean) => void) | undefined>(undefined);
  const timer = useRef<number | undefined>(undefined);
  const dialog = useRef<HTMLDialogElement>(null);
  const cancelBtn = useRef<HTMLButtonElement>(null);

  const confirm = useCallback((o: ConfirmOptions) => {
    resolver.current?.(false); // если вдруг уже был открыт другой
    window.clearTimeout(timer.current);
    setOpts(o);
    setOpen(true);
    return new Promise<boolean>((res) => {
      resolver.current = res;
    });
  }, []);

  useEffect(() => {
    if (!open) return;
    const d = dialog.current!;
    if (!d.open) d.showModal();
    cancelBtn.current?.focus();
    void d.offsetWidth; // принудительный reflow, чтобы переход стартовал с opacity-0
    setShown(true);
  }, [open, opts]);

  useEffect(() => () => window.clearTimeout(timer.current), []);

  const finish = (v: boolean) => {
    if (!resolver.current) return; // уже закрываемся
    resolver.current(v);
    resolver.current = undefined;
    setShown(false);
    timer.current = window.setTimeout(() => {
      dialog.current?.close();
      setOpen(false);
    }, DURATION);
  };

  return (
    <ConfirmContext.Provider value={confirm}>
      {children}
      <dialog
        ref={dialog}
        data-shown={shown || undefined}
        onCancel={(e) => {
          e.preventDefault(); // Esc идёт через анимацию
          finish(false);
        }}
        onClose={() => finish(false)} // страховка на принудительное закрытие
        onClick={(e) => e.target === dialog.current && finish(false)}
        className="m-auto w-full max-w-sm overflow-visible bg-transparent p-0 text-zinc-100
                   translate-y-3 scale-[0.97] opacity-0
                   transition duration-300 ease-[cubic-bezier(0.16,1,0.3,1)]
                   data-shown:translate-y-0 data-shown:scale-100 data-shown:opacity-100
                   motion-reduce:transition-none
                   backdrop:bg-black/40 backdrop:opacity-0 backdrop:backdrop-blur-sm
                   backdrop:transition-opacity backdrop:duration-300
                   data-shown:backdrop:opacity-100"
      >
        {opts && (
          <div className="rounded-2xl border border-zinc-800 bg-zinc-950 p-6 shadow-2xl shadow-black/20">
            <div className="flex flex-col items-center text-center">
              <div
                className={`mb-5 flex size-12 items-center justify-center rounded-2xl border ${
                  opts.danger
                    ? "border-red-500/20 bg-red-500/10 text-red-400"
                    : "border-zinc-800 bg-zinc-900/80 text-zinc-300"
                }`}
              >
                {opts.danger ? (
                  <TriangleAlert className="size-5" strokeWidth={1.7} />
                ) : (
                  <CircleHelp className="size-5" strokeWidth={1.7} />
                )}
              </div>

              <h3 className="text-xl font-medium tracking-tight text-zinc-100">
                {opts.title}
              </h3>
              {opts.text && (
                <p className="mt-2 text-sm leading-relaxed text-zinc-400">
                  {opts.text}
                </p>
              )}
            </div>

            <div className="mt-6 grid grid-cols-2 gap-3">
              <button
                ref={cancelBtn}
                onClick={() => finish(false)}
                className="rounded-xl border border-zinc-800 bg-zinc-900/50 px-4 py-2.5 text-sm font-medium text-zinc-300
                           transition-colors hover:bg-zinc-800 hover:text-zinc-100
                           focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-600"
              >
                Отмена
              </button>
              <button
                onClick={() => finish(true)}
                className={`rounded-xl px-4 py-2.5 text-sm font-medium transition-colors
                            focus-visible:outline-none focus-visible:ring-2 ${
                  opts.danger
                    ? "bg-red-500/90 text-white hover:bg-red-500 focus-visible:ring-red-400/50"
                    : "bg-zinc-100 text-zinc-900 hover:bg-white focus-visible:ring-zinc-400"
                }`}
              >
                {opts.confirmLabel ?? "Подтвердить"}
              </button>
            </div>
          </div>
        )}
      </dialog>
    </ConfirmContext.Provider>
  );
}