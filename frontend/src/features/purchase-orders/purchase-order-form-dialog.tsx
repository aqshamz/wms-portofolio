"use client";

import { useEffect } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, Plus, Trash2, X } from "lucide-react";
import {
  Controller,
  useFieldArray,
  useForm,
  useWatch,
  type Control,
  type FieldErrors,
  type UseFormRegister,
  type UseFormSetValue,
} from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  businessPartnerKeys,
  listBusinessPartners,
} from "@/features/business-partners/business-partner-api";
import {
  getItem,
  itemCatalogKeys,
  listItems,
  listUOMs,
} from "@/features/item-catalog/item-catalog-api";
import type {
  CatalogItem,
  UOM,
} from "@/features/item-catalog/item-catalog-types";
import {
  createPurchaseOrder,
  purchaseOrderKeys,
} from "@/features/purchase-orders/purchase-order-api";
import {
  emptyPurchaseOrderLine,
  purchaseOrderFormSchema,
  type PurchaseOrderFormValues,
} from "@/features/purchase-orders/purchase-order-schema";

const allActive = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

function localDateTime(date = new Date()) {
  const offset = date.getTimezoneOffset() * 60_000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

function optional(value: string) {
  return value.trim() || undefined;
}

function LineEditor({
  index,
  items,
  uoms,
  control,
  register,
  setValue,
  errors,
  removable,
  onRemove,
}: {
  index: number;
  items: CatalogItem[];
  uoms: UOM[];
  control: Control<PurchaseOrderFormValues>;
  register: UseFormRegister<PurchaseOrderFormValues>;
  setValue: UseFormSetValue<PurchaseOrderFormValues>;
  errors?: FieldErrors<PurchaseOrderFormValues["lines"][number]>;
  removable: boolean;
  onRemove: () => void;
}) {
  const itemId = useWatch({ control, name: `lines.${index}.item_id` });
  const itemDetail = useQuery({
    queryKey: itemCatalogKeys.item(itemId || "none"),
    queryFn: () => getItem(itemId),
    enabled: Boolean(itemId),
  });
  const uomById = new Map(uoms.map((uom) => [uom.uom_id, uom]));
  const receivingUOMs = (itemDetail.data?.uoms ?? [])
    .filter((itemUOM) => itemUOM.is_active && itemUOM.is_receiving_uom)
    .map((itemUOM) => {
      const uom = uomById.get(itemUOM.uom_id);
      return {
        value: itemUOM.uom_id,
        label: uom
          ? `${uom.name} (${uom.code})`
          : `UOM ${itemUOM.uom_id.slice(0, 8)}`,
      };
    });

  return (
    <section className="rounded-xl border border-slate-200 p-4">
      <div className="flex items-center justify-between gap-3">
        <h3 className="font-bold text-slate-950">Line {index + 1}</h3>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          disabled={!removable}
          className="text-rose-700"
          onClick={onRemove}
        >
          <Trash2 className="size-4" />
          Remove
        </Button>
      </div>
      <div className="mt-4 grid gap-4 sm:grid-cols-2">
        <FormField label="Item" htmlFor={`po-item-${index}`} required>
          <Controller
            control={control}
            name={`lines.${index}.item_id`}
            render={({ field, fieldState }) => (
              <>
                <Select
                  id={`po-item-${index}`}
                  ariaLabel={`Item for line ${index + 1}`}
                  value={field.value}
                  options={items.map((item) => ({
                    value: item.item_id,
                    label: `${item.name} (${item.code})`,
                  }))}
                  placeholder="Select item"
                  invalid={Boolean(fieldState.error)}
                  className="mt-2"
                  onValueChange={(value) => {
                    field.onChange(value);
                    setValue(`lines.${index}.uom_id`, "");
                  }}
                />
                {fieldState.error ? (
                  <p className="mt-1.5 text-xs text-rose-700">
                    {fieldState.error.message}
                  </p>
                ) : null}
              </>
            )}
          />
        </FormField>
        <FormField label="Receiving UOM" htmlFor={`po-uom-${index}`} required>
          <Controller
            control={control}
            name={`lines.${index}.uom_id`}
            render={({ field, fieldState }) => (
              <>
                <Select
                  id={`po-uom-${index}`}
                  ariaLabel={`Receiving UOM for line ${index + 1}`}
                  value={field.value}
                  options={receivingUOMs}
                  placeholder={
                    !itemId
                      ? "Select an item first"
                      : itemDetail.isPending
                        ? "Loading receiving UOMs…"
                        : receivingUOMs.length
                          ? "Select receiving UOM"
                          : "No receiving UOM configured"
                  }
                  disabled={!itemId || itemDetail.isPending}
                  invalid={Boolean(fieldState.error)}
                  className="mt-2"
                  onValueChange={field.onChange}
                />
                {fieldState.error ? (
                  <p className="mt-1.5 text-xs text-rose-700">
                    {fieldState.error.message}
                  </p>
                ) : null}
              </>
            )}
          />
        </FormField>
        <FormField
          label="Ordered quantity"
          htmlFor={`po-qty-${index}`}
          required
          error={errors?.ordered_qty?.message}
        >
          <Input
            id={`po-qty-${index}`}
            inputMode="decimal"
            placeholder="0.000000"
            invalid={Boolean(errors?.ordered_qty)}
            {...register(`lines.${index}.ordered_qty`)}
          />
        </FormField>
        <FormField
          label="Vendor item code"
          htmlFor={`po-vendor-code-${index}`}
          error={errors?.vendor_item_code?.message}
        >
          <Input
            id={`po-vendor-code-${index}`}
            placeholder="Optional supplier SKU"
            {...register(`lines.${index}.vendor_item_code`)}
          />
        </FormField>
        <FormField label="Expected lot" htmlFor={`po-lot-${index}`}>
          <Input
            id={`po-lot-${index}`}
            placeholder="Optional"
            {...register(`lines.${index}.expected_lot_no`)}
          />
        </FormField>
        <FormField label="Expected expiry" htmlFor={`po-expiry-${index}`}>
          <Input
            id={`po-expiry-${index}`}
            type="date"
            {...register(`lines.${index}.expected_expiry_date`)}
          />
        </FormField>
        <FormField
          label="Over-receipt tolerance (%)"
          htmlFor={`po-over-${index}`}
          error={errors?.over_receipt_tolerance_pct?.message}
        >
          <Input
            id={`po-over-${index}`}
            inputMode="decimal"
            invalid={Boolean(errors?.over_receipt_tolerance_pct)}
            {...register(`lines.${index}.over_receipt_tolerance_pct`)}
          />
        </FormField>
        <FormField
          label="Under-receipt tolerance (%)"
          htmlFor={`po-under-${index}`}
          error={errors?.under_receipt_tolerance_pct?.message}
        >
          <Input
            id={`po-under-${index}`}
            inputMode="decimal"
            invalid={Boolean(errors?.under_receipt_tolerance_pct)}
            {...register(`lines.${index}.under_receipt_tolerance_pct`)}
          />
        </FormField>
      </div>
    </section>
  );
}

export function PurchaseOrderFormDialog({
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
  const suppliers = useQuery({
    queryKey: businessPartnerKeys.partnerList({
      ...allActive,
      ownerId,
      partnerTypeCode: "SUPPLIER",
    }),
    queryFn: () =>
      listBusinessPartners({
        ...allActive,
        ownerId,
        partnerTypeCode: "SUPPLIER",
      }),
  });
  const factories = useQuery({
    queryKey: businessPartnerKeys.partnerList({
      ...allActive,
      ownerId,
      partnerTypeCode: "FACTORY",
    }),
    queryFn: () =>
      listBusinessPartners({
        ...allActive,
        ownerId,
        partnerTypeCode: "FACTORY",
      }),
  });
  const vendors = Array.from(
    new Map(
      [...(suppliers.data?.items ?? []), ...(factories.data?.items ?? [])].map(
        (vendor) => [vendor.partner_id, vendor],
      ),
    ).values(),
  );
  const items = useQuery({
    queryKey: itemCatalogKeys.itemList({
      ...allActive,
      ownerId,
      categoryId: "",
    }),
    queryFn: () => listItems({ ...allActive, ownerId, categoryId: "" }),
  });
  const uoms = useQuery({
    queryKey: itemCatalogKeys.uomList(allActive),
    queryFn: () => listUOMs(allActive),
  });
  const {
    control,
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<PurchaseOrderFormValues>({
    resolver: zodResolver(purchaseOrderFormSchema),
    defaultValues: {
      owner_id: ownerId,
      warehouse_id: warehouseId,
      vendor_id: "",
      business_date: new Date().toISOString().slice(0, 10),
      purchase_order_no: "",
      ordered_at: localDateTime(),
      expected_arrival_at: "",
      notes: "",
      lines: [{ ...emptyPurchaseOrderLine }],
    },
  });
  const lines = useFieldArray({ control, name: "lines" });
  const save = useMutation({
    mutationFn: (values: PurchaseOrderFormValues) =>
      createPurchaseOrder({
        owner_id: values.owner_id,
        warehouse_id: values.warehouse_id,
        vendor_id: values.vendor_id,
        business_date: values.business_date,
        purchase_order_no: values.purchase_order_no.trim(),
        ordered_at: new Date(values.ordered_at).toISOString(),
        expected_arrival_at: values.expected_arrival_at
          ? new Date(values.expected_arrival_at).toISOString()
          : undefined,
        notes: optional(values.notes),
        lines: values.lines.map((line) => ({
          item_id: line.item_id,
          ordered_qty: line.ordered_qty.trim(),
          uom_id: line.uom_id,
          vendor_item_code: optional(line.vendor_item_code),
          expected_lot_no: optional(line.expected_lot_no),
          expected_expiry_date: optional(line.expected_expiry_date),
          notes: optional(line.notes),
          over_receipt_tolerance_pct: optional(line.over_receipt_tolerance_pct),
          under_receipt_tolerance_pct: optional(
            line.under_receipt_tolerance_pct,
          ),
        })),
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: purchaseOrderKeys.lists(),
      });
      toast.success("Purchase order draft created.");
      onOpenChange(false);
    },
  });

  useEffect(() => {
    reset((values) => ({
      ...values,
      owner_id: ownerId,
      warehouse_id: warehouseId,
    }));
  }, [ownerId, warehouseId, reset]);

  return (
    <Dialog.Root
      open
      onOpenChange={(open) => !save.isPending && onOpenChange(open)}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[94vh] w-[calc(100%-1.5rem)] max-w-5xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="sticky top-0 z-10 flex items-start justify-between border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                Create purchase order
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                Create a scoped draft with at least one receiving line.
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close purchase order form"
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
                Order header
              </h2>
              <div className="mt-4 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
                <FormField label="Owner" htmlFor="po-owner" required>
                  <Input id="po-owner" value={ownerLabel} readOnly />
                </FormField>
                <FormField label="Warehouse" htmlFor="po-warehouse" required>
                  <Input id="po-warehouse" value={warehouseLabel} readOnly />
                </FormField>
                <FormField
                  label="Supplier"
                  htmlFor="po-supplier"
                  required
                  error={errors.vendor_id?.message}
                >
                  <Controller
                    control={control}
                    name="vendor_id"
                    render={({ field }) => (
                      <Select
                        id="po-supplier"
                        ariaLabel="Supplier"
                        value={field.value}
                        onValueChange={field.onChange}
                        options={vendors.map((vendor) => ({
                          value: vendor.partner_id,
                          label: `${vendor.name} (${vendor.code})`,
                        }))}
                        placeholder={
                          suppliers.isPending || factories.isPending
                            ? "Loading suppliers…"
                            : "Select supplier"
                        }
                        disabled={
                          suppliers.isPending ||
                          factories.isPending ||
                          suppliers.isError ||
                          factories.isError
                        }
                        invalid={Boolean(errors.vendor_id)}
                        className="mt-2"
                      />
                    )}
                  />
                </FormField>
                <FormField
                  label="PO number"
                  htmlFor="po-number"
                  required
                  error={errors.purchase_order_no?.message}
                >
                  <Input
                    id="po-number"
                    placeholder="CLIENT-PO-001"
                    autoFocus
                    invalid={Boolean(errors.purchase_order_no)}
                    {...register("purchase_order_no")}
                  />
                </FormField>
                <FormField
                  label="Business date"
                  htmlFor="po-business-date"
                  required
                  error={errors.business_date?.message}
                >
                  <Input
                    id="po-business-date"
                    type="date"
                    invalid={Boolean(errors.business_date)}
                    {...register("business_date")}
                  />
                </FormField>
                <FormField
                  label="Ordered at"
                  htmlFor="po-ordered-at"
                  required
                  error={errors.ordered_at?.message}
                >
                  <Input
                    id="po-ordered-at"
                    type="datetime-local"
                    invalid={Boolean(errors.ordered_at)}
                    {...register("ordered_at")}
                  />
                </FormField>
                <FormField
                  label="Expected arrival"
                  htmlFor="po-arrival"
                  error={errors.expected_arrival_at?.message}
                >
                  <Input
                    id="po-arrival"
                    type="datetime-local"
                    invalid={Boolean(errors.expected_arrival_at)}
                    {...register("expected_arrival_at")}
                  />
                </FormField>
                <div className="sm:col-span-2">
                  <FormField
                    label="Notes"
                    htmlFor="po-notes"
                    error={errors.notes?.message}
                  >
                    <textarea
                      id="po-notes"
                      rows={3}
                      className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                      {...register("notes")}
                    />
                  </FormField>
                </div>
              </div>
            </section>
            <section className="mt-7 border-t border-slate-200 pt-6">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <h2 className="text-xs font-bold tracking-wide text-slate-500 uppercase">
                    Order lines
                  </h2>
                  <p className="mt-1 text-sm text-slate-600">
                    Only active items and their receiving UOMs are available.
                  </p>
                </div>
                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  onClick={() => lines.append({ ...emptyPurchaseOrderLine })}
                >
                  <Plus className="size-4" />
                  Add line
                </Button>
              </div>
              <div className="mt-4 space-y-4">
                {lines.fields.map((line, index) => (
                  <LineEditor
                    key={line.id}
                    index={index}
                    items={items.data?.items ?? []}
                    uoms={uoms.data?.items ?? []}
                    control={control}
                    register={register}
                    setValue={setValue}
                    errors={errors.lines?.[index]}
                    removable={lines.fields.length > 1}
                    onRemove={() => lines.remove(index)}
                  />
                ))}
              </div>
              {typeof errors.lines?.root?.message === "string" ? (
                <p className="mt-2 text-sm text-rose-700">
                  {errors.lines.root.message}
                </p>
              ) : null}
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
                  suppliers.isPending ||
                  factories.isPending ||
                  items.isPending ||
                  uoms.isPending
                }
              >
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
