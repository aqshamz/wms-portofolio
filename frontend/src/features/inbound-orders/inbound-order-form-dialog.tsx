"use client";

import { useEffect } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Decimal from "decimal.js";
import { LoaderCircle, Trash2, X } from "lucide-react";
import {
  Controller,
  useFieldArray,
  useForm,
  useWatch,
  type FieldErrors,
} from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  createInboundOrder,
  inboundOrderKeys,
} from "@/features/inbound-orders/inbound-order-api";
import {
  inboundOrderFormSchema,
  type InboundOrderFormValues,
} from "@/features/inbound-orders/inbound-order-schema";
import {
  getPurchaseOrder,
  listPurchaseOrders,
  purchaseOrderKeys,
} from "@/features/purchase-orders/purchase-order-api";
import type { PurchaseOrderLine } from "@/features/purchase-orders/purchase-order-types";

const allPurchaseOrders = {
  status: "",
  search: "",
  page: 1,
  pageSize: 100,
};

function optional(value: string) {
  return value.trim() || undefined;
}

function localDateTime(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  const offset = date.getTimezoneOffset() * 60_000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

function remainingQuantity(line: PurchaseOrderLine) {
  return Decimal.max(
    new Decimal(line.ordered_qty).minus(line.scheduled_qty),
    0,
  ).toFixed(6);
}

function CreateLine({
  index,
  source,
  errors,
  register,
  removable,
  onRemove,
}: {
  index: number;
  source?: PurchaseOrderLine;
  errors?: FieldErrors<InboundOrderFormValues["lines"][number]>;
  register: ReturnType<typeof useForm<InboundOrderFormValues>>["register"];
  removable: boolean;
  onRemove: () => void;
}) {
  return (
    <article className="rounded-xl border border-slate-200 p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="font-semibold text-slate-950">
            {source?.item_name ?? "Purchase Order line"}
          </h3>
          <p className="mt-0.5 font-mono text-xs text-slate-500">
            {source
              ? `PO line ${source.line_no} · ${source.item_code} · ${source.uom_code}`
              : "Loading line details…"}
          </p>
        </div>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="text-rose-700"
          disabled={!removable}
          onClick={onRemove}
        >
          <Trash2 className="size-4" />
          Remove
        </Button>
      </div>
      <input
        type="hidden"
        {...register(`lines.${index}.purchase_order_line_id`)}
      />
      <div className="mt-4 grid gap-4 sm:grid-cols-3">
        <FormField
          label="Expected quantity"
          htmlFor={`inbound-qty-${index}`}
          required
          error={errors?.expected_qty?.message}
        >
          <Input
            id={`inbound-qty-${index}`}
            inputMode="decimal"
            invalid={Boolean(errors?.expected_qty)}
            {...register(`lines.${index}.expected_qty`)}
          />
          {source ? (
            <p className="mt-1 text-xs text-slate-500">
              Unscheduled: {remainingQuantity(source)} {source.uom_code}
            </p>
          ) : null}
        </FormField>
        <FormField
          label="Customer line reference"
          htmlFor={`inbound-reference-${index}`}
          error={errors?.customer_line_reference?.message}
        >
          <Input
            id={`inbound-reference-${index}`}
            placeholder="Optional"
            {...register(`lines.${index}.customer_line_reference`)}
          />
        </FormField>
        <FormField
          label="Line notes"
          htmlFor={`inbound-line-notes-${index}`}
          error={errors?.notes?.message}
        >
          <Input
            id={`inbound-line-notes-${index}`}
            placeholder="Optional"
            {...register(`lines.${index}.notes`)}
          />
        </FormField>
      </div>
    </article>
  );
}

export function InboundOrderFormDialog({
  ownerId,
  ownerLabel,
  warehouseId,
  warehouseLabel,
  onOpenChange,
}: {
  ownerId: string;
  ownerLabel: string;
  warehouseId: string;
  warehouseLabel: string;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const purchaseOrders = useQuery({
    queryKey: purchaseOrderKeys.list({
      ...allPurchaseOrders,
      ownerId,
      warehouseId,
    }),
    queryFn: () =>
      listPurchaseOrders({
        ...allPurchaseOrders,
        ownerId,
        warehouseId,
      }),
  });
  const {
    control,
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<InboundOrderFormValues>({
    resolver: zodResolver(inboundOrderFormSchema),
    defaultValues: {
      purchase_order_id: "",
      business_date: new Date().toISOString().slice(0, 10),
      expected_arrival_at: "",
      external_reference: "",
      supplier_reference: "",
      notes: "",
      lines: [],
    },
  });
  const { fields, remove, replace } = useFieldArray({
    control,
    name: "lines",
  });
  const purchaseOrderId = useWatch({ control, name: "purchase_order_id" });
  const purchaseOrder = useQuery({
    queryKey: purchaseOrderKeys.detail(purchaseOrderId || "none"),
    queryFn: () => getPurchaseOrder(purchaseOrderId),
    enabled: Boolean(purchaseOrderId),
  });
  const eligiblePurchaseOrders = (purchaseOrders.data?.items ?? []).filter(
    (item) =>
      item.status_code === "APPROVED" ||
      item.status_code === "PARTIALLY_RECEIVED",
  );

  useEffect(() => {
    if (!purchaseOrder.data) return;
    const availableLines = (purchaseOrder.data.lines ?? [])
      .filter((line) => new Decimal(remainingQuantity(line)).greaterThan(0))
      .map((line) => ({
        purchase_order_line_id: line.purchase_order_line_id,
        expected_qty: remainingQuantity(line),
        customer_line_reference: "",
        notes: "",
      }));
    replace(availableLines);
    setValue(
      "expected_arrival_at",
      localDateTime(purchaseOrder.data.expected_arrival_at),
    );
  }, [purchaseOrder.data, replace, setValue]);

  const save = useMutation({
    mutationFn: (values: InboundOrderFormValues) =>
      createInboundOrder({
        purchase_order_id: values.purchase_order_id,
        business_date: values.business_date,
        expected_arrival_at: values.expected_arrival_at
          ? new Date(values.expected_arrival_at).toISOString()
          : undefined,
        external_reference: optional(values.external_reference),
        supplier_reference: optional(values.supplier_reference),
        notes: optional(values.notes),
        lines: values.lines.map((line) => ({
          purchase_order_line_id: line.purchase_order_line_id,
          expected_qty: line.expected_qty.trim(),
          customer_line_reference: optional(line.customer_line_reference),
          notes: optional(line.notes),
        })),
      }),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: inboundOrderKeys.lists() }),
        queryClient.invalidateQueries({ queryKey: purchaseOrderKeys.all }),
      ]);
      toast.success("Inbound Order draft created.");
      onOpenChange(false);
    },
  });
  const sourceByLineId = new Map(
    (purchaseOrder.data?.lines ?? []).map((line) => [
      line.purchase_order_line_id,
      line,
    ]),
  );

  return (
    <Dialog.Root
      open
      onOpenChange={(open) => !save.isPending && onOpenChange(open)}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[94vh] w-[calc(100%-1.5rem)] max-w-5xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                Create Inbound Order
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                Schedule remaining quantities from an approved Purchase Order.
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close Inbound Order form"
                className="grid size-10 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>
          <form
            noValidate
            className="p-5 sm:p-6"
            onSubmit={handleSubmit((values) => save.mutate(values))}
          >
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
                Inbound plan
              </h2>
              <div className="mt-4 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
                <FormField label="Owner" htmlFor="inbound-owner" required>
                  <Input id="inbound-owner" value={ownerLabel} readOnly />
                </FormField>
                <FormField
                  label="Warehouse"
                  htmlFor="inbound-warehouse"
                  required
                >
                  <Input
                    id="inbound-warehouse"
                    value={warehouseLabel}
                    readOnly
                  />
                </FormField>
                <FormField
                  label="Source Purchase Order"
                  htmlFor="inbound-po"
                  required
                  error={errors.purchase_order_id?.message}
                >
                  <Controller
                    control={control}
                    name="purchase_order_id"
                    render={({ field }) => (
                      <Select
                        id="inbound-po"
                        ariaLabel="Source Purchase Order"
                        value={field.value}
                        options={eligiblePurchaseOrders.map((item) => ({
                          value: item.purchase_order_id,
                          label: `${item.purchase_order_no} · ${item.vendor_name}`,
                        }))}
                        placeholder={
                          purchaseOrders.isPending
                            ? "Loading approved orders…"
                            : eligiblePurchaseOrders.length
                              ? "Select Purchase Order"
                              : "No schedulable Purchase Orders"
                        }
                        disabled={purchaseOrders.isPending}
                        invalid={Boolean(errors.purchase_order_id)}
                        className="mt-2"
                        onValueChange={(value) => {
                          field.onChange(value);
                          replace([]);
                        }}
                      />
                    )}
                  />
                </FormField>
                <FormField
                  label="Business date"
                  htmlFor="inbound-business-date"
                  required
                  error={errors.business_date?.message}
                >
                  <Input
                    id="inbound-business-date"
                    type="date"
                    invalid={Boolean(errors.business_date)}
                    {...register("business_date")}
                  />
                </FormField>
                <FormField
                  label="Expected arrival"
                  htmlFor="inbound-arrival"
                  error={errors.expected_arrival_at?.message}
                >
                  <Input
                    id="inbound-arrival"
                    type="datetime-local"
                    {...register("expected_arrival_at")}
                  />
                </FormField>
                <FormField
                  label="External reference"
                  htmlFor="inbound-external-reference"
                  error={errors.external_reference?.message}
                >
                  <Input
                    id="inbound-external-reference"
                    placeholder="Appointment or ASN"
                    {...register("external_reference")}
                  />
                </FormField>
                <FormField
                  label="Supplier reference"
                  htmlFor="inbound-supplier-reference"
                  error={errors.supplier_reference?.message}
                >
                  <Input
                    id="inbound-supplier-reference"
                    placeholder="Delivery reference"
                    {...register("supplier_reference")}
                  />
                </FormField>
                <div className="sm:col-span-2">
                  <FormField
                    label="Notes"
                    htmlFor="inbound-notes"
                    error={errors.notes?.message}
                  >
                    <textarea
                      id="inbound-notes"
                      rows={3}
                      className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                      {...register("notes")}
                    />
                  </FormField>
                </div>
              </div>
            </section>
            <section className="mt-7 border-t border-slate-200 pt-6">
              <h2 className="text-xs font-bold tracking-wide text-slate-500 uppercase">
                Expected lines
              </h2>
              <p className="mt-1 text-sm text-slate-600">
                All remaining Purchase Order quantities are included initially;
                remove lines that are not part of this arrival.
              </p>
              {purchaseOrder.isPending ? (
                <div className="mt-4 flex items-center gap-2 rounded-xl bg-slate-50 p-4 text-sm text-slate-600">
                  <LoaderCircle className="size-4 animate-spin" />
                  Loading Purchase Order lines…
                </div>
              ) : (
                <div className="mt-4 space-y-3">
                  {fields.map((line, index) => (
                    <CreateLine
                      key={line.id}
                      index={index}
                      source={sourceByLineId.get(line.purchase_order_line_id)}
                      errors={errors.lines?.[index]}
                      register={register}
                      removable={fields.length > 1}
                      onRemove={() => remove(index)}
                    />
                  ))}
                  {purchaseOrderId &&
                  !purchaseOrder.isPending &&
                  !fields.length ? (
                    <p className="rounded-xl bg-amber-50 p-4 text-sm text-amber-900">
                      This Purchase Order has no unscheduled quantity remaining.
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
              <Button type="submit" disabled={save.isPending || !fields.length}>
                {save.isPending ? (
                  <LoaderCircle className="size-4 animate-spin" />
                ) : null}
                {save.isPending ? "Creating…" : "Create draft"}
              </Button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
