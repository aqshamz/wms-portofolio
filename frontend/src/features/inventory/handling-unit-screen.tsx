"use client";

import { useEffect, useMemo, useState } from "react";
import { useQueries, useQuery, useQueryClient } from "@tanstack/react-query";
import { parseAsInteger, parseAsString, useQueryStates } from "nuqs";
import {
  Boxes,
  ChevronLeft,
  ChevronRight,
  Eye,
  LoaderCircle,
  Plus,
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
  getLocation,
  storageLayoutKeys,
} from "@/features/storage-layout/storage-layout-api";
import {
  getHandlingUnitType,
  unitsPackagingKeys,
} from "@/features/units-packaging/units-packaging-api";
import { handlingUnitKeys, listHandlingUnits } from "./inventory-api";
import { HandlingUnitCreateDialog } from "./handling-unit-create-dialog";
import { HandlingUnitDetailDialog } from "./handling-unit-detail-dialog";
import type { HandlingUnit, HandlingUnitFilters } from "./inventory-types";

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

export function InventoryHandlingUnitsScreen({
  timezone,
  canCreate = false,
}: {
  timezone: string;
  canCreate?: boolean;
}) {
  const queryClient = useQueryClient();
  const [creating, setCreating] = useState(false);
  const [filters, setFilters] = useQueryStates({
    owner: parseAsString.withDefault(""),
    warehouse: parseAsString.withDefault(""),
    status: parseAsString.withDefault("all"),
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

  const status: HandlingUnitFilters["status"] = ["open", "closed"].includes(
    filters.status,
  )
    ? (filters.status as HandlingUnitFilters["status"])
    : "all";
  const request: HandlingUnitFilters = {
    ownerId: filters.owner,
    warehouseId: filters.warehouse,
    status,
    search: filters.search,
    page: Math.max(1, filters.page),
    pageSize: 10,
  };
  const units = useQuery({
    queryKey: handlingUnitKeys.list(request),
    queryFn: () => listHandlingUnits(request),
    enabled: Boolean(owner && warehouse),
  });
  const typeIds = [
    ...new Set(
      (units.data?.items ?? []).map((unit) => unit.handling_unit_type_id),
    ),
  ];
  const typeQueries = useQueries({
    queries: typeIds.map((id) => ({
      queryKey: unitsPackagingKeys.handling(id),
      queryFn: () => getHandlingUnitType(id),
    })),
  });
  const typeNames = new Map(
    typeQueries.flatMap((query) =>
      query.data
        ? [
            [
              query.data.handling_unit_type_id,
              `${query.data.code} · ${query.data.name}`,
            ] as const,
          ]
        : [],
    ),
  );
  const locationIds = [
    ...new Set(
      (units.data?.items ?? [])
        .map((unit) => unit.current_location_id)
        .filter((id): id is string => Boolean(id)),
    ),
  ];
  const locationQueries = useQueries({
    queries: locationIds.map((id) => ({
      queryKey: storageLayoutKeys.location(id),
      queryFn: () => getLocation(id),
    })),
  });
  const locationNames = new Map(
    locationQueries.flatMap((query) =>
      query.data
        ? [
            [
              query.data.location_id,
              `${query.data.code}${query.data.zone_code ? ` · ${query.data.zone_code}` : ""}`,
            ] as const,
          ]
        : [],
    ),
  );
  const error = warehouses.error ?? warehouseOwners.error ?? units.error;
  const page = units.data?.page ?? request.page;
  const totalPages = units.data?.total_pages ?? 0;

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold text-slate-600">Inventory</p>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <Boxes className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Handling units
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Pallets, cartons, totes and other trackable containers.
              </p>
            </div>
          </div>
        </div>
        {canCreate ? (
          <Button
            disabled={!owner || !warehouse}
            onClick={() => setCreating(true)}
          >
            <Plus className="size-4" /> Create handling unit
          </Button>
        ) : null}
      </header>

      <section className="rounded-2xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950 sm:p-5">
        Handling units group stock under a scannable container. A child handling
        unit follows its parent hierarchy and location; closed units remain
        visible for traceability.
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
            ariaLabel="Filter handling units by warehouse"
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
            ariaLabel="Filter handling units by owner"
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
            ariaLabel="Filter handling units by status"
            value={status}
            options={[
              { value: "all", label: "All statuses" },
              { value: "open", label: "Open" },
              { value: "closed", label: "Closed" },
            ]}
            onValueChange={(value) =>
              void setFilters({ status: value, record: "", page: 1 })
            }
          />
          <div className="relative">
            <Search className="pointer-events-none absolute top-3.5 left-3 size-4 text-slate-400" />
            <Input
              className="mt-0 pl-9"
              aria-label="Search handling units"
              placeholder="Barcode or internal reference"
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
              ? `${units.data?.total_items ?? 0} handling units`
              : "Select your working scope"}
          </p>
          <Button
            variant="ghost"
            size="sm"
            disabled={units.isFetching}
            onClick={() => {
              void warehouses.refetch();
              if (warehouse) void warehouseOwners.refetch();
              if (owner && warehouse) void units.refetch();
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
        ) : units.isPending ? (
          <div className="flex min-h-56 items-center justify-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-5 animate-spin" /> Loading handling
            units…
          </div>
        ) : !units.data?.items.length ? (
          <EmptyState text="Try another barcode, reference, or status." empty />
        ) : (
          <HandlingUnitRows
            rows={units.data.items}
            typeNames={typeNames}
            locationNames={locationNames}
            timezone={timezone}
            onSelect={(record) => void setFilters({ record })}
          />
        )}

        {owner && warehouse && units.data && units.data.total_items > 0 ? (
          <footer className="flex flex-col gap-3 border-t border-slate-200 p-4 text-sm sm:flex-row sm:items-center sm:justify-between">
            <p className="text-slate-600">
              {units.data.total_items} handling units · page {page} of{" "}
              {Math.max(1, totalPages)}
            </p>
            <div className="flex gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1 || units.isFetching}
                onClick={() => void setFilters({ page: page - 1 })}
              >
                <ChevronLeft className="size-4" /> Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages || units.isFetching}
                onClick={() => void setFilters({ page: page + 1 })}
              >
                Next <ChevronRight className="size-4" />
              </Button>
            </div>
          </footer>
        ) : null}
      </Panel>

      {filters.record ? (
        <HandlingUnitDetailDialog
          key={filters.record}
          id={filters.record}
          timezone={timezone}
          onOpenChange={(open) => {
            if (!open) void setFilters({ record: "" });
          }}
        />
      ) : null}
      {creating && owner && warehouse ? (
        <HandlingUnitCreateDialog
          ownerId={owner.owner_id}
          warehouseId={warehouse.warehouse_id}
          onCreated={async (unit) => {
            await queryClient.invalidateQueries({
              queryKey: ["inventory", "handling-units", "list"],
            });
            void setFilters({ record: unit.handling_unit_id });
          }}
          onOpenChange={setCreating}
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
        <Boxes className="mx-auto size-8 text-slate-400" />
        <p className="mt-3 font-semibold">
          {empty
            ? "No handling units found"
            : "Select a warehouse and served owner"}
        </p>
        <p className="mt-1 text-sm text-slate-600">{text}</p>
      </div>
    </div>
  );
}

function Status({ closed }: { closed: boolean }) {
  return (
    <span
      className={
        closed
          ? "inline-flex rounded-full bg-slate-100 px-2 py-1 text-xs font-semibold text-slate-700"
          : "inline-flex rounded-full bg-emerald-50 px-2 py-1 text-xs font-semibold text-emerald-700"
      }
    >
      {closed ? "Closed" : "Open"}
    </span>
  );
}

function HandlingUnitRows({
  rows,
  typeNames,
  locationNames,
  timezone,
  onSelect,
}: {
  rows: HandlingUnit[];
  typeNames: Map<string, string>;
  locationNames: Map<string, string>;
  timezone: string;
  onSelect: (id: string) => void;
}) {
  const location = (unit: HandlingUnit) =>
    unit.current_location_id
      ? (locationNames.get(unit.current_location_id) ?? "Loading location…")
      : "Not placed";
  return (
    <>
      <div className="hidden overflow-x-auto md:block">
        <table className="min-w-full text-left text-sm">
          <thead className="bg-slate-50 text-xs font-bold text-slate-500 uppercase">
            <tr>
              {[
                "Barcode",
                "Type",
                "Location",
                "Status",
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
            {rows.map((unit) => (
              <tr
                key={unit.handling_unit_id}
                className="border-t border-slate-100 hover:bg-slate-50"
              >
                <td className="px-5 py-4 font-semibold text-slate-950">
                  {unit.barcode}
                </td>
                <td className="px-5 py-4">
                  {typeNames.get(unit.handling_unit_type_id) ?? "Loading type…"}
                </td>
                <td className="px-5 py-4">{location(unit)}</td>
                <td className="px-5 py-4">
                  <Status closed={unit.is_closed} />
                </td>
                <td className="px-5 py-4">
                  {registeredAt(unit.created_at, timezone)}
                </td>
                <td className="px-5 py-4">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onSelect(unit.handling_unit_id)}
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
        {rows.map((unit) => (
          <article key={unit.handling_unit_id} className="p-4">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="font-semibold text-slate-950">{unit.barcode}</p>
                <p className="mt-1 text-xs text-slate-500">
                  {typeNames.get(unit.handling_unit_type_id) ?? "Loading type…"}
                </p>
              </div>
              <Status closed={unit.is_closed} />
            </div>
            <p className="mt-3 text-xs text-slate-500">Current location</p>
            <p className="mt-1 text-sm">{location(unit)}</p>
            <Button
              className="mt-4 w-full"
              variant="secondary"
              onClick={() => onSelect(unit.handling_unit_id)}
            >
              <Eye className="size-4" /> View handling unit
            </Button>
          </article>
        ))}
      </div>
    </>
  );
}
