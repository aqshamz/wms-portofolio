"use client";

import { useEffect, useMemo, useState } from "react";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import {
  Archive,
  ArrowRight,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  Eye,
  LoaderCircle,
  Plus,
  Search,
  Settings2,
} from "lucide-react";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  listPurchaseOrders,
  purchaseOrderKeys,
} from "@/features/purchase-orders/purchase-order-api";
import { PurchaseOrderDetailDialog } from "@/features/purchase-orders/purchase-order-detail-dialog";
import { PurchaseOrderFormDialog } from "@/features/purchase-orders/purchase-order-form-dialog";
import type { PurchaseOrderStatus } from "@/features/purchase-orders/purchase-order-types";
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
const statusOptions = [
  { value: "all", label: "All statuses" },
  { value: "DRAFT", label: "Draft" },
  { value: "APPROVED", label: "Approved" },
  { value: "PARTIALLY_RECEIVED", label: "Partially received" },
  { value: "RECEIVED", label: "Received" },
  { value: "CLOSED", label: "Closed" },
  { value: "CANCELLED", label: "Cancelled" },
] as const;

const dateFormatter = new Intl.DateTimeFormat("en-ID", {
  day: "2-digit",
  month: "short",
  year: "numeric",
});

function statusTone(status: PurchaseOrderStatus) {
  if (status === "APPROVED" || status === "RECEIVED") return "success";
  if (status === "PARTIALLY_RECEIVED") return "warning";
  if (status === "CANCELLED") return "danger";
  if (status === "DRAFT") return "info";
  return "neutral";
}

function statusLabel(status: string) {
  return status
    .replaceAll("_", " ")
    .toLowerCase()
    .replace(/^./, (value) => value.toUpperCase());
}

export function PurchaseOrdersScreen({
  canPlan,
  canApprove,
  canCancel,
}: {
  canPlan: boolean;
  canApprove: boolean;
  canCancel: boolean;
}) {
  const [filters, setFilters] = useQueryStates({
    owner: parseAsString.withDefault(""),
    warehouse: parseAsString.withDefault(""),
    status: parseAsString.withDefault("all"),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
  });
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
    () => (warehouseOwners.data ?? []).filter((item) => item.is_active),
    [warehouseOwners.data],
  );
  const owner = owners.find((item) => item.owner_id === filters.owner);

  useEffect(() => {
    if (!filters.owner && owners.length === 1)
      void setFilters({
        owner: owners[0].owner_id,
        page: 1,
      });
  }, [filters.owner, owners, setFilters]);
  useEffect(() => {
    if (!filters.warehouse && warehouses.data?.items.length === 1)
      void setFilters({
        warehouse: warehouses.data.items[0].warehouse_id,
        page: 1,
      });
  }, [filters.warehouse, warehouses.data?.items, setFilters]);

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
  const orders = useQuery({
    queryKey: purchaseOrderKeys.list(queryFilters),
    queryFn: () => listPurchaseOrders(queryFilters),
    enabled: Boolean(owner && filters.warehouse),
    placeholderData: keepPreviousData,
  });
  const warehouse = warehouses.data?.items.find(
    (item) => item.warehouse_id === filters.warehouse,
  );
  const totalPages = orders.data?.total_pages ?? 0;
  const currentPage = orders.data?.page ?? queryFilters.page;
  const viewOnly = !canPlan && !canApprove && !canCancel;

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-600">
            Inbound operations
          </p>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <Archive className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Purchase orders
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Plan supplier commitments and track receiving progress.
              </p>
            </div>
          </div>
        </div>
        {canPlan ? (
          <Button
            className="w-full sm:w-auto"
            disabled={!owner || !warehouse}
            onClick={() => setCreating(true)}
          >
            <Plus className="size-4" />
            Create purchase order
          </Button>
        ) : null}
      </header>
      <div className="rounded-2xl border border-slate-200 bg-white p-4 sm:p-5">
        <div className="flex flex-wrap items-center gap-2 text-xs font-semibold text-slate-700">
          <span className="rounded-full bg-cyan-50 px-3 py-1.5 text-cyan-800">
            Draft
          </span>
          <ArrowRight className="size-4 text-slate-400" />
          <span className="rounded-full bg-emerald-50 px-3 py-1.5 text-emerald-800">
            Approved
          </span>
          <ArrowRight className="size-4 text-slate-400" />
          <span className="rounded-full bg-amber-50 px-3 py-1.5 text-amber-800">
            Partially received
          </span>
          <ArrowRight className="size-4 text-slate-400" />
          <span className="rounded-full bg-emerald-50 px-3 py-1.5 text-emerald-800">
            Received / closed
          </span>
          <span className="text-slate-400">or</span>
          <span className="rounded-full bg-rose-50 px-3 py-1.5 text-rose-800">
            Cancelled
          </span>
        </div>
        <p className="mt-3 text-sm text-slate-600">
          Drafts can be edited line by line. Purchase Orders are not hard
          deleted; cancellation preserves the audit trail before an inbound
          order is created.
        </p>
        {viewOnly ? (
          <p className="mt-3 rounded-xl bg-amber-50 px-3 py-2 text-sm text-amber-900">
            This account is view-only. Assign INBOUND.PLAN for create/edit,
            INBOUND.APPROVE for approval, or INBOUND.CANCEL for cancellation.
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
              onValueChange={(warehouse) =>
                void setFilters({ warehouse, owner: "", page: 1 })
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
              onValueChange={(owner) => void setFilters({ owner, page: 1 })}
            />
            <Select
              ariaLabel="Filter by status"
              value={filters.status}
              options={statusOptions}
              onValueChange={(status) => void setFilters({ status, page: 1 })}
            />
            <form
              className="flex gap-2"
              onSubmit={(event) => {
                event.preventDefault();
                const data = new FormData(event.currentTarget);
                void setFilters({
                  search: String(data.get("search") ?? "").trim(),
                  page: 1,
                });
              }}
            >
              <label className="relative flex-1">
                <span className="sr-only">Search purchase orders</span>
                <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
                <input
                  key={filters.search}
                  name="search"
                  defaultValue={filters.search}
                  placeholder="PO number or supplier"
                  className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-3 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                />
              </label>
              <Button type="submit" variant="secondary">
                Search
              </Button>
            </form>
          </div>
          {!filters.owner || !filters.warehouse ? (
            <p className="mt-3 text-sm text-amber-800">
              Select an owner and warehouse to load only the Purchase Orders
              inside your access scope.
            </p>
          ) : null}
        </div>
        <div className="flex min-h-12 items-center justify-between border-b border-slate-200 px-4 py-3 text-sm sm:px-5">
          <p className="font-semibold text-slate-700">
            {!filters.owner || !filters.warehouse
              ? "Scope required"
              : orders.isPending
                ? "Loading purchase orders…"
                : `${(orders.data?.total_items ?? 0).toLocaleString()} purchase order${orders.data?.total_items === 1 ? "" : "s"}`}
          </p>
          {orders.isFetching && !orders.isPending ? (
            <span className="inline-flex items-center gap-2 text-xs text-slate-500">
              <LoaderCircle className="size-3.5 animate-spin" />
              Refreshing
            </span>
          ) : null}
        </div>
        {!filters.owner || !filters.warehouse ? (
          <div className="px-5 py-16 text-center">
            <Archive className="mx-auto size-9 text-slate-300" />
            <h2 className="mt-4 font-bold text-slate-950">
              Choose an operational scope
            </h2>
            <p className="mt-1 text-sm text-slate-500">
              Both fields are required by the API to prevent cross-owner or
              cross-warehouse access.
            </p>
          </div>
        ) : orders.isPending ? (
          <div className="space-y-3 p-5">
            {Array.from({ length: 5 }, (_, index) => (
              <div
                key={index}
                className="h-16 animate-pulse rounded-xl bg-slate-100"
              />
            ))}
          </div>
        ) : orders.isError ? (
          <div className="p-5">
            <div
              role="alert"
              className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
            >
              <div className="flex gap-3">
                <CircleAlert className="mt-0.5 size-5" />
                <div>
                  <p className="font-semibold">
                    Purchase orders could not be loaded
                  </p>
                  <p className="mt-1 text-sm">{orders.error.message}</p>
                </div>
              </div>
              <Button
                variant="secondary"
                className="mt-4"
                onClick={() => orders.refetch()}
              >
                Try again
              </Button>
            </div>
          </div>
        ) : !orders.data.items.length ? (
          <div className="px-5 py-16 text-center">
            <Archive className="mx-auto size-9 text-slate-300" />
            <h2 className="mt-4 font-bold text-slate-950">
              No purchase orders found
            </h2>
            <p className="mt-1 text-sm text-slate-500">
              Create a draft or adjust the filters.
            </p>
          </div>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full border-collapse text-left">
                <thead>
                  <tr className="bg-slate-50 text-xs font-bold tracking-wide text-slate-500 uppercase">
                    <th className="px-5 py-3">Purchase order</th>
                    <th className="px-5 py-3">Supplier</th>
                    <th className="px-5 py-3">Ordered</th>
                    <th className="px-5 py-3">Expected arrival</th>
                    <th className="px-5 py-3">Status</th>
                    <th className="px-5 py-3 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {orders.data.items.map((order) => (
                    <tr
                      key={order.purchase_order_id}
                      className="hover:bg-slate-50/70"
                    >
                      <td className="px-5 py-4">
                        <p className="font-semibold text-slate-950">
                          {order.purchase_order_no}
                        </p>
                        <p className="mt-0.5 font-mono text-xs text-slate-500">
                          {order.owner_code} · {order.warehouse_code}
                        </p>
                      </td>
                      <td className="px-5 py-4">
                        <p className="text-sm font-medium text-slate-800">
                          {order.vendor_name}
                        </p>
                        <p className="font-mono text-xs text-slate-500">
                          {order.vendor_code}
                        </p>
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-600">
                        {dateFormatter.format(new Date(order.ordered_at))}
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-600">
                        {order.expected_arrival_at
                          ? dateFormatter.format(
                              new Date(order.expected_arrival_at),
                            )
                          : "Not set"}
                      </td>
                      <td className="px-5 py-4">
                        <StatusBadge tone={statusTone(order.status_code)}>
                          {statusLabel(order.status_code)}
                        </StatusBadge>
                      </td>
                      <td className="px-5 py-4 text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setSelectedId(order.purchase_order_id)}
                        >
                          {viewOnly ? (
                            <Eye className="size-4" />
                          ) : (
                            <Settings2 className="size-4" />
                          )}
                          {viewOnly ? "View" : "Manage"}
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <ul className="divide-y divide-slate-200 md:hidden">
              {orders.data.items.map((order) => (
                <li key={order.purchase_order_id} className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="font-semibold text-slate-950">
                        {order.purchase_order_no}
                      </p>
                      <p className="mt-0.5 text-sm text-slate-600">
                        {order.vendor_name}
                      </p>
                    </div>
                    <StatusBadge tone={statusTone(order.status_code)}>
                      {statusLabel(order.status_code)}
                    </StatusBadge>
                  </div>
                  <div className="mt-3 flex items-center justify-between border-t border-slate-100 pt-2">
                    <p className="text-xs text-slate-500">
                      Ordered {dateFormatter.format(new Date(order.ordered_at))}
                    </p>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setSelectedId(order.purchase_order_id)}
                    >
                      {viewOnly ? (
                        <Eye className="size-4" />
                      ) : (
                        <Settings2 className="size-4" />
                      )}
                      {viewOnly ? "View" : "Manage"}
                    </Button>
                  </div>
                </li>
              ))}
            </ul>
          </>
        )}
        {!orders.isPending && !orders.isError && totalPages > 0 ? (
          <div className="flex flex-col gap-3 border-t border-slate-200 px-4 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-5">
            <p className="text-sm text-slate-600">
              Page{" "}
              <span className="font-semibold text-slate-900">
                {currentPage}
              </span>{" "}
              of {totalPages}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={currentPage <= 1}
                onClick={() => void setFilters({ page: currentPage - 1 })}
              >
                <ChevronLeft className="size-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={currentPage >= totalPages}
                onClick={() => void setFilters({ page: currentPage + 1 })}
              >
                Next
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </div>
        ) : null}
      </Panel>
      {creating && owner && warehouse ? (
        <PurchaseOrderFormDialog
          ownerId={owner.owner_id}
          ownerLabel={`${owner.owner_name} (${owner.owner_code})`}
          warehouseId={warehouse.warehouse_id}
          warehouseLabel={`${warehouse.name} (${warehouse.code})`}
          onOpenChange={(open) => !open && setCreating(false)}
        />
      ) : null}
      {selectedId ? (
        <PurchaseOrderDetailDialog
          purchaseOrderId={selectedId}
          canPlan={canPlan}
          canApprove={canApprove}
          canCancel={canCancel}
          onOpenChange={(open) => !open && setSelectedId(undefined)}
        />
      ) : null}
    </div>
  );
}
