"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import {
  AlertTriangle,
  ChevronLeft,
  ChevronRight,
  Eye,
  LoaderCircle,
  RefreshCw,
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
import { exceptionKeys, listInboundExceptions } from "./inbound-exception-api";
import { ExceptionDetailDialog } from "./exception-detail-dialog";
import {
  EXCEPTION_TYPES,
  exceptionLabel,
  exceptionTime,
  exceptionTone,
} from "./inbound-exception-types";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
export function InboundExceptionsScreen({ timezone }: { timezone: string }) {
  const [filters, setFilters] = useQueryStates({
    owner: parseAsString.withDefault(""),
    warehouse: parseAsString.withDefault(""),
    type: parseAsString.withDefault("all"),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    exception: parseAsString.withDefault(""),
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
      void setFilters({ owner: "", exception: "", page: 1 });
  }, [warehouseOwners.isSuccess, filters.owner, owner, setFilters]);
  const queryFilters = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    exceptionType: filters.type === "all" ? "" : filters.type,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const query = useQuery({
    queryKey: exceptionKeys.list(queryFilters),
    queryFn: () => listInboundExceptions(queryFilters),
    enabled: Boolean(owner && warehouse),
  });
  const error = warehouses.error ?? warehouseOwners.error ?? query.error;
  const rows = query.data?.items ?? [];
  const page = query.data?.page ?? queryFilters.page;
  const totalPages = query.data?.total_pages ?? 0;
  const refresh = () => {
    void warehouses.refetch();
    if (warehouse) void warehouseOwners.refetch();
    if (owner && warehouse) void query.refetch();
  };
  const select = (id: string) => void setFilters({ exception: id });
  return (
    <div className="space-y-6">
      <header>
        <p className="text-sm font-semibold text-slate-600">
          Inbound operations
        </p>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <AlertTriangle className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Inbound exceptions
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Receipt variances, rejection reasons, cancellations and reversals.
            </p>
          </div>
        </div>
      </header>
      <section className="rounded-2xl border border-slate-200 bg-white p-4 text-sm text-slate-600 sm:p-5">
        Exceptions are created automatically by inbound actions. These read-only
        audit records do not have an open/closed status; follow-up is handled
        through the source document’s workflow.
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
              void setFilters({
                warehouse: id,
                owner: "",
                exception: "",
                page: 1,
              })
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
              void setFilters({ owner: id, exception: "", page: 1 })
            }
          />
          <Select
            ariaLabel="Filter exception type"
            value={filters.type}
            options={[
              { value: "all", label: "All exception types" },
              ...EXCEPTION_TYPES.map((code) => ({
                value: code,
                label: exceptionLabel(code),
              })),
              ...(!["all", ...EXCEPTION_TYPES].includes(filters.type)
                ? [{ value: filters.type, label: exceptionLabel(filters.type) }]
                : []),
            ]}
            onValueChange={(type) => void setFilters({ type, page: 1 })}
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search inbound exceptions"
              placeholder="Source document, line or notes"
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
          <p className="text-slate-600">
            {owner && warehouse
              ? `${query.data?.total_items ?? 0} exceptions`
              : "Select your working scope"}
          </p>
          <Button
            variant="ghost"
            size="sm"
            disabled={
              warehouses.isFetching ||
              warehouseOwners.isFetching ||
              query.isFetching
            }
            onClick={refresh}
          >
            <RefreshCw className="size-4" />
            Refresh list
          </Button>
        </div>
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
              <AlertTriangle className="mx-auto size-8 text-slate-400" />
              <p className="mt-3 font-semibold">
                Select a warehouse and served owner
              </p>
              <p className="mt-1 text-sm text-slate-600">
                Only warehouses and owners within your access scope are shown.
                Superadmins can select any warehouse and served owner.
              </p>
            </div>
          </div>
        ) : query.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" />
            Loading exceptions…
          </div>
        ) : !rows.length ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <p className="font-semibold">No inbound exceptions found</p>
              <p className="mt-1 text-sm text-slate-600">
                Try another type or search. Inbound actions create these records
                automatically.
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
                      "Exception / recorded",
                      "Type",
                      "Source document / line",
                      "Expected / actual",
                      "Variance",
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
                      key={value.inbound_exception_id}
                      className="border-t border-slate-100 hover:bg-slate-50"
                    >
                      <td className="px-5 py-4">
                        <p className="font-semibold text-slate-950">
                          {value.inbound_exception_id}
                        </p>
                        <p className="mt-1 text-xs text-slate-500">
                          {exceptionTime(value.created_at, timezone)}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <StatusBadge
                          tone={exceptionTone(value.exception_type_code)}
                        >
                          {exceptionLabel(value.exception_type_code)}
                        </StatusBadge>
                      </td>
                      <td className="px-5 py-4">
                        <p>{value.source_document_id}</p>
                        <p className="mt-1 text-xs text-slate-500">
                          {value.source_line_id ?? "Document-level exception"}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        {value.expected_qty ?? "—"} / {value.actual_qty ?? "—"}
                      </td>
                      <td className="px-5 py-4">{value.variance_qty ?? "—"}</td>
                      <td className="px-5 py-4">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => select(value.inbound_exception_id)}
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
                <article key={value.inbound_exception_id} className="p-4">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <p className="min-w-0 font-semibold break-all text-slate-950">
                      {value.inbound_exception_id}
                    </p>
                    <StatusBadge
                      tone={exceptionTone(value.exception_type_code)}
                    >
                      {exceptionLabel(value.exception_type_code)}
                    </StatusBadge>
                  </div>
                  <p className="mt-2 text-sm break-words text-slate-600">
                    {value.source_document_id}
                  </p>
                  <p className="mt-1 text-xs break-words text-slate-500">
                    {value.source_line_id ?? "Document-level exception"}
                  </p>
                  <dl className="mt-3 grid grid-cols-2 gap-3 text-sm">
                    {[
                      ["Expected", value.expected_qty],
                      ["Actual", value.actual_qty],
                      ["Variance", value.variance_qty],
                      ["Recorded", exceptionTime(value.created_at, timezone)],
                    ].map(([label, quantity]) => (
                      <div key={label}>
                        <dt className="text-xs text-slate-500">{label}</dt>
                        <dd className="break-words">{quantity ?? "—"}</dd>
                      </div>
                    ))}
                  </dl>
                  <Button
                    className="mt-4 w-full"
                    variant="secondary"
                    onClick={() => select(value.inbound_exception_id)}
                  >
                    <Eye className="size-4" />
                    View exception
                  </Button>
                </article>
              ))}
            </div>
          </>
        )}
        {owner && warehouse && query.data && query.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {query.data.total_items} exceptions · page {page} of{" "}
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
      {filters.exception ? (
        <ExceptionDetailDialog
          key={filters.exception}
          exceptionId={filters.exception}
          timezone={timezone}
          onOpenChange={(open) => {
            if (!open) void setFilters({ exception: "" });
          }}
        />
      ) : null}
    </div>
  );
}
