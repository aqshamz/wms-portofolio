"use client";

import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { LoaderCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  getReceipt,
  listReceipts,
  receiptKeys,
} from "@/features/receipts/receipt-api";
import {
  createInspection,
  inspectionKeys,
  listReceiptInspections,
} from "./quality-inspection-api";
import { createInspectionSchema } from "./quality-inspection-schema";
import { InspectionDialog } from "./inspection-dialog";

export function InspectionCreateDialog({
  ownerId,
  warehouseId,
  scopeLabel,
  onOpenChange,
  onCreated,
}: {
  ownerId: string;
  warehouseId: string;
  scopeLabel: string;
  onOpenChange: (open: boolean) => void;
  onCreated: (id: string) => void;
}) {
  const queryClient = useQueryClient();
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [receiptId, setReceiptId] = useState("");
  const filters = {
    ownerId,
    warehouseId,
    status: "COMPLETED",
    search,
    page,
    pageSize: 10,
  };
  const receipts = useQuery({
    queryKey: receiptKeys.list(filters),
    queryFn: () => listReceipts(filters),
  });
  const receipt = useQuery({
    queryKey: receiptKeys.detail(receiptId),
    queryFn: () => getReceipt(receiptId),
    enabled: Boolean(receiptId),
  });
  const existing = useQuery({
    queryKey: inspectionKeys.receipt(ownerId, warehouseId, receiptId),
    queryFn: () => listReceiptInspections(ownerId, warehouseId, receiptId),
    enabled: Boolean(receiptId),
  });
  const form = useForm<{ receipt_inventory_id: string; notes: string }>({
    resolver: zodResolver(createInspectionSchema),
    defaultValues: { receipt_inventory_id: "", notes: "" },
  });
  const batchId = useWatch({
    control: form.control,
    name: "receipt_inventory_id",
  });
  const inspectedIds = new Set(
    existing.data?.map((row) => row.receipt_inventory_id) ?? [],
  );
  const inScope =
    receipt.data?.owner_id === ownerId &&
    receipt.data?.warehouse_id === warehouseId &&
    receipt.data.status_code === "COMPLETED";
  const batches = inScope
    ? (receipt.data?.lines ?? []).flatMap((line) =>
        line.batches
          .filter(
            (batch) =>
              batch.initial_balance_id &&
              !inspectedIds.has(batch.receipt_inventory_id),
          )
          .map((batch) => ({
            ...batch,
            item_code: line.item_code,
            item_name: line.item_name,
          })),
      )
    : [];
  const selected = batches.find(
    (batch) => batch.receipt_inventory_id === batchId,
  );
  const ready = inScope && existing.isSuccess && Boolean(selected);
  const receiptOptions = (receipts.data?.items ?? []).map((row) => ({
    value: row.receipt_id,
    label: `${row.receipt_id} · ${row.delivery_note_no || row.business_date}`,
  }));
  if (
    inScope &&
    receipt.data &&
    !receiptOptions.some((option) => option.value === receiptId)
  ) {
    receiptOptions.unshift({
      value: receiptId,
      label: `${receiptId} · ${receipt.data.delivery_note_no || receipt.data.business_date}`,
    });
  }
  const create = useMutation({
    mutationFn: (values: { receipt_inventory_id: string; notes: string }) => {
      if (!ready)
        throw new Error(
          "Select an eligible batch in this warehouse and owner scope.",
        );
      return createInspection({
        receipt_inventory_id: values.receipt_inventory_id,
        notes: values.notes.trim() || undefined,
      });
    },
    onSuccess: async (inspection) => {
      await queryClient.invalidateQueries({ queryKey: inspectionKeys.all });
      toast.success("Quality inspection opened.");
      onCreated(inspection.inspection_id);
    },
    onError: (error) => toast.error(error.message),
  });
  const error =
    receipts.error ?? receipt.error ?? existing.error ?? create.error;

  return (
    <InspectionDialog
      title="Start quality inspection"
      description={scopeLabel}
      busy={create.isPending}
      onOpenChange={onOpenChange}
    >
      <form
        className="space-y-5"
        onSubmit={form.handleSubmit(
          (values) => create.mutate(values),
          () => toast.error("Please check the highlighted fields."),
        )}
      >
        <p className="rounded-xl bg-cyan-50 p-3 text-sm text-cyan-900">
          Inspect one accepted batch from a completed receipt. Quantities are in
          the item’s base unit, not the purchase or packaging unit.
        </p>
        <fieldset className="space-y-5" disabled={create.isPending}>
          <FormField
            label="Search completed receipts"
            htmlFor="qc-receipt-search"
          >
            <Input
              id="qc-receipt-search"
              value={search}
              maxLength={160}
              placeholder="Receipt or delivery note"
              onChange={(event) => {
                setSearch(event.target.value);
                setPage(1);
              }}
            />
          </FormField>
          <FormField label="Completed receipt" htmlFor="qc-receipt" required>
            <Select
              id="qc-receipt"
              ariaLabel="Completed receipt"
              className="mt-2"
              value={receiptId}
              options={receiptOptions}
              placeholder={
                receipts.isPending
                  ? "Loading receipts…"
                  : "Select completed receipt"
              }
              disabled={receipts.isPending || create.isPending}
              onValueChange={(id) => {
                setReceiptId(id);
                form.setValue("receipt_inventory_id", "");
                form.clearErrors();
              }}
            />
            {receipts.data && receipts.data.total_pages > 1 ? (
              <div className="mt-2 flex items-center gap-3 text-xs">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                >
                  Previous
                </Button>
                <span>
                  Page {page} of {receipts.data.total_pages}
                </span>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  disabled={page >= receipts.data.total_pages}
                  onClick={() => setPage(page + 1)}
                >
                  Next
                </Button>
              </div>
            ) : null}
            {receipts.isSuccess && !receipts.data.items.length ? (
              <p className="mt-2 text-sm text-slate-600">
                No completed receipts match. Complete a receipt first.
              </p>
            ) : null}
          </FormField>
          <FormField
            label="Receipt batch"
            htmlFor="qc-batch"
            required
            error={form.formState.errors.receipt_inventory_id?.message}
          >
            <Select
              id="qc-batch"
              ariaLabel="Receipt batch"
              className="mt-2"
              value={batchId}
              options={batches.map((batch) => ({
                value: batch.receipt_inventory_id,
                label: `${batch.item_code} · ${batch.base_qty} ${batch.base_uom_code} · ${batch.lot_number || batch.serial_no || batch.handling_unit_barcode || batch.receipt_inventory_id} · ${batch.received_location_code}`,
              }))}
              disabled={
                !receiptId ||
                receipt.isPending ||
                existing.isPending ||
                existing.isError ||
                create.isPending
              }
              placeholder="Select uninspected batch"
              invalid={Boolean(form.formState.errors.receipt_inventory_id)}
              onValueChange={(id) =>
                form.setValue("receipt_inventory_id", id, {
                  shouldValidate: true,
                })
              }
            />
            {receiptId &&
            receipt.isSuccess &&
            existing.isSuccess &&
            !batches.length ? (
              <p className="mt-2 text-sm text-amber-800">
                No uninspected batches remain. Open an existing pending
                inspection or its replacement from the list.
              </p>
            ) : null}
          </FormField>
          {selected ? (
            <dl className="grid gap-3 rounded-xl bg-slate-50 p-4 text-sm sm:grid-cols-2">
              <div>
                <dt className="text-xs text-slate-500">Item / base quantity</dt>
                <dd className="font-semibold">
                  {selected.item_name} · {selected.base_qty}{" "}
                  {selected.base_uom_code}
                </dd>
              </div>
              <div>
                <dt className="text-xs text-slate-500">Received location</dt>
                <dd>{selected.received_location_code}</dd>
              </div>
              {selected.lot_number ? (
                <div>
                  <dt className="text-xs text-slate-500">Lot</dt>
                  <dd>{selected.lot_number}</dd>
                </div>
              ) : null}
              {selected.serial_no || selected.handling_unit_barcode ? (
                <div>
                  <dt className="text-xs text-slate-500">
                    Serial / handling unit (cannot split)
                  </dt>
                  <dd>
                    {selected.serial_no || selected.handling_unit_barcode}
                  </dd>
                </div>
              ) : null}
            </dl>
          ) : null}
          <FormField
            label="Notes"
            htmlFor="qc-create-notes"
            error={form.formState.errors.notes?.message}
          >
            <textarea
              id="qc-create-notes"
              className="mt-2 min-h-24 w-full rounded-xl border border-slate-300 p-3 text-sm"
              maxLength={4000}
              {...form.register("notes")}
            />
          </FormField>
        </fieldset>
        {error ? (
          <p
            role="alert"
            className="rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
          >
            {error.message}
          </p>
        ) : null}
        <div className="flex flex-col-reverse gap-3 border-t border-slate-200 pt-4 sm:flex-row sm:justify-end">
          <Button
            type="button"
            variant="secondary"
            disabled={create.isPending}
            onClick={() => onOpenChange(false)}
          >
            Close
          </Button>
          <Button type="submit" disabled={!ready || create.isPending}>
            {create.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            Start inspection
          </Button>
        </div>
      </form>
    </InspectionDialog>
  );
}
