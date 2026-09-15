"use client";

import { useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  AlertTriangle,
  Beaker,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  ClipboardCheck,
  LoaderCircle,
  Pencil,
  Plus,
  Search,
  X,
  XCircle,
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
import {
  deactivateInspectionResult,
  deactivateQualityStatus,
  listInspectionResults,
  listQualityStatuses,
  qualitySetupKeys,
} from "@/features/quality-setup/quality-setup-api";
import {
  QualitySetupDialog,
  type QualityEditor,
} from "@/features/quality-setup/quality-setup-dialog";
import type {
  InspectionResult,
  QualityStatus,
} from "@/features/quality-setup/quality-setup-types";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 10;
const views = ["statuses", "results"] as const;
const activeValues = ["all", "active", "inactive"] as const;
const statusOptions = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;

type DeactivateTarget =
  | { kind: "status"; item: QualityStatus }
  | { kind: "result"; item: InspectionResult };

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
            <p className="font-semibold">Quality setup could not be loaded</p>
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
        Page <span className="font-semibold text-slate-900">{page}</span> of{" "}
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
            Deactivate {target?.item.name ?? "record"}?
          </Dialog.Title>
          <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
            Existing quality records retain this value, but it will no longer be
            available to new inspections or inventory-lot workflows. Confirm
            that operational processes no longer depend on its code.
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

export function QualitySetupScreen({ canWrite }: { canWrite: boolean }) {
  const [filters, setFilters] = useQueryStates({
    view: parseAsStringLiteral(views).withDefault("statuses"),
    search: parseAsString.withDefault(""),
    active: parseAsStringLiteral(activeValues).withDefault("all"),
    page: parseAsInteger.withDefault(1),
  });
  const [editor, setEditor] = useState<QualityEditor>();
  const [deactivateTarget, setDeactivateTarget] = useState<DeactivateTarget>();
  const queryClient = useQueryClient();
  const listFilters = {
    search: filters.search,
    active: filters.active,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const statuses = useQuery({
    queryKey: qualitySetupKeys.statusList(listFilters),
    queryFn: () => listQualityStatuses(listFilters),
    enabled: filters.view === "statuses",
    placeholderData: keepPreviousData,
  });
  const results = useQuery({
    queryKey: qualitySetupKeys.resultList(listFilters),
    queryFn: () => listInspectionResults(listFilters),
    enabled: filters.view === "results",
    placeholderData: keepPreviousData,
  });
  const currentQuery = filters.view === "statuses" ? statuses : results;
  const currentData =
    filters.view === "statuses" ? statuses.data : results.data;
  const deactivate = useMutation<
    QualityStatus | InspectionResult,
    Error,
    DeactivateTarget
  >({
    mutationFn: (target) =>
      target.kind === "status"
        ? deactivateQualityStatus(target.item.quality_status_id)
        : deactivateInspectionResult(target.item.inspection_result_id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: qualitySetupKeys.all });
      toast.success("Quality setup record deactivated.");
      setDeactivateTarget(undefined);
    },
  });
  const currentPage = currentData?.page ?? listFilters.page;
  const totalPages = currentData?.total_pages ?? 0;
  const totalItems = currentData?.total_items ?? 0;

  function startCreate() {
    setEditor(
      filters.view === "statuses" ? { kind: "status" } : { kind: "result" },
    );
  }

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <Link
            href="/master-data/catalog"
            className="inline-flex min-h-9 items-center gap-2 text-sm font-semibold text-slate-600 hover:text-slate-950"
          >
            <ChevronLeft className="size-4" />
            Catalog
          </Link>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <Beaker className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Quality setup
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Standardize quality lifecycle states and inspection outcomes.
              </p>
            </div>
          </div>
        </div>
        {canWrite ? (
          <Button className="w-full sm:w-auto" onClick={startCreate}>
            <Plus className="size-4" />
            Create{" "}
            {filters.view === "statuses"
              ? "quality status"
              : "inspection result"}
          </Button>
        ) : null}
      </header>
      <div
        className="flex gap-2 overflow-x-auto pb-1"
        role="tablist"
        aria-label="Quality setup sections"
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
            {view === "statuses" ? "Quality statuses" : "Inspection results"}
          </button>
        ))}
      </div>
      <Panel className="overflow-hidden">
        <form
          className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:p-5"
          onSubmit={(event) => {
            event.preventDefault();
            const data = new FormData(event.currentTarget);
            void setFilters({
              search: String(data.get("search") ?? "").trim(),
              page: 1,
            });
          }}
        >
          <label className="relative flex-1">
            <span className="sr-only">Search quality setup</span>
            <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
            <input
              key={`${filters.view}-${filters.search}`}
              name="search"
              defaultValue={filters.search}
              placeholder={`Search ${filters.view === "statuses" ? "quality statuses" : "inspection results"}`}
              className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
            />
          </label>
          <Select
            ariaLabel="Filter by status"
            value={filters.active}
            options={statusOptions}
            onValueChange={(active) => void setFilters({ active, page: 1 })}
            className="sm:w-40"
          />
          <Button type="submit" variant="secondary">
            Search
          </Button>
          {filters.search ? (
            <Button
              type="button"
              variant="ghost"
              onClick={() => void setFilters({ search: "", page: 1 })}
            >
              Clear
            </Button>
          ) : null}
        </form>
        {!currentQuery.isPending && !currentQuery.isError ? (
          <div className="border-b border-slate-200 px-4 py-3 text-sm text-slate-600 sm:px-5">
            <span className="font-semibold text-slate-950">{totalItems}</span>{" "}
            {filters.view === "statuses"
              ? "quality statuses"
              : "inspection results"}
          </div>
        ) : null}
        {currentQuery.isPending ? (
          <LoadingState />
        ) : currentQuery.isError ? (
          <ErrorState
            error={currentQuery.error}
            retry={() => void currentQuery.refetch()}
          />
        ) : filters.view === "statuses" ? (
          !statuses.data?.items.length ? (
            <div className="px-5 py-14 text-center">
              <ClipboardCheck className="mx-auto size-9 text-slate-300" />
              <h2 className="mt-4 font-bold text-slate-950">
                No quality statuses found
              </h2>
              <p className="mt-1 text-sm text-slate-500">
                Adjust the filters or create the first status.
              </p>
            </div>
          ) : (
            <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
              {statuses.data.items.map((item) => (
                <QualityCard
                  key={item.quality_status_id}
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
        ) : !results.data?.items.length ? (
          <div className="px-5 py-14 text-center">
            <Beaker className="mx-auto size-9 text-slate-300" />
            <h2 className="mt-4 font-bold text-slate-950">
              No inspection results found
            </h2>
            <p className="mt-1 text-sm text-slate-500">
              Adjust the filters or create the first result.
            </p>
          </div>
        ) : (
          <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
            {results.data.items.map((item) => (
              <ResultCard
                key={item.inspection_result_id}
                item={item}
                canWrite={canWrite}
                onEdit={() => setEditor({ kind: "result", item })}
                onDeactivate={() =>
                  setDeactivateTarget({ kind: "result", item })
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
        <QualitySetupDialog
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

function QualityCard({
  item,
  canWrite,
  onEdit,
  onDeactivate,
}: {
  item: QualityStatus;
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

function ResultCard({
  item,
  canWrite,
  onEdit,
  onDeactivate,
}: {
  item: InspectionResult;
  canWrite: boolean;
  onEdit: () => void;
  onDeactivate: () => void;
}) {
  return (
    <article className="flex min-h-52 flex-col rounded-xl border border-slate-200 p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="font-bold text-slate-950">{item.name}</h2>
          <p className="mt-0.5 font-mono text-xs text-slate-500">{item.code}</p>
        </div>
        <StatusBadge tone={item.is_active ? "success" : "neutral"}>
          {item.is_active ? "Active" : "Inactive"}
        </StatusBadge>
      </div>
      <p className="mt-4 text-sm leading-6 text-slate-600">
        {item.description || "No description"}
      </p>
      <div
        className={
          item.is_accepted
            ? "mt-4 flex flex-1 items-center gap-3 rounded-xl bg-emerald-50 p-3 text-emerald-800"
            : "mt-4 flex flex-1 items-center gap-3 rounded-xl bg-rose-50 p-3 text-rose-800"
        }
      >
        {item.is_accepted ? (
          <CheckCircle2 className="size-5 shrink-0" />
        ) : (
          <XCircle className="size-5 shrink-0" />
        )}
        <p className="text-sm font-semibold">
          {item.is_accepted ? "Accepted outcome" : "Rejected outcome"}
        </p>
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
