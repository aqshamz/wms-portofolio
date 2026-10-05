"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import {
  ChevronLeft,
  ChevronRight,
  ClipboardCheck,
  Eye,
  Plus,
  Search,
} from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  listWarehouseOwners,
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import { PERMISSIONS } from "@/lib/auth/permissions";
import { listCycleCounts, stockControlKeys } from "./stock-control-api";
import { CycleCountCreateDialog } from "./cycle-count-create-dialog";
import { CycleCountDetailDialog } from "./cycle-count-detail-dialog";
import { GrandStockOpnameCreateDialog } from "./grand-stock-opname-create-dialog";
import type { CycleCountFilters } from "./stock-control-types";

const warehouseFilter = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
const statuses = [
  { value: "", label: "All statuses" },
  { value: "DRAFT", label: "Draft" },
  { value: "COUNTING", label: "Counting / recount" },
  { value: "REVIEW", label: "Ready for review" },
  { value: "PARTIALLY_POSTED", label: "Partially posted" },
  { value: "POSTED", label: "Posted" },
  { value: "CANCELLED", label: "Cancelled" },
];
export function CycleCountsScreen({
  timezone,
  countType = "CYCLE",
}: {
  timezone: string;
  countType?: CycleCountFilters["countType"];
}) {
  const { can } = useAuth();
  const [filters, setFilters] = useQueryStates({
    warehouse: parseAsString.withDefault(""),
    owner: parseAsString.withDefault(""),
    status: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    count: parseAsString.withDefault(""),
    create: parseAsString.withDefault(""),
  });
  const [search, setSearch] = useState(filters.search);
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(warehouseFilter),
    queryFn: () => listWarehouses(warehouseFilter),
  });
  const warehouse = warehouses.data?.items.find(
    (entry) => entry.warehouse_id === filters.warehouse,
  );
  const warehouseOwners = useQuery({
    queryKey: warehouseKeys.owners(filters.warehouse || "none"),
    queryFn: () => listWarehouseOwners(filters.warehouse),
    enabled: Boolean(warehouse),
  });
  const owners = useMemo(
    () => (warehouseOwners.data ?? []).filter((entry) => entry.is_active),
    [warehouseOwners.data],
  );
  const owner = owners.find((entry) => entry.owner_id === filters.owner);
  useEffect(() => {
    if (!filters.warehouse && warehouses.data?.items.length === 1)
      void setFilters({ warehouse: warehouses.data.items[0].warehouse_id });
  }, [filters.warehouse, setFilters, warehouses.data?.items]);
  useEffect(() => {
    if (warehouse && !filters.owner && owners.length === 1)
      void setFilters({ owner: owners[0].owner_id });
  }, [filters.owner, owner, owners, setFilters, warehouse]);
  const request = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    countType,
    status: filters.status,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 20,
  };
  const counts = useQuery({
    queryKey: stockControlKeys.cycleCountList(request),
    queryFn: () => listCycleCounts(request),
    enabled: Boolean(owner && warehouse),
  });
  const page = Math.max(1, filters.page);
  const totalPages = counts.data?.total_pages ?? 0;
  const isGrand = countType === "GRAND";
  const singularLabel = isGrand ? "grand stock opname" : "cycle count";
  const pluralLabel = isGrand ? "grand stock opnames" : "cycle counts";
  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-600">Stock control</p>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <ClipboardCheck className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold">
                {isGrand ? "Grand stock opname" : "Cycle counts"}
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                {isGrand
                  ? "Full blind stock counts for one warehouse and served customer."
                  : "Blind physical counts, recounts and approved variance posting."}
              </p>
            </div>
          </div>
        </div>
        {can(PERMISSIONS.INVENTORY.COUNT) ? (
          <Button
            disabled={!owner || !warehouse}
            onClick={() => void setFilters({ create: "true" })}
          >
            <Plus className="size-4" /> Create {singularLabel}
          </Button>
        ) : null}
      </header>
      <section className="rounded-2xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950">
        {isGrand
          ? "Each document snapshots every positive balance for one served owner in active STORAGE and PICK_FACE locations. Operational locations such as receiving, QC, quarantine and docks remain outside the count."
          : "Counts are blind until review. Variances outside tolerance require a second count. Approval posts only selected lines; matching lines create no inventory movement."}
      </section>
      <Panel className="overflow-hidden">
        <form
          className="grid gap-3 border-b border-slate-200 p-4 md:grid-cols-2 xl:grid-cols-[1fr_1fr_1fr_2fr_auto]"
          onSubmit={(event) => {
            event.preventDefault();
            void setFilters({ search: search.trim(), page: 1 });
          }}
        >
          <Select
            ariaLabel={`${isGrand ? "Grand stock opname" : "Cycle count"} warehouse`}
            value={filters.warehouse}
            options={(warehouses.data?.items ?? []).map((entry) => ({
              value: entry.warehouse_id,
              label: `${entry.name} (${entry.code})`,
            }))}
            placeholder="Select warehouse"
            onValueChange={(warehouseId) =>
              void setFilters({
                warehouse: warehouseId,
                owner: "",
                count: "",
                page: 1,
              })
            }
          />
          <Select
            ariaLabel={`${isGrand ? "Grand stock opname" : "Cycle count"} owner`}
            value={filters.owner}
            options={owners.map((entry) => ({
              value: entry.owner_id,
              label: `${entry.owner_name} (${entry.owner_code})`,
            }))}
            placeholder="Select owner"
            disabled={!warehouse}
            onValueChange={(ownerId) =>
              void setFilters({ owner: ownerId, count: "", page: 1 })
            }
          />
          <Select
            ariaLabel={`${isGrand ? "Grand stock opname" : "Cycle count"} status`}
            value={filters.status}
            options={statuses}
            onValueChange={(status) => void setFilters({ status, page: 1 })}
          />
          <div className="relative">
            <Search className="absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              placeholder="Document, item or location"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>
          <Button type="submit" variant="secondary">
            Search
          </Button>
        </form>
        {counts.error ? (
          <p className="m-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
            {counts.error.message}
          </p>
        ) : !owner || !warehouse ? (
          <Empty text="Select a warehouse and owner." />
        ) : counts.isPending ? (
          <Empty text={`Loading ${pluralLabel}…`} />
        ) : !counts.data?.items.length ? (
          <Empty text={`No ${pluralLabel} match this scope.`} />
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full text-left text-sm">
              <thead className="bg-slate-50 text-xs font-bold text-slate-500 uppercase">
                <tr>
                  {[
                    "Document",
                    "Business date",
                    "Progress",
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
                {counts.data.items.map((entry) => (
                  <tr
                    key={entry.cycle_count_id}
                    className="border-t border-slate-100"
                  >
                    <td className="px-5 py-4 font-semibold">
                      {entry.cycle_count_id}
                    </td>
                    <td className="px-5 py-4">{entry.business_date}</td>
                    <td className="px-5 py-4">
                      {entry.final_lines}/{entry.total_lines} final
                      <p className="text-xs text-slate-500">
                        {entry.recount_lines} recount · {entry.open_lines} open
                      </p>
                    </td>
                    <td className="px-5 py-4">
                      {entry.created_by_display_name}
                    </td>
                    <td className="px-5 py-4">
                      <StatusBadge>{entry.status_code}</StatusBadge>
                    </td>
                    <td className="px-5 py-4">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() =>
                          void setFilters({ count: entry.cycle_count_id })
                        }
                      >
                        <Eye className="size-4" /> View
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {counts.data && counts.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {counts.data.total_items} {pluralLabel} · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || counts.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" /> Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || counts.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>
      {filters.create && owner && warehouse ? (
        isGrand ? (
          <GrandStockOpnameCreateDialog
            ownerId={owner.owner_id}
            ownerName={`${owner.owner_name} (${owner.owner_code})`}
            warehouseId={warehouse.warehouse_id}
            warehouseName={`${warehouse.name} (${warehouse.code})`}
            timezone={timezone}
            onOpenChange={(open) => {
              if (!open) void setFilters({ create: "" });
            }}
          />
        ) : (
          <CycleCountCreateDialog
            ownerId={owner.owner_id}
            warehouseId={warehouse.warehouse_id}
            timezone={timezone}
            onOpenChange={(open) => {
              if (!open) void setFilters({ create: "" });
            }}
          />
        )
      ) : null}
      {filters.count ? (
        <CycleCountDetailDialog
          cycleCountId={filters.count}
          onOpenChange={(open) => {
            if (!open) void setFilters({ count: "" });
          }}
        />
      ) : null}
    </div>
  );
}
function Empty({ text }: { text: string }) {
  return (
    <div className="grid min-h-56 place-items-center p-6 text-sm text-slate-600">
      {text}
    </div>
  );
}
