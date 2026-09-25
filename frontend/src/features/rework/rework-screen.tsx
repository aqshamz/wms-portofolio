"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import {
  ChevronLeft,
  ChevronRight,
  Eye,
  LoaderCircle,
  Search,
  Wrench,
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
import { exceptionTime } from "@/features/inbound-exceptions/inbound-exception-types";
import { listReworkTasks, reworkKeys } from "./rework-api";
import { ReworkDetailDialog } from "./rework-detail-dialog";
import {
  REWORK_STATUSES,
  reworkAssignee,
  reworkLabel,
  reworkTone,
  type ReworkCapabilities,
} from "./rework-types";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
export function ReworkScreen({
  capabilities,
}: {
  capabilities: ReworkCapabilities;
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
  }, [filters.warehouse, warehouses.data?.items, setFilters]);
  useEffect(() => {
    if (warehouse && !filters.owner && owners.length === 1)
      void setFilters({ owner: owners[0].owner_id, page: 1 });
  }, [warehouse, filters.owner, owners, setFilters]);
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
  const query = useQuery({
    queryKey: reworkKeys.list(queryFilters),
    queryFn: () => listReworkTasks(queryFilters),
    enabled: Boolean(owner && warehouse),
  });
  const error = warehouses.error ?? warehouseOwners.error ?? query.error;
  const rows = query.data?.items ?? [];
  const page = query.data?.page ?? queryFilters.page;
  const totalPages = query.data?.total_pages ?? 0;
  const select = (id: string) => void setFilters({ task: id });
  const refresh = () => {
    void warehouses.refetch();
    if (warehouse) void warehouseOwners.refetch();
    if (owner && warehouse) void query.refetch();
  };
  return (
    <div className="space-y-6">
      <header>
        <p className="text-sm font-semibold text-slate-600">
          Inbound operations
        </p>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <Wrench className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Rework tasks
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Repair quarantined stock and send it back for quality inspection.
            </p>
          </div>
        </div>
      </header>
      <section className="rounded-2xl border border-slate-200 bg-white p-4 text-sm text-slate-600 sm:p-5">
        Tasks are created by quarantine REWORK decisions. Starting an open task
        assigns it to you; only the assigned worker can complete it. Completion
        processes the full planned quantity and opens reinspection, not
        available stock.
        {!capabilities.canRework ? (
          <p className="mt-3 rounded-xl bg-amber-50 p-3 text-amber-900">
            This account has read-only access. INBOUND.REWORK is required to
            execute tasks.
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
            onValueChange={(warehouse) =>
              void setFilters({ warehouse, owner: "", task: "", page: 1 })
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
            onValueChange={(owner) =>
              void setFilters({ owner, task: "", page: 1 })
            }
          />
          <Select
            ariaLabel="Filter rework status"
            value={filters.status}
            options={[
              { value: "all", label: "All statuses" },
              ...REWORK_STATUSES.map((status) => ({
                value: status,
                label: reworkLabel(status),
              })),
            ]}
            onValueChange={(status) => void setFilters({ status, page: 1 })}
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search rework tasks"
              maxLength={160}
              placeholder="Task, quarantine case or item"
              value={searchDraft}
              onChange={(event) => setSearchDraft(event.target.value)}
            />
          </div>
          <Button type="submit" variant="secondary">
            Search
          </Button>
        </form>
        {error ? (
          <div
            role="alert"
            className="m-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            <p>{error.message}</p>
            <Button variant="ghost" size="sm" onClick={refresh}>
              Try again
            </Button>
          </div>
        ) : !owner || !warehouse ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <Wrench className="mx-auto size-8 text-slate-400" />
              <p className="mt-3 font-semibold">
                Select a warehouse and served owner
              </p>
              <p className="mt-1 text-sm text-slate-600">
                Tasks are filtered to your account’s access scope. Superadmins
                can select any warehouse and served owner.
              </p>
            </div>
          </div>
        ) : query.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" />
            Loading rework tasks…
          </div>
        ) : !rows.length ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <p className="font-semibold">No rework tasks found</p>
              <p className="mt-1 text-sm text-slate-600">
                Record a REWORK decision on a quarantine case to create a task.
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
                      "Quarantine case",
                      "Planned / completed",
                      "Status / priority",
                      "Assigned account",
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
                      key={task.rework_task_id}
                      className="border-t border-slate-100 hover:bg-slate-50"
                    >
                      <td className="px-5 py-4">
                        <p className="font-semibold text-slate-950">
                          {task.rework_task_id}
                        </p>
                        <p className="mt-1 text-xs text-slate-500">
                          {task.item_code} ·{" "}
                          {exceptionTime(
                            task.created_at,
                            capabilities.timezone,
                          )}
                        </p>
                      </td>
                      <td className="px-5 py-4">{task.quarantine_case_id}</td>
                      <td className="px-5 py-4">
                        {task.planned_qty} / {task.completed_qty} {task.base_uom_code}
                        <p className="mt-1 text-xs text-slate-500">
                          Planned / completed
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <StatusBadge tone={reworkTone(task.task_status_code)}>
                          {reworkLabel(task.task_status_code)}
                        </StatusBadge>
                        <p className="mt-1 text-xs text-slate-500">
                          {reworkLabel(task.task_priority_code)}
                        </p>
                      </td>
                      <td className="max-w-48 px-5 py-4 break-words">
                        {reworkAssignee(task, capabilities.accountId)}
                      </td>
                      <td className="px-5 py-4">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => select(task.rework_task_id)}
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
                <article key={task.rework_task_id} className="p-4">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <p className="min-w-0 font-semibold break-all text-slate-950">
                      {task.rework_task_id}
                    </p>
                    <StatusBadge tone={reworkTone(task.task_status_code)}>
                      {reworkLabel(task.task_status_code)}
                    </StatusBadge>
                  </div>
                  <p className="mt-2 text-sm break-words text-slate-600">
                    {task.item_code} · {task.quarantine_case_id}
                  </p>
                  <dl className="mt-3 grid grid-cols-2 gap-3 text-sm">
                    {[
                      [
                        "Planned quantity",
                        `${task.planned_qty} ${task.base_uom_code}`,
                      ],
                      [
                        "Completed quantity",
                        `${task.completed_qty} ${task.base_uom_code}`,
                      ],
                      ["Priority", reworkLabel(task.task_priority_code)],
                      [
                        "Assigned account",
                        reworkAssignee(task, capabilities.accountId),
                      ],
                    ].map(([label, value]) => (
                      <div key={label}>
                        <dt className="text-xs text-slate-500">{label}</dt>
                        <dd className="break-words">{value}</dd>
                      </div>
                    ))}
                  </dl>
                  <Button
                    className="mt-4 w-full"
                    variant="secondary"
                    onClick={() => select(task.rework_task_id)}
                  >
                    <Eye className="size-4" />
                    View task
                  </Button>
                </article>
              ))}
            </div>
          </>
        )}
        {owner && warehouse && query.data && query.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {query.data.total_items} tasks · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || query.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || query.isFetching}
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
        <ReworkDetailDialog
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
