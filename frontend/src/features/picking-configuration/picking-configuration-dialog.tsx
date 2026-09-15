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
import { listInventoryStatuses } from "@/features/inventory-classifications/inventory-classification-api";
import { listOrganizations } from "@/features/organizations/organization-api";
import {
  createPickingSortMethod,
  createPickingStrategy,
  createPickingStrategyRule,
  getPickingSortMethod,
  getPickingStrategy,
  getPickingStrategyRule,
  listPickingSortMethods,
  pickingConfigurationKeys,
  updatePickingSortMethod,
  updatePickingStrategy,
  updatePickingStrategyRule,
} from "@/features/picking-configuration/picking-configuration-api";
import {
  pickingRecordFormSchema,
  pickingRuleFormSchema,
  type PickingRecordFormValues,
  type PickingRuleFormValues,
} from "@/features/picking-configuration/picking-configuration-schema";
import type {
  PickingSortMethod,
  PickingStrategy,
  PickingStrategyRule,
} from "@/features/picking-configuration/picking-configuration-types";
import { listZones } from "@/features/storage-layout/storage-layout-api";
import { listWarehouses } from "@/features/warehouses/warehouse-api";

export type PickingEditor =
  | { kind: "method"; item?: PickingSortMethod }
  | { kind: "strategy"; item?: PickingStrategy }
  | {
      kind: "rule";
      strategy: PickingStrategy;
      nextSequence: number;
      item?: PickingStrategyRule;
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
          aria-label="Close picking configuration form"
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

function RecordDialog({
  target,
  onOpenChange,
}: {
  target: Extract<PickingEditor, { kind: "method" | "strategy" }>;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = target.item !== undefined;
  const queryClient = useQueryClient();
  const id =
    target.kind === "method"
      ? target.item?.picking_sort_method_id
      : target.item?.picking_strategy_id;
  const detail = useQuery<PickingSortMethod | PickingStrategy>({
    queryKey:
      target.kind === "method"
        ? pickingConfigurationKeys.method(id ?? "new")
        : pickingConfigurationKeys.strategy(id ?? "new"),
    queryFn: () =>
      target.kind === "method"
        ? getPickingSortMethod(id!)
        : getPickingStrategy(id!),
    enabled: editing,
  });
  const organizations = useQuery({
    queryKey: pickingConfigurationKeys.organizations(),
    queryFn: () => listOrganizations(allRecords),
    enabled: target.kind === "strategy",
  });
  const warehouses = useQuery({
    queryKey: pickingConfigurationKeys.warehouses(),
    queryFn: () => listWarehouses(allRecords),
    enabled: target.kind === "strategy",
  });
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<PickingRecordFormValues>({
    resolver: zodResolver(pickingRecordFormSchema),
    defaultValues: {
      code: "",
      name: "",
      description: "",
      owner_id: "none",
      warehouse_id: "none",
      is_active: true,
    },
  });
  const save = useMutation<
    PickingSortMethod | PickingStrategy,
    Error,
    PickingRecordFormValues
  >({
    mutationFn: (values) => {
      const common = {
        name: values.name.trim(),
        description: optional(values.description),
      };
      if (target.kind === "method")
        return target.item
          ? updatePickingSortMethod(target.item.picking_sort_method_id, {
              ...common,
              is_active: values.is_active,
            })
          : createPickingSortMethod({
              code: values.code.trim().toUpperCase(),
              ...common,
            });
      return target.item
        ? updatePickingStrategy(target.item.picking_strategy_id, {
            ...common,
            is_active: values.is_active,
          })
        : createPickingStrategy({
            code: values.code.trim().toUpperCase(),
            ...common,
            owner_id: optionalId(values.owner_id),
            warehouse_id: optionalId(values.warehouse_id),
          });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: pickingConfigurationKeys.all,
      });
      toast.success(
        `${target.kind === "method" ? "Sort method" : "Picking strategy"} ${editing ? "updated" : "created"}.`,
      );
      onOpenChange(false);
    },
  });
  useEffect(() => {
    const value = detail.data ?? target.item;
    reset({
      code: value?.code ?? "",
      name: value?.name ?? "",
      description: value?.description ?? "",
      owner_id:
        target.kind === "strategy"
          ? ((value as PickingStrategy | undefined)?.owner_id ?? "none")
          : "none",
      warehouse_id:
        target.kind === "strategy"
          ? ((value as PickingStrategy | undefined)?.warehouse_id ?? "none")
          : "none",
      is_active: value?.is_active ?? true,
    });
  }, [detail.data, reset, target]);
  const ownerOptions = [
    { value: "none", label: "All owners (global)" },
    ...(organizations.data?.items ?? []).map((owner) => ({
      value: owner.organization_id,
      label: `${owner.name} (${owner.code})${owner.is_active ? "" : " · Inactive"}`,
      disabled:
        !owner.is_active &&
        owner.organization_id !==
          (target.kind === "strategy" ? target.item?.owner_id : undefined),
    })),
  ];
  const warehouseOptions = [
    { value: "none", label: "All warehouses" },
    ...(warehouses.data?.items ?? []).map((warehouse) => ({
      value: warehouse.warehouse_id,
      label: `${warehouse.name} (${warehouse.code})${warehouse.is_active ? "" : " · Inactive"}`,
      disabled:
        !warehouse.is_active &&
        warehouse.warehouse_id !==
          (target.kind === "strategy" ? target.item?.warehouse_id : undefined),
    })),
  ];
  const subject = target.kind === "method" ? "sort method" : "picking strategy";
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
            title={`${editing ? "Edit" : "Create"} ${subject}`}
            description={
              target.kind === "method"
                ? "Define how candidate inventory locations are ordered for picking."
                : "Define the owner and warehouse scope where an ordered rule set applies."
            }
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
                  htmlFor="picking-record-code"
                  required
                  error={errors.code?.message}
                >
                  <Input
                    id="picking-record-code"
                    readOnly={editing}
                    placeholder={
                      target.kind === "method" ? "FIFO" : "STANDARD_PICK"
                    }
                    invalid={Boolean(errors.code)}
                    {...register("code")}
                  />
                </FormField>
                <FormField
                  label="Name"
                  htmlFor="picking-record-name"
                  required
                  error={errors.name?.message}
                >
                  <Input
                    id="picking-record-name"
                    placeholder={
                      target.kind === "method"
                        ? "First in, first out"
                        : "Standard picking"
                    }
                    invalid={Boolean(errors.name)}
                    {...register("name")}
                  />
                </FormField>
                {target.kind === "strategy" ? (
                  <>
                    <FormField label="Owner scope" htmlFor="picking-owner">
                      <Controller
                        name="owner_id"
                        control={control}
                        render={({ field }) => (
                          <Select
                            id="picking-owner"
                            ariaLabel="Owner scope"
                            value={field.value}
                            options={ownerOptions}
                            onValueChange={field.onChange}
                            disabled={editing || organizations.isPending}
                          />
                        )}
                      />
                      <p className="mt-1.5 text-xs text-slate-500">
                        Leave global to apply across owners.
                      </p>
                    </FormField>
                    <FormField
                      label="Warehouse scope"
                      htmlFor="picking-warehouse"
                    >
                      <Controller
                        name="warehouse_id"
                        control={control}
                        render={({ field }) => (
                          <Select
                            id="picking-warehouse"
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
                  </>
                ) : null}
                <div className="sm:col-span-2">
                  <FormField
                    label="Description"
                    htmlFor="picking-record-description"
                  >
                    <textarea
                      id="picking-record-description"
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
                    Configuration is active
                  </label>
                ) : null}
              </div>
              <Actions
                editing={editing}
                pending={save.isPending}
                subject={subject}
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
  target: Extract<PickingEditor, { kind: "rule" }>;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = target.item !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: pickingConfigurationKeys.rule(
      target.strategy.picking_strategy_id,
      target.item?.rule_id ?? "new",
    ),
    queryFn: () =>
      getPickingStrategyRule(
        target.strategy.picking_strategy_id,
        target.item!.rule_id,
      ),
    enabled: editing,
  });
  const methods = useQuery({
    queryKey: pickingConfigurationKeys.methods(allRecords),
    queryFn: () => listPickingSortMethods(allRecords),
  });
  const inventoryStatuses = useQuery({
    queryKey: pickingConfigurationKeys.inventoryStatuses(),
    queryFn: () => listInventoryStatuses(allRecords),
  });
  const zones = useQuery({
    queryKey: pickingConfigurationKeys.zones(
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
  } = useForm<PickingRuleFormValues>({
    resolver: zodResolver(pickingRuleFormSchema),
    defaultValues: {
      sequence_no: target.nextSequence,
      inventory_status_id: "none",
      zone_id: "none",
      picking_sort_method_id: "",
      is_active: true,
    },
  });
  const save = useMutation({
    mutationFn: (values: PickingRuleFormValues) => {
      const common = {
        sequence_no: values.sequence_no,
        inventory_status_id: optionalId(values.inventory_status_id),
        zone_id: optionalId(values.zone_id),
        picking_sort_method_id: values.picking_sort_method_id,
      };
      return target.item
        ? updatePickingStrategyRule(
            target.strategy.picking_strategy_id,
            target.item.rule_id,
            { ...common, is_active: values.is_active },
          )
        : createPickingStrategyRule(
            target.strategy.picking_strategy_id,
            common,
          );
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: pickingConfigurationKeys.all,
      });
      toast.success(`Picking rule ${editing ? "updated" : "created"}.`);
      onOpenChange(false);
    },
  });
  useEffect(() => {
    const value = detail.data ?? target.item;
    reset({
      sequence_no: value?.sequence_no ?? target.nextSequence,
      inventory_status_id: value?.inventory_status_id ?? "none",
      zone_id: value?.zone_id ?? "none",
      picking_sort_method_id: value?.picking_sort_method_id ?? "",
      is_active: value?.is_active ?? true,
    });
  }, [detail.data, reset, target.item, target.nextSequence]);
  const methodOptions = (methods.data?.items ?? []).map((method) => ({
    value: method.picking_sort_method_id,
    label: `${method.name} (${method.code})${method.is_active ? "" : " · Inactive"}`,
    disabled:
      !method.is_active &&
      method.picking_sort_method_id !== target.item?.picking_sort_method_id,
  }));
  const inventoryOptions = [
    { value: "none", label: "Any pickable inventory status" },
    ...(inventoryStatuses.data?.items ?? [])
      .filter(
        (status) =>
          (status.is_active && status.is_allocatable && status.is_pickable) ||
          status.inventory_status_id === target.item?.inventory_status_id,
      )
      .map((status) => ({
        value: status.inventory_status_id,
        label: `${status.name} (${status.code})${status.is_active ? "" : " · Inactive"}`,
        disabled:
          (!status.is_active ||
            !status.is_allocatable ||
            !status.is_pickable) &&
          status.inventory_status_id !== target.item?.inventory_status_id,
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
            title={`${editing ? "Edit" : "Create"} picking rule`}
            description="Rules are evaluated in sequence; optional filters narrow where a sort method applies."
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
                  htmlFor="picking-rule-sequence"
                  required
                  error={errors.sequence_no?.message}
                >
                  <Input
                    id="picking-rule-sequence"
                    type="number"
                    min={1}
                    invalid={Boolean(errors.sequence_no)}
                    {...register("sequence_no")}
                  />
                </FormField>
                <FormField
                  label="Inventory status"
                  htmlFor="picking-rule-status"
                >
                  <Controller
                    name="inventory_status_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="picking-rule-status"
                        ariaLabel="Inventory status filter"
                        value={field.value}
                        options={inventoryOptions}
                        onValueChange={field.onChange}
                        disabled={inventoryStatuses.isPending}
                      />
                    )}
                  />
                </FormField>
                <FormField label="Warehouse zone" htmlFor="picking-rule-zone">
                  <Controller
                    name="zone_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="picking-rule-zone"
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
                      ? "Optional. Limit this rule to one zone."
                      : "Zone filters require a warehouse-scoped strategy."}
                  </p>
                </FormField>
                <FormField
                  label="Sort method"
                  htmlFor="picking-rule-method"
                  required
                  error={errors.picking_sort_method_id?.message}
                >
                  <Controller
                    name="picking_sort_method_id"
                    control={control}
                    render={({ field }) => (
                      <Select
                        id="picking-rule-method"
                        ariaLabel="Picking sort method"
                        placeholder="Select a sort method"
                        value={field.value}
                        options={methodOptions}
                        onValueChange={field.onChange}
                        disabled={methods.isPending}
                        invalid={Boolean(errors.picking_sort_method_id)}
                      />
                    )}
                  />
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
              <Actions
                editing={editing}
                pending={save.isPending}
                subject="picking rule"
              />
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function PickingConfigurationDialog({
  target,
  onOpenChange,
}: {
  target: PickingEditor;
  onOpenChange: (open: boolean) => void;
}) {
  return target.kind === "rule" ? (
    <RuleDialog target={target} onOpenChange={onOpenChange} />
  ) : (
    <RecordDialog target={target} onOpenChange={onOpenChange} />
  );
}
