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
import {
  createItem,
  createItemCategory,
  getItem,
  getItemCategory,
  itemCatalogKeys,
  updateItem,
  updateItemCategory,
} from "@/features/item-catalog/item-catalog-api";
import {
  emptyItemCategoryForm,
  emptyItemForm,
  itemCategoryFormSchema,
  itemFormSchema,
  type ItemCategoryFormValues,
  type ItemFormValues,
} from "@/features/item-catalog/item-catalog-schema";
import type {
  CatalogItem,
  ItemCategory,
  ItemProfileRequest,
  UOM,
} from "@/features/item-catalog/item-catalog-types";
import { ApiError } from "@/lib/api/client";

function optional(value: string) {
  const normalized = value.trim();
  return normalized || undefined;
}

function optionalDays(value: string) {
  const normalized = value.trim();
  return normalized ? Number(normalized) : undefined;
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

function categoryValues(category: ItemCategory): ItemCategoryFormValues {
  return {
    code: category.code,
    name: category.name,
    parent_category_id: category.parent_category_id ?? "",
    is_active: category.is_active,
  };
}

export function ItemCategoryDialog({
  ownerId,
  ownerLabel,
  category,
  categories,
  onOpenChange,
}: {
  ownerId: string;
  ownerLabel: string;
  category?: ItemCategory;
  categories: ItemCategory[];
  onOpenChange: (open: boolean) => void;
}) {
  const editing = category !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: itemCatalogKeys.category(category?.category_id ?? "new"),
    queryFn: () => getItemCategory(category!.category_id),
    enabled: editing,
  });
  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ItemCategoryFormValues>({
    resolver: zodResolver(itemCategoryFormSchema),
    defaultValues: emptyItemCategoryForm,
  });
  const save = useMutation({
    mutationFn: (values: ItemCategoryFormValues) =>
      editing
        ? updateItemCategory(category.category_id, {
            parent_category_id: optional(values.parent_category_id),
            name: values.name.trim(),
            is_active: values.is_active,
          })
        : createItemCategory({
            owner_id: ownerId,
            parent_category_id: optional(values.parent_category_id),
            code: values.code.trim().toUpperCase(),
            name: values.name.trim(),
          }),
    onSuccess: async (item) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: itemCatalogKeys.categoryLists(),
        }),
        queryClient.invalidateQueries({
          queryKey: itemCatalogKeys.category(item.category_id),
        }),
        queryClient.invalidateQueries({
          queryKey: itemCatalogKeys.itemLists(),
        }),
      ]);
      toast.success(
        editing ? "Item category updated." : "Item category created.",
      );
      onOpenChange(false);
    },
  });

  useEffect(() => {
    reset(
      editing ? categoryValues(detail.data ?? category) : emptyItemCategoryForm,
    );
  }, [detail.data, editing, category, reset]);

  const parentOptions = [
    { value: "root", label: "No parent (top level)" },
    ...categories
      .filter((item) => item.category_id !== category?.category_id)
      .map((item) => ({
        value: item.category_id,
        label: `${item.name} (${item.code})${item.is_active ? "" : " · Inactive"}`,
        disabled: !item.is_active,
      })),
  ];

  return (
    <DialogFrame
      title={editing ? "Edit item category" : "Create item category"}
      description="Categories are owner-specific and may be nested under another active category."
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      {editing && detail.isPending ? (
        <div className="grid min-h-56 place-items-center p-6 text-sm text-slate-600">
          <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
          Loading category…
        </div>
      ) : editing && detail.isError ? (
        <div className="p-5 sm:p-6">
          <p
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {detail.error.message}
          </p>
          <Button
            className="mt-4"
            variant="secondary"
            onClick={() => detail.refetch()}
          >
            Try again
          </Button>
        </div>
      ) : (
        <form
          noValidate
          className="space-y-5 p-5 sm:p-6"
          onSubmit={handleSubmit((values) => save.mutate(values))}
        >
          {save.error ? (
            <p
              role="alert"
              className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
            >
              {save.error.message}
            </p>
          ) : null}
          <FormField label="Owner" htmlFor="category-owner" required>
            <Input id="category-owner" value={ownerLabel} readOnly />
          </FormField>
          <div className="grid gap-5 sm:grid-cols-2">
            <FormField
              label="Code"
              htmlFor="category-code"
              required
              error={errors.code?.message}
            >
              <Input
                id="category-code"
                placeholder="BEVERAGES"
                readOnly={editing}
                autoFocus={!editing}
                invalid={Boolean(errors.code)}
                {...register("code")}
              />
            </FormField>
            <FormField
              label="Name"
              htmlFor="category-name"
              required
              error={errors.name?.message}
            >
              <Input
                id="category-name"
                placeholder="Beverages"
                autoFocus={editing}
                invalid={Boolean(errors.name)}
                {...register("name")}
              />
            </FormField>
            <div className="sm:col-span-2">
              <FormField
                label="Parent category"
                htmlFor="category-parent"
                error={errors.parent_category_id?.message}
              >
                <Controller
                  control={control}
                  name="parent_category_id"
                  render={({ field }) => (
                    <Select
                      id="category-parent"
                      ariaLabel="Parent category"
                      value={field.value || "root"}
                      options={parentOptions}
                      onValueChange={(value) =>
                        field.onChange(value === "root" ? "" : value)
                      }
                      className="mt-2"
                    />
                  )}
                />
              </FormField>
            </div>
            {editing ? (
              <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800 sm:col-span-2">
                <input
                  type="checkbox"
                  className="size-4 accent-cyan-600"
                  {...register("is_active")}
                />
                Category is active
              </label>
            ) : null}
          </div>
          <div className="flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
            <Dialog.Close asChild>
              <Button type="button" variant="secondary">
                Cancel
              </Button>
            </Dialog.Close>
            <Button type="submit" disabled={save.isPending}>
              {save.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              {save.isPending
                ? "Saving…"
                : editing
                  ? "Save changes"
                  : "Create category"}
            </Button>
          </div>
        </form>
      )}
    </DialogFrame>
  );
}

function itemValues(item: CatalogItem): ItemFormValues {
  return {
    code: item.code,
    name: item.name,
    description: item.description ?? "",
    category_id: item.category_id ?? "",
    base_uom_id: item.base_uom_id,
    weight: item.weight ?? "",
    volume: item.volume ?? "",
    lot_controlled: item.lot_controlled,
    serial_controlled: item.serial_controlled,
    shelf_life_days: item.shelf_life_days?.toString() ?? "",
    minimum_receive_days: item.minimum_receive_days?.toString() ?? "",
    is_active: item.is_active,
  };
}

function itemProfile(values: ItemFormValues): ItemProfileRequest {
  return {
    category_id: optional(values.category_id),
    name: values.name.trim(),
    description: optional(values.description),
    weight: optional(values.weight),
    volume: optional(values.volume),
    lot_controlled: values.lot_controlled,
    serial_controlled: values.serial_controlled,
    shelf_life_days: optionalDays(values.shelf_life_days),
    minimum_receive_days: optionalDays(values.minimum_receive_days),
  };
}

export function ItemDialog({
  ownerId,
  ownerLabel,
  item,
  categories,
  uoms,
  onOpenChange,
}: {
  ownerId: string;
  ownerLabel: string;
  item?: CatalogItem;
  categories: ItemCategory[];
  uoms: UOM[];
  onOpenChange: (open: boolean) => void;
}) {
  const editing = item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: itemCatalogKeys.item(item?.item_id ?? "new"),
    queryFn: () => getItem(item!.item_id),
    enabled: editing,
  });
  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ItemFormValues>({
    resolver: zodResolver(itemFormSchema),
    defaultValues: emptyItemForm,
  });
  const save = useMutation({
    mutationFn: (values: ItemFormValues) => {
      const profile = itemProfile(values);
      if (!editing) {
        return createItem({
          owner_id: ownerId,
          code: values.code.trim().toUpperCase(),
          base_uom_id: values.base_uom_id,
          ...profile,
        });
      }
      if (!detail.data) throw new Error("Item details are not loaded.");
      return updateItem(detail.data.item_id, {
        ...profile,
        is_active: values.is_active,
        expected_updated_at: detail.data.updated_at,
      });
    },
    onSuccess: async (saved) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: itemCatalogKeys.itemLists(),
        }),
        queryClient.invalidateQueries({
          queryKey: itemCatalogKeys.item(saved.item_id),
        }),
      ]);
      toast.success(
        editing ? "Item updated." : "Item created with its base UOM.",
      );
      onOpenChange(false);
    },
  });

  useEffect(() => {
    reset(editing ? itemValues(detail.data ?? item) : emptyItemForm);
  }, [detail.data, editing, item, reset]);

  const categoryOptions = [
    { value: "uncategorized", label: "Uncategorized" },
    ...categories.map((category) => ({
      value: category.category_id,
      label: `${category.name} (${category.code})${category.is_active ? "" : " · Inactive"}`,
      disabled: !category.is_active,
    })),
  ];
  const activeUOMOptions = uoms
    .filter((uom) => uom.is_active)
    .map((uom) => ({ value: uom.uom_id, label: `${uom.name} (${uom.code})` }));
  const currentUOM = uoms.find(
    (uom) => uom.uom_id === (detail.data ?? item)?.base_uom_id,
  );

  return (
    <DialogFrame
      title={editing ? "Edit item" : "Create item"}
      description={
        editing
          ? "Update the item profile and inventory controls."
          : "Create the owner SKU and its mandatory base unit of measure."
      }
      pending={save.isPending}
      onOpenChange={onOpenChange}
      wide
    >
      {editing && detail.isPending ? (
        <div className="grid min-h-72 place-items-center p-6 text-sm text-slate-600">
          <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
          Loading item…
        </div>
      ) : editing && detail.isError ? (
        <div className="p-5 sm:p-6">
          <p
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {detail.error.message}
          </p>
          <Button
            className="mt-4"
            variant="secondary"
            onClick={() => detail.refetch()}
          >
            Try again
          </Button>
        </div>
      ) : (
        <form
          noValidate
          className="p-5 sm:p-6"
          onSubmit={handleSubmit((values) => save.mutate(values))}
        >
          {save.error ? (
            <div
              role="alert"
              className="mb-5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
            >
              <p>{save.error.message}</p>
              {editing &&
              save.error instanceof ApiError &&
              save.error.status === 409 ? (
                <Button
                  type="button"
                  size="sm"
                  variant="secondary"
                  className="mt-3"
                  onClick={async () => {
                    save.reset();
                    await detail.refetch();
                  }}
                >
                  Reload latest data
                </Button>
              ) : null}
            </div>
          ) : null}

          <section>
            <h2 className="text-xs font-bold tracking-wide text-slate-500 uppercase">
              Item profile
            </h2>
            <div className="mt-4 grid gap-5 sm:grid-cols-2">
              <div className="sm:col-span-2">
                <FormField label="Owner" htmlFor="item-owner" required>
                  <Input id="item-owner" value={ownerLabel} readOnly />
                </FormField>
              </div>
              <FormField
                label="Item code"
                htmlFor="item-code"
                required
                error={errors.code?.message}
              >
                <Input
                  id="item-code"
                  placeholder="SKU_001"
                  readOnly={editing}
                  autoFocus={!editing}
                  invalid={Boolean(errors.code)}
                  {...register("code")}
                />
              </FormField>
              <FormField
                label="Item name"
                htmlFor="item-name"
                required
                error={errors.name?.message}
              >
                <Input
                  id="item-name"
                  placeholder="Item name"
                  autoFocus={editing}
                  invalid={Boolean(errors.name)}
                  {...register("name")}
                />
              </FormField>
              <div className="sm:col-span-2">
                <FormField
                  label="Description"
                  htmlFor="item-description"
                  error={errors.description?.message}
                >
                  <textarea
                    id="item-description"
                    rows={3}
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                    {...register("description")}
                  />
                </FormField>
              </div>
              <FormField
                label="Category"
                htmlFor="item-category"
                error={errors.category_id?.message}
              >
                <Controller
                  control={control}
                  name="category_id"
                  render={({ field }) => (
                    <Select
                      id="item-category"
                      ariaLabel="Item category"
                      value={field.value || "uncategorized"}
                      options={categoryOptions}
                      onValueChange={(value) =>
                        field.onChange(value === "uncategorized" ? "" : value)
                      }
                      className="mt-2"
                    />
                  )}
                />
              </FormField>
              <FormField
                label="Base UOM"
                htmlFor="item-base-uom"
                required
                error={errors.base_uom_id?.message}
              >
                {editing ? (
                  <Input
                    id="item-base-uom"
                    readOnly
                    value={
                      currentUOM
                        ? `${currentUOM.name} (${currentUOM.code})`
                        : (detail.data?.base_uom_id ?? item.base_uom_id)
                    }
                  />
                ) : (
                  <Controller
                    control={control}
                    name="base_uom_id"
                    render={({ field }) => (
                      <Select
                        id="item-base-uom"
                        ariaLabel="Base unit of measure"
                        value={field.value}
                        options={activeUOMOptions}
                        onValueChange={field.onChange}
                        placeholder={
                          activeUOMOptions.length
                            ? "Select base UOM"
                            : "No active UOMs"
                        }
                        invalid={Boolean(errors.base_uom_id)}
                        className="mt-2"
                      />
                    )}
                  />
                )}
              </FormField>
            </div>
          </section>

          <section className="mt-7 border-t border-slate-200 pt-6">
            <h2 className="text-xs font-bold tracking-wide text-slate-500 uppercase">
              Physical values
            </h2>
            <div className="mt-4 grid gap-5 sm:grid-cols-2">
              <FormField
                label="Weight"
                htmlFor="item-weight"
                error={errors.weight?.message}
              >
                <Input
                  id="item-weight"
                  inputMode="decimal"
                  placeholder="0.000000"
                  invalid={Boolean(errors.weight)}
                  {...register("weight")}
                />
              </FormField>
              <FormField
                label="Volume"
                htmlFor="item-volume"
                error={errors.volume?.message}
              >
                <Input
                  id="item-volume"
                  inputMode="decimal"
                  placeholder="0.000000"
                  invalid={Boolean(errors.volume)}
                  {...register("volume")}
                />
              </FormField>
            </div>
            <p className="mt-2 text-xs text-slate-500">
              Values use the item&apos;s base UOM and allow up to six decimal
              places.
            </p>
          </section>

          <section className="mt-7 border-t border-slate-200 pt-6">
            <h2 className="text-xs font-bold tracking-wide text-slate-500 uppercase">
              Inventory controls
            </h2>
            <div className="mt-4 grid gap-4 sm:grid-cols-2">
              <label className="flex min-h-12 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
                <input
                  type="checkbox"
                  className="size-4 accent-cyan-600"
                  {...register("lot_controlled")}
                />
                Lot controlled
              </label>
              <label className="flex min-h-12 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
                <input
                  type="checkbox"
                  className="size-4 accent-cyan-600"
                  {...register("serial_controlled")}
                />
                Serial controlled
              </label>
              <FormField
                label="Shelf life (days)"
                htmlFor="item-shelf-life"
                error={errors.shelf_life_days?.message}
              >
                <Input
                  id="item-shelf-life"
                  inputMode="numeric"
                  min={0}
                  placeholder="Optional"
                  invalid={Boolean(errors.shelf_life_days)}
                  {...register("shelf_life_days")}
                />
              </FormField>
              <FormField
                label="Minimum life at receipt (days)"
                htmlFor="item-minimum-receive"
                error={errors.minimum_receive_days?.message}
              >
                <Input
                  id="item-minimum-receive"
                  inputMode="numeric"
                  min={0}
                  placeholder="Optional"
                  invalid={Boolean(errors.minimum_receive_days)}
                  {...register("minimum_receive_days")}
                />
              </FormField>
              {editing ? (
                <label className="flex min-h-12 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800 sm:col-span-2">
                  <input
                    type="checkbox"
                    className="size-4 accent-cyan-600"
                    {...register("is_active")}
                  />
                  Item is active
                </label>
              ) : null}
            </div>
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
                save.isPending || (!editing && activeUOMOptions.length === 0)
              }
            >
              {save.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              {save.isPending
                ? "Saving…"
                : editing
                  ? "Save changes"
                  : "Create item"}
            </Button>
          </div>
        </form>
      )}
    </DialogFrame>
  );
}
