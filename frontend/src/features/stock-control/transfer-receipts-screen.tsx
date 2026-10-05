"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsString, useQueryStates } from "nuqs";
import {
  ArrowDownToLine,
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
import { listWarehouseTransfers, stockControlKeys } from "./stock-control-api";
import { WarehouseTransferDetailDialog } from "./warehouse-transfer-detail-dialog";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
export function TransferReceiptsScreen({
  timezone,
  canReceive,
  canPutaway,
}: {
  timezone: string;
  canReceive: boolean;
  canPutaway: boolean;
}) {
  const [filters, setFilters] = useQueryStates({
    warehouse: parseAsString.withDefault(""),
    owner: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    status: parseAsString.withDefault("ALL"),
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
    if (!filters.warehouse && warehouses.data?.items.length === 1)
      void setFilters({ warehouse: warehouses.data.items[0].warehouse_id });
  }, [filters.warehouse, setFilters, warehouses.data?.items]);
  useEffect(() => {
    if (warehouse && !filters.owner && owners.length === 1)
      void setFilters({ owner: owners[0].owner_id });
  }, [filters.owner, owners, setFilters, warehouse]);
  const request = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    side: "TARGET" as const,
    status: filters.status === "ALL" ? "" : filters.status,
    search: filters.search,
    page: 1,
    pageSize: 100,
  };
  const transfers = useQuery({
    queryKey: stockControlKeys.warehouseTransferList(request),
    queryFn: () => listWarehouseTransfers(request),
    enabled: Boolean(owner && warehouse),
  });
  const error = warehouses.error ?? warehouseOwners.error ?? transfers.error;
  return (
    <div className="space-y-6">
      <header>
        <p className="text-sm font-semibold text-slate-600">
          Inbound operations
        </p>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <ArrowDownToLine className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold sm:text-3xl">
              Transfer receipts
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Receive dispatched inter-warehouse stock and put it away.
            </p>
          </div>
        </div>
      </header>
      <section className="rounded-2xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950">
        Only transfers addressed to the selected warehouse are shown. Receipt
        posts stock into PUTAWAY_PENDING at a receiving-capable location.
      </section>
      <Panel className="overflow-hidden">
        <form
          className="grid gap-3 border-b border-slate-200 p-4 md:grid-cols-2 xl:grid-cols-[1fr_1fr_1fr_2fr_auto]"
          onSubmit={(event) => {
            event.preventDefault();
            void setFilters({ search: searchDraft.trim() });
          }}
        >
          <Select
            ariaLabel="Transfer destination warehouse"
            value={filters.warehouse}
            options={(warehouses.data?.items ?? []).map((entry) => ({
              value: entry.warehouse_id,
              label: `${entry.name} (${entry.code})`,
            }))}
            placeholder="Destination warehouse"
            onValueChange={(value) =>
              void setFilters({ warehouse: value, owner: "" })
            }
          />
          <Select
            ariaLabel="Transfer inventory owner"
            value={filters.owner}
            options={owners.map((entry) => ({
              value: entry.owner_id,
              label: `${entry.owner_name} (${entry.owner_code})`,
            }))}
            placeholder="Inventory owner"
            disabled={!warehouse}
            onValueChange={(value) => void setFilters({ owner: value })}
          />
          <Select
            ariaLabel="Transfer status"
            value={filters.status}
            options={[
              { value: "ALL", label: "All statuses" },
              { value: "IN_TRANSIT", label: "In transit" },
              { value: "RECEIVED", label: "Received" },
              { value: "CANCELLED", label: "Cancelled" },
            ]}
            onValueChange={(value) => void setFilters({ status: value })}
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              value={searchDraft}
              placeholder="Transfer or item"
              onChange={(event) => setSearchDraft(event.target.value)}
            />
          </div>
          <Button type="submit" variant="secondary">
            Search
          </Button>
        </form>
        <div className="flex items-center justify-between border-b border-slate-200 px-5 py-3">
          <p className="text-sm font-semibold">Destination transfer queue</p>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => void transfers.refetch()}
          >
            <RefreshCw className="size-4" /> Refresh
          </Button>
        </div>
        {error ? (
          <p className="m-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
            {error.message}
          </p>
        ) : !owner || !warehouse ? (
          <p className="p-8 text-center text-sm text-slate-500">
            Select a destination warehouse and served owner.
          </p>
        ) : transfers.isPending ? (
          <div className="flex min-h-48 items-center justify-center gap-2 text-sm">
            <LoaderCircle className="size-5 animate-spin" /> Loading transfers…
          </div>
        ) : !transfers.data?.items.length ? (
          <p className="p-8 text-center text-sm text-slate-500">
            No destination transfers match these filters.
          </p>
        ) : (
          <div className="divide-y divide-slate-100">
            {transfers.data.items.map((entry) => (
              <button
                key={entry.warehouse_transfer_id}
                className="flex w-full items-center justify-between gap-4 p-5 text-left hover:bg-slate-50"
                onClick={() =>
                  void setFilters({ transfer: entry.warehouse_transfer_id })
                }
              >
                <div>
                  <p className="font-semibold">{entry.warehouse_transfer_id}</p>
                  <p className="mt-1 text-xs text-slate-500">
                    From {entry.source_warehouse_name} · {entry.owner_name}
                  </p>
                </div>
                <div className="flex items-center gap-3">
                  <StatusBadge
                    tone={entry.status_code === "RECEIVED" ? "success" : "info"}
                  >
                    {entry.status_code}
                  </StatusBadge>
                  <Eye className="size-4" />
                </div>
              </button>
            ))}
          </div>
        )}
      </Panel>
      {filters.transfer ? (
        <WarehouseTransferDetailDialog
          id={filters.transfer}
          timezone={timezone}
          canTransfer={false}
          canReceive={canReceive}
          canPutaway={canPutaway}
          onOpenChange={(open) => {
            if (!open) void setFilters({ transfer: "" });
          }}
        />
      ) : null}
    </div>
  );
}
