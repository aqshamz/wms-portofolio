"use client";

import { useEffect, type ReactNode } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, X } from "lucide-react";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { itemCatalogKeys } from "@/features/item-catalog/item-catalog-api";
import type { CatalogItem } from "@/features/item-catalog/item-catalog-types";
import {
  createHandlingUnitType,
  createItemBarcode,
  createItemUOM,
  createUOM,
  getHandlingUnitType,
  getUOM,
  unitsPackagingKeys,
  updateHandlingUnitType,
  updateItemBarcode,
  updateItemUOM,
  updateUOM,
} from "@/features/units-packaging/units-packaging-api";
import {
  barcodeFormSchema,
  emptyBarcodeForm,
  emptyHandlingUnitForm,
  emptyItemUOMForm,
  emptyUOMForm,
  handlingUnitFormSchema,
  itemUOMFormSchema,
  uomFormSchema,
  type BarcodeFormValues,
  type HandlingUnitFormValues,
  type ItemUOMFormValues,
  type UOMFormValues,
} from "@/features/units-packaging/units-packaging-schema";
import type {
  HandlingUnitType,
  ItemBarcode,
  ItemUOM,
  UOM,
} from "@/features/units-packaging/units-packaging-types";

function optional(value: string) {
  const normalized = value.trim();
  return normalized || undefined;
}

function DialogFrame({
  title,
  description,
  pending,
  onOpenChange,
  children,
  wide,
}: {
  title: string;
  description: string;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
  wide?: boolean;
}) {
  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!pending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content
          className={`fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none ${wide ? "max-w-3xl" : "max-w-lg"}`}
        >
          <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
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
                aria-label="Close dialog"
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
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

export function UOMDialog({
  uom,
  onOpenChange,
}: {
  uom?: UOM;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = uom !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: unitsPackagingKeys.uom(uom?.uom_id ?? "new"),
    queryFn: () => getUOM(uom!.uom_id),
    enabled: editing,
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<UOMFormValues>({
    resolver: zodResolver(uomFormSchema),
    defaultValues: emptyUOMForm,
  });
  const save = useMutation({
    mutationFn: (values: UOMFormValues) =>
      editing
        ? updateUOM(uom.uom_id, {
            name: values.name.trim(),
            decimal_scale: values.decimal_scale,
            is_active: values.is_active,
          })
        : createUOM({
            code: values.code.trim().toUpperCase(),
            name: values.name.trim(),
            decimal_scale: values.decimal_scale,
          }),
    onSuccess: async (saved) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: unitsPackagingKeys.uomLists(),
        }),
        queryClient.invalidateQueries({
          queryKey: unitsPackagingKeys.uom(saved.uom_id),
        }),
        queryClient.invalidateQueries({ queryKey: itemCatalogKeys.uomLists() }),
      ]);
      toast.success(editing ? "UOM updated." : "UOM created.");
      onOpenChange(false);
    },
  });
  useEffect(() => {
    const value = detail.data ?? uom;
    reset(
      value
        ? {
            code: value.code,
            name: value.name,
            decimal_scale: value.decimal_scale,
            is_active: value.is_active,
          }
        : emptyUOMForm,
    );
  }, [detail.data, reset, uom]);

  return (
    <DialogFrame
      title={editing ? "Edit UOM" : "Create UOM"}
      description="Define a global unit and the decimal precision allowed for quantities."
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      {editing && detail.isPending ? (
        <Loading label="Loading UOM…" />
      ) : editing && detail.isError ? (
        <LoadError error={detail.error} retry={() => detail.refetch()} />
      ) : (
        <form
          noValidate
          className="space-y-5 p-5 sm:p-6"
          onSubmit={handleSubmit((values) => save.mutate(values))}
        >
          <SaveError error={save.error} />
          <div className="grid gap-5 sm:grid-cols-2">
            <FormField
              label="Code"
              htmlFor="uom-code"
              required
              error={errors.code?.message}
            >
              <Input
                id="uom-code"
                placeholder="EA"
                readOnly={editing}
                autoFocus={!editing}
                invalid={Boolean(errors.code)}
                {...register("code")}
              />
            </FormField>
            <FormField
              label="Name"
              htmlFor="uom-name"
              required
              error={errors.name?.message}
            >
              <Input
                id="uom-name"
                placeholder="Each"
                autoFocus={editing}
                invalid={Boolean(errors.name)}
                {...register("name")}
              />
            </FormField>
            <FormField
              label="Decimal scale"
              htmlFor="uom-scale"
              required
              error={errors.decimal_scale?.message}
            >
              <Input
                id="uom-scale"
                type="number"
                min={0}
                max={6}
                invalid={Boolean(errors.decimal_scale)}
                {...register("decimal_scale", { valueAsNumber: true })}
              />
            </FormField>
            {editing ? (
              <label className="flex min-h-11 items-center gap-3 self-end rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
                <input
                  type="checkbox"
                  className="size-4 accent-cyan-600"
                  {...register("is_active")}
                />
                UOM is active
              </label>
            ) : null}
          </div>
          <FormActions
            pending={save.isPending}
            editing={editing}
            createLabel="Create UOM"
          />
        </form>
      )}
    </DialogFrame>
  );
}

export function HandlingUnitDialog({
  handlingUnit,
  onOpenChange,
}: {
  handlingUnit?: HandlingUnitType;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = handlingUnit !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: unitsPackagingKeys.handling(
      handlingUnit?.handling_unit_type_id ?? "new",
    ),
    queryFn: () => getHandlingUnitType(handlingUnit!.handling_unit_type_id),
    enabled: editing,
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<HandlingUnitFormValues>({
    resolver: zodResolver(handlingUnitFormSchema),
    defaultValues: emptyHandlingUnitForm,
  });
  const save = useMutation({
    mutationFn: (values: HandlingUnitFormValues) =>
      editing
        ? updateHandlingUnitType(handlingUnit.handling_unit_type_id, {
            name: values.name.trim(),
            max_weight: optional(values.max_weight),
            max_volume: optional(values.max_volume),
            is_active: values.is_active,
          })
        : createHandlingUnitType({
            code: values.code.trim().toUpperCase(),
            name: values.name.trim(),
            max_weight: optional(values.max_weight),
            max_volume: optional(values.max_volume),
          }),
    onSuccess: async (saved) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: unitsPackagingKeys.handlingLists(),
        }),
        queryClient.invalidateQueries({
          queryKey: unitsPackagingKeys.handling(saved.handling_unit_type_id),
        }),
      ]);
      toast.success(
        editing ? "Handling unit type updated." : "Handling unit type created.",
      );
      onOpenChange(false);
    },
  });
  useEffect(() => {
    const value = detail.data ?? handlingUnit;
    reset(
      value
        ? {
            code: value.code,
            name: value.name,
            max_weight: value.max_weight ?? "",
            max_volume: value.max_volume ?? "",
            is_active: value.is_active,
          }
        : emptyHandlingUnitForm,
    );
  }, [detail.data, handlingUnit, reset]);

  return (
    <DialogFrame
      title={editing ? "Edit handling unit type" : "Create handling unit type"}
      description="Define reusable container types and their optional capacity limits."
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      {editing && detail.isPending ? (
        <Loading label="Loading handling unit…" />
      ) : editing && detail.isError ? (
        <LoadError error={detail.error} retry={() => detail.refetch()} />
      ) : (
        <form
          noValidate
          className="space-y-5 p-5 sm:p-6"
          onSubmit={handleSubmit((values) => save.mutate(values))}
        >
          <SaveError error={save.error} />
          <div className="grid gap-5 sm:grid-cols-2">
            <FormField
              label="Code"
              htmlFor="handling-code"
              required
              error={errors.code?.message}
            >
              <Input
                id="handling-code"
                placeholder="PALLET"
                readOnly={editing}
                autoFocus={!editing}
                invalid={Boolean(errors.code)}
                {...register("code")}
              />
            </FormField>
            <FormField
              label="Name"
              htmlFor="handling-name"
              required
              error={errors.name?.message}
            >
              <Input
                id="handling-name"
                placeholder="Pallet"
                autoFocus={editing}
                invalid={Boolean(errors.name)}
                {...register("name")}
              />
            </FormField>
            <FormField
              label="Maximum weight"
              htmlFor="handling-weight"
              error={errors.max_weight?.message}
            >
              <Input
                id="handling-weight"
                inputMode="decimal"
                placeholder="Optional"
                invalid={Boolean(errors.max_weight)}
                {...register("max_weight")}
              />
            </FormField>
            <FormField
              label="Maximum volume"
              htmlFor="handling-volume"
              error={errors.max_volume?.message}
            >
              <Input
                id="handling-volume"
                inputMode="decimal"
                placeholder="Optional"
                invalid={Boolean(errors.max_volume)}
                {...register("max_volume")}
              />
            </FormField>
            {editing ? (
              <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800 sm:col-span-2">
                <input
                  type="checkbox"
                  className="size-4 accent-cyan-600"
                  {...register("is_active")}
                />
                Handling unit type is active
              </label>
            ) : null}
          </div>
          <FormActions
            pending={save.isPending}
            editing={editing}
            createLabel="Create handling unit"
          />
        </form>
      )}
    </DialogFrame>
  );
}

export function ItemUOMDialog({
  item,
  itemUOM,
  itemUOMs,
  uoms,
  onOpenChange,
}: {
  item: CatalogItem;
  itemUOM?: ItemUOM;
  itemUOMs: ItemUOM[];
  uoms: UOM[];
  onOpenChange: (open: boolean) => void;
}) {
  const editing = itemUOM !== undefined;
  const isBase = itemUOM?.uom_id === item.base_uom_id;
  const queryClient = useQueryClient();
  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ItemUOMFormValues>({
    resolver: zodResolver(itemUOMFormSchema),
    defaultValues: emptyItemUOMForm,
  });
  const save = useMutation({
    mutationFn: (values: ItemUOMFormValues) => {
      const profile = {
        conversion_to_base: values.conversion_to_base.trim(),
        length: optional(values.length),
        width: optional(values.width),
        height: optional(values.height),
        weight: optional(values.weight),
        is_receiving_uom: values.is_receiving_uom,
        is_picking_uom: values.is_picking_uom,
      };
      return editing
        ? updateItemUOM(item.item_id, itemUOM.item_uom_id, {
            ...profile,
            is_active: isBase ? true : values.is_active,
          })
        : createItemUOM(item.item_id, { uom_id: values.uom_id, ...profile });
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: unitsPackagingKeys.itemUOMs(item.item_id),
        }),
        queryClient.invalidateQueries({
          queryKey: itemCatalogKeys.item(item.item_id),
        }),
      ]);
      toast.success(editing ? "Item UOM updated." : "Item UOM added.");
      onOpenChange(false);
    },
  });
  useEffect(() => {
    reset(
      itemUOM
        ? {
            uom_id: itemUOM.uom_id,
            conversion_to_base: itemUOM.conversion_to_base,
            length: itemUOM.length ?? "",
            width: itemUOM.width ?? "",
            height: itemUOM.height ?? "",
            weight: itemUOM.weight ?? "",
            is_receiving_uom: itemUOM.is_receiving_uom,
            is_picking_uom: itemUOM.is_picking_uom,
            is_active: itemUOM.is_active,
          }
        : emptyItemUOMForm,
    );
  }, [itemUOM, reset]);
  const assigned = new Set(itemUOMs.map((value) => value.uom_id));
  const options = uoms
    .filter((uom) => uom.is_active && !assigned.has(uom.uom_id))
    .map((uom) => ({ value: uom.uom_id, label: `${uom.name} (${uom.code})` }));
  const selectedUOM = uoms.find((uom) => uom.uom_id === itemUOM?.uom_id);

  return (
    <DialogFrame
      title={editing ? "Edit item UOM" : "Add item UOM"}
      description={`${item.name} (${item.code}) · conversions are measured against its base UOM.`}
      pending={save.isPending}
      onOpenChange={onOpenChange}
      wide
    >
      <form
        noValidate
        className="p-5 sm:p-6"
        onSubmit={handleSubmit((values) => save.mutate(values))}
      >
        <SaveError error={save.error} />
        {isBase ? (
          <p className="mb-5 rounded-xl bg-cyan-50 p-3 text-sm text-cyan-900">
            This is the base UOM. It must remain active with conversion 1.
          </p>
        ) : null}
        {!item.is_active ? (
          <p className="mb-5 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
            This item is inactive. Existing packaging can be reviewed or
            deactivated, but it cannot be added or reactivated.
          </p>
        ) : null}
        <div className="grid gap-5 sm:grid-cols-2">
          <div className="sm:col-span-2">
            <FormField
              label="Unit of measure"
              htmlFor="item-uom-unit"
              required
              error={errors.uom_id?.message}
            >
              {editing ? (
                <Input
                  id="item-uom-unit"
                  readOnly
                  value={
                    selectedUOM
                      ? `${selectedUOM.name} (${selectedUOM.code})`
                      : itemUOM.uom_id
                  }
                />
              ) : (
                <Controller
                  control={control}
                  name="uom_id"
                  render={({ field }) => (
                    <Select
                      id="item-uom-unit"
                      ariaLabel="Unit of measure"
                      value={field.value}
                      options={options}
                      onValueChange={field.onChange}
                      placeholder={
                        options.length
                          ? "Select a UOM"
                          : "All active UOMs assigned"
                      }
                      invalid={Boolean(errors.uom_id)}
                      className="mt-2"
                    />
                  )}
                />
              )}
            </FormField>
          </div>
          <FormField
            label="Conversion to base"
            htmlFor="item-uom-conversion"
            required
            error={errors.conversion_to_base?.message}
          >
            <Input
              id="item-uom-conversion"
              inputMode="decimal"
              readOnly={isBase}
              placeholder="Example: 12"
              invalid={Boolean(errors.conversion_to_base)}
              {...register("conversion_to_base")}
            />
          </FormField>
          <div />
          <FormField
            label="Length"
            htmlFor="item-uom-length"
            error={errors.length?.message}
          >
            <Input
              id="item-uom-length"
              inputMode="decimal"
              placeholder="Optional"
              invalid={Boolean(errors.length)}
              {...register("length")}
            />
          </FormField>
          <FormField
            label="Width"
            htmlFor="item-uom-width"
            error={errors.width?.message}
          >
            <Input
              id="item-uom-width"
              inputMode="decimal"
              placeholder="Optional"
              invalid={Boolean(errors.width)}
              {...register("width")}
            />
          </FormField>
          <FormField
            label="Height"
            htmlFor="item-uom-height"
            error={errors.height?.message}
          >
            <Input
              id="item-uom-height"
              inputMode="decimal"
              placeholder="Optional"
              invalid={Boolean(errors.height)}
              {...register("height")}
            />
          </FormField>
          <FormField
            label="Weight"
            htmlFor="item-uom-weight"
            error={errors.weight?.message}
          >
            <Input
              id="item-uom-weight"
              inputMode="decimal"
              placeholder="Optional"
              invalid={Boolean(errors.weight)}
              {...register("weight")}
            />
          </FormField>
          <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
            <input
              type="checkbox"
              className="size-4 accent-cyan-600"
              {...register("is_receiving_uom")}
            />
            Receiving UOM
          </label>
          <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
            <input
              type="checkbox"
              className="size-4 accent-cyan-600"
              {...register("is_picking_uom")}
            />
            Picking UOM
          </label>
          {editing && !isBase ? (
            <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800 sm:col-span-2">
              <input
                type="checkbox"
                className="size-4 accent-cyan-600"
                {...register("is_active")}
              />
              Item UOM is active
            </label>
          ) : null}
        </div>
        <FormActions
          pending={save.isPending}
          editing={editing}
          createLabel="Add item UOM"
          disabled={!editing && (!item.is_active || options.length === 0)}
        />
      </form>
    </DialogFrame>
  );
}

export function BarcodeDialog({
  item,
  barcode,
  itemUOMs,
  uoms,
  onOpenChange,
}: {
  item: CatalogItem;
  barcode?: ItemBarcode;
  itemUOMs: ItemUOM[];
  uoms: UOM[];
  onOpenChange: (open: boolean) => void;
}) {
  const editing = barcode !== undefined;
  const queryClient = useQueryClient();
  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<BarcodeFormValues>({
    resolver: zodResolver(barcodeFormSchema),
    defaultValues: emptyBarcodeForm,
  });
  const save = useMutation({
    mutationFn: (values: BarcodeFormValues) =>
      editing
        ? updateItemBarcode(item.item_id, barcode.item_barcode_id, {
            uom_id: optional(values.uom_id),
            barcode: values.barcode.trim(),
            is_active: values.is_active,
          })
        : createItemBarcode(item.item_id, {
            uom_id: optional(values.uom_id),
            barcode: values.barcode.trim(),
            is_primary: values.is_primary,
          }),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: unitsPackagingKeys.barcodes(item.item_id),
        }),
        queryClient.invalidateQueries({
          queryKey: itemCatalogKeys.item(item.item_id),
        }),
      ]);
      toast.success(editing ? "Barcode updated." : "Barcode created.");
      onOpenChange(false);
    },
  });
  useEffect(() => {
    reset(
      barcode
        ? {
            uom_id: barcode.uom_id ?? "",
            barcode: barcode.barcode,
            is_primary: barcode.is_primary,
            is_active: barcode.is_active,
          }
        : emptyBarcodeForm,
    );
  }, [barcode, reset]);
  const options = [
    { value: "item-wide", label: "Item-wide barcode" },
    ...itemUOMs.map((unit) => {
      const uom = uoms.find((value) => value.uom_id === unit.uom_id);
      return {
        value: unit.uom_id,
        label: `${uom?.name ?? unit.uom_id} (${uom?.code ?? "UOM"})${unit.is_active ? "" : " · Inactive"}`,
        disabled: !unit.is_active,
      };
    }),
  ];

  return (
    <DialogFrame
      title={editing ? "Edit barcode" : "Create barcode"}
      description={`${item.name} (${item.code}) · optionally bind the code to an active item UOM.`}
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      <form
        noValidate
        className="space-y-5 p-5 sm:p-6"
        onSubmit={handleSubmit((values) => save.mutate(values))}
      >
        <SaveError error={save.error} />
        {!item.is_active ? (
          <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
            This item is inactive. Barcodes can be reviewed or deactivated, but
            not created or reactivated.
          </p>
        ) : null}
        <FormField
          label="Barcode"
          htmlFor="barcode-value"
          required
          error={errors.barcode?.message}
        >
          <Input
            id="barcode-value"
            placeholder="Scan or enter barcode"
            autoFocus
            invalid={Boolean(errors.barcode)}
            {...register("barcode")}
          />
        </FormField>
        <FormField
          label="Applies to"
          htmlFor="barcode-uom"
          error={errors.uom_id?.message}
        >
          <Controller
            control={control}
            name="uom_id"
            render={({ field }) => (
              <Select
                id="barcode-uom"
                ariaLabel="Barcode UOM"
                value={field.value || "item-wide"}
                options={options}
                onValueChange={(value) =>
                  field.onChange(value === "item-wide" ? "" : value)
                }
                className="mt-2"
              />
            )}
          />
        </FormField>
        {!editing ? (
          <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
            <input
              type="checkbox"
              className="size-4 accent-cyan-600"
              {...register("is_primary")}
            />
            Make primary barcode
          </label>
        ) : (
          <>
            <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
              <input
                type="checkbox"
                className="size-4 accent-cyan-600"
                {...register("is_active")}
              />
              Barcode is active
            </label>
            <p className="text-xs text-slate-500">
              Use “Make primary” from the list to change the primary barcode.
            </p>
          </>
        )}
        <FormActions
          pending={save.isPending}
          editing={editing}
          createLabel="Create barcode"
          disabled={!editing && !item.is_active}
        />
      </form>
    </DialogFrame>
  );
}

function Loading({ label }: { label: string }) {
  return (
    <div className="grid min-h-56 place-items-center p-6 text-sm text-slate-600">
      <div className="text-center">
        <LoaderCircle className="mx-auto mb-3 size-6 animate-spin text-cyan-700" />
        {label}
      </div>
    </div>
  );
}

function LoadError({ error, retry }: { error: Error; retry: () => void }) {
  return (
    <div className="p-5 sm:p-6">
      <p
        role="alert"
        className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
      >
        {error.message}
      </p>
      <Button className="mt-4" variant="secondary" onClick={retry}>
        Try again
      </Button>
    </div>
  );
}

function SaveError({ error }: { error?: Error | null }) {
  return error ? (
    <p
      role="alert"
      className="mb-5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
    >
      {error.message}
    </p>
  ) : null;
}

function FormActions({
  pending,
  editing,
  createLabel,
  disabled,
}: {
  pending: boolean;
  editing: boolean;
  createLabel: string;
  disabled?: boolean;
}) {
  return (
    <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
      <Dialog.Close asChild>
        <Button type="button" variant="secondary">
          Cancel
        </Button>
      </Dialog.Close>
      <Button type="submit" disabled={pending || disabled}>
        {pending ? <LoaderCircle className="size-4 animate-spin" /> : null}
        {pending ? "Saving…" : editing ? "Save changes" : createLabel}
      </Button>
    </div>
  );
}
