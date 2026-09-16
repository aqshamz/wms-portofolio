"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, X } from "lucide-react";
import { Controller, useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  getItem,
  itemCatalogKeys,
  listItems,
  listUOMs,
} from "@/features/item-catalog/item-catalog-api";
import {
  addPurchaseOrderLine,
  purchaseOrderKeys,
  updatePurchaseOrder,
  updatePurchaseOrderLine,
} from "@/features/purchase-orders/purchase-order-api";
import {
  emptyPurchaseOrderLine,
  purchaseOrderHeaderSchema,
  purchaseOrderLineSchema,
  type PurchaseOrderFormValues,
  type PurchaseOrderHeaderValues,
} from "@/features/purchase-orders/purchase-order-schema";
import type {
  PurchaseOrder,
  PurchaseOrderLine,
} from "@/features/purchase-orders/purchase-order-types";
import { ApiError } from "@/lib/api/client";

const allActive = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

function optional(value: string) {
  return value.trim() || undefined;
}

function localDateTime(value: string) {
  const date = new Date(value);
  const offset = date.getTimezoneOffset() * 60_000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

function Frame({
  title,
  description,
  pending,
  onOpenChange,
  children,
}: {
  title: string;
  description: string;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
}) {
  return (
    <Dialog.Root open onOpenChange={(open) => !pending && onOpenChange(open)}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-[60] bg-slate-950/60" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-[70] max-h-[92vh] w-[calc(100%-2rem)] max-w-2xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="flex items-start justify-between gap-4 border-b border-slate-200 px-5 py-4 sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                {title}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                {description}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close draft editor"
                className="grid size-10 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

function MutationError({
  error,
  reload,
}: {
  error: Error;
  reload: () => void;
}) {
  return (
    <div
      role="alert"
      className="mb-5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
    >
      <p>{error.message}</p>
      {error instanceof ApiError && error.status === 409 ? (
        <Button
          type="button"
          size="sm"
          variant="secondary"
          className="mt-3"
          onClick={reload}
        >
          Reload latest data
        </Button>
      ) : null}
    </div>
  );
}

export function PurchaseOrderHeaderDialog({
  order,
  onOpenChange,
}: {
  order: PurchaseOrder;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<PurchaseOrderHeaderValues>({
    resolver: zodResolver(purchaseOrderHeaderSchema),
    defaultValues: {
      purchase_order_no: order.purchase_order_no,
      ordered_at: localDateTime(order.ordered_at),
      expected_arrival_at: order.expected_arrival_at
        ? localDateTime(order.expected_arrival_at)
        : "",
      notes: order.notes ?? "",
    },
  });
  const save = useMutation({
    mutationFn: (values: PurchaseOrderHeaderValues) =>
      updatePurchaseOrder(order.purchase_order_id, {
        expected_version: order.version_no,
        purchase_order_no: values.purchase_order_no.trim(),
        ordered_at: new Date(values.ordered_at).toISOString(),
        expected_arrival_at: values.expected_arrival_at
          ? new Date(values.expected_arrival_at).toISOString()
          : undefined,
        notes: optional(values.notes),
      }),
    onSuccess: async (updated) => {
      queryClient.setQueryData(
        purchaseOrderKeys.detail(order.purchase_order_id),
        updated,
      );
      await queryClient.invalidateQueries({
        queryKey: purchaseOrderKeys.lists(),
      });
      toast.success("Purchase order header updated.");
      onOpenChange(false);
    },
  });

  return (
    <Frame
      title="Edit draft header"
      description="Owner, warehouse, supplier, and business date remain fixed."
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      <form
        className="p-5 sm:p-6"
        onSubmit={handleSubmit((values) => save.mutate(values))}
      >
        {save.error ? (
          <MutationError
            error={save.error}
            reload={() => window.location.reload()}
          />
        ) : null}
        <div className="grid gap-5 sm:grid-cols-2">
          <div className="sm:col-span-2">
            <FormField
              label="PO number"
              htmlFor="edit-po-number"
              required
              error={errors.purchase_order_no?.message}
            >
              <Input
                id="edit-po-number"
                autoFocus
                invalid={Boolean(errors.purchase_order_no)}
                {...register("purchase_order_no")}
              />
            </FormField>
          </div>
          <FormField
            label="Ordered at"
            htmlFor="edit-po-ordered"
            required
            error={errors.ordered_at?.message}
          >
            <Input
              id="edit-po-ordered"
              type="datetime-local"
              invalid={Boolean(errors.ordered_at)}
              {...register("ordered_at")}
            />
          </FormField>
          <FormField
            label="Expected arrival"
            htmlFor="edit-po-arrival"
            error={errors.expected_arrival_at?.message}
          >
            <Input
              id="edit-po-arrival"
              type="datetime-local"
              invalid={Boolean(errors.expected_arrival_at)}
              {...register("expected_arrival_at")}
            />
          </FormField>
          <div className="sm:col-span-2">
            <FormField
              label="Notes"
              htmlFor="edit-po-notes"
              error={errors.notes?.message}
            >
              <textarea
                id="edit-po-notes"
                rows={4}
                className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                {...register("notes")}
              />
            </FormField>
          </div>
        </div>
        <div className="mt-6 flex justify-end gap-2 border-t border-slate-200 pt-5">
          <Dialog.Close asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </Dialog.Close>
          <Button type="submit" disabled={save.isPending}>
            {save.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            Save changes
          </Button>
        </div>
      </form>
    </Frame>
  );
}

type LineValues = PurchaseOrderFormValues["lines"][number];

function lineValues(line?: PurchaseOrderLine): LineValues {
  return line
    ? {
        item_id: line.item_id,
        ordered_qty: line.ordered_qty,
        uom_id: line.uom_id,
        vendor_item_code: line.vendor_item_code ?? "",
        expected_lot_no: line.expected_lot_no ?? "",
        expected_expiry_date: line.expected_expiry_date ?? "",
        notes: line.notes ?? "",
        over_receipt_tolerance_pct: line.over_receipt_tolerance_pct,
        under_receipt_tolerance_pct: line.under_receipt_tolerance_pct,
      }
    : { ...emptyPurchaseOrderLine };
}

export function PurchaseOrderLineDialog({
  order,
  line,
  onOpenChange,
}: {
  order: PurchaseOrder;
  line?: PurchaseOrderLine;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = Boolean(line);
  const queryClient = useQueryClient();
  const {
    control,
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<LineValues>({
    resolver: zodResolver(purchaseOrderLineSchema),
    defaultValues: lineValues(line),
  });
  const itemId = useWatch({ control, name: "item_id" });
  const items = useQuery({
    queryKey: itemCatalogKeys.itemList({
      ...allActive,
      ownerId: order.owner_id,
      categoryId: "",
    }),
    queryFn: () =>
      listItems({ ...allActive, ownerId: order.owner_id, categoryId: "" }),
    enabled: !editing,
  });
  const itemDetail = useQuery({
    queryKey: itemCatalogKeys.item(itemId || "none"),
    queryFn: () => getItem(itemId),
    enabled: Boolean(itemId) && !editing,
  });
  const uoms = useQuery({
    queryKey: itemCatalogKeys.uomList(allActive),
    queryFn: () => listUOMs(allActive),
    enabled: !editing,
  });
  const uomById = new Map(
    (uoms.data?.items ?? []).map((uom) => [uom.uom_id, uom]),
  );
  const receivingUOMs = (itemDetail.data?.uoms ?? [])
    .filter((itemUOM) => itemUOM.is_active && itemUOM.is_receiving_uom)
    .map((itemUOM) => {
      const uom = uomById.get(itemUOM.uom_id);
      return {
        value: itemUOM.uom_id,
        label: uom ? `${uom.name} (${uom.code})` : itemUOM.uom_id,
      };
    });
  const save = useMutation({
    mutationFn: (values: LineValues) =>
      editing && line
        ? updatePurchaseOrderLine(
            order.purchase_order_id,
            line.purchase_order_line_id,
            {
              expected_version: order.version_no,
              ordered_qty: values.ordered_qty.trim(),
              vendor_item_code: optional(values.vendor_item_code),
              expected_lot_no: optional(values.expected_lot_no),
              expected_expiry_date: optional(values.expected_expiry_date),
              notes: optional(values.notes),
              over_receipt_tolerance_pct: optional(
                values.over_receipt_tolerance_pct,
              ),
              under_receipt_tolerance_pct: optional(
                values.under_receipt_tolerance_pct,
              ),
            },
          )
        : addPurchaseOrderLine(order.purchase_order_id, order.version_no, {
            item_id: values.item_id,
            uom_id: values.uom_id,
            ordered_qty: values.ordered_qty.trim(),
            vendor_item_code: optional(values.vendor_item_code),
            expected_lot_no: optional(values.expected_lot_no),
            expected_expiry_date: optional(values.expected_expiry_date),
            notes: optional(values.notes),
            over_receipt_tolerance_pct: optional(
              values.over_receipt_tolerance_pct,
            ),
            under_receipt_tolerance_pct: optional(
              values.under_receipt_tolerance_pct,
            ),
          }),
    onSuccess: async (updated) => {
      queryClient.setQueryData(
        purchaseOrderKeys.detail(order.purchase_order_id),
        updated,
      );
      await queryClient.invalidateQueries({
        queryKey: purchaseOrderKeys.lists(),
      });
      toast.success(
        editing ? "Purchase order line updated." : "Purchase order line added.",
      );
      onOpenChange(false);
    },
  });

  return (
    <Frame
      title={editing ? "Edit draft line" : "Add draft line"}
      description={
        editing
          ? "The item and receiving UOM stay fixed; update the commercial values."
          : "Add an active owner item with an allowed receiving UOM."
      }
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      <form
        className="p-5 sm:p-6"
        onSubmit={handleSubmit((values) => save.mutate(values))}
      >
        {save.error ? (
          <MutationError
            error={save.error}
            reload={() => window.location.reload()}
          />
        ) : null}
        <div className="grid gap-5 sm:grid-cols-2">
          <FormField
            label="Item"
            htmlFor="edit-line-item"
            required
            error={errors.item_id?.message}
          >
            {editing ? (
              <Input
                id="edit-line-item"
                value={`${line?.item_name} (${line?.item_code})`}
                readOnly
              />
            ) : (
              <Controller
                control={control}
                name="item_id"
                render={({ field }) => (
                  <Select
                    id="edit-line-item"
                    ariaLabel="Purchase order item"
                    value={field.value}
                    options={(items.data?.items ?? []).map((item) => ({
                      value: item.item_id,
                      label: `${item.name} (${item.code})`,
                    }))}
                    placeholder="Select item"
                    className="mt-2"
                    onValueChange={(value) => {
                      field.onChange(value);
                      setValue("uom_id", "");
                    }}
                  />
                )}
              />
            )}
          </FormField>
          <FormField
            label="Receiving UOM"
            htmlFor="edit-line-uom"
            required
            error={errors.uom_id?.message}
          >
            {editing ? (
              <Input id="edit-line-uom" value={line?.uom_code ?? ""} readOnly />
            ) : (
              <Controller
                control={control}
                name="uom_id"
                render={({ field }) => (
                  <Select
                    id="edit-line-uom"
                    ariaLabel="Receiving UOM"
                    value={field.value}
                    options={receivingUOMs}
                    placeholder={
                      !itemId
                        ? "Select item first"
                        : itemDetail.isPending
                          ? "Loading UOMs…"
                          : "Select receiving UOM"
                    }
                    disabled={!itemId || itemDetail.isPending}
                    className="mt-2"
                    onValueChange={field.onChange}
                  />
                )}
              />
            )}
          </FormField>
          <FormField
            label="Ordered quantity"
            htmlFor="edit-line-qty"
            required
            error={errors.ordered_qty?.message}
          >
            <Input
              id="edit-line-qty"
              inputMode="decimal"
              invalid={Boolean(errors.ordered_qty)}
              {...register("ordered_qty")}
            />
          </FormField>
          <FormField
            label="Vendor item code"
            htmlFor="edit-line-vendor"
            error={errors.vendor_item_code?.message}
          >
            <Input id="edit-line-vendor" {...register("vendor_item_code")} />
          </FormField>
          <FormField
            label="Expected lot"
            htmlFor="edit-line-lot"
            error={errors.expected_lot_no?.message}
          >
            <Input id="edit-line-lot" {...register("expected_lot_no")} />
          </FormField>
          <FormField label="Expected expiry" htmlFor="edit-line-expiry">
            <Input
              id="edit-line-expiry"
              type="date"
              {...register("expected_expiry_date")}
            />
          </FormField>
          <FormField
            label="Over tolerance (%)"
            htmlFor="edit-line-over"
            error={errors.over_receipt_tolerance_pct?.message}
          >
            <Input
              id="edit-line-over"
              inputMode="decimal"
              {...register("over_receipt_tolerance_pct")}
            />
          </FormField>
          <FormField
            label="Under tolerance (%)"
            htmlFor="edit-line-under"
            error={errors.under_receipt_tolerance_pct?.message}
          >
            <Input
              id="edit-line-under"
              inputMode="decimal"
              {...register("under_receipt_tolerance_pct")}
            />
          </FormField>
        </div>
        <div className="mt-6 flex justify-end gap-2 border-t border-slate-200 pt-5">
          <Dialog.Close asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </Dialog.Close>
          <Button type="submit" disabled={save.isPending}>
            {save.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            {editing ? "Save line" : "Add line"}
          </Button>
        </div>
      </form>
    </Frame>
  );
}
