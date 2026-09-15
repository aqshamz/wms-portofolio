"use client";

import { useEffect } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, X } from "lucide-react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  createInventoryStatus,
  getInventoryStatus,
  inventoryClassificationKeys,
  updateInventoryStatus,
} from "@/features/inventory-classifications/inventory-classification-api";
import {
  emptyInventoryStatusForm,
  inventoryStatusFormSchema,
  type InventoryStatusFormValues,
} from "@/features/inventory-classifications/inventory-classification-schema";
import type { InventoryStatus } from "@/features/inventory-classifications/inventory-classification-types";

function optional(value: string) {
  const normalized = value.trim();
  return normalized || undefined;
}

function formValues(item: InventoryStatus): InventoryStatusFormValues {
  return {
    code: item.code,
    name: item.name,
    description: item.description ?? "",
    is_allocatable: item.is_allocatable,
    is_pickable: item.is_pickable,
    is_active: item.is_active,
  };
}

export function InventoryStatusDialog({
  item,
  onOpenChange,
}: {
  item?: InventoryStatus;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: inventoryClassificationKeys.detail(
      item?.inventory_status_id ?? "new",
    ),
    queryFn: () => getInventoryStatus(item!.inventory_status_id),
    enabled: editing,
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<InventoryStatusFormValues>({
    resolver: zodResolver(inventoryStatusFormSchema),
    defaultValues: emptyInventoryStatusForm,
  });
  const save = useMutation({
    mutationFn: (values: InventoryStatusFormValues) =>
      editing
        ? updateInventoryStatus(item.inventory_status_id, {
            name: values.name.trim(),
            description: optional(values.description),
            is_allocatable: values.is_allocatable,
            is_pickable: values.is_pickable,
            is_active: values.is_active,
          })
        : createInventoryStatus({
            code: values.code.trim().toUpperCase(),
            name: values.name.trim(),
            description: optional(values.description),
            is_allocatable: values.is_allocatable,
            is_pickable: values.is_pickable,
          }),
    onSuccess: async (saved) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: inventoryClassificationKeys.lists(),
        }),
        queryClient.invalidateQueries({
          queryKey: inventoryClassificationKeys.detail(
            saved.inventory_status_id,
          ),
        }),
      ]);
      toast.success(
        editing
          ? "Inventory classification updated."
          : "Inventory classification created.",
      );
      onOpenChange(false);
    },
  });

  useEffect(() => {
    reset(item ? formValues(detail.data ?? item) : emptyInventoryStatusForm);
  }, [detail.data, item, reset]);

  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!save.isPending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] max-w-xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                {editing
                  ? "Edit inventory classification"
                  : "Create inventory classification"}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                Control whether stock in this status can be allocated and
                picked.
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close inventory classification form"
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>

          {editing && detail.isPending ? (
            <div className="grid min-h-64 place-items-center p-6 text-sm text-slate-600">
              <div className="text-center">
                <LoaderCircle className="mx-auto mb-3 size-6 animate-spin text-cyan-700" />
                Loading classification…
              </div>
            </div>
          ) : editing && detail.isError ? (
            <div className="p-5 sm:p-6">
              <p
                role="alert"
                className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
              >
                {detail.error.message}
              </p>
              <Button
                className="mt-4"
                variant="secondary"
                onClick={() => detail.refetch()}
              >
                Try again
              </Button>
            </div>
          ) : (
            <form
              noValidate
              className="p-5 sm:p-6"
              onSubmit={handleSubmit((values) => save.mutate(values))}
            >
              {save.error ? (
                <p
                  role="alert"
                  className="mb-5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
                >
                  {save.error.message}
                </p>
              ) : null}

              <div className="grid gap-5 sm:grid-cols-2">
                <FormField
                  label="Code"
                  htmlFor="inventory-status-code"
                  required
                  error={errors.code?.message}
                >
                  <Input
                    id="inventory-status-code"
                    placeholder="AVAILABLE"
                    readOnly={editing}
                    autoFocus={!editing}
                    invalid={Boolean(errors.code)}
                    {...register("code")}
                  />
                </FormField>
                <FormField
                  label="Name"
                  htmlFor="inventory-status-name"
                  required
                  error={errors.name?.message}
                >
                  <Input
                    id="inventory-status-name"
                    placeholder="Available"
                    autoFocus={editing}
                    invalid={Boolean(errors.name)}
                    {...register("name")}
                  />
                </FormField>
                <div className="sm:col-span-2">
                  <FormField
                    label="Description"
                    htmlFor="inventory-status-description"
                    error={errors.description?.message}
                  >
                    <textarea
                      id="inventory-status-description"
                      rows={3}
                      className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                      {...register("description")}
                    />
                  </FormField>
                </div>

                <label className="rounded-xl border border-slate-200 p-4">
                  <span className="flex items-center gap-3 text-sm font-semibold text-slate-900">
                    <input
                      type="checkbox"
                      className="size-4 accent-cyan-600"
                      {...register("is_allocatable")}
                    />
                    Allocatable
                  </span>
                  <span className="mt-2 block text-xs leading-5 text-slate-500">
                    Stock may be reserved for demand or outbound orders.
                  </span>
                </label>
                <label className="rounded-xl border border-slate-200 p-4">
                  <span className="flex items-center gap-3 text-sm font-semibold text-slate-900">
                    <input
                      type="checkbox"
                      className="size-4 accent-cyan-600"
                      {...register("is_pickable")}
                    />
                    Pickable
                  </span>
                  <span className="mt-2 block text-xs leading-5 text-slate-500">
                    Stock may be physically selected by picking workflows.
                  </span>
                </label>
                {editing ? (
                  <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800 sm:col-span-2">
                    <input
                      type="checkbox"
                      className="size-4 accent-cyan-600"
                      {...register("is_active")}
                    />
                    Inventory classification is active
                  </label>
                ) : null}
              </div>

              <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
                <Dialog.Close asChild>
                  <Button type="button" variant="secondary">
                    Cancel
                  </Button>
                </Dialog.Close>
                <Button type="submit" disabled={save.isPending}>
                  {save.isPending ? (
                    <LoaderCircle className="size-4 animate-spin" />
                  ) : null}
                  {save.isPending
                    ? "Saving…"
                    : editing
                      ? "Save changes"
                      : "Create classification"}
                </Button>
              </div>
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
