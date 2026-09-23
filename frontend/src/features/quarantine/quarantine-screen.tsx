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
  ShieldAlert,
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
import { listQuarantineCases, quarantineKeys } from "./quarantine-api";
import { QuarantineDetailDialog } from "./quarantine-detail-dialog";
import {
  quarantineLabel,
  quarantineTone,
  remainingQuantity,
  type QuarantineCapabilities,
} from "./quarantine-types";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
export function QuarantineScreen({
  capabilities,
}: {
  capabilities: QuarantineCapabilities;
}) {
  const [filters, setFilters] = useQueryStates({
    owner: parseAsString.withDefault(""),
    warehouse: parseAsString.withDefault(""),
    status: parseAsString.withDefault("all"),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    case: parseAsString.withDefault(""),
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
    () => (warehouseOwners.data ?? []).filter((row) => row.is_active),
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
      void setFilters({ owner: "", case: "", page: 1 });
  }, [warehouseOwners.isSuccess, filters.owner, owner, setFilters]);
  const queryFilters = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    status: filters.status === "all" ? "" : filters.status,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const cases = useQuery({
    queryKey: quarantineKeys.list(queryFilters),
    queryFn: () => listQuarantineCases(queryFilters),
    enabled: Boolean(owner && warehouse),
  });
  const error = warehouses.error ?? warehouseOwners.error ?? cases.error;
  const rows = cases.data?.items ?? [];
  const page = cases.data?.page ?? queryFilters.page;
  const totalPages = cases.data?.total_pages ?? 0;
  const select = (id: string) => void setFilters({ case: id });
  return (
    <div className="space-y-6">
      <header>
        <p className="text-sm font-semibold text-slate-600">
          Inbound operations
        </p>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <ShieldAlert className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Quarantine
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Decide how failed QC stock is accepted, returned, disposed, or
              reworked.
            </p>
          </div>
        </div>
      </header>
      <section className="rounded-2xl border border-slate-200 bg-white p-4 sm:p-5">
        <p className="text-sm text-slate-600">
          Cases are created by failed quality inspections, not manually. Partial
          decisions leave the case partially decided; deciding the full
          quarantine quantity closes it. Closing the case does not mean its
          rework or reinspection has finished.
        </p>
        <p className="mt-2 text-xs text-slate-500">
          INBOUND.READ: view cases and history · INBOUND.QUARANTINE_DISPOSE:
          record decisions. Every decision posts stock immediately; acceptance
          moves it directly into available storage.
        </p>
        {!capabilities.canDispose ? (
          <p className="mt-3 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
            This account has read-only quarantine access.
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
              void setFilters({ warehouse: id, owner: "", case: "", page: 1 })
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
              void setFilters({ owner: id, case: "", page: 1 })
            }
          />
          <Select
            ariaLabel="Filter quarantine status"
            value={filters.status}
            options={[
              { value: "all", label: "All statuses" },
              ...["OPEN", "PARTIALLY_DECIDED", "CLOSED"].map((status) => ({
                value: status,
                label: quarantineLabel(status),
              })),
            ]}
            onValueChange={(status) => void setFilters({ status, page: 1 })}
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search quarantine cases"
              placeholder="Case, item or lot"
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
          <div
            role="alert"
            className="m-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            <p>{error.message}</p>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                void warehouses.refetch();
                if (filters.warehouse) void warehouseOwners.refetch();
                if (owner && warehouse) void cases.refetch();
              }}
            >
              Try again
            </Button>
          </div>
        ) : !owner || !warehouse ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <ShieldAlert className="mx-auto size-8 text-slate-400" />
              <p className="mt-3 font-semibold">
                Select a warehouse and served owner
              </p>
              <p className="mt-1 text-sm text-slate-600">
                Cases are filtered to your account’s access scope. Superadmins
                can select any warehouse and served owner.
              </p>
            </div>
          </div>
        ) : cases.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" />
            Loading quarantine cases…
          </div>
        ) : !rows.length ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <p className="font-semibold">No quarantine cases found</p>
              <p className="mt-1 text-sm text-slate-600">
                Failed inspection quantities create quarantine cases.
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
                      "Case / item",
                      "Lot / location",
                      "Quarantined",
                      "Decided / remaining",
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
                  {rows.map((value) => (
                    <tr
                      key={value.quarantine_case_id}
                      className="border-t border-slate-100 hover:bg-slate-50"
                    >
                      <td className="px-5 py-4">
                        <p className="font-semibold text-slate-950">
                          {value.quarantine_case_id}
                        </p>
                        <p className="mt-1 text-xs text-slate-500">
                          {value.item_code}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <p>{value.lot_number || "No lot"}</p>
                        <p className="mt-1 text-xs text-slate-500">
                          {value.location_code}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        {value.quarantine_qty} {value.base_uom_code}
                      </td>
                      <td className="px-5 py-4">
                        <p>Decided {value.disposed_qty}</p>
                        <p className="mt-1 text-xs font-semibold text-cyan-800">
                          Remaining {remainingQuantity(value)}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <StatusBadge tone={quarantineTone(value.status_code)}>
                          {quarantineLabel(value.status_code)}
                        </StatusBadge>
                      </td>
                      <td className="px-5 py-4">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => select(value.quarantine_case_id)}
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
              {rows.map((value) => (
                <article key={value.quarantine_case_id} className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="font-semibold break-all text-slate-950">
                        {value.quarantine_case_id}
                      </p>
                      <p className="mt-1 text-xs text-slate-500">
                        {value.item_code} · {value.lot_number || "No lot"}
                      </p>
                    </div>
                    <StatusBadge tone={quarantineTone(value.status_code)}>
                      {quarantineLabel(value.status_code)}
                    </StatusBadge>
                  </div>
                  <dl className="mt-3 grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <dt className="text-xs text-slate-500">Location</dt>
                      <dd className="break-words">{value.location_code}</dd>
                    </div>
                    <div>
                      <dt className="text-xs text-slate-500">Quarantined</dt>
                      <dd>
                        {value.quarantine_qty} {value.base_uom_code}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-xs text-slate-500">Decided</dt>
                      <dd>{value.disposed_qty}</dd>
                    </div>
                    <div>
                      <dt className="text-xs text-slate-500">Remaining</dt>
                      <dd>{remainingQuantity(value)}</dd>
                    </div>
                  </dl>
                  <Button
                    className="mt-4 w-full"
                    variant="secondary"
                    onClick={() => select(value.quarantine_case_id)}
                  >
                    <Eye className="size-4" />
                    View case
                  </Button>
                </article>
              ))}
            </div>
          </>
        )}
        {owner && warehouse && cases.data && cases.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {cases.data.total_items} cases · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || cases.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || cases.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>
      {filters.case ? (
        <QuarantineDetailDialog
          key={filters.case}
          caseId={filters.case}
          capabilities={capabilities}
          onOpenChange={(open) => {
            if (!open) void setFilters({ case: "" });
          }}
        />
      ) : null}
    </div>
  );
}
