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
  createAppModule,
  getAppModule,
  modulesPermissionsKeys,
  updateAppModule,
} from "@/features/modules-permissions/modules-permissions-api";
import {
  appModuleFormSchema,
  emptyAppModuleForm,
  type AppModuleFormValues,
} from "@/features/modules-permissions/modules-permissions-schema";
import type { AppModule } from "@/features/modules-permissions/modules-permissions-types";

function formValues(item: AppModule): AppModuleFormValues {
  return {
    code: item.code,
    name: item.name,
    display_order: item.display_order,
    is_active: item.is_active,
  };
}

export function AppModuleDialog({
  item,
  onOpenChange,
}: {
  item?: AppModule;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: modulesPermissionsKeys.moduleDetail(item?.module_id ?? "new"),
    queryFn: () => getAppModule(item!.module_id),
    enabled: editing,
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AppModuleFormValues>({
    resolver: zodResolver(appModuleFormSchema),
    defaultValues: emptyAppModuleForm,
  });
  const save = useMutation({
    mutationFn: (values: AppModuleFormValues) =>
      item
        ? updateAppModule(item.module_id, {
            name: values.name.trim(),
            display_order: values.display_order,
            is_active: values.is_active,
          })
        : createAppModule({
            code: values.code.trim().toUpperCase(),
            name: values.name.trim(),
            display_order: values.display_order,
          }),
    onSuccess: async (saved) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: modulesPermissionsKeys.modules(),
        }),
        queryClient.invalidateQueries({
          queryKey: modulesPermissionsKeys.moduleDetail(saved.module_id),
        }),
      ]);
      toast.success(`Application module ${editing ? "updated" : "created"}.`);
      onOpenChange(false);
    },
  });

  useEffect(() => {
    reset(item ? formValues(detail.data ?? item) : emptyAppModuleForm);
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
                  ? "Edit application module"
                  : "Create application module"}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                Name and order modules used to group workflow permissions.
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close application module form"
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
                Loading module…
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
                  htmlFor="app-module-code"
                  required
                  error={errors.code?.message}
                >
                  <Input
                    id="app-module-code"
                    placeholder="INBOUND"
                    readOnly={editing}
                    autoFocus={!editing}
                    invalid={Boolean(errors.code)}
                    {...register("code")}
                  />
                </FormField>
                <FormField
                  label="Display order"
                  htmlFor="app-module-order"
                  required
                  error={errors.display_order?.message}
                >
                  <Input
                    id="app-module-order"
                    type="number"
                    min={0}
                    inputMode="numeric"
                    invalid={Boolean(errors.display_order)}
                    {...register("display_order")}
                  />
                </FormField>
                <div className="sm:col-span-2">
                  <FormField
                    label="Name"
                    htmlFor="app-module-name"
                    required
                    error={errors.name?.message}
                  >
                    <Input
                      id="app-module-name"
                      placeholder="Inbound"
                      autoFocus={editing}
                      invalid={Boolean(errors.name)}
                      {...register("name")}
                    />
                  </FormField>
                </div>
                {editing ? (
                  <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800 sm:col-span-2">
                    <input
                      type="checkbox"
                      className="size-4 accent-cyan-600"
                      {...register("is_active")}
                    />
                    Application module is active
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
                      : "Create module"}
                </Button>
              </div>
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
