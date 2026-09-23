"use client";

import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select, type SelectOption } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { ApiError } from "@/lib/api/client";
import { businessDateToday } from "@/features/putaway/putaway-schema";
import {
  createDisposition,
  listQuarantineTargets,
  quarantineKeys,
} from "./quarantine-api";
import { dispositionSchema, type DispositionValues } from "./quarantine-schema";
import {
  canDecide,
  dispositionEffect,
  remainingQuantity,
  type DispositionType,
  type QuarantineCapabilities,
  type QuarantineCase,
} from "./quarantine-types";

export function DispositionForm({
  value,
  types,
  capabilities,
  onDone,
  onBusyChange,
}: {
  value: QuarantineCase;
  types: DispositionType[];
  capabilities: QuarantineCapabilities;
  onDone: (value: QuarantineCase) => void;
  onBusyChange: (busy: boolean) => void;
}) {
  const [decision, setDecision] = useState<DispositionValues>();
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [selection, setSelection] = useState<SelectOption<string>>();
  const form = useForm<DispositionValues>({
    resolver: zodResolver(dispositionSchema(value, types)),
    defaultValues: {
      disposition_type_code: "",
      disposition_qty: remainingQuantity(value),
      business_date: businessDateToday(capabilities.timezone),
      target_location_id: "",
      client_decision_reference: "",
      decision_notes: "",
      work_instructions: "",
    },
  });
  const typeCode = useWatch({
    control: form.control,
    name: "disposition_type_code",
  });
  const targetId = useWatch({
    control: form.control,
    name: "target_location_id",
  });
  const selectedType = types.find((row) => row.code === typeCode);
  const effect = dispositionEffect(selectedType);
  const filters = { search, page, pageSize: 20 };
  const targets = useQuery({
    queryKey: quarantineKeys.targets(value.quarantine_case_id, filters),
    queryFn: () => listQuarantineTargets(value.quarantine_case_id, filters),
    enabled: effect === "accept",
  });
  const options = (targets.data?.items ?? []).map((row) => ({
    value: row.location_id,
    label: `${row.code} · ${row.zone_code}`,
  }));
  if (
    selection &&
    targetId === selection.value &&
    !options.some((option) => option.value === selection.value)
  )
    options.unshift(selection);
  const snapshotReady = Boolean(
    value.quarantine_balance_version_no &&
    value.quarantine_balance_version_no > 0 &&
    value.available_qty !== undefined &&
    value.is_indivisible !== undefined,
  );
  const mutation = useMutation({
    mutationFn: (fields: DispositionValues) => {
      if (!canDecide(value, capabilities) || !snapshotReady)
        throw new Error("Refresh the case before recording a disposition.");
      const type = types.find(
        (row) => row.code === fields.disposition_type_code && row.is_active,
      );
      const action = dispositionEffect(type);
      if (!action) throw new Error("Disposition type is unavailable.");
      return createDisposition(value.quarantine_case_id, {
        expected_case_version: value.version_no,
        expected_balance_version: value.quarantine_balance_version_no!,
        disposition_type_code: fields.disposition_type_code,
        disposition_qty: fields.disposition_qty,
        business_date: fields.business_date,
        decided_at: new Date().toISOString(),
        ...(action === "accept"
          ? { target_location_id: fields.target_location_id }
          : {}),
        ...(action === "rework"
          ? { work_instructions: fields.work_instructions }
          : {}),
        ...(fields.client_decision_reference
          ? { client_decision_reference: fields.client_decision_reference }
          : {}),
        ...(fields.decision_notes
          ? { decision_notes: fields.decision_notes }
          : {}),
      });
    },
    onSuccess: (updated) => {
      toast.success("Disposition recorded. Inventory updated.");
      onDone(updated);
    },
    onError: (error) => toast.error(error.message),
    onSettled: () => onBusyChange(false),
  });
  const errors = form.formState.errors;
  const locked = mutation.isPending || Boolean(decision);
  return (
    <form
      className="space-y-4 rounded-xl border border-cyan-200 bg-cyan-50/30 p-4"
      onSubmit={form.handleSubmit(
        (fields) => {
          mutation.reset();
          setDecision(fields);
        },
        () => toast.error("Please check the highlighted fields."),
      )}
    >
      <h3 className="font-bold text-slate-950">Record disposition</h3>
      <p className="text-sm text-slate-700">
        Undecided: {remainingQuantity(value)}{" "}
        {value.base_uom_code || "base units"}. Unreserved quarantine stock:{" "}
        {value.available_qty ?? "Unavailable"}. Each confirmed decision posts
        inventory immediately and cannot be edited here.
      </p>
      {!snapshotReady ? (
        <p
          role="alert"
          className="rounded-lg bg-amber-50 p-3 text-sm text-amber-900"
        >
          Stock version or batch information is unavailable. Refresh this case
          before recording a disposition.
        </p>
      ) : null}
      {value.is_indivisible ? (
        <p className="rounded-lg bg-amber-50 p-3 text-sm text-amber-900">
          Serial-numbered or handling-unit stock cannot be split. Process the
          entire unreserved source quantity.
        </p>
      ) : null}
      <FormField
        label="Disposition type"
        htmlFor="disposition-type"
        required
        error={errors.disposition_type_code?.message}
      >
        <Select
          id="disposition-type"
          ariaLabel="Disposition type"
          value={typeCode}
          options={types
            .filter((row) => row.is_active && dispositionEffect(row))
            .map((row) => ({
              value: row.code,
              label: `${row.name} (${row.code})`,
            }))}
          disabled={locked}
          invalid={Boolean(errors.disposition_type_code)}
          ariaDescribedBy={
            errors.disposition_type_code ? "disposition-type-error" : undefined
          }
          placeholder="Select a decision"
          onValueChange={(code) => {
            form.setValue("disposition_type_code", code, {
              shouldValidate: true,
            });
            form.setValue("target_location_id", "");
            setSelection(undefined);
            setSearch("");
            setPage(1);
          }}
        />
      </FormField>
      {selectedType?.description ? (
        <p className="text-sm text-slate-600">{selectedType.description}</p>
      ) : null}
      {effect ? (
        <p className="rounded-lg bg-white p-3 text-sm text-slate-700">
          {effect === "accept"
            ? "Acceptance moves stock from quarantine into AVAILABLE at the selected storage location immediately. It does not create a putaway task. Confirm only after the physical move is ready."
            : effect === "rework"
              ? "Rework changes this quantity to QC_PENDING at its current location and creates a rework task. After rework is completed, it must pass a new quality inspection."
              : "This decision removes the quantity from warehouse inventory. Confirm the physical return or disposal before posting."}
        </p>
      ) : null}
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField
          label="Quantity (base units)"
          htmlFor="disposition-qty"
          required
          error={errors.disposition_qty?.message}
        >
          <Input
            id="disposition-qty"
            inputMode="decimal"
            disabled={locked}
            aria-invalid={Boolean(errors.disposition_qty)}
            aria-describedby={
              errors.disposition_qty ? "disposition-qty-error" : undefined
            }
            {...form.register("disposition_qty")}
          />
        </FormField>
        <FormField
          label="Business date"
          htmlFor="disposition-date"
          required
          error={errors.business_date?.message}
        >
          <Input
            id="disposition-date"
            type="date"
            disabled={locked}
            {...form.register("business_date")}
          />
        </FormField>
      </div>
      {effect === "accept" ? (
        <div className="space-y-3">
          <FormField
            label="Search storage targets"
            htmlFor="quarantine-target-search"
          >
            <Input
              id="quarantine-target-search"
              className="mt-1.5"
              maxLength={160}
              disabled={locked}
              value={search}
              placeholder="Location or zone code"
              onChange={(event) => {
                setSearch(event.target.value);
                setPage(1);
              }}
            />
          </FormField>
          <FormField
            label="Acceptance target"
            htmlFor="quarantine-target"
            required
            error={errors.target_location_id?.message}
          >
            <Select
              id="quarantine-target"
              ariaLabel="Acceptance target"
              value={targetId}
              options={options}
              disabled={locked || targets.isPending || targets.isError}
              placeholder={
                targets.isPending
                  ? "Loading eligible locations…"
                  : "Select storage location"
              }
              invalid={Boolean(errors.target_location_id)}
              ariaDescribedBy={
                errors.target_location_id
                  ? "quarantine-target-error"
                  : undefined
              }
              onValueChange={(id) => {
                form.setValue("target_location_id", id, {
                  shouldValidate: true,
                });
                setSelection(options.find((option) => option.value === id));
              }}
            />
          </FormField>
          <p className="text-xs text-slate-500">
            Only active, unlocked, storage-capable locations in this warehouse
            matching the item’s active putaway strategy are shown. Rules are
            rechecked when posting.
          </p>
          {targets.error ? (
            <div role="alert" className="text-sm text-rose-900">
              <p>{targets.error.message}</p>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                disabled={locked}
                onClick={() => void targets.refetch()}
              >
                Retry targets
              </Button>
            </div>
          ) : null}
          {targets.isSuccess && !targets.data.items.length ? (
            <p className="text-sm text-amber-900">
              No storage targets match the active strategy and your search.
            </p>
          ) : null}
          {targets.data && targets.data.total_pages > 1 ? (
            <div className="flex flex-wrap items-center gap-2 text-xs text-slate-600">
              <Button
                type="button"
                variant="secondary"
                size="sm"
                disabled={locked || page <= 1 || targets.isFetching}
                onClick={() => setPage(page - 1)}
              >
                Previous targets
              </Button>
              <span>
                Page {page} of {targets.data.total_pages}
              </span>
              <Button
                type="button"
                variant="secondary"
                size="sm"
                disabled={
                  locked ||
                  page >= targets.data.total_pages ||
                  targets.isFetching
                }
                onClick={() => setPage(page + 1)}
              >
                Next targets
              </Button>
            </div>
          ) : null}
        </div>
      ) : null}
      {effect === "rework" ? (
        <FormField
          label="Work instructions"
          htmlFor="disposition-instructions"
          required
          error={errors.work_instructions?.message}
        >
          <Textarea
            id="disposition-instructions"
            disabled={locked}
            rows={3}
            {...form.register("work_instructions")}
          />
        </FormField>
      ) : null}
      <FormField
        label="Decision reference (optional)"
        htmlFor="disposition-reference"
        error={errors.client_decision_reference?.message}
      >
        <Input
          id="disposition-reference"
          disabled={locked}
          placeholder="Owner approval or return reference"
          {...form.register("client_decision_reference")}
        />
      </FormField>
      <FormField
        label="Decision notes (optional)"
        htmlFor="disposition-notes"
        error={errors.decision_notes?.message}
      >
        <Textarea
          id="disposition-notes"
          disabled={locked}
          rows={3}
          {...form.register("decision_notes")}
        />
      </FormField>
      {decision ? (
        <section
          className="space-y-3 rounded-xl border border-amber-200 bg-amber-50 p-4"
          aria-label="Confirm quarantine disposition"
        >
          <p className="text-sm text-amber-950">
            Confirm {decision.disposition_type_code} for{" "}
            {decision.disposition_qty} {value.base_uom_code || "base units"}
            {effect === "accept"
              ? ` into ${selection?.label || decision.target_location_id}`
              : ""}
            ? This posts inventory immediately. The case closes when its entire
            quarantine quantity has been decided. Decision time is recorded when
            confirmed.
          </p>
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              disabled={mutation.isPending || !snapshotReady}
              onClick={() => {
                onBusyChange(true);
                mutation.mutate(decision);
              }}
            >
              {mutation.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              Confirm disposition
            </Button>
            <Button
              type="button"
              variant="secondary"
              disabled={mutation.isPending}
              onClick={() => setDecision(undefined)}
            >
              Back to decision
            </Button>
          </div>
        </section>
      ) : (
        <Button type="submit" disabled={!snapshotReady || mutation.isPending}>
          Review decision
        </Button>
      )}
      {mutation.error ? (
        <p
          role="alert"
          className="rounded-lg bg-rose-50 p-3 text-sm text-rose-900"
        >
          {mutation.error.message}
          {mutation.error instanceof ApiError && mutation.error.status === 409
            ? " Refresh this case and review the latest quantities before trying again."
            : ""}
        </p>
      ) : null}
    </form>
  );
}
