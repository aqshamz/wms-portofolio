"use client";

import { useMemo, useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
  type UseQueryResult,
} from "@tanstack/react-query";
import {
  AlertTriangle,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  FileDigit,
  Hash,
  LoaderCircle,
  Plus,
  Search,
  Sparkles,
  X,
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
import { StatusBadge } from "@/components/ui/status-badge";
import {
  deactivateDocumentNumberRule,
  documentNumberingKeys,
  listDocumentDailyCounters,
  listDocumentNumberRules,
} from "@/features/document-numbering/document-numbering-api";
import {
  AllocateDocumentIdDialog,
  NumberRuleDialog,
} from "@/features/document-numbering/document-numbering-dialogs";
import type {
  DocumentNumberRule,
  DocumentDailyCounterPage,
} from "@/features/document-numbering/document-numbering-types";
import {
  documentWorkflowKeys,
  listDocumentTypes,
} from "@/features/document-workflows/document-workflow-api";
import type { DocumentType } from "@/features/document-workflows/document-workflow-types";
import { cn } from "@/lib/utils";

const views = ["rules", "allocate", "counters"] as const;
const PAGE_SIZE = 10;
const dateFormatter = new Intl.DateTimeFormat("en-GB", {
  day: "2-digit",
  month: "short",
  year: "numeric",
});
const dateTimeFormatter = new Intl.DateTimeFormat("en-GB", {
  day: "2-digit",
  month: "short",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
});

function isoDate(offsetDays = 0) {
  const date = new Date();
  date.setUTCDate(date.getUTCDate() + offsetDays);
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Jakarta",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(date);
  const value = Object.fromEntries(
    parts.map((part) => [part.type, part.value]),
  );
  return `${value.year}-${value.month}-${value.day}`;
}

function LoadingState() {
  return (
    <div className="space-y-3 p-4">
      {Array.from({ length: 4 }, (_, index) => (
        <div
          key={index}
          className="h-24 animate-pulse rounded-xl bg-slate-100"
        />
      ))}
    </div>
  );
}

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

function RuleFormat({ rule }: { rule: DocumentNumberRule }) {
  const parts = [
    rule.prefix,
    rule.include_partner_code ? "PARTNER" : "",
    rule.include_warehouse_code ? "WAREHOUSE" : "",
    "YYYYMMDD",
    "1".padStart(rule.sequence_length, "0"),
  ].filter(Boolean);
  return (
    <span className="font-mono break-all">{parts.join(rule.separator)}</span>
  );
}

export function DocumentNumberingScreen({ canWrite }: { canWrite: boolean }) {
  const [filters, setFilters] = useQueryStates({
    documentType: parseAsString.withDefault(""),
    view: parseAsStringLiteral(views).withDefault("rules"),
    typeSearch: parseAsString.withDefault(""),
    dateFrom: parseAsString.withDefault(isoDate(-29)),
    dateTo: parseAsString.withDefault(isoDate()),
    page: parseAsInteger.withDefault(1),
  });
  const [showRuleDialog, setShowRuleDialog] = useState(false);
  const [showAllocationDialog, setShowAllocationDialog] = useState(false);
  const [deactivateRule, setDeactivateRule] = useState<DocumentNumberRule>();
  const queryClient = useQueryClient();
  const typeFilters = {
    search: filters.typeSearch,
    active: "all" as const,
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
  const rules = useQuery({
    queryKey: documentNumberingKeys.rules(
      selectedType?.document_type_id ?? "none",
    ),
    queryFn: () => listDocumentNumberRules(selectedType!.document_type_id),
    enabled: Boolean(selectedType),
  });
  const counterFilters = {
    dateFrom: filters.dateFrom,
    dateTo: filters.dateTo,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const counters = useQuery({
    queryKey: documentNumberingKeys.counters(
      selectedType?.document_type_id ?? "none",
      counterFilters,
    ),
    queryFn: () =>
      listDocumentDailyCounters(selectedType!.document_type_id, counterFilters),
    enabled: Boolean(selectedType) && filters.view === "counters",
    placeholderData: keepPreviousData,
  });
  const deactivate = useMutation({
    mutationFn: (rule: DocumentNumberRule) =>
      deactivateDocumentNumberRule(
        rule.document_type_id,
        rule.document_number_rule_id,
      ),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: documentNumberingKeys.all,
      });
      toast.success("Numbering rule deactivated.");
      setDeactivateRule(undefined);
    },
  });
  const activeRule = rules.data?.items.find((rule) => rule.is_active);
  const sortedRules = [...(rules.data?.items ?? [])].sort((a, b) =>
    b.effective_from.localeCompare(a.effective_from),
  );
  const totalPages = counters.data?.total_pages ?? 0;

  return (
    <div className="space-y-6">
      <header>
        <Link
          href="/master-data/operational"
          className="inline-flex min-h-9 items-center gap-2 text-sm font-semibold text-slate-600 hover:text-slate-950"
        >
          <ChevronLeft className="size-4" />
          Operational setup
        </Link>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <FileDigit className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Document numbering
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Manage identifier formats, controlled allocation, and daily
              sequence usage.
            </p>
          </div>
        </div>
      </header>

      <div className="grid gap-5 lg:grid-cols-[19rem_minmax(0,1fr)]">
        <Panel className="h-fit overflow-hidden">
          <div className="border-b border-slate-200 p-4">
            <h2 className="font-bold text-slate-950">Document types</h2>
            <p className="mt-1 text-xs text-slate-500">
              Choose which identifier series to manage.
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
              <FileDigit className="mx-auto size-8 text-slate-300" />
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
                        page: 1,
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
              <FileDigit className="mx-auto size-10 text-slate-300" />
              <h2 className="mt-4 font-bold text-slate-950">
                Select a document type
              </h2>
              <p className="mt-1 text-sm text-slate-500">
                Its numbering rules and counters will appear here.
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
                </div>
                {rules.isPending ? (
                  <span className="text-sm text-slate-500">Loading rule…</span>
                ) : activeRule ? (
                  <div className="rounded-xl bg-emerald-50 px-3 py-2 text-sm text-emerald-900">
                    <span className="font-semibold">Active format</span>
                    <span className="mt-1 block text-xs">
                      <RuleFormat rule={activeRule} />
                    </span>
                  </div>
                ) : (
                  <div className="rounded-xl bg-amber-50 px-3 py-2 text-sm font-semibold text-amber-900">
                    No active rule
                  </div>
                )}
              </div>
            </Panel>

            <div
              className="flex gap-2 overflow-x-auto pb-1"
              role="tablist"
              aria-label="Document numbering sections"
            >
              {views.map((view) => (
                <button
                  key={view}
                  type="button"
                  role="tab"
                  aria-selected={filters.view === view}
                  onClick={() => void setFilters({ view, page: 1 })}
                  className={cn(
                    "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold transition-colors",
                    filters.view === view
                      ? "bg-slate-950 text-white"
                      : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
                  )}
                >
                  {view === "rules"
                    ? "Number rules"
                    : view === "allocate"
                      ? "Allocate ID"
                      : "Daily counters"}
                </button>
              ))}
            </div>

            {filters.view === "rules" ? (
              <RulesPanel
                documentType={selectedType}
                rules={sortedRules}
                activeRule={activeRule}
                pending={rules.isPending}
                error={rules.error}
                canWrite={canWrite}
                onRetry={() => void rules.refetch()}
                onReplace={() => setShowRuleDialog(true)}
                onDeactivate={setDeactivateRule}
              />
            ) : filters.view === "allocate" ? (
              <AllocationPanel
                documentType={selectedType}
                activeRule={activeRule}
                pending={rules.isPending}
                canWrite={canWrite}
                onAllocate={() => setShowAllocationDialog(true)}
              />
            ) : (
              <CountersPanel
                filters={counterFilters}
                query={counters}
                totalPages={totalPages}
                onFilter={(dateFrom, dateTo) =>
                  void setFilters({ dateFrom, dateTo, page: 1 })
                }
                onPage={(page) => void setFilters({ page })}
              />
            )}
          </div>
        )}
      </div>

      {showRuleDialog && selectedType ? (
        <NumberRuleDialog
          documentType={selectedType}
          currentRule={activeRule}
          onOpenChange={(open) => {
            if (!open) setShowRuleDialog(false);
          }}
        />
      ) : null}
      {showAllocationDialog && selectedType && activeRule ? (
        <AllocateDocumentIdDialog
          documentType={selectedType}
          rule={activeRule}
          onOpenChange={(open) => {
            if (!open) setShowAllocationDialog(false);
          }}
        />
      ) : null}
      <DeactivateRuleDialog
        rule={deactivateRule}
        pending={deactivate.isPending}
        error={deactivate.error}
        onConfirm={() => {
          if (deactivateRule) deactivate.mutate(deactivateRule);
        }}
        onOpenChange={(open) => {
          if (!open) {
            deactivate.reset();
            setDeactivateRule(undefined);
          }
        }}
      />
    </div>
  );
}

function RulesPanel({
  documentType,
  rules,
  activeRule,
  pending,
  error,
  canWrite,
  onRetry,
  onReplace,
  onDeactivate,
}: {
  documentType: DocumentType;
  rules: DocumentNumberRule[];
  activeRule?: DocumentNumberRule;
  pending: boolean;
  error: Error | null;
  canWrite: boolean;
  onRetry: () => void;
  onReplace: () => void;
  onDeactivate: (rule: DocumentNumberRule) => void;
}) {
  return (
    <Panel className="overflow-hidden">
      <div className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5">
        <div>
          <h2 className="font-bold text-slate-950">Numbering rule history</h2>
          <p className="mt-1 text-sm text-slate-500">
            Rules are immutable so previously issued IDs remain explainable.
          </p>
        </div>
        {canWrite ? (
          <Button disabled={!documentType.is_active} onClick={onReplace}>
            <Plus className="size-4" />
            {activeRule ? "Replace active rule" : "Create rule"}
          </Button>
        ) : null}
      </div>
      {pending ? (
        <LoadingState />
      ) : error ? (
        <ErrorState
          title="Numbering rules could not be loaded"
          error={error}
          retry={onRetry}
        />
      ) : !rules.length ? (
        <div className="px-5 py-14 text-center">
          <Hash className="mx-auto size-9 text-slate-300" />
          <h3 className="mt-4 font-bold text-slate-950">
            No numbering rules yet
          </h3>
          <p className="mt-1 text-sm text-slate-500">
            Create the first rule before allocating document IDs.
          </p>
        </div>
      ) : (
        <div className="space-y-3 p-4 sm:p-5">
          {rules.map((rule) => (
            <article
              key={rule.document_number_rule_id}
              className="rounded-xl border border-slate-200 p-4"
            >
              <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <RuleFormat rule={rule} />
                    <StatusBadge tone={rule.is_active ? "success" : "neutral"}>
                      {rule.is_active ? "Active" : "Historical"}
                    </StatusBadge>
                  </div>
                  <div className="mt-3 flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-500">
                    <span>
                      Effective{" "}
                      {dateFormatter.format(
                        new Date(`${rule.effective_from}T00:00:00`),
                      )}
                    </span>
                    <span>
                      {rule.effective_until
                        ? `Until ${dateFormatter.format(new Date(`${rule.effective_until}T00:00:00`))}`
                        : "No end date"}
                    </span>
                    <span>{rule.sequence_length} sequence digits</span>
                  </div>
                </div>
                {canWrite && rule.is_active ? (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-rose-700"
                    onClick={() => onDeactivate(rule)}
                  >
                    Deactivate
                  </Button>
                ) : null}
              </div>
            </article>
          ))}
        </div>
      )}
    </Panel>
  );
}

function AllocationPanel({
  documentType,
  activeRule,
  pending,
  canWrite,
  onAllocate,
}: {
  documentType: DocumentType;
  activeRule?: DocumentNumberRule;
  pending: boolean;
  canWrite: boolean;
  onAllocate: () => void;
}) {
  return (
    <Panel className="overflow-hidden">
      <div className="border-b border-slate-200 p-4 sm:p-5">
        <h2 className="font-bold text-slate-950">Controlled ID allocation</h2>
        <p className="mt-1 text-sm text-slate-500">
          Use this only when a transaction workflow needs a real reserved
          identifier.
        </p>
      </div>
      {pending ? (
        <LoadingState />
      ) : !activeRule ? (
        <div className="px-5 py-14 text-center">
          <CircleAlert className="mx-auto size-9 text-amber-400" />
          <h3 className="mt-4 font-bold text-slate-950">
            No active numbering rule
          </h3>
          <p className="mt-1 text-sm text-slate-500">
            Create an active rule before allocating an ID.
          </p>
        </div>
      ) : (
        <div className="p-4 sm:p-6">
          <div className="rounded-2xl bg-slate-950 p-5 text-white">
            <p className="text-xs font-semibold tracking-wide text-cyan-300 uppercase">
              Current format
            </p>
            <p className="mt-3 text-lg font-bold">
              <RuleFormat rule={activeRule} />
            </p>
            <div className="mt-4 flex flex-wrap gap-2 text-xs text-slate-300">
              <span className="rounded-full bg-white/10 px-2.5 py-1">
                Daily sequence
              </span>
              {activeRule.include_partner_code ? (
                <span className="rounded-full bg-white/10 px-2.5 py-1">
                  Partner required
                </span>
              ) : null}
              {activeRule.include_warehouse_code ? (
                <span className="rounded-full bg-white/10 px-2.5 py-1">
                  Warehouse required
                </span>
              ) : null}
            </div>
          </div>
          <div className="mt-5 flex gap-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-amber-950">
            <AlertTriangle className="mt-0.5 size-5 shrink-0" />
            <p className="text-sm leading-6">
              Allocation advances the counter permanently. For visual checks,
              use the format shown above instead of allocating a test number.
            </p>
          </div>
          {canWrite ? (
            <Button
              className="mt-5 w-full sm:w-auto"
              disabled={!documentType.is_active}
              onClick={onAllocate}
            >
              <Sparkles className="size-4" />
              Allocate next ID
            </Button>
          ) : (
            <p className="mt-5 text-sm text-slate-500">
              Write permission is required to allocate an ID.
            </p>
          )}
        </div>
      )}
    </Panel>
  );
}

function CountersPanel({
  filters,
  query,
  totalPages,
  onFilter,
  onPage,
}: {
  filters: { dateFrom: string; dateTo: string; page: number; pageSize: number };
  query: UseQueryResult<DocumentDailyCounterPage, Error>;
  totalPages: number;
  onFilter: (from: string, to: string) => void;
  onPage: (page: number) => void;
}) {
  return (
    <Panel className="overflow-hidden">
      <form
        className="grid gap-3 border-b border-slate-200 p-4 sm:grid-cols-[1fr_1fr_auto] sm:p-5"
        onSubmit={(event) => {
          event.preventDefault();
          const data = new FormData(event.currentTarget);
          onFilter(
            String(data.get("dateFrom") ?? ""),
            String(data.get("dateTo") ?? ""),
          );
        }}
      >
        <label className="text-sm font-semibold text-slate-800">
          From
          <input
            name="dateFrom"
            type="date"
            required
            max={filters.dateTo}
            defaultValue={filters.dateFrom}
            className="mt-2 h-11 w-full rounded-xl border border-slate-300 px-3 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
          />
        </label>
        <label className="text-sm font-semibold text-slate-800">
          To
          <input
            name="dateTo"
            type="date"
            required
            min={filters.dateFrom}
            defaultValue={filters.dateTo}
            className="mt-2 h-11 w-full rounded-xl border border-slate-300 px-3 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
          />
        </label>
        <Button type="submit" variant="secondary" className="self-end">
          Apply dates
        </Button>
      </form>
      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState
          title="Daily counters could not be loaded"
          error={query.error}
          retry={() => void query.refetch()}
        />
      ) : !query.data?.items.length ? (
        <div className="px-5 py-14 text-center">
          <CalendarDays className="mx-auto size-9 text-slate-300" />
          <h3 className="mt-4 font-bold text-slate-950">
            No allocation activity
          </h3>
          <p className="mt-1 text-sm text-slate-500">
            No numbers were allocated within this business-date range.
          </p>
        </div>
      ) : (
        <div className="divide-y divide-slate-100">
          {query.data.items.map((counter) => (
            <div
              key={counter.business_date}
              className="flex flex-col gap-2 px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5"
            >
              <div>
                <p className="font-semibold text-slate-950">
                  {dateFormatter.format(
                    new Date(`${counter.business_date}T00:00:00`),
                  )}
                </p>
                <p className="mt-1 text-xs text-slate-500">
                  Updated{" "}
                  {dateTimeFormatter.format(new Date(counter.updated_at))}
                </p>
              </div>
              <div className="rounded-xl bg-slate-100 px-3 py-2 text-right">
                <p className="text-xs text-slate-500">Last sequence</p>
                <p className="font-mono font-bold text-slate-950">
                  {counter.last_number}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
      {!query.isPending && !query.isError && totalPages > 0 ? (
        <div className="flex flex-col gap-3 border-t border-slate-200 px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5">
          <p className="text-sm text-slate-600">
            Page{" "}
            <span className="font-semibold text-slate-950">
              {query.data?.page ?? filters.page}
            </span>{" "}
            of {totalPages}
          </p>
          <div className="flex gap-2">
            <Button
              variant="secondary"
              size="sm"
              disabled={filters.page <= 1}
              onClick={() => onPage(filters.page - 1)}
            >
              <ChevronLeft className="size-4" />
              Previous
            </Button>
            <Button
              variant="secondary"
              size="sm"
              disabled={filters.page >= totalPages}
              onClick={() => onPage(filters.page + 1)}
            >
              Next
              <ChevronRight className="size-4" />
            </Button>
          </div>
        </div>
      ) : null}
    </Panel>
  );
}

function DeactivateRuleDialog({
  rule,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: {
  rule?: DocumentNumberRule;
  pending: boolean;
  error?: Error | null;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog.Root
      open={rule !== undefined}
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
            Deactivate this numbering rule?
          </Dialog.Title>
          <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
            Document ID allocation will stop immediately for this document type.
            Historical rules and counters remain available.
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
