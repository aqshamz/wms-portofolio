"use client";

import { useId, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import { HandlingUnitCreateDialog } from "@/features/inventory/handling-unit-create-dialog";
import {
  handlingUnitKeys,
  listHandlingUnits,
} from "@/features/inventory/inventory-api";
import { localDateTime } from "@/features/outbound-orders/outbound-order-schema";
import {
  allPages,
  assignDriver,
  dispatchHandlingUnit,
  dispatchLine,
  fulfillmentKeys,
  getFulfillment,
  listCarriers,
  listDrivers,
  packLine,
  transitionFulfillment,
  type FulfillmentMode,
  type PackingLine,
  type ShipmentLine,
} from "./fulfillment-api";

type Action =
  | { kind: "pack"; line: PackingLine; key: string }
  | { kind: "line"; line: ShipmentLine; key: string }
  | {
      kind: "hu";
      huId: string;
      label: string;
      lines: ShipmentLine[];
      key: string;
    }
  | { kind: "complete" | "cancel" }
  | { kind: "driver"; driverId: string };
export function dispatchBalances(lines: ShipmentLine[]) {
  const versions = new Map<string, number>();
  for (const line of lines) {
    const existing = versions.get(line.source_balance_id);
    if (existing !== undefined && existing !== line.source_balance_version)
      throw new Error("Refresh the HU: balance versions are inconsistent.");
    versions.set(line.source_balance_id, line.source_balance_version);
  }
  return [...versions].map(([balance_id, expected_version]) => ({
    balance_id,
    expected_version,
  }));
}
export function FulfillmentDetailDialog({
  mode,
  id,
  canWrite,
  canCancel,
  canCreateHU,
  onOpenChange,
}: {
  mode: FulfillmentMode;
  id: string;
  canWrite: boolean;
  canCancel: boolean;
  canCreateHU: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const fieldId = useId();
  const client = useQueryClient();
  const query = useQuery({
    queryKey: fulfillmentKeys.detail(mode, id),
    queryFn: () => getFulfillment(mode, id),
  });
  const doc = query.data;
  const packing = doc && "packing_id" in doc ? doc : undefined;
  const shipment = doc && "shipment_id" in doc ? doc : undefined;
  const [action, setAction] = useState<Action>();
  const [huId, setHuId] = useState("");
  const [huSearch, setHuSearch] = useState("");
  const [businessDate, setBusinessDate] = useState("");
  const [creatingHU, setCreatingHU] = useState(false);
  const [driverId, setDriverId] = useState("");
  const date =
    businessDate || (shipment?.business_date ?? localDateTime().slice(0, 10));
  const unitFilters = {
    ownerId: doc?.owner_id ?? "",
    warehouseId: doc?.warehouse_id ?? "",
    status: "open" as const,
    search: "",
    page: 1,
    pageSize: 100,
  };
  const units = useQuery({
    queryKey: [...handlingUnitKeys.list(unitFilters), "all-pages"],
    queryFn: () =>
      allPages((page) => listHandlingUnits({ ...unitFilters, page })),
    enabled: Boolean(packing && canWrite && packing.status_code === "OPEN"),
  });
  const drivers = useQuery({
    queryKey: [
      "outbound-fulfillment",
      "drivers",
      shipment?.carrier_code ?? "all",
    ],
    queryFn: async () => {
      const carriers = await listCarriers();
      const chosen = shipment?.carrier_code
        ? carriers.filter((c) => c.code === shipment.carrier_code)
        : carriers;
      const rows = await Promise.all(
        chosen.map(async (c) =>
          (await listDrivers(c.carrier_id)).map((d) => ({
            ...d,
            label: `${d.name} (${d.code}) · ${c.name}`,
          })),
        ),
      );
      return rows.flat();
    },
    enabled: Boolean(
      shipment &&
      canWrite &&
      shipment.status_code === "PLANNED" &&
      !shipment.drivers?.some((d) => d.is_primary),
    ),
  });
  const mutation = useMutation({
    mutationFn: async (selected: Action) => {
      if (!doc) throw new Error("Refresh the document first.");
      if (selected.kind === "pack")
        return packLine(id, selected.line.outbound_check_line_id, {
          expected_balance_version: selected.line.source_balance_version,
          handling_unit_id: huId || undefined,
          business_date: date,
          operation_key: selected.key,
        });
      if (selected.kind === "line")
        return dispatchLine(id, selected.line.packing_line_id, {
          expected_balance_version: selected.line.source_balance_version,
          business_date: date,
          operation_key: selected.key,
        });
      if (selected.kind === "hu")
        return dispatchHandlingUnit(id, selected.huId, {
          business_date: date,
          operation_key: selected.key,
          expected_balances: dispatchBalances(selected.lines),
        });
      if (selected.kind === "driver") {
        await assignDriver(id, selected.driverId);
        return getFulfillment(mode, id);
      }
      return transitionFulfillment(mode, id, selected.kind);
    },
    onSuccess: async (next) => {
      client.setQueryData(fulfillmentKeys.detail(mode, id), next);
      setAction(undefined);
      setDriverId("");
      await client.invalidateQueries();
      toast.success("Document updated.");
    },
    onError: (error: Error) => toast.error(error.message),
  });
  const busy = mutation.isPending;
  const error = mutation.error || query.error;
  const packingLines = packing?.lines ?? [];
  const shipmentLines = shipment?.lines ?? [];
  const posted = packing
    ? packingLines.filter((l) => l.packing_line_id).length
    : shipmentLines.filter((l) => l.shipment_line_id).length;
  const total = packing ? packingLines.length : shipmentLines.length;
  const open = packing
    ? packing.status_code === "OPEN"
    : shipment?.status_code === "PLANNED";
  const hasPrimary = shipment?.drivers?.some((d) => d.is_primary);
  const huGroups = new Map<string, ShipmentLine[]>();
  for (const line of shipmentLines)
    if (line.handling_unit_id)
      huGroups.set(line.handling_unit_id, [
        ...(huGroups.get(line.handling_unit_id) ?? []),
        line,
      ]);
  const selectableUnits = (units.data ?? []).filter(
    (unit) =>
      !unit.parent_handling_unit_id &&
      !unit.child_count &&
      (unit.current_location_id === packing?.packing_location_id ||
        unit.current_location_id === null ||
        unit.handling_unit_id ===
          (action?.kind === "pack"
            ? action.line.handling_unit_id
            : undefined)) &&
      unit.barcode.toLowerCase().includes(huSearch.trim().toLowerCase()),
  );
  const unitOptions = [
    { value: "__loose__", label: "No handling unit (loose stock)" },
    ...selectableUnits.map((unit) => ({
      value: unit.handling_unit_id,
      label: unit.barcode,
    })),
  ];
  if (
    action?.kind === "pack" &&
    action.line.handling_unit_id &&
    !unitOptions.some((o) => o.value === action.line.handling_unit_id)
  )
    unitOptions.push({
      value: action.line.handling_unit_id,
      label: action.line.handling_unit_barcode || action.line.handling_unit_id,
    });
  if (huId && !unitOptions.some((o) => o.value === huId)) {
    const unit = units.data?.find((u) => u.handling_unit_id === huId);
    unitOptions.push({ value: huId, label: unit?.barcode || huId });
  }
  const startPack = (line: PackingLine) => {
    mutation.reset();
    setHuId(line.handling_unit_id ?? "");
    setHuSearch("");
    setAction({ kind: "pack", line, key: crypto.randomUUID() });
  };
  const startAction = (selected: Action) => {
    mutation.reset();
    setAction(selected);
  };
  return (
    <>
      <OperationDialog
        title={id}
        description={
          mode === "packings"
            ? "Checked stock, packing and handling units"
            : "Shipment manifest and atomic handling-unit dispatch"
        }
        className="max-w-5xl"
        busy={busy || creatingHU}
        onOpenChange={onOpenChange}
      >
        <div className="space-y-5">
          <Button
            variant="secondary"
            disabled={busy || query.isFetching}
            onClick={() => {
              setAction(undefined);
              mutation.reset();
              void query.refetch();
            }}
          >
            Refresh document
          </Button>
          {query.isPending && <p>Loading document…</p>}
          {error && (
            <p
              role="alert"
              className="rounded-xl bg-rose-50 p-4 text-sm text-rose-700"
            >
              {error.message} If stock changed, refresh before starting a new
              action.
            </p>
          )}
          {doc && (
            <>
              <section className="rounded-xl bg-slate-50 p-4">
                <StatusBadge
                  tone={doc.status_code === "CANCELLED" ? "danger" : "info"}
                >
                  {doc.status_code}
                </StatusBadge>
                <dl className="mt-4 grid gap-4 text-sm sm:grid-cols-2 lg:grid-cols-3">
                  <div>
                    <dt className="text-slate-500">Progress</dt>
                    <dd>
                      {posted} / {total} lines{" "}
                      {packing ? "packed" : "dispatched"}
                    </dd>
                  </div>
                  {packing ? (
                    <>
                      <div>
                        <dt className="text-slate-500">Client DO</dt>
                        <dd>{packing.client_delivery_order_no}</dd>
                      </div>
                      <div>
                        <dt className="text-slate-500">Packing location</dt>
                        <dd>{packing.packing_location_code || "—"}</dd>
                      </div>
                    </>
                  ) : (
                    shipment && (
                      <>
                        <div>
                          <dt className="text-slate-500">Business date</dt>
                          <dd>{shipment.business_date}</dd>
                        </div>
                        <div>
                          <dt className="text-slate-500">Vehicle / seal</dt>
                          <dd>
                            {shipment.vehicle_number || "—"} /{" "}
                            {shipment.seal_number || "—"}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-slate-500">Carrier service</dt>
                          <dd>
                            {shipment.carrier_code || "Self-managed"}{" "}
                            {shipment.carrier_service_code}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-slate-500">Tracking / route</dt>
                          <dd>
                            {shipment.tracking_number || "—"} /{" "}
                            {shipment.route_reference || "—"}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-slate-500">Notes</dt>
                          <dd>{shipment.notes || "—"}</dd>
                        </div>
                      </>
                    )
                  )}
                </dl>
              </section>
              {open && !action && (
                <div className="flex flex-wrap gap-3">
                  {canWrite && (
                    <Button
                      disabled={
                        busy ||
                        total === 0 ||
                        posted !== total ||
                        Boolean(shipment?.carrier_service_id && !hasPrimary)
                      }
                      onClick={() => startAction({ kind: "complete" })}
                    >
                      {packing ? "Complete packing" : "Complete shipment"}
                    </Button>
                  )}
                  {canCancel && (
                    <Button
                      variant="secondary"
                      disabled={busy || posted > 0}
                      onClick={() => startAction({ kind: "cancel" })}
                    >
                      {packing
                        ? "Cancel empty packing"
                        : "Cancel undispatched shipment"}
                    </Button>
                  )}
                </div>
              )}
              {shipment && (
                <section className="space-y-3 rounded-xl border border-slate-200 p-4">
                  <h3 className="font-bold">Drivers</h3>
                  {shipment.drivers?.map((d) => (
                    <p key={d.driver_id} className="text-sm">
                      {d.driver_name} ({d.driver_code})
                      {d.is_primary ? " · Primary" : ""}
                    </p>
                  ))}
                  {open && canWrite && !hasPrimary && !action && (
                    <div className="flex flex-col gap-3 sm:flex-row">
                      <Select
                        ariaLabel="Primary driver"
                        value={driverId}
                        options={(drivers.data ?? []).map((d) => ({
                          value: d.driver_id,
                          label: d.label,
                        }))}
                        placeholder="Select an active driver"
                        onValueChange={setDriverId}
                      />
                      <Button
                        variant="secondary"
                        disabled={!driverId || busy}
                        onClick={() =>
                          startAction({ kind: "driver", driverId })
                        }
                      >
                        Assign primary driver
                      </Button>
                    </div>
                  )}
                  {drivers.error && (
                    <p role="alert" className="text-sm text-rose-700">
                      {drivers.error.message}
                    </p>
                  )}
                  {shipment.carrier_service_id && !hasPrimary && (
                    <p className="text-sm text-amber-800">
                      Assign a primary driver before completing this
                      carrier-service shipment.
                    </p>
                  )}
                </section>
              )}
              {action && (
                <section className="space-y-4 rounded-xl border border-amber-200 bg-amber-50 p-4">
                  <p className="text-sm">
                    {action.kind === "pack"
                      ? `Pack the full checked quantity: ${action.line.expected_qty} ${action.line.uom_code} of ${action.line.item_code}?`
                      : action.kind === "hu"
                        ? `Dispatch the entire HU ${action.label}, covering ${action.lines.length} packing lines? This removes its stock from the warehouse in one transaction.`
                        : action.kind === "line"
                          ? `Dispatch ${action.line.shipped_qty} ${action.line.uom_code} of ${action.line.item_code}? This removes stock from the warehouse.`
                          : action.kind === "complete"
                            ? packing
                              ? "Complete packing? Every checked line must already be posted."
                              : "Complete the shipment? Every manifest line must already be dispatched."
                            : action.kind === "cancel"
                              ? "Cancel this empty document? It will remain in audit history."
                              : "Assign this driver as the shipment’s primary driver?"}
                  </p>
                  {(action.kind === "pack" ||
                    action.kind === "hu" ||
                    action.kind === "line") && (
                    <FormField
                      label="Movement business date"
                      htmlFor={`${fieldId}-date`}
                      required
                    >
                      <Input
                        id={`${fieldId}-date`}
                        type="date"
                        value={date}
                        disabled={busy}
                        onChange={(e) => setBusinessDate(e.target.value)}
                      />
                    </FormField>
                  )}
                  {action.kind === "pack" && (
                    <>
                      <FormField
                        label="Search / scan HU barcode"
                        htmlFor={`${fieldId}-search`}
                      >
                        <Input
                          id={`${fieldId}-search`}
                          value={huSearch}
                          disabled={busy}
                          onChange={(e) => setHuSearch(e.target.value)}
                        />
                      </FormField>
                      <FormField
                        label="Handling unit"
                        htmlFor={`${fieldId}-hu`}
                      >
                        <Select
                          className="mt-2"
                          id={`${fieldId}-hu`}
                          ariaLabel="Handling unit"
                          value={huId || "__loose__"}
                          options={unitOptions}
                          disabled={
                            busy || Boolean(action.line.handling_unit_id)
                          }
                          onValueChange={(value) =>
                            setHuId(value === "__loose__" ? "" : value)
                          }
                        />
                      </FormField>
                      {action.line.handling_unit_id && (
                        <p className="text-sm text-slate-600">
                          Stock already in a handling unit must stay in that
                          unit.
                        </p>
                      )}
                      {units.error && (
                        <p role="alert" className="text-sm text-rose-700">
                          Unable to load HUs: {units.error.message}
                        </p>
                      )}
                      {canCreateHU && !action.line.handling_unit_id && (
                        <Button
                          variant="secondary"
                          disabled={busy}
                          onClick={() => setCreatingHU(true)}
                        >
                          Create empty handling unit
                        </Button>
                      )}
                    </>
                  )}
                  <div className="flex gap-2">
                    <Button
                      disabled={
                        busy ||
                        Boolean(
                          (action.kind === "pack" ||
                            action.kind === "line" ||
                            action.kind === "hu") &&
                          !date,
                        )
                      }
                      onClick={() => {
                        if (!busy) mutation.mutate(action);
                      }}
                    >
                      {busy ? "Posting…" : "Confirm"}
                    </Button>
                    <Button
                      variant="secondary"
                      disabled={busy}
                      onClick={() => setAction(undefined)}
                    >
                      Back
                    </Button>
                  </div>
                </section>
              )}
              <section className="space-y-4">
                <h3 className="font-bold">
                  {packing ? "Checked lines" : "Manifest lines"} ({total})
                </h3>
                {packingLines.map((line) => (
                  <article
                    key={line.outbound_check_line_id}
                    className="space-y-3 rounded-xl border border-slate-200 p-4 text-sm"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <h4 className="font-bold">
                        {line.item_code} · {line.expected_qty} {line.uom_code}
                      </h4>
                      {line.serial_no && (
                        <p className="text-slate-600">
                          Serial: {line.serial_no}
                        </p>
                      )}
                      <StatusBadge
                        tone={line.packing_line_id ? "success" : "info"}
                      >
                        {line.packing_line_id ? "Packed" : "Waiting"}
                      </StatusBadge>
                    </div>
                    <p className="text-slate-600">
                      Lot: {line.lot_number || "—"} · HU:{" "}
                      {line.handling_unit_barcode ||
                        line.handling_unit_id ||
                        "Loose stock"}
                    </p>
                    {line.movement_id && (
                      <p className="text-xs text-slate-500">
                        Movement: {line.movement_id}
                      </p>
                    )}
                    {open && canWrite && !line.packing_line_id && !action && (
                      <Button
                        size="sm"
                        variant="secondary"
                        disabled={busy}
                        onClick={() => startPack(line)}
                      >
                        Pack line
                      </Button>
                    )}
                  </article>
                ))}
                {shipmentLines.map((line) => (
                  <article
                    key={line.packing_line_id}
                    className="space-y-3 rounded-xl border border-slate-200 p-4 text-sm"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <h4 className="font-bold">
                        {line.item_code} · {line.shipped_qty} {line.uom_code}
                      </h4>
                      {line.serial_no && (
                        <p className="text-slate-600">
                          Serial: {line.serial_no}
                        </p>
                      )}
                      <StatusBadge
                        tone={line.shipment_line_id ? "success" : "info"}
                      >
                        {line.shipment_line_id ? "Dispatched" : "Waiting"}
                      </StatusBadge>
                    </div>
                    <p className="text-slate-600">
                      DO: {line.outbound_id} · Lot: {line.lot_number || "—"}
                    </p>
                    <p className="text-slate-600">
                      HU:{" "}
                      {line.handling_unit_barcode ||
                        line.handling_unit_id ||
                        "Loose stock"}
                    </p>
                    {line.movement_id && (
                      <p className="text-xs text-slate-500">
                        Movement: {line.movement_id}
                      </p>
                    )}
                    {open &&
                      canWrite &&
                      !line.shipment_line_id &&
                      !line.handling_unit_id &&
                      !action && (
                        <Button
                          size="sm"
                          variant="secondary"
                          disabled={busy}
                          onClick={() =>
                            startAction({
                              kind: "line",
                              line,
                              key: crypto.randomUUID(),
                            })
                          }
                        >
                          Dispatch loose line
                        </Button>
                      )}
                  </article>
                ))}
              </section>
              {huGroups.size > 0 && (
                <section className="space-y-3 rounded-xl border border-cyan-200 p-4">
                  <h3 className="font-bold">Handling-unit dispatch</h3>
                  <p className="text-sm text-slate-600">
                    A container cannot be split across shipments. Nested
                    containers are not supported by this dispatch action.
                  </p>
                  {[...huGroups].map(([unitId, lines]) => {
                    const label = lines[0].handling_unit_barcode || unitId;
                    return (
                      <div
                        key={unitId}
                        className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 pt-3"
                      >
                        <p className="text-sm font-semibold">
                          {label} · {lines.length} packing lines
                        </p>
                        {open && canWrite && !action && (
                          <Button
                            variant="secondary"
                            size="sm"
                            disabled={
                              busy ||
                              lines.some((l) => Boolean(l.shipment_line_id))
                            }
                            onClick={() =>
                              startAction({
                                kind: "hu",
                                huId: unitId,
                                label,
                                lines,
                                key: crypto.randomUUID(),
                              })
                            }
                          >
                            Dispatch entire HU
                          </Button>
                        )}
                      </div>
                    );
                  })}
                </section>
              )}
            </>
          )}
        </div>
      </OperationDialog>
      {creatingHU && packing && (
        <HandlingUnitCreateDialog
          ownerId={packing.owner_id}
          warehouseId={packing.warehouse_id}
          defaultLocationId={packing.packing_location_id}
          onOpenChange={setCreatingHU}
          onCreated={(unit) => {
            setHuId(unit.handling_unit_id);
            setHuSearch("");
            void units.refetch();
          }}
        />
      )}
    </>
  );
}
