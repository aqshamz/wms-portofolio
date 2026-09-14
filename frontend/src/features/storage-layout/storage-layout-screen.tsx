"use client";

import { useEffect, useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  AlertTriangle,
  Boxes,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  Layers3,
  LoaderCircle,
  LockKeyhole,
  MapPin,
  Pencil,
  Plus,
  Search,
  X,
} from "lucide-react";
import Link from "next/link";
import {
  parseAsInteger,
  parseAsString,
  parseAsStringLiteral,
  useQueryStates,
} from "nuqs";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Panel } from "@/components/ui/panel";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  deactivateLocation,
  deactivateZone,
  listLocations,
  listLocationTypes,
  listZones,
  storageLayoutKeys,
  updateLocationType,
} from "@/features/storage-layout/storage-layout-api";
import {
  LocationDialog,
  LocationTypeDialog,
  ZoneDialog,
} from "@/features/storage-layout/storage-layout-dialogs";
import type {
  LocationType,
  WarehouseLocation,
  WarehouseZone,
} from "@/features/storage-layout/storage-layout-types";
import {
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 10;
const views = ["locations", "zones", "types"] as const;
const activeValues = ["all", "active", "inactive"] as const;
const statusOptions = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;
const warehouseFilters = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

type Editor =
  | { kind: "type"; item?: LocationType }
  | { kind: "zone"; item?: WarehouseZone }
  | { kind: "location"; item?: WarehouseLocation }
  | null;
type DeactivateTarget =
  | { kind: "type"; item: LocationType }
  | { kind: "zone"; item: WarehouseZone }
  | { kind: "location"; item: WarehouseLocation };

const viewLabels = {
  locations: "Locations",
  zones: "Zones",
  types: "Location types",
} as const;

function EmptyState({
  icon: Icon,
  title,
  message,
}: {
  icon: typeof MapPin;
  title: string;
  message: string;
}) {
  return (
    <div className="px-5 py-14 text-center">
      <Icon className="mx-auto size-9 text-slate-300" />
      <h2 className="mt-4 font-bold text-slate-950">{title}</h2>
      <p className="mt-1 text-sm text-slate-500">{message}</p>
    </div>
  );
}

function ErrorState({ error, retry }: { error: Error; retry: () => void }) {
  return (
    <div className="p-5">
      <div
        role="alert"
        className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900"
      >
        <div className="flex gap-3">
          <CircleAlert className="mt-0.5 size-5 shrink-0" />
          <div>
            <p className="font-semibold">Storage layout could not be loaded</p>
            <p className="mt-1 text-sm">{error.message}</p>
          </div>
        </div>
        <Button variant="secondary" className="mt-4" onClick={retry}>
          Try again
        </Button>
      </div>
    </div>
  );
}

function LoadingState() {
  return (
    <div className="space-y-3 p-5">
      {Array.from({ length: 5 }, (_, index) => (
        <div
          key={index}
          className="h-16 animate-pulse rounded-xl bg-slate-100"
        />
      ))}
    </div>
  );
}

function DeactivateDialog({
  target,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: {
  target?: DeactivateTarget;
  pending: boolean;
  error?: Error | null;
  onConfirm: () => void;
  onOpenChange: (open: boolean) => void;
}) {
  const name = target
    ? target.kind === "location"
      ? target.item.code
      : target.item.name
    : "record";
  return (
    <Dialog.Root
      open={target !== undefined}
      onOpenChange={(open) => {
        if (!pending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content
          role="alertdialog"
          className="fixed top-1/2 left-1/2 z-50 w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-white p-5 shadow-2xl focus:outline-none sm:p-6"
        >
          <div className="flex items-start justify-between">
            <div className="grid size-11 place-items-center rounded-xl bg-amber-100 text-amber-700">
              <AlertTriangle className="size-5" />
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close confirmation"
                className="grid size-9 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-4" />
              </button>
            </Dialog.Close>
          </div>
          <Dialog.Title className="mt-5 text-lg font-bold text-slate-950">
            Deactivate {name}?
          </Dialog.Title>
          <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
            It remains available in historical records but cannot be selected
            for new warehouse operations.
          </Dialog.Description>
          {error ? (
            <p
              role="alert"
              className="mt-4 rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
            >
              {error.message}
            </p>
          ) : null}
          <div className="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <Dialog.Close asChild>
              <Button type="button" variant="secondary">
                Keep active
              </Button>
            </Dialog.Close>
            <Button
              type="button"
              disabled={pending}
              onClick={onConfirm}
              className="bg-rose-700 hover:bg-rose-800"
            >
              {pending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              {pending ? "Deactivating…" : "Deactivate"}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function StorageLayoutScreen({ canWrite }: { canWrite: boolean }) {
  const [filters, setFilters] = useQueryStates({
    view: parseAsStringLiteral(views).withDefault("locations"),
    warehouse: parseAsString.withDefault(""),
    search: parseAsString.withDefault(""),
    active: parseAsStringLiteral(activeValues).withDefault("all"),
    page: parseAsInteger.withDefault(1),
  });
  const [editor, setEditor] = useState<Editor>(null);
  const [deactivateTarget, setDeactivateTarget] = useState<DeactivateTarget>();
  const queryClient = useQueryClient();
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(warehouseFilters),
    queryFn: () => listWarehouses(warehouseFilters),
  });
  const locationTypes = useQuery({
    queryKey: storageLayoutKeys.locationTypes(filters.active),
    queryFn: () => listLocationTypes(filters.active),
  });
  const allLocationTypes = useQuery({
    queryKey: storageLayoutKeys.locationTypes("all"),
    queryFn: () => listLocationTypes("all"),
    enabled: filters.view === "locations",
  });
  const zones = useQuery({
    queryKey: storageLayoutKeys.zones(
      filters.warehouse,
      filters.active,
      filters.search,
    ),
    queryFn: () => listZones(filters.warehouse, filters.active, filters.search),
    enabled: Boolean(filters.warehouse) && filters.view === "zones",
  });
  const allZones = useQuery({
    queryKey: storageLayoutKeys.zones(filters.warehouse, "all"),
    queryFn: () => listZones(filters.warehouse, "all"),
    enabled: Boolean(filters.warehouse) && filters.view === "locations",
  });
  const locationFilters = {
    warehouseId: filters.warehouse,
    search: filters.search,
    active: filters.active,
    page: Math.max(1, filters.page),
    pageSize: PAGE_SIZE,
  };
  const locations = useQuery({
    queryKey: storageLayoutKeys.locations(locationFilters),
    queryFn: () => listLocations(locationFilters),
    enabled: Boolean(filters.warehouse) && filters.view === "locations",
    placeholderData: keepPreviousData,
  });
  const deactivate = useMutation({
    mutationFn: async (target: DeactivateTarget) => {
      if (target.kind === "zone") return deactivateZone(target.item.zone_id);
      if (target.kind === "location")
        return deactivateLocation(target.item.location_id);
      return updateLocationType(target.item.location_type_id, {
        name: target.item.name,
        description: target.item.description,
        allows_receiving: target.item.allows_receiving,
        allows_storage: target.item.allows_storage,
        allows_picking: target.item.allows_picking,
        allows_shipping: target.item.allows_shipping,
        is_active: false,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: storageLayoutKeys.all });
      toast.success("Storage layout record deactivated.");
      setDeactivateTarget(undefined);
    },
  });

  const warehouseItems = warehouses.data?.items ?? [];
  const firstWarehouseId = warehouseItems[0]?.warehouse_id;
  useEffect(() => {
    if (!filters.warehouse && firstWarehouseId) {
      void setFilters({ warehouse: firstWarehouseId, page: 1 });
    }
  }, [filters.warehouse, firstWarehouseId, setFilters]);

  const warehouseOptions = warehouseItems.map((warehouse) => ({
    value: warehouse.warehouse_id,
    label: `${warehouse.name} (${warehouse.code})`,
  }));
  const currentView = filters.view;
  const locationItems = locations.data?.items ?? [];
  const currentPage = locations.data?.page ?? locationFilters.page;
  const totalPages = locations.data?.total_pages ?? 0;
  const createDisabled =
    (currentView !== "types" && !filters.warehouse) ||
    (currentView === "locations" &&
      (!allZones.data?.some((zone) => zone.is_active) ||
        !allLocationTypes.data?.some((type) => type.is_active)));

  function startCreate() {
    setEditor(
      currentView === "types"
        ? { kind: "type" }
        : currentView === "zones"
          ? { kind: "zone" }
          : { kind: "location" },
    );
  }

  return (
    <div className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <Link
            href="/master-data/core"
            className="inline-flex min-h-9 items-center gap-2 text-sm font-semibold text-slate-600 hover:text-slate-950"
          >
            <ChevronLeft className="size-4" />
            Core masters
          </Link>
          <div className="mt-3 flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
              <Layers3 className="size-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                Storage layout
              </h1>
              <p className="mt-1 text-sm text-slate-600">
                Configure location capabilities and each warehouse&apos;s
                physical structure.
              </p>
            </div>
          </div>
        </div>
        {canWrite ? (
          <Button
            className="w-full sm:w-auto"
            disabled={createDisabled}
            onClick={startCreate}
          >
            <Plus className="size-4" />
            Create{" "}
            {currentView === "types"
              ? "location type"
              : currentView.slice(0, -1)}
          </Button>
        ) : null}
      </header>

      <div
        className="flex gap-2 overflow-x-auto pb-1"
        role="tablist"
        aria-label="Storage layout sections"
      >
        {views.map((view) => (
          <button
            key={view}
            type="button"
            role="tab"
            aria-selected={currentView === view}
            onClick={() =>
              void setFilters({ view, search: "", active: "all", page: 1 })
            }
            className={cn(
              "min-h-10 shrink-0 rounded-xl px-4 text-sm font-semibold transition-colors",
              currentView === view
                ? "bg-slate-950 text-white"
                : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
            )}
          >
            {viewLabels[view]}
          </button>
        ))}
      </div>

      {currentView !== "types" ? (
        <Panel className="p-4 sm:p-5">
          <p className="mb-2 text-xs font-bold tracking-wide text-slate-500 uppercase">
            Warehouse context
          </p>
          <Select
            ariaLabel="Warehouse"
            value={filters.warehouse}
            options={warehouseOptions}
            onValueChange={(warehouse) =>
              void setFilters({ warehouse, search: "", page: 1 })
            }
            placeholder={
              warehouses.isPending
                ? "Loading warehouses…"
                : "Select a warehouse"
            }
            disabled={warehouses.isPending || warehouseOptions.length === 0}
            className="max-w-xl"
          />
        </Panel>
      ) : null}

      <Panel className="overflow-hidden">
        <form
          className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:p-5"
          onSubmit={(event) => {
            event.preventDefault();
            const data = new FormData(event.currentTarget);
            void setFilters({
              search: String(data.get("search") ?? "").trim(),
              page: 1,
            });
          }}
        >
          {currentView !== "types" ? (
            <label className="relative flex-1">
              <span className="sr-only">Search {currentView}</span>
              <Search className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
              <input
                key={`${currentView}-${filters.search}`}
                name="search"
                defaultValue={filters.search}
                placeholder={`Search ${currentView}`}
                className="h-11 w-full rounded-xl border border-slate-300 bg-white pr-4 pl-10 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
              />
            </label>
          ) : (
            <div className="flex-1" />
          )}
          <Select
            ariaLabel="Filter by status"
            value={filters.active}
            options={statusOptions}
            onValueChange={(active) => void setFilters({ active, page: 1 })}
            className="sm:w-40"
          />
          {currentView !== "types" ? (
            <Button type="submit" variant="secondary">
              Search
            </Button>
          ) : null}
          {currentView !== "types" && filters.search ? (
            <Button
              type="button"
              variant="ghost"
              onClick={() => void setFilters({ search: "", page: 1 })}
            >
              Clear
            </Button>
          ) : null}
        </form>

        {currentView === "types" ? (
          locationTypes.isPending ? (
            <LoadingState />
          ) : locationTypes.isError ? (
            <ErrorState
              error={locationTypes.error}
              retry={() => void locationTypes.refetch()}
            />
          ) : locationTypes.data.length === 0 ? (
            <EmptyState
              icon={Boxes}
              title="No location types found"
              message="Create a location type to describe warehouse capabilities."
            />
          ) : (
            <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
              {locationTypes.data.map((item) => (
                <article
                  key={item.location_type_id}
                  className="rounded-xl border border-slate-200 p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h2 className="font-bold text-slate-950">{item.name}</h2>
                      <p className="mt-0.5 font-mono text-xs text-slate-500">
                        {item.code}
                      </p>
                    </div>
                    <StatusBadge tone={item.is_active ? "success" : "neutral"}>
                      {item.is_active ? "Active" : "Inactive"}
                    </StatusBadge>
                  </div>
                  <p className="mt-3 min-h-10 text-sm text-slate-600">
                    {item.description || "No description"}
                  </p>
                  <div className="mt-3 flex flex-wrap gap-1.5">
                    {[
                      ["Receive", item.allows_receiving],
                      ["Store", item.allows_storage],
                      ["Pick", item.allows_picking],
                      ["Ship", item.allows_shipping],
                    ]
                      .filter(([, allowed]) => allowed)
                      .map(([label]) => (
                        <span
                          key={String(label)}
                          className="rounded-full bg-cyan-50 px-2 py-1 text-xs font-semibold text-cyan-800"
                        >
                          {label}
                        </span>
                      ))}
                  </div>
                  {canWrite ? (
                    <div className="mt-4 flex justify-end gap-1 border-t border-slate-100 pt-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setEditor({ kind: "type", item })}
                      >
                        <Pencil className="size-4" />
                        Edit
                      </Button>
                      {item.is_active ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-rose-700"
                          onClick={() =>
                            setDeactivateTarget({ kind: "type", item })
                          }
                        >
                          Deactivate
                        </Button>
                      ) : null}
                    </div>
                  ) : null}
                </article>
              ))}
            </div>
          )
        ) : !filters.warehouse ? (
          <EmptyState
            icon={Layers3}
            title="Select a warehouse"
            message="Choose the warehouse whose storage layout you want to manage."
          />
        ) : currentView === "zones" ? (
          zones.isPending ? (
            <LoadingState />
          ) : zones.isError ? (
            <ErrorState
              error={zones.error}
              retry={() => void zones.refetch()}
            />
          ) : zones.data.length === 0 ? (
            <EmptyState
              icon={Layers3}
              title="No zones found"
              message="Create the first operational zone for this warehouse."
            />
          ) : (
            <div className="grid gap-3 p-4 sm:p-5 md:grid-cols-2 xl:grid-cols-3">
              {zones.data.map((item) => (
                <article
                  key={item.zone_id}
                  className="rounded-xl border border-slate-200 p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h2 className="font-bold text-slate-950">{item.name}</h2>
                      <p className="mt-0.5 font-mono text-xs text-slate-500">
                        {item.code}
                      </p>
                    </div>
                    <StatusBadge tone={item.is_active ? "success" : "neutral"}>
                      {item.is_active ? "Active" : "Inactive"}
                    </StatusBadge>
                  </div>
                  <p className="mt-3 text-sm text-slate-600">
                    {item.description || "No description"}
                  </p>
                  <p className="mt-3 text-xs font-semibold text-slate-500">
                    {item.location_count.toLocaleString()} locations
                  </p>
                  {canWrite ? (
                    <div className="mt-4 flex justify-end gap-1 border-t border-slate-100 pt-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setEditor({ kind: "zone", item })}
                      >
                        <Pencil className="size-4" />
                        Edit
                      </Button>
                      {item.is_active ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-rose-700"
                          onClick={() =>
                            setDeactivateTarget({ kind: "zone", item })
                          }
                        >
                          Deactivate
                        </Button>
                      ) : null}
                    </div>
                  ) : null}
                </article>
              ))}
            </div>
          )
        ) : locations.isPending ? (
          <LoadingState />
        ) : locations.isError ? (
          <ErrorState
            error={locations.error}
            retry={() => void locations.refetch()}
          />
        ) : locationItems.length === 0 ? (
          <EmptyState
            icon={MapPin}
            title="No locations found"
            message="Create a physical location after configuring a zone and location type."
          />
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full border-collapse text-left">
                <thead>
                  <tr className="bg-slate-50 text-xs font-bold tracking-wide text-slate-500 uppercase">
                    <th className="px-5 py-3">Location</th>
                    <th className="px-5 py-3">Zone</th>
                    <th className="px-5 py-3">Type</th>
                    <th className="px-5 py-3">Coordinates</th>
                    <th className="px-5 py-3">Flags</th>
                    <th className="px-5 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {locationItems.map((item) => (
                    <tr key={item.location_id} className="hover:bg-slate-50/70">
                      <td className="px-5 py-4">
                        <p className="font-semibold text-slate-950">
                          {item.code}
                        </p>
                        <p className="mt-0.5 text-xs text-slate-500">
                          {item.barcode || "No barcode"}
                        </p>
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-700">
                        {item.zone_code}
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-700">
                        {item.location_type_code}
                      </td>
                      <td className="px-5 py-4 text-sm text-slate-600">
                        {[item.aisle, item.bay, item.level_no, item.position_no]
                          .filter(Boolean)
                          .join(" / ") || "—"}
                      </td>
                      <td className="px-5 py-4">
                        <div className="flex flex-wrap gap-1">
                          {item.is_active ? (
                            <StatusBadge tone="success">Active</StatusBadge>
                          ) : (
                            <StatusBadge tone="neutral">Inactive</StatusBadge>
                          )}
                          {item.is_pick_face ? (
                            <StatusBadge tone="info">Pick face</StatusBadge>
                          ) : null}
                          {item.is_locked ? (
                            <StatusBadge tone="warning">
                              <LockKeyhole className="size-3" />
                              Locked
                            </StatusBadge>
                          ) : null}
                        </div>
                      </td>
                      <td className="px-5 py-4">
                        <div className="flex justify-end gap-1">
                          {canWrite ? (
                            <>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() =>
                                  setEditor({ kind: "location", item })
                                }
                              >
                                <Pencil className="size-4" />
                                Edit
                              </Button>
                              {item.is_active ? (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  className="text-rose-700"
                                  onClick={() =>
                                    setDeactivateTarget({
                                      kind: "location",
                                      item,
                                    })
                                  }
                                >
                                  Deactivate
                                </Button>
                              ) : null}
                            </>
                          ) : null}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <ul className="divide-y divide-slate-200 md:hidden">
              {locationItems.map((item) => (
                <li key={item.location_id} className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="font-semibold text-slate-950">
                        {item.code}
                      </p>
                      <p className="mt-1 text-xs text-slate-500">
                        {item.zone_code} · {item.location_type_code}
                      </p>
                    </div>
                    <StatusBadge tone={item.is_active ? "success" : "neutral"}>
                      {item.is_active ? "Active" : "Inactive"}
                    </StatusBadge>
                  </div>
                  <p className="mt-3 text-sm text-slate-600">
                    {[item.aisle, item.bay, item.level_no, item.position_no]
                      .filter(Boolean)
                      .join(" / ") || "No coordinates"}
                  </p>
                  {canWrite ? (
                    <div className="mt-3 flex justify-end gap-1 border-t border-slate-100 pt-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setEditor({ kind: "location", item })}
                      >
                        <Pencil className="size-4" />
                        Edit
                      </Button>
                      {item.is_active ? (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-rose-700"
                          onClick={() =>
                            setDeactivateTarget({ kind: "location", item })
                          }
                        >
                          Deactivate
                        </Button>
                      ) : null}
                    </div>
                  ) : null}
                </li>
              ))}
            </ul>
          </>
        )}

        {currentView === "locations" &&
        !locations.isPending &&
        !locations.isError &&
        totalPages > 0 ? (
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

      {editor?.kind === "type" ? (
        <LocationTypeDialog
          open
          locationType={editor.item}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      {editor?.kind === "zone" && filters.warehouse ? (
        <ZoneDialog
          open
          warehouseId={filters.warehouse}
          zone={editor.item}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      {editor?.kind === "location" && filters.warehouse ? (
        <LocationDialog
          open
          warehouseId={filters.warehouse}
          location={editor.item}
          zones={allZones.data ?? []}
          locationTypes={allLocationTypes.data ?? []}
          onOpenChange={(open) => {
            if (!open) setEditor(null);
          }}
        />
      ) : null}
      <DeactivateDialog
        target={deactivateTarget}
        pending={deactivate.isPending}
        error={deactivate.error}
        onConfirm={() => {
          if (deactivateTarget) deactivate.mutate(deactivateTarget);
        }}
        onOpenChange={(open) => {
          if (!open) {
            deactivate.reset();
            setDeactivateTarget(undefined);
          }
        }}
      />
    </div>
  );
}
