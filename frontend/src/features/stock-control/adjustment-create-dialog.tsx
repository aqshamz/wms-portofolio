"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Decimal from "decimal.js";
import { LoaderCircle, Plus, Trash2 } from "lucide-react";
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
  listSerials,
  listSerialStates,
  serialKeys,
  serialStateKeys,
} from "@/features/inventory/inventory-api";
import type { Balance } from "@/features/inventory/inventory-types";
import type { AdjustmentDirection } from "./stock-control-types";
import {
  createAdjustment,
  listStockControlReasons,
  stockControlKeys,
} from "./stock-control-api";

const quantityPattern = /^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$/;
const adjustmentReasonCodes = new Set([
  "DAMAGE",
  "EXPIRY",
  "MANUAL_ADJUSTMENT",
  "TRANSFER_VARIANCE",
]);
type DraftLine = {
  key: string;
  balance: Balance;
  quantity: string;
  serialId?: string;
  serialNumber?: string;
};
function currentBusinessDate(timezone: string) {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}

export function AdjustmentCreateDialog({
  ownerId,
  warehouseId,
  timezone,
  onOpenChange,
}: {
  ownerId: string;
  warehouseId: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [balanceSearch, setBalanceSearch] = useState("");
  const [balanceId, setBalanceId] = useState("");
  const [direction, setDirection] = useState<AdjustmentDirection>("DECREASE");
  const [quantity, setQuantity] = useState("");
  const [date, setDate] = useState(() => currentBusinessDate(timezone));
  const [serialSearch, setSerialSearch] = useState("");
  const [serialId, setSerialId] = useState("");
  const [reasonCode, setReasonCode] = useState("MANUAL_ADJUSTMENT");
  const [notes, setNotes] = useState("");
  const [lines, setLines] = useState<DraftLine[]>([]);
  const [formError, setFormError] = useState("");
  const filters = useMemo(
    () => ({
      ownerId,
      warehouseId,
      search: balanceSearch,
      includeZero: true,
      locationTypeCode: "STORAGE",
      page: 1,
      pageSize: 100,
    }),
    [balanceSearch, ownerId, warehouseId],
  );
  const balances = useQuery({
    queryKey: balanceKeys.list(filters),
    queryFn: () => listBalances(filters),
  });
  const selected = balances.data?.items.find(
    (balance) => balance.balance_id === balanceId,
  );
  const reasons = useQuery({
    queryKey: stockControlKeys.reasons(),
    queryFn: listStockControlReasons,
  });
  const adjustmentReasons = (reasons.data ?? []).filter((reason) =>
    adjustmentReasonCodes.has(reason.code),
  );
  const selectedReason = adjustmentReasons.find(
    (reason) => reason.code === reasonCode,
  );
  const serialStateFilters = useMemo(
    () => ({
      ownerId,
      warehouseId,
      locationId: selected?.location_id,
      itemId: selected?.item_id,
      lotId: selected?.lot_id ?? undefined,
      handlingUnitId: selected?.handling_unit_id ?? undefined,
      inventoryStatusId: selected?.inventory_status_id,
      search: serialSearch,
      page: 1,
      pageSize: 100,
    }),
    [ownerId, selected, serialSearch, warehouseId],
  );
  const serialStates = useQuery({
    queryKey: serialStateKeys.list(serialStateFilters),
    queryFn: () => listSerialStates(serialStateFilters),
    enabled: Boolean(selected?.serial_controlled && direction === "DECREASE"),
  });
  const serialFilters = useMemo(
    () => ({
      ownerId,
      itemId: selected?.item_id,
      search: serialSearch,
      page: 1,
      pageSize: 100,
    }),
    [ownerId, selected?.item_id, serialSearch],
  );
  const serials = useQuery({
    queryKey: serialKeys.list(serialFilters),
    queryFn: () => listSerials(serialFilters),
    enabled: Boolean(selected?.serial_controlled && direction === "INCREASE"),
  });
  const serialOptions =
    direction === "DECREASE"
      ? (serialStates.data?.items ?? []).map((row) => ({
          value: row.serial_id,
          label: row.serial_no,
        }))
      : (serials.data?.items ?? []).map((row) => ({
          value: row.serial_id,
          label: row.serial_no,
        }));

  function addLine() {
    setFormError("");
    if (!selected) return setFormError("Select an inventory balance.");
    const value = selected.serial_controlled ? "1" : quantity;
    if (!quantityPattern.test(value) || new Decimal(value).lte(0))
      return setFormError("Quantity must be positive with up to 6 decimals.");
    if (
      direction === "DECREASE" &&
      new Decimal(value).gt(new Decimal(selected.available_qty))
    )
      return setFormError(
        `Decrease cannot exceed ${selected.available_qty} ${selected.uom_code} available.`,
      );
    if (selected.serial_controlled && !serialId)
      return setFormError("Select a serial number.");
    const key = selected.balance_id;
    if (lines.some((line) => line.balance.balance_id === selected.balance_id))
      return setFormError("That inventory balance is already added.");
    setLines((current) => [
      ...current,
      {
        key,
        balance: selected,
        quantity: value,
        serialId: serialId || undefined,
        serialNumber: serialOptions.find((option) => option.value === serialId)
          ?.label,
      },
    ]);
    setBalanceId("");
    setQuantity("");
    setSerialId("");
    setSerialSearch("");
  }

  const create = useMutation({
    mutationFn: async () => {
      if (!date) throw new Error("Business date is required.");
      if (!lines.length) throw new Error("Add at least one adjustment line.");
      if (!reasonCode) throw new Error("Select an adjustment reason.");
      if (selectedReason?.requires_note && !notes.trim())
        throw new Error(`Notes are required for ${selectedReason.name}.`);
      return createAdjustment({
        business_date: date,
        direction,
        reason_code: reasonCode,
        notes: notes.trim() || undefined,
        lines: lines.map((line) => ({
          balance_id: line.balance.balance_id,
          quantity: line.quantity,
          serial_id: line.serialId,
          expected_balance_version: line.balance.version_no,
        })),
      });
    },
    onSuccess: (adjustment) => {
      toast.success(
        `${adjustment.inventory_adjustment_id} created with ${adjustment.total_lines} lines. Stock was not changed.`,
      );
      void queryClient.invalidateQueries({
        queryKey: stockControlKeys.adjustments(),
      });
      onOpenChange(false);
    },
    onError: (error) => toast.error(error.message),
  });
  const error =
    balances.error ??
    reasons.error ??
    serialStates.error ??
    serials.error ??
    create.error;

  return (
    <OperationDialog
      title="Request inventory adjustment"
      description="Create one controlled document for several inventory balances."
      busy={create.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close adjustment request"
    >
      <form
        className="space-y-5"
        onSubmit={(event) => {
          event.preventDefault();
          create.mutate();
        }}
      >
        <section className="rounded-xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950">
          Direction, business date, reason, and notes apply to every line.
          Reviewers may approve or reject selected lines independently.
        </section>
        {error || formError ? (
          <p
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {error?.message ?? formError}
          </p>
        ) : null}
        <fieldset
          disabled={create.isPending}
          className="grid gap-4 sm:grid-cols-2"
        >
          <FormField label="Direction" htmlFor="adjustment-direction" required>
            <Select
              id="adjustment-direction"
              ariaLabel="Adjustment direction"
              className="mt-2"
              value={direction}
              options={[
                { value: "DECREASE", label: "Decrease on-hand stock" },
                { value: "INCREASE", label: "Increase on-hand stock" },
              ]}
              onValueChange={(value) => {
                setDirection(value as AdjustmentDirection);
                setLines([]);
                setBalanceId("");
              }}
            />
          </FormField>
          <FormField label="Business date" htmlFor="adjustment-date" required>
            <Input
              id="adjustment-date"
              type="date"
              value={date}
              onChange={(event) => setDate(event.target.value)}
            />
          </FormField>
          <FormField label="Reason" htmlFor="adjustment-reason" required>
            <Select
              id="adjustment-reason"
              ariaLabel="Adjustment reason"
              className="mt-2"
              value={reasonCode}
              options={adjustmentReasons.map((reason) => ({
                value: reason.code,
                label: `${reason.name} (${reason.code})`,
              }))}
              placeholder="Select reason"
              onValueChange={setReasonCode}
            />
          </FormField>
          <FormField
            label="Document notes"
            htmlFor="adjustment-notes"
            required={selectedReason?.requires_note}
          >
            <Textarea
              id="adjustment-notes"
              maxLength={4000}
              value={notes}
              onChange={(event) => setNotes(event.target.value)}
            />
          </FormField>
        </fieldset>

        <section className="rounded-xl border border-slate-200 p-4">
          <h3 className="font-semibold">Add inventory line</h3>
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <FormField label="Search inventory" htmlFor="adjustment-search">
              <Input
                id="adjustment-search"
                placeholder="Item, location or lot"
                value={balanceSearch}
                onChange={(event) => setBalanceSearch(event.target.value)}
              />
            </FormField>
            <FormField
              label="Inventory balance"
              htmlFor="adjustment-balance"
              required
            >
              <Select
                id="adjustment-balance"
                ariaLabel="Inventory balance to adjust"
                className="mt-2"
                value={balanceId}
                options={(balances.data?.items ?? []).map((balance) => ({
                  value: balance.balance_id,
                  label: `${balance.item_code} · ${balance.location_code} · ${balance.inventory_status_code} · ${balance.on_hand_qty} ${balance.uom_code}`,
                }))}
                placeholder="Select balance"
                onValueChange={(value) => {
                  setBalanceId(value);
                  setQuantity("");
                  setSerialId("");
                }}
              />
            </FormField>
            <FormField
              label="Quantity (base units)"
              htmlFor="adjustment-quantity"
              required
            >
              <Input
                id="adjustment-quantity"
                inputMode="decimal"
                value={selected?.serial_controlled ? "1" : quantity}
                disabled={Boolean(selected?.serial_controlled)}
                onChange={(event) => setQuantity(event.target.value)}
              />
            </FormField>
            {selected?.serial_controlled ? (
              <FormField
                label="Serial number"
                htmlFor="adjustment-serial"
                required
              >
                <Input
                  className="mb-2"
                  aria-label="Search serial numbers"
                  placeholder="Search serial"
                  value={serialSearch}
                  onChange={(event) => setSerialSearch(event.target.value)}
                />
                <Select
                  id="adjustment-serial"
                  ariaLabel="Serial number to adjust"
                  value={serialId}
                  options={serialOptions}
                  placeholder="Select serial"
                  onValueChange={setSerialId}
                />
              </FormField>
            ) : null}
          </div>
          <div className="mt-4 flex justify-end">
            <Button type="button" variant="secondary" onClick={addLine}>
              <Plus className="size-4" /> Add line
            </Button>
          </div>
        </section>

        <section className="overflow-hidden rounded-xl border border-slate-200">
          <div className="border-b border-slate-200 bg-slate-50 px-4 py-3">
            <p className="font-semibold">Document lines ({lines.length})</p>
          </div>
          {!lines.length ? (
            <p className="p-4 text-sm text-slate-500">No balances added yet.</p>
          ) : (
            <div className="divide-y divide-slate-200">
              {lines.map((line, index) => (
                <div
                  key={line.key}
                  className="flex items-start justify-between gap-3 p-4 text-sm"
                >
                  <div>
                    <p className="font-semibold">
                      {index + 1}. {line.balance.item_code} · {line.quantity}{" "}
                      {line.balance.uom_code}
                    </p>
                    <p className="mt-1 text-slate-600">
                      {line.balance.location_code} ·{" "}
                      {line.balance.inventory_status_code}
                      {line.serialNumber ? ` · ${line.serialNumber}` : ""}
                    </p>
                    <p className="mt-1 text-xs text-slate-500">
                      Balance version {line.balance.version_no}
                    </p>
                  </div>
                  <Button
                    type="button"
                    size="sm"
                    variant="ghost"
                    aria-label={`Remove line ${index + 1}`}
                    onClick={() =>
                      setLines((current) =>
                        current.filter((entry) => entry.key !== line.key),
                      )
                    }
                  >
                    <Trash2 className="size-4" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </section>
        <div className="flex flex-col-reverse gap-3 border-t border-slate-200 pt-4 sm:flex-row sm:justify-end">
          <Button
            type="button"
            variant="secondary"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={create.isPending || !lines.length}>
            {create.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            Submit {lines.length} line{lines.length === 1 ? "" : "s"}
          </Button>
        </div>
      </form>
    </OperationDialog>
  );
}
