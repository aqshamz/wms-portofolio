"use client";

import { useEffect } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, X } from "lucide-react";
import { useForm, type UseFormRegisterReturn } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import {
  createOrganization,
  getOrganization,
  organizationKeys,
  updateOrganization,
} from "@/features/organizations/organization-api";
import {
  emptyOrganizationForm,
  organizationFormSchema,
  type OrganizationFormValues,
} from "@/features/organizations/organization-schema";
import type {
  CreateOrganizationRequest,
  Organization,
  UpdateOrganizationRequest,
} from "@/features/organizations/organization-types";
import { ApiError } from "@/lib/api/client";

function Field({
  label,
  name,
  error,
  required,
  children,
}: {
  label: string;
  name: string;
  error?: string;
  required?: boolean;
  children: React.ReactNode;
}) {
  return (
    <FormField label={label} htmlFor={name} error={error} required={required}>
      {children}
    </FormField>
  );
}

function TextInput({
  registration,
  error,
  ...props
}: React.InputHTMLAttributes<HTMLInputElement> & {
  registration: UseFormRegisterReturn;
  error?: string;
}) {
  return (
    <Input
      {...props}
      {...registration}
      invalid={Boolean(error)}
      aria-describedby={error ? `${props.id}-error` : undefined}
    />
  );
}

function optional(value: string) {
  const trimmed = value.trim();
  return trimmed || undefined;
}

function formValues(organization: Organization): OrganizationFormValues {
  return {
    code: organization.code,
    name: organization.name,
    legal_name: organization.legal_name ?? "",
    tax_number: organization.tax_number ?? "",
    timezone_name: organization.timezone_name,
    address_line_1: organization.address_line_1 ?? "",
    address_line_2: organization.address_line_2 ?? "",
    city: organization.city ?? "",
    province: organization.province ?? "",
    postal_code: organization.postal_code ?? "",
    country_code: organization.country_code ?? "",
    is_active: organization.is_active,
  };
}

function mutableRequest(
  values: OrganizationFormValues,
): Omit<CreateOrganizationRequest, "code"> {
  return {
    name: values.name.trim(),
    legal_name: optional(values.legal_name),
    tax_number: optional(values.tax_number),
    timezone_name: values.timezone_name.trim(),
    address_line_1: optional(values.address_line_1),
    address_line_2: optional(values.address_line_2),
    city: optional(values.city),
    province: optional(values.province),
    postal_code: optional(values.postal_code),
    country_code: optional(values.country_code)?.toUpperCase(),
  };
}

export function OrganizationFormDialog({
  open,
  organizationId,
  onOpenChange,
}: {
  open: boolean;
  organizationId?: string;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = organizationId !== undefined;
  const queryClient = useQueryClient();
  const detail = useQuery({
    queryKey: organizationKeys.detail(organizationId ?? "new"),
    queryFn: () => getOrganization(organizationId!),
    enabled: open && editing,
  });
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<OrganizationFormValues>({
    resolver: zodResolver(organizationFormSchema),
    defaultValues: emptyOrganizationForm,
  });
  const save = useMutation({
    mutationFn: async (values: OrganizationFormValues) => {
      const request = mutableRequest(values);
      if (!editing) {
        return createOrganization({
          code: values.code.trim().toUpperCase(),
          ...request,
        });
      }
      if (!detail.data) throw new Error("Organization details are not loaded.");

      const updateRequest: UpdateOrganizationRequest = {
        ...request,
        is_active: values.is_active,
        expected_updated_at: detail.data.updated_at,
      };
      return updateOrganization(detail.data.organization_id, updateRequest);
    },
    onSuccess: async (organization) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: organizationKeys.lists() }),
        queryClient.invalidateQueries({
          queryKey: organizationKeys.detail(organization.organization_id),
        }),
      ]);
      toast.success(
        editing ? "Organization updated." : "Organization created.",
      );
      onOpenChange(false);
    },
  });

  useEffect(() => {
    if (!open) return;
    if (!editing) reset(emptyOrganizationForm);
  }, [editing, open, reset]);

  useEffect(() => {
    if (open && detail.data) reset(formValues(detail.data));
  }, [detail.data, open, reset]);

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
                {editing ? "Edit organization" : "Create organization"}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                {editing
                  ? "Update the organization profile and operational address."
                  : "Add an organization that can own or operate warehouse data."}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100 hover:text-slate-950"
                aria-label="Close organization form"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>

          {editing && detail.isPending ? (
            <div className="grid min-h-72 place-items-center p-6 text-sm text-slate-600">
              <LoaderCircle className="mb-3 size-6 animate-spin text-cyan-700" />
              Loading organization…
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
                <Field
                  label="Code"
                  name="organization-code"
                  required
                  error={errors.code?.message}
                >
                  <TextInput
                    id="organization-code"
                    placeholder="Example: NEXA_ID"
                    readOnly={editing}
                    registration={register("code")}
                    error={errors.code?.message}
                  />
                </Field>
                <Field
                  label="Display name"
                  name="organization-name"
                  required
                  error={errors.name?.message}
                >
                  <TextInput
                    id="organization-name"
                    placeholder="Organization name"
                    autoFocus={!editing}
                    registration={register("name")}
                    error={errors.name?.message}
                  />
                </Field>
                <Field
                  label="Legal name"
                  name="organization-legal-name"
                  error={errors.legal_name?.message}
                >
                  <TextInput
                    id="organization-legal-name"
                    registration={register("legal_name")}
                    error={errors.legal_name?.message}
                  />
                </Field>
                <Field
                  label="Tax number"
                  name="organization-tax-number"
                  error={errors.tax_number?.message}
                >
                  <TextInput
                    id="organization-tax-number"
                    registration={register("tax_number")}
                    error={errors.tax_number?.message}
                  />
                </Field>
                <Field
                  label="Timezone"
                  name="organization-timezone"
                  required
                  error={errors.timezone_name?.message}
                >
                  <TextInput
                    id="organization-timezone"
                    list="timezone-options"
                    placeholder="Asia/Jakarta"
                    registration={register("timezone_name")}
                    error={errors.timezone_name?.message}
                  />
                  <datalist id="timezone-options">
                    <option value="Asia/Jakarta" />
                    <option value="Asia/Makassar" />
                    <option value="Asia/Jayapura" />
                    <option value="UTC" />
                  </datalist>
                </Field>
                <Field
                  label="Country code"
                  name="organization-country"
                  error={errors.country_code?.message}
                >
                  <TextInput
                    id="organization-country"
                    placeholder="ID"
                    maxLength={2}
                    registration={register("country_code")}
                    error={errors.country_code?.message}
                  />
                </Field>
                <div className="sm:col-span-2">
                  <Field
                    label="Address line 1"
                    name="organization-address-1"
                    error={errors.address_line_1?.message}
                  >
                    <TextInput
                      id="organization-address-1"
                      registration={register("address_line_1")}
                      error={errors.address_line_1?.message}
                    />
                  </Field>
                </div>
                <div className="sm:col-span-2">
                  <Field
                    label="Address line 2"
                    name="organization-address-2"
                    error={errors.address_line_2?.message}
                  >
                    <TextInput
                      id="organization-address-2"
                      registration={register("address_line_2")}
                      error={errors.address_line_2?.message}
                    />
                  </Field>
                </div>
                <Field
                  label="City"
                  name="organization-city"
                  error={errors.city?.message}
                >
                  <TextInput
                    id="organization-city"
                    registration={register("city")}
                    error={errors.city?.message}
                  />
                </Field>
                <Field
                  label="Province"
                  name="organization-province"
                  error={errors.province?.message}
                >
                  <TextInput
                    id="organization-province"
                    registration={register("province")}
                    error={errors.province?.message}
                  />
                </Field>
                <Field
                  label="Postal code"
                  name="organization-postal-code"
                  error={errors.postal_code?.message}
                >
                  <TextInput
                    id="organization-postal-code"
                    registration={register("postal_code")}
                    error={errors.postal_code?.message}
                  />
                </Field>
                {editing ? (
                  <label className="flex min-h-11 items-center gap-3 self-end rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
                    <input
                      type="checkbox"
                      className="size-4 accent-cyan-600"
                      {...register("is_active")}
                    />
                    Organization is active
                  </label>
                ) : null}
              </div>

              <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
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
                      : "Create organization"}
                </Button>
              </div>
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
