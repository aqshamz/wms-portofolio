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
  Eye,
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
import { WarehouseTransferDialog } from "./warehouse-transfer-dialog";
import { WarehouseTransferDetailDialog } from "./warehouse-transfer-detail-dialog";
import { listWarehouseTransfers, stockControlKeys } from "./stock-control-api";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

export function WarehouseTransfersScreen({ timezone }: { timezone: string }) {
  const [filters, setFilters] = useQueryStates({
    warehouse: parseAsString.withDefault(""),
    owner: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
    balance: parseAsString.withDefault(""),
    transfer: parseAsString.withDefault(""),
  });
  const [searchDraft, setSearchDraft] = useState(filters.search);
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(activeWarehouses),
    queryFn: () => listWarehouses(activeWarehouses),
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
  const transferRequest = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    side: "SOURCE" as const,
    status: "",
    search: "",
    page: 1,
    pageSize: 20,
  };
  const transfers = useQuery({
    queryKey: stockControlKeys.warehouseTransferList(transferRequest),
    queryFn: () => listWarehouseTransfers(transferRequest),
    enabled: Boolean(owner && warehouse),
  });
  const error =
    warehouses.error ??
    warehouseOwners.error ??
    balances.error ??
    transfers.error;
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
                Warehouse transfers
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Create, approve and dispatch inventory to another served
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
        Draft transfers do not change stock. Dispatch posts stock out and marks
        the document in transit; the target warehouse receives and puts it away
        separately.
      </section>

      {owner && warehouse ? (
        <Panel className="overflow-hidden">
          <div className="flex items-center justify-between border-b border-slate-200 px-5 py-4">
            <div>
              <h2 className="font-bold">Transfer documents</h2>
              <p className="text-xs text-slate-500">
                Source-side approval and dispatch
              </p>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => void transfers.refetch()}
            >
              <RefreshCw className="size-4" /> Refresh
            </Button>
          </div>
          {transfers.isPending ? (
            <div className="flex min-h-28 items-center justify-center gap-2 text-sm">
              <LoaderCircle className="size-4 animate-spin" /> Loading
              transfers…
            </div>
          ) : !transfers.data?.items.length ? (
            <p className="p-5 text-sm text-slate-500">
              No transfer documents yet. Choose a source balance below to create
              one.
            </p>
          ) : (
            <div className="divide-y divide-slate-100">
              {transfers.data.items.map((entry) => (
                <button
                  key={entry.warehouse_transfer_id}
                  type="button"
                  className="flex w-full items-center justify-between gap-4 p-4 text-left hover:bg-slate-50"
                  onClick={() =>
                    void setFilters({ transfer: entry.warehouse_transfer_id })
                  }
                >
                  <div>
                    <p className="font-semibold">
                      {entry.warehouse_transfer_id}
                    </p>
                    <p className="mt-1 text-xs text-slate-500">
                      {entry.target_warehouse_name} ·{" "}
                      {entry.created_at.slice(0, 10)}
                    </p>
                  </div>
                  <div className="flex items-center gap-3">
                    <StatusBadge
                      tone={
                        entry.status_code === "CANCELLED"
                          ? "danger"
                          : entry.status_code === "RECEIVED"
                            ? "success"
                            : "info"
                      }
                    >
                      {entry.status_code}
                    </StatusBadge>
                    <Eye className="size-4 text-slate-500" />
                  </div>
                </button>
              ))}
            </div>
          )}
        </Panel>
      ) : null}

      <Panel className="overflow-hidden">
        <form
          className="grid gap-3 border-b border-slate-200 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-[1fr_1fr_2fr_auto]"
          onSubmit={(event) => {
            event.preventDefault();
            void setFilters({ search: searchDraft.trim(), page: 1 });
          }}
        >
          <Select
            ariaLabel="Warehouse transfer source warehouse"
            value={filters.warehouse}
            options={(warehouses.data?.items ?? []).map((entry) => ({
              value: entry.warehouse_id,
              label: `${entry.name} (${entry.code})`,
            }))}
            placeholder={
              warehouses.isPending ? "Loading warehouses…" : "Source warehouse"
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
            ariaLabel="Warehouse transfer inventory owner"
            value={filters.owner}
            options={owners.map((entry) => ({
              value: entry.owner_id,
              label: `${entry.owner_name} (${entry.owner_code})`,
            }))}
            placeholder="Select inventory owner"
            disabled={!warehouse || warehouseOwners.isPending}
            onValueChange={(ownerId) =>
              void setFilters({ owner: ownerId, balance: "", page: 1 })
            }
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search transferable inventory"
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
            <p className="font-semibold text-slate-950">
              Transferable source stock
            </p>
            <p className="text-xs text-slate-500">
              {owner && warehouse
                ? `${balances.data?.total_items ?? 0} AVAILABLE balances in STORAGE locations`
                : "Select the source warehouse and owner"}
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
          <p
            role="alert"
            className="m-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {error.message}
          </p>
        ) : !owner || !warehouse ? (
          <EmptyState text="Only inventory within your account’s access scope is shown." />
        ) : balances.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" /> Loading inventory…
          </div>
        ) : !balances.data?.items.length ? (
          <EmptyState text="No AVAILABLE stock in STORAGE locations matches this scope." />
        ) : (
          <TransferRows
            rows={balances.data.items}
            onTransfer={(balanceId) => void setFilters({ balance: balanceId })}
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
        <WarehouseTransferDialog
          key={filters.balance}
          balanceId={filters.balance}
          timezone={timezone}
          onOpenChange={(open) => {
            if (!open) void setFilters({ balance: "" });
          }}
        />
      ) : null}
      {filters.transfer ? (
        <WarehouseTransferDetailDialog
          id={filters.transfer}
          timezone={timezone}
          canTransfer
          canReceive={false}
          canPutaway={false}
          onOpenChange={(open) => {
            if (!open) void setFilters({ transfer: "" });
          }}
        />
      ) : null}
    </div>
  );
}

function TransferRows({
  rows,
  onTransfer,
}: {
  rows: Balance[];
  onTransfer: (id: string) => void;
}) {
  return (
    <>
      <div className="hidden overflow-x-auto md:block">
        <table className="min-w-full text-left text-sm">
          <thead className="bg-slate-50 text-xs font-bold text-slate-500 uppercase">
            <tr>
              {[
                "Item",
                "Source",
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
              const transferable =
                new Decimal(balance.available_qty).gt(0) &&
                !balance.handling_unit_id;
              return (
                <tr
                  key={balance.balance_id}
                  className="border-t border-slate-100 hover:bg-slate-50"
                >
                  <td className="px-5 py-4">
                    <p className="font-semibold">{balance.item_code}</p>
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
                    {balance.handling_unit_id ? (
                      <p className="mt-1 text-xs text-amber-700">
                        HU transfer unavailable
                      </p>
                    ) : null}
                  </td>
                  <td className="px-5 py-4">
                    {balance.on_hand_qty} / {balance.reserved_qty} /{" "}
                    {balance.available_qty} {balance.uom_code}
                  </td>
                  <td className="px-5 py-4">
                    <Button
                      size="sm"
                      disabled={!transferable}
                      onClick={() => onTransfer(balance.balance_id)}
                    >
                      Transfer <ArrowRight className="size-4" />
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
          const transferable =
            new Decimal(balance.available_qty).gt(0) &&
            !balance.handling_unit_id;
          return (
            <article key={balance.balance_id} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-semibold">{balance.item_code}</p>
                  <p className="mt-1 text-xs text-slate-500">
                    {balance.item_name}
                  </p>
                </div>
                <StatusBadge tone="info">
                  {balance.inventory_status_code}
                </StatusBadge>
              </div>
              <p className="mt-3 text-sm">
                {balance.location_code} ·{" "}
                {balance.lot_number ??
                  balance.handling_unit_barcode ??
                  "No lot"}
              </p>
              <p className="mt-1 text-xs text-slate-500">
                Available {balance.available_qty} {balance.uom_code}
              </p>
              {balance.handling_unit_id ? (
                <p className="mt-2 text-xs text-amber-700">
                  Handling-unit transfer is unavailable in the immediate flow.
                </p>
              ) : null}
              <Button
                className="mt-4 w-full"
                disabled={!transferable}
                onClick={() => onTransfer(balance.balance_id)}
              >
                Transfer <ArrowRight className="size-4" />
              </Button>
            </article>
          );
        })}
      </div>
    </>
  );
}

function EmptyState({ text }: { text: string }) {
  return (
    <div className="grid min-h-56 place-items-center p-6 text-center">
      <div>
        <ArrowRightLeft className="mx-auto size-8 text-slate-400" />
        <p className="mt-3 font-semibold">No transferable stock</p>
        <p className="mt-1 text-sm text-slate-600">{text}</p>
      </div>
    </div>
  );
}
