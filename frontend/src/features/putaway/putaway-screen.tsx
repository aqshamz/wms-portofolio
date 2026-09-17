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
  PackageOpen,
  Search,
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
import { listPutawayTasks, putawayKeys } from "./putaway-api";
import { PutawayDetailDialog } from "./putaway-detail-dialog";
import {
  putawayLabel,
  putawayTone,
  type PutawayCapabilities,
  type PutawayStatus,
} from "./putaway-types";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
const statuses: PutawayStatus[] = [
  "OPEN",
  "ASSIGNED",
  "IN_PROGRESS",
  "COMPLETED",
  "CANCELLED",
  "REVERSED",
];

export function PutawayScreen({
  capabilities,
}: {
  capabilities: PutawayCapabilities;
}) {
  const [filters, setFilters] = useQueryStates({
    owner: parseAsString.withDefault(""),
    warehouse: parseAsString.withDefault(""),
    status: parseAsString.withDefault("all"),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    task: parseAsString.withDefault(""),
  });
  const [searchDraft, setSearchDraft] = useState(filters.search);
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(activeWarehouses),
    queryFn: () => listWarehouses(activeWarehouses),
  });
  const warehouseOwners = useQuery({
    queryKey: warehouseKeys.owners(filters.warehouse || "none"),
    queryFn: () => listWarehouseOwners(filters.warehouse),
    enabled: Boolean(filters.warehouse),
  });
  const owners = useMemo(
    () => (warehouseOwners.data ?? []).filter((owner) => owner.is_active),
    [warehouseOwners.data],
  );
  const owner = owners.find((row) => row.owner_id === filters.owner);
  const warehouse = warehouses.data?.items.find(
    (row) => row.warehouse_id === filters.warehouse,
  );
  useEffect(() => {
    if (!filters.warehouse && warehouses.data?.items.length === 1)
      void setFilters({
        warehouse: warehouses.data.items[0].warehouse_id,
        page: 1,
      });
  }, [filters.warehouse, warehouses.data?.items, setFilters]);
  useEffect(() => {
    if (!filters.owner && owners.length === 1)
      void setFilters({ owner: owners[0].owner_id, page: 1 });
  }, [filters.owner, owners, setFilters]);
  useEffect(() => {
    if (warehouseOwners.isSuccess && filters.owner && !owner)
      void setFilters({ owner: "", task: "", page: 1 });
  }, [warehouseOwners.isSuccess, filters.owner, owner, setFilters]);
  const queryFilters = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    status: filters.status === "all" ? "" : filters.status,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const tasks = useQuery({
    queryKey: putawayKeys.list(queryFilters),
    queryFn: () => listPutawayTasks(queryFilters),
    enabled: Boolean(owner && warehouse),
  });
  const error = warehouses.error ?? warehouseOwners.error ?? tasks.error;
  const rows = tasks.data?.items ?? [];
  const page = tasks.data?.page ?? queryFilters.page;
  const totalPages = tasks.data?.total_pages ?? 0;
  const select = (id: string) => void setFilters({ task: id });

  return (
    <div className="space-y-6">
      <header>
        <p className="text-sm font-semibold text-slate-600">
          Inbound operations
        </p>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <PackageOpen className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Putaway tasks
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Move passed QC stock from receiving to its planned storage
              location.
            </p>
          </div>
        </div>
      </header>
      <section className="rounded-2xl border border-slate-200 bg-white p-4 sm:p-5">
        <div className="flex flex-wrap items-center gap-2 text-xs font-semibold">
          <span className="rounded-full bg-slate-100 px-3 py-1.5 text-slate-700">
            Open
          </span>
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
            Completed · Available in storage
          </span>
        </div>
        <p className="mt-3 text-sm text-slate-600">
          Tasks are created by passed quality inspections, not manually.
          Starting an open task assigns it to you; only its assignee can
          complete it. Assignment and target changes are allowed only before
          work starts.
        </p>
        <p className="mt-2 text-xs text-slate-500">
          INBOUND.PUTAWAY: start / complete · INBOUND.ASSIGN: assign / retarget
          · INBOUND.CANCEL: cancel / reverse. Cancellation and reversal return
          stock to QC with a replacement inspection.
        </p>
        {!capabilities.canPutaway &&
        !capabilities.canAssign &&
        !capabilities.canCancel ? (
          <p className="mt-3 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
            This account has read-only putaway access.
          </p>
        ) : null}
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
            ariaLabel="Filter by warehouse"
            value={filters.warehouse}
            options={(warehouses.data?.items ?? []).map((row) => ({
              value: row.warehouse_id,
              label: `${row.name} (${row.code})`,
            }))}
            placeholder={
              warehouses.isPending ? "Loading warehouses…" : "Select warehouse"
            }
            onValueChange={(id) =>
              void setFilters({ warehouse: id, owner: "", task: "", page: 1 })
            }
          />
          <Select
            ariaLabel="Filter by owner"
            value={filters.owner}
            options={owners.map((row) => ({
              value: row.owner_id,
              label: `${row.owner_name} (${row.owner_code})`,
            }))}
            placeholder="Select served owner"
            disabled={!warehouse || warehouseOwners.isPending}
            onValueChange={(id) =>
              void setFilters({ owner: id, task: "", page: 1 })
            }
          />
          <Select
            ariaLabel="Filter putaway status"
            value={filters.status}
            options={[
              { value: "all", label: "All statuses" },
              ...statuses.map((status) => ({
                value: status,
                label: putawayLabel(status),
              })),
            ]}
            onValueChange={(status) => void setFilters({ status, page: 1 })}
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search putaway tasks"
              placeholder="Task, item or location"
              maxLength={160}
              value={searchDraft}
              onChange={(event) => setSearchDraft(event.target.value)}
            />
          </div>
          <Button type="submit" variant="secondary">
            Search
          </Button>
        </form>
        {error ? (
          <p
            role="alert"
            className="m-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {error.message}
          </p>
        ) : !owner || !warehouse ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <PackageOpen className="mx-auto size-8 text-slate-400" />
              <p className="mt-3 font-semibold">
                Select a warehouse and served owner
              </p>
              <p className="mt-1 text-sm text-slate-600">
                Tasks are filtered to your account’s configured access scope.
              </p>
            </div>
          </div>
        ) : tasks.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" />
            Loading putaway tasks…
          </div>
        ) : !rows.length ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <p className="font-semibold">No putaway tasks found</p>
              <p className="mt-1 text-sm text-slate-600">
                Complete an inspection with passed stock to create a task.
              </p>
            </div>
          </div>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="min-w-full text-left text-sm">
                <thead className="bg-slate-50 text-xs font-bold text-slate-500 uppercase">
                  <tr>
                    {[
                      "Task / item",
                      "Source → target",
                      "Quantity",
                      "Assigned account",
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
                  {rows.map((task) => (
                    <tr
                      key={task.putaway_task_id}
                      className="border-t border-slate-100 hover:bg-slate-50"
                    >
                      <td className="px-5 py-4">
                        <p className="font-semibold text-slate-950">
                          {task.putaway_task_id}
                        </p>
                        <p className="mt-1 text-xs text-slate-500">
                          {task.item_code}
                          {task.lot_number ? ` · ${task.lot_number}` : ""}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <p>{task.source_location_code}</p>
                        <p className="mt-1 text-xs font-semibold text-cyan-800">
                          → {task.target_location_code}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <p>
                          {task.planned_qty} {task.base_uom_code}
                        </p>
                        <p className="mt-1 text-xs text-slate-500">
                          Completed {task.completed_qty}
                        </p>
                      </td>
                      <td className="px-5 py-4 text-xs">
                        {task.assigned_display_name ||
                          task.assigned_to ||
                          "Unassigned"}
                        {task.assigned_to === capabilities.accountId ? (
                          <span className="ml-1 text-cyan-800">(you)</span>
                        ) : null}
                      </td>
                      <td className="px-5 py-4">
                        <StatusBadge tone={putawayTone(task.task_status_code)}>
                          {putawayLabel(task.task_status_code)}
                        </StatusBadge>
                        <p className="mt-1 text-xs text-slate-500">
                          {task.task_priority_code}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => select(task.putaway_task_id)}
                        >
                          <Eye className="size-4" />
                          View
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div className="divide-y divide-slate-100 md:hidden">
              {rows.map((task) => (
                <article key={task.putaway_task_id} className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="font-semibold break-all text-slate-950">
                        {task.putaway_task_id}
                      </p>
                      <p className="mt-1 text-xs text-slate-500">
                        {task.item_code} · {task.lot_number || "No lot"}
                      </p>
                    </div>
                    <StatusBadge tone={putawayTone(task.task_status_code)}>
                      {putawayLabel(task.task_status_code)}
                    </StatusBadge>
                  </div>
                  <dl className="mt-3 grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <dt className="text-xs text-slate-500">From → to</dt>
                      <dd className="break-words">
                        {task.source_location_code} →{" "}
                        {task.target_location_code}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-xs text-slate-500">
                        Planned quantity
                      </dt>
                      <dd>
                        {task.planned_qty} {task.base_uom_code}
                      </dd>
                    </div>
                    <div className="col-span-2">
                      <dt className="text-xs text-slate-500">
                        Assigned account
                      </dt>
                      <dd>
                        {task.assigned_display_name ||
                          task.assigned_to ||
                          "Unassigned"}
                      </dd>
                    </div>
                  </dl>
                  <Button
                    className="mt-4 w-full"
                    variant="secondary"
                    onClick={() => select(task.putaway_task_id)}
                  >
                    <Eye className="size-4" />
                    View task
                  </Button>
                </article>
              ))}
            </div>
          </>
        )}
        {owner && warehouse && tasks.data && tasks.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {tasks.data.total_items} tasks · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || tasks.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || tasks.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>
      {filters.task ? (
        <PutawayDetailDialog
          key={filters.task}
          taskId={filters.task}
          capabilities={capabilities}
          onOpenChange={(open) => {
            if (!open) void setFilters({ task: "" });
          }}
        />
      ) : null}
    </div>
  );
}
