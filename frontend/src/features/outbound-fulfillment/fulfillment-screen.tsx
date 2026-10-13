"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import { PackageCheck, Plus, Search, Truck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  listWarehouseOwners,
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import {
  fulfillmentId,
  fulfillmentKeys,
  listFulfillment,
  type FulfillmentMode,
} from "./fulfillment-api";
import { FulfillmentCreateDialog } from "./fulfillment-create-dialog";
import { FulfillmentDetailDialog } from "./fulfillment-detail-dialog";

const active = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
export function FulfillmentScreen({
  mode,
  canWrite,
  canCancel,
  canCreateHU,
}: {
  mode: FulfillmentMode;
  canWrite: boolean;
  canCancel: boolean;
  canCreateHU: boolean;
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
  const [selected, setSelected] = useState<string>();
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
    () => (ownersQuery.data ?? []).filter((o) => o.is_active),
    [ownersQuery.data],
  );
  const warehouse = warehouses.data?.items.find(
    (w) => w.warehouse_id === filters.warehouse,
  );
  const owner = owners.find((o) => o.owner_id === filters.owner);
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
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    status: filters.status === "all" ? "" : filters.status,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const query = useQuery({
    queryKey: fulfillmentKeys.list(mode, request),
    queryFn: () => listFulfillment(mode, request),
    enabled: Boolean(owner && warehouse),
  });
  const packing = mode === "packings";
  const title = packing ? "Packing" : "Shipments / HU dispatch";
  const Icon = packing ? PackageCheck : Truck;
  const error = warehouses.error || ownersQuery.error || query.error;
  const changeScope = (next: { warehouse?: string; owner?: string }) => {
    setCreating(false);
    setSelected(undefined);
    void setFilters({ ...next, page: 1 });
  };
  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-sm font-semibold text-slate-600">
            Outbound operations
          </p>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <Icon className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                {title}
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                {packing
                  ? "Pack passed check lines into loose stock or scannable handling units."
                  : "Build a manifest and dispatch loose stock or an entire handling unit."}
              </p>
            </div>
          </div>
        </div>
        {canWrite && (
          <Button
            className="w-full sm:w-auto"
            disabled={!owner || !warehouse}
            onClick={() => setCreating(true)}
          >
            <Plus className="size-4" />
            {packing ? "Create packing" : "Create shipment"}
          </Button>
        )}
      </header>
      <p className="rounded-2xl border border-slate-200 bg-white p-4 text-sm text-slate-600 sm:p-5">
        {packing
          ? "Passed check → Packing → Completed packing. Packing moves checked stock into the packing area; it does not remove inventory from the warehouse."
          : "Completed packing → Planned shipment → Dispatch stock → Complete shipment. HU dispatch is all-or-nothing across every positive balance in the container. Carrier-service shipments require a primary driver."}
      </p>
      <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white">
        <form
          className="border-b border-slate-200 p-4 sm:p-5"
          onSubmit={(e) => {
            e.preventDefault();
            void setFilters({ search: search.trim(), page: 1 });
          }}
        >
          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-[1fr_1fr_1fr_2fr_auto]">
            <Select
              ariaLabel="Warehouse"
              value={filters.warehouse}
              options={(warehouses.data?.items ?? []).map((w) => ({
                value: w.warehouse_id,
                label: `${w.name} (${w.code})`,
              }))}
              placeholder="Select warehouse"
              onValueChange={(value) =>
                changeScope({ warehouse: value, owner: "" })
              }
            />
            <Select
              ariaLabel="Owner"
              value={filters.owner}
              options={owners.map((o) => ({
                value: o.owner_id,
                label: `${o.owner_name} (${o.owner_code})`,
              }))}
              placeholder={
                !warehouse ? "Select warehouse first" : "Select owner"
              }
              disabled={!warehouse || ownersQuery.isPending}
              onValueChange={(value) => changeScope({ owner: value })}
            />
            <Select
              ariaLabel="Status"
              value={filters.status}
              options={[
                { value: "all", label: "All statuses" },
                ...(packing
                  ? ["OPEN", "COMPLETED", "CANCELLED"]
                  : ["PLANNED", "SHIPPED", "CANCELLED"]
                ).map((value) => ({
                  value,
                  label: value.replaceAll("_", " "),
                })),
              ]}
              onValueChange={(status) => void setFilters({ status, page: 1 })}
            />
            <label className="relative">
              <span className="sr-only">Search {title}</span>
              <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
              <Input
                className="mt-0 pl-10"
                type="search"
                value={search}
                placeholder={
                  packing
                    ? "Packing or client DO"
                    : "Shipment, tracking or route"
                }
                onChange={(e) => setSearch(e.target.value)}
              />
            </label>
            <Button variant="secondary">Search</Button>
          </div>
        </form>
        {error && (
          <p role="alert" className="p-5 text-rose-700">
            {error.message}
          </p>
        )}
        {!owner || !warehouse ? (
          <div className="px-5 py-16 text-center">
            <Icon className="mx-auto size-9 text-slate-300" />
            <h2 className="mt-4 font-bold">Choose an operational scope</h2>
            <p className="mt-1 text-sm text-slate-500">
              Select an accessible warehouse and served owner.
            </p>
          </div>
        ) : query.isPending ? (
          <p className="p-5">Loading {title.toLowerCase()}…</p>
        ) : (
          !error && (
            <>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm">
                  <thead className="bg-slate-50 text-xs text-slate-500">
                    <tr>
                      {[
                        "DOCUMENT",
                        packing ? "CLIENT DO / LOCATION" : "DATE / VEHICLE",
                        "STATUS",
                        "POSTED LINES",
                        "ACTION",
                      ].map((label) => (
                        <th key={label} className="px-5 py-3 font-semibold">
                          {label}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {query.data?.items.map((doc) => (
                      <tr
                        key={fulfillmentId(doc)}
                        className="border-t border-slate-200"
                      >
                        <td className="px-5 py-4 font-semibold">
                          {fulfillmentId(doc)}
                        </td>
                        <td className="px-5 py-4">
                          {"packing_id" in doc ? (
                            <>
                              {doc.client_delivery_order_no}
                              <p className="mt-1 text-xs text-slate-500">
                                {doc.packing_location_code}
                              </p>
                            </>
                          ) : (
                            <>
                              {doc.business_date}
                              <p className="mt-1 text-xs text-slate-500">
                                {doc.vehicle_number || "No vehicle reference"}
                              </p>
                            </>
                          )}
                        </td>
                        <td className="px-5 py-4">
                          <StatusBadge
                            tone={
                              doc.status_code === "CANCELLED"
                                ? "danger"
                                : "info"
                            }
                          >
                            {doc.status_code}
                          </StatusBadge>
                        </td>
                        <td className="px-5 py-4">{doc.line_count}</td>
                        <td className="px-5 py-4">
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => setSelected(fulfillmentId(doc))}
                          >
                            View
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {!query.data?.items.length && (
                <p className="p-6 text-sm text-slate-500">
                  No documents match these filters.
                </p>
              )}
              <footer className="flex items-center justify-between gap-3 border-t border-slate-200 p-4 text-xs text-slate-500">
                <span>
                  {query.data?.total_items ?? 0} documents · page {request.page}{" "}
                  of {Math.max(1, query.data?.total_pages ?? 1)}
                </span>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="secondary"
                    disabled={request.page <= 1 || query.isFetching}
                    onClick={() => void setFilters({ page: request.page - 1 })}
                  >
                    Previous
                  </Button>
                  <Button
                    size="sm"
                    variant="secondary"
                    disabled={
                      request.page >= (query.data?.total_pages ?? 0) ||
                      query.isFetching
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
        <FulfillmentCreateDialog
          mode={mode}
          ownerId={owner.owner_id}
          warehouseId={warehouse.warehouse_id}
          scopeLabel={`${owner.owner_name} · ${warehouse.name}`}
          onOpenChange={(open) => {
            if (!open) setCreating(false);
          }}
          onCreated={(id) => {
            setCreating(false);
            setSelected(id);
          }}
        />
      )}
      {selected && (
        <FulfillmentDetailDialog
          key={`${mode}-${selected}`}
          mode={mode}
          id={selected}
          canWrite={canWrite}
          canCancel={canCancel}
          canCreateHU={canCreateHU}
          onOpenChange={(open) => {
            if (!open) setSelected(undefined);
          }}
        />
      )}
    </div>
  );
}
