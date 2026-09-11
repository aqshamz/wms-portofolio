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
import {
  listOrganizations,
  organizationKeys,
} from "@/features/organizations/organization-api";
import {
  createWarehouse,
  getWarehouse,
  updateWarehouse,
  warehouseKeys,
} from "@/features/warehouses/warehouse-api";
import {
  emptyWarehouseForm,
  warehouseFormSchema,
  type WarehouseFormValues,
} from "@/features/warehouses/warehouse-schema";
import type {
  CreateWarehouseRequest,
  UpdateWarehouseRequest,
  Warehouse,
} from "@/features/warehouses/warehouse-types";
import { ApiError } from "@/lib/api/client";

const activeOperatorFilters = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

function optional(value: string) {
  const trimmed = value.trim();
  return trimmed || undefined;
}

function formValues(warehouse: Warehouse): WarehouseFormValues {
  return {
    operator_id: warehouse.operator_id,
    code: warehouse.code,
    name: warehouse.name,
    timezone_name: warehouse.timezone_name,
    address_line_1: warehouse.address_line_1 ?? "",
    address_line_2: warehouse.address_line_2 ?? "",
    city: warehouse.city ?? "",
    province: warehouse.province ?? "",
    postal_code: warehouse.postal_code ?? "",
    country_code: warehouse.country_code ?? "",
    is_active: warehouse.is_active,
  };
}

function mutableRequest(
  values: WarehouseFormValues,
): Omit<CreateWarehouseRequest, "operator_id" | "code"> {
  return {
    name: values.name.trim(),
    timezone_name: values.timezone_name.trim(),
    address_line_1: optional(values.address_line_1),
    address_line_2: optional(values.address_line_2),
    city: optional(values.city),
    province: optional(values.province),
    postal_code: optional(values.postal_code),
    country_code: optional(values.country_code)?.toUpperCase(),
  };
}

function describedBy(id: string, error?: string) {
  return error ? `${id}-error` : undefined;
}

export function WarehouseFormDialog({
  open,
  warehouseId,
  onOpenChange,
}: {
  open: boolean;
  warehouseId?: string;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = warehouseId !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: warehouseKeys.detail(warehouseId ?? "new"),
    queryFn: () => getWarehouse(warehouseId!),
    enabled: open && editing,
  });
  const operators = useQuery({
    queryKey: organizationKeys.list(activeOperatorFilters),
    queryFn: () => listOrganizations(activeOperatorFilters),
    enabled: open && !editing,
  });
  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<WarehouseFormValues>({
    resolver: zodResolver(warehouseFormSchema),
    defaultValues: emptyWarehouseForm,
  });
  const save = useMutation({
    mutationFn: async (values: WarehouseFormValues) => {
      const request = mutableRequest(values);
      if (!editing) {
        return createWarehouse({
          operator_id: values.operator_id,
          code: values.code.trim().toUpperCase(),
          ...request,
        });
      }
      if (!detail.data) throw new Error("Warehouse details are not loaded.");

      const updateRequest: UpdateWarehouseRequest = {
        ...request,
        is_active: values.is_active,
        expected_updated_at: detail.data.updated_at,
      };
      return updateWarehouse(detail.data.warehouse_id, updateRequest);
    },
    onSuccess: async (warehouse) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: warehouseKeys.lists() }),
        queryClient.invalidateQueries({
          queryKey: warehouseKeys.detail(warehouse.warehouse_id),
        }),
      ]);
      toast.success(editing ? "Warehouse updated." : "Warehouse created.");
      onOpenChange(false);
    },
  });

  useEffect(() => {
    if (!open) return;
    if (!editing) reset(emptyWarehouseForm);
  }, [editing, open, reset]);

  useEffect(() => {
    if (open && detail.data) reset(formValues(detail.data));
  }, [detail.data, open, reset]);

  const operatorOptions =
    operators.data?.items.map((organization) => ({
      value: organization.organization_id,
      label: `${organization.name} (${organization.code})`,
    })) ?? [];

  return (
    <Dialog.Root
      open={open}
      onOpenChange={(nextOpen) => {
        if (!save.isPending) onOpenChange(nextOpen);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] max-w-2xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                {editing ? "Edit warehouse" : "Create warehouse"}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                {editing
                  ? "Update the facility profile and operational address."
                  : "Add a facility and identify the organization operating it."}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100 hover:text-slate-950"
                aria-label="Close warehouse form"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>

          {editing && detail.isPending ? (
            <div className="grid min-h-72 place-items-center p-6 text-sm text-slate-600">
              <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
              Loading warehouse…
            </div>
          ) : editing && detail.isError ? (
            <div className="p-6">
              <div
                role="alert"
                className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
              >
                {detail.error.message}
              </div>
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
              onSubmit={handleSubmit((values) => save.mutate(values))}
              className="p-5 sm:p-6"
            >
              {save.error ? (
                <div
                  role="alert"
                  className="mb-5 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-900"
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

              <div className="grid gap-5 sm:grid-cols-2">
                <div className="sm:col-span-2">
                  <FormField
                    label="Operating organization"
                    htmlFor="warehouse-operator"
                    required
                    error={errors.operator_id?.message}
                  >
                    {editing ? (
                      <Input
                        id="warehouse-operator"
                        readOnly
                        value={
                          detail.data
                            ? `${detail.data.operator_name} (${detail.data.operator_code})`
                            : ""
                        }
                      />
                    ) : (
                      <Controller
                        control={control}
                        name="operator_id"
                        render={({ field }) => (
                          <Select
                            id="warehouse-operator"
                            ariaLabel="Operating organization"
                            ariaDescribedBy={describedBy(
                              "warehouse-operator",
                              errors.operator_id?.message,
                            )}
                            value={field.value}
                            onValueChange={field.onChange}
                            options={operatorOptions}
                            placeholder={
                              operators.isPending
                                ? "Loading organizations…"
                                : operatorOptions.length === 0
                                  ? "No active organizations"
                                  : "Select an organization"
                            }
                            disabled={operators.isPending || operators.isError}
                            invalid={Boolean(errors.operator_id)}
                            className="mt-2"
                          />
                        )}
                      />
                    )}
                    {!editing && operators.isError ? (
                      <p className="mt-1.5 text-xs text-rose-700">
                        Organizations could not be loaded.
                      </p>
                    ) : null}
                  </FormField>
                </div>

                <FormField
                  label="Code"
                  htmlFor="warehouse-code"
                  required
                  error={errors.code?.message}
                >
                  <Input
                    id="warehouse-code"
                    placeholder="Example: JKT_DC"
                    readOnly={editing}
                    invalid={Boolean(errors.code)}
                    aria-describedby={describedBy(
                      "warehouse-code",
                      errors.code?.message,
                    )}
                    {...register("code")}
                  />
                </FormField>
                <FormField
                  label="Display name"
                  htmlFor="warehouse-name"
                  required
                  error={errors.name?.message}
                >
                  <Input
                    id="warehouse-name"
                    placeholder="Warehouse name"
                    autoFocus={!editing}
                    invalid={Boolean(errors.name)}
                    aria-describedby={describedBy(
                      "warehouse-name",
                      errors.name?.message,
                    )}
                    {...register("name")}
                  />
                </FormField>
                <FormField
                  label="Timezone"
                  htmlFor="warehouse-timezone"
                  required
                  error={errors.timezone_name?.message}
                >
                  <Input
                    id="warehouse-timezone"
                    list="warehouse-timezone-options"
                    placeholder="Asia/Jakarta"
                    invalid={Boolean(errors.timezone_name)}
                    aria-describedby={describedBy(
                      "warehouse-timezone",
                      errors.timezone_name?.message,
                    )}
                    {...register("timezone_name")}
                  />
                  <datalist id="warehouse-timezone-options">
                    <option value="Asia/Jakarta" />
                    <option value="Asia/Makassar" />
                    <option value="Asia/Jayapura" />
                    <option value="UTC" />
                  </datalist>
                </FormField>
                <FormField
                  label="Country code"
                  htmlFor="warehouse-country"
                  error={errors.country_code?.message}
                >
                  <Input
                    id="warehouse-country"
                    placeholder="ID"
                    maxLength={2}
                    invalid={Boolean(errors.country_code)}
                    aria-describedby={describedBy(
                      "warehouse-country",
                      errors.country_code?.message,
                    )}
                    {...register("country_code")}
                  />
                </FormField>
                <div className="sm:col-span-2">
                  <FormField
                    label="Address line 1"
                    htmlFor="warehouse-address-1"
                    error={errors.address_line_1?.message}
                  >
                    <Input
                      id="warehouse-address-1"
                      invalid={Boolean(errors.address_line_1)}
                      {...register("address_line_1")}
                    />
                  </FormField>
                </div>
                <div className="sm:col-span-2">
                  <FormField
                    label="Address line 2"
                    htmlFor="warehouse-address-2"
                    error={errors.address_line_2?.message}
                  >
                    <Input
                      id="warehouse-address-2"
                      invalid={Boolean(errors.address_line_2)}
                      {...register("address_line_2")}
                    />
                  </FormField>
                </div>
                <FormField
                  label="City"
                  htmlFor="warehouse-city"
                  error={errors.city?.message}
                >
                  <Input
                    id="warehouse-city"
                    invalid={Boolean(errors.city)}
                    {...register("city")}
                  />
                </FormField>
                <FormField
                  label="Province"
                  htmlFor="warehouse-province"
                  error={errors.province?.message}
                >
                  <Input
                    id="warehouse-province"
                    invalid={Boolean(errors.province)}
                    {...register("province")}
                  />
                </FormField>
                <FormField
                  label="Postal code"
                  htmlFor="warehouse-postal-code"
                  error={errors.postal_code?.message}
                >
                  <Input
                    id="warehouse-postal-code"
                    invalid={Boolean(errors.postal_code)}
                    {...register("postal_code")}
                  />
                </FormField>
                {editing ? (
                  <label className="flex min-h-11 items-center gap-3 self-end rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
                    <input
                      type="checkbox"
                      className="size-4 accent-cyan-600"
                      {...register("is_active")}
                    />
                    Warehouse is active
                  </label>
                ) : null}
              </div>

              <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
                <Dialog.Close asChild>
                  <Button type="button" variant="secondary">
                    Cancel
                  </Button>
                </Dialog.Close>
                <Button
                  type="submit"
                  disabled={save.isPending || (!editing && operators.isPending)}
                >
                  {save.isPending ? (
                    <LoaderCircle className="size-4 animate-spin" />
                  ) : null}
                  {save.isPending
                    ? "Saving…"
                    : editing
                      ? "Save changes"
                      : "Create warehouse"}
                </Button>
              </div>
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
