"use client";

import { useState } from "react";
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
  serialStateKeys,
} from "@/features/inventory/inventory-api";
import {
  listWarehouses,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import { createWarehouseTransfer, stockControlKeys } from "./stock-control-api";

const quantityPattern = /^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$/;
function businessDate(timezone: string) {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}

export function WarehouseTransferDialog({
  balanceId,
  timezone,
  onOpenChange,
}: {
  balanceId: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [targetWarehouseId, setTargetWarehouseId] = useState("");
  const [serialSearch, setSerialSearch] = useState("");
  const [serialId, setSerialId] = useState("");
  const [quantity, setQuantity] = useState("");
  const [date, setDate] = useState(() => businessDate(timezone));
  const [notes, setNotes] = useState("");
  const balance = useQuery({
    queryKey: balanceKeys.detail(balanceId),
    queryFn: () => getBalance(balanceId),
  });
  const source = balance.data;
  const warehouseFilters = {
    search: "",
    active: "active" as const,
    ownerId: source?.owner_id,
    page: 1,
    pageSize: 100,
  };
  const warehouses = useQuery({
    queryKey: warehouseKeys.list(warehouseFilters),
    queryFn: () => listWarehouses(warehouseFilters),
    enabled: Boolean(source?.owner_id),
  });
  const serialFilters = {
    ownerId: source?.owner_id ?? "",
    warehouseId: source?.warehouse_id ?? "",
    locationId: source?.location_id,
    itemId: source?.item_id,
    lotId: source?.lot_id ?? undefined,
    inventoryStatusId: source?.inventory_status_id,
    search: serialSearch,
    page: 1,
    pageSize: 100,
  };
  const serials = useQuery({
    queryKey: serialStateKeys.list(serialFilters),
    queryFn: () => listSerialStates(serialFilters),
    enabled: Boolean(source?.serial_controlled),
  });
  const targets = (warehouses.data?.items ?? []).filter(
    (entry) => entry.warehouse_id !== source?.warehouse_id,
  );
  const effectiveQuantity = source?.serial_controlled ? "1" : quantity;
  const create = useMutation({
    mutationFn: async () => {
      if (!source) throw new Error("Source balance is unavailable.");
      if (source.handling_unit_id)
        throw new Error("Handling-unit transfer is not supported yet.");
      if (!targetWarehouseId) throw new Error("Select a target warehouse.");
      if (!date) throw new Error("Business date is required.");
      if (
        !quantityPattern.test(effectiveQuantity) ||
        new Decimal(effectiveQuantity).lte(0) ||
        new Decimal(effectiveQuantity).gt(new Decimal(source.available_qty))
      )
        throw new Error(
          `Quantity must be positive and cannot exceed ${source.available_qty} ${source.uom_code}.`,
        );
      if (source.serial_controlled && !serialId)
        throw new Error("Select the serial number to transfer.");
      return createWarehouseTransfer({
        business_date: date,
        source_balance_id: source.balance_id,
        target_warehouse_id: targetWarehouseId,
        quantity: effectiveQuantity,
        serial_ids: serialId ? [serialId] : [],
        notes: notes.trim() || undefined,
      });
    },
    onSuccess: (document) => {
      toast.success(
        `Transfer ${document.warehouse_transfer_id} created as draft.`,
      );
      void queryClient.invalidateQueries({
        queryKey: stockControlKeys.warehouseTransfers(),
      });
      onOpenChange(false);
    },
    onError: (error) => toast.error(error.message),
  });
  const error =
    balance.error ?? warehouses.error ?? serials.error ?? create.error;
  return (
    <OperationDialog
      title="Create warehouse transfer"
      description="Create a dispatch-and-receipt document. Stock remains at the source until dispatch."
      busy={create.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close warehouse transfer"
    >
      {balance.isPending ? (
        <div className="flex min-h-48 items-center justify-center gap-2 text-sm text-slate-600">
          <LoaderCircle className="size-5 animate-spin" /> Loading source stock…
        </div>
      ) : !source ? (
        <p className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
          Source balance could not be loaded.
        </p>
      ) : (
        <form
          className="space-y-5"
          onSubmit={(event) => {
            event.preventDefault();
            create.mutate();
          }}
        >
          <section className="rounded-xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950">
            The transfer starts in Draft. Approve and dispatch it from Stock
            Control; the destination warehouse then receives it into a receiving
            location and puts it away.
          </section>
          <section className="grid gap-3 rounded-xl bg-slate-50 p-4 text-sm sm:grid-cols-3">
            <Info
              label="Item"
              value={source.item_code}
              detail={source.item_name}
            />
            <Info
              label="Source"
              value={source.location_code}
              detail={source.inventory_status_code}
            />
            <Info
              label="Available"
              value={`${source.available_qty} ${source.uom_code}`}
              detail={source.lot_number ?? "No lot"}
            />
          </section>
          {error ? (
            <p
              role="alert"
              className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
            >
              {error.message}
            </p>
          ) : null}
          <fieldset
            disabled={create.isPending || Boolean(source.handling_unit_id)}
            className="grid gap-4 sm:grid-cols-2"
          >
            <FormField
              label="Target warehouse"
              htmlFor="transfer-warehouse"
              required
            >
              <Select
                id="transfer-warehouse"
                ariaLabel="Warehouse transfer target warehouse"
                className="mt-2"
                value={targetWarehouseId}
                options={targets.map((entry) => ({
                  value: entry.warehouse_id,
                  label: `${entry.name} (${entry.code})`,
                }))}
                placeholder={
                  warehouses.isPending
                    ? "Loading warehouses…"
                    : targets.length
                      ? "Select served warehouse"
                      : "No other served warehouse"
                }
                onValueChange={setTargetWarehouseId}
              />
            </FormField>
            <FormField label="Business date" htmlFor="transfer-date" required>
              <Input
                id="transfer-date"
                type="date"
                value={date}
                onChange={(event) => setDate(event.target.value)}
              />
            </FormField>
            <FormField
              label="Quantity (base units)"
              htmlFor="transfer-quantity"
              required
            >
              <Input
                id="transfer-quantity"
                inputMode="decimal"
                value={effectiveQuantity}
                disabled={source.serial_controlled}
                placeholder={`Maximum ${source.available_qty}`}
                onChange={(event) => setQuantity(event.target.value)}
              />
            </FormField>
            {source.serial_controlled ? (
              <>
                <FormField
                  label="Search serial"
                  htmlFor="transfer-serial-search"
                >
                  <Input
                    id="transfer-serial-search"
                    value={serialSearch}
                    onChange={(event) => setSerialSearch(event.target.value)}
                  />
                </FormField>
                <FormField
                  label="Serial number"
                  htmlFor="transfer-serial"
                  required
                >
                  <Select
                    id="transfer-serial"
                    ariaLabel="Serial number to transfer"
                    className="mt-2"
                    value={serialId}
                    options={(serials.data?.items ?? []).map((entry) => ({
                      value: entry.serial_id,
                      label: entry.serial_no,
                    }))}
                    placeholder="Select serial number"
                    onValueChange={setSerialId}
                  />
                </FormField>
              </>
            ) : null}
            <div className="sm:col-span-2">
              <FormField label="Notes" htmlFor="transfer-notes">
                <Textarea
                  id="transfer-notes"
                  maxLength={4000}
                  value={notes}
                  onChange={(event) => setNotes(event.target.value)}
                />
              </FormField>
            </div>
          </fieldset>
          <div className="flex justify-end gap-3 border-t border-slate-200 pt-4">
            <Button
              type="button"
              variant="secondary"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={
                create.isPending ||
                !targetWarehouseId ||
                Boolean(source.handling_unit_id)
              }
            >
              {create.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              Create draft
            </Button>
          </div>
        </form>
      )}
    </OperationDialog>
  );
}
function Info({
  label,
  value,
  detail,
}: {
  label: string;
  value: string;
  detail: string;
}) {
  return (
    <div>
      <p className="text-xs font-semibold text-slate-500 uppercase">{label}</p>
      <p className="mt-1 font-semibold">{value}</p>
      <p className="text-xs text-slate-500">{detail}</p>
    </div>
  );
}
