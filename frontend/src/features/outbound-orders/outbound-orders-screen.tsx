"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import { ArrowRight, ClipboardList, Plus, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  listWarehouseOwners,
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import { listOutboundOrders, outboundOrderKeys } from "./outbound-order-api";
import { OutboundOrderDetailDialog } from "./outbound-order-detail-dialog";
import { OutboundOrderFormDialog } from "./outbound-order-form-dialog";

const active = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
const statuses = [
  "DRAFT",
  "VALIDATED",
  "RELEASED",
  "PARTIALLY_ALLOCATED",
  "ALLOCATED",
  "WAVED",
  "PICKING",
  "STAGED",
  "CHECKING",
  "CHECK_FAILED",
  "CHECKED",
  "PACKING",
  "PACKED",
  "SHIPPED",
  "DELIVERED",
  "DELIVERY_FAILED",
  "PARTIALLY_DELIVERED",
  "RETURNED",
  "CANCELLED",
];
export function OutboundOrdersScreen({
  canPlan,
  canCancel,
}: {
  canPlan: boolean;
  canCancel: boolean;
}) {
  const [filters, setFilters] = useQueryStates({
    warehouse: parseAsString.withDefault(""),
    owner: parseAsString.withDefault(""),
    status: parseAsString.withDefault("all"),
    search: parseAsString.withDefault(""),
    page: parseAsInteger.withDefault(1),
  });
  const [search, setSearch] = useState(filters.search);
  const [creating, setCreating] = useState(false);
  const [selectedId, setSelectedId] = useState<string>();
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(active),
    queryFn: () => listWarehouses(active),
  });
  const ownersQuery = useQuery({
    queryKey: warehouseKeys.owners(filters.warehouse || "none"),
    queryFn: () => listWarehouseOwners(filters.warehouse),
    enabled: Boolean(filters.warehouse),
  });
  const owners = useMemo(
    () => (ownersQuery.data ?? []).filter((owner) => owner.is_active),
    [ownersQuery.data],
  );
  const warehouse = warehouses.data?.items.find(
    (item) => item.warehouse_id === filters.warehouse,
  );
  const owner = owners.find((item) => item.owner_id === filters.owner);
  useEffect(() => {
    if (!filters.warehouse && warehouses.data?.items.length === 1)
      void setFilters({
        warehouse: warehouses.data.items[0].warehouse_id,
        page: 1,
      });
  }, [filters.warehouse, warehouses.data?.items, setFilters]);
  useEffect(() => {
    if (!filters.owner && owners.length === 1)
      void setFilters({ owner: owners[0].owner_id, page: 1 });
  }, [filters.owner, owners, setFilters]);
  useEffect(() => {
    if (ownersQuery.isSuccess && filters.owner && !owner)
      void setFilters({ owner: "", page: 1 });
  }, [filters.owner, owner, ownersQuery.isSuccess, setFilters]);
  const request = {
    warehouseId: filters.warehouse,
    ownerId: filters.owner,
    status: filters.status === "all" ? "" : filters.status,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const orders = useQuery({
    queryKey: outboundOrderKeys.list(request),
    queryFn: () => listOutboundOrders(request),
    enabled: Boolean(owner && warehouse),
  });
  const error = warehouses.error || ownersQuery.error || orders.error;
  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-600">
            Outbound operations
          </p>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <ClipboardList className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Delivery orders
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Client demand, frozen quantities, validation and release.
              </p>
            </div>
          </div>
        </div>
        {canPlan && (
          <Button
            className="w-full sm:w-auto"
            disabled={!owner || !warehouse}
            onClick={() => setCreating(true)}
          >
            <Plus className="size-4" />
            Create delivery order
          </Button>
        )}
      </header>
      <div className="rounded-2xl border border-slate-200 bg-white p-4 sm:p-5">
        <div className="flex flex-wrap items-center gap-2 text-xs font-medium">
          <span className="rounded-full bg-cyan-50 px-3 py-1 text-cyan-800">
            Draft
          </span>
          <ArrowRight className="size-4 text-slate-400" />
          <span className="rounded-full bg-emerald-50 px-3 py-1 text-emerald-800">
            Validated
          </span>
          <ArrowRight className="size-4 text-slate-400" />
          <span className="rounded-full bg-emerald-50 px-3 py-1 text-emerald-800">
            Released
          </span>
          <ArrowRight className="size-4 text-slate-400" />
          <span className="rounded-full bg-slate-100 px-3 py-1 text-slate-600">
            Allocation (next step)
          </span>
        </div>
        <p className="mt-3 text-sm text-slate-600">
          Order UOMs are frozen on each line. Warehouse work uses base units. No
          stock moves when creating or releasing a delivery order.
        </p>
      </div>
      <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
        <form
          className="border-b border-slate-200 p-4 sm:p-5"
          onSubmit={(event) => {
            event.preventDefault();
            void setFilters({ search: search.trim(), page: 1 });
          }}
        >
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-[1fr_1fr_1fr_2fr_auto]">
            <Select
              ariaLabel="Warehouse"
              value={filters.warehouse}
              options={(warehouses.data?.items ?? []).map((item) => ({
                value: item.warehouse_id,
                label: `${item.name} (${item.code})`,
              }))}
              placeholder="Select warehouse"
              onValueChange={(value) => {
                setSelectedId(undefined);
                setCreating(false);
                void setFilters({ warehouse: value, owner: "", page: 1 });
              }}
            />
            <Select
              ariaLabel="Owner"
              value={filters.owner}
              options={owners.map((item) => ({
                value: item.owner_id,
                label: `${item.owner_name} (${item.owner_code})`,
              }))}
              placeholder={
                !warehouse
                  ? "Select warehouse first"
                  : ownersQuery.isPending
                    ? "Loading served owners…"
                    : !owners.length
                      ? "No active served owners"
                      : "Select owner"
              }
              disabled={!warehouse || ownersQuery.isPending}
              onValueChange={(value) => {
                setSelectedId(undefined);
                setCreating(false);
                void setFilters({ owner: value, page: 1 });
              }}
            />
            <Select
              ariaLabel="Delivery Order status"
              value={filters.status}
              options={[
                { value: "all", label: "All statuses" },
                ...statuses.map((status) => ({
                  value: status,
                  label: status.replaceAll("_", " "),
                })),
              ]}
              onValueChange={(value) =>
                void setFilters({ status: value, page: 1 })
              }
            />
            <label className="relative">
              <span className="sr-only">Search delivery orders</span>
              <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
              <Input
                className="mt-0 pl-10"
                type="search"
                aria-label="Search delivery orders"
                placeholder="Client DO / order number"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
              />
            </label>
            <Button type="submit" variant="secondary">
              Search
            </Button>
          </div>
          {(!owner || !warehouse) && (
            <p className="mt-3 text-sm text-amber-800">
              Select an owner and warehouse to load only delivery orders inside
              your access scope.
            </p>
          )}
        </form>
        {error && (
          <p role="alert" className="p-4 text-rose-700">
            {error.message}
          </p>
        )}
        {!owner || !warehouse ? (
          <>
            <div className="min-h-12 border-b border-slate-200 px-4 py-3 text-sm sm:px-5">
              Scope required
            </div>
            <div className="px-5 py-16 text-center">
              <ClipboardList className="mx-auto size-9 text-slate-300" />
              <h2 className="mt-4 font-bold text-slate-950">
                Choose an operational scope
              </h2>
              <p className="mt-1 text-sm text-slate-500">
                Select an accessible warehouse and served owner to view delivery
                orders.
              </p>
            </div>
          </>
        ) : orders.isPending ? (
          <p className="p-6">Loading Delivery Orders…</p>
        ) : (
          !error && (
            <>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm">
                  <thead className="bg-slate-50 text-xs text-slate-500">
                    <tr>
                      {[
                        "CLIENT DO / DOCUMENT",
                        "CUSTOMER / SHIP TO",
                        "BUSINESS DATE",
                        "STATUS",
                        "ACTION",
                      ].map((label) => (
                        <th key={label} className="px-4 py-3 font-semibold">
                          {label}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {orders.data?.items.map((order) => (
                      <tr
                        key={order.outbound_id}
                        className="border-t border-slate-200"
                      >
                        <td className="px-4 py-4">
                          <p className="font-semibold">
                            {order.client_delivery_order_no}
                          </p>
                          <p className="mt-1 text-xs text-slate-500">
                            {order.outbound_id}
                          </p>
                        </td>
                        <td className="px-4 py-4">
                          {order.customer_name}
                          <p className="text-xs text-slate-500">
                            {order.ship_to_name}
                          </p>
                        </td>
                        <td className="px-4 py-4">{order.business_date}</td>
                        <td className="px-4 py-4">
                          <StatusBadge
                            tone={
                              order.status_code === "CANCELLED"
                                ? "danger"
                                : "info"
                            }
                          >
                            {order.status_code.replaceAll("_", " ")}
                          </StatusBadge>
                        </td>
                        <td className="px-4 py-4">
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => setSelectedId(order.outbound_id)}
                          >
                            View
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {!orders.data?.items.length && (
                <p className="p-6 text-sm text-slate-500">
                  No Delivery Orders match these filters.
                </p>
              )}
              <footer className="flex items-center justify-between gap-3 border-t border-slate-200 p-4 text-xs text-slate-500">
                <span>
                  {orders.data?.total_items ?? 0} orders · page{" "}
                  {orders.data?.page ?? 1} of{" "}
                  {Math.max(1, orders.data?.total_pages ?? 1)}
                </span>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="secondary"
                    disabled={request.page <= 1 || orders.isFetching}
                    onClick={() => void setFilters({ page: request.page - 1 })}
                  >
                    Previous
                  </Button>
                  <Button
                    size="sm"
                    variant="secondary"
                    disabled={
                      request.page >= (orders.data?.total_pages ?? 0) ||
                      orders.isFetching
                    }
                    onClick={() => void setFilters({ page: request.page + 1 })}
                  >
                    Next
                  </Button>
                </div>
              </footer>
            </>
          )
        )}
      </section>
      {creating && owner && warehouse && (
        <OutboundOrderFormDialog
          ownerId={owner.owner_id}
          warehouseId={warehouse.warehouse_id}
          ownerLabel={`${owner.owner_name} (${owner.owner_code})`}
          warehouseLabel={`${warehouse.name} (${warehouse.code})`}
          scopeLabel={`${owner.owner_name} · ${warehouse.name}`}
          onOpenChange={(open) => {
            if (!open) setCreating(false);
          }}
          onCreated={(id) => {
            setCreating(false);
            setSelectedId(id);
          }}
        />
      )}
      {selectedId && (
        <OutboundOrderDetailDialog
          id={selectedId}
          canPlan={canPlan}
          canCancel={canCancel}
          onOpenChange={(open) => {
            if (!open) setSelectedId(undefined);
          }}
        />
      )}
    </div>
  );
}
