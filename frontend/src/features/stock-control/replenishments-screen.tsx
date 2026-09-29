"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import {
  ArrowRight,
  ChevronLeft,
  ChevronRight,
  Eye,
  LoaderCircle,
  Plus,
  RefreshCw,
  Search,
  SendToBack,
} from "lucide-react";
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
import { listReplenishments, stockControlKeys } from "./stock-control-api";
import { ReplenishmentCreateDialog } from "./replenishment-create-dialog";
import { ReplenishmentDetailDialog } from "./replenishment-detail-dialog";
import {
  replenishmentLabel,
  replenishmentTone,
  type ReplenishmentStatus,
} from "./replenishment-types";

const statuses: ReplenishmentStatus[] = [
  "OPEN",
  "ASSIGNED",
  "IN_PROGRESS",
  "COMPLETED",
  "CANCELLED",
];
const warehouseFilters = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

export function ReplenishmentsScreen({
  accountId,
  timezone,
}: {
  accountId: string;
  timezone: string;
}) {
  const [filters, setFilters] = useQueryStates({
    warehouse: parseAsString.withDefault(""),
    owner: parseAsString.withDefault(""),
    status: parseAsString.withDefault("all"),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    task: parseAsString.withDefault(""),
    create: parseAsString.withDefault(""),
  });
  const [searchDraft, setSearchDraft] = useState(filters.search);
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(warehouseFilters),
    queryFn: () => listWarehouses(warehouseFilters),
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
    if (!filters.warehouse && warehouses.data?.items.length === 1)
      void setFilters({
        warehouse: warehouses.data.items[0].warehouse_id,
        page: 1,
      });
  }, [filters.warehouse, setFilters, warehouses.data?.items]);
  useEffect(() => {
    if (!filters.owner && owners.length === 1)
      void setFilters({ owner: owners[0].owner_id, page: 1 });
  }, [filters.owner, owners, setFilters]);
  useEffect(() => {
    if (warehouseOwners.isSuccess && filters.owner && !owner)
      void setFilters({ owner: "", task: "", page: 1 });
  }, [filters.owner, owner, setFilters, warehouseOwners.isSuccess]);
  const request = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    status: filters.status === "all" ? "" : filters.status,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const tasks = useQuery({
    queryKey: stockControlKeys.replenishmentList(request),
    queryFn: () => listReplenishments(request),
    enabled: Boolean(owner && warehouse),
  });
  const error = warehouses.error ?? warehouseOwners.error ?? tasks.error;
  const rows = tasks.data?.items ?? [];
  const page = tasks.data?.page ?? request.page;
  const totalPages = tasks.data?.total_pages ?? 0;

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-600">Stock control</p>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <SendToBack className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
                Replenishments
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Move reserve stock into forward pick faces before outbound
                picking.
              </p>
            </div>
          </div>
        </div>
        <Button
          disabled={!owner || !warehouse}
          onClick={() => void setFilters({ create: "new" })}
        >
          <Plus className="size-4" /> Create replenishment
        </Button>
      </header>

      <section className="rounded-2xl border border-slate-200 bg-white p-4 sm:p-5">
        <div className="flex flex-wrap items-center gap-2 text-xs font-semibold">
          <span className="rounded-full bg-slate-100 px-3 py-1.5">Open</span>
          <ArrowRight className="size-4 text-slate-400" />
          <span className="rounded-full bg-cyan-50 px-3 py-1.5 text-cyan-900">
            Assigned (optional)
          </span>
          <ArrowRight className="size-4 text-slate-400" />
          <span className="rounded-full bg-amber-50 px-3 py-1.5 text-amber-900">
            In progress
          </span>
          <ArrowRight className="size-4 text-slate-400" />
          <span className="rounded-full bg-emerald-50 px-3 py-1.5 text-emerald-900">
            Completed · Stock at pick face
          </span>
        </div>
        <p className="mt-3 text-sm text-slate-600">
          Task creation reserves stock without moving it. An open task can be
          claimed by a worker; an assigned task is restricted to that account.
          Completion posts one immutable REPLENISHMENT movement.
        </p>
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
            ariaLabel="Filter replenishments by warehouse"
            value={filters.warehouse}
            options={(warehouses.data?.items ?? []).map((row) => ({
              value: row.warehouse_id,
              label: `${row.name} (${row.code})`,
            }))}
            placeholder={
              warehouses.isPending ? "Loading warehouses…" : "Select warehouse"
            }
            onValueChange={(value) =>
              void setFilters({
                warehouse: value,
                owner: "",
                task: "",
                page: 1,
              })
            }
          />
          <Select
            ariaLabel="Filter replenishments by owner"
            value={filters.owner}
            options={owners.map((row) => ({
              value: row.owner_id,
              label: `${row.owner_name} (${row.owner_code})`,
            }))}
            placeholder="Select served owner"
            disabled={!warehouse || warehouseOwners.isPending}
            onValueChange={(value) =>
              void setFilters({ owner: value, task: "", page: 1 })
            }
          />
          <Select
            ariaLabel="Filter replenishment status"
            value={filters.status}
            options={[
              { value: "all", label: "All statuses" },
              ...statuses.map((status) => ({
                value: status,
                label: replenishmentLabel(status),
              })),
            ]}
            onValueChange={(value) =>
              void setFilters({ status: value, page: 1 })
            }
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search replenishments"
              placeholder="Task, item or location"
              value={searchDraft}
              onChange={(event) => setSearchDraft(event.target.value)}
            />
          </div>
          <Button type="submit" variant="secondary">
            Search
          </Button>
        </form>
        <div className="flex items-center justify-between border-b border-slate-200 px-5 py-3 text-sm">
          <p className="font-semibold">
            {tasks.data?.total_items ?? 0} replenishment tasks
          </p>
          <Button
            variant="ghost"
            size="sm"
            disabled={tasks.isFetching}
            onClick={() => void tasks.refetch()}
          >
            <RefreshCw className="size-4" /> Refresh
          </Button>
        </div>
        {error ? (
          <p
            role="alert"
            className="m-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {error.message}
          </p>
        ) : !owner || !warehouse ? (
          <Empty text="Select a warehouse and served owner." />
        ) : tasks.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" /> Loading
            replenishments…
          </div>
        ) : !rows.length ? (
          <Empty text="No replenishment tasks match this scope." />
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full text-left text-sm">
              <thead className="bg-slate-50 text-xs font-bold text-slate-500 uppercase">
                <tr>
                  {[
                    "Task / item",
                    "Source → pick face",
                    "Planned / completed",
                    "Status / priority",
                    "Assigned",
                    "Action",
                  ].map((heading) => (
                    <th key={heading} className="px-5 py-3">
                      {heading}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rows.map((task) => (
                  <tr
                    key={task.replenishment_task_id}
                    className="border-t border-slate-100 hover:bg-slate-50"
                  >
                    <td className="px-5 py-4">
                      <p className="font-semibold">
                        {task.replenishment_task_id}
                      </p>
                      <p className="mt-1 text-xs text-slate-500">
                        {task.item_code} ·{" "}
                        {task.lot_number ??
                          task.serial_number ??
                          task.handling_unit_barcode ??
                          "No tracked identity"}
                      </p>
                    </td>
                    <td className="px-5 py-4">
                      {task.source_location_code} → {task.target_location_code}
                    </td>
                    <td className="px-5 py-4">
                      {task.planned_qty} / {task.completed_qty} {task.uom_code}
                    </td>
                    <td className="px-5 py-4">
                      <StatusBadge
                        tone={replenishmentTone(task.task_status_code)}
                      >
                        {replenishmentLabel(task.task_status_code)}
                      </StatusBadge>
                      <p className="mt-1 text-xs text-slate-500">
                        {task.task_priority_code}
                      </p>
                    </td>
                    <td className="px-5 py-4">
                      {task.assigned_display_name ??
                        task.assigned_username ??
                        "Unassigned"}
                    </td>
                    <td className="px-5 py-4">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() =>
                          void setFilters({ task: task.replenishment_task_id })
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
        {tasks.data && tasks.data.total_items > 0 ? (
          <footer className="flex items-center justify-between border-t border-slate-200 p-4 text-sm">
            <p>
              Page {page} of {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" /> Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>

      {filters.create && owner && warehouse ? (
        <ReplenishmentCreateDialog
          ownerId={owner.owner_id}
          warehouseId={warehouse.warehouse_id}
          onOpenChange={(open) => {
            if (!open) void setFilters({ create: "" });
          }}
        />
      ) : null}
      {filters.task ? (
        <ReplenishmentDetailDialog
          taskId={filters.task}
          accountId={accountId}
          timezone={timezone}
          onOpenChange={(open) => {
            if (!open) void setFilters({ task: "" });
          }}
        />
      ) : null}
    </div>
  );
}

function Empty({ text }: { text: string }) {
  return (
    <div className="grid min-h-56 place-items-center p-6 text-center">
      <div>
        <SendToBack className="mx-auto size-8 text-slate-400" />
        <p className="mt-3 font-semibold">No replenishment work</p>
        <p className="mt-1 text-sm text-slate-600">{text}</p>
      </div>
    </div>
  );
}
