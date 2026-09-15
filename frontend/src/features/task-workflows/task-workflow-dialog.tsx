"use client";

import { useEffect } from "react";
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
import { listWorkflowPermissions } from "@/features/document-workflows/document-workflow-api";
import {
  createTaskPriority,
  createTaskStatus,
  createTaskTransition,
  createTaskType,
  getTaskPriority,
  getTaskStatus,
  getTaskTransition,
  getTaskType,
  taskWorkflowKeys,
  updateTaskPriority,
  updateTaskStatus,
  updateTaskTransition,
  updateTaskType,
} from "@/features/task-workflows/task-workflow-api";
import {
  taskRecordFormSchema,
  taskTransitionFormSchema,
  type TaskRecordFormValues,
  type TaskTransitionFormValues,
} from "@/features/task-workflows/task-workflow-schema";
import type {
  TaskPriority,
  TaskStatus,
  TaskTransition,
  TaskType,
} from "@/features/task-workflows/task-workflow-types";

export type TaskWorkflowEditor =
  | { kind: "type"; item?: TaskType }
  | { kind: "status"; item?: TaskStatus }
  | { kind: "priority"; item?: TaskPriority }
  | { kind: "transition"; statuses: TaskStatus[]; item?: TaskTransition };

function optional(value: string) {
  return value.trim() || undefined;
}

function Header({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
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
          aria-label="Close task workflow form"
          className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
        >
          <X className="size-5" />
        </button>
      </Dialog.Close>
    </div>
  );
}

function LoadingDetail() {
  return (
    <div className="grid min-h-64 place-items-center p-6 text-sm text-slate-600">
      <div className="text-center">
        <LoaderCircle className="mx-auto mb-3 size-6 animate-spin text-cyan-700" />
        Loading configuration…
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

function Actions({
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

function RecordDialog({
  target,
  onOpenChange,
}: {
  target: Exclude<TaskWorkflowEditor, { kind: "transition" }>;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = target.item !== undefined;
  const queryClient = useQueryClient();
  const id =
    target.kind === "type"
      ? target.item?.task_type_id
      : target.kind === "status"
        ? target.item?.task_status_id
        : target.item?.task_priority_id;
  const detail = useQuery<TaskType | TaskStatus | TaskPriority>({
    queryKey: taskWorkflowKeys.detail(target.kind, id ?? "new"),
    queryFn: () =>
      target.kind === "type"
        ? getTaskType(id!)
        : target.kind === "status"
          ? getTaskStatus(id!)
          : getTaskPriority(id!),
    enabled: editing,
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<TaskRecordFormValues>({
    resolver: zodResolver(taskRecordFormSchema),
    defaultValues: {
      code: "",
      name: "",
      description: "",
      priority_value: 0,
      is_initial: false,
      is_final: false,
      is_cancelled: false,
      is_active: true,
    },
  });
  const save = useMutation<
    TaskType | TaskStatus | TaskPriority,
    Error,
    TaskRecordFormValues
  >({
    mutationFn: (values) => {
      if (target.kind === "type") {
        return target.item
          ? updateTaskType(target.item.task_type_id, {
              name: values.name.trim(),
              description: optional(values.description),
              is_active: values.is_active,
            })
          : createTaskType({
              code: values.code.trim().toUpperCase(),
              name: values.name.trim(),
              description: optional(values.description),
            });
      }
      if (target.kind === "status") {
        const common = {
          name: values.name.trim(),
          is_initial: values.is_initial,
          is_final: values.is_final,
          is_cancelled: values.is_cancelled,
        };
        return target.item
          ? updateTaskStatus(target.item.task_status_id, {
              ...common,
              is_active: values.is_active,
            })
          : createTaskStatus({
              code: values.code.trim().toUpperCase(),
              ...common,
            });
      }
      return target.item
        ? updateTaskPriority(target.item.task_priority_id, {
            name: values.name.trim(),
            priority_value: values.priority_value,
            is_active: values.is_active,
          })
        : createTaskPriority({
            code: values.code.trim().toUpperCase(),
            name: values.name.trim(),
            priority_value: values.priority_value,
          });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: taskWorkflowKeys.all });
      const subject =
        target.kind === "type"
          ? "Task type"
          : target.kind === "status"
            ? "Task status"
            : "Task priority";
      toast.success(`${subject} ${editing ? "updated" : "created"}.`);
      onOpenChange(false);
    },
  });

  useEffect(() => {
    const value = detail.data ?? target.item;
    reset({
      code: value?.code ?? "",
      name: value?.name ?? "",
      description:
        target.kind === "type"
          ? ((value as TaskType | undefined)?.description ?? "")
          : "",
      priority_value:
        target.kind === "priority"
          ? ((value as TaskPriority | undefined)?.priority_value ?? 0)
          : 0,
      is_initial:
        target.kind === "status"
          ? ((value as TaskStatus | undefined)?.is_initial ?? false)
          : false,
      is_final:
        target.kind === "status"
          ? ((value as TaskStatus | undefined)?.is_final ?? false)
          : false,
      is_cancelled:
        target.kind === "status"
          ? ((value as TaskStatus | undefined)?.is_cancelled ?? false)
          : false,
      is_active: value?.is_active ?? true,
    });
  }, [detail.data, reset, target]);

  const subject =
    target.kind === "type"
      ? "task type"
      : target.kind === "status"
        ? "task status"
        : "task priority";
  const description =
    target.kind === "type"
      ? "Define a category of warehouse work such as picking or putaway."
      : target.kind === "status"
        ? "Define a shared lifecycle state used by every warehouse task."
        : "Define a numeric priority level used to order task execution.";
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
          <Header
            title={`${editing ? "Edit" : "Create"} ${subject}`}
            description={description}
          />
          {editing && detail.isPending ? (
            <LoadingDetail />
          ) : editing && detail.isError ? (
            <DetailError
              error={detail.error}
              retry={() => void detail.refetch()}
            />
          ) : (
            <form
              noValidate
              className="p-5 sm:p-6"
              onSubmit={handleSubmit((values) => save.mutate(values))}
            >
              {save.error ? (
                <p
                  role="alert"
                  className="mb-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
                >
                  {save.error.message}
                </p>
              ) : null}
              <div className="grid gap-5 sm:grid-cols-2">
                <FormField
                  label="Code"
                  htmlFor="task-record-code"
                  required
                  error={errors.code?.message}
                >
                  <Input
                    id="task-record-code"
                    readOnly={editing}
                    placeholder={
                      target.kind === "type"
                        ? "PICKING"
                        : target.kind === "status"
                          ? "IN_PROGRESS"
                          : "HIGH"
                    }
                    invalid={Boolean(errors.code)}
                    {...register("code")}
                  />
                </FormField>
                <FormField
                  label="Name"
                  htmlFor="task-record-name"
                  required
                  error={errors.name?.message}
                >
                  <Input
                    id="task-record-name"
                    placeholder={
                      target.kind === "type"
                        ? "Picking"
                        : target.kind === "status"
                          ? "In progress"
                          : "High"
                    }
                    invalid={Boolean(errors.name)}
                    {...register("name")}
                  />
                </FormField>
                {target.kind === "type" ? (
                  <div className="sm:col-span-2">
                    <FormField
                      label="Description"
                      htmlFor="task-record-description"
                    >
                      <textarea
                        id="task-record-description"
                        rows={3}
                        className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                        {...register("description")}
                      />
                    </FormField>
                  </div>
                ) : null}
                {target.kind === "priority" ? (
                  <FormField
                    label="Priority value"
                    htmlFor="task-priority-value"
                    required
                    error={errors.priority_value?.message}
                  >
                    <Input
                      id="task-priority-value"
                      type="number"
                      min={0}
                      invalid={Boolean(errors.priority_value)}
                      {...register("priority_value")}
                    />
                  </FormField>
                ) : null}
                {target.kind === "status" ? (
                  <div className="sm:col-span-2">
                    <div className="grid gap-3 sm:grid-cols-3">
                      <Flag
                        label="Initial state"
                        help="Becomes the only entry state for new tasks."
                        input={register("is_initial")}
                      />
                      <Flag
                        label="Final state"
                        help="Task execution is complete."
                        input={register("is_final")}
                      />
                      <Flag
                        label="Cancelled state"
                        help="Task ended by cancellation."
                        input={register("is_cancelled")}
                      />
                    </div>
                    {errors.is_initial?.message || errors.is_final?.message ? (
                      <p className="mt-2 text-xs text-rose-700">
                        {errors.is_initial?.message ?? errors.is_final?.message}
                      </p>
                    ) : null}
                  </div>
                ) : null}
                {editing ? (
                  <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800 sm:col-span-2">
                    <input
                      type="checkbox"
                      className="size-4 accent-cyan-600"
                      {...register("is_active")}
                    />
                    Configuration is active
                  </label>
                ) : null}
              </div>
              <Actions
                editing={editing}
                pending={save.isPending}
                subject={subject}
              />
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

function Flag({
  label,
  help,
  input,
}: {
  label: string;
  help: string;
  input: UseFormRegisterReturn;
}) {
  return (
    <label className="rounded-xl border border-slate-200 p-3">
      <span className="flex items-center gap-2 text-sm font-semibold text-slate-900">
        <input type="checkbox" className="size-4 accent-cyan-600" {...input} />
        {label}
      </span>
      <span className="mt-1 block pl-6 text-xs leading-5 text-slate-500">
        {help}
      </span>
    </label>
  );
}

function TransitionDialog({
  target,
  onOpenChange,
}: {
  target: Extract<TaskWorkflowEditor, { kind: "transition" }>;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = target.item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: taskWorkflowKeys.detail(
      "transition",
      target.item?.task_status_transition_id ?? "new",
    ),
    queryFn: () => getTaskTransition(target.item!.task_status_transition_id),
    enabled: editing,
  });
  const permissions = useQuery({
    queryKey: taskWorkflowKeys.permissions(),
    queryFn: listWorkflowPermissions,
  });
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<TaskTransitionFormValues>({
    resolver: zodResolver(taskTransitionFormSchema),
    defaultValues: {
      from_status_id: "",
      to_status_id: "",
      required_permission_id: "none",
      is_active: true,
    },
  });
  const fromStatusId = useWatch({ control, name: "from_status_id" });
  const save = useMutation({
    mutationFn: (values: TaskTransitionFormValues) => {
      const required_permission_id =
        values.required_permission_id === "none"
          ? undefined
          : values.required_permission_id;
      return target.item
        ? updateTaskTransition(target.item.task_status_transition_id, {
            required_permission_id,
            is_active: values.is_active,
          })
        : createTaskTransition({
            from_status_id: values.from_status_id,
            to_status_id: values.to_status_id,
            required_permission_id,
          });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: taskWorkflowKeys.all });
      toast.success(`Task transition ${editing ? "updated" : "created"}.`);
      onOpenChange(false);
    },
  });
  useEffect(() => {
    const value = detail.data ?? target.item;
    reset({
      from_status_id: value?.from_status_id ?? "",
      to_status_id: value?.to_status_id ?? "",
      required_permission_id: value?.required_permission_id ?? "none",
      is_active: value?.is_active ?? true,
    });
  }, [detail.data, reset, target.item]);
  const statusOptions = target.statuses.map((status) => ({
    value: status.task_status_id,
    label: `${status.name} (${status.code})${status.is_active ? "" : " · Inactive"}`,
    disabled:
      !status.is_active &&
      status.task_status_id !== target.item?.from_status_id &&
      status.task_status_id !== target.item?.to_status_id,
  }));
  const fromOptions = statusOptions.map((option) => {
    const status = target.statuses.find(
      (item) => item.task_status_id === option.value,
    );
    return {
      ...option,
      disabled:
        option.disabled ||
        (Boolean(status?.is_final) &&
          option.value !== target.item?.from_status_id),
    };
  });
  const toOptions = statusOptions.map((option) => ({
    ...option,
    disabled: option.disabled || option.value === fromStatusId,
  }));
  const permissionOptions = [
    { value: "none", label: "No additional permission" },
    ...(permissions.data?.items ?? []).map((permission) => ({
      value: permission.permission_id,
      label: `${permission.name} (${permission.code})${permission.is_active ? "" : " · Inactive"}`,
      disabled:
        !permission.is_active &&
        permission.permission_id !== target.item?.required_permission_id,
    })),
  ];
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
          <Header
            title={`${editing ? "Edit" : "Create"} task transition`}
            description={
              editing
                ? "Change its permission guard or reactivate it. Transition endpoints are immutable."
                : "Allow tasks to move from one shared status to another."
            }
          />
          {editing && detail.isPending ? (
            <LoadingDetail />
          ) : editing && detail.isError ? (
            <DetailError
              error={detail.error}
              retry={() => void detail.refetch()}
            />
          ) : (
            <form
              noValidate
              className="p-5 sm:p-6"
              onSubmit={handleSubmit((values) => save.mutate(values))}
            >
              {save.error ? (
                <p
                  role="alert"
                  className="mb-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
                >
                  {save.error.message}
                </p>
              ) : null}
              <div className="grid gap-5">
                <FormField
                  label="From status"
                  htmlFor="task-transition-from"
                  required
                  error={errors.from_status_id?.message}
                >
                  <Controller
                    name="from_status_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="task-transition-from"
                        ariaLabel="From task status"
                        placeholder="Select starting status"
                        value={field.value}
                        options={fromOptions}
                        onValueChange={field.onChange}
                        disabled={editing}
                        invalid={Boolean(errors.from_status_id)}
                      />
                    )}
                  />
                </FormField>
                <FormField
                  label="To status"
                  htmlFor="task-transition-to"
                  required
                  error={errors.to_status_id?.message}
                >
                  <Controller
                    name="to_status_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="task-transition-to"
                        ariaLabel="To task status"
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
                  htmlFor="task-transition-permission"
                >
                  <Controller
                    name="required_permission_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="task-transition-permission"
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
                  <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
                    <input
                      type="checkbox"
                      className="size-4 accent-cyan-600"
                      {...register("is_active")}
                    />
                    Transition is active
                  </label>
                ) : null}
              </div>
              <Actions
                editing={editing}
                pending={save.isPending}
                subject="task transition"
              />
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function TaskWorkflowDialog({
  target,
  onOpenChange,
}: {
  target: TaskWorkflowEditor;
  onOpenChange: (open: boolean) => void;
}) {
  return target.kind === "transition" ? (
    <TransitionDialog target={target} onOpenChange={onOpenChange} />
  ) : (
    <RecordDialog target={target} onOpenChange={onOpenChange} />
  );
}
