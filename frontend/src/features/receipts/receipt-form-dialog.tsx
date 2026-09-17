"use client";

import { useEffect, useMemo } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Decimal from "decimal.js";
import { LoaderCircle, Plus, Trash2, X } from "lucide-react";
import {
  Controller,
  useFieldArray,
  useForm,
  useWatch,
  type Control,
  type FieldErrors,
  type UseFormRegister,
} from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  getInboundOrder,
  inboundOrderKeys,
  listInboundOrders,
} from "@/features/inbound-orders/inbound-order-api";
import type { InboundOrderLine } from "@/features/inbound-orders/inbound-order-types";
import {
  itemCatalogKeys,
  listItems,
} from "@/features/item-catalog/item-catalog-api";
import type { CatalogItem } from "@/features/item-catalog/item-catalog-types";
import {
  createReceipt,
  receiptKeys,
  updateReceipt,
} from "@/features/receipts/receipt-api";
import {
  emptyReceiptBatch,
  receiptFormSchema,
  type ReceiptFormValues,
} from "@/features/receipts/receipt-schema";
import type {
  Receipt,
  ReceiptLineRequest,
} from "@/features/receipts/receipt-types";
import { receiptLocationOptions } from "@/features/receipts/receipt-location-options";
import {
  listLocationTypes,
  listLocations,
  storageLayoutKeys,
} from "@/features/storage-layout/storage-layout-api";
import type { WarehouseLocation } from "@/features/storage-layout/storage-layout-types";

const pageSize = 100;

function optional(value: string) {
  return value.trim() || undefined;
}

function localDateTime(value?: string) {
  const date = value ? new Date(value) : new Date();
  const offset = date.getTimezoneOffset() * 60_000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

function toRequestLine(
  line: ReceiptFormValues["lines"][number],
): ReceiptLineRequest {
  return {
    inbound_line_id: line.inbound_line_id,
    received_qty: line.received_qty.trim(),
    rejected_qty: line.rejected_qty.trim(),
    exception_type_code:
      line.exception_type_code === "NONE"
        ? undefined
        : line.exception_type_code,
    exception_notes: optional(line.exception_notes),
    batches: line.batches.map((batch) => ({
      source_qty: batch.source_qty.trim(),
      received_location_id: batch.received_location_id,
      lot: line.lot_controlled
        ? {
            lot_number: batch.lot_number.trim(),
            manufacture_date: optional(batch.manufacture_date),
            expiry_date: optional(batch.expiry_date),
          }
        : undefined,
      handling_unit_id: optional(batch.handling_unit_id),
      serial_no: line.serial_controlled ? optional(batch.serial_no) : undefined,
    })),
  };
}

function remaining(line: InboundOrderLine) {
  return Decimal.max(
    new Decimal(line.expected_qty).minus(line.completed_receipt_qty),
    0,
  ).toFixed(6);
}

function BatchEditor({
  lineIndex,
  batchIndex,
  control,
  register,
  locations,
  lotControlled,
  serialControlled,
  error,
  canRemove,
  onRemove,
}: {
  lineIndex: number;
  batchIndex: number;
  control: Control<ReceiptFormValues>;
  register: UseFormRegister<ReceiptFormValues>;
  locations: WarehouseLocation[];
  lotControlled: boolean;
  serialControlled: boolean;
  error?: FieldErrors<ReceiptFormValues["lines"][number]["batches"][number]>;
  canRemove: boolean;
  onRemove: () => void;
}) {
  return (
    <div className="rounded-xl bg-slate-50 p-3">
      <div className="flex items-center justify-between gap-3">
        <p className="text-xs font-bold tracking-wide text-slate-500 uppercase">
          Batch {batchIndex + 1}
        </p>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="text-rose-700"
          disabled={!canRemove}
          onClick={onRemove}
        >
          <Trash2 className="size-4" /> Remove
        </Button>
      </div>
      <div className="mt-3 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <FormField
          label="Batch quantity"
          htmlFor={`receipt-${lineIndex}-${batchIndex}-qty`}
          required
          error={error?.source_qty?.message}
        >
          <Input
            id={`receipt-${lineIndex}-${batchIndex}-qty`}
            inputMode="decimal"
            invalid={Boolean(error?.source_qty)}
            {...register(`lines.${lineIndex}.batches.${batchIndex}.source_qty`)}
          />
        </FormField>
        <FormField
          label="Received location"
          htmlFor={`receipt-${lineIndex}-${batchIndex}-location`}
          required
          error={error?.received_location_id?.message}
        >
          <Controller
            control={control}
            name={`lines.${lineIndex}.batches.${batchIndex}.received_location_id`}
            render={({ field }) => (
              <Select
                id={`receipt-${lineIndex}-${batchIndex}-location`}
                ariaLabel={`Batch ${batchIndex + 1} received location`}
                value={field.value}
                options={locations.map((location) => ({
                  value: location.location_id,
                  label: `${location.code}${location.zone_code ? ` · ${location.zone_code}` : ""}`,
                }))}
                placeholder="Select location"
                invalid={Boolean(error?.received_location_id)}
                className="mt-2"
                onValueChange={field.onChange}
              />
            )}
          />
        </FormField>
        {lotControlled ? (
          <FormField
            label="Lot number"
            htmlFor={`receipt-${lineIndex}-${batchIndex}-lot`}
            required
            error={error?.lot_number?.message}
          >
            <Input
              id={`receipt-${lineIndex}-${batchIndex}-lot`}
              invalid={Boolean(error?.lot_number)}
              {...register(
                `lines.${lineIndex}.batches.${batchIndex}.lot_number`,
              )}
            />
          </FormField>
        ) : null}
        {lotControlled ? (
          <>
            <FormField
              label="Manufacture date"
              htmlFor={`receipt-${lineIndex}-${batchIndex}-manufacture`}
              error={error?.manufacture_date?.message}
            >
              <Input
                id={`receipt-${lineIndex}-${batchIndex}-manufacture`}
                type="date"
                {...register(
                  `lines.${lineIndex}.batches.${batchIndex}.manufacture_date`,
                )}
              />
            </FormField>
            <FormField
              label="Expiry date"
              htmlFor={`receipt-${lineIndex}-${batchIndex}-expiry`}
              error={error?.expiry_date?.message}
            >
              <Input
                id={`receipt-${lineIndex}-${batchIndex}-expiry`}
                type="date"
                {...register(
                  `lines.${lineIndex}.batches.${batchIndex}.expiry_date`,
                )}
              />
            </FormField>
          </>
        ) : null}
        {serialControlled ? (
          <FormField
            label="Serial number"
            htmlFor={`receipt-${lineIndex}-${batchIndex}-serial`}
            required
            error={error?.serial_no?.message}
          >
            <Input
              id={`receipt-${lineIndex}-${batchIndex}-serial`}
              invalid={Boolean(error?.serial_no)}
              {...register(
                `lines.${lineIndex}.batches.${batchIndex}.serial_no`,
              )}
            />
          </FormField>
        ) : null}
        <FormField
          label="Handling unit ID"
          htmlFor={`receipt-${lineIndex}-${batchIndex}-hu`}
          error={error?.handling_unit_id?.message}
        >
          <Input
            id={`receipt-${lineIndex}-${batchIndex}-hu`}
            placeholder="Optional existing HU"
            {...register(
              `lines.${lineIndex}.batches.${batchIndex}.handling_unit_id`,
            )}
          />
        </FormField>
      </div>
    </div>
  );
}

function ReceiptLineEditor({
  index,
  source,
  control,
  register,
  locations,
  error,
  removable,
  onRemove,
}: {
  index: number;
  source?: InboundOrderLine;
  control: Control<ReceiptFormValues>;
  register: UseFormRegister<ReceiptFormValues>;
  locations: WarehouseLocation[];
  error?: FieldErrors<ReceiptFormValues["lines"][number]>;
  removable: boolean;
  onRemove: () => void;
}) {
  const line = useWatch({ control, name: `lines.${index}` });
  const batches = useFieldArray({ control, name: `lines.${index}.batches` });
  const accepted = useMemo(() => {
    try {
      return Decimal.max(
        new Decimal(line.received_qty || 0).minus(line.rejected_qty || 0),
        0,
      ).toFixed(6);
    } catch {
      return "—";
    }
  }, [line.received_qty, line.rejected_qty]);
  const batchTotal = useMemo(() => {
    try {
      return line.batches
        .reduce(
          (total, batch) => total.plus(batch.source_qty || 0),
          new Decimal(0),
        )
        .toFixed(6);
    } catch {
      return "—";
    }
  }, [line.batches]);
  const batchError = error?.batches?.root?.message ?? error?.batches?.message;

  return (
    <article className="rounded-2xl border border-slate-200 p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="font-semibold text-slate-950">
            {source?.item_name ?? `Receipt line ${index + 1}`}
          </h3>
          <p className="mt-0.5 font-mono text-xs text-slate-500">
            {source
              ? `${source.item_code} · ${source.uom_code} · remaining ${remaining(source)}`
              : line.item_id}
          </p>
          <div className="mt-2 flex flex-wrap gap-2 text-xs font-semibold">
            {line.lot_controlled ? (
              <span className="rounded-full bg-violet-50 px-2 py-1 text-violet-700">
                Lot controlled
              </span>
            ) : null}
            {line.serial_controlled ? (
              <span className="rounded-full bg-amber-50 px-2 py-1 text-amber-800">
                Serialized
              </span>
            ) : null}
          </div>
        </div>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="text-rose-700"
          disabled={!removable}
          onClick={onRemove}
        >
          <Trash2 className="size-4" /> Remove line
        </Button>
      </div>
      <input type="hidden" {...register(`lines.${index}.inbound_line_id`)} />
      <input type="hidden" {...register(`lines.${index}.item_id`)} />
      <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <FormField
          label="Received quantity"
          htmlFor={`receipt-${index}-received`}
          required
          error={error?.received_qty?.message}
        >
          <Input
            id={`receipt-${index}-received`}
            inputMode="decimal"
            invalid={Boolean(error?.received_qty)}
            {...register(`lines.${index}.received_qty`)}
          />
        </FormField>
        <FormField
          label="Rejected quantity"
          htmlFor={`receipt-${index}-rejected`}
          required
          error={error?.rejected_qty?.message}
        >
          <Input
            id={`receipt-${index}-rejected`}
            inputMode="decimal"
            invalid={Boolean(error?.rejected_qty)}
            {...register(`lines.${index}.rejected_qty`)}
          />
        </FormField>
        <FormField
          label="Accepted quantity"
          htmlFor={`receipt-${index}-accepted`}
        >
          <Input id={`receipt-${index}-accepted`} value={accepted} readOnly />
        </FormField>
        <FormField
          label="Exception type"
          htmlFor={`receipt-${index}-exception-type`}
          error={error?.exception_type_code?.message}
        >
          <Controller
            control={control}
            name={`lines.${index}.exception_type_code`}
            render={({ field }) => (
              <Select
                id={`receipt-${index}-exception-type`}
                ariaLabel={`Line ${index + 1} exception type`}
                value={field.value}
                options={[
                  { value: "NONE", label: "No exception" },
                  { value: "REJECTED_AT_DOCK", label: "Rejected at dock" },
                  { value: "DAMAGED", label: "Damaged" },
                  { value: "WRONG_ITEM", label: "Wrong item" },
                ]}
                className="mt-2"
                onValueChange={field.onChange}
              />
            )}
          />
        </FormField>
        <div className="sm:col-span-2 lg:col-span-4">
          <FormField
            label="Exception notes"
            htmlFor={`receipt-${index}-exception-notes`}
            error={error?.exception_notes?.message}
          >
            <textarea
              id={`receipt-${index}-exception-notes`}
              rows={2}
              placeholder="Required for rejected or over-received quantity"
              className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
              {...register(`lines.${index}.exception_notes`)}
            />
          </FormField>
        </div>
      </div>
      <div className="mt-5 flex items-center justify-between gap-3">
        <div>
          <p className="text-sm font-semibold text-slate-900">
            Inventory batches
          </p>
          <p className="text-xs text-slate-500">
            Batch total must equal accepted quantity. Serialized items need one
            serial per batch.
          </p>
        </div>
        <Button
          type="button"
          variant="secondary"
          size="sm"
          onClick={() =>
            batches.append({
              ...emptyReceiptBatch,
              received_location_id: locations[0]?.location_id ?? "",
            })
          }
        >
          <Plus className="size-4" /> Add batch
        </Button>
      </div>
      <p
        aria-live="polite"
        className={`mt-3 rounded-lg px-3 py-2 text-sm font-semibold ${batchTotal === accepted ? "bg-emerald-50 text-emerald-800" : "bg-amber-50 text-amber-900"}`}
      >
        Batch total: {batchTotal} · Accepted: {accepted}
        {batchTotal !== accepted ? " · Quantities do not match" : " · Balanced"}
      </p>
      {typeof batchError === "string" ? (
        <p role="alert" className="mt-2 text-sm font-semibold text-rose-700">
          {batchError}
        </p>
      ) : null}
      <div className="mt-3 space-y-3">
        {batches.fields.map((batch, batchIndex) => (
          <BatchEditor
            key={batch.id}
            lineIndex={index}
            batchIndex={batchIndex}
            control={control}
            register={register}
            locations={locations}
            lotControlled={line.lot_controlled}
            serialControlled={line.serial_controlled}
            error={error?.batches?.[batchIndex]}
            canRemove={batches.fields.length > 1 || accepted === "0.000000"}
            onRemove={() => batches.remove(batchIndex)}
          />
        ))}
      </div>
    </article>
  );
}

export function ReceiptFormDialog({
  ownerId,
  warehouseId,
  scopeLabel,
  receipt,
  replacementFor,
  onOpenChange,
}: {
  ownerId: string;
  warehouseId: string;
  scopeLabel: string;
  receipt?: Receipt;
  replacementFor?: Receipt;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const editing = Boolean(receipt);
  const replacing = Boolean(replacementFor);
  const locationsFilter = {
    warehouseId,
    search: "",
    active: "active" as const,
    page: 1,
    pageSize,
  };
  const locationsQuery = useQuery({
    queryKey: storageLayoutKeys.locations(locationsFilter),
    queryFn: () => listLocations(locationsFilter),
  });
  const locations = (locationsQuery.data?.items ?? []).filter(
    (location) => !location.is_locked,
  );
  const locationTypes = useQuery({
    queryKey: storageLayoutKeys.locationTypes("active"),
    queryFn: () => listLocationTypes("active"),
  });
  const { receiving: receivingLocations, docks: dockOptions } =
    receiptLocationOptions(locations, locationTypes.data ?? []);
  const itemsFilter = {
    ownerId,
    categoryId: "",
    search: "",
    active: "active" as const,
    page: 1,
    pageSize,
  };
  const itemsQuery = useQuery({
    queryKey: itemCatalogKeys.itemList(itemsFilter),
    queryFn: () => listItems(itemsFilter),
  });
  const itemsById = useMemo(
    () =>
      new Map<string, CatalogItem>(
        (itemsQuery.data?.items ?? []).map((item) => [item.item_id, item]),
      ),
    [itemsQuery.data?.items],
  );
  const inboundFilters = {
    ownerId,
    warehouseId,
    status: "",
    search: "",
    page: 1,
    pageSize,
  };
  const inboundOrders = useQuery({
    queryKey: inboundOrderKeys.list(inboundFilters),
    queryFn: () => listInboundOrders(inboundFilters),
    enabled: !editing && !replacing,
  });
  const {
    control,
    register,
    handleSubmit,
    getValues,
    setValue,
    formState: { errors },
  } = useForm<ReceiptFormValues>({
    resolver: zodResolver(receiptFormSchema),
    defaultValues: {
      inbound_id: receipt?.inbound_id ?? replacementFor?.inbound_id ?? "",
      business_date:
        receipt?.business_date ?? new Date().toISOString().slice(0, 10),
      received_at: localDateTime(receipt?.received_at),
      dock_location_id: receipt?.dock_location_id ?? "",
      vehicle_number: receipt?.vehicle_number ?? "",
      seal_number: receipt?.seal_number ?? "",
      delivery_note_no: receipt?.delivery_note_no ?? "",
      notes: receipt?.notes ?? "",
      lines: (receipt?.lines ?? []).map((line) => {
        const item = itemsById.get(line.item_id);
        return {
          inbound_line_id: line.inbound_line_id ?? "",
          item_id: line.item_id,
          received_qty: line.received_qty,
          rejected_qty: line.rejected_qty,
          exception_type_code:
            (line.exception_type_code as ReceiptFormValues["lines"][number]["exception_type_code"]) ??
            "NONE",
          exception_notes: line.exception_notes ?? "",
          lot_controlled:
            item?.lot_controlled ??
            Boolean(line.batches.some((batch) => batch.lot_number)),
          serial_controlled:
            item?.serial_controlled ??
            Boolean(line.batches.some((batch) => batch.serial_no)),
          batches: line.batches.map((batch) => ({
            source_qty: batch.source_qty,
            received_location_id: batch.received_location_id,
            lot_number: batch.lot_number ?? "",
            manufacture_date: "",
            expiry_date: "",
            handling_unit_id: batch.handling_unit_id ?? "",
            serial_no: batch.serial_no ?? "",
          })),
        };
      }),
    },
  });
  const {
    fields: lineFields,
    remove: removeLine,
    replace: replaceLines,
  } = useFieldArray({ control, name: "lines" });
  const inboundId = useWatch({ control, name: "inbound_id" });
  const dockLocationId = useWatch({ control, name: "dock_location_id" });
  const inboundOrder = useQuery({
    queryKey: inboundOrderKeys.detail(inboundId || "none"),
    queryFn: () => getInboundOrder(inboundId),
    enabled: Boolean(inboundId),
  });

  useEffect(() => {
    if (editing || !inboundOrder.data || itemsQuery.isPending) return;
    replaceLines(
      (inboundOrder.data.lines ?? [])
        .filter((line) => new Decimal(remaining(line)).gt(0))
        .map((line) => {
          const item = itemsById.get(line.item_id);
          const quantity = remaining(line);
          return {
            inbound_line_id: line.inbound_line_id,
            item_id: line.item_id,
            received_qty: quantity,
            rejected_qty: "0",
            exception_type_code: "NONE" as const,
            exception_notes: "",
            lot_controlled: item?.lot_controlled ?? false,
            serial_controlled: item?.serial_controlled ?? false,
            batches: [
              {
                ...emptyReceiptBatch,
                source_qty: quantity,
                received_location_id: getValues("dock_location_id"),
                lot_number: line.expected_lot_no ?? "",
                expiry_date: line.expected_expiry_date ?? "",
              },
            ],
          };
        }),
    );
    setValue("business_date", inboundOrder.data.business_date);
  }, [
    editing,
    getValues,
    inboundOrder.data,
    itemsById,
    itemsQuery.isPending,
    replaceLines,
    setValue,
  ]);

  useEffect(() => {
    if (!editing || itemsQuery.isPending) return;
    getValues("lines").forEach((line, index) => {
      const item = itemsById.get(line.item_id);
      if (!item) return;
      setValue(`lines.${index}.lot_controlled`, item.lot_controlled);
      setValue(`lines.${index}.serial_controlled`, item.serial_controlled);
    });
  }, [editing, getValues, itemsById, itemsQuery.isPending, setValue]);

  useEffect(() => {
    if (!dockLocationId) return;
    const lines = getValues("lines");
    lines.forEach((line, lineIndex) =>
      line.batches.forEach((batch, batchIndex) => {
        if (!batch.received_location_id) {
          setValue(
            `lines.${lineIndex}.batches.${batchIndex}.received_location_id`,
            dockLocationId,
          );
        }
      }),
    );
  }, [dockLocationId, getValues, setValue]);

  const save = useMutation({
    mutationFn: (values: ReceiptFormValues) => {
      const receivingIds = new Set(
        receivingLocations.map((location) => location.location_id),
      );
      const dockIds = new Set(
        dockOptions.map((location) => location.location_id),
      );
      if (
        !dockIds.has(values.dock_location_id) ||
        values.lines.some((line) =>
          line.batches.some(
            (batch) => !receivingIds.has(batch.received_location_id),
          ),
        )
      ) {
        throw new Error(
          "Select an active, unlocked Dock for the header and a receiving-enabled location for every batch.",
        );
      }
      const lines = values.lines.map(toRequestLine);
      if (receipt) {
        return updateReceipt(receipt.receipt_id, {
          expected_version: receipt.version_no,
          received_at: new Date(values.received_at).toISOString(),
          dock_location_id: values.dock_location_id,
          vehicle_number: optional(values.vehicle_number),
          seal_number: optional(values.seal_number),
          delivery_note_no: optional(values.delivery_note_no),
          notes: optional(values.notes),
          lines,
        });
      }
      return createReceipt({
        inbound_id: values.inbound_id,
        business_date: values.business_date,
        received_at: new Date(values.received_at).toISOString(),
        dock_location_id: values.dock_location_id,
        vehicle_number: optional(values.vehicle_number),
        seal_number: optional(values.seal_number),
        delivery_note_no: optional(values.delivery_note_no),
        notes: optional(values.notes),
        supersedes_receipt_id: replacementFor?.receipt_id,
        lines,
      });
    },
    onSuccess: async (saved) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: receiptKeys.lists() }),
        queryClient.invalidateQueries({
          queryKey: receiptKeys.detail(saved.receipt_id),
        }),
        queryClient.invalidateQueries({ queryKey: inboundOrderKeys.all }),
      ]);
      toast.success(
        receipt
          ? "Receipt draft updated."
          : replacementFor
            ? "Replacement receipt opened."
            : "Receipt opened.",
      );
      onOpenChange(false);
    },
  });
  const sourceById = new Map(
    (inboundOrder.data?.lines ?? []).map((line) => [
      line.inbound_line_id,
      line,
    ]),
  );
  const eligibleInboundOrders = (inboundOrders.data?.items ?? []).filter(
    (order) =>
      order.status_code === "RELEASED" ||
      order.status_code === "PARTIALLY_RECEIVED",
  );

  return (
    <Dialog.Root
      open
      onOpenChange={(open) => !save.isPending && onOpenChange(open)}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[96vh] w-[calc(100%-1rem)] max-w-6xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="sticky top-0 z-20 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                {editing
                  ? `Edit ${receipt?.receipt_id}`
                  : replacing
                    ? "Open replacement receipt"
                    : "Open receipt"}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                Record delivered, rejected, lot, serial, and receiving-location
                details.
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close receipt form"
                className="grid size-10 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>
          <form
            noValidate
            className="p-5 sm:p-6"
            onSubmit={handleSubmit(
              (values) => save.mutate(values),
              (invalidErrors) => {
                const batchIssue = getValues("lines")
                  .map((_, index) => {
                    const batchErrors = invalidErrors.lines?.[index]?.batches;
                    const message =
                      batchErrors?.root?.message ?? batchErrors?.message;
                    return typeof message === "string"
                      ? `Line ${index + 1}: ${message}`
                      : undefined;
                  })
                  .find(Boolean);
                toast.error("Receipt could not be saved.", {
                  description:
                    batchIssue ?? "Check the highlighted fields and try again.",
                });
              },
            )}
          >
            {locationsQuery.error || locationTypes.error ? (
              <p
                role="alert"
                className="mb-5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
              >
                Receiving locations could not be loaded.{" "}
                {locationsQuery.error?.message || locationTypes.error?.message}
              </p>
            ) : null}
            {save.error ? (
              <p
                role="alert"
                className="mb-5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
              >
                {save.error.message}
              </p>
            ) : null}
            <section>
              <h2 className="text-xs font-bold tracking-wide text-slate-500 uppercase">
                Receipt header
              </h2>
              <p className="mt-1 text-sm text-slate-600">{scopeLabel}</p>
              <div className="mt-4 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
                <FormField
                  label="Inbound Order"
                  htmlFor="receipt-inbound"
                  required
                  error={errors.inbound_id?.message}
                >
                  {editing || replacing ? (
                    <Input
                      id="receipt-inbound"
                      value={
                        receipt?.inbound_id ?? replacementFor?.inbound_id ?? ""
                      }
                      readOnly
                    />
                  ) : (
                    <Controller
                      control={control}
                      name="inbound_id"
                      render={({ field }) => (
                        <Select
                          id="receipt-inbound"
                          ariaLabel="Source Inbound Order"
                          value={field.value}
                          options={eligibleInboundOrders.map((order) => ({
                            value: order.inbound_id,
                            label: `${order.inbound_id} · ${order.vendor_name}`,
                          }))}
                          placeholder={
                            inboundOrders.isPending
                              ? "Loading released orders…"
                              : eligibleInboundOrders.length
                                ? "Select released Inbound Order"
                                : "No receivable Inbound Orders"
                          }
                          invalid={Boolean(errors.inbound_id)}
                          className="mt-2"
                          onValueChange={(value) => {
                            field.onChange(value);
                            replaceLines([]);
                          }}
                        />
                      )}
                    />
                  )}
                </FormField>
                <FormField
                  label="Business date"
                  htmlFor="receipt-date"
                  required
                  error={errors.business_date?.message}
                >
                  <Input
                    id="receipt-date"
                    type="date"
                    readOnly={editing}
                    invalid={Boolean(errors.business_date)}
                    {...register("business_date")}
                  />
                </FormField>
                <FormField
                  label="Received at"
                  htmlFor="receipt-at"
                  required
                  error={errors.received_at?.message}
                >
                  <Input
                    id="receipt-at"
                    type="datetime-local"
                    invalid={Boolean(errors.received_at)}
                    {...register("received_at")}
                  />
                </FormField>
                <FormField
                  label="Receiving dock"
                  htmlFor="receipt-dock"
                  required
                  error={errors.dock_location_id?.message}
                >
                  <Controller
                    control={control}
                    name="dock_location_id"
                    render={({ field }) => (
                      <Select
                        id="receipt-dock"
                        ariaLabel="Receiving dock"
                        value={field.value}
                        options={dockOptions.map((location) => ({
                          value: location.location_id,
                          label: `${location.code}${location.location_type_name ? ` · ${location.location_type_name}` : ""}`,
                        }))}
                        placeholder={
                          locationsQuery.isPending || locationTypes.isPending
                            ? "Loading locations…"
                            : dockOptions.length
                              ? "Select receiving dock"
                              : "No active Dock location"
                        }
                        disabled={
                          locationsQuery.isPending || locationTypes.isPending
                        }
                        invalid={Boolean(errors.dock_location_id)}
                        className="mt-2"
                        onValueChange={field.onChange}
                      />
                    )}
                  />
                </FormField>
                <FormField
                  label="Vehicle number"
                  htmlFor="receipt-vehicle"
                  error={errors.vehicle_number?.message}
                >
                  <Input id="receipt-vehicle" {...register("vehicle_number")} />
                </FormField>
                <FormField
                  label="Seal number"
                  htmlFor="receipt-seal"
                  error={errors.seal_number?.message}
                >
                  <Input id="receipt-seal" {...register("seal_number")} />
                </FormField>
                <FormField
                  label="Delivery note"
                  htmlFor="receipt-delivery"
                  error={errors.delivery_note_no?.message}
                >
                  <Input
                    id="receipt-delivery"
                    {...register("delivery_note_no")}
                  />
                </FormField>
                <FormField
                  label="Notes"
                  htmlFor="receipt-notes"
                  error={errors.notes?.message}
                >
                  <Input id="receipt-notes" {...register("notes")} />
                </FormField>
              </div>
            </section>
            <section className="mt-7 border-t border-slate-200 pt-6">
              <h2 className="text-xs font-bold tracking-wide text-slate-500 uppercase">
                Received lines
              </h2>
              {inboundOrder.isPending || itemsQuery.isPending ? (
                <div className="mt-4 flex items-center gap-2 rounded-xl bg-slate-50 p-4 text-sm text-slate-600">
                  <LoaderCircle className="size-4 animate-spin" /> Loading item
                  controls and expected lines…
                </div>
              ) : (
                <div className="mt-4 space-y-4">
                  {lineFields.map((field, index) => (
                    <ReceiptLineEditor
                      key={field.id}
                      index={index}
                      source={sourceById.get(field.inbound_line_id)}
                      control={control}
                      register={register}
                      locations={receivingLocations}
                      error={errors.lines?.[index]}
                      removable={lineFields.length > 1}
                      onRemove={() => removeLine(index)}
                    />
                  ))}
                  {inboundId && !lineFields.length ? (
                    <p className="rounded-xl bg-amber-50 p-4 text-sm text-amber-900">
                      This Inbound Order has no remaining quantity to receive.
                    </p>
                  ) : null}
                </div>
              )}
            </section>
            <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
              <Dialog.Close asChild>
                <Button type="button" variant="secondary">
                  Cancel
                </Button>
              </Dialog.Close>
              <Button
                type="submit"
                disabled={
                  save.isPending ||
                  !lineFields.length ||
                  itemsQuery.isPending ||
                  !locationsQuery.isSuccess ||
                  !locationTypes.isSuccess
                }
              >
                {save.isPending ? (
                  <LoaderCircle className="size-4 animate-spin" />
                ) : null}
                {save.isPending
                  ? "Saving…"
                  : editing
                    ? "Save draft"
                    : "Open receipt"}
              </Button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
