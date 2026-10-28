// shared/ui/confirm/useConfirm.ts
import { createContext, useContext } from "react";

export type ConfirmOptions = {
  title: string;
  text?: string;
  confirmLabel?: string;
  danger?: boolean;
};

export type ConfirmFn = (options: ConfirmOptions) => Promise<boolean>;

export const ConfirmContext = createContext<ConfirmFn | null>(null);

export function useConfirm(): ConfirmFn {
  const confirm = useContext(ConfirmContext);
  if (!confirm) {
    throw new Error("useConfirm должен вызываться внутри <ConfirmProvider>");
  }
  return confirm;
}