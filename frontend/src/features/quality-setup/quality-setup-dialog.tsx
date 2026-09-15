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
  createInspectionResult,
  createQualityStatus,
  getInspectionResult,
  getQualityStatus,
  qualitySetupKeys,
  updateInspectionResult,
  updateQualityStatus,
} from "@/features/quality-setup/quality-setup-api";
import {
  emptyQualitySetupForm,
  qualitySetupFormSchema,
  type QualitySetupFormValues,
} from "@/features/quality-setup/quality-setup-schema";
import type {
  InspectionResult,
  QualityStatus,
} from "@/features/quality-setup/quality-setup-types";

export type QualityEditor =
  | { kind: "status"; item?: QualityStatus }
  | { kind: "result"; item?: InspectionResult };

function optional(value: string) {
  const normalized = value.trim();
  return normalized || undefined;
}

function formValues(target: QualityEditor): QualitySetupFormValues {
  if (!target.item) return emptyQualitySetupForm;
  return {
    code: target.item.code,
    name: target.item.name,
    description: target.item.description ?? "",
    is_accepted: target.kind === "result" ? target.item.is_accepted : false,
    is_active: target.item.is_active,
  };
}

export function QualitySetupDialog({
  target,
  onOpenChange,
}: {
  target: QualityEditor;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = target.item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery<QualityStatus | InspectionResult>({
    queryKey:
      target.kind === "status"
        ? qualitySetupKeys.status(target.item?.quality_status_id ?? "new")
        : qualitySetupKeys.result(target.item?.inspection_result_id ?? "new"),
    queryFn: () =>
      target.kind === "status"
        ? getQualityStatus(target.item!.quality_status_id)
        : getInspectionResult(target.item!.inspection_result_id),
    enabled: editing,
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<QualitySetupFormValues>({
    resolver: zodResolver(qualitySetupFormSchema),
    defaultValues: emptyQualitySetupForm,
  });
  const save = useMutation<
    QualityStatus | InspectionResult,
    Error,
    QualitySetupFormValues
  >({
    mutationFn: (values: QualitySetupFormValues) => {
      const common = {
        name: values.name.trim(),
        description: optional(values.description),
      };
      if (target.kind === "status") {
        return target.item
          ? updateQualityStatus(target.item.quality_status_id, {
              ...common,
              is_active: values.is_active,
            })
          : createQualityStatus({
              code: values.code.trim().toUpperCase(),
              ...common,
            });
      }
      return target.item
        ? updateInspectionResult(target.item.inspection_result_id, {
            ...common,
            is_accepted: values.is_accepted,
            is_active: values.is_active,
          })
        : createInspectionResult({
            code: values.code.trim().toUpperCase(),
            ...common,
            is_accepted: values.is_accepted,
          });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: qualitySetupKeys.all });
      const label =
        target.kind === "status" ? "Quality status" : "Inspection result";
      toast.success(`${label} ${editing ? "updated" : "created"}.`);
      onOpenChange(false);
    },
  });

  useEffect(() => {
    if (!detail.data) {
      reset(formValues(target));
      return;
    }
    reset(
      formValues(
        target.kind === "status"
          ? { kind: "status", item: detail.data as QualityStatus }
          : { kind: "result", item: detail.data as InspectionResult },
      ),
    );
  }, [detail.data, reset, target]);

  const subject =
    target.kind === "status" ? "quality status" : "inspection result";

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
                {editing ? "Edit" : "Create"} {subject}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                {target.kind === "status"
                  ? "Describe the lifecycle state applied to quality-controlled inventory."
                  : "Define an inspection outcome and whether it is treated as accepted."}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close quality setup form"
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
                Loading {subject}…
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
                  htmlFor="quality-code"
                  required
                  error={errors.code?.message}
                >
                  <Input
                    id="quality-code"
                    placeholder={
                      target.kind === "status" ? "PENDING" : "ACCEPTED"
                    }
                    readOnly={editing}
                    autoFocus={!editing}
                    invalid={Boolean(errors.code)}
                    {...register("code")}
                  />
                </FormField>
                <FormField
                  label="Name"
                  htmlFor="quality-name"
                  required
                  error={errors.name?.message}
                >
                  <Input
                    id="quality-name"
                    placeholder={
                      target.kind === "status" ? "Pending" : "Accepted"
                    }
                    autoFocus={editing}
                    invalid={Boolean(errors.name)}
                    {...register("name")}
                  />
                </FormField>
                <div className="sm:col-span-2">
                  <FormField
                    label="Description"
                    htmlFor="quality-description"
                    error={errors.description?.message}
                  >
                    <textarea
                      id="quality-description"
                      rows={3}
                      className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                      {...register("description")}
                    />
                  </FormField>
                </div>
                {target.kind === "result" ? (
                  <label className="rounded-xl border border-slate-200 p-4 sm:col-span-2">
                    <span className="flex items-center gap-3 text-sm font-semibold text-slate-900">
                      <input
                        type="checkbox"
                        className="size-4 accent-cyan-600"
                        {...register("is_accepted")}
                      />
                      Accepted outcome
                    </span>
                    <span className="mt-2 block text-xs leading-5 text-slate-500">
                      Accepted and partially accepted outcomes may continue to
                      downstream putaway according to the operational workflow.
                    </span>
                  </label>
                ) : null}
                {editing ? (
                  <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800 sm:col-span-2">
                    <input
                      type="checkbox"
                      className="size-4 accent-cyan-600"
                      {...register("is_active")}
                    />
                    {target.kind === "status"
                      ? "Quality status"
                      : "Inspection result"}{" "}
                    is active
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
                      : `Create ${subject}`}
                </Button>
              </div>
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
