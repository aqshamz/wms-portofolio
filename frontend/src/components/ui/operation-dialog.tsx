"use client";

import type { ReactNode } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { X } from "lucide-react";
import { Button } from "@/components/ui/button";

export interface OperationDialogProps {
  title: string;
  description: string;
  busy?: boolean;
  closeLabel?: string;
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
}
export function OperationDialog({
  title,
  description,
  busy = false,
  closeLabel = "Close dialog",
  onOpenChange,
  children,
}: OperationDialogProps) {
  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!busy) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/50 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[94dvh] w-[calc(100%-1.5rem)] max-w-4xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl border border-slate-200 bg-white shadow-2xl">
          <header className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white p-5 sm:p-6">
            <div className="min-w-0">
              <Dialog.Title className="text-xl font-bold break-all text-slate-950">
                {title}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                {description}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <Button
                variant="ghost"
                size="icon"
                aria-label={closeLabel}
                disabled={busy}
              >
                <X className="size-5" />
              </Button>
            </Dialog.Close>
          </header>
          <div className="p-5 sm:p-6">{children}</div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
