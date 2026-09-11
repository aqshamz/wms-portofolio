"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, LoaderCircle, X } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  deactivateWarehouse,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import type { Warehouse } from "@/features/warehouses/warehouse-types";

export function WarehouseDeactivateDialog({
  warehouse,
  onOpenChange,
}: {
  warehouse?: Warehouse;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const deactivate = useMutation({
    mutationFn: () =>
      deactivateWarehouse(warehouse!.warehouse_id, warehouse!.updated_at),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: warehouseKeys.all });
      toast.success("Warehouse deactivated.");
      onOpenChange(false);
    },
  });

  return (
    <Dialog.Root
      open={warehouse !== undefined}
      onOpenChange={(open) => {
        if (!deactivate.isPending) {
          if (open) deactivate.reset();
          onOpenChange(open);
        }
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content
          role="alertdialog"
          className="fixed top-1/2 left-1/2 z-50 w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-white p-5 shadow-2xl focus:outline-none sm:p-6"
        >
          <div className="flex items-start justify-between gap-4">
            <div className="grid size-11 place-items-center rounded-xl bg-amber-100 text-amber-700">
              <AlertTriangle className="size-5" />
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close confirmation"
                className="grid size-9 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-4" />
              </button>
            </Dialog.Close>
          </div>
          <Dialog.Title className="mt-5 text-lg font-bold text-slate-950">
            Deactivate {warehouse?.name}?
          </Dialog.Title>
          <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
            The warehouse remains in historical records but cannot be selected
            for new operations. You can reactivate it later through Edit.
          </Dialog.Description>
          {deactivate.error ? (
            <div
              role="alert"
              className="mt-4 rounded-xl bg-rose-50 px-4 py-3 text-sm text-rose-900"
            >
              {deactivate.error.message}
            </div>
          ) : null}
          <div className="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <Dialog.Close asChild>
              <Button type="button" variant="secondary">
                Keep active
              </Button>
            </Dialog.Close>
            <Button
              type="button"
              disabled={deactivate.isPending}
              onClick={() => deactivate.mutate()}
              className="bg-rose-700 hover:bg-rose-800 active:bg-rose-900"
            >
              {deactivate.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              {deactivate.isPending ? "Deactivating…" : "Deactivate"}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
