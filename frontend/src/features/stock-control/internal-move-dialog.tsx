"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Decimal from "decimal.js";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
  balanceKeys,
  getBalance,
  listSerialStates,
  movementKeys,
  serialStateKeys,
} from "@/features/inventory/inventory-api";
import {
  listLocations,
  listLocationTypes,
  storageLayoutKeys,
} from "@/features/storage-layout/storage-layout-api";
import {
  listStockControlReasons,
  postInternalMove,
  stockControlKeys,
} from "./stock-control-api";

const quantityPattern = /^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$/;

function businessDate(timezone: string) {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}

export function InternalMoveDialog({
  balanceId,
  timezone,
  onOpenChange,
}: {
  balanceId: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [targetSearch, setTargetSearch] = useState("");
  const [targetLocationId, setTargetLocationId] = useState("");
  const [serialSearch, setSerialSearch] = useState("");
  const [serialId, setSerialId] = useState("");
  const [quantity, setQuantity] = useState("");
  const [date, setDate] = useState(() => businessDate(timezone));
  const [reference, setReference] = useState(
    () => `IM-${businessDate(timezone).replaceAll("-", "")}`,
  );
  const [reasonCode, setReasonCode] = useState("");
  const [notes, setNotes] = useState("");

  const balance = useQuery({
    queryKey: balanceKeys.detail(balanceId),
    queryFn: () => getBalance(balanceId),
  });
  const source = balance.data;
  const locationTypes = useQuery({
    queryKey: storageLayoutKeys.locationTypes("active"),
    queryFn: () => listLocationTypes("active"),
  });
  const storageTypeId = locationTypes.data?.find(
    (type) => type.code === "STORAGE" && type.is_active,
  )?.location_type_id;
  const locationFilters = useMemo(
    () => ({
      warehouseId: source?.warehouse_id ?? "",
      locationTypeId: storageTypeId,
      search: targetSearch,
      active: "active" as const,
      page: 1,
      pageSize: 100,
    }),
    [source?.warehouse_id, storageTypeId, targetSearch],
  );
  const locations = useQuery({
    queryKey: storageLayoutKeys.locations(locationFilters),
    queryFn: () => listLocations(locationFilters),
    enabled: Boolean(source?.warehouse_id && storageTypeId),
  });
  const reasons = useQuery({
    queryKey: stockControlKeys.reasons(),
    queryFn: listStockControlReasons,
  });
  const serialFilters = useMemo(
    () => ({
      ownerId: source?.owner_id ?? "",
      warehouseId: source?.warehouse_id ?? "",
      locationId: source?.location_id,
      itemId: source?.item_id,
      lotId: source?.lot_id ?? undefined,
      handlingUnitId: source?.handling_unit_id ?? undefined,
      inventoryStatusId: source?.inventory_status_id,
      search: serialSearch,
      page: 1,
      pageSize: 100,
    }),
    [serialSearch, source],
  );
  const serials = useQuery({
    queryKey: serialStateKeys.list(serialFilters),
    queryFn: () => listSerialStates(serialFilters),
    enabled: Boolean(source?.serial_controlled),
  });

  const usableLocations = (locations.data?.items ?? []).filter(
    (location) =>
      !location.is_locked &&
      location.location_type_code === "STORAGE" &&
      location.location_id !== source?.location_id,
  );
  const internalMoveReasons = (reasons.data ?? []).filter(
    (reason) => reason.code === "RELOCATION" || reason.code === "CONSOLIDATION",
  );
  const effectiveReasonCode =
    reasonCode ||
    (internalMoveReasons.some((reason) => reason.code === "RELOCATION")
      ? "RELOCATION"
      : "");
  const effectiveQuantity = source?.serial_controlled
    ? "1"
    : source?.handling_unit_id
      ? source.available_qty
      : quantity;
  const selectedReason = internalMoveReasons.find(
    (reason) => reason.code === effectiveReasonCode,
  );

  const move = useMutation({
    mutationFn: async () => {
      if (!source) throw new Error("Source balance is unavailable.");
      if (!targetLocationId) throw new Error("Select a target location.");
      if (!date) throw new Error("Business date is required.");
      if (!reference.trim()) throw new Error("Movement reference is required.");
      if (!quantityPattern.test(effectiveQuantity)) {
        throw new Error(
          "Quantity must be a positive number with up to 6 decimals.",
        );
      }
      const requested = new Decimal(effectiveQuantity);
      if (requested.lte(0) || requested.gt(new Decimal(source.available_qty))) {
        throw new Error(
          `Quantity cannot exceed ${source.available_qty} ${source.uom_code} available.`,
        );
      }
      if (
        source.serial_controlled &&
        (effectiveQuantity !== "1" || !serialId)
      ) {
        throw new Error(
          "Select one serial number; serialized moves always move exactly 1 base unit.",
        );
      }
      if (source.handling_unit_id && !requested.eq(source.available_qty)) {
        throw new Error(
          "A handling unit must move with its full available balance.",
        );
      }
      if (selectedReason?.requires_note && !notes.trim()) {
        throw new Error(`Notes are required for ${selectedReason.name}.`);
      }
      return postInternalMove({
        operation_key: `stock.internal-move.${crypto.randomUUID()}`,
        business_date: date,
        source_document_id: reference.trim(),
        reason_code: effectiveReasonCode || undefined,
        notes: notes.trim() || undefined,
        source_balance_id: source.balance_id,
        target_location_id: targetLocationId,
        quantity: effectiveQuantity,
        serial_ids: serialId ? [serialId] : [],
        expected_version: source.version_no,
      });
    },
    onSuccess: (result) => {
      const movement = result.movements[0];
      toast.success(
        `Moved ${movement?.quantity ?? effectiveQuantity} ${movement?.uom_code ?? source?.uom_code ?? "units"} to ${movement?.to_location_code ?? "the target location"}.`,
      );
      void queryClient.invalidateQueries({
        queryKey: ["inventory", "balances"],
      });
      void queryClient.invalidateQueries({
        queryKey: ["inventory", "movements"],
      });
      void queryClient.invalidateQueries({ queryKey: movementKeys.types });
      onOpenChange(false);
    },
    onError: (error) => toast.error(error.message),
  });
  const error =
    balance.error ??
    locationTypes.error ??
    locations.error ??
    reasons.error ??
    serials.error ??
    move.error;

  return (
    <OperationDialog
      title="Post internal movement"
      description="Relocate available stock inside the same warehouse and write an immutable movement record."
      busy={move.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close internal movement"
    >
      {balance.isPending ? (
        <div className="flex min-h-48 items-center justify-center gap-2 text-sm text-slate-600">
          <LoaderCircle className="size-5 animate-spin" /> Loading source stock…
        </div>
      ) : !source ? (
        <p
          role="alert"
          className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
        >
          {balance.error?.message ?? "Source balance could not be loaded."}
        </p>
      ) : (
        <form
          className="space-y-5"
          onSubmit={(event) => {
            event.preventDefault();
            move.mutate();
          }}
        >
          <section className="grid gap-3 rounded-xl bg-slate-50 p-4 text-sm sm:grid-cols-3">
            <div>
              <p className="text-xs font-semibold text-slate-500 uppercase">
                Item
              </p>
              <p className="mt-1 font-semibold">{source.item_code}</p>
              <p className="text-xs text-slate-500">{source.item_name}</p>
            </div>
            <div>
              <p className="text-xs font-semibold text-slate-500 uppercase">
                Source
              </p>
              <p className="mt-1 font-semibold">{source.location_code}</p>
              <p className="text-xs text-slate-500">
                {source.inventory_status_code}
              </p>
            </div>
            <div>
              <p className="text-xs font-semibold text-slate-500 uppercase">
                Available
              </p>
              <p className="mt-1 font-semibold">
                {source.available_qty} {source.uom_code}
              </p>
              <p className="text-xs text-slate-500">
                {source.lot_number ??
                  source.handling_unit_barcode ??
                  "No lot or HU"}
              </p>
            </div>
          </section>

          {source.handling_unit_id ? (
            <p className="rounded-xl border border-amber-200 bg-amber-50 p-3 text-sm text-amber-900">
              This relocates the whole handling unit between STORAGE locations.
              It succeeds only when the HU has one positive, fully unreserved
              balance and no child handling units.
            </p>
          ) : null}
          {error ? (
            <p
              role="alert"
              className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
            >
              {error.message}
            </p>
          ) : null}

          <fieldset
            disabled={move.isPending}
            className="grid gap-4 sm:grid-cols-2"
          >
            <FormField
              label="Search target locations"
              htmlFor="internal-move-location-search"
            >
              <Input
                id="internal-move-location-search"
                maxLength={160}
                placeholder="Location or barcode"
                value={targetSearch}
                onChange={(event) => setTargetSearch(event.target.value)}
              />
            </FormField>
            <FormField
              label="Target location"
              htmlFor="internal-move-target"
              required
            >
              <Select
                id="internal-move-target"
                ariaLabel="Internal movement target location"
                className="mt-2"
                value={targetLocationId}
                options={usableLocations.map((location) => ({
                  value: location.location_id,
                  label: `${location.code}${location.zone_code ? ` · ${location.zone_code}` : ""}${location.is_pick_face ? " · Pick face" : ""}`,
                }))}
                placeholder={
                  locationTypes.isSuccess && !storageTypeId
                    ? "No active STORAGE location type"
                    : locations.isPending
                      ? "Loading locations…"
                      : "Select STORAGE location"
                }
                disabled={
                  locationTypes.isPending ||
                  locationTypes.isError ||
                  locations.isPending ||
                  locations.isError ||
                  !storageTypeId
                }
                onValueChange={setTargetLocationId}
              />
            </FormField>
            <FormField
              label="Quantity (base units)"
              htmlFor="internal-move-quantity"
              required
            >
              <Input
                id="internal-move-quantity"
                inputMode="decimal"
                value={effectiveQuantity}
                disabled={Boolean(
                  source.serial_controlled || source.handling_unit_id,
                )}
                placeholder={`Maximum ${source.available_qty}`}
                onChange={(event) => setQuantity(event.target.value)}
              />
            </FormField>
            <FormField
              label="Business date"
              htmlFor="internal-move-date"
              required
            >
              <Input
                id="internal-move-date"
                type="date"
                value={date}
                onChange={(event) => setDate(event.target.value)}
              />
            </FormField>
            {source.serial_controlled ? (
              <>
                <FormField
                  label="Search serial"
                  htmlFor="internal-move-serial-search"
                >
                  <Input
                    id="internal-move-serial-search"
                    maxLength={160}
                    placeholder="Serial number"
                    value={serialSearch}
                    onChange={(event) => setSerialSearch(event.target.value)}
                  />
                </FormField>
                <FormField
                  label="Serial number"
                  htmlFor="internal-move-serial"
                  required
                >
                  <Select
                    id="internal-move-serial"
                    ariaLabel="Serial number to move"
                    className="mt-2"
                    value={serialId}
                    options={(serials.data?.items ?? []).map((serial) => ({
                      value: serial.serial_id,
                      label: serial.serial_no,
                    }))}
                    placeholder={
                      serials.isPending
                        ? "Loading serials…"
                        : "Select serial number"
                    }
                    disabled={serials.isPending || serials.isError}
                    onValueChange={setSerialId}
                  />
                </FormField>
              </>
            ) : null}
            <FormField
              label="Movement reference"
              htmlFor="internal-move-reference"
              required
            >
              <Input
                id="internal-move-reference"
                maxLength={140}
                value={reference}
                onChange={(event) => setReference(event.target.value)}
              />
            </FormField>
            <FormField label="Reason" htmlFor="internal-move-reason">
              <Select
                id="internal-move-reason"
                ariaLabel="Internal movement reason"
                className="mt-2"
                value={effectiveReasonCode}
                options={internalMoveReasons.map((reason) => ({
                  value: reason.code,
                  label: `${reason.name} (${reason.code})`,
                }))}
                placeholder={
                  reasons.isPending ? "Loading reasons…" : "Optional reason"
                }
                disabled={reasons.isPending || reasons.isError}
                onValueChange={setReasonCode}
              />
            </FormField>
            <div className="sm:col-span-2">
              <FormField
                label="Notes"
                htmlFor="internal-move-notes"
                required={selectedReason?.requires_note}
              >
                <Textarea
                  id="internal-move-notes"
                  maxLength={4000}
                  placeholder="Why is this stock being moved?"
                  value={notes}
                  onChange={(event) => setNotes(event.target.value)}
                />
              </FormField>
            </div>
          </fieldset>

          <div className="flex flex-col-reverse gap-3 border-t border-slate-200 pt-4 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="secondary"
              disabled={move.isPending}
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={move.isPending || !targetLocationId}
            >
              {move.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              Post movement
            </Button>
          </div>
        </form>
      )}
    </OperationDialog>
  );
}
