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
  Boxes,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  LoaderCircle,
  Pencil,
  Plus,
  Search,
  ShieldAlert,
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
import {
  deactivateInventoryStatus,
  inventoryClassificationKeys,
  listInventoryStatuses,
} from "@/features/inventory-classifications/inventory-classification-api";
import { InventoryStatusDialog } from "@/features/inventory-classifications/inventory-status-dialog";
import type { InventoryStatus } from "@/features/inventory-classifications/inventory-classification-types";

const PAGE_SIZE = 10;
const activeValues = ["all", "active", "inactive"] as const;
const statusOptions = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;

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
            <p className="font-semibold">
              Inventory classifications could not be loaded
            </p>
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

function DeactivateDialog({
  item,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: {
  item?: InventoryStatus;
  pending: boolean;
  error?: Error | null;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog.Root
      open={item !== undefined}
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
            Deactivate {item?.name ?? "classification"}?
          </Dialog.Title>
          <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
            Existing inventory retains this status, but it cannot be selected as
            the target status for new inventory movements. Confirm that
            operational workflows no longer depend on it.
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

export function InventoryClassificationsScreen({
  canWrite,
}: {
  canWrite: boolean;
}) {
  const [filters, setFilters] = useQueryStates({
    search: parseAsString.withDefault(""),
    active: parseAsStringLiteral(activeValues).withDefault("all"),
    page: parseAsInteger.withDefault(1),
  });
  const [editor, setEditor] = useState<{ item?: InventoryStatus }>();
  const [deactivateTarget, setDeactivateTarget] = useState<InventoryStatus>();
  const queryClient = useQueryClient();
  const listFilters = {
    search: filters.search,
    active: filters.active,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const statuses = useQuery({
    queryKey: inventoryClassificationKeys.list(listFilters),
    queryFn: () => listInventoryStatuses(listFilters),
    placeholderData: keepPreviousData,
  });
  const deactivate = useMutation({
    mutationFn: (item: InventoryStatus) =>
      deactivateInventoryStatus(item.inventory_status_id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: inventoryClassificationKeys.all,
      });
      toast.success("Inventory classification deactivated.");
      setDeactivateTarget(undefined);
    },
  });
  const items = statuses.data?.items ?? [];
  const currentPage = statuses.data?.page ?? listFilters.page;
  const totalPages = statuses.data?.total_pages ?? 0;

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
              <ShieldAlert className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Inventory classifications
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Define whether inventory states are available for allocation and
                picking.
              </p>
            </div>
          </div>
        </div>
        {canWrite ? (
          <Button className="w-full sm:w-auto" onClick={() => setEditor({})}>
            <Plus className="size-4" />
            Create classification
          </Button>
        ) : null}
      </header>

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
            <span className="sr-only">Search classifications</span>
            <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
            <input
              key={filters.search}
              name="search"
              defaultValue={filters.search}
              placeholder="Search by code or name"
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
        {!statuses.isPending && !statuses.isError ? (
          <div className="border-b border-slate-200 px-4 py-3 text-sm text-slate-600 sm:px-5">
            <span className="font-semibold text-slate-950">
              {statuses.data?.total_items ?? 0}
            </span>{" "}
            classifications
          </div>
        ) : null}
        {statuses.isPending ? (
          <LoadingState />
        ) : statuses.isError ? (
          <ErrorState
            error={statuses.error}
            retry={() => void statuses.refetch()}
          />
        ) : !items.length ? (
          <div className="px-5 py-14 text-center">
            <Boxes className="mx-auto size-9 text-slate-300" />
            <h2 className="mt-4 font-bold text-slate-950">
              No inventory classifications found
            </h2>
            <p className="mt-1 text-sm text-slate-500">
              Adjust the filters or create the first classification.
            </p>
          </div>
        ) : (
          <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
            {items.map((item) => (
              <article
                key={item.inventory_status_id}
                className="flex min-h-56 flex-col rounded-xl border border-slate-200 p-4"
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
                <p className="mt-3 min-h-10 text-sm leading-6 text-slate-600">
                  {item.description || "No description"}
                </p>
                <div className="mt-4 grid flex-1 grid-cols-2 gap-2">
                  <div
                    className={
                      item.is_allocatable
                        ? "rounded-xl bg-emerald-50 p-3 text-emerald-800"
                        : "rounded-xl bg-slate-50 p-3 text-slate-500"
                    }
                  >
                    {item.is_allocatable ? (
                      <CheckCircle2 className="size-4" />
                    ) : (
                      <X className="size-4" />
                    )}
                    <p className="mt-2 text-xs font-semibold">
                      {item.is_allocatable ? "Allocatable" : "Not allocatable"}
                    </p>
                  </div>
                  <div
                    className={
                      item.is_pickable
                        ? "rounded-xl bg-cyan-50 p-3 text-cyan-800"
                        : "rounded-xl bg-slate-50 p-3 text-slate-500"
                    }
                  >
                    {item.is_pickable ? (
                      <CheckCircle2 className="size-4" />
                    ) : (
                      <X className="size-4" />
                    )}
                    <p className="mt-2 text-xs font-semibold">
                      {item.is_pickable ? "Pickable" : "Not pickable"}
                    </p>
                  </div>
                </div>
                {canWrite ? (
                  <div className="mt-4 flex justify-end gap-1 border-t border-slate-100 pt-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setEditor({ item })}
                    >
                      <Pencil className="size-4" />
                      Edit
                    </Button>
                    {item.is_active ? (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-rose-700"
                        onClick={() => setDeactivateTarget(item)}
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
        {!statuses.isPending && !statuses.isError && totalPages > 0 ? (
          <div className="flex flex-col gap-3 border-t border-slate-200 px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5">
            <p className="text-sm text-slate-600">
              Page{" "}
              <span className="font-semibold text-slate-900">
                {currentPage}
              </span>{" "}
              of {totalPages}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={currentPage <= 1}
                onClick={() => void setFilters({ page: currentPage - 1 })}
              >
                <ChevronLeft className="size-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={currentPage >= totalPages}
                onClick={() => void setFilters({ page: currentPage + 1 })}
              >
                Next
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </div>
        ) : null}
      </Panel>
      {editor ? (
        <InventoryStatusDialog
          item={editor.item}
          onOpenChange={(open) => {
            if (!open) setEditor(undefined);
          }}
        />
      ) : null}
      <DeactivateDialog
        item={deactivateTarget}
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
