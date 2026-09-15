"use client";

import { type ReactNode, useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  AlertTriangle,
  ArrowRight,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  ClipboardList,
  GitBranch,
  LoaderCircle,
  Pencil,
  Plus,
  Search,
  X,
  Zap,
} from "lucide-react";
import Link from "next/link";
import {
  parseAsInteger,
  parseAsString,
  parseAsStringLiteral,
  useQueryStates,
} from "nuqs";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import { listWorkflowPermissions } from "@/features/document-workflows/document-workflow-api";
import {
  taskWorkflowKeys,
  deactivateTaskPriority,
  deactivateTaskStatus,
  deactivateTaskTransition,
  deactivateTaskType,
  listTaskPriorities,
  listTaskStatuses,
  listTaskTransitions,
  listTaskTypes,
} from "@/features/task-workflows/task-workflow-api";
import {
  TaskWorkflowDialog,
  type TaskWorkflowEditor,
} from "@/features/task-workflows/task-workflow-dialog";
import type {
  TaskPriority,
  TaskStatus,
  TaskTransition,
  TaskType,
  TaskWorkflowView,
} from "@/features/task-workflows/task-workflow-types";
import { cn } from "@/lib/utils";

const views = ["types", "statuses", "transitions", "priorities"] as const;
const activeValues = ["all", "active", "inactive"] as const;
const PAGE_SIZE = 10;
const allStatusesFilter = {
  search: "",
  active: "all" as const,
  page: 1,
  pageSize: 100,
};
const statusOptions = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;

type DeactivateTarget =
  | { kind: "type"; item: TaskType }
  | { kind: "status"; item: TaskStatus }
  | { kind: "transition"; item: TaskTransition }
  | { kind: "priority"; item: TaskPriority };

function LoadingState() {
  return (
    <div className="space-y-3 p-5">
      {Array.from({ length: 5 }, (_, index) => (
        <div
          key={index}
          className="h-28 animate-pulse rounded-xl bg-slate-100"
        />
      ))}
    </div>
  );
}

function ErrorState({ error, retry }: { error: Error; retry: () => void }) {
  return (
    <div className="p-5">
      <div
        role="alert"
        className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
      >
        <div className="flex gap-3">
          <CircleAlert className="mt-0.5 size-5 shrink-0" />
          <div>
            <p className="font-semibold">Task workflow could not be loaded</p>
            <p className="mt-1 text-sm">{error.message}</p>
          </div>
        </div>
        <Button variant="secondary" className="mt-4" onClick={retry}>
          Try again
        </Button>
      </div>
    </div>
  );
}

function Pagination({
  page,
  totalPages,
  onPage,
}: {
  page: number;
  totalPages: number;
  onPage: (page: number) => void;
}) {
  if (totalPages <= 0) return null;
  return (
    <div className="flex flex-col gap-3 border-t border-slate-200 px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5">
      <p className="text-sm text-slate-600">
        Page <span className="font-semibold text-slate-950">{page}</span> of{" "}
        {totalPages}
      </p>
      <div className="flex gap-2">
        <Button
          variant="secondary"
          size="sm"
          disabled={page <= 1}
          onClick={() => onPage(page - 1)}
        >
          <ChevronLeft className="size-4" />
          Previous
        </Button>
        <Button
          variant="secondary"
          size="sm"
          disabled={page >= totalPages}
          onClick={() => onPage(page + 1)}
        >
          Next
          <ChevronRight className="size-4" />
        </Button>
      </div>
    </div>
  );
}

function EmptyState({ icon, title }: { icon: ReactNode; title: string }) {
  return (
    <div className="px-5 py-14 text-center text-slate-300">
      {icon}
      <h2 className="mt-4 font-bold text-slate-950">{title}</h2>
      <p className="mt-1 text-sm text-slate-500">
        Adjust the filters or create the first configuration.
      </p>
    </div>
  );
}

function DeactivateDialog({
  target,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: {
  target?: DeactivateTarget;
  pending: boolean;
  error?: Error | null;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
}) {
  const label =
    target?.kind === "transition"
      ? "this transition"
      : (target?.item.name ?? "record");
  return (
    <Dialog.Root
      open={target !== undefined}
      onOpenChange={(open) => {
        if (!pending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content
          role="alertdialog"
          className="fixed top-1/2 left-1/2 z-50 w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-white p-5 shadow-2xl focus:outline-none sm:p-6"
        >
          <div className="flex items-start justify-between">
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
            Deactivate {label}?
          </Dialog.Title>
          <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
            {target?.kind === "status"
              ? "Existing tasks retain this status. Deactivation also removes its initial-state designation."
              : target?.kind === "transition"
                ? "Tasks will no longer be allowed to move through this route."
                : "Existing task records retain this value, but it will no longer be available for new work."}
          </Dialog.Description>
          {error ? (
            <p
              role="alert"
              className="mt-4 rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
            >
              {error.message}
            </p>
          ) : null}
          <div className="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <Dialog.Close asChild>
              <Button type="button" variant="secondary">
                Keep active
              </Button>
            </Dialog.Close>
            <Button
              type="button"
              className="bg-rose-700 hover:bg-rose-800"
              disabled={pending}
              onClick={onConfirm}
            >
              {pending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              {pending ? "Deactivating…" : "Deactivate"}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function TaskWorkflowScreen({ canWrite }: { canWrite: boolean }) {
  const [filters, setFilters] = useQueryStates({
    view: parseAsStringLiteral(views).withDefault("types"),
    search: parseAsString.withDefault(""),
    active: parseAsStringLiteral(activeValues).withDefault("all"),
    page: parseAsInteger.withDefault(1),
  });
  const [editor, setEditor] = useState<TaskWorkflowEditor>();
  const [deactivateTarget, setDeactivateTarget] = useState<DeactivateTarget>();
  const queryClient = useQueryClient();
  const listFilters = {
    search: filters.view === "transitions" ? "" : filters.search,
    active: filters.active,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const types = useQuery({
    queryKey: taskWorkflowKeys.list("types", listFilters),
    queryFn: () => listTaskTypes(listFilters),
    enabled: filters.view === "types",
    placeholderData: keepPreviousData,
  });
  const statuses = useQuery({
    queryKey: taskWorkflowKeys.list("statuses", listFilters),
    queryFn: () => listTaskStatuses(listFilters),
    enabled: filters.view === "statuses",
    placeholderData: keepPreviousData,
  });
  const transitions = useQuery({
    queryKey: taskWorkflowKeys.list("transitions", listFilters),
    queryFn: () => listTaskTransitions(listFilters),
    enabled: filters.view === "transitions",
    placeholderData: keepPreviousData,
  });
  const priorities = useQuery({
    queryKey: taskWorkflowKeys.list("priorities", listFilters),
    queryFn: () => listTaskPriorities(listFilters),
    enabled: filters.view === "priorities",
    placeholderData: keepPreviousData,
  });
  const allStatuses = useQuery({
    queryKey: taskWorkflowKeys.list("statuses-all", allStatusesFilter),
    queryFn: () => listTaskStatuses(allStatusesFilter),
    enabled: filters.view === "transitions",
  });
  const permissions = useQuery({
    queryKey: taskWorkflowKeys.permissions(),
    queryFn: listWorkflowPermissions,
    enabled: filters.view === "transitions",
  });
  const deactivate = useMutation<
    TaskType | TaskStatus | TaskTransition | TaskPriority,
    Error,
    DeactivateTarget
  >({
    mutationFn: (target) =>
      target.kind === "type"
        ? deactivateTaskType(target.item.task_type_id)
        : target.kind === "status"
          ? deactivateTaskStatus(target.item.task_status_id)
          : target.kind === "transition"
            ? deactivateTaskTransition(target.item.task_status_transition_id)
            : deactivateTaskPriority(target.item.task_priority_id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: taskWorkflowKeys.all });
      toast.success("Task workflow configuration deactivated.");
      setDeactivateTarget(undefined);
    },
  });
  const currentQuery =
    filters.view === "types"
      ? types
      : filters.view === "statuses"
        ? statuses
        : filters.view === "transitions"
          ? transitions
          : priorities;
  const currentData =
    filters.view === "types"
      ? types.data
      : filters.view === "statuses"
        ? statuses.data
        : filters.view === "transitions"
          ? transitions.data
          : priorities.data;
  const currentPage = currentData?.page ?? listFilters.page;
  const totalPages = currentData?.total_pages ?? 0;
  const totalItems = currentData?.total_items ?? 0;
  const statusById = new Map(
    (allStatuses.data?.items ?? []).map((status) => [
      status.task_status_id,
      status,
    ]),
  );
  const permissionById = new Map(
    (permissions.data?.items ?? []).map((permission) => [
      permission.permission_id,
      permission,
    ]),
  );
  const activeStatusCount = (allStatuses.data?.items ?? []).filter(
    (status) => status.is_active,
  ).length;

  function startCreate() {
    setEditor(
      filters.view === "types"
        ? { kind: "type" }
        : filters.view === "statuses"
          ? { kind: "status" }
          : filters.view === "priorities"
            ? { kind: "priority" }
            : { kind: "transition", statuses: allStatuses.data?.items ?? [] },
    );
  }

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <Link
            href="/master-data/operational"
            className="inline-flex min-h-9 items-center gap-2 text-sm font-semibold text-slate-600 hover:text-slate-950"
          >
            <ChevronLeft className="size-4" />
            Operational setup
          </Link>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <ClipboardList className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Task workflows
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Configure warehouse work categories, lifecycle rules, and
                execution priorities.
              </p>
            </div>
          </div>
        </div>
        {canWrite ? (
          <Button
            className="w-full sm:w-auto"
            disabled={filters.view === "transitions" && activeStatusCount < 2}
            onClick={startCreate}
          >
            <Plus className="size-4" />
            Create {viewSubject(filters.view)}
          </Button>
        ) : null}
      </header>
      <div
        className="flex gap-2 overflow-x-auto pb-1"
        role="tablist"
        aria-label="Task workflow sections"
      >
        {views.map((view) => (
          <button
            key={view}
            type="button"
            role="tab"
            aria-selected={filters.view === view}
            onClick={() =>
              void setFilters({ view, search: "", active: "all", page: 1 })
            }
            className={cn(
              "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold transition-colors",
              filters.view === view
                ? "bg-slate-950 text-white"
                : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
            )}
          >
            {view === "types"
              ? "Task types"
              : view === "statuses"
                ? "Statuses"
                : view === "transitions"
                  ? "Transitions"
                  : "Priorities"}
          </button>
        ))}
      </div>
      <Panel className="overflow-hidden">
        <div className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:p-5">
          {filters.view !== "transitions" ? (
            <form
              className="flex min-w-0 flex-1 flex-col gap-3 sm:flex-row"
              onSubmit={(event) => {
                event.preventDefault();
                const data = new FormData(event.currentTarget);
                void setFilters({
                  search: String(data.get("search") ?? "").trim(),
                  page: 1,
                });
              }}
            >
              <label className="relative min-w-0 flex-1">
                <span className="sr-only">Search task configuration</span>
                <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
                <input
                  key={`${filters.view}-${filters.search}`}
                  name="search"
                  defaultValue={filters.search}
                  placeholder={`Search ${viewSubject(filters.view)}s`}
                  className="h-11 w-full rounded-xl border border-slate-300 pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                />
              </label>
              <Select
                ariaLabel="Filter by active status"
                value={filters.active}
                options={statusOptions}
                onValueChange={(active) => void setFilters({ active, page: 1 })}
                className="sm:w-40"
              />
              <Button type="submit" variant="secondary">
                Search
              </Button>
            </form>
          ) : (
            <div className="flex flex-1 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <p className="text-sm text-slate-500">
                Transition endpoints are shown by status name and code.
              </p>
              <Select
                ariaLabel="Filter transitions by active status"
                value={filters.active}
                options={statusOptions}
                onValueChange={(active) => void setFilters({ active, page: 1 })}
                className="sm:w-40"
              />
            </div>
          )}
        </div>
        {filters.view === "transitions" &&
        activeStatusCount < 2 &&
        !allStatuses.isPending ? (
          <div className="border-b border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
            Create at least two active task statuses before adding a transition.
          </div>
        ) : null}
        {!currentQuery.isPending && !currentQuery.isError ? (
          <div className="border-b border-slate-200 px-4 py-3 text-sm text-slate-600 sm:px-5">
            <span className="font-semibold text-slate-950">{totalItems}</span>{" "}
            {viewCountLabel(filters.view, totalItems)}
          </div>
        ) : null}
        {currentQuery.isPending ? (
          <LoadingState />
        ) : currentQuery.isError ? (
          <ErrorState
            error={currentQuery.error}
            retry={() => void currentQuery.refetch()}
          />
        ) : filters.view === "types" ? (
          !types.data?.items.length ? (
            <EmptyState
              icon={<ClipboardList className="mx-auto size-9" />}
              title="No task types found"
            />
          ) : (
            <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
              {types.data.items.map((item) => (
                <TypeCard
                  key={item.task_type_id}
                  item={item}
                  canWrite={canWrite}
                  onEdit={() => setEditor({ kind: "type", item })}
                  onDeactivate={() =>
                    setDeactivateTarget({ kind: "type", item })
                  }
                />
              ))}
            </div>
          )
        ) : filters.view === "statuses" ? (
          !statuses.data?.items.length ? (
            <EmptyState
              icon={<GitBranch className="mx-auto size-9" />}
              title="No task statuses found"
            />
          ) : (
            <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
              {statuses.data.items.map((item) => (
                <StatusCard
                  key={item.task_status_id}
                  item={item}
                  canWrite={canWrite}
                  onEdit={() => setEditor({ kind: "status", item })}
                  onDeactivate={() =>
                    setDeactivateTarget({ kind: "status", item })
                  }
                />
              ))}
            </div>
          )
        ) : filters.view === "transitions" ? (
          !transitions.data?.items.length ? (
            <EmptyState
              icon={<GitBranch className="mx-auto size-9" />}
              title="No task transitions found"
            />
          ) : (
            <div className="space-y-3 p-4 sm:p-5">
              {transitions.data.items.map((item) => (
                <TransitionCard
                  key={item.task_status_transition_id}
                  item={item}
                  from={statusById.get(item.from_status_id)}
                  to={statusById.get(item.to_status_id)}
                  permission={
                    item.required_permission_id
                      ? permissionById.get(item.required_permission_id)
                      : undefined
                  }
                  canWrite={canWrite}
                  onEdit={() =>
                    setEditor({
                      kind: "transition",
                      statuses: allStatuses.data?.items ?? [],
                      item,
                    })
                  }
                  onDeactivate={() =>
                    setDeactivateTarget({ kind: "transition", item })
                  }
                />
              ))}
            </div>
          )
        ) : !priorities.data?.items.length ? (
          <EmptyState
            icon={<Zap className="mx-auto size-9" />}
            title="No task priorities found"
          />
        ) : (
          <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
            {priorities.data.items.map((item) => (
              <PriorityCard
                key={item.task_priority_id}
                item={item}
                canWrite={canWrite}
                onEdit={() => setEditor({ kind: "priority", item })}
                onDeactivate={() =>
                  setDeactivateTarget({ kind: "priority", item })
                }
              />
            ))}
          </div>
        )}
        {!currentQuery.isPending && !currentQuery.isError ? (
          <Pagination
            page={currentPage}
            totalPages={totalPages}
            onPage={(page) => void setFilters({ page })}
          />
        ) : null}
      </Panel>
      {editor ? (
        <TaskWorkflowDialog
          target={editor}
          onOpenChange={(open) => {
            if (!open) setEditor(undefined);
          }}
        />
      ) : null}
      <DeactivateDialog
        target={deactivateTarget}
        pending={deactivate.isPending}
        error={deactivate.error}
        onConfirm={() => {
          if (deactivateTarget) deactivate.mutate(deactivateTarget);
        }}
        onOpenChange={(open) => {
          if (!open) {
            deactivate.reset();
            setDeactivateTarget(undefined);
          }
        }}
      />
    </div>
  );
}

function viewSubject(view: TaskWorkflowView) {
  return view === "types"
    ? "task type"
    : view === "statuses"
      ? "status"
      : view === "transitions"
        ? "transition"
        : "priority";
}
function viewCountLabel(view: TaskWorkflowView, count: number) {
  const label =
    view === "types"
      ? "task type"
      : view === "statuses"
        ? "status"
        : view === "transitions"
          ? "transition"
          : "priority";
  return `${label}${count === 1 ? "" : "s"}`;
}

function TypeCard({
  item,
  canWrite,
  onEdit,
  onDeactivate,
}: {
  item: TaskType;
  canWrite: boolean;
  onEdit: () => void;
  onDeactivate: () => void;
}) {
  return (
    <article className="flex min-h-48 flex-col rounded-xl border border-slate-200 p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="font-bold text-slate-950">{item.name}</h2>
          <p className="mt-0.5 font-mono text-xs text-slate-500">{item.code}</p>
        </div>
        <StatusBadge tone={item.is_active ? "success" : "neutral"}>
          {item.is_active ? "Active" : "Inactive"}
        </StatusBadge>
      </div>
      <p className="mt-4 flex-1 text-sm leading-6 text-slate-600">
        {item.description || "No description"}
      </p>
      <CardActions
        active={item.is_active}
        canWrite={canWrite}
        onEdit={onEdit}
        onDeactivate={onDeactivate}
      />
    </article>
  );
}

function StatusCard({
  item,
  canWrite,
  onEdit,
  onDeactivate,
}: {
  item: TaskStatus;
  canWrite: boolean;
  onEdit: () => void;
  onDeactivate: () => void;
}) {
  const flags = [
    item.is_initial ? "Initial" : "",
    item.is_final ? "Final" : "",
    item.is_cancelled ? "Cancelled" : "",
  ].filter(Boolean);
  return (
    <article className="flex min-h-44 flex-col rounded-xl border border-slate-200 p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="font-bold text-slate-950">{item.name}</h2>
          <p className="mt-0.5 font-mono text-xs text-slate-500">{item.code}</p>
        </div>
        <StatusBadge tone={item.is_active ? "success" : "neutral"}>
          {item.is_active ? "Active" : "Inactive"}
        </StatusBadge>
      </div>
      <div className="mt-5 flex flex-1 flex-wrap content-start gap-2">
        {flags.length ? (
          flags.map((flag) => (
            <span
              key={flag}
              className="h-fit rounded-full bg-cyan-50 px-2.5 py-1 text-xs font-semibold text-cyan-800"
            >
              {flag}
            </span>
          ))
        ) : (
          <span className="text-sm text-slate-500">
            Intermediate workflow state
          </span>
        )}
      </div>
      <CardActions
        active={item.is_active}
        canWrite={canWrite}
        onEdit={onEdit}
        onDeactivate={onDeactivate}
      />
    </article>
  );
}

function PriorityCard({
  item,
  canWrite,
  onEdit,
  onDeactivate,
}: {
  item: TaskPriority;
  canWrite: boolean;
  onEdit: () => void;
  onDeactivate: () => void;
}) {
  return (
    <article className="flex min-h-44 flex-col rounded-xl border border-slate-200 p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="font-bold text-slate-950">{item.name}</h2>
          <p className="mt-0.5 font-mono text-xs text-slate-500">{item.code}</p>
        </div>
        <StatusBadge tone={item.is_active ? "success" : "neutral"}>
          {item.is_active ? "Active" : "Inactive"}
        </StatusBadge>
      </div>
      <div className="my-5 flex flex-1 items-center gap-3 rounded-xl bg-slate-50 p-3">
        <Zap className="size-5 text-cyan-700" />
        <div>
          <p className="text-xs text-slate-500">Priority value</p>
          <p className="font-mono text-xl font-bold text-slate-950">
            {item.priority_value}
          </p>
        </div>
      </div>
      <CardActions
        active={item.is_active}
        canWrite={canWrite}
        onEdit={onEdit}
        onDeactivate={onDeactivate}
      />
    </article>
  );
}

function CardActions({
  active,
  canWrite,
  onEdit,
  onDeactivate,
}: {
  active: boolean;
  canWrite: boolean;
  onEdit: () => void;
  onDeactivate: () => void;
}) {
  return canWrite ? (
    <div className="mt-4 flex justify-end gap-1 border-t border-slate-100 pt-2">
      <Button variant="ghost" size="sm" onClick={onEdit}>
        <Pencil className="size-4" />
        Edit
      </Button>
      {active ? (
        <Button
          variant="ghost"
          size="sm"
          className="text-rose-700"
          onClick={onDeactivate}
        >
          Deactivate
        </Button>
      ) : null}
    </div>
  ) : null;
}

function TransitionCard({
  item,
  from,
  to,
  permission,
  canWrite,
  onEdit,
  onDeactivate,
}: {
  item: TaskTransition;
  from?: TaskStatus;
  to?: TaskStatus;
  permission?: { name: string; code: string };
  canWrite: boolean;
  onEdit: () => void;
  onDeactivate: () => void;
}) {
  return (
    <article className="rounded-xl border border-slate-200 p-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="min-w-0 flex-1">
          <div className="flex min-w-0 items-center gap-3">
            <StatusNode status={from} fallback={item.from_status_id} />
            <ArrowRight className="size-5 shrink-0 text-cyan-700" />
            <StatusNode status={to} fallback={item.to_status_id} />
          </div>
          <p className="mt-3 text-xs text-slate-500">
            {permission ? (
              <>
                Requires{" "}
                <span className="font-semibold text-slate-700">
                  {permission.name}
                </span>{" "}
                <span className="font-mono">({permission.code})</span>
              </>
            ) : (
              "No additional permission required"
            )}
          </p>
        </div>
        <div className="flex items-center justify-between gap-2 sm:justify-end">
          <StatusBadge tone={item.is_active ? "success" : "neutral"}>
            {item.is_active ? "Active" : "Inactive"}
          </StatusBadge>
          {canWrite ? (
            <div className="flex gap-1">
              <Button variant="ghost" size="sm" onClick={onEdit}>
                <Pencil className="size-4" />
                Edit
              </Button>
              {item.is_active ? (
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-rose-700"
                  onClick={onDeactivate}
                >
                  Deactivate
                </Button>
              ) : null}
            </div>
          ) : null}
        </div>
      </div>
    </article>
  );
}

function StatusNode({
  status,
  fallback,
}: {
  status?: TaskStatus;
  fallback: string;
}) {
  return (
    <div className="min-w-0 flex-1 rounded-xl bg-slate-50 px-3 py-2">
      <p className="truncate text-sm font-bold text-slate-950">
        {status?.name ?? "Unknown status"}
      </p>
      <p className="truncate font-mono text-xs text-slate-500">
        {status?.code ?? fallback}
      </p>
    </div>
  );
}
