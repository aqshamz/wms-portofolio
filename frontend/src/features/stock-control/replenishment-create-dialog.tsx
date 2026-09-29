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
  listBalances,
  listSerialStates,
  serialStateKeys,
} from "@/features/inventory/inventory-api";
import {
  createReplenishment,
  listReplenishmentTargets,
  stockControlKeys,
} from "./stock-control-api";

const quantityPattern = /^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$/;

export function ReplenishmentCreateDialog({
  ownerId,
  warehouseId,
  onOpenChange,
}: {
  ownerId: string;
  warehouseId: string;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [sourceSearch, setSourceSearch] = useState("");
  const [sourceBalanceId, setSourceBalanceId] = useState("");
  const [targetSearch, setTargetSearch] = useState("");
  const [targetLocationId, setTargetLocationId] = useState("");
  const [serialId, setSerialId] = useState("");
  const [quantity, setQuantity] = useState("");
  const [priority, setPriority] = useState("NORMAL");
  const [notes, setNotes] = useState("");
  const balanceFilters = useMemo(
    () => ({
      ownerId,
      warehouseId,
      search: sourceSearch,
      includeZero: false,
      page: 1,
      pageSize: 100,
    }),
    [ownerId, sourceSearch, warehouseId],
  );
  const balances = useQuery({
    queryKey: balanceKeys.list(balanceFilters),
    queryFn: () => listBalances(balanceFilters),
  });
  const targets = useQuery({
    queryKey: stockControlKeys.replenishmentTargets(
      ownerId,
      warehouseId,
      targetSearch,
    ),
    queryFn: () => listReplenishmentTargets(ownerId, warehouseId, targetSearch),
  });
  const source = balances.data?.items.find(
    (row) => row.balance_id === sourceBalanceId,
  );
  const serialFilters = useMemo(
    () => ({
      ownerId,
      warehouseId,
      locationId: source?.location_id,
      itemId: source?.item_id,
      lotId: source?.lot_id ?? undefined,
      handlingUnitId: source?.handling_unit_id ?? undefined,
      inventoryStatusId: source?.inventory_status_id,
      search: "",
      page: 1,
      pageSize: 100,
    }),
    [ownerId, source, warehouseId],
  );
  const serials = useQuery({
    queryKey: serialStateKeys.list(serialFilters),
    queryFn: () => listSerialStates(serialFilters),
    enabled: Boolean(source?.serial_controlled),
  });
  const sourceOptions = (balances.data?.items ?? []).filter((balance) => {
    return (
      new Decimal(balance.available_qty).gt(0) &&
      balance.location_allows_storage &&
      !balance.location_is_pick_face &&
      !balance.location_is_locked &&
      balance.inventory_status_is_allocatable &&
      balance.inventory_status_is_pickable
    );
  });
  const targetOptions = (targets.data?.items ?? []).filter(
    (location) => location.location_id !== source?.location_id,
  );
  const effectiveQuantity = source?.serial_controlled
    ? "1"
    : source?.handling_unit_id
      ? source.available_qty
      : quantity;

  const create = useMutation({
    mutationFn: () => {
      if (!source) throw new Error("Select reserve stock to replenish.");
      if (!targetLocationId) throw new Error("Select a pick-face target.");
      if (!quantityPattern.test(effectiveQuantity))
        throw new Error("Enter a positive quantity with up to 6 decimals.");
      const requested = new Decimal(effectiveQuantity);
      if (requested.lte(0) || requested.gt(source.available_qty))
        throw new Error(
          `Quantity cannot exceed ${source.available_qty} ${source.uom_code}.`,
        );
      if (source.serial_controlled && !serialId)
        throw new Error("Select the serial number to replenish.");
      return createReplenishment({
        source_balance_id: source.balance_id,
        target_location_id: targetLocationId,
        serial_id: serialId || undefined,
        quantity: effectiveQuantity,
        priority_code: priority,
        notes: notes.trim() || undefined,
        expected_balance_version: source.version_no,
      });
    },
    onSuccess: (task) => {
      toast.success(`Replenishment ${task.replenishment_task_id} created.`);
      void queryClient.invalidateQueries({ queryKey: stockControlKeys.all });
      void queryClient.invalidateQueries({
        queryKey: ["inventory", "balances"],
      });
      onOpenChange(false);
    },
    onError: (error) => toast.error(error.message),
  });
  const error =
    balances.error ?? targets.error ?? serials.error ?? create.error;

  return (
    <OperationDialog
      title="Create replenishment"
      description="Reserve stock in storage and plan its movement into a forward pick face."
      busy={create.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close replenishment creation"
    >
      <form
        className="space-y-5"
        onSubmit={(event) => {
          event.preventDefault();
          create.mutate();
        }}
      >
        <p className="rounded-xl border border-cyan-200 bg-cyan-50 p-3 text-sm text-cyan-950">
          Creating the task reserves its planned quantity. Inventory stays at
          the source until the assigned worker completes the task.
        </p>
        {error ? (
          <p
            role="alert"
            className="rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
          >
            {error.message}
          </p>
        ) : null}
        <fieldset
          disabled={create.isPending}
          className="grid gap-4 sm:grid-cols-2"
        >
          <FormField
            label="Search reserve stock"
            htmlFor="replenishment-source-search"
          >
            <Input
              id="replenishment-source-search"
              maxLength={160}
              value={sourceSearch}
              placeholder="Item, location or lot"
              onChange={(event) => setSourceSearch(event.target.value)}
            />
          </FormField>
          <FormField
            label="Source balance"
            htmlFor="replenishment-source"
            required
          >
            <Select
              id="replenishment-source"
              ariaLabel="Replenishment source balance"
              className="mt-2"
              value={sourceBalanceId}
              options={sourceOptions.map((balance) => ({
                value: balance.balance_id,
                label: `${balance.item_code} · ${balance.location_code} · ${balance.available_qty} ${balance.uom_code}${balance.lot_number ? ` · ${balance.lot_number}` : ""}${balance.handling_unit_barcode ? ` · ${balance.handling_unit_barcode}` : ""}`,
              }))}
              placeholder={
                balances.isPending ? "Loading stock…" : "Select reserve stock"
              }
              disabled={balances.isPending}
              onValueChange={(value) => {
                setSourceBalanceId(value);
                setTargetLocationId("");
                setSerialId("");
                setQuantity("");
              }}
            />
          </FormField>
          <FormField
            label="Search pick faces"
            htmlFor="replenishment-target-search"
          >
            <Input
              id="replenishment-target-search"
              maxLength={160}
              value={targetSearch}
              placeholder="Location or zone"
              onChange={(event) => setTargetSearch(event.target.value)}
            />
          </FormField>
          <FormField
            label="Target pick face"
            htmlFor="replenishment-target"
            required
          >
            <Select
              id="replenishment-target"
              ariaLabel="Replenishment target pick face"
              className="mt-2"
              value={targetLocationId}
              options={targetOptions.map((location) => ({
                value: location.location_id,
                label: `${location.code}${location.zone_code ? ` · ${location.zone_code}` : ""} · ${location.location_type_code}`,
              }))}
              placeholder={
                targets.isPending ? "Loading pick faces…" : "Select pick face"
              }
              disabled={!source || targets.isPending}
              onValueChange={setTargetLocationId}
            />
          </FormField>
          <FormField
            label="Quantity (base units)"
            htmlFor="replenishment-quantity"
            required
          >
            <Input
              id="replenishment-quantity"
              inputMode="decimal"
              value={effectiveQuantity}
              placeholder={
                source
                  ? `Maximum ${source.available_qty}`
                  : "Select source first"
              }
              disabled={
                !source ||
                Boolean(source.serial_controlled || source.handling_unit_id)
              }
              onChange={(event) => setQuantity(event.target.value)}
            />
          </FormField>
          <FormField label="Priority" htmlFor="replenishment-priority" required>
            <Select
              id="replenishment-priority"
              ariaLabel="Replenishment priority"
              className="mt-2"
              value={priority}
              options={["LOW", "NORMAL", "HIGH", "URGENT"].map((code) => ({
                value: code,
                label: code,
              }))}
              onValueChange={setPriority}
            />
          </FormField>
          {source?.serial_controlled ? (
            <div className="sm:col-span-2">
              <FormField
                label="Serial number"
                htmlFor="replenishment-serial"
                required
              >
                <Select
                  id="replenishment-serial"
                  ariaLabel="Replenishment serial number"
                  className="mt-2"
                  value={serialId}
                  options={(serials.data?.items ?? []).map((serial) => ({
                    value: serial.serial_id,
                    label: serial.serial_no,
                  }))}
                  placeholder={
                    serials.isPending ? "Loading serials…" : "Select serial"
                  }
                  disabled={serials.isPending}
                  onValueChange={setSerialId}
                />
              </FormField>
            </div>
          ) : null}
          <div className="sm:col-span-2">
            <FormField
              label="Instructions / notes"
              htmlFor="replenishment-notes"
            >
              <Textarea
                id="replenishment-notes"
                maxLength={4000}
                value={notes}
                placeholder="Optional handling instructions"
                onChange={(event) => setNotes(event.target.value)}
              />
            </FormField>
          </div>
        </fieldset>
        <div className="flex flex-col-reverse gap-3 border-t border-slate-200 pt-4 sm:flex-row sm:justify-end">
          <Button
            type="button"
            variant="secondary"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={create.isPending || !source || !targetLocationId}
          >
            {create.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            Create task
          </Button>
        </div>
      </form>
    </OperationDialog>
  );
}
