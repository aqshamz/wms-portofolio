"use client";

import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import {
  ArrowRightLeft,
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
import { StatusBadge } from "@/components/ui/status-badge";
import {
  listWarehouseOwners,
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import {
  listMovements,
  listMovementTypes,
  movementKeys,
} from "./inventory-api";
import { InventoryMovementDetailDialog } from "./movement-detail-dialog";
import type { Movement, MovementFilters } from "./inventory-types";

const activeWarehouses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

function occurredAt(value: string, timezone: string) {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: timezone,
  }).format(new Date(value));
}

export function InventoryMovementsScreen({ timezone }: { timezone: string }) {
  const [filters, setFilters] = useQueryStates({
    owner: parseAsString.withDefault(""),
    warehouse: parseAsString.withDefault(""),
    movementType: parseAsString.withDefault("all"),
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
  const movementTypes = useQuery({
    queryKey: movementKeys.types,
    queryFn: listMovementTypes,
  });

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

  const request: MovementFilters = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    movementTypeId: filters.movementType === "all" ? "" : filters.movementType,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const movements = useQuery({
    queryKey: movementKeys.list(request),
    queryFn: () => listMovements(request),
    enabled: Boolean(owner && warehouse),
  });
  const error =
    warehouses.error ??
    warehouseOwners.error ??
    movementTypes.error ??
    movements.error;
  const page = movements.data?.page ?? request.page;
  const totalPages = movements.data?.total_pages ?? 0;

  return (
    <div className="space-y-6">
      <header>
        <p className="text-sm font-semibold text-slate-600">Inventory</p>
        <div className="mt-3 flex items-center gap-3">
          <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <ArrowRightLeft className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Inventory movements
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Immutable stock history created by warehouse workflows.
            </p>
          </div>
        </div>
      </header>

      <section className="rounded-2xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950 sm:p-5">
        Movement records cannot be edited or deleted. Each row explains when,
        where and why inventory changed.
      </section>

      <Panel className="overflow-hidden">
        <form
          className="grid gap-3 border-b border-slate-200 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-[1fr_1fr_1fr_2fr_auto]"
          onSubmit={(event) => {
            event.preventDefault();
            void setFilters({ search: searchDraft.trim(), page: 1 });
          }}
        >
          <Select
            ariaLabel="Filter movements by warehouse"
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
            ariaLabel="Filter movements by owner"
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
          <Select
            ariaLabel="Filter by movement type"
            value={filters.movementType}
            options={[
              { value: "all", label: "All movement types" },
              ...(movementTypes.data ?? []).map((value) => ({
                value: value.movement_type_id,
                label: `${value.name} (${value.code})`,
              })),
            ]}
            onValueChange={(movementType) =>
              void setFilters({ movementType, page: 1 })
            }
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search inventory movements"
              placeholder="Item, source document or lot"
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
              ? `${movements.data?.total_items ?? 0} movements`
              : "Select your working scope"}
          </p>
          <Button
            variant="ghost"
            size="sm"
            disabled={movements.isFetching}
            onClick={() => {
              void warehouses.refetch();
              void movementTypes.refetch();
              if (warehouse) void warehouseOwners.refetch();
              if (owner && warehouse) void movements.refetch();
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
          <EmptyState text="Only movements within your access scope will be requested." />
        ) : movements.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" /> Loading movements…
          </div>
        ) : !movements.data?.items.length ? (
          <EmptyState text="Try another movement type or search value." empty />
        ) : (
          <MovementRows
            rows={movements.data.items}
            timezone={timezone}
            onSelect={(record) => void setFilters({ record })}
          />
        )}

        {owner &&
        warehouse &&
        movements.data &&
        movements.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {movements.data.total_items} movements · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || movements.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" /> Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || movements.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>

      {filters.record ? (
        <InventoryMovementDetailDialog
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
        <ArrowRightLeft className="mx-auto size-8 text-slate-400" />
        <p className="mt-3 font-semibold">
          {empty
            ? "No inventory movements found"
            : "Select a warehouse and served owner"}
        </p>
        <p className="mt-1 text-sm text-slate-600">{text}</p>
      </div>
    </div>
  );
}

function MovementRows({
  rows,
  timezone,
  onSelect,
}: {
  rows: Movement[];
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
                "Movement / occurred",
                "Item / quantity",
                "From → to",
                "Source",
                "Action",
              ].map((heading) => (
                <th key={heading} className="px-5 py-3">
                  {heading}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((movement) => (
              <tr
                key={movement.movement_id}
                className="border-t border-slate-100 hover:bg-slate-50"
              >
                <td className="px-5 py-4">
                  <StatusBadge tone="info">
                    {movement.movement_type_code}
                  </StatusBadge>
                  <p className="mt-2 text-xs text-slate-500">
                    {occurredAt(movement.occurred_at, timezone)}
                  </p>
                </td>
                <td className="px-5 py-4">
                  <p className="font-semibold">{movement.item_code}</p>
                  <p className="mt-1 text-xs text-slate-500">
                    {movement.quantity} {movement.uom_code}
                  </p>
                </td>
                <td className="px-5 py-4">
                  {movement.from_location_code ?? "Outside WMS"} →{" "}
                  {movement.to_location_code ?? "Outside WMS"}
                </td>
                <td className="px-5 py-4">{movement.source_document_id}</td>
                <td className="px-5 py-4">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onSelect(movement.movement_id)}
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
        {rows.map((movement) => (
          <article key={movement.movement_id} className="p-4">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="font-semibold text-slate-950">
                  {movement.item_code}
                </p>
                <p className="mt-1 text-xs text-slate-500">
                  {occurredAt(movement.occurred_at, timezone)}
                </p>
              </div>
              <StatusBadge tone="info">
                {movement.movement_type_code}
              </StatusBadge>
            </div>
            <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-xs text-slate-500">Quantity</dt>
                <dd className="mt-1">
                  {movement.quantity} {movement.uom_code}
                </dd>
              </div>
              <div>
                <dt className="text-xs text-slate-500">Source</dt>
                <dd className="mt-1 break-words">
                  {movement.source_document_id}
                </dd>
              </div>
              <div className="col-span-2">
                <dt className="text-xs text-slate-500">From → to</dt>
                <dd className="mt-1">
                  {movement.from_location_code ?? "Outside WMS"} →{" "}
                  {movement.to_location_code ?? "Outside WMS"}
                </dd>
              </div>
            </dl>
            <Button
              className="mt-4 w-full"
              variant="secondary"
              onClick={() => onSelect(movement.movement_id)}
            >
              <Eye className="size-4" /> View movement
            </Button>
          </article>
        ))}
      </div>
    </>
  );
}
