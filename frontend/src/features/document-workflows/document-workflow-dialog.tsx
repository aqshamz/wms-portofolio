"use client";

import { type ReactNode, useEffect } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, X } from "lucide-react";
import {
  Controller,
  useForm,
  useWatch,
  type UseFormRegisterReturn,
} from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  createDocumentStatus,
  createDocumentTransition,
  createDocumentType,
  documentWorkflowKeys,
  getDocumentStatus,
  getDocumentTransition,
  getDocumentType,
  listWorkflowModules,
  listWorkflowPermissions,
  updateDocumentStatus,
  updateDocumentTransition,
  updateDocumentType,
} from "@/features/document-workflows/document-workflow-api";
import {
  documentStatusFormSchema,
  documentTransitionFormSchema,
  documentTypeFormSchema,
  type DocumentStatusFormValues,
  type DocumentTransitionFormValues,
  type DocumentTypeFormValues,
} from "@/features/document-workflows/document-workflow-schema";
import type {
  DocumentStatus,
  DocumentTransition,
  DocumentType,
} from "@/features/document-workflows/document-workflow-types";

export type WorkflowEditor =
  | { kind: "type"; item?: DocumentType }
  | { kind: "status"; typeId: string; item?: DocumentStatus }
  | {
      kind: "transition";
      typeId: string;
      statuses: DocumentStatus[];
      item?: DocumentTransition;
    };

function optional(value: string) {
  return value.trim() || undefined;
}

function DialogFrame({
  title,
  description,
  pending,
  children,
  onOpenChange,
}: {
  title: string;
  description: string;
  pending: boolean;
  children: ReactNode;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!pending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] max-w-xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                {title}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                {description}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close workflow form"
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

function FormActions({
  editing,
  pending,
  subject,
}: {
  editing: boolean;
  pending: boolean;
  subject: string;
}) {
  return (
    <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
      <Dialog.Close asChild>
        <Button type="button" variant="secondary">
          Cancel
        </Button>
      </Dialog.Close>
      <Button type="submit" disabled={pending}>
        {pending ? <LoaderCircle className="size-4 animate-spin" /> : null}
        {pending ? "Saving…" : editing ? "Save changes" : `Create ${subject}`}
      </Button>
    </div>
  );
}

function TypeDialog({
  item,
  onOpenChange,
}: {
  item?: DocumentType;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: documentWorkflowKeys.type(item?.document_type_id ?? "new"),
    queryFn: () => getDocumentType(item!.document_type_id),
    enabled: editing,
  });
  const modules = useQuery({
    queryKey: documentWorkflowKeys.modules(),
    queryFn: listWorkflowModules,
  });
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<DocumentTypeFormValues>({
    resolver: zodResolver(documentTypeFormSchema),
    defaultValues: {
      code: "",
      name: "",
      module_code: "",
      description: "",
      is_active: true,
    },
  });
  const save = useMutation({
    mutationFn: (values: DocumentTypeFormValues) => {
      const common = {
        name: values.name.trim(),
        module_code: values.module_code,
        description: optional(values.description),
      };
      return item
        ? updateDocumentType(item.document_type_id, {
            ...common,
            is_active: values.is_active,
          })
        : createDocumentType({
            code: values.code.trim().toUpperCase(),
            ...common,
          });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: documentWorkflowKeys.all,
      });
      toast.success(`Document type ${editing ? "updated" : "created"}.`);
      onOpenChange(false);
    },
  });

  useEffect(() => {
    const value = detail.data ?? item;
    reset({
      code: value?.code ?? "",
      name: value?.name ?? "",
      module_code: value?.module_code ?? "",
      description: value?.description ?? "",
      is_active: value?.is_active ?? true,
    });
  }, [detail.data, item, reset]);

  const moduleOptions = (modules.data?.items ?? []).map((module) => ({
    value: module.code,
    label: `${module.name} (${module.code})${module.is_active ? "" : " · Inactive"}`,
    disabled: !module.is_active && module.code !== item?.module_code,
  }));

  return (
    <DialogFrame
      title={`${editing ? "Edit" : "Create"} document type`}
      description="Define the document family and the application module that owns it."
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      {editing && detail.isPending ? (
        <LoadingDetail label="document type" />
      ) : editing && detail.isError ? (
        <DetailError error={detail.error} retry={() => void detail.refetch()} />
      ) : (
        <form
          noValidate
          className="p-5 sm:p-6"
          onSubmit={handleSubmit((values) => save.mutate(values))}
        >
          <SaveError error={save.error} />
          <div className="grid gap-5 sm:grid-cols-2">
            <FormField
              label="Code"
              htmlFor="document-type-code"
              required
              error={errors.code?.message}
            >
              <Input
                id="document-type-code"
                readOnly={editing}
                placeholder="PURCHASE_ORDER"
                invalid={Boolean(errors.code)}
                {...register("code")}
              />
            </FormField>
            <FormField
              label="Name"
              htmlFor="document-type-name"
              required
              error={errors.name?.message}
            >
              <Input
                id="document-type-name"
                placeholder="Purchase order"
                invalid={Boolean(errors.name)}
                {...register("name")}
              />
            </FormField>
            <div className="sm:col-span-2">
              <FormField
                label="Module"
                htmlFor="document-type-module"
                required
                error={errors.module_code?.message}
              >
                <Controller
                  name="module_code"
                  control={control}
                  render={({ field }) => (
                    <Select
                      id="document-type-module"
                      ariaLabel="Document module"
                      placeholder={
                        modules.isPending
                          ? "Loading modules…"
                          : "Select a module"
                      }
                      value={field.value}
                      options={moduleOptions}
                      onValueChange={field.onChange}
                      disabled={modules.isPending}
                      invalid={Boolean(errors.module_code)}
                    />
                  )}
                />
              </FormField>
            </div>
            <div className="sm:col-span-2">
              <FormField
                label="Description"
                htmlFor="document-type-description"
              >
                <textarea
                  id="document-type-description"
                  rows={3}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                  {...register("description")}
                />
              </FormField>
            </div>
            {editing ? (
              <ActiveField
                label="Document type is active"
                register={register("is_active")}
              />
            ) : null}
          </div>
          <FormActions
            editing={editing}
            pending={save.isPending}
            subject="document type"
          />
        </form>
      )}
    </DialogFrame>
  );
}

function StatusDialog({
  typeId,
  item,
  onOpenChange,
}: {
  typeId: string;
  item?: DocumentStatus;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: documentWorkflowKeys.status(typeId, item?.status_id ?? "new"),
    queryFn: () => getDocumentStatus(typeId, item!.status_id),
    enabled: editing,
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<DocumentStatusFormValues>({
    resolver: zodResolver(documentStatusFormSchema),
    defaultValues: {
      code: "",
      name: "",
      description: "",
      display_order: 0,
      is_initial: false,
      is_final: false,
      is_cancelled: false,
      is_active: true,
    },
  });
  const save = useMutation({
    mutationFn: (values: DocumentStatusFormValues) => {
      const common = {
        name: values.name.trim(),
        description: optional(values.description),
        display_order: values.display_order,
        is_initial: values.is_initial,
        is_final: values.is_final,
        is_cancelled: values.is_cancelled,
      };
      return item
        ? updateDocumentStatus(typeId, item.status_id, {
            ...common,
            is_active: values.is_active,
          })
        : createDocumentStatus(typeId, {
            code: values.code.trim().toUpperCase(),
            ...common,
          });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: documentWorkflowKeys.all,
      });
      toast.success(`Document status ${editing ? "updated" : "created"}.`);
      onOpenChange(false);
    },
  });

  useEffect(() => {
    const value = detail.data ?? item;
    reset({
      code: value?.code ?? "",
      name: value?.name ?? "",
      description: value?.description ?? "",
      display_order: value?.display_order ?? 0,
      is_initial: value?.is_initial ?? false,
      is_final: value?.is_final ?? false,
      is_cancelled: value?.is_cancelled ?? false,
      is_active: value?.is_active ?? true,
    });
  }, [detail.data, item, reset]);

  return (
    <DialogFrame
      title={`${editing ? "Edit" : "Create"} document status`}
      description="Place a named state in this document's lifecycle."
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      {editing && detail.isPending ? (
        <LoadingDetail label="status" />
      ) : editing && detail.isError ? (
        <DetailError error={detail.error} retry={() => void detail.refetch()} />
      ) : (
        <form
          noValidate
          className="p-5 sm:p-6"
          onSubmit={handleSubmit((values) => save.mutate(values))}
        >
          <SaveError error={save.error} />
          <div className="grid gap-5 sm:grid-cols-2">
            <FormField
              label="Code"
              htmlFor="document-status-code"
              required
              error={errors.code?.message}
            >
              <Input
                id="document-status-code"
                readOnly={editing}
                placeholder="RELEASED"
                invalid={Boolean(errors.code)}
                {...register("code")}
              />
            </FormField>
            <FormField
              label="Name"
              htmlFor="document-status-name"
              required
              error={errors.name?.message}
            >
              <Input
                id="document-status-name"
                placeholder="Released"
                invalid={Boolean(errors.name)}
                {...register("name")}
              />
            </FormField>
            <FormField
              label="Display order"
              htmlFor="document-status-order"
              required
              error={errors.display_order?.message}
            >
              <Input
                id="document-status-order"
                type="number"
                min={0}
                invalid={Boolean(errors.display_order)}
                {...register("display_order")}
              />
            </FormField>
            <div className="sm:col-span-2">
              <FormField
                label="Description"
                htmlFor="document-status-description"
              >
                <textarea
                  id="document-status-description"
                  rows={3}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                  {...register("description")}
                />
              </FormField>
            </div>
            <div className="grid gap-3 sm:col-span-2 sm:grid-cols-3">
              <BooleanField
                label="Initial state"
                help="Entry point for new documents."
                register={register("is_initial")}
              />
              <BooleanField
                label="Final state"
                help="Document processing is complete."
                register={register("is_final")}
              />
              <BooleanField
                label="Cancelled state"
                help="Document ended by cancellation."
                register={register("is_cancelled")}
              />
            </div>
            {editing ? (
              <ActiveField
                label="Document status is active"
                register={register("is_active")}
              />
            ) : null}
          </div>
          <FormActions
            editing={editing}
            pending={save.isPending}
            subject="status"
          />
        </form>
      )}
    </DialogFrame>
  );
}

function TransitionDialog({
  typeId,
  statuses,
  item,
  onOpenChange,
}: {
  typeId: string;
  statuses: DocumentStatus[];
  item?: DocumentTransition;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: documentWorkflowKeys.transition(
      typeId,
      item?.transition_id ?? "new",
    ),
    queryFn: () => getDocumentTransition(typeId, item!.transition_id),
    enabled: editing,
  });
  const permissions = useQuery({
    queryKey: documentWorkflowKeys.permissions(),
    queryFn: listWorkflowPermissions,
  });
  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<DocumentTransitionFormValues>({
    resolver: zodResolver(documentTransitionFormSchema),
    defaultValues: {
      from_status_id: "",
      to_status_id: "",
      required_permission_id: "none",
      is_active: true,
    },
  });
  const fromStatusId = useWatch({ control, name: "from_status_id" });
  const save = useMutation({
    mutationFn: (values: DocumentTransitionFormValues) => {
      const permission =
        values.required_permission_id === "none"
          ? undefined
          : values.required_permission_id;
      return item
        ? updateDocumentTransition(typeId, item.transition_id, {
            required_permission_id: permission,
            is_active: values.is_active,
          })
        : createDocumentTransition(typeId, {
            from_status_id: values.from_status_id,
            to_status_id: values.to_status_id,
            required_permission_id: permission,
          });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: documentWorkflowKeys.all,
      });
      toast.success(`Transition ${editing ? "updated" : "created"}.`);
      onOpenChange(false);
    },
  });

  useEffect(() => {
    const value = detail.data ?? item;
    reset({
      from_status_id: value?.from_status_id ?? "",
      to_status_id: value?.to_status_id ?? "",
      required_permission_id: value?.required_permission_id ?? "none",
      is_active: value?.is_active ?? true,
    });
  }, [detail.data, item, reset]);

  const statusOptions = statuses.map((status) => ({
    value: status.status_id,
    label: `${status.name} (${status.code})${status.is_active ? "" : " · Inactive"}`,
    disabled:
      !status.is_active &&
      status.status_id !== item?.from_status_id &&
      status.status_id !== item?.to_status_id,
  }));
  const toOptions = statusOptions.map((option) => ({
    ...option,
    disabled: option.value === fromStatusId,
  }));
  const permissionOptions = [
    { value: "none", label: "No additional permission" },
    ...(permissions.data?.items ?? []).map((permission) => ({
      value: permission.permission_id,
      label: `${permission.name} (${permission.code})${permission.is_active ? "" : " · Inactive"}`,
      disabled:
        !permission.is_active &&
        permission.permission_id !== item?.required_permission_id,
    })),
  ];

  return (
    <DialogFrame
      title={`${editing ? "Edit" : "Create"} transition`}
      description={
        editing
          ? "Change the permission guard or reactivate this transition. Endpoints are immutable."
          : "Allow a document to move from one status to another."
      }
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      {editing && detail.isPending ? (
        <LoadingDetail label="transition" />
      ) : editing && detail.isError ? (
        <DetailError error={detail.error} retry={() => void detail.refetch()} />
      ) : (
        <form
          noValidate
          className="p-5 sm:p-6"
          onSubmit={handleSubmit((values) => save.mutate(values))}
        >
          <SaveError error={save.error} />
          <div className="grid gap-5">
            <FormField
              label="From status"
              htmlFor="transition-from"
              required
              error={errors.from_status_id?.message}
            >
              <Controller
                name="from_status_id"
                control={control}
                render={({ field }) => (
                  <Select
                    id="transition-from"
                    ariaLabel="From status"
                    placeholder="Select starting status"
                    value={field.value}
                    options={statusOptions}
                    onValueChange={field.onChange}
                    disabled={editing}
                    invalid={Boolean(errors.from_status_id)}
                  />
                )}
              />
            </FormField>
            <FormField
              label="To status"
              htmlFor="transition-to"
              required
              error={errors.to_status_id?.message}
            >
              <Controller
                name="to_status_id"
                control={control}
                render={({ field }) => (
                  <Select
                    id="transition-to"
                    ariaLabel="To status"
                    placeholder="Select destination status"
                    value={field.value}
                    options={toOptions}
                    onValueChange={field.onChange}
                    disabled={editing}
                    invalid={Boolean(errors.to_status_id)}
                  />
                )}
              />
            </FormField>
            <FormField
              label="Required permission"
              htmlFor="transition-permission"
            >
              <Controller
                name="required_permission_id"
                control={control}
                render={({ field }) => (
                  <Select
                    id="transition-permission"
                    ariaLabel="Required permission"
                    value={field.value}
                    options={permissionOptions}
                    onValueChange={field.onChange}
                    disabled={permissions.isPending}
                  />
                )}
              />
              <p className="mt-1.5 text-xs text-slate-500">
                Optional. Accounts need this permission before making the
                transition.
              </p>
            </FormField>
            {editing ? (
              <ActiveField
                label="Transition is active"
                register={register("is_active")}
              />
            ) : null}
          </div>
          <FormActions
            editing={editing}
            pending={save.isPending}
            subject="transition"
          />
        </form>
      )}
    </DialogFrame>
  );
}

function LoadingDetail({ label }: { label: string }) {
  return (
    <div className="grid min-h-64 place-items-center p-6 text-sm text-slate-600">
      <div className="text-center">
        <LoaderCircle className="mx-auto mb-3 size-6 animate-spin text-cyan-700" />
        Loading {label}…
      </div>
    </div>
  );
}

function DetailError({ error, retry }: { error: Error; retry: () => void }) {
  return (
    <div className="p-5 sm:p-6">
      <p
        role="alert"
        className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
      >
        {error.message}
      </p>
      <Button className="mt-4" variant="secondary" onClick={retry}>
        Try again
      </Button>
    </div>
  );
}

function SaveError({ error }: { error: Error | null }) {
  return error ? (
    <p
      role="alert"
      className="mb-5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
    >
      {error.message}
    </p>
  ) : null;
}

function BooleanField({
  label,
  help,
  register,
}: {
  label: string;
  help: string;
  register: UseFormRegisterReturn;
}) {
  return (
    <label className="rounded-xl border border-slate-200 p-3">
      <span className="flex items-center gap-2 text-sm font-semibold text-slate-900">
        <input
          type="checkbox"
          className="size-4 accent-cyan-600"
          {...register}
        />
        {label}
      </span>
      <span className="mt-1 block pl-6 text-xs leading-5 text-slate-500">
        {help}
      </span>
    </label>
  );
}

function ActiveField({
  label,
  register,
}: {
  label: string;
  register: UseFormRegisterReturn;
}) {
  return (
    <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800 sm:col-span-2">
      <input type="checkbox" className="size-4 accent-cyan-600" {...register} />
      {label}
    </label>
  );
}

export function DocumentWorkflowDialog({
  target,
  onOpenChange,
}: {
  target: WorkflowEditor;
  onOpenChange: (open: boolean) => void;
}) {
  if (target.kind === "type")
    return <TypeDialog item={target.item} onOpenChange={onOpenChange} />;
  if (target.kind === "status")
    return (
      <StatusDialog
        typeId={target.typeId}
        item={target.item}
        onOpenChange={onOpenChange}
      />
    );
  return (
    <TransitionDialog
      typeId={target.typeId}
      statuses={target.statuses}
      item={target.item}
      onOpenChange={onOpenChange}
    />
  );
}
