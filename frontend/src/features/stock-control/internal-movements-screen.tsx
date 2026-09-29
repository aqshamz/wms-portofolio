"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import Decimal from "decimal.js";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import Link from "next/link";
import {
  ArrowRight,
  ArrowRightLeft,
  ChevronLeft,
  ChevronRight,
  LoaderCircle,
  RefreshCw,
  Search,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import { balanceKeys, listBalances } from "@/features/inventory/inventory-api";
import type {
  Balance,
  BalanceFilters,
} from "@/features/inventory/inventory-types";
import {
  listWarehouseOwners,
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import { InternalMoveDialog } from "./internal-move-dialog";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

export function InternalMovementsScreen({ timezone }: { timezone: string }) {
  const [filters, setFilters] = useQueryStates({
    warehouse: parseAsString.withDefault(""),
    owner: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    balance: parseAsString.withDefault(""),
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
  }, [filters.owner, owner, owners, setFilters, warehouse]);
  useEffect(() => {
    if (warehouseOwners.isSuccess && filters.owner && !owner) {
      void setFilters({ owner: "", balance: "", page: 1 });
    }
  }, [filters.owner, owner, setFilters, warehouseOwners.isSuccess]);

  const request: BalanceFilters = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    search: filters.search,
    includeZero: false,
    locationTypeCode: "STORAGE",
    inventoryStatusCode: "AVAILABLE",
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const balances = useQuery({
    queryKey: balanceKeys.list(request),
    queryFn: () => listBalances(request),
    enabled: Boolean(owner && warehouse),
  });
  const error = warehouses.error ?? warehouseOwners.error ?? balances.error;
  const page = balances.data?.page ?? request.page;
  const totalPages = balances.data?.total_pages ?? 0;

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-600">Stock control</p>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <ArrowRightLeft className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Internal movements
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Relocate available stock between locations in the same
                warehouse.
              </p>
            </div>
          </div>
        </div>
        <Button asChild variant="secondary">
          <Link href="/inventory/movements">View movement history</Link>
        </Button>
      </header>

      <section className="rounded-2xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950 sm:p-5">
        Each confirmation posts immediately and creates an immutable inventory
        movement. Reserved stock cannot move. Serialized stock moves one serial
        at a time; an eligible handling unit moves as a whole. Replenishment
        will reuse this posting foundation but remain a separately assigned task
        workflow.
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
            ariaLabel="Filter internal movements by warehouse"
            value={filters.warehouse}
            options={(warehouses.data?.items ?? []).map((row) => ({
              value: row.warehouse_id,
              label: `${row.name} (${row.code})`,
            }))}
            placeholder={
              warehouses.isPending ? "Loading warehouses…" : "Select warehouse"
            }
            onValueChange={(warehouseId) =>
              void setFilters({
                warehouse: warehouseId,
                owner: "",
                balance: "",
                page: 1,
              })
            }
          />
          <Select
            ariaLabel="Filter internal movements by owner"
            value={filters.owner}
            options={owners.map((row) => ({
              value: row.owner_id,
              label: `${row.owner_name} (${row.owner_code})`,
            }))}
            placeholder="Select served owner"
            disabled={!warehouse || warehouseOwners.isPending}
            onValueChange={(ownerId) =>
              void setFilters({ owner: ownerId, balance: "", page: 1 })
            }
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search movable inventory"
              placeholder="Item, source location or lot"
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
            <p className="font-semibold text-slate-950">Movable inventory</p>
            <p className="text-xs text-slate-500">
              {owner && warehouse
                ? `${balances.data?.total_items ?? 0} available balances in STORAGE locations`
                : "Select your working scope"}
            </p>
          </div>
          <Button
            variant="ghost"
            size="sm"
            disabled={balances.isFetching}
            onClick={() => {
              void warehouses.refetch();
              if (warehouse) void warehouseOwners.refetch();
              if (owner && warehouse) void balances.refetch();
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
          <EmptyState text="Only stock within your account’s owner and warehouse access is requested." />
        ) : balances.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" /> Loading inventory…
          </div>
        ) : !balances.data?.items.length ? (
          <EmptyState text="No available stock in STORAGE locations matches this scope and search." />
        ) : (
          <BalanceRows
            rows={balances.data.items}
            onMove={(balanceId) => void setFilters({ balance: balanceId })}
          />
        )}

        {owner &&
        warehouse &&
        balances.data &&
        balances.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {balances.data.total_items} balances · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || balances.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" /> Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || balances.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>

      {filters.balance ? (
        <InternalMoveDialog
          key={filters.balance}
          balanceId={filters.balance}
          timezone={timezone}
          onOpenChange={(open) => {
            if (!open) void setFilters({ balance: "" });
          }}
        />
      ) : null}
    </div>
  );
}

function EmptyState({ text }: { text: string }) {
  return (
    <div className="grid min-h-56 place-items-center p-6 text-center">
      <div>
        <ArrowRightLeft className="mx-auto size-8 text-slate-400" />
        <p className="mt-3 font-semibold">No source stock selected</p>
        <p className="mt-1 text-sm text-slate-600">{text}</p>
      </div>
    </div>
  );
}

function BalanceRows({
  rows,
  onMove,
}: {
  rows: Balance[];
  onMove: (id: string) => void;
}) {
  return (
    <>
      <div className="hidden overflow-x-auto md:block">
        <table className="min-w-full text-left text-sm">
          <thead className="bg-slate-50 text-xs font-bold text-slate-500 uppercase">
            <tr>
              {[
                "Item",
                "Source / status",
                "Lot / handling unit",
                "On hand / reserved / available",
                "Action",
              ].map((heading) => (
                <th key={heading} className="px-5 py-3">
                  {heading}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((balance) => {
              const available = new Decimal(balance.available_qty).gt(0);
              return (
                <tr
                  key={balance.balance_id}
                  className="border-t border-slate-100 hover:bg-slate-50"
                >
                  <td className="px-5 py-4">
                    <p className="font-semibold text-slate-950">
                      {balance.item_code}
                    </p>
                    <p className="mt-1 text-xs text-slate-500">
                      {balance.item_name}
                    </p>
                    {balance.serial_controlled ? (
                      <StatusBadge tone="info">Serialized</StatusBadge>
                    ) : null}
                  </td>
                  <td className="px-5 py-4">
                    <p>{balance.location_code}</p>
                    <p className="mt-1 text-xs text-slate-500">
                      {balance.inventory_status_code}
                    </p>
                  </td>
                  <td className="px-5 py-4">
                    {balance.lot_number ?? balance.handling_unit_barcode ?? "—"}
                  </td>
                  <td className="px-5 py-4">
                    {balance.on_hand_qty} / {balance.reserved_qty} /{" "}
                    {balance.available_qty} {balance.uom_code}
                  </td>
                  <td className="px-5 py-4">
                    <Button
                      size="sm"
                      disabled={!available}
                      onClick={() => onMove(balance.balance_id)}
                    >
                      Move <ArrowRight className="size-4" />
                    </Button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      <div className="divide-y divide-slate-100 md:hidden">
        {rows.map((balance) => {
          const available = new Decimal(balance.available_qty).gt(0);
          return (
            <article key={balance.balance_id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-semibold text-slate-950">
                    {balance.item_code}
                  </p>
                  <p className="mt-1 text-xs text-slate-500">
                    {balance.item_name}
                  </p>
                </div>
                <StatusBadge tone="info">
                  {balance.inventory_status_code}
                </StatusBadge>
              </div>
              <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
                <div>
                  <dt className="text-xs text-slate-500">Source</dt>
                  <dd>{balance.location_code}</dd>
                </div>
                <div>
                  <dt className="text-xs text-slate-500">Lot / HU</dt>
                  <dd>
                    {balance.lot_number ?? balance.handling_unit_barcode ?? "—"}
                  </dd>
                </div>
                <div className="col-span-2">
                  <dt className="text-xs text-slate-500">Available</dt>
                  <dd>
                    {balance.available_qty} {balance.uom_code}
                  </dd>
                </div>
              </dl>
              <Button
                className="mt-4 w-full"
                disabled={!available}
                onClick={() => onMove(balance.balance_id)}
              >
                Move stock <ArrowRight className="size-4" />
              </Button>
            </article>
          );
        })}
      </div>
    </>
  );
}
