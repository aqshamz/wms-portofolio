"use client";

import { useRef, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
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
  inventoryClassificationKeys,
  listInventoryStatuses,
} from "@/features/inventory-classifications/inventory-classification-api";
import {
  listStockControlReasons,
  postInventoryStatusChange,
  stockControlKeys,
} from "./stock-control-api";
import type { InventoryStatusChangeRequest } from "./stock-control-types";
import {
  isManualStatus,
  statusChangeValidation,
} from "./inventory-status-change-validation";

const activeStatuses = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

function businessDate(timezone: string) {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}

export function InventoryStatusChangeDialog({
  balanceId,
  timezone,
  onOpenChange,
}: {
  balanceId: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const client = useQueryClient();
  const [targetId, setTargetId] = useState("");
  const [quantity, setQuantity] = useState("");
  const [serialId, setSerialId] = useState("");
  const [serialSearch, setSerialSearch] = useState("");
  const [date, setDate] = useState(() => businessDate(timezone));
  const [reference, setReference] = useState(
    () => `ISC-${businessDate(timezone).replaceAll("-", "")}`,
  );
  const [reasonCode, setReasonCode] = useState("");
  const [notes, setNotes] = useState("");
  const [validationError, setValidationError] = useState<string>();
  const [review, setReview] = useState<InventoryStatusChangeRequest | null>(
    null,
  );
  const retryKey = useRef<{ fingerprint: string; key: string } | null>(null);

  const balance = useQuery({
    queryKey: balanceKeys.detail(balanceId),
    queryFn: () => getBalance(balanceId),
  });
  const source = balance.data;
  const statuses = useQuery({
    queryKey: inventoryClassificationKeys.list(activeStatuses),
    queryFn: () => listInventoryStatuses(activeStatuses),
  });
  const reasons = useQuery({
    queryKey: stockControlKeys.reasons(),
    queryFn: listStockControlReasons,
  });
  const targets = (statuses.data?.items ?? []).filter(
    (status) =>
      status.is_active &&
      isManualStatus(status.code) &&
      status.inventory_status_id !== source?.inventory_status_id,
  );
  const target = targets.find(
    (status) => status.inventory_status_id === targetId,
  );
  const reason = reasons.data?.find((entry) => entry.code === reasonCode);
  const serialFilters = {
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
  };
  const serials = useQuery({
    queryKey: serialStateKeys.list(serialFilters),
    queryFn: () => listSerialStates(serialFilters),
    enabled: Boolean(source?.serial_controlled),
  });
  const matchingSerials = (serials.data?.items ?? []).filter(
    (entry) => entry.balance.balance_id === balanceId,
  );
  const effectiveQuantity = source?.serial_controlled ? "1" : quantity.trim();
  const post = useMutation({
    mutationFn: (request: InventoryStatusChangeRequest) =>
      postInventoryStatusChange(request),
    retry: false,
    onSuccess: (result) => {
      if (result.source_balance)
        client.setQueryData(
          balanceKeys.detail(result.source_balance.balance_id),
          result.source_balance,
        );
      if (result.destination_balance)
        client.setQueryData(
          balanceKeys.detail(result.destination_balance.balance_id),
          result.destination_balance,
        );
      void client.invalidateQueries({ queryKey: ["inventory", "balances"] });
      void client.invalidateQueries({ queryKey: ["inventory", "movements"] });
      void client.invalidateQueries({
        queryKey: ["inventory", "serial-states"],
      });
      toast.success(
        `Inventory status changed. Movement ${result.movements[0]?.movement_id ?? "recorded"}.`,
      );
      onOpenChange(false);
    },
    onError: (error) => toast.error(error.message),
  });
  const lookupError =
    balance.error ??
    statuses.error ??
    reasons.error ??
    (source?.serial_controlled ? serials.error : null);
  const errorMessage =
    post.error?.message ?? lookupError?.message ?? validationError;
  const loadingOptions =
    statuses.isFetching ||
    reasons.isFetching ||
    (Boolean(source?.serial_controlled) && serials.isFetching);

  function prepareReview() {
    if (!source) return;
    const error = statusChangeValidation({
      source,
      target,
      reason,
      quantity: effectiveQuantity,
      serial: matchingSerials.find((entry) => entry.serial_id === serialId),
      date,
      reference,
      notes,
    });
    setValidationError(error);
    if (error || loadingOptions || lookupError || !target || !reason) return;
    const command = {
      business_date: date,
      source_document_id: reference.trim(),
      reason_code: reason.code,
      notes: notes.trim() || undefined,
      source_balance_id: source.balance_id,
      target_inventory_status_id: target.inventory_status_id,
      quantity: effectiveQuantity,
      serial_ids: source.serial_controlled ? [serialId] : [],
      expected_version: source.version_no,
    };
    const fingerprint = JSON.stringify(command);
    if (retryKey.current?.fingerprint !== fingerprint)
      retryKey.current = {
        fingerprint,
        key: `stock.status-change.${crypto.randomUUID()}`,
      };
    setReview({ ...command, operation_key: retryKey.current.key });
    post.reset();
  }

  return (
    <OperationDialog
      title="Change inventory status"
      description="Change stock condition without moving its location or changing the total on-hand quantity."
      busy={post.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close inventory status change"
    >
      {balance.isPending ? (
        <div className="flex min-h-48 items-center justify-center gap-2 text-sm">
          <LoaderCircle className="size-5 animate-spin" /> Loading stock…
        </div>
      ) : !source ? (
        <p role="alert" className="rounded-xl bg-rose-50 p-4 text-rose-900">
          {balance.error?.message ?? "Source balance is unavailable."}
        </p>
      ) : (
        <form
          className="space-y-5"
          onSubmit={(event) => {
            event.preventDefault();
            if (!review) prepareReview();
          }}
        >
          <section className="grid gap-3 rounded-xl bg-slate-50 p-4 text-sm sm:grid-cols-3">
            <div>
              <p className="text-xs text-slate-500">Item</p>
              <p className="font-semibold">{source.item_code}</p>
              <p>{source.item_name}</p>
            </div>
            <div>
              <p className="text-xs text-slate-500">
                Location / current status
              </p>
              <p>
                {source.location_code} · {source.inventory_status_code}
              </p>
              <p className="text-xs text-slate-500">
                {source.lot_number ? `Lot ${source.lot_number}` : "No lot"}
                {source.handling_unit_barcode
                  ? ` · HU ${source.handling_unit_barcode}`
                  : ""}
              </p>
            </div>
            <div>
              <p className="text-xs text-slate-500">Unreserved / reserved</p>
              <p className="font-semibold">
                {source.available_qty} / {source.reserved_qty} {source.uom_code}
              </p>
            </div>
          </section>
          <p className="rounded-xl bg-cyan-50 p-3 text-sm text-cyan-950">
            This posts immediately after confirmation. Reserved stock cannot
            change status. QC, putaway, quarantine and return-pending stock must
            use their owning transactions. An active grand opname blocks
            posting.
          </p>
          {errorMessage ? (
            <p
              role="alert"
              className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
            >
              {errorMessage}
            </p>
          ) : null}
          <div className="flex justify-end">
            <Button
              type="button"
              variant="secondary"
              size="sm"
              disabled={post.isPending || balance.isFetching}
              onClick={() => {
                setReview(null);
                post.reset();
                setValidationError(undefined);
                void balance.refetch();
                void statuses.refetch();
                void reasons.refetch();
                if (source.serial_controlled) void serials.refetch();
              }}
            >
              <RefreshCw className="size-4" /> Refresh stock and options
            </Button>
          </div>
          <fieldset
            disabled={post.isPending || Boolean(review)}
            className="grid gap-4 sm:grid-cols-2"
          >
            <FormField
              label="Target status"
              htmlFor="status-change-target"
              required
            >
              <Select
                id="status-change-target"
                ariaLabel="Target inventory status"
                value={targetId}
                options={targets.map((status) => ({
                  value: status.inventory_status_id,
                  label: `${status.name} (${status.code})`,
                }))}
                placeholder="Select a different status"
                onValueChange={setTargetId}
                disabled={statuses.isPending || statuses.isError}
              />
              {target ? (
                <p className="mt-1 text-xs text-slate-600">
                  Allocation: {target.is_allocatable ? "allowed" : "blocked"} ·
                  Picking: {target.is_pickable ? "allowed" : "blocked"}
                </p>
              ) : null}
            </FormField>
            <FormField
              label={`Quantity (${source.uom_code}, base units)`}
              htmlFor="status-change-quantity"
              required
            >
              <Input
                id="status-change-quantity"
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
                  label="Search serial numbers"
                  htmlFor="status-change-serial-search"
                >
                  <Input
                    id="status-change-serial-search"
                    value={serialSearch}
                    maxLength={160}
                    onChange={(event) => {
                      setSerialSearch(event.target.value);
                      setSerialId("");
                    }}
                  />
                </FormField>
                <FormField
                  label="Serial number"
                  htmlFor="status-change-serial"
                  required
                >
                  <Select
                    id="status-change-serial"
                    ariaLabel="Serial number to change status"
                    value={serialId}
                    options={matchingSerials.map((entry) => ({
                      value: entry.serial_id,
                      label: entry.serial_no,
                    }))}
                    placeholder="Select one serial"
                    onValueChange={setSerialId}
                    disabled={serials.isFetching || serials.isError}
                  />
                </FormField>
              </>
            ) : null}
            <FormField
              label="Business date"
              htmlFor="status-change-date"
              required
            >
              <Input
                id="status-change-date"
                type="date"
                value={date}
                onChange={(event) => setDate(event.target.value)}
              />
            </FormField>
            <FormField
              label="Status-change reference"
              htmlFor="status-change-reference"
              required
            >
              <Input
                id="status-change-reference"
                value={reference}
                maxLength={140}
                onChange={(event) => setReference(event.target.value)}
              />
            </FormField>
            <FormField label="Reason" htmlFor="status-change-reason" required>
              <Select
                id="status-change-reason"
                ariaLabel="Inventory status-change reason"
                value={reasonCode}
                options={(reasons.data ?? [])
                  .filter((entry) => entry.is_active)
                  .map((entry) => ({
                    value: entry.code,
                    label: `${entry.name} (${entry.code})`,
                  }))}
                placeholder="Select inventory reason"
                onValueChange={setReasonCode}
                disabled={reasons.isPending || reasons.isError}
              />
            </FormField>
            <FormField
              label="Notes"
              htmlFor="status-change-notes"
              required={reason?.requires_note}
            >
              <Textarea
                id="status-change-notes"
                value={notes}
                maxLength={4000}
                placeholder="Why is the inventory status changing?"
                onChange={(event) => setNotes(event.target.value)}
              />
            </FormField>
          </fieldset>
          {review ? (
            <section className="space-y-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950">
              <p>
                Change {review.quantity} {source.uom_code} from{" "}
                {source.inventory_status_code} to{" "}
                {target?.name ?? "the selected status"} at{" "}
                {source.location_code}? Total stock and physical location stay
                unchanged. This creates an immutable movement and cannot be
                edited.
              </p>
              <div className="flex flex-wrap justify-end gap-3">
                <Button
                  type="button"
                  variant="secondary"
                  disabled={post.isPending}
                  onClick={() => {
                    setReview(null);
                    post.reset();
                  }}
                >
                  Back to edit
                </Button>
                <Button
                  type="button"
                  disabled={post.isPending}
                  onClick={() => post.mutate(review)}
                >
                  {post.isPending ? (
                    <LoaderCircle className="size-4 animate-spin" />
                  ) : null}
                  Confirm status change
                </Button>
              </div>
            </section>
          ) : (
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
                  loadingOptions || Boolean(lookupError) || balance.isFetching
                }
              >
                Review status change
              </Button>
            </div>
          )}
        </form>
      )}
    </OperationDialog>
  );
}
