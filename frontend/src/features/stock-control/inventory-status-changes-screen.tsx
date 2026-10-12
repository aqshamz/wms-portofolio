"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import Decimal from "decimal.js";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import Link from "next/link";
import {
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
import type { BalanceFilters } from "@/features/inventory/inventory-types";
import {
  listWarehouseOwners,
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import { InventoryStatusChangeDialog } from "./inventory-status-change-dialog";
import { manualStatusCodes } from "./inventory-status-change-validation";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

export function InventoryStatusChangesScreen({
  timezone,
}: {
  timezone: string;
}) {
  const [filters, setFilters] = useQueryStates({
    warehouse: parseAsString.withDefault(""),
    owner: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    status: parseAsString.withDefault("AVAILABLE"),
    locationType: parseAsString.withDefault("STORAGE"),
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
  const validStatus = manualStatusCodes.some((code) => code === filters.status);
  const validLocation = ["STORAGE", "PICK_FACE"].includes(filters.locationType);

  useEffect(() => {
    if (!filters.warehouse && warehouses.data?.items.length === 1)
      void setFilters({
        warehouse: warehouses.data.items[0].warehouse_id,
        page: 1,
      });
  }, [filters.warehouse, setFilters, warehouses.data?.items]);
  useEffect(() => {
    if (warehouse && !filters.owner && owners.length === 1)
      void setFilters({ owner: owners[0].owner_id, page: 1 });
  }, [filters.owner, owners, setFilters, warehouse]);
  useEffect(() => {
    if (warehouseOwners.isSuccess && filters.owner && !owner)
      void setFilters({ owner: "", balance: "", page: 1 });
  }, [filters.owner, owner, setFilters, warehouseOwners.isSuccess]);

  const request: BalanceFilters = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    search: filters.search,
    includeZero: false,
    inventoryStatusCode: validStatus ? filters.status : "AVAILABLE",
    locationTypeCode: validLocation ? filters.locationType : "STORAGE",
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const balances = useQuery({
    queryKey: balanceKeys.list(request),
    queryFn: () => listBalances(request),
    enabled: Boolean(owner && warehouse && validStatus && validLocation),
  });
  const error = warehouses.error ?? warehouseOwners.error ?? balances.error;
  const page = balances.data?.page ?? request.page;
  const totalPages = balances.data?.total_pages ?? 0;
  const historyQuery = new URLSearchParams();
  if (warehouse) historyQuery.set("warehouse", warehouse.warehouse_id);
  if (owner) historyQuery.set("owner", owner.owner_id);

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
                Inventory status changes
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Place stock on hold, release it, or mark it damaged or expired.
              </p>
            </div>
          </div>
        </div>
        <Button asChild variant="secondary">
          <Link href={`/inventory/movements?${historyQuery}`}>
            View movement history
          </Link>
        </Button>
      </header>
      <section className="rounded-2xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950 sm:p-5">
        Changes post immediately after review and confirmation, with a required
        reason and an immutable movement record. Total stock and location stay
        unchanged. Only unreserved stock in STORAGE or PICK_FACE can change
        between AVAILABLE, HOLD, DAMAGED and EXPIRED. Receiving, quality,
        quarantine and return transactions remain separate.
      </section>
      <Panel className="overflow-hidden">
        <form
          className="grid gap-3 border-b border-slate-200 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3"
          onSubmit={(event) => {
            event.preventDefault();
            void setFilters({
              search: searchDraft.trim(),
              page: 1,
              balance: "",
            });
          }}
        >
          <Select
            ariaLabel="Filter status changes by warehouse"
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
                balance: "",
                page: 1,
              })
            }
          />
          <Select
            ariaLabel="Filter status changes by owner"
            value={filters.owner}
            options={owners.map((row) => ({
              value: row.owner_id,
              label: `${row.owner_name} (${row.owner_code})`,
            }))}
            placeholder="Select served owner"
            disabled={!warehouse || warehouseOwners.isPending}
            onValueChange={(value) =>
              void setFilters({ owner: value, balance: "", page: 1 })
            }
          />
          <Select
            ariaLabel="Filter current inventory status"
            value={filters.status}
            options={manualStatusCodes.map((code) => ({
              value: code,
              label: code,
            }))}
            onValueChange={(value) =>
              void setFilters({ status: value, balance: "", page: 1 })
            }
          />
          <Select
            ariaLabel="Filter inventory location type"
            value={filters.locationType}
            options={[
              { value: "STORAGE", label: "Storage locations" },
              { value: "PICK_FACE", label: "Pick-face locations" },
            ]}
            onValueChange={(value) =>
              void setFilters({ locationType: value, balance: "", page: 1 })
            }
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search inventory for status change"
              placeholder="Item, location or lot"
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
            <p className="font-semibold">Source inventory</p>
            <p className="text-xs text-slate-500">
              {owner && warehouse
                ? `${balances.data?.total_items ?? 0} balances in the selected scope`
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
              if (owner && warehouse && validStatus && validLocation)
                void balances.refetch();
            }}
          >
            <RefreshCw className="size-4" />
            Refresh
          </Button>
        </div>
        {error ? (
          <p
            role="alert"
            className="m-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {error.message}
          </p>
        ) : !validStatus || !validLocation ? (
          <p role="alert" className="p-6 text-sm">
            Choose a supported inventory status and location type above.
          </p>
        ) : !owner || !warehouse ? (
          <p className="p-8 text-center text-sm text-slate-600">
            Select an accessible warehouse and its served owner. Only stock
            within your access scopes is requested.
          </p>
        ) : balances.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm">
            <LoaderCircle className="size-5 animate-spin" />
            Loading inventory…
          </div>
        ) : !balances.data?.items.length ? (
          <p className="p-8 text-center text-sm text-slate-600">
            No positive stock matches this scope, status and location filter.
          </p>
        ) : (
          <div className="divide-y divide-slate-100">
            {balances.data.items.map((row) => {
              const canChange =
                new Decimal(row.available_qty).gt(0) && !row.location_is_locked;
              return (
                <article
                  key={row.balance_id}
                  className="grid items-center gap-4 p-4 text-sm sm:p-5 md:grid-cols-[1.5fr_1fr_1fr_1.2fr_auto]"
                >
                  <div>
                    <p className="font-semibold">{row.item_code}</p>
                    <p className="text-xs text-slate-500">{row.item_name}</p>
                    {row.serial_controlled ? (
                      <StatusBadge tone="info">Serialized</StatusBadge>
                    ) : null}
                  </div>
                  <div>
                    <p className="text-xs text-slate-500">Location / status</p>
                    <p>{row.location_code}</p>
                    <StatusBadge>{row.inventory_status_code}</StatusBadge>
                  </div>
                  <div>
                    <p className="text-xs text-slate-500">
                      Lot / handling unit
                    </p>
                    <p>{row.lot_number ?? "No lot"}</p>
                    <p>{row.handling_unit_barcode ?? "Loose stock"}</p>
                  </div>
                  <div>
                    <p className="text-xs text-slate-500">
                      On hand / reserved / unreserved
                    </p>
                    <p>
                      {row.on_hand_qty} / {row.reserved_qty} /{" "}
                      {row.available_qty} {row.uom_code}
                    </p>
                    {row.location_is_locked ? (
                      <p className="text-xs text-amber-800">Location locked</p>
                    ) : !canChange ? (
                      <p className="text-xs text-amber-800">
                        No unreserved quantity
                      </p>
                    ) : null}
                  </div>
                  <Button
                    size="sm"
                    disabled={!canChange}
                    onClick={() => void setFilters({ balance: row.balance_id })}
                  >
                    Change status
                  </Button>
                </article>
              );
            })}
          </div>
        )}
        {owner &&
        warehouse &&
        balances.data &&
        balances.data.total_items > 0 ? (
          <footer className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 p-4 text-sm">
            <p>
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
                <ChevronLeft className="size-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || balances.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>
      {filters.balance && owner && warehouse ? (
        <InventoryStatusChangeDialog
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
