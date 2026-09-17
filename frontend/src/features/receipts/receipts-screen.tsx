"use client";

import { useEffect, useMemo, useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import {
  ArrowRight,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  Eye,
  LoaderCircle,
  PackageCheck,
  Plus,
  Search,
} from "lucide-react";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import { ReceiptDetailDialog } from "@/features/receipts/receipt-detail-dialog";
import { ReceiptFormDialog } from "@/features/receipts/receipt-form-dialog";
import { listReceipts, receiptKeys } from "@/features/receipts/receipt-api";
import type { ReceiptStatus } from "@/features/receipts/receipt-types";
import {
  listWarehouseOwners,
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";

const PAGE_SIZE = 10;
const allActive = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
const statuses = [
  { value: "all", label: "All statuses" },
  { value: "OPEN", label: "Open" },
  { value: "COMPLETED", label: "Completed" },
  { value: "CANCELLED", label: "Cancelled" },
  { value: "REVERSED", label: "Reversed" },
] as const;

function statusTone(status: ReceiptStatus) {
  if (status === "COMPLETED") return "success";
  if (status === "OPEN") return "info";
  if (status === "CANCELLED") return "danger";
  return "neutral";
}

function label(value: string) {
  return value
    .replaceAll("_", " ")
    .toLowerCase()
    .replace(/^./, (first) => first.toUpperCase());
}

export function ReceiptsScreen({
  canReceive,
  canCancel,
  canReadInventory,
}: {
  canReceive: boolean;
  canCancel: boolean;
  canReadInventory: boolean;
}) {
  const [filters, setFilters] = useQueryStates({
    owner: parseAsString.withDefault(""),
    warehouse: parseAsString.withDefault(""),
    status: parseAsString.withDefault("all"),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
  });
  const [searchDraft, setSearchDraft] = useState(filters.search);
  const [creating, setCreating] = useState(false);
  const [selectedId, setSelectedId] = useState<string>();
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(allActive),
    queryFn: () => listWarehouses(allActive),
  });
  const warehouseOwners = useQuery({
    queryKey: warehouseKeys.owners(filters.warehouse || "none"),
    queryFn: () => listWarehouseOwners(filters.warehouse),
    enabled: Boolean(filters.warehouse),
  });
  const owners = useMemo(
    () => (warehouseOwners.data ?? []).filter((owner) => owner.is_active),
    [warehouseOwners.data],
  );
  const owner = owners.find((item) => item.owner_id === filters.owner);
  const warehouse = warehouses.data?.items.find(
    (item) => item.warehouse_id === filters.warehouse,
  );

  useEffect(() => {
    if (!filters.warehouse && warehouses.data?.items.length === 1) {
      void setFilters({
        warehouse: warehouses.data.items[0].warehouse_id,
        page: 1,
      });
    }
  }, [filters.warehouse, setFilters, warehouses.data?.items]);
  useEffect(() => {
    if (!filters.owner && owners.length === 1) {
      void setFilters({ owner: owners[0].owner_id, page: 1 });
    }
  }, [filters.owner, owners, setFilters]);
  useEffect(() => {
    if (warehouseOwners.isSuccess && filters.owner && !owner) {
      void setFilters({ owner: "", page: 1 });
    }
  }, [filters.owner, owner, setFilters, warehouseOwners.isSuccess]);
  const queryFilters = useMemo(
    () => ({
      ownerId: filters.owner,
      warehouseId: filters.warehouse,
      status: filters.status === "all" ? "" : filters.status,
      search: filters.search,
      page: Math.max(1, filters.page),
      pageSize: PAGE_SIZE,
    }),
    [filters],
  );
  const receipts = useQuery({
    queryKey: receiptKeys.list(queryFilters),
    queryFn: () => listReceipts(queryFilters),
    enabled: Boolean(owner && warehouse),
    placeholderData: keepPreviousData,
  });
  const currentPage = receipts.data?.page ?? queryFilters.page;
  const totalPages = receipts.data?.total_pages ?? 0;
  const scopeLabel =
    owner && warehouse ? `${owner.owner_name} · ${warehouse.name}` : "";

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-600">
            Inbound operations
          </p>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <PackageCheck className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Receipts
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Capture arrivals and post accepted inventory into quality
                control.
              </p>
            </div>
          </div>
        </div>
        {canReceive ? (
          <Button
            className="w-full sm:w-auto"
            disabled={!owner || !warehouse}
            onClick={() => setCreating(true)}
          >
            <Plus className="size-4" /> Open receipt
          </Button>
        ) : null}
      </header>

      <div className="rounded-2xl border border-slate-200 bg-white p-4 sm:p-5">
        <div className="flex flex-wrap items-center gap-2 text-xs font-semibold text-slate-700">
          <span className="rounded-full bg-cyan-50 px-3 py-1.5 text-cyan-800">
            Open draft
          </span>
          <ArrowRight className="size-4 text-slate-400" />
          <span className="rounded-full bg-emerald-50 px-3 py-1.5 text-emerald-800">
            Completed · QC pending
          </span>
          <span className="text-slate-400">or</span>
          <span className="rounded-full bg-rose-50 px-3 py-1.5 text-rose-800">
            Cancelled
          </span>
          <span className="text-slate-400">/</span>
          <span className="rounded-full bg-slate-100 px-3 py-1.5 text-slate-700">
            Reversed
          </span>
        </div>
        <p className="mt-3 text-sm text-slate-600">
          Completing posts each accepted batch to inventory. A completed receipt
          can only be reversed while stock is untouched and no quality
          inspection has started.
        </p>
        {!canReceive && !canCancel ? (
          <p className="mt-3 rounded-xl bg-amber-50 px-3 py-2 text-sm text-amber-900">
            This account is view-only. INBOUND.RECEIVE opens, edits, and
            completes receipts; INBOUND.CANCEL cancels or reverses them.
          </p>
        ) : null}
      </div>

      <Panel className="overflow-hidden">
        <div className="border-b border-slate-200 p-4 sm:p-5">
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-[1fr_1fr_1fr_2fr_auto]">
            <Select
              ariaLabel="Filter by warehouse"
              value={filters.warehouse}
              options={(warehouses.data?.items ?? []).map((item) => ({
                value: item.warehouse_id,
                label: `${item.name} (${item.code})`,
              }))}
              placeholder={
                warehouses.isPending
                  ? "Loading warehouses…"
                  : "Select warehouse"
              }
              onValueChange={(warehouseId) =>
                void setFilters({ warehouse: warehouseId, owner: "", page: 1 })
              }
            />
            <Select
              ariaLabel="Filter by owner"
              value={filters.owner}
              options={owners.map((item) => ({
                value: item.owner_id,
                label: `${item.owner_name} (${item.owner_code})`,
              }))}
              placeholder={
                !filters.warehouse
                  ? "Select warehouse first"
                  : warehouseOwners.isPending
                    ? "Loading served owners…"
                    : owners.length
                      ? "Select owner"
                      : "No active served owners"
              }
              disabled={!filters.warehouse || warehouseOwners.isPending}
              onValueChange={(ownerId) =>
                void setFilters({ owner: ownerId, page: 1 })
              }
            />
            <Select
              ariaLabel="Filter by receipt status"
              value={filters.status}
              options={statuses}
              onValueChange={(status) => void setFilters({ status, page: 1 })}
            />
            <Input
              aria-label="Search receipts"
              className="mt-0"
              value={searchDraft}
              placeholder="Search receipt, delivery note, or reference"
              onChange={(event) => setSearchDraft(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter")
                  void setFilters({ search: searchDraft, page: 1 });
              }}
            />
            <Button
              variant="secondary"
              onClick={() => void setFilters({ search: searchDraft, page: 1 })}
            >
              <Search className="size-4" /> Search
            </Button>
          </div>
        </div>
        {!owner || !warehouse ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <CircleAlert className="mx-auto size-8 text-amber-600" />
              <p className="mt-3 font-semibold text-slate-900">
                Select a warehouse and served owner
              </p>
              <p className="mt-1 text-sm text-slate-600">
                Receipts are filtered to the account’s configured access scope.
              </p>
            </div>
          </div>
        ) : receipts.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" /> Loading receipts…
          </div>
        ) : receipts.error ? (
          <p
            role="alert"
            className="m-5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
          >
            {receipts.error.message}
          </p>
        ) : !receipts.data?.items.length ? (
          <div className="grid min-h-56 place-items-center p-6 text-center">
            <div>
              <PackageCheck className="mx-auto size-8 text-slate-400" />
              <p className="mt-3 font-semibold text-slate-900">
                No receipts found
              </p>
              <p className="mt-1 text-sm text-slate-600">
                Release an Inbound Order, then open its receipt here.
              </p>
            </div>
          </div>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="min-w-full text-left text-sm">
                <thead className="bg-slate-50 text-xs font-bold tracking-wide text-slate-500 uppercase">
                  <tr>
                    <th className="px-5 py-3">Receipt</th>
                    <th className="px-5 py-3">Inbound Order</th>
                    <th className="px-5 py-3">Received</th>
                    <th className="px-5 py-3">Delivery / vehicle</th>
                    <th className="px-5 py-3">Status</th>
                    <th className="px-5 py-3 text-right">Action</th>
                  </tr>
                </thead>
                <tbody>
                  {receipts.data.items.map((receipt) => (
                    <tr
                      key={receipt.receipt_id}
                      className="border-t border-slate-100 hover:bg-slate-50/70"
                    >
                      <td className="px-5 py-4">
                        <p className="font-semibold text-slate-950">
                          {receipt.receipt_id}
                        </p>
                        <p className="mt-0.5 text-xs text-slate-500">
                          {receipt.owner_code} · {receipt.warehouse_code}
                        </p>
                      </td>
                      <td className="px-5 py-4 font-mono text-xs text-slate-700">
                        {receipt.inbound_id || "—"}
                      </td>
                      <td className="px-5 py-4">
                        <p>
                          {new Date(receipt.received_at).toLocaleString(
                            "en-ID",
                          )}
                        </p>
                        <p className="text-xs text-slate-500">
                          Business date {receipt.business_date}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <p>{receipt.delivery_note_no || "No delivery note"}</p>
                        <p className="text-xs text-slate-500">
                          {receipt.vehicle_number || "No vehicle"}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <StatusBadge tone={statusTone(receipt.status_code)}>
                          {label(receipt.status_code)}
                        </StatusBadge>
                      </td>
                      <td className="px-5 py-4 text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setSelectedId(receipt.receipt_id)}
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
              {receipts.data.items.map((receipt) => (
                <article key={receipt.receipt_id} className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="font-semibold break-all text-slate-950">
                        {receipt.receipt_id}
                      </p>
                      <p className="mt-1 font-mono text-xs text-slate-500">
                        {receipt.inbound_id}
                      </p>
                    </div>
                    <StatusBadge tone={statusTone(receipt.status_code)}>
                      {label(receipt.status_code)}
                    </StatusBadge>
                  </div>
                  <div className="mt-3 grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <p className="text-xs text-slate-500">Received</p>
                      <p>
                        {new Date(receipt.received_at).toLocaleDateString(
                          "en-ID",
                        )}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-slate-500">Delivery note</p>
                      <p>{receipt.delivery_note_no || "—"}</p>
                    </div>
                  </div>
                  <Button
                    className="mt-4 w-full"
                    variant="secondary"
                    onClick={() => setSelectedId(receipt.receipt_id)}
                  >
                    <Eye className="size-4" /> View receipt
                  </Button>
                </article>
              ))}
            </div>
          </>
        )}
        {receipts.data && receipts.data.total_items > 0 ? (
          <div className="flex flex-col gap-3 border-t border-slate-200 px-4 py-3 text-sm sm:flex-row sm:items-center sm:justify-between sm:px-5">
            <p className="text-slate-600">
              {receipts.data.total_items} receipt
              {receipts.data.total_items === 1 ? "" : "s"} · page {currentPage}{" "}
              of {Math.max(totalPages, 1)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={currentPage <= 1}
                onClick={() => void setFilters({ page: currentPage - 1 })}
              >
                <ChevronLeft className="size-4" /> Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={currentPage >= totalPages}
                onClick={() => void setFilters({ page: currentPage + 1 })}
              >
                Next <ChevronRight className="size-4" />
              </Button>
            </div>
          </div>
        ) : null}
      </Panel>
      {creating && owner && warehouse ? (
        <ReceiptFormDialog
          ownerId={owner.owner_id}
          warehouseId={warehouse.warehouse_id}
          scopeLabel={scopeLabel}
          onOpenChange={setCreating}
        />
      ) : null}
      {selectedId ? (
        <ReceiptDetailDialog
          receiptId={selectedId}
          canReceive={canReceive}
          canCancel={canCancel}
          canReadInventory={canReadInventory}
          onOpenChange={(open) => {
            if (!open) setSelectedId(undefined);
          }}
        />
      ) : null}
    </div>
  );
}
