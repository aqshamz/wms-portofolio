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
  ArrowDown,
  Boxes,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  GitBranch,
  ListOrdered,
  LoaderCircle,
  Pencil,
  Plus,
  Search,
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
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import { listInventoryStatuses } from "@/features/inventory-classifications/inventory-classification-api";
import { listOrganizations } from "@/features/organizations/organization-api";
import {
  deactivatePickingSortMethod,
  deactivatePickingStrategy,
  deactivatePickingStrategyRule,
  listPickingSortMethods,
  listPickingStrategies,
  listPickingStrategyRules,
  pickingConfigurationKeys,
} from "@/features/picking-configuration/picking-configuration-api";
import {
  PickingConfigurationDialog,
  type PickingEditor,
} from "@/features/picking-configuration/picking-configuration-dialog";
import type {
  PickingSortMethod,
  PickingSortMethodPage,
  PickingStrategy,
  PickingStrategyRule,
} from "@/features/picking-configuration/picking-configuration-types";
import { listZones } from "@/features/storage-layout/storage-layout-api";
import { listWarehouses } from "@/features/warehouses/warehouse-api";
import { cn } from "@/lib/utils";

const views = ["methods", "strategies"] as const;
const activeValues = ["all", "active", "inactive"] as const;
const PAGE_SIZE = 10;
const allRecords = {
  search: "",
  active: "all" as const,
  page: 1,
  pageSize: 100,
};
const activeOptions = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;

type DeactivateTarget =
  | { kind: "method"; item: PickingSortMethod }
  | { kind: "strategy"; item: PickingStrategy }
  | { kind: "rule"; strategy: PickingStrategy; item: PickingStrategyRule };

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
    target?.kind === "rule"
      ? "this picking rule"
      : (target?.item.name ?? "configuration");
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
            {target?.kind === "strategy"
              ? "Its ordered rules remain recorded, but new picking work will no longer select this strategy."
              : target?.kind === "rule"
                ? "This condition and sort method will no longer participate in strategy evaluation."
                : "Rules using this method must be changed before they can be reactivated."}
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

export function PickingConfigurationScreen({
  canWrite,
}: {
  canWrite: boolean;
}) {
  const [filters, setFilters] = useQueryStates({
    view: parseAsStringLiteral(views).withDefault("methods"),
    strategy: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    active: parseAsStringLiteral(activeValues).withDefault("all"),
    page: parseAsInteger.withDefault(1),
  });
  const [editor, setEditor] = useState<PickingEditor>();
  const [deactivateTarget, setDeactivateTarget] = useState<DeactivateTarget>();
  const queryClient = useQueryClient();
  const listFilters = {
    search: filters.search,
    active: filters.active,
    page: Math.max(1, filters.page),
    pageSize: filters.view === "strategies" ? 100 : PAGE_SIZE,
  };
  const methods = useQuery({
    queryKey: pickingConfigurationKeys.methods(listFilters),
    queryFn: () => listPickingSortMethods(listFilters),
    enabled: filters.view === "methods",
    placeholderData: keepPreviousData,
  });
  const strategies = useQuery({
    queryKey: pickingConfigurationKeys.strategies(listFilters),
    queryFn: () => listPickingStrategies(listFilters),
    enabled: filters.view === "strategies",
    placeholderData: keepPreviousData,
  });
  const selectedStrategy = useMemo(() => {
    const items = strategies.data?.items ?? [];
    return (
      items.find((item) => item.picking_strategy_id === filters.strategy) ??
      items[0]
    );
  }, [filters.strategy, strategies.data?.items]);
  const rules = useQuery({
    queryKey: pickingConfigurationKeys.rules(
      selectedStrategy?.picking_strategy_id ?? "none",
      allRecords,
    ),
    queryFn: () =>
      listPickingStrategyRules(
        selectedStrategy!.picking_strategy_id,
        allRecords,
      ),
    enabled: Boolean(selectedStrategy) && filters.view === "strategies",
  });
  const allMethods = useQuery({
    queryKey: pickingConfigurationKeys.methods(allRecords),
    queryFn: () => listPickingSortMethods(allRecords),
    enabled: filters.view === "strategies",
  });
  const organizations = useQuery({
    queryKey: pickingConfigurationKeys.organizations(),
    queryFn: () => listOrganizations(allRecords),
    enabled: filters.view === "strategies",
  });
  const warehouses = useQuery({
    queryKey: pickingConfigurationKeys.warehouses(),
    queryFn: () => listWarehouses(allRecords),
    enabled: filters.view === "strategies",
  });
  const inventoryStatuses = useQuery({
    queryKey: pickingConfigurationKeys.inventoryStatuses(),
    queryFn: () => listInventoryStatuses(allRecords),
    enabled: filters.view === "strategies",
  });
  const zones = useQuery({
    queryKey: pickingConfigurationKeys.zones(
      selectedStrategy?.warehouse_id ?? "global",
    ),
    queryFn: () => listZones(selectedStrategy!.warehouse_id!, "all"),
    enabled:
      Boolean(selectedStrategy?.warehouse_id) && filters.view === "strategies",
  });
  const deactivate = useMutation<
    PickingSortMethod | PickingStrategy | PickingStrategyRule,
    Error,
    DeactivateTarget
  >({
    mutationFn: (target) =>
      target.kind === "method"
        ? deactivatePickingSortMethod(target.item.picking_sort_method_id)
        : target.kind === "strategy"
          ? deactivatePickingStrategy(target.item.picking_strategy_id)
          : deactivatePickingStrategyRule(
              target.strategy.picking_strategy_id,
              target.item.rule_id,
            ),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: pickingConfigurationKeys.all,
      });
      toast.success("Picking configuration deactivated.");
      setDeactivateTarget(undefined);
    },
  });
  const ownerById = new Map(
    (organizations.data?.items ?? []).map((item) => [
      item.organization_id,
      item,
    ]),
  );
  const warehouseById = new Map(
    (warehouses.data?.items ?? []).map((item) => [item.warehouse_id, item]),
  );
  const methodById = new Map(
    (allMethods.data?.items ?? []).map((item) => [
      item.picking_sort_method_id,
      item,
    ]),
  );
  const statusById = new Map(
    (inventoryStatuses.data?.items ?? []).map((item) => [
      item.inventory_status_id,
      item,
    ]),
  );
  const zoneById = new Map(
    (zones.data ?? []).map((item) => [item.zone_id, item]),
  );
  const sortedRules = [...(rules.data?.items ?? [])].sort(
    (a, b) => a.sequence_no - b.sequence_no,
  );
  const nextSequence =
    Math.max(0, ...sortedRules.map((rule) => rule.sequence_no)) + 1;
  const activeMethodCount = (allMethods.data?.items ?? []).filter(
    (item) => item.is_active,
  ).length;

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
              <ListOrdered className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Picking configuration
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Control strategy scope, eligible inventory, and location
                ordering.
              </p>
            </div>
          </div>
        </div>
        {canWrite ? (
          <Button
            className="w-full sm:w-auto"
            onClick={() =>
              setEditor(
                filters.view === "methods"
                  ? { kind: "method" }
                  : { kind: "strategy" },
              )
            }
          >
            <Plus className="size-4" />
            Create {filters.view === "methods" ? "sort method" : "strategy"}
          </Button>
        ) : null}
      </header>
      <div
        className="flex gap-2 overflow-x-auto pb-1"
        role="tablist"
        aria-label="Picking configuration sections"
      >
        {views.map((view) => (
          <button
            key={view}
            type="button"
            role="tab"
            aria-selected={filters.view === view}
            onClick={() =>
              void setFilters({
                view,
                search: "",
                active: "all",
                page: 1,
                strategy: "",
              })
            }
            className={cn(
              "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold transition-colors",
              filters.view === view
                ? "bg-slate-950 text-white"
                : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
            )}
          >
            {view === "methods" ? "Sort methods" : "Picking strategies"}
          </button>
        ))}
      </div>
      {filters.view === "methods" ? (
        <MethodsPanel
          query={methods}
          filters={listFilters}
          canWrite={canWrite}
          onSearch={(search) => void setFilters({ search, page: 1 })}
          onActive={(active) => void setFilters({ active, page: 1 })}
          onPage={(page) => void setFilters({ page })}
          onEdit={(item) => setEditor({ kind: "method", item })}
          onDeactivate={(item) => setDeactivateTarget({ kind: "method", item })}
        />
      ) : (
        <div className="grid gap-5 lg:grid-cols-[19rem_minmax(0,1fr)]">
          <Panel className="h-fit overflow-hidden">
            <div className="border-b border-slate-200 p-4">
              <h2 className="font-bold text-slate-950">Picking strategies</h2>
              <p className="mt-1 text-xs text-slate-500">
                Select a strategy to edit its ordered rules.
              </p>
            </div>
            <StrategyFilters
              filters={listFilters}
              onSearch={(search) => void setFilters({ search, strategy: "" })}
              onActive={(active) => void setFilters({ active, strategy: "" })}
            />
            {strategies.isPending ? (
              <LoadingState />
            ) : strategies.isError ? (
              <ErrorState
                title="Strategies could not be loaded"
                error={strategies.error}
                retry={() => void strategies.refetch()}
              />
            ) : !strategies.data.items.length ? (
              <div className="p-8 text-center">
                <Boxes className="mx-auto size-8 text-slate-300" />
                <p className="mt-3 text-sm font-semibold text-slate-900">
                  No strategies found
                </p>
              </div>
            ) : (
              <div className="max-h-[34rem] space-y-1 overflow-y-auto p-2">
                {strategies.data.items.map((item) => {
                  const active =
                    selectedStrategy?.picking_strategy_id ===
                    item.picking_strategy_id;
                  return (
                    <button
                      key={item.picking_strategy_id}
                      type="button"
                      onClick={() =>
                        void setFilters({ strategy: item.picking_strategy_id })
                      }
                      className={cn(
                        "w-full rounded-xl p-3 text-left transition-colors",
                        active
                          ? "bg-slate-950 text-white"
                          : "hover:bg-slate-50",
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
                          "mt-2 block truncate text-xs",
                          active ? "text-slate-300" : "text-slate-500",
                        )}
                      >
                        {scopeLabel(item, ownerById, warehouseById)}
                      </span>
                    </button>
                  );
                })}
              </div>
            )}
          </Panel>
          {!selectedStrategy ? (
            <Panel className="grid min-h-80 place-items-center p-8 text-center">
              <div>
                <GitBranch className="mx-auto size-10 text-slate-300" />
                <h2 className="mt-4 font-bold text-slate-950">
                  Select a strategy
                </h2>
                <p className="mt-1 text-sm text-slate-500">
                  Its ordered rules will appear here.
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
                        {selectedStrategy.name}
                      </h2>
                      <StatusBadge
                        tone={
                          selectedStrategy.is_active ? "success" : "neutral"
                        }
                      >
                        {selectedStrategy.is_active ? "Active" : "Inactive"}
                      </StatusBadge>
                    </div>
                    <p className="mt-1 font-mono text-xs text-slate-500">
                      {selectedStrategy.code}
                    </p>
                    <p className="mt-2 text-sm font-semibold text-cyan-800">
                      {scopeLabel(selectedStrategy, ownerById, warehouseById)}
                    </p>
                    <p className="mt-3 max-w-2xl text-sm leading-6 text-slate-600">
                      {selectedStrategy.description || "No description"}
                    </p>
                  </div>
                  {canWrite ? (
                    <div className="flex shrink-0 gap-1">
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() =>
                          setEditor({
                            kind: "strategy",
                            item: selectedStrategy,
                          })
                        }
                      >
                        <Pencil className="size-4" />
                        Edit
                      </Button>
                      {selectedStrategy.is_active ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-rose-700"
                          onClick={() =>
                            setDeactivateTarget({
                              kind: "strategy",
                              item: selectedStrategy,
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
              <Panel className="overflow-hidden">
                <div className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5">
                  <div>
                    <h2 className="font-bold text-slate-950">
                      Ordered picking rules
                    </h2>
                    <p className="mt-1 text-sm text-slate-500">
                      Rules are evaluated from the lowest sequence number
                      upward.
                    </p>
                  </div>
                  {canWrite ? (
                    <Button
                      disabled={
                        !selectedStrategy.is_active || activeMethodCount === 0
                      }
                      onClick={() =>
                        setEditor({
                          kind: "rule",
                          strategy: selectedStrategy,
                          nextSequence,
                        })
                      }
                    >
                      <Plus className="size-4" />
                      Create rule
                    </Button>
                  ) : null}
                </div>
                {activeMethodCount === 0 && !allMethods.isPending ? (
                  <div className="border-b border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
                    Create an active sort method before adding a picking rule.
                  </div>
                ) : null}
                {rules.isPending ? (
                  <LoadingState />
                ) : rules.isError ? (
                  <ErrorState
                    title="Picking rules could not be loaded"
                    error={rules.error}
                    retry={() => void rules.refetch()}
                  />
                ) : !sortedRules.length ? (
                  <div className="px-5 py-14 text-center">
                    <ListOrdered className="mx-auto size-9 text-slate-300" />
                    <h3 className="mt-4 font-bold text-slate-950">
                      No picking rules found
                    </h3>
                    <p className="mt-1 text-sm text-slate-500">
                      Create the first ordered rule for this strategy.
                    </p>
                  </div>
                ) : (
                  <div className="p-4 sm:p-5">
                    {sortedRules.map((item, index) => (
                      <div key={item.rule_id}>
                        <RuleCard
                          item={item}
                          method={methodById.get(item.picking_sort_method_id)}
                          inventoryStatus={
                            item.inventory_status_id
                              ? statusById.get(item.inventory_status_id)
                              : undefined
                          }
                          zone={
                            item.zone_id
                              ? zoneById.get(item.zone_id)
                              : undefined
                          }
                          canWrite={canWrite}
                          onEdit={() =>
                            setEditor({
                              kind: "rule",
                              strategy: selectedStrategy,
                              nextSequence,
                              item,
                            })
                          }
                          onDeactivate={() =>
                            setDeactivateTarget({
                              kind: "rule",
                              strategy: selectedStrategy,
                              item,
                            })
                          }
                        />
                        {index < sortedRules.length - 1 ? (
                          <ArrowDown className="mx-auto my-2 size-4 text-slate-300" />
                        ) : null}
                      </div>
                    ))}
                  </div>
                )}
              </Panel>
            </div>
          )}
        </div>
      )}
      {editor ? (
        <PickingConfigurationDialog
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

function StrategyFilters({
  filters,
  onSearch,
  onActive,
}: {
  filters: { search: string; active: "all" | "active" | "inactive" };
  onSearch: (search: string) => void;
  onActive: (active: "all" | "active" | "inactive") => void;
}) {
  return (
    <form
      className="space-y-2 border-b border-slate-200 p-3"
      onSubmit={(event) => {
        event.preventDefault();
        const data = new FormData(event.currentTarget);
        onSearch(String(data.get("search") ?? "").trim());
      }}
    >
      <label className="relative block">
        <span className="sr-only">Search strategies</span>
        <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-slate-400" />
        <input
          key={filters.search}
          name="search"
          defaultValue={filters.search}
          placeholder="Search strategies"
          className="h-10 w-full rounded-xl border border-slate-300 pr-3 pl-9 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
        />
      </label>
      <Select
        ariaLabel="Filter strategies by status"
        value={filters.active}
        options={activeOptions}
        onValueChange={onActive}
      />
    </form>
  );
}

function MethodsPanel({
  query,
  filters,
  canWrite,
  onSearch,
  onActive,
  onPage,
  onEdit,
  onDeactivate,
}: {
  query: UseQueryResult<PickingSortMethodPage, Error>;
  filters: {
    search: string;
    active: "all" | "active" | "inactive";
    page: number;
  };
  canWrite: boolean;
  onSearch: (search: string) => void;
  onActive: (active: "all" | "active" | "inactive") => void;
  onPage: (page: number) => void;
  onEdit: (item: PickingSortMethod) => void;
  onDeactivate: (item: PickingSortMethod) => void;
}) {
  const totalPages = query.data?.total_pages ?? 0;
  return (
    <Panel className="overflow-hidden">
      <form
        className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:p-5"
        onSubmit={(event) => {
          event.preventDefault();
          const data = new FormData(event.currentTarget);
          onSearch(String(data.get("search") ?? "").trim());
        }}
      >
        <label className="relative flex-1">
          <span className="sr-only">Search sort methods</span>
          <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
          <input
            key={filters.search}
            name="search"
            defaultValue={filters.search}
            placeholder="Search sort methods"
            className="h-11 w-full rounded-xl border border-slate-300 pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
          />
        </label>
        <Select
          ariaLabel="Filter sort methods by status"
          value={filters.active}
          options={activeOptions}
          onValueChange={onActive}
          className="sm:w-40"
        />
        <Button type="submit" variant="secondary">
          Search
        </Button>
      </form>
      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState
          title="Sort methods could not be loaded"
          error={query.error}
          retry={() => void query.refetch()}
        />
      ) : !query.data.items.length ? (
        <div className="px-5 py-14 text-center">
          <ListOrdered className="mx-auto size-9 text-slate-300" />
          <h2 className="mt-4 font-bold text-slate-950">
            No sort methods found
          </h2>
        </div>
      ) : (
        <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
          {query.data.items.map((item) => (
            <article
              key={item.picking_sort_method_id}
              className="flex min-h-48 flex-col rounded-xl border border-slate-200 p-4"
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <h2 className="font-bold text-slate-950">{item.name}</h2>
                  <p className="mt-0.5 font-mono text-xs text-slate-500">
                    {item.code}
                  </p>
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
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onEdit(item)}
                  >
                    <Pencil className="size-4" />
                    Edit
                  </Button>
                  {item.is_active ? (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-rose-700"
                      onClick={() => onDeactivate(item)}
                    >
                      Deactivate
                    </Button>
                  ) : null}
                </div>
              ) : null}
            </article>
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

function RuleCard({
  item,
  method,
  inventoryStatus,
  zone,
  canWrite,
  onEdit,
  onDeactivate,
}: {
  item: PickingStrategyRule;
  method?: PickingSortMethod;
  inventoryStatus?: { name: string; code: string };
  zone?: { name: string; code: string };
  canWrite: boolean;
  onEdit: () => void;
  onDeactivate: () => void;
}) {
  return (
    <article className="rounded-xl border border-slate-200 p-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex min-w-0 gap-3">
          <div className="grid size-10 shrink-0 place-items-center rounded-xl bg-slate-950 font-mono text-sm font-bold text-cyan-300">
            {item.sequence_no}
          </div>
          <div className="min-w-0">
            <h3 className="font-bold text-slate-950">
              {method?.name ?? "Unknown sort method"}
            </h3>
            <p className="mt-0.5 font-mono text-xs text-slate-500">
              {method?.code ?? item.picking_sort_method_id}
            </p>
            <div className="mt-3 flex flex-wrap gap-2">
              <span className="rounded-full bg-slate-100 px-2.5 py-1 text-xs text-slate-600">
                {inventoryStatus
                  ? `Status: ${inventoryStatus.name} (${inventoryStatus.code})`
                  : "Any pickable status"}
              </span>
              <span className="rounded-full bg-slate-100 px-2.5 py-1 text-xs text-slate-600">
                {zone ? `Zone: ${zone.name} (${zone.code})` : "Any zone"}
              </span>
            </div>
          </div>
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

function scopeLabel(
  strategy: PickingStrategy,
  owners: Map<string, { name: string; code: string }>,
  warehouses: Map<string, { name: string; code: string }>,
) {
  const owner = strategy.owner_id ? owners.get(strategy.owner_id) : undefined;
  const warehouse = strategy.warehouse_id
    ? warehouses.get(strategy.warehouse_id)
    : undefined;
  const ownerLabel = strategy.owner_id
    ? owner
      ? `${owner.name} (${owner.code})`
      : "Specific owner"
    : "All owners";
  const warehouseLabel = strategy.warehouse_id
    ? warehouse
      ? `${warehouse.name} (${warehouse.code})`
      : "Specific warehouse"
    : "All warehouses";
  return `${ownerLabel} · ${warehouseLabel}`;
}
