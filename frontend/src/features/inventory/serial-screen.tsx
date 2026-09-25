"use client";

import { useEffect, useMemo, useState } from "react";
import { useQueries, useQuery } from "@tanstack/react-query";
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
import {
  listWarehouseOwners,
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import {
  getItem,
  itemCatalogKeys,
} from "@/features/item-catalog/item-catalog-api";
import { listSerials, serialKeys } from "./inventory-api";
import { SerialDetailDialog } from "./serial-detail-dialog";
import type { InventorySerial, SerialFilters } from "./inventory-types";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

function registeredAt(value: string, timezone: string) {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: timezone,
  }).format(new Date(value));
}

export function InventorySerialsScreen({ timezone }: { timezone: string }) {
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

  const request: SerialFilters = {
    ownerId: filters.owner,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const serials = useQuery({
    queryKey: serialKeys.list(request),
    queryFn: () => listSerials(request),
    enabled: Boolean(owner && warehouse),
  });
  const itemIds = [
    ...new Set((serials.data?.items ?? []).map((serial) => serial.item_id)),
  ];
  const itemQueries = useQueries({
    queries: itemIds.map((itemId) => ({
      queryKey: itemCatalogKeys.item(itemId),
      queryFn: () => getItem(itemId),
    })),
  });
  const itemNames = new Map(
    itemQueries.flatMap((query) =>
      query.data
        ? [[query.data.item_id, `${query.data.code} · ${query.data.name}`]]
        : [],
    ),
  );
  const error = warehouses.error ?? warehouseOwners.error ?? serials.error;
  const page = serials.data?.page ?? request.page;
  const totalPages = serials.data?.total_pages ?? 0;

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
              Serial numbers
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Owner-wide identities for individually tracked items.
            </p>
          </div>
        </div>
      </header>

      <section className="rounded-2xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950 sm:p-5">
        A serial number identifies one physical unit throughout its lifetime.
        Use Serial states to see whether it is currently in stock and where it
        is located; shipped serials remain in this identity registry.
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
            ariaLabel="Select warehouse for serial owner access"
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
            ariaLabel="Filter serial numbers by owner"
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
              aria-label="Search serial numbers"
              placeholder="Serial number"
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
              ? `${serials.data?.total_items ?? 0} serial numbers for this owner`
              : "Select your working owner"}
          </p>
          <Button
            variant="ghost"
            size="sm"
            disabled={serials.isFetching}
            onClick={() => {
              void warehouses.refetch();
              if (warehouse) void warehouseOwners.refetch();
              if (owner && warehouse) void serials.refetch();
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
          <EmptyState text="Select a warehouse, then an owner you are allowed to serve." />
        ) : serials.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" /> Loading serial
            numbers…
          </div>
        ) : !serials.data?.items.length ? (
          <EmptyState text="Try another serial-number search." empty />
        ) : (
          <SerialRows
            rows={serials.data.items}
            itemNames={itemNames}
            timezone={timezone}
            onSelect={(record) => void setFilters({ record })}
          />
        )}

        {owner && warehouse && serials.data && serials.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {serials.data.total_items} serial numbers · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || serials.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" /> Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || serials.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>

      {filters.record ? (
        <SerialDetailDialog
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
            ? "No serial numbers found"
            : "Select a warehouse and served owner"}
        </p>
        <p className="mt-1 text-sm text-slate-600">{text}</p>
      </div>
    </div>
  );
}

function SerialRows({
  rows,
  itemNames,
  timezone,
  onSelect,
}: {
  rows: InventorySerial[];
  itemNames: Map<string, string>;
  timezone: string;
  onSelect: (id: string) => void;
}) {
  return (
    <>
      <div className="hidden overflow-x-auto md:block">
        <table className="min-w-full text-left text-sm">
          <thead className="bg-slate-50 text-xs font-bold text-slate-500 uppercase">
            <tr>
              {["Serial number", "Item", "Registered", "Action"].map(
                (heading) => (
                  <th key={heading} className="px-5 py-3">
                    {heading}
                  </th>
                ),
              )}
            </tr>
          </thead>
          <tbody>
            {rows.map((serial) => (
              <tr
                key={serial.serial_id}
                className="border-t border-slate-100 hover:bg-slate-50"
              >
                <td className="px-5 py-4 font-semibold text-slate-950">
                  {serial.serial_no}
                </td>
                <td className="px-5 py-4">
                  {itemNames.get(serial.item_id) ?? "Loading item…"}
                </td>
                <td className="px-5 py-4">
                  {registeredAt(serial.created_at, timezone)}
                </td>
                <td className="px-5 py-4">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onSelect(serial.serial_id)}
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
        {rows.map((serial) => (
          <article key={serial.serial_id} className="p-4">
            <p className="font-semibold text-slate-950">{serial.serial_no}</p>
            <p className="mt-1 text-xs text-slate-500">
              {itemNames.get(serial.item_id) ?? "Loading item…"}
            </p>
            <p className="mt-3 text-xs text-slate-500">Registered</p>
            <p className="mt-1 text-sm">
              {registeredAt(serial.created_at, timezone)}
            </p>
            <Button
              className="mt-4 w-full"
              variant="secondary"
              onClick={() => onSelect(serial.serial_id)}
            >
              <Eye className="size-4" /> View serial
            </Button>
          </article>
        ))}
      </div>
    </>
  );
}
