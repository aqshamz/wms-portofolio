"use client";

import { useEffect } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertTriangle,
  CheckCircle2,
  Copy,
  LoaderCircle,
  X,
} from "lucide-react";
import { Controller, useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { listBusinessPartners } from "@/features/business-partners/business-partner-api";
import {
  documentNumberingKeys,
  generateDocumentId,
  replaceDocumentNumberRule,
} from "@/features/document-numbering/document-numbering-api";
import {
  allocationFormSchema,
  numberRuleFormSchema,
  type AllocationFormValues,
  type NumberRuleFormValues,
} from "@/features/document-numbering/document-numbering-schema";
import type {
  DocumentNumberRule,
  GeneratedDocumentId,
} from "@/features/document-numbering/document-numbering-types";
import type { DocumentType } from "@/features/document-workflows/document-workflow-types";
import { listWarehouses } from "@/features/warehouses/warehouse-api";

function jakartaToday() {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Jakarta",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(new Date());
  const value = Object.fromEntries(
    parts.map((part) => [part.type, part.value]),
  );
  return `${value.year}-${value.month}-${value.day}`;
}

function DialogHeader({
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
          aria-label="Close numbering form"
          className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
        >
          <X className="size-5" />
        </button>
      </Dialog.Close>
    </div>
  );
}

export function NumberRuleDialog({
  documentType,
  currentRule,
  onOpenChange,
}: {
  documentType: DocumentType;
  currentRule?: DocumentNumberRule;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<NumberRuleFormValues>({
    resolver: zodResolver(numberRuleFormSchema),
    defaultValues: {
      prefix: currentRule?.prefix ?? documentType.code,
      separator: currentRule?.separator ?? "-",
      sequence_length: currentRule?.sequence_length ?? 6,
      effective_from: jakartaToday(),
      include_partner_code: currentRule?.include_partner_code ?? false,
      include_warehouse_code: currentRule?.include_warehouse_code ?? true,
    },
  });
  const values = useWatch({ control });
  const save = useMutation({
    mutationFn: (form: NumberRuleFormValues) =>
      replaceDocumentNumberRule(documentType.document_type_id, {
        prefix: form.prefix.trim().toUpperCase(),
        separator: form.separator,
        sequence_length: form.sequence_length,
        effective_from: form.effective_from,
        include_partner_code: form.include_partner_code,
        include_warehouse_code: form.include_warehouse_code,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: documentNumberingKeys.all,
      });
      toast.success("Numbering rule activated.");
      onOpenChange(false);
    },
  });

  useEffect(() => {
    reset({
      prefix: currentRule?.prefix ?? documentType.code,
      separator: currentRule?.separator ?? "-",
      sequence_length: currentRule?.sequence_length ?? 6,
      effective_from: jakartaToday(),
      include_partner_code: currentRule?.include_partner_code ?? false,
      include_warehouse_code: currentRule?.include_warehouse_code ?? true,
    });
  }, [currentRule, documentType.code, reset]);

  const previewParts = [
    values.prefix?.trim().toUpperCase() || "PREFIX",
    values.include_partner_code ? "PARTNER" : "",
    values.include_warehouse_code ? "WAREHOUSE" : "",
    (values.effective_from || jakartaToday()).replaceAll("-", ""),
    "1".padStart(Number(values.sequence_length) || 6, "0"),
  ].filter(Boolean);
  const preview = previewParts.join(values.separator ?? "-");

  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!save.isPending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] max-w-2xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <DialogHeader
            title={
              currentRule ? "Replace numbering rule" : "Create numbering rule"
            }
            description={`Configure identifiers for ${documentType.name}.`}
          />
          <form
            noValidate
            className="p-5 sm:p-6"
            onSubmit={handleSubmit((form) => save.mutate(form))}
          >
            {currentRule ? (
              <div className="mb-5 flex gap-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-amber-950">
                <AlertTriangle className="mt-0.5 size-5 shrink-0" />
                <p className="text-sm leading-6">
                  Saving creates a new active rule and retires the current rule.
                  Numbering history is preserved and cannot be edited in place.
                </p>
              </div>
            ) : null}
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
                label="Prefix"
                htmlFor="number-prefix"
                required
                error={errors.prefix?.message}
              >
                <Input
                  id="number-prefix"
                  placeholder="PO"
                  invalid={Boolean(errors.prefix)}
                  {...register("prefix")}
                />
              </FormField>
              <FormField
                label="Separator"
                htmlFor="number-separator"
                error={errors.separator?.message}
              >
                <Input
                  id="number-separator"
                  maxLength={3}
                  placeholder="Leave blank for none"
                  invalid={Boolean(errors.separator)}
                  {...register("separator")}
                />
              </FormField>
              <FormField
                label="Sequence digits"
                htmlFor="sequence-length"
                required
                error={errors.sequence_length?.message}
              >
                <Input
                  id="sequence-length"
                  type="number"
                  min={3}
                  max={18}
                  invalid={Boolean(errors.sequence_length)}
                  {...register("sequence_length")}
                />
              </FormField>
              <FormField
                label="Effective from"
                htmlFor="effective-from"
                required
                error={errors.effective_from?.message}
              >
                <Input
                  id="effective-from"
                  type="date"
                  max={jakartaToday()}
                  invalid={Boolean(errors.effective_from)}
                  {...register("effective_from")}
                />
              </FormField>
              <label className="rounded-xl border border-slate-200 p-4">
                <span className="flex items-center gap-3 text-sm font-semibold text-slate-900">
                  <input
                    type="checkbox"
                    className="size-4 accent-cyan-600"
                    {...register("include_partner_code")}
                  />
                  Include partner code
                </span>
                <span className="mt-2 block pl-7 text-xs leading-5 text-slate-500">
                  A business partner must be supplied when allocating an ID.
                </span>
              </label>
              <label className="rounded-xl border border-slate-200 p-4">
                <span className="flex items-center gap-3 text-sm font-semibold text-slate-900">
                  <input
                    type="checkbox"
                    className="size-4 accent-cyan-600"
                    {...register("include_warehouse_code")}
                  />
                  Include warehouse code
                </span>
                <span className="mt-2 block pl-7 text-xs leading-5 text-slate-500">
                  A warehouse must be supplied when allocating an ID.
                </span>
              </label>
            </div>
            <div className="mt-5 rounded-xl bg-slate-950 p-4 text-white">
              <p className="text-xs font-semibold tracking-wide text-cyan-300 uppercase">
                Format preview
              </p>
              <p className="mt-2 font-mono text-sm break-all sm:text-base">
                {preview}
              </p>
              <p className="mt-2 text-xs text-slate-400">
                Preview only. No sequence number is consumed.
              </p>
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
                  ? "Activating…"
                  : currentRule
                    ? "Replace and activate"
                    : "Create and activate"}
              </Button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function AllocateDocumentIdDialog({
  documentType,
  rule,
  onOpenChange,
}: {
  documentType: DocumentType;
  rule: DocumentNumberRule;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const partners = useQuery({
    queryKey: documentNumberingKeys.partners(),
    queryFn: () =>
      listBusinessPartners({
        ownerId: "",
        partnerTypeCode: "",
        search: "",
        active: "active",
        page: 1,
        pageSize: 100,
      }),
    enabled: rule.include_partner_code,
  });
  const warehouses = useQuery({
    queryKey: documentNumberingKeys.warehouses(),
    queryFn: () =>
      listWarehouses({ search: "", active: "active", page: 1, pageSize: 100 }),
    enabled: rule.include_warehouse_code,
  });
  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<AllocationFormValues>({
    resolver: zodResolver(allocationFormSchema),
    defaultValues: {
      business_date: jakartaToday(),
      partner_id: "none",
      warehouse_id: "none",
    },
  });
  const allocate = useMutation<
    GeneratedDocumentId,
    Error,
    AllocationFormValues
  >({
    mutationFn: (form) =>
      generateDocumentId(documentType.document_type_id, {
        business_date: form.business_date,
        partner_id: form.partner_id === "none" ? undefined : form.partner_id,
        warehouse_id:
          form.warehouse_id === "none" ? undefined : form.warehouse_id,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: documentNumberingKeys.all,
      });
      toast.success("Document ID allocated.");
    },
  });

  function submit(form: AllocationFormValues) {
    let valid = true;
    if (rule.include_partner_code && form.partner_id === "none") {
      setError("partner_id", {
        message: "Business partner is required by this rule.",
      });
      valid = false;
    }
    if (rule.include_warehouse_code && form.warehouse_id === "none") {
      setError("warehouse_id", {
        message: "Warehouse is required by this rule.",
      });
      valid = false;
    }
    if (valid) allocate.mutate(form);
  }

  const partnerOptions = [
    { value: "none", label: "Select business partner" },
    ...(partners.data?.items ?? []).map((partner) => ({
      value: partner.partner_id,
      label: `${partner.name} (${partner.code})`,
    })),
  ];
  const warehouseOptions = [
    { value: "none", label: "Select warehouse" },
    ...(warehouses.data?.items ?? []).map((warehouse) => ({
      value: warehouse.warehouse_id,
      label: `${warehouse.name} (${warehouse.code})`,
    })),
  ];

  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!allocate.isPending) onOpenChange(open);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] max-w-xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <DialogHeader
            title="Allocate document ID"
            description={`Consume the next ${documentType.name} sequence number.`}
          />
          {allocate.data ? (
            <div className="p-5 sm:p-6">
              <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-5 text-emerald-950">
                <CheckCircle2 className="size-7" />
                <p className="mt-4 text-sm font-semibold">
                  Document ID allocated
                </p>
                <p className="mt-2 font-mono text-lg font-bold break-all">
                  {allocate.data.document_id}
                </p>
                <p className="mt-2 text-xs">
                  Business date {allocate.data.business_date} · Sequence{" "}
                  {allocate.data.sequence_number}
                </p>
              </div>
              <div className="mt-5 flex flex-col gap-2 sm:flex-row sm:justify-end">
                <Button
                  variant="secondary"
                  onClick={() => {
                    void navigator.clipboard.writeText(
                      allocate.data!.document_id,
                    );
                    toast.success("Document ID copied.");
                  }}
                >
                  <Copy className="size-4" />
                  Copy ID
                </Button>
                <Button onClick={() => onOpenChange(false)}>Done</Button>
              </div>
            </div>
          ) : (
            <form
              noValidate
              className="p-5 sm:p-6"
              onSubmit={handleSubmit(submit)}
            >
              <div className="mb-5 flex gap-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-amber-950">
                <AlertTriangle className="mt-0.5 size-5 shrink-0" />
                <p className="text-sm leading-6">
                  <span className="font-semibold">This is not a preview.</span>{" "}
                  Allocation permanently advances the daily counter even if no
                  document is created afterward.
                </p>
              </div>
              {allocate.error ? (
                <p
                  role="alert"
                  className="mb-5 rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
                >
                  {allocate.error.message}
                </p>
              ) : null}
              <div className="grid gap-5">
                <FormField
                  label="Business date"
                  htmlFor="allocation-date"
                  required
                  error={errors.business_date?.message}
                >
                  <Input
                    id="allocation-date"
                    type="date"
                    min={rule.effective_from}
                    max={rule.effective_until}
                    invalid={Boolean(errors.business_date)}
                    {...register("business_date")}
                  />
                </FormField>
                {rule.include_partner_code ? (
                  <FormField
                    label="Business partner"
                    htmlFor="allocation-partner"
                    required
                    error={errors.partner_id?.message}
                  >
                    <Controller
                      name="partner_id"
                      control={control}
                      render={({ field }) => (
                        <Select
                          id="allocation-partner"
                          ariaLabel="Business partner"
                          value={field.value}
                          options={partnerOptions}
                          onValueChange={field.onChange}
                          disabled={partners.isPending}
                          invalid={Boolean(errors.partner_id)}
                        />
                      )}
                    />
                  </FormField>
                ) : null}
                {rule.include_warehouse_code ? (
                  <FormField
                    label="Warehouse"
                    htmlFor="allocation-warehouse"
                    required
                    error={errors.warehouse_id?.message}
                  >
                    <Controller
                      name="warehouse_id"
                      control={control}
                      render={({ field }) => (
                        <Select
                          id="allocation-warehouse"
                          ariaLabel="Warehouse"
                          value={field.value}
                          options={warehouseOptions}
                          onValueChange={field.onChange}
                          disabled={warehouses.isPending}
                          invalid={Boolean(errors.warehouse_id)}
                        />
                      )}
                    />
                  </FormField>
                ) : null}
              </div>
              <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
                <Dialog.Close asChild>
                  <Button type="button" variant="secondary">
                    Cancel
                  </Button>
                </Dialog.Close>
                <Button type="submit" disabled={allocate.isPending}>
                  {allocate.isPending ? (
                    <LoaderCircle className="size-4 animate-spin" />
                  ) : null}
                  {allocate.isPending ? "Allocating…" : "Allocate next ID"}
                </Button>
              </div>
            </form>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
