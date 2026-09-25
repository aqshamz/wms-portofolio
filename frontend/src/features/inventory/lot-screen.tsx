"use client";

import { useEffect, useMemo, useState } from "react";
import { useQueries, useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import {
  CalendarDays,
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
import { listLots, lotKeys } from "./inventory-api";
import { LotDetailDialog } from "./lot-detail-dialog";
import type { InventoryLot, LotFilters } from "./inventory-types";

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

export function InventoryLotsScreen({ timezone }: { timezone: string }) {
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

  const request: LotFilters = {
    ownerId: filters.owner,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const lots = useQuery({
    queryKey: lotKeys.list(request),
    queryFn: () => listLots(request),
    enabled: Boolean(owner && warehouse),
  });
  const itemIds = [
    ...new Set((lots.data?.items ?? []).map((lot) => lot.item_id)),
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
  const error = warehouses.error ?? warehouseOwners.error ?? lots.error;
  const page = lots.data?.page ?? request.page;
  const totalPages = lots.data?.total_pages ?? 0;

  return (
    <div className="space-y-6">
      <header>
        <p className="text-sm font-semibold text-slate-600">Inventory</p>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <CalendarDays className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Lots
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Owner-wide lot identities, manufacture dates and expiry dates.
            </p>
          </div>
        </div>
      </header>

      <section className="rounded-2xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950 sm:p-5">
        Lots identify batches but do not represent stock quantity or warehouse
        placement. Use Balances to see where a lot currently exists and how much
        is available.
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
            ariaLabel="Select warehouse for lot owner access"
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
            ariaLabel="Filter lots by owner"
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
              aria-label="Search lots"
              placeholder="Lot number"
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
              ? `${lots.data?.total_items ?? 0} lots for this owner`
              : "Select your working owner"}
          </p>
          <Button
            variant="ghost"
            size="sm"
            disabled={lots.isFetching}
            onClick={() => {
              void warehouses.refetch();
              if (warehouse) void warehouseOwners.refetch();
              if (owner && warehouse) void lots.refetch();
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
        ) : lots.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" /> Loading lots…
          </div>
        ) : !lots.data?.items.length ? (
          <EmptyState text="Try another lot number search." empty />
        ) : (
          <LotRows
            rows={lots.data.items}
            itemNames={itemNames}
            timezone={timezone}
            onSelect={(record) => void setFilters({ record })}
          />
        )}

        {owner && warehouse && lots.data && lots.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {lots.data.total_items} lots · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || lots.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" /> Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || lots.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>

      {filters.record ? (
        <LotDetailDialog
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
        <CalendarDays className="mx-auto size-8 text-slate-400" />
        <p className="mt-3 font-semibold">
          {empty ? "No lots found" : "Select a warehouse and served owner"}
        </p>
        <p className="mt-1 text-sm text-slate-600">{text}</p>
      </div>
    </div>
  );
}

function LotRows({
  rows,
  itemNames,
  timezone,
  onSelect,
}: {
  rows: InventoryLot[];
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
              {[
                "Lot",
                "Item",
                "Manufactured",
                "Expires",
                "Registered",
                "Action",
              ].map((heading) => (
                <th key={heading} className="px-5 py-3">
                  {heading}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((lot) => (
              <tr
                key={lot.lot_id}
                className="border-t border-slate-100 hover:bg-slate-50"
              >
                <td className="px-5 py-4 font-semibold text-slate-950">
                  {lot.lot_number}
                </td>
                <td className="px-5 py-4">
                  {itemNames.get(lot.item_id) ?? "Loading item…"}
                </td>
                <td className="px-5 py-4">{lot.manufacture_date ?? "—"}</td>
                <td className="px-5 py-4">{lot.expiry_date ?? "—"}</td>
                <td className="px-5 py-4">
                  {registeredAt(lot.created_at, timezone)}
                </td>
                <td className="px-5 py-4">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onSelect(lot.lot_id)}
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
        {rows.map((lot) => (
          <article key={lot.lot_id} className="p-4">
            <p className="font-semibold text-slate-950">{lot.lot_number}</p>
            <p className="mt-1 text-xs text-slate-500">
              {itemNames.get(lot.item_id) ?? "Loading item…"}
            </p>
            <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-xs text-slate-500">Manufactured</dt>
                <dd className="mt-1">{lot.manufacture_date ?? "—"}</dd>
              </div>
              <div>
                <dt className="text-xs text-slate-500">Expires</dt>
                <dd className="mt-1">{lot.expiry_date ?? "—"}</dd>
              </div>
            </dl>
            <Button
              className="mt-4 w-full"
              variant="secondary"
              onClick={() => onSelect(lot.lot_id)}
            >
              <Eye className="size-4" /> View lot
            </Button>
          </article>
        ))}
      </div>
    </>
  );
}
