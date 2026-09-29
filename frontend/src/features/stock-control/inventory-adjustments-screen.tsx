"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import {
  ChevronLeft,
  ChevronRight,
  Eye,
  Plus,
  RefreshCw,
  Search,
  ShieldCheck,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/components/auth/auth-provider";
import { Input } from "@/components/ui/input";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import { PERMISSIONS } from "@/lib/auth/permissions";
import {
  listWarehouseOwners,
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import { listAdjustments, stockControlKeys } from "./stock-control-api";
import type { InventoryAdjustment } from "./stock-control-types";
import { AdjustmentCreateDialog } from "./adjustment-create-dialog";
import { AdjustmentDetailDialog } from "./adjustment-detail-dialog";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

const statusOptions = [
  { value: "", label: "All statuses" },
  { value: "DRAFT", label: "Pending approval" },
  { value: "PARTIALLY_POSTED", label: "Partially posted" },
  { value: "POSTED", label: "Posted" },
  { value: "CANCELLED", label: "Cancelled / rejected" },
];

export function InventoryAdjustmentsScreen({ timezone }: { timezone: string }) {
  const { can } = useAuth();
  const [filters, setFilters] = useQueryStates({
    warehouse: parseAsString.withDefault(""),
    owner: parseAsString.withDefault(""),
    status: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    adjustment: parseAsString.withDefault(""),
    create: parseAsString.withDefault(""),
  });
  const [searchDraft, setSearchDraft] = useState(filters.search);
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(activeWarehouses),
    queryFn: () => listWarehouses(activeWarehouses),
  });
  const warehouse = warehouses.data?.items.find(
    (row) => row.warehouse_id === filters.warehouse,
  );
  const warehouseOwners = useQuery({
    queryKey: warehouseKeys.owners(filters.warehouse || "none"),
    queryFn: () => listWarehouseOwners(filters.warehouse),
    enabled: Boolean(warehouse),
  });
  const owners = useMemo(
    () => (warehouseOwners.data ?? []).filter((row) => row.is_active),
    [warehouseOwners.data],
  );
  const owner = owners.find((row) => row.owner_id === filters.owner);

  useEffect(() => {
    if (!filters.warehouse && warehouses.data?.items.length === 1) {
      void setFilters({ warehouse: warehouses.data.items[0].warehouse_id });
    }
  }, [filters.warehouse, setFilters, warehouses.data?.items]);
  useEffect(() => {
    if (warehouse && !filters.owner && owners.length === 1) {
      void setFilters({ owner: owners[0].owner_id });
    }
  }, [filters.owner, owners, setFilters, warehouse]);
  useEffect(() => {
    if (warehouseOwners.isSuccess && filters.owner && !owner) {
      void setFilters({ owner: "", adjustment: "", page: 1 });
    }
  }, [filters.owner, owner, setFilters, warehouseOwners.isSuccess]);

  const request = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    status: filters.status,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const adjustments = useQuery({
    queryKey: stockControlKeys.adjustmentList(request),
    queryFn: () => listAdjustments(request),
    enabled: Boolean(owner && warehouse),
  });
  const error = warehouses.error ?? warehouseOwners.error ?? adjustments.error;
  const page = adjustments.data?.page ?? request.page;
  const totalPages = adjustments.data?.total_pages ?? 0;

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-600">Stock control</p>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <ShieldCheck className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Inventory adjustments
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Request controlled quantity corrections and review their audit
                trail.
              </p>
            </div>
          </div>
        </div>
        {can(PERMISSIONS.INVENTORY.ADJUST) ? (
          <Button
            disabled={!owner || !warehouse}
            onClick={() => void setFilters({ create: "true" })}
          >
            <Plus className="size-4" /> Request adjustment
          </Button>
        ) : null}
      </header>

      <section className="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950 sm:p-5">
        Creating a request does not change inventory. A different account with
        approval permission can review selected lines. Approved lines post
        atomically, while other lines remain pending for a later decision.
      </section>

      <Panel className="overflow-hidden">
        <form
          className="grid gap-3 border-b border-slate-200 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-[1fr_1fr_1fr_2fr_auto]"
          onSubmit={(event) => {
            event.preventDefault();
            void setFilters({ search: searchDraft.trim(), page: 1 });
          }}
        >
          <Select
            ariaLabel="Filter adjustments by warehouse"
            value={filters.warehouse}
            options={(warehouses.data?.items ?? []).map((row) => ({
              value: row.warehouse_id,
              label: `${row.name} (${row.code})`,
            }))}
            placeholder="Select warehouse"
            onValueChange={(warehouseId) =>
              void setFilters({
                warehouse: warehouseId,
                owner: "",
                adjustment: "",
                page: 1,
              })
            }
          />
          <Select
            ariaLabel="Filter adjustments by owner"
            value={filters.owner}
            options={owners.map((row) => ({
              value: row.owner_id,
              label: `${row.owner_name} (${row.owner_code})`,
            }))}
            placeholder="Select served owner"
            disabled={!warehouse || warehouseOwners.isPending}
            onValueChange={(ownerId) =>
              void setFilters({ owner: ownerId, adjustment: "", page: 1 })
            }
          />
          <Select
            ariaLabel="Filter adjustments by status"
            value={filters.status}
            options={statusOptions}
            onValueChange={(status) => void setFilters({ status, page: 1 })}
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search adjustment requests"
              placeholder="Request, item, location or reason"
              maxLength={160}
              value={searchDraft}
              onChange={(event) => setSearchDraft(event.target.value)}
            />
          </div>
          <Button type="submit" variant="secondary">
            Search
          </Button>
        </form>

        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3 text-sm sm:px-5">
          <div>
            <p className="font-semibold text-slate-950">Adjustment requests</p>
            <p className="text-xs text-slate-500">
              {owner && warehouse
                ? `${adjustments.data?.total_items ?? 0} requests in this scope`
                : "Select your working scope"}
            </p>
          </div>
          <Button
            variant="ghost"
            size="sm"
            disabled={adjustments.isFetching}
            onClick={() => {
              void warehouses.refetch();
              if (warehouse) void warehouseOwners.refetch();
              if (owner && warehouse) void adjustments.refetch();
            }}
          >
            <RefreshCw className="size-4" /> Refresh
          </Button>
        </div>

        {error ? (
          <div
            role="alert"
            className="m-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {error.message}
          </div>
        ) : !owner || !warehouse ? (
          <EmptyState text="Only requests within your account’s owner and warehouse access are shown." />
        ) : adjustments.isPending ? (
          <div className="grid min-h-56 place-items-center text-sm text-slate-600">
            Loading adjustment requests…
          </div>
        ) : !adjustments.data?.items.length ? (
          <EmptyState text="No adjustment request matches this scope and filter." />
        ) : (
          <AdjustmentRows
            rows={adjustments.data.items}
            onView={(id) => void setFilters({ adjustment: id })}
          />
        )}

        {adjustments.data && adjustments.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {adjustments.data.total_items} requests · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || adjustments.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" /> Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || adjustments.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>

      {filters.create && owner && warehouse ? (
        <AdjustmentCreateDialog
          ownerId={owner.owner_id}
          warehouseId={warehouse.warehouse_id}
          timezone={timezone}
          onOpenChange={(open) => {
            if (!open) void setFilters({ create: "" });
          }}
        />
      ) : null}
      {filters.adjustment ? (
        <AdjustmentDetailDialog
          adjustmentId={filters.adjustment}
          onOpenChange={(open) => {
            if (!open) void setFilters({ adjustment: "" });
          }}
        />
      ) : null}
    </div>
  );
}

function tone(status: string) {
  if (status === "POSTED") return "success" as const;
  if (status === "CANCELLED") return "danger" as const;
  return "warning" as const;
}

function AdjustmentRows({
  rows,
  onView,
}: {
  rows: InventoryAdjustment[];
  onView: (id: string) => void;
}) {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full text-left text-sm">
        <thead className="bg-slate-50 text-xs font-bold text-slate-500 uppercase">
          <tr>
            {[
              "Document",
              "Direction / reason",
              "Line progress",
              "Requested by",
              "Status",
              "Action",
            ].map((heading) => (
              <th key={heading} className="px-5 py-3">
                {heading}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr
              key={row.inventory_adjustment_id}
              className="border-t border-slate-100 hover:bg-slate-50"
            >
              <td className="px-5 py-4">
                <p className="font-semibold text-slate-950">
                  {row.inventory_adjustment_id}
                </p>
                <p className="mt-1 text-xs text-slate-500">
                  {row.owner_code} · {row.warehouse_code}
                </p>
              </td>
              <td className="px-5 py-4">
                <p
                  className={
                    row.direction === "INCREASE"
                      ? "font-semibold text-emerald-700"
                      : "font-semibold text-rose-700"
                  }
                >
                  {row.direction}
                </p>
                <p className="mt-1 text-xs text-slate-500">{row.reason_name}</p>
              </td>
              <td className="px-5 py-4">
                <p className="font-semibold">{row.total_lines} lines</p>
                <p className="mt-1 text-xs text-slate-500">
                  {row.posted_lines} posted · {row.pending_lines} pending ·{" "}
                  {row.rejected_lines + row.cancelled_lines} closed
                </p>
              </td>
              <td className="px-5 py-4">
                <p>{row.created_by_display_name}</p>
                <p className="mt-1 text-xs text-slate-500">
                  @{row.created_by_username}
                </p>
              </td>
              <td className="px-5 py-4">
                <StatusBadge tone={tone(row.status_code)}>
                  {row.status_code}
                </StatusBadge>
              </td>
              <td className="px-5 py-4">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onView(row.inventory_adjustment_id)}
                >
                  <Eye className="size-4" /> View
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function EmptyState({ text }: { text: string }) {
  return (
    <div className="grid min-h-56 place-items-center p-6 text-center">
      <div>
        <ShieldCheck className="mx-auto size-8 text-slate-400" />
        <p className="mt-3 font-semibold">No adjustment requests</p>
        <p className="mt-1 text-sm text-slate-600">{text}</p>
      </div>
    </div>
  );
}
