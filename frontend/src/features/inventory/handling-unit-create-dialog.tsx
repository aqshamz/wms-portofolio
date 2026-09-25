"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { Select } from "@/components/ui/select";
import {
  listLocations,
  storageLayoutKeys,
} from "@/features/storage-layout/storage-layout-api";
import {
  listHandlingUnitTypes,
  unitsPackagingKeys,
} from "@/features/units-packaging/units-packaging-api";
import { createHandlingUnit } from "./inventory-api";
import type { HandlingUnit } from "./inventory-types";

const activeTypes = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

export function HandlingUnitCreateDialog({
  ownerId,
  warehouseId,
  defaultLocationId = "",
  onCreated,
  onOpenChange,
}: {
  ownerId: string;
  warehouseId: string;
  defaultLocationId?: string;
  onCreated: (unit: HandlingUnit) => void;
  onOpenChange: (open: boolean) => void;
}) {
  const [barcode, setBarcode] = useState("");
  const [typeId, setTypeId] = useState("");
  const [locationId, setLocationId] = useState(defaultLocationId);
  const types = useQuery({
    queryKey: unitsPackagingKeys.handlingList(activeTypes),
    queryFn: () => listHandlingUnitTypes(activeTypes),
  });
  const locationFilters = useMemo(
    () => ({
      warehouseId,
      search: "",
      active: "active" as const,
      page: 1,
      pageSize: 100,
    }),
    [warehouseId],
  );
  const locations = useQuery({
    queryKey: storageLayoutKeys.locations(locationFilters),
    queryFn: () => listLocations(locationFilters),
  });
  const usableLocations = (locations.data?.items ?? []).filter(
    (location) => !location.is_locked,
  );
  const create = useMutation({
    mutationFn: () => {
      if (!barcode.trim()) throw new Error("Enter or scan a barcode.");
      if (!typeId) throw new Error("Select a handling-unit type.");
      if (!locationId) throw new Error("Select the current location.");
      return createHandlingUnit({
        owner_id: ownerId,
        warehouse_id: warehouseId,
        handling_unit_type_id: typeId,
        current_location_id: locationId,
        barcode: barcode.trim(),
      });
    },
    onSuccess: (unit) => {
      toast.success(`Handling unit ${unit.barcode} created.`);
      onCreated(unit);
      onOpenChange(false);
    },
  });
  const error = create.error ?? types.error ?? locations.error;

  return (
    <OperationDialog
      title="Create handling unit"
      description="Register an empty, scannable container at its current warehouse location."
      busy={create.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close handling-unit form"
    >
      <form
        className="space-y-5"
        onSubmit={(event) => {
          event.preventDefault();
          create.mutate();
        }}
      >
        <div className="rounded-xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950">
          Creating a handling unit registers an empty container only. Stock is
          attached later by a receipt or another controlled inventory workflow.
        </div>
        {error ? (
          <div
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {error.message}
          </div>
        ) : null}
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Barcode" htmlFor="handling-unit-barcode" required>
            <Input
              id="handling-unit-barcode"
              autoFocus
              maxLength={120}
              placeholder="Scan or enter pallet/carton barcode"
              value={barcode}
              onChange={(event) => setBarcode(event.target.value)}
            />
          </FormField>
          <FormField
            label="Handling-unit type"
            htmlFor="handling-unit-type"
            required
          >
            <Select
              id="handling-unit-type"
              ariaLabel="Handling-unit type"
              value={typeId}
              options={(types.data?.items ?? []).map((type) => ({
                value: type.handling_unit_type_id,
                label: `${type.code} · ${type.name}`,
              }))}
              placeholder={types.isPending ? "Loading types…" : "Select type"}
              disabled={types.isPending}
              onValueChange={setTypeId}
            />
          </FormField>
          <FormField
            label="Current location"
            htmlFor="handling-unit-location"
            required
          >
            <Select
              id="handling-unit-location"
              ariaLabel="Handling-unit current location"
              value={locationId}
              options={usableLocations.map((location) => ({
                value: location.location_id,
                label: `${location.code}${location.zone_code ? ` · ${location.zone_code}` : ""}`,
              }))}
              placeholder={
                locations.isPending ? "Loading locations…" : "Select location"
              }
              disabled={locations.isPending}
              onValueChange={setLocationId}
            />
          </FormField>
        </div>
        <p className="text-sm text-slate-600">
          This first operational foundation creates root handling units. Nested
          carton/pallet composition will be managed by a later pack/unpack
          workflow.
        </p>
        <div className="flex justify-end gap-3 border-t border-slate-200 pt-4">
          <Button
            type="button"
            variant="secondary"
            disabled={create.isPending}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={create.isPending}>
            {create.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            Create handling unit
          </Button>
        </div>
      </form>
    </OperationDialog>
  );
}
