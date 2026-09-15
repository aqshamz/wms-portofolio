"use client";

import { type ReactNode, useMemo, useState } from "react";
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
  CircleAlert,
  FileCog,
  GitBranch,
  LoaderCircle,
  Pencil,
  Plus,
  Search,
  X,
} from "lucide-react";
import Link from "next/link";
import { parseAsString, parseAsStringLiteral, useQueryStates } from "nuqs";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  deactivateDocumentStatus,
  deactivateDocumentTransition,
  deactivateDocumentType,
  documentWorkflowKeys,
  listDocumentStatuses,
  listDocumentTransitions,
  listDocumentTypes,
  listWorkflowPermissions,
} from "@/features/document-workflows/document-workflow-api";
import {
  DocumentWorkflowDialog,
  type WorkflowEditor,
} from "@/features/document-workflows/document-workflow-dialog";
import type {
  DocumentStatus,
  DocumentTransition,
  DocumentType,
} from "@/features/document-workflows/document-workflow-types";
import { cn } from "@/lib/utils";

const activeValues = ["all", "active", "inactive"] as const;
const views = ["statuses", "transitions"] as const;
const activeOptions = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;
const allRecords = {
  search: "",
  active: "all" as const,
  page: 1,
  pageSize: 100,
};

type DeactivateTarget =
  | { kind: "type"; item: DocumentType }
  | { kind: "status"; typeId: string; item: DocumentStatus }
  | { kind: "transition"; typeId: string; item: DocumentTransition };

function ErrorState({
  title,
  error,
  retry,
}: {
  title: string;
  error: Error;
  retry: () => void;
}) {
  return (
    <div
      role="alert"
      className="m-4 rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
    >
      <div className="flex gap-3">
        <CircleAlert className="mt-0.5 size-5 shrink-0" />
        <div>
          <p className="font-semibold">{title}</p>
          <p className="mt-1 text-sm">{error.message}</p>
        </div>
      </div>
      <Button variant="secondary" className="mt-4" onClick={retry}>
        Try again
      </Button>
    </div>
  );
}

function LoadingState() {
  return (
    <div className="space-y-3 p-4">
      {Array.from({ length: 4 }, (_, index) => (
        <div
          key={index}
          className="h-28 animate-pulse rounded-xl bg-slate-100"
        />
      ))}
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
            {target?.kind === "type"
              ? "Its statuses and transition rules remain recorded, but new workflow activity should no longer use this document type."
              : target?.kind === "status"
                ? "Existing documents retain this status. Deactivation also removes its initial-state designation."
                : "Documents will no longer be allowed to use this route through the workflow."}
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

export function DocumentWorkflowScreen({ canWrite }: { canWrite: boolean }) {
  const [filters, setFilters] = useQueryStates({
    documentType: parseAsString.withDefault(""),
    view: parseAsStringLiteral(views).withDefault("statuses"),
    typeSearch: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    active: parseAsStringLiteral(activeValues).withDefault("all"),
  });
  const [editor, setEditor] = useState<WorkflowEditor>();
  const [deactivateTarget, setDeactivateTarget] = useState<DeactivateTarget>();
  const queryClient = useQueryClient();
  const typeFilters = {
    search: filters.typeSearch,
    active: "all" as const,
    page: 1,
    pageSize: 100,
  };
  const childFilters = {
    search: filters.search,
    active: filters.active,
    page: 1,
    pageSize: 100,
  };
  const types = useQuery({
    queryKey: documentWorkflowKeys.typeList(typeFilters),
    queryFn: () => listDocumentTypes(typeFilters),
    placeholderData: keepPreviousData,
  });
  const selectedType = useMemo(() => {
    const items = types.data?.items ?? [];
    return (
      items.find((item) => item.document_type_id === filters.documentType) ??
      items[0]
    );
  }, [filters.documentType, types.data?.items]);
  const statuses = useQuery({
    queryKey: documentWorkflowKeys.statuses(
      selectedType?.document_type_id ?? "none",
      childFilters,
    ),
    queryFn: () =>
      listDocumentStatuses(selectedType!.document_type_id, childFilters),
    enabled: Boolean(selectedType) && filters.view === "statuses",
    placeholderData: keepPreviousData,
  });
  const allStatuses = useQuery({
    queryKey: documentWorkflowKeys.statuses(
      selectedType?.document_type_id ?? "none",
      allRecords,
    ),
    queryFn: () =>
      listDocumentStatuses(selectedType!.document_type_id, allRecords),
    enabled: Boolean(selectedType),
  });
  const transitions = useQuery({
    queryKey: documentWorkflowKeys.transitions(
      selectedType?.document_type_id ?? "none",
      childFilters,
    ),
    queryFn: () =>
      listDocumentTransitions(selectedType!.document_type_id, childFilters),
    enabled: Boolean(selectedType) && filters.view === "transitions",
    placeholderData: keepPreviousData,
  });
  const permissions = useQuery({
    queryKey: documentWorkflowKeys.permissions(),
    queryFn: listWorkflowPermissions,
  });
  const deactivate = useMutation<
    DocumentType | DocumentStatus | DocumentTransition,
    Error,
    DeactivateTarget
  >({
    mutationFn: (target: DeactivateTarget) => {
      if (target.kind === "type")
        return deactivateDocumentType(target.item.document_type_id);
      if (target.kind === "status")
        return deactivateDocumentStatus(target.typeId, target.item.status_id);
      return deactivateDocumentTransition(
        target.typeId,
        target.item.transition_id,
      );
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: documentWorkflowKeys.all,
      });
      toast.success("Workflow configuration deactivated.");
      setDeactivateTarget(undefined);
    },
  });
  const currentQuery = filters.view === "statuses" ? statuses : transitions;
  const statusById = new Map(
    (allStatuses.data?.items ?? []).map((status) => [status.status_id, status]),
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

  function startCreateChild() {
    if (!selectedType) return;
    setEditor(
      filters.view === "statuses"
        ? { kind: "status", typeId: selectedType.document_type_id }
        : {
            kind: "transition",
            typeId: selectedType.document_type_id,
            statuses: allStatuses.data?.items ?? [],
          },
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
              <GitBranch className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Document workflows
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Design document lifecycles and control who can move work
                forward.
              </p>
            </div>
          </div>
        </div>
        {canWrite ? (
          <Button
            className="w-full sm:w-auto"
            onClick={() => setEditor({ kind: "type" })}
          >
            <Plus className="size-4" />
            Create document type
          </Button>
        ) : null}
      </header>

      <div className="grid gap-5 lg:grid-cols-[19rem_minmax(0,1fr)]">
        <Panel className="h-fit overflow-hidden">
          <div className="border-b border-slate-200 p-4">
            <h2 className="font-bold text-slate-950">Document types</h2>
            <p className="mt-1 text-xs text-slate-500">
              Choose a document family to configure.
            </p>
          </div>
          <form
            className="border-b border-slate-200 p-3"
            onSubmit={(event) => {
              event.preventDefault();
              const data = new FormData(event.currentTarget);
              void setFilters({
                typeSearch: String(data.get("typeSearch") ?? "").trim(),
                documentType: "",
              });
            }}
          >
            <label className="relative block">
              <span className="sr-only">Search document types</span>
              <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-slate-400" />
              <input
                key={filters.typeSearch}
                name="typeSearch"
                defaultValue={filters.typeSearch}
                placeholder="Search document types"
                className="h-10 w-full rounded-xl border border-slate-300 pr-3 pl-9 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
              />
            </label>
          </form>
          {types.isPending ? (
            <LoadingState />
          ) : types.isError ? (
            <ErrorState
              title="Document types could not be loaded"
              error={types.error}
              retry={() => void types.refetch()}
            />
          ) : !types.data.items.length ? (
            <div className="p-8 text-center">
              <FileCog className="mx-auto size-8 text-slate-300" />
              <p className="mt-3 text-sm font-semibold text-slate-900">
                No document types found
              </p>
            </div>
          ) : (
            <div className="max-h-[32rem] space-y-1 overflow-y-auto p-2">
              {types.data.items.map((item) => {
                const active =
                  selectedType?.document_type_id === item.document_type_id;
                return (
                  <button
                    key={item.document_type_id}
                    type="button"
                    onClick={() =>
                      void setFilters({
                        documentType: item.document_type_id,
                        search: "",
                        active: "all",
                      })
                    }
                    className={cn(
                      "w-full rounded-xl p-3 text-left transition-colors",
                      active ? "bg-slate-950 text-white" : "hover:bg-slate-50",
                    )}
                  >
                    <span className="flex items-start justify-between gap-2">
                      <span className="font-semibold">{item.name}</span>
                      <span
                        className={cn(
                          "mt-1 size-2 shrink-0 rounded-full",
                          item.is_active ? "bg-emerald-400" : "bg-slate-300",
                        )}
                      />
                    </span>
                    <span
                      className={cn(
                        "mt-1 block font-mono text-xs",
                        active ? "text-cyan-200" : "text-slate-500",
                      )}
                    >
                      {item.code}
                    </span>
                    <span
                      className={cn(
                        "mt-2 block text-xs",
                        active ? "text-slate-300" : "text-slate-500",
                      )}
                    >
                      {item.module_code}
                    </span>
                  </button>
                );
              })}
            </div>
          )}
        </Panel>

        {!selectedType ? (
          <Panel className="grid min-h-80 place-items-center p-8 text-center">
            <div>
              <FileCog className="mx-auto size-10 text-slate-300" />
              <h2 className="mt-4 font-bold text-slate-950">
                Select a document type
              </h2>
              <p className="mt-1 text-sm text-slate-500">
                Its statuses and transitions will appear here.
              </p>
            </div>
          </Panel>
        ) : (
          <div className="min-w-0 space-y-4">
            <Panel className="p-4 sm:p-5">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h2 className="text-xl font-bold text-slate-950">
                      {selectedType.name}
                    </h2>
                    <StatusBadge
                      tone={selectedType.is_active ? "success" : "neutral"}
                    >
                      {selectedType.is_active ? "Active" : "Inactive"}
                    </StatusBadge>
                  </div>
                  <p className="mt-1 font-mono text-xs text-slate-500">
                    {selectedType.code} · {selectedType.module_code}
                  </p>
                  <p className="mt-3 max-w-2xl text-sm leading-6 text-slate-600">
                    {selectedType.description || "No description"}
                  </p>
                </div>
                {canWrite ? (
                  <div className="flex shrink-0 gap-1">
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() =>
                        setEditor({ kind: "type", item: selectedType })
                      }
                    >
                      <Pencil className="size-4" />
                      Edit
                    </Button>
                    {selectedType.is_active ? (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-rose-700"
                        onClick={() =>
                          setDeactivateTarget({
                            kind: "type",
                            item: selectedType,
                          })
                        }
                      >
                        Deactivate
                      </Button>
                    ) : null}
                  </div>
                ) : null}
              </div>
            </Panel>

            <div
              className="flex gap-2 overflow-x-auto pb-1"
              role="tablist"
              aria-label="Document workflow sections"
            >
              {views.map((view) => (
                <button
                  key={view}
                  type="button"
                  role="tab"
                  aria-selected={filters.view === view}
                  onClick={() =>
                    void setFilters({ view, search: "", active: "all" })
                  }
                  className={cn(
                    "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold transition-colors",
                    filters.view === view
                      ? "bg-slate-950 text-white"
                      : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
                  )}
                >
                  {view === "statuses" ? "Statuses" : "Transitions"}
                </button>
              ))}
            </div>

            <Panel className="overflow-hidden">
              <div className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-center">
                <form
                  className="flex min-w-0 flex-1 flex-col gap-3 sm:flex-row"
                  onSubmit={(event) => {
                    event.preventDefault();
                    const data = new FormData(event.currentTarget);
                    void setFilters({
                      search: String(data.get("search") ?? "").trim(),
                    });
                  }}
                >
                  <label className="relative min-w-0 flex-1">
                    <span className="sr-only">Search workflow records</span>
                    <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
                    <input
                      key={`${filters.view}-${filters.search}`}
                      name="search"
                      defaultValue={filters.search}
                      placeholder={`Search ${filters.view}`}
                      className="h-11 w-full rounded-xl border border-slate-300 pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                    />
                  </label>
                  <Select
                    ariaLabel="Filter by status"
                    value={filters.active}
                    options={activeOptions}
                    onValueChange={(active) => void setFilters({ active })}
                    className="sm:w-40"
                  />
                  <Button type="submit" variant="secondary">
                    Search
                  </Button>
                </form>
                {canWrite ? (
                  <Button
                    disabled={
                      !selectedType.is_active ||
                      (filters.view === "transitions" && activeStatusCount < 2)
                    }
                    onClick={startCreateChild}
                  >
                    <Plus className="size-4" />
                    Create{" "}
                    {filters.view === "statuses" ? "status" : "transition"}
                  </Button>
                ) : null}
              </div>
              {filters.view === "transitions" && activeStatusCount < 2 ? (
                <div className="border-b border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
                  Create at least two active statuses before adding a
                  transition.
                </div>
              ) : null}
              {currentQuery.isPending ? (
                <LoadingState />
              ) : currentQuery.isError ? (
                <ErrorState
                  title={`${filters.view === "statuses" ? "Statuses" : "Transitions"} could not be loaded`}
                  error={currentQuery.error}
                  retry={() => void currentQuery.refetch()}
                />
              ) : filters.view === "statuses" ? (
                !statuses.data?.items.length ? (
                  <EmptyState
                    icon={<FileCog className="size-9" />}
                    title="No statuses found"
                  />
                ) : (
                  <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2">
                    {statuses.data.items.map((item) => (
                      <StatusCard
                        key={item.status_id}
                        item={item}
                        canWrite={canWrite}
                        onEdit={() =>
                          setEditor({
                            kind: "status",
                            typeId: selectedType.document_type_id,
                            item,
                          })
                        }
                        onDeactivate={() =>
                          setDeactivateTarget({
                            kind: "status",
                            typeId: selectedType.document_type_id,
                            item,
                          })
                        }
                      />
                    ))}
                  </div>
                )
              ) : !transitions.data?.items.length ? (
                <EmptyState
                  icon={<GitBranch className="size-9" />}
                  title="No transitions found"
                />
              ) : (
                <div className="space-y-3 p-4 sm:p-5">
                  {transitions.data.items.map((item) => (
                    <TransitionCard
                      key={item.transition_id}
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
                          typeId: selectedType.document_type_id,
                          statuses: allStatuses.data?.items ?? [],
                          item,
                        })
                      }
                      onDeactivate={() =>
                        setDeactivateTarget({
                          kind: "transition",
                          typeId: selectedType.document_type_id,
                          item,
                        })
                      }
                    />
                  ))}
                </div>
              )}
            </Panel>
          </div>
        )}
      </div>

      {editor ? (
        <DocumentWorkflowDialog
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

function EmptyState({ icon, title }: { icon: ReactNode; title: string }) {
  return (
    <div className="px-5 py-14 text-center text-slate-300">
      {icon}
      <h3 className="mt-4 font-bold text-slate-950">{title}</h3>
      <p className="mt-1 text-sm text-slate-500">
        Adjust the filters or create the first workflow record.
      </p>
    </div>
  );
}

function StatusCard({
  item,
  canWrite,
  onEdit,
  onDeactivate,
}: {
  item: DocumentStatus;
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
    <article className="flex min-h-48 flex-col rounded-xl border border-slate-200 p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="font-bold text-slate-950">{item.name}</h3>
          <p className="mt-0.5 font-mono text-xs text-slate-500">{item.code}</p>
        </div>
        <StatusBadge tone={item.is_active ? "success" : "neutral"}>
          {item.is_active ? "Active" : "Inactive"}
        </StatusBadge>
      </div>
      <p className="mt-3 flex-1 text-sm leading-6 text-slate-600">
        {item.description || "No description"}
      </p>
      <div className="mt-3 flex flex-wrap items-center gap-2">
        <span className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-semibold text-slate-600">
          Order {item.display_order}
        </span>
        {flags.map((flag) => (
          <span
            key={flag}
            className="rounded-full bg-cyan-50 px-2.5 py-1 text-xs font-semibold text-cyan-800"
          >
            {flag}
          </span>
        ))}
      </div>
      {canWrite ? (
        <div className="mt-4 flex justify-end gap-1 border-t border-slate-100 pt-2">
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
    </article>
  );
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
  item: DocumentTransition;
  from?: DocumentStatus;
  to?: DocumentStatus;
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
  status?: DocumentStatus;
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
