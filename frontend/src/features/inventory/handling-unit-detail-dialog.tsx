"use client";

import { useQuery } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import {
  getLocation,
  storageLayoutKeys,
} from "@/features/storage-layout/storage-layout-api";
import {
  getHandlingUnitType,
  unitsPackagingKeys,
} from "@/features/units-packaging/units-packaging-api";
import {
  balanceKeys,
  getHandlingUnit,
  handlingUnitKeys,
  listBalances,
} from "./inventory-api";

function Detail({ label, value }: { label: string; value?: string | null }) {
  return (
    <div>
      <dt className="text-xs font-semibold text-slate-500">{label}</dt>
      <dd className="mt-1 text-sm break-words text-slate-900">
        {value === null || value === undefined || value === "" ? "—" : value}
      </dd>
    </div>
  );
}

export function HandlingUnitDetailDialog({
  id,
  timezone,
  onOpenChange,
}: {
  id: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const unit = useQuery({
    queryKey: handlingUnitKeys.detail(id),
    queryFn: () => getHandlingUnit(id),
  });
  const type = useQuery({
    queryKey: unitsPackagingKeys.handling(
      unit.data?.handling_unit_type_id ?? "none",
    ),
    queryFn: () => getHandlingUnitType(unit.data!.handling_unit_type_id),
    enabled: Boolean(unit.data?.handling_unit_type_id),
  });
  const location = useQuery({
    queryKey: storageLayoutKeys.location(
      unit.data?.current_location_id ?? "none",
    ),
    queryFn: () => getLocation(unit.data!.current_location_id!),
    enabled: Boolean(unit.data?.current_location_id),
  });
  const parent = useQuery({
    queryKey: handlingUnitKeys.detail(
      unit.data?.parent_handling_unit_id ?? "none",
    ),
    queryFn: () => getHandlingUnit(unit.data!.parent_handling_unit_id!),
    enabled: Boolean(unit.data?.parent_handling_unit_id),
  });
  const contentFilters = {
    ownerId: unit.data?.owner_id ?? "",
    warehouseId: unit.data?.warehouse_id ?? "",
    handlingUnitId: unit.data?.handling_unit_id,
    search: "",
    includeZero: false,
    page: 1,
    pageSize: 100,
  };
  const contents = useQuery({
    queryKey: balanceKeys.list(contentFilters),
    queryFn: () => listBalances(contentFilters),
    enabled: Boolean(unit.data),
  });

  return (
    <OperationDialog
      title="Handling unit"
      description="Container identity, hierarchy and current placement"
      onOpenChange={onOpenChange}
      closeLabel="Close handling-unit details"
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={unit.isFetching}
          onClick={() => void unit.refetch()}
        >
          <RefreshCw className="size-4" /> Refresh handling unit
        </Button>
        {unit.error ? (
          <div
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {unit.error.message}
          </div>
        ) : null}
        {unit.isPending ? (
          <p className="flex items-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-4 animate-spin" /> Loading handling
            unit…
          </p>
        ) : null}
        {unit.data ? (
          <>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2">
              <Detail label="Barcode" value={unit.data.barcode} />
              <Detail
                label="Type"
                value={
                  type.data
                    ? `${type.data.code} · ${type.data.name}`
                    : type.isPending
                      ? "Loading type…"
                      : "Type unavailable"
                }
              />
              <Detail
                label="Status"
                value={unit.data.is_closed ? "Closed" : "Open"}
              />
              <Detail
                label="Current location"
                value={
                  !unit.data.current_location_id
                    ? "Not placed"
                    : location.data
                      ? `${location.data.code}${location.data.zone_code ? ` · ${location.data.zone_code}` : ""}`
                      : location.isPending
                        ? "Loading location…"
                        : "Location unavailable"
                }
              />
              <Detail
                label="Parent handling unit"
                value={
                  !unit.data.parent_handling_unit_id
                    ? "Root unit"
                    : parent.data
                      ? parent.data.barcode
                      : parent.isPending
                        ? "Loading parent…"
                        : "Parent unavailable"
                }
              />
              <Detail
                label="Registered"
                value={new Intl.DateTimeFormat("en-GB", {
                  dateStyle: "medium",
                  timeStyle: "short",
                  timeZone: timezone,
                }).format(new Date(unit.data.created_at))}
              />
              <Detail
                label="Internal handling-unit reference"
                value={unit.data.handling_unit_id}
              />
            </dl>
            <section className="mt-5 overflow-hidden rounded-xl border border-slate-200">
              <div className="border-b border-slate-200 bg-slate-50 px-4 py-3">
                <h3 className="font-bold text-slate-950">
                  Current stock contents
                </h3>
                <p className="mt-1 text-xs text-slate-600">
                  Positive inventory balances currently assigned to this
                  container.
                </p>
              </div>
              {contents.isPending ? (
                <p className="flex items-center gap-2 p-4 text-sm text-slate-600">
                  <LoaderCircle className="size-4 animate-spin" /> Loading
                  contents…
                </p>
              ) : contents.error ? (
                <p
                  role="alert"
                  className="m-4 rounded-lg bg-rose-50 p-3 text-sm text-rose-900"
                >
                  {contents.error.message}
                </p>
              ) : !contents.data?.items.length ? (
                <p className="p-4 text-sm text-slate-600">
                  {unit.data.child_count > 0
                    ? `No direct stock balances. This unit contains ${unit.data.child_count} child handling unit${unit.data.child_count === 1 ? "" : "s"}.`
                    : "Empty handling unit—ready to receive stock."}
                </p>
              ) : (
                <div className="divide-y divide-slate-200">
                  {contents.data.items.map((balance) => (
                    <div
                      key={balance.balance_id}
                      className="grid gap-2 p-4 text-sm sm:grid-cols-[minmax(0,1fr)_auto]"
                    >
                      <div>
                        <p className="font-semibold text-slate-950">
                          {balance.item_code} · {balance.item_name}
                        </p>
                        <p className="mt-1 text-xs text-slate-600">
                          {balance.lot_number
                            ? `Lot ${balance.lot_number} · `
                            : ""}
                          {balance.inventory_status_code} ·{" "}
                          {balance.location_code}
                        </p>
                      </div>
                      <div className="sm:text-right">
                        <p className="font-semibold text-slate-950">
                          {balance.on_hand_qty} {balance.uom_code}
                        </p>
                        <p className="mt-1 text-xs text-slate-600">
                          {balance.reserved_qty} reserved
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>
          </>
        ) : null}
      </div>
    </OperationDialog>
  );
}
