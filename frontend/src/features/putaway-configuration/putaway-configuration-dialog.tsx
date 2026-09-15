"use client";

import { useEffect } from "react";
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
import { listItemCategories } from "@/features/item-catalog/item-catalog-api";
import { listOrganizations } from "@/features/organizations/organization-api";
import {
  createPutawayStrategy,
  createPutawayStrategyRule,
  getPutawayStrategy,
  getPutawayStrategyRule,
  putawayConfigurationKeys,
  updatePutawayStrategy,
  updatePutawayStrategyRule,
} from "@/features/putaway-configuration/putaway-configuration-api";
import {
  putawayRuleFormSchema,
  putawayStrategyFormSchema,
  type PutawayRuleFormValues,
  type PutawayStrategyFormValues,
} from "@/features/putaway-configuration/putaway-configuration-schema";
import type {
  PutawayStrategy,
  PutawayStrategyRule,
} from "@/features/putaway-configuration/putaway-configuration-types";
import {
  listLocationTypes,
  listZones,
} from "@/features/storage-layout/storage-layout-api";
import { listWarehouses } from "@/features/warehouses/warehouse-api";

export type PutawayEditor =
  | { kind: "strategy"; item?: PutawayStrategy }
  | {
      kind: "rule";
      strategy: PutawayStrategy;
      nextSequence: number;
      item?: PutawayStrategyRule;
    };

const allRecords = {
  search: "",
  active: "all" as const,
  page: 1,
  pageSize: 100,
};
function optional(value: string) {
  return value.trim() || undefined;
}
function optionalId(value: string) {
  return value === "none" ? undefined : value;
}
function Header({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
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
          aria-label="Close putaway configuration form"
          className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
        >
          <X className="size-5" />
        </button>
      </Dialog.Close>
    </div>
  );
}
function LoadingDetail() {
  return (
    <div className="grid min-h-64 place-items-center p-6 text-sm text-slate-600">
      <div className="text-center">
        <LoaderCircle className="mx-auto mb-3 size-6 animate-spin text-cyan-700" />
        Loading configuration…
      </div>
    </div>
  );
}
function DetailError({ error, retry }: { error: Error; retry: () => void }) {
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
function Actions({
  editing,
  pending,
  subject,
}: {
  editing: boolean;
  pending: boolean;
  subject: string;
}) {
  return (
    <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
      <Dialog.Close asChild>
        <Button type="button" variant="secondary">
          Cancel
        </Button>
      </Dialog.Close>
      <Button type="submit" disabled={pending}>
        {pending ? <LoaderCircle className="size-4 animate-spin" /> : null}
        {pending ? "Saving…" : editing ? "Save changes" : `Create ${subject}`}
      </Button>
    </div>
  );
}

function StrategyDialog({
  item,
  onOpenChange,
}: {
  item?: PutawayStrategy;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: putawayConfigurationKeys.strategy(
      item?.putaway_strategy_id ?? "new",
    ),
    queryFn: () => getPutawayStrategy(item!.putaway_strategy_id),
    enabled: editing,
  });
  const organizations = useQuery({
    queryKey: putawayConfigurationKeys.organizations(),
    queryFn: () => listOrganizations(allRecords),
  });
  const warehouses = useQuery({
    queryKey: putawayConfigurationKeys.warehouses(),
    queryFn: () => listWarehouses(allRecords),
  });
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<PutawayStrategyFormValues>({
    resolver: zodResolver(putawayStrategyFormSchema),
    defaultValues: {
      code: "",
      name: "",
      description: "",
      owner_id: "none",
      warehouse_id: "none",
      is_active: true,
    },
  });
  const save = useMutation({
    mutationFn: (values: PutawayStrategyFormValues) => {
      const common = {
        name: values.name.trim(),
        description: optional(values.description),
      };
      return item
        ? updatePutawayStrategy(item.putaway_strategy_id, {
            ...common,
            is_active: values.is_active,
          })
        : createPutawayStrategy({
            code: values.code.trim().toUpperCase(),
            ...common,
            owner_id: optionalId(values.owner_id),
            warehouse_id: optionalId(values.warehouse_id),
          });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: putawayConfigurationKeys.all,
      });
      toast.success(`Putaway strategy ${editing ? "updated" : "created"}.`);
      onOpenChange(false);
    },
  });
  useEffect(() => {
    const value = detail.data ?? item;
    reset({
      code: value?.code ?? "",
      name: value?.name ?? "",
      description: value?.description ?? "",
      owner_id: value?.owner_id ?? "none",
      warehouse_id: value?.warehouse_id ?? "none",
      is_active: value?.is_active ?? true,
    });
  }, [detail.data, item, reset]);
  const ownerOptions = [
    { value: "none", label: "All owners (global)" },
    ...(organizations.data?.items ?? []).map((owner) => ({
      value: owner.organization_id,
      label: `${owner.name} (${owner.code})${owner.is_active ? "" : " · Inactive"}`,
      disabled: !owner.is_active && owner.organization_id !== item?.owner_id,
    })),
  ];
  const warehouseOptions = [
    { value: "none", label: "All warehouses" },
    ...(warehouses.data?.items ?? []).map((warehouse) => ({
      value: warehouse.warehouse_id,
      label: `${warehouse.name} (${warehouse.code})${warehouse.is_active ? "" : " · Inactive"}`,
      disabled:
        !warehouse.is_active && warehouse.warehouse_id !== item?.warehouse_id,
    })),
  ];
  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!save.isPending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] max-w-xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <Header
            title={`${editing ? "Edit" : "Create"} putaway strategy`}
            description="Define the owner and warehouse scope where ordered destination rules apply."
          />
          {editing && detail.isPending ? (
            <LoadingDetail />
          ) : editing && detail.isError ? (
            <DetailError
              error={detail.error}
              retry={() => void detail.refetch()}
            />
          ) : (
            <form
              noValidate
              className="p-5 sm:p-6"
              onSubmit={handleSubmit((values) => save.mutate(values))}
            >
              {save.error ? (
                <p
                  role="alert"
                  className="mb-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
                >
                  {save.error.message}
                </p>
              ) : null}
              <div className="grid gap-5 sm:grid-cols-2">
                <FormField
                  label="Code"
                  htmlFor="putaway-strategy-code"
                  required
                  error={errors.code?.message}
                >
                  <Input
                    id="putaway-strategy-code"
                    readOnly={editing}
                    placeholder="STANDARD_PUTAWAY"
                    invalid={Boolean(errors.code)}
                    {...register("code")}
                  />
                </FormField>
                <FormField
                  label="Name"
                  htmlFor="putaway-strategy-name"
                  required
                  error={errors.name?.message}
                >
                  <Input
                    id="putaway-strategy-name"
                    placeholder="Standard putaway"
                    invalid={Boolean(errors.name)}
                    {...register("name")}
                  />
                </FormField>
                <FormField label="Owner scope" htmlFor="putaway-owner">
                  <Controller
                    name="owner_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="putaway-owner"
                        ariaLabel="Owner scope"
                        value={field.value}
                        options={ownerOptions}
                        onValueChange={field.onChange}
                        disabled={editing || organizations.isPending}
                      />
                    )}
                  />
                  <p className="mt-1.5 text-xs text-slate-500">
                    An owner scope enables category-specific rules.
                  </p>
                </FormField>
                <FormField label="Warehouse scope" htmlFor="putaway-warehouse">
                  <Controller
                    name="warehouse_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="putaway-warehouse"
                        ariaLabel="Warehouse scope"
                        value={field.value}
                        options={warehouseOptions}
                        onValueChange={field.onChange}
                        disabled={editing || warehouses.isPending}
                      />
                    )}
                  />
                  <p className="mt-1.5 text-xs text-slate-500">
                    A warehouse scope enables zone-specific rules.
                  </p>
                </FormField>
                <div className="sm:col-span-2">
                  <FormField
                    label="Description"
                    htmlFor="putaway-strategy-description"
                  >
                    <textarea
                      id="putaway-strategy-description"
                      rows={3}
                      className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                      {...register("description")}
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
                    Strategy is active
                  </label>
                ) : null}
              </div>
              <Actions
                editing={editing}
                pending={save.isPending}
                subject="putaway strategy"
              />
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

function RuleDialog({
  target,
  onOpenChange,
}: {
  target: Extract<PutawayEditor, { kind: "rule" }>;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = target.item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: putawayConfigurationKeys.rule(
      target.strategy.putaway_strategy_id,
      target.item?.rule_id ?? "new",
    ),
    queryFn: () =>
      getPutawayStrategyRule(
        target.strategy.putaway_strategy_id,
        target.item!.rule_id,
      ),
    enabled: editing,
  });
  const categories = useQuery({
    queryKey: putawayConfigurationKeys.categories(
      target.strategy.owner_id ?? "global",
    ),
    queryFn: () =>
      listItemCategories({
        ownerId: target.strategy.owner_id!,
        search: "",
        active: "all",
        page: 1,
        pageSize: 100,
      }),
    enabled: Boolean(target.strategy.owner_id),
  });
  const locationTypes = useQuery({
    queryKey: putawayConfigurationKeys.locationTypes(),
    queryFn: () => listLocationTypes("all"),
  });
  const zones = useQuery({
    queryKey: putawayConfigurationKeys.zones(
      target.strategy.warehouse_id ?? "global",
    ),
    queryFn: () => listZones(target.strategy.warehouse_id!, "all"),
    enabled: Boolean(target.strategy.warehouse_id),
  });
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<PutawayRuleFormValues>({
    resolver: zodResolver(putawayRuleFormSchema),
    defaultValues: {
      sequence_no: target.nextSequence,
      category_id: "none",
      location_type_id: "none",
      zone_id: "none",
      minimum_empty_percent: "",
      is_active: true,
    },
  });
  const save = useMutation({
    mutationFn: (values: PutawayRuleFormValues) => {
      const common = {
        sequence_no: values.sequence_no,
        category_id: optionalId(values.category_id),
        location_type_id: optionalId(values.location_type_id),
        zone_id: optionalId(values.zone_id),
        minimum_empty_percent: optional(values.minimum_empty_percent),
      };
      return target.item
        ? updatePutawayStrategyRule(
            target.strategy.putaway_strategy_id,
            target.item.rule_id,
            { ...common, is_active: values.is_active },
          )
        : createPutawayStrategyRule(
            target.strategy.putaway_strategy_id,
            common,
          );
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: putawayConfigurationKeys.all,
      });
      toast.success(`Putaway rule ${editing ? "updated" : "created"}.`);
      onOpenChange(false);
    },
  });
  useEffect(() => {
    const value = detail.data ?? target.item;
    reset({
      sequence_no: value?.sequence_no ?? target.nextSequence,
      category_id: value?.category_id ?? "none",
      location_type_id: value?.location_type_id ?? "none",
      zone_id: value?.zone_id ?? "none",
      minimum_empty_percent: value?.minimum_empty_percent ?? "",
      is_active: value?.is_active ?? true,
    });
  }, [detail.data, reset, target.item, target.nextSequence]);
  const categoryOptions = [
    { value: "none", label: "Any item category" },
    ...(categories.data?.items ?? [])
      .filter(
        (category) =>
          category.is_active ||
          category.category_id === target.item?.category_id,
      )
      .map((category) => ({
        value: category.category_id,
        label: `${category.name} (${category.code})${category.is_active ? "" : " · Inactive"}`,
        disabled:
          !category.is_active &&
          category.category_id !== target.item?.category_id,
      })),
  ];
  const locationTypeOptions = [
    { value: "none", label: "Any location type" },
    ...(locationTypes.data ?? []).map((locationType) => ({
      value: locationType.location_type_id,
      label: `${locationType.name} (${locationType.code})${locationType.is_active ? "" : " · Inactive"}`,
      disabled:
        !locationType.is_active &&
        locationType.location_type_id !== target.item?.location_type_id,
    })),
  ];
  const zoneOptions = [
    { value: "none", label: "Any zone" },
    ...(zones.data ?? []).map((zone) => ({
      value: zone.zone_id,
      label: `${zone.name} (${zone.code})${zone.is_active ? "" : " · Inactive"}`,
      disabled: !zone.is_active && zone.zone_id !== target.item?.zone_id,
    })),
  ];
  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!save.isPending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] max-w-xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <Header
            title={`${editing ? "Edit" : "Create"} putaway rule`}
            description="Rules are evaluated in sequence; optional conditions narrow eligible destinations."
          />
          {editing && detail.isPending ? (
            <LoadingDetail />
          ) : editing && detail.isError ? (
            <DetailError
              error={detail.error}
              retry={() => void detail.refetch()}
            />
          ) : (
            <form
              noValidate
              className="p-5 sm:p-6"
              onSubmit={handleSubmit((values) => save.mutate(values))}
            >
              {save.error ? (
                <p
                  role="alert"
                  className="mb-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
                >
                  {save.error.message}
                </p>
              ) : null}
              <div className="grid gap-5">
                <FormField
                  label="Sequence"
                  htmlFor="putaway-rule-sequence"
                  required
                  error={errors.sequence_no?.message}
                >
                  <Input
                    id="putaway-rule-sequence"
                    type="number"
                    min={1}
                    invalid={Boolean(errors.sequence_no)}
                    {...register("sequence_no")}
                  />
                </FormField>
                <FormField
                  label="Item category"
                  htmlFor="putaway-rule-category"
                >
                  <Controller
                    name="category_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="putaway-rule-category"
                        ariaLabel="Item category filter"
                        value={field.value}
                        options={categoryOptions}
                        onValueChange={field.onChange}
                        disabled={
                          !target.strategy.owner_id || categories.isPending
                        }
                      />
                    )}
                  />
                  <p className="mt-1.5 text-xs text-slate-500">
                    {target.strategy.owner_id
                      ? "Optional. Limit the rule to one owner category."
                      : "Category filters require an owner-scoped strategy."}
                  </p>
                </FormField>
                <FormField
                  label="Location type"
                  htmlFor="putaway-rule-location-type"
                >
                  <Controller
                    name="location_type_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="putaway-rule-location-type"
                        ariaLabel="Location type filter"
                        value={field.value}
                        options={locationTypeOptions}
                        onValueChange={field.onChange}
                        disabled={locationTypes.isPending}
                      />
                    )}
                  />
                </FormField>
                <FormField label="Warehouse zone" htmlFor="putaway-rule-zone">
                  <Controller
                    name="zone_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="putaway-rule-zone"
                        ariaLabel="Warehouse zone filter"
                        value={field.value}
                        options={zoneOptions}
                        onValueChange={field.onChange}
                        disabled={
                          !target.strategy.warehouse_id || zones.isPending
                        }
                      />
                    )}
                  />
                  <p className="mt-1.5 text-xs text-slate-500">
                    {target.strategy.warehouse_id
                      ? "Optional. Limit the rule to one warehouse zone."
                      : "Zone filters require a warehouse-scoped strategy."}
                  </p>
                </FormField>
                <FormField
                  label="Minimum empty capacity (%)"
                  htmlFor="putaway-rule-empty"
                  error={errors.minimum_empty_percent?.message}
                >
                  <Input
                    id="putaway-rule-empty"
                    inputMode="decimal"
                    placeholder="For example, 25"
                    invalid={Boolean(errors.minimum_empty_percent)}
                    {...register("minimum_empty_percent")}
                  />
                  <p className="mt-1.5 text-xs text-slate-500">
                    Optional value from 0 to 100, with up to four decimal
                    places.
                  </p>
                </FormField>
                {editing ? (
                  <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
                    <input
                      type="checkbox"
                      className="size-4 accent-cyan-600"
                      {...register("is_active")}
                    />
                    Rule is active
                  </label>
                ) : null}
              </div>
              <div className="mt-5 rounded-xl bg-cyan-50 p-4 text-sm leading-6 text-cyan-950">
                Leaving all conditions empty creates a catch-all destination
                rule. Place broad rules after more specific ones.
              </div>
              <Actions
                editing={editing}
                pending={save.isPending}
                subject="putaway rule"
              />
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function PutawayConfigurationDialog({
  target,
  onOpenChange,
}: {
  target: PutawayEditor;
  onOpenChange: (open: boolean) => void;
}) {
  return target.kind === "strategy" ? (
    <StrategyDialog item={target.item} onOpenChange={onOpenChange} />
  ) : (
    <RuleDialog target={target} onOpenChange={onOpenChange} />
  );
}
