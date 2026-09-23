"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import {
  Barcode,
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
import { listSerialStates, serialStateKeys } from "./inventory-api";
import { SerialStateDetailDialog } from "./serial-state-detail-dialog";
import type { SerialState, SerialStateFilters } from "./inventory-types";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

function updatedAt(value: string, timezone: string) {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: timezone,
  }).format(new Date(value));
}

export function InventorySerialStatesScreen({
  timezone,
}: {
  timezone: string;
}) {
  const [filters, setFilters] = useQueryStates({
    owner: parseAsString.withDefault(""),
    warehouse: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    record: parseAsString.withDefault(""),
  });
  const [searchDraft, setSearchDraft] = useState(filters.search);
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(activeWarehouses),
    queryFn: () => listWarehouses(activeWarehouses),
  });
  const warehouse = warehouses.data?.items.find(
    (value) => value.warehouse_id === filters.warehouse,
  );
  const warehouseOwners = useQuery({
    queryKey: warehouseKeys.owners(filters.warehouse || "none"),
    queryFn: () => listWarehouseOwners(filters.warehouse),
    enabled: Boolean(warehouse),
  });
  const owners = useMemo(
    () => (warehouseOwners.data ?? []).filter((value) => value.is_active),
    [warehouseOwners.data],
  );
  const owner = owners.find((value) => value.owner_id === filters.owner);

  useEffect(() => {
    if (!filters.warehouse && warehouses.data?.items.length === 1) {
      void setFilters({
        warehouse: warehouses.data.items[0].warehouse_id,
        page: 1,
      });
    }
  }, [filters.warehouse, setFilters, warehouses.data?.items]);
  useEffect(() => {
    if (warehouse && !filters.owner && owners.length === 1) {
      void setFilters({ owner: owners[0].owner_id, page: 1 });
    }
  }, [filters.owner, owners, setFilters, warehouse]);
  useEffect(() => {
    if (warehouseOwners.isSuccess && filters.owner && !owner) {
      void setFilters({ owner: "", record: "", page: 1 });
    }
  }, [filters.owner, owner, setFilters, warehouseOwners.isSuccess]);

  const request: SerialStateFilters = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const serialStates = useQuery({
    queryKey: serialStateKeys.list(request),
    queryFn: () => listSerialStates(request),
    enabled: Boolean(owner && warehouse),
  });
  const error = warehouses.error ?? warehouseOwners.error ?? serialStates.error;
  const page = serialStates.data?.page ?? request.page;
  const totalPages = serialStates.data?.total_pages ?? 0;

  return (
    <div className="space-y-6">
      <header>
        <p className="text-sm font-semibold text-slate-600">Inventory</p>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <Barcode className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Serial states
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Current warehouse position and inventory state of serialized
              units.
            </p>
          </div>
        </div>
      </header>

      <section className="rounded-2xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950 sm:p-5">
        This list contains serials currently in WMS stock. Shipped or removed
        serials leave current state, while their immutable history remains in
        Inventory movements.
      </section>

      <Panel className="overflow-hidden">
        <form
          className="grid gap-3 border-b border-slate-200 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-[1fr_1fr_2fr_auto]"
          onSubmit={(event) => {
            event.preventDefault();
            void setFilters({ search: searchDraft.trim(), page: 1 });
          }}
        >
          <Select
            ariaLabel="Filter serial states by warehouse"
            value={filters.warehouse}
            options={(warehouses.data?.items ?? []).map((value) => ({
              value: value.warehouse_id,
              label: `${value.name} (${value.code})`,
            }))}
            placeholder={
              warehouses.isPending ? "Loading warehouses…" : "Select warehouse"
            }
            onValueChange={(warehouseId) =>
              void setFilters({
                warehouse: warehouseId,
                owner: "",
                record: "",
                page: 1,
              })
            }
          />
          <Select
            ariaLabel="Filter serial states by owner"
            value={filters.owner}
            options={owners.map((value) => ({
              value: value.owner_id,
              label: `${value.owner_name} (${value.owner_code})`,
            }))}
            placeholder="Select served owner"
            disabled={!warehouse || warehouseOwners.isPending}
            onValueChange={(ownerId) =>
              void setFilters({ owner: ownerId, record: "", page: 1 })
            }
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search serial states"
              placeholder="Serial number, item, location or lot"
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
              ? `${serialStates.data?.total_items ?? 0} serial states`
              : "Select your working scope"}
          </p>
          <Button
            variant="ghost"
            size="sm"
            disabled={serialStates.isFetching}
            onClick={() => {
              void warehouses.refetch();
              if (warehouse) void warehouseOwners.refetch();
              if (owner && warehouse) void serialStates.refetch();
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
          <EmptyState text="Only serial states within your access scope will be requested." />
        ) : serialStates.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" /> Loading serial
            states…
          </div>
        ) : !serialStates.data?.items.length ? (
          <EmptyState
            text="Try another serial, item, location or lot search."
            empty
          />
        ) : (
          <SerialStateRows
            rows={serialStates.data.items}
            timezone={timezone}
            onSelect={(record) => void setFilters({ record })}
          />
        )}

        {owner &&
        warehouse &&
        serialStates.data &&
        serialStates.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {serialStates.data.total_items} serial states · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || serialStates.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" /> Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || serialStates.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>

      {filters.record ? (
        <SerialStateDetailDialog
          key={filters.record}
          id={filters.record}
          timezone={timezone}
          onOpenChange={(open) => {
            if (!open) void setFilters({ record: "" });
          }}
        />
      ) : null}
    </div>
  );
}

function EmptyState({
  text,
  empty = false,
}: {
  text: string;
  empty?: boolean;
}) {
  return (
    <div className="grid min-h-56 place-items-center p-6 text-center">
      <div>
        <Barcode className="mx-auto size-8 text-slate-400" />
        <p className="mt-3 font-semibold">
          {empty
            ? "No current serial states found"
            : "Select a warehouse and served owner"}
        </p>
        <p className="mt-1 text-sm text-slate-600">{text}</p>
      </div>
    </div>
  );
}

function SerialStateRows({
  rows,
  timezone,
  onSelect,
}: {
  rows: SerialState[];
  timezone: string;
  onSelect: (id: string) => void;
}) {
  return (
    <>
      <div className="hidden overflow-x-auto md:block">
        <table className="min-w-full text-left text-sm">
          <thead className="bg-slate-50 text-xs font-bold text-slate-500 uppercase">
            <tr>
              {[
                "Serial",
                "Item",
                "Location / status",
                "Lot / HU",
                "Updated",
                "Action",
              ].map((heading) => (
                <th key={heading} className="px-5 py-3">
                  {heading}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((state) => (
              <tr
                key={state.serial_id}
                className="border-t border-slate-100 hover:bg-slate-50"
              >
                <td className="px-5 py-4 font-semibold text-slate-950">
                  {state.serial_no}
                </td>
                <td className="px-5 py-4">
                  <p className="font-semibold">{state.balance.item_code}</p>
                  <p className="mt-1 text-xs text-slate-500">
                    {state.balance.item_name}
                  </p>
                </td>
                <td className="px-5 py-4">
                  <p>{state.balance.location_code}</p>
                  <StatusBadge tone="info">
                    {state.balance.inventory_status_code}
                  </StatusBadge>
                </td>
                <td className="px-5 py-4">
                  {state.balance.lot_number ??
                    (state.balance.handling_unit_id ? "Handling unit" : "—")}
                </td>
                <td className="px-5 py-4">
                  {updatedAt(state.updated_at, timezone)}
                </td>
                <td className="px-5 py-4">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onSelect(state.serial_id)}
                  >
                    <Eye className="size-4" /> View
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="divide-y divide-slate-100 md:hidden">
        {rows.map((state) => (
          <article key={state.serial_id} className="p-4">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="font-semibold text-slate-950">
                  {state.serial_no}
                </p>
                <p className="mt-1 text-xs text-slate-500">
                  {state.balance.item_code} · {state.balance.item_name}
                </p>
              </div>
              <StatusBadge tone="info">
                {state.balance.inventory_status_code}
              </StatusBadge>
            </div>
            <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-xs text-slate-500">Location</dt>
                <dd className="mt-1">{state.balance.location_code}</dd>
              </div>
              <div>
                <dt className="text-xs text-slate-500">Lot / HU</dt>
                <dd className="mt-1">
                  {state.balance.lot_number ??
                    (state.balance.handling_unit_id ? "Handling unit" : "—")}
                </dd>
              </div>
            </dl>
            <Button
              className="mt-4 w-full"
              variant="secondary"
              onClick={() => onSelect(state.serial_id)}
            >
              <Eye className="size-4" /> View serial state
            </Button>
          </article>
        ))}
      </div>
    </>
  );
}
