"use client";

import { useEffect, useMemo, useState, type ReactNode } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, Plus, Tag, Trash2, X } from "lucide-react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  assignBusinessPartnerType,
  businessPartnerKeys,
  createBusinessPartner,
  createPartnerType,
  getBusinessPartner,
  getPartnerType,
  listPartnerTypes,
  removeBusinessPartnerType,
  updateBusinessPartner,
  updatePartnerType,
} from "@/features/business-partners/business-partner-api";
import {
  businessPartnerFormSchema,
  emptyBusinessPartnerForm,
  emptyPartnerTypeForm,
  partnerTypeFormSchema,
  type BusinessPartnerFormValues,
  type PartnerTypeFormValues,
} from "@/features/business-partners/business-partner-schema";
import type {
  BusinessPartner,
  BusinessPartnerDetail,
  BusinessPartnerProfileRequest,
  PartnerType,
} from "@/features/business-partners/business-partner-types";
import { ApiError } from "@/lib/api/client";

const activeTypeFilters = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};

function optional(value: string) {
  const normalized = value.trim();
  return normalized || undefined;
}

function DialogFrame({
  open,
  title,
  description,
  pending,
  onOpenChange,
  children,
  wide,
}: {
  open: boolean;
  title: string;
  description: string;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
  wide?: boolean;
}) {
  return (
    <Dialog.Root
      open={open}
      onOpenChange={(nextOpen) => {
        if (!pending) onOpenChange(nextOpen);
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

function partnerTypeValues(item: PartnerType): PartnerTypeFormValues {
  return {
    code: item.code,
    name: item.name,
    description: item.description ?? "",
    is_active: item.is_active,
  };
}

export function PartnerTypeDialog({
  open,
  partnerType,
  onOpenChange,
}: {
  open: boolean;
  partnerType?: PartnerType;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = partnerType !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: businessPartnerKeys.type(partnerType?.partner_type_id ?? "new"),
    queryFn: () => getPartnerType(partnerType!.partner_type_id),
    enabled: open && editing,
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<PartnerTypeFormValues>({
    resolver: zodResolver(partnerTypeFormSchema),
    defaultValues: emptyPartnerTypeForm,
  });
  const save = useMutation({
    mutationFn: (values: PartnerTypeFormValues) =>
      editing
        ? updatePartnerType(partnerType.partner_type_id, {
            name: values.name.trim(),
            description: optional(values.description),
            is_active: values.is_active,
          })
        : createPartnerType({
            code: values.code.trim().toUpperCase(),
            name: values.name.trim(),
            description: optional(values.description),
          }),
    onSuccess: async (item) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: businessPartnerKeys.typeLists(),
        }),
        queryClient.invalidateQueries({
          queryKey: businessPartnerKeys.type(item.partner_type_id),
        }),
        queryClient.invalidateQueries({
          queryKey: businessPartnerKeys.partnerLists(),
        }),
      ]);
      toast.success(
        editing ? "Partner type updated." : "Partner type created.",
      );
      onOpenChange(false);
    },
  });

  useEffect(() => {
    if (!open) return;
    reset(
      editing
        ? partnerTypeValues(detail.data ?? partnerType)
        : emptyPartnerTypeForm,
    );
  }, [detail.data, editing, open, partnerType, reset]);

  return (
    <DialogFrame
      open={open}
      title={editing ? "Edit partner type" : "Create partner type"}
      description="Partner types are shared labels used to classify suppliers, customers, carriers, and other partners."
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      {editing && detail.isPending ? (
        <div className="grid min-h-56 place-items-center p-6 text-sm text-slate-600">
          <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
          Loading partner type…
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
          <div className="grid gap-5 sm:grid-cols-2">
            <FormField
              label="Code"
              htmlFor="partner-type-code"
              required
              error={errors.code?.message}
            >
              <Input
                id="partner-type-code"
                placeholder="SUPPLIER"
                readOnly={editing}
                autoFocus={!editing}
                invalid={Boolean(errors.code)}
                {...register("code")}
              />
            </FormField>
            <FormField
              label="Name"
              htmlFor="partner-type-name"
              required
              error={errors.name?.message}
            >
              <Input
                id="partner-type-name"
                placeholder="Supplier"
                autoFocus={editing}
                invalid={Boolean(errors.name)}
                {...register("name")}
              />
            </FormField>
            <div className="sm:col-span-2">
              <FormField
                label="Description"
                htmlFor="partner-type-description"
                error={errors.description?.message}
              >
                <textarea
                  id="partner-type-description"
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
                Partner type is active
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
                  : "Create partner type"}
            </Button>
          </div>
        </form>
      )}
    </DialogFrame>
  );
}

function businessPartnerValues(
  item: BusinessPartnerDetail | BusinessPartner,
): BusinessPartnerFormValues {
  return {
    code: item.code,
    name: item.name,
    legal_name: item.legal_name ?? "",
    tax_number: item.tax_number ?? "",
    email: item.email ?? "",
    phone: item.phone ?? "",
    address_line_1: item.address_line_1 ?? "",
    address_line_2: item.address_line_2 ?? "",
    city: item.city ?? "",
    province: item.province ?? "",
    postal_code: item.postal_code ?? "",
    country_code: item.country_code ?? "ID",
    is_active: item.is_active,
  };
}

function profileRequest(
  values: BusinessPartnerFormValues,
): BusinessPartnerProfileRequest {
  return {
    name: values.name.trim(),
    legal_name: optional(values.legal_name),
    tax_number: optional(values.tax_number),
    email: optional(values.email),
    phone: optional(values.phone),
    address_line_1: optional(values.address_line_1),
    address_line_2: optional(values.address_line_2),
    city: optional(values.city),
    province: optional(values.province),
    postal_code: optional(values.postal_code),
    country_code: optional(values.country_code)?.toUpperCase(),
  };
}

export function BusinessPartnerDialog({
  open,
  ownerId,
  ownerLabel,
  partner,
  onCreated,
  onOpenChange,
}: {
  open: boolean;
  ownerId: string;
  ownerLabel: string;
  partner?: BusinessPartner;
  onCreated?: (partner: BusinessPartner) => void;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = partner !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: businessPartnerKeys.partner(partner?.partner_id ?? "new"),
    queryFn: () => getBusinessPartner(partner!.partner_id),
    enabled: open && editing,
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<BusinessPartnerFormValues>({
    resolver: zodResolver(businessPartnerFormSchema),
    defaultValues: emptyBusinessPartnerForm,
  });
  const save = useMutation({
    mutationFn: (values: BusinessPartnerFormValues) => {
      const profile = profileRequest(values);
      if (!editing) {
        return createBusinessPartner({
          owner_id: ownerId,
          code: values.code.trim().toUpperCase(),
          ...profile,
        });
      }
      if (!detail.data)
        throw new Error("Business partner details are not loaded.");
      return updateBusinessPartner(detail.data.partner_id, {
        ...profile,
        is_active: values.is_active,
        expected_updated_at: detail.data.updated_at,
      });
    },
    onSuccess: async (item) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: businessPartnerKeys.partnerLists(),
        }),
        queryClient.invalidateQueries({
          queryKey: businessPartnerKeys.partner(item.partner_id),
        }),
      ]);
      toast.success(
        editing ? "Business partner updated." : "Business partner created.",
      );
      onOpenChange(false);
      if (!editing) onCreated?.(item);
    },
  });

  useEffect(() => {
    if (!open) return;
    reset(
      editing
        ? businessPartnerValues(detail.data ?? partner)
        : emptyBusinessPartnerForm,
    );
  }, [detail.data, editing, open, partner, reset]);

  return (
    <DialogFrame
      open={open}
      title={editing ? "Edit business partner" : "Create business partner"}
      description={
        editing
          ? "Update the partner profile and contact details."
          : `Add a partner for ${ownerLabel}. You can assign its business types next.`
      }
      pending={save.isPending}
      onOpenChange={onOpenChange}
      wide
    >
      {editing && detail.isPending ? (
        <div className="grid min-h-72 place-items-center p-6 text-sm text-slate-600">
          <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
          Loading business partner…
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
          <div className="grid gap-5 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <FormField label="Owner" htmlFor="partner-owner" required>
                <Input id="partner-owner" value={ownerLabel} readOnly />
              </FormField>
            </div>
            <FormField
              label="Code"
              htmlFor="partner-code"
              required
              error={errors.code?.message}
            >
              <Input
                id="partner-code"
                placeholder="SUPPLIER_JKT"
                readOnly={editing}
                autoFocus={!editing}
                invalid={Boolean(errors.code)}
                {...register("code")}
              />
            </FormField>
            <FormField
              label="Display name"
              htmlFor="partner-name"
              required
              error={errors.name?.message}
            >
              <Input
                id="partner-name"
                placeholder="Partner name"
                autoFocus={editing}
                invalid={Boolean(errors.name)}
                {...register("name")}
              />
            </FormField>
            <FormField
              label="Legal name"
              htmlFor="partner-legal-name"
              error={errors.legal_name?.message}
            >
              <Input id="partner-legal-name" {...register("legal_name")} />
            </FormField>
            <FormField
              label="Tax number"
              htmlFor="partner-tax-number"
              error={errors.tax_number?.message}
            >
              <Input id="partner-tax-number" {...register("tax_number")} />
            </FormField>
            <FormField
              label="Email"
              htmlFor="partner-email"
              error={errors.email?.message}
            >
              <Input
                id="partner-email"
                type="email"
                invalid={Boolean(errors.email)}
                {...register("email")}
              />
            </FormField>
            <FormField
              label="Phone"
              htmlFor="partner-phone"
              error={errors.phone?.message}
            >
              <Input id="partner-phone" type="tel" {...register("phone")} />
            </FormField>
            <div className="sm:col-span-2">
              <FormField
                label="Address line 1"
                htmlFor="partner-address-1"
                error={errors.address_line_1?.message}
              >
                <Input id="partner-address-1" {...register("address_line_1")} />
              </FormField>
            </div>
            <div className="sm:col-span-2">
              <FormField
                label="Address line 2"
                htmlFor="partner-address-2"
                error={errors.address_line_2?.message}
              >
                <Input id="partner-address-2" {...register("address_line_2")} />
              </FormField>
            </div>
            <FormField
              label="City"
              htmlFor="partner-city"
              error={errors.city?.message}
            >
              <Input id="partner-city" {...register("city")} />
            </FormField>
            <FormField
              label="Province"
              htmlFor="partner-province"
              error={errors.province?.message}
            >
              <Input id="partner-province" {...register("province")} />
            </FormField>
            <FormField
              label="Postal code"
              htmlFor="partner-postal-code"
              error={errors.postal_code?.message}
            >
              <Input id="partner-postal-code" {...register("postal_code")} />
            </FormField>
            <FormField
              label="Country code"
              htmlFor="partner-country-code"
              error={errors.country_code?.message}
            >
              <Input
                id="partner-country-code"
                placeholder="ID"
                maxLength={2}
                invalid={Boolean(errors.country_code)}
                {...register("country_code")}
              />
            </FormField>
            {editing ? (
              <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800 sm:col-span-2">
                <input
                  type="checkbox"
                  className="size-4 accent-cyan-600"
                  {...register("is_active")}
                />
                Business partner is active
              </label>
            ) : null}
          </div>
          <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
            <Dialog.Close asChild>
              <Button type="button" variant="secondary">
                Cancel
              </Button>
            </Dialog.Close>
            <Button type="submit" disabled={save.isPending || !ownerId}>
              {save.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              {save.isPending
                ? "Saving…"
                : editing
                  ? "Save changes"
                  : "Create partner"}
            </Button>
          </div>
        </form>
      )}
    </DialogFrame>
  );
}

export function PartnerTypeAssignmentsDialog({
  open,
  partner,
  canWrite,
  onOpenChange,
}: {
  open: boolean;
  partner: BusinessPartner;
  canWrite: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [selectedTypeId, setSelectedTypeId] = useState("");
  const [removingId, setRemovingId] = useState<string>();
  const detail = useQuery({
    queryKey: businessPartnerKeys.partner(partner.partner_id),
    queryFn: () => getBusinessPartner(partner.partner_id),
    enabled: open,
  });
  const availableTypes = useQuery({
    queryKey: businessPartnerKeys.typeList(activeTypeFilters),
    queryFn: () => listPartnerTypes(activeTypeFilters),
    enabled: open,
  });
  const refresh = async () => {
    await Promise.all([
      queryClient.invalidateQueries({
        queryKey: businessPartnerKeys.partner(partner.partner_id),
      }),
      queryClient.invalidateQueries({
        queryKey: businessPartnerKeys.partnerLists(),
      }),
    ]);
  };
  const assign = useMutation({
    mutationFn: () =>
      assignBusinessPartnerType(partner.partner_id, selectedTypeId),
    onSuccess: async () => {
      await refresh();
      setSelectedTypeId("");
      toast.success("Partner type assigned.");
    },
  });
  const remove = useMutation({
    mutationFn: (partnerTypeId: string) =>
      removeBusinessPartnerType(partner.partner_id, partnerTypeId),
    onMutate: setRemovingId,
    onSuccess: async () => {
      await refresh();
      toast.success("Partner type removed.");
    },
    onSettled: () => setRemovingId(undefined),
  });
  const assignedIds = useMemo(
    () =>
      new Set(
        detail.data?.partner_types.map((item) => item.partner_type_id) ?? [],
      ),
    [detail.data?.partner_types],
  );
  const typeOptions =
    availableTypes.data?.items
      .filter((item) => !assignedIds.has(item.partner_type_id))
      .map((item) => ({
        value: item.partner_type_id,
        label: `${item.name} (${item.code})`,
      })) ?? [];

  return (
    <DialogFrame
      open={open}
      title="Business types"
      description={`${partner.name} (${partner.code})`}
      pending={assign.isPending || remove.isPending}
      onOpenChange={onOpenChange}
    >
      <div className="p-5 sm:p-6">
        {detail.isPending || availableTypes.isPending ? (
          <div className="grid min-h-48 place-items-center text-sm text-slate-600">
            <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
            Loading business types…
          </div>
        ) : detail.isError || availableTypes.isError ? (
          <div>
            <p
              role="alert"
              className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
            >
              {detail.error?.message ?? availableTypes.error?.message}
            </p>
            <Button
              className="mt-4"
              variant="secondary"
              onClick={() => {
                void detail.refetch();
                void availableTypes.refetch();
              }}
            >
              Try again
            </Button>
          </div>
        ) : (
          <>
            {canWrite && partner.is_active ? (
              <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
                <p className="text-sm font-semibold text-slate-900">
                  Assign another type
                </p>
                <div className="mt-3 flex flex-col gap-2 sm:flex-row">
                  <Select
                    ariaLabel="Partner type"
                    value={selectedTypeId}
                    options={typeOptions}
                    onValueChange={setSelectedTypeId}
                    placeholder={
                      typeOptions.length
                        ? "Select a partner type"
                        : "All active types assigned"
                    }
                    disabled={!typeOptions.length || assign.isPending}
                    className="flex-1"
                  />
                  <Button
                    type="button"
                    disabled={!selectedTypeId || assign.isPending}
                    onClick={() => assign.mutate()}
                  >
                    {assign.isPending ? (
                      <LoaderCircle className="size-4 animate-spin" />
                    ) : (
                      <Plus className="size-4" />
                    )}
                    Assign
                  </Button>
                </div>
              </div>
            ) : !partner.is_active ? (
              <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
                Reactivate this partner before assigning new types.
              </p>
            ) : null}

            {assign.error || remove.error ? (
              <p
                role="alert"
                className="mt-4 rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
              >
                {assign.error?.message ?? remove.error?.message}
              </p>
            ) : null}

            <div className="mt-5">
              <p className="text-xs font-bold tracking-wide text-slate-500 uppercase">
                Assigned types
              </p>
              {detail.data?.partner_types.length ? (
                <ul className="mt-2 divide-y divide-slate-200 rounded-xl border border-slate-200">
                  {detail.data.partner_types.map((item) => (
                    <li
                      key={item.partner_type_id}
                      className="flex items-center justify-between gap-3 p-3"
                    >
                      <div className="flex min-w-0 items-center gap-3">
                        <div className="grid size-9 shrink-0 place-items-center rounded-lg bg-cyan-50 text-cyan-800">
                          <Tag className="size-4" />
                        </div>
                        <div className="min-w-0">
                          <p className="truncate text-sm font-semibold text-slate-950">
                            {item.name}
                          </p>
                          <p className="font-mono text-xs text-slate-500">
                            {item.code}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-1">
                        <StatusBadge
                          tone={item.is_active ? "success" : "neutral"}
                        >
                          {item.is_active ? "Active" : "Inactive"}
                        </StatusBadge>
                        {canWrite ? (
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            aria-label={`Remove ${item.name}`}
                            className="text-rose-700"
                            disabled={remove.isPending}
                            onClick={() => remove.mutate(item.partner_type_id)}
                          >
                            {removingId === item.partner_type_id ? (
                              <LoaderCircle className="size-4 animate-spin" />
                            ) : (
                              <Trash2 className="size-4" />
                            )}
                          </Button>
                        ) : null}
                      </div>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="mt-2 rounded-xl border border-dashed border-slate-300 p-5 text-center text-sm text-slate-500">
                  No business types assigned yet.
                </p>
              )}
            </div>
          </>
        )}
        <div className="mt-6 flex justify-end border-t border-slate-200 pt-5">
          <Dialog.Close asChild>
            <Button type="button" variant="secondary">
              Done
            </Button>
          </Dialog.Close>
        </div>
      </div>
    </DialogFrame>
  );
}
