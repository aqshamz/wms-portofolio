"use client";

import { useState } from "react";
import Link from "next/link";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import Decimal from "decimal.js";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  listLocations,
  listLocationTypes,
  storageLayoutKeys,
} from "@/features/storage-layout/storage-layout-api";
import {
  cancelInspection,
  completeInspection,
  getInspection,
  inspectionKeys,
} from "./quality-inspection-api";
import {
  cancelInspectionSchema,
  completeInspectionSchema,
  quantityTotal,
  type CompleteInspectionValues,
} from "./quality-inspection-schema";
import {
  inspectionState,
  inspectionTone,
  type QualityInspection,
} from "./quality-inspection-types";
import { InspectionDialog } from "./inspection-dialog";

function Detail({ name, value }: { name: string; value?: string | null }) {
  return (
    <div>
      <dt className="text-xs font-semibold text-slate-500">{name}</dt>
      <dd className="mt-1 text-sm break-words text-slate-900">
        {value || "—"}
      </dd>
    </div>
  );
}

function InspectionCompletion({
  inspection,
  onBusyChange,
  onDone,
}: {
  inspection: QualityInspection;
  onBusyChange: (busy: boolean) => void;
  onDone: (updated: QualityInspection) => void;
}) {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [selectedLocation, setSelectedLocation] = useState<{
    value: string;
    label: string;
  }>();
  const [confirmation, setConfirmation] = useState<CompleteInspectionValues>();
  const indivisible = Boolean(
    inspection.is_indivisible ||
    inspection.serial_no ||
    inspection.handling_unit_barcode,
  );
  const form = useForm<CompleteInspectionValues>({
    resolver: zodResolver(
      completeInspectionSchema(inspection.inspected_qty, indivisible),
    ),
    defaultValues: {
      passed_qty: inspection.inspected_qty,
      failed_qty: "0",
      putaway_target_location_id: "",
      notes: inspection.notes ?? "",
    },
  });
  const passed = useWatch({ control: form.control, name: "passed_qty" });
  const failed = useWatch({ control: form.control, name: "failed_qty" });
  const target = useWatch({
    control: form.control,
    name: "putaway_target_location_id",
  });
  const total = quantityTotal(passed, failed);
  const hasPassed = /^\d+(\.\d+)?$/.test(passed) && new Decimal(passed).gt(0);
  const filters = {
    warehouseId: inspection.warehouse_id,
    active: "active" as const,
    search,
    page,
    pageSize: 100,
  };
  const locations = useQuery({
    queryKey: storageLayoutKeys.locations(filters),
    queryFn: () => listLocations(filters),
    enabled: hasPassed,
  });
  const types = useQuery({
    queryKey: storageLayoutKeys.locationTypes("active"),
    queryFn: () => listLocationTypes("active"),
    enabled: hasPassed,
  });
  const storageTypes = new Set(
    types.data
      ?.filter((type) => type.is_active && type.allows_storage)
      .map((type) => type.location_type_id),
  );
  const options = (locations.data?.items ?? [])
    .filter(
      (location) =>
        location.warehouse_id === inspection.warehouse_id &&
        location.is_active &&
        !location.is_locked &&
        storageTypes.has(location.location_type_id),
    )
    .map((location) => ({
      value: location.location_id,
      label: `${location.code}${location.zone_code ? ` · ${location.zone_code}` : ""}`,
    }));
  if (
    selectedLocation &&
    !options.some((option) => option.value === selectedLocation.value)
  )
    options.unshift(selectedLocation);
  const mutation = useMutation({
    mutationFn: (values: CompleteInspectionValues) => {
      if (!inspection.source_balance_version_no)
        throw new Error(
          "Stock version is unavailable. Refresh the inspection before completing.",
        );
      return completeInspection(inspection.inspection_id, {
        expected_version: inspection.version_no,
        expected_balance_version: inspection.source_balance_version_no,
        passed_qty: values.passed_qty,
        failed_qty: values.failed_qty,
        putaway_target_location_id: new Decimal(values.passed_qty).gt(0)
          ? values.putaway_target_location_id
          : undefined,
        notes: values.notes || undefined,
      });
    },
    onSuccess: (updated) => {
      toast.success("Inspection completed. Outcome tasks have been created.");
      onDone(updated);
    },
    onError: (error) => toast.error(error.message),
    onSettled: () => onBusyChange(false),
  });
  return (
    <form
      className="space-y-4 rounded-xl border border-slate-200 p-4"
      onSubmit={form.handleSubmit(
        (values) => setConfirmation(values),
        () => {
          setConfirmation(undefined);
          toast.error(
            "Please check the highlighted quantities and storage location.",
          );
        },
      )}
    >
      <h3 className="font-bold text-slate-950">Record inspection result</h3>
      {!inspection.source_balance_version_no ? (
        <p
          role="alert"
          className="rounded-lg bg-amber-50 p-3 text-sm text-amber-900"
        >
          Stock version is unavailable. Refresh this inspection before
          completing it.
        </p>
      ) : null}
      <p className="text-sm text-slate-600">
        Inspected: {inspection.inspected_qty} {inspection.base_uom_code}.{" "}
        {indivisible
          ? "This serial or handling-unit batch must pass or fail in full."
          : "Passed and failed quantities must account for the entire batch."}
      </p>
      <fieldset
        disabled={mutation.isPending || Boolean(confirmation)}
        className="space-y-4"
      >
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={() => {
              form.setValue("passed_qty", inspection.inspected_qty);
              form.setValue("failed_qty", "0", { shouldValidate: true });
            }}
          >
            Pass all
          </Button>
          <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={() => {
              form.setValue("passed_qty", "0");
              form.setValue("failed_qty", inspection.inspected_qty, {
                shouldValidate: true,
              });
              form.setValue("putaway_target_location_id", "", {
                shouldValidate: true,
              });
            }}
          >
            Fail all
          </Button>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField
            label={`Passed quantity (${inspection.base_uom_code || "base units"})`}
            htmlFor="qc-passed"
            required
            error={form.formState.errors.passed_qty?.message}
          >
            <Input
              id="qc-passed"
              inputMode="decimal"
              maxLength={30}
              invalid={Boolean(form.formState.errors.passed_qty)}
              {...form.register("passed_qty")}
            />
          </FormField>
          <FormField
            label={`Failed quantity (${inspection.base_uom_code || "base units"})`}
            htmlFor="qc-failed"
            required
            error={form.formState.errors.failed_qty?.message}
          >
            <Input
              id="qc-failed"
              inputMode="decimal"
              maxLength={30}
              invalid={Boolean(form.formState.errors.failed_qty)}
              {...form.register("failed_qty")}
            />
          </FormField>
        </div>
        <p
          role="status"
          className={`rounded-lg p-3 text-sm ${total !== undefined && new Decimal(total).eq(inspection.inspected_qty) ? "bg-emerald-50 text-emerald-900" : "bg-amber-50 text-amber-900"}`}
        >
          Accounted quantity: {total ?? "Invalid quantity"} /{" "}
          {inspection.inspected_qty} {inspection.base_uom_code}
        </p>
        {hasPassed ? (
          <div className="space-y-3">
            <FormField
              label="Search storage locations"
              htmlFor="qc-storage-search"
            >
              <Input
                id="qc-storage-search"
                value={search}
                maxLength={160}
                placeholder="Location code"
                onChange={(event) => {
                  setSearch(event.target.value);
                  setPage(1);
                }}
              />
            </FormField>
            <FormField
              label="Putaway target"
              htmlFor="qc-target"
              required
              error={form.formState.errors.putaway_target_location_id?.message}
            >
              <Select
                id="qc-target"
                ariaLabel="Putaway target"
                className="mt-2"
                value={target}
                options={options}
                disabled={
                  locations.isPending ||
                  types.isPending ||
                  types.isError ||
                  locations.isError
                }
                invalid={Boolean(
                  form.formState.errors.putaway_target_location_id,
                )}
                placeholder="Select storage location"
                onValueChange={(value) => {
                  setSelectedLocation(
                    options.find((option) => option.value === value),
                  );
                  form.setValue("putaway_target_location_id", value, {
                    shouldValidate: true,
                  });
                }}
              />
            </FormField>
            <p className="text-xs text-slate-600">
              Only active, unlocked storage-capable locations in this warehouse
              are shown. The backend also enforces the item’s configured putaway
              strategy.
            </p>
            {locations.data && locations.data.total_pages > 1 ? (
              <div className="flex items-center gap-2 text-xs">
                <Button
                  type="button"
                  size="sm"
                  variant="ghost"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                >
                  Previous locations
                </Button>
                <span>
                  {page} / {locations.data.total_pages}
                </span>
                <Button
                  type="button"
                  size="sm"
                  variant="ghost"
                  disabled={page >= locations.data.total_pages}
                  onClick={() => setPage(page + 1)}
                >
                  Next locations
                </Button>
              </div>
            ) : null}
            {locations.isSuccess && types.isSuccess && !options.length ? (
              <p className="text-sm text-amber-800">
                No eligible storage locations on this page. Try another search
                or location page.
              </p>
            ) : null}
            {locations.error || types.error ? (
              <p role="alert" className="text-sm text-rose-800">
                {(locations.error ?? types.error)?.message}
              </p>
            ) : null}
          </div>
        ) : null}
        <FormField
          label="Inspection notes"
          htmlFor="qc-notes"
          error={form.formState.errors.notes?.message}
        >
          <textarea
            id="qc-notes"
            maxLength={4000}
            className="mt-2 min-h-24 w-full rounded-xl border border-slate-300 p-3 text-sm"
            {...form.register("notes")}
          />
        </FormField>
      </fieldset>
      {confirmation ? (
        <div className="space-y-3 rounded-xl border border-amber-200 bg-amber-50 p-4">
          <p className="text-sm text-amber-950">
            Complete with {confirmation.passed_qty} passed and{" "}
            {confirmation.failed_qty} failed? Passed stock will create a putaway
            task; failed stock will create a quarantine case. Stock remains at
            its received location until physically moved. This result cannot be
            edited.
          </p>
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              disabled={mutation.isPending}
              onClick={() => {
                onBusyChange(true);
                mutation.mutate(confirmation);
              }}
            >
              {mutation.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              Confirm completion
            </Button>
            <Button
              type="button"
              variant="secondary"
              disabled={mutation.isPending}
              onClick={() => {
                setConfirmation(undefined);
                mutation.reset();
              }}
            >
              Back to result
            </Button>
          </div>
        </div>
      ) : (
        <Button type="submit" disabled={!inspection.source_balance_version_no}>
          Complete inspection
        </Button>
      )}
      {mutation.error ? (
        <p
          role="alert"
          className="rounded-lg bg-rose-50 p-3 text-sm text-rose-900"
        >
          {mutation.error.message} If stock changed, refresh before trying
          again.
        </p>
      ) : null}
    </form>
  );
}

export function InspectionDetailDialog({
  inspectionId,
  canQC,
  canCancel,
  onOpenChange,
  onSelect,
}: {
  inspectionId: string;
  canQC: boolean;
  canCancel: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (id: string) => void;
}) {
  const queryClient = useQueryClient();
  const [busy, setBusy] = useState(false);
  const [cancelling, setCancelling] = useState(false);
  const form = useForm<{ reason: string }>({
    resolver: zodResolver(cancelInspectionSchema),
    defaultValues: { reason: "" },
  });
  const query = useQuery({
    queryKey: inspectionKeys.detail(inspectionId),
    queryFn: () => getInspection(inspectionId),
    refetchOnWindowFocus: false,
  });
  const inspection = query.data;
  const pending =
    inspection && !inspection.inspected_at && !inspection.cancelled_at;
  const updated = (value: QualityInspection) => {
    queryClient.setQueryData(inspectionKeys.detail(value.inspection_id), value);
    void queryClient.invalidateQueries({ queryKey: inspectionKeys.all });
    void queryClient.invalidateQueries({ queryKey: ["putaway-tasks"] });
    void queryClient.invalidateQueries({ queryKey: ["quarantine-cases"] });
    setCancelling(false);
    setBusy(false);
  };
  const cancel = useMutation({
    mutationFn: (values: { reason: string }) => {
      if (!inspection || !canCancel || !pending)
        throw new Error("Only pending inspections can be cancelled.");
      return cancelInspection(
        inspection.inspection_id,
        inspection.version_no,
        values.reason,
      );
    },
    onSuccess: (value) => {
      toast.success("Inspection cancelled. A replacement inspection is ready.");
      updated(value);
    },
    onError: (error) => toast.error(error.message),
    onSettled: () => setBusy(false),
  });
  return (
    <InspectionDialog
      title={inspectionId}
      description="Quality inspection details and inventory outcomes"
      busy={busy}
      onOpenChange={onOpenChange}
    >
      {query.isPending ? (
        <p className="flex items-center gap-2 text-sm text-slate-600">
          <LoaderCircle className="size-4 animate-spin" />
          Loading inspection…
        </p>
      ) : null}
      {query.error ? (
        <p
          role="alert"
          className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
        >
          {query.error.message}
        </p>
      ) : null}
      <div className="space-y-5">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          disabled={busy || query.isFetching}
          onClick={() => {
            setCancelling(false);
            cancel.reset();
            void query.refetch();
          }}
        >
          <RefreshCw className="size-4" />
          Refresh inspection
        </Button>
        {inspection ? (
          <>
            <StatusBadge tone={inspectionTone(inspection)}>
              {inspectionState(inspection)}
            </StatusBadge>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2 lg:grid-cols-3">
              <Detail name="Receipt" value={inspection.receipt_id} />
              <Detail name="Batch" value={inspection.receipt_inventory_id} />
              <Detail name="Item" value={inspection.item_code} />
              <Detail name="Lot" value={inspection.lot_number} />
              <Detail
                name="Serial / handling unit"
                value={[inspection.serial_no, inspection.handling_unit_barcode]
                  .filter(Boolean)
                  .join(" / ")}
              />
              <Detail
                name="Received location"
                value={inspection.location_code}
              />
              <Detail
                name="Inspected (base units)"
                value={`${inspection.inspected_qty} ${inspection.base_uom_code ?? ""}`}
              />
              <Detail
                name="Passed / failed"
                value={`${inspection.passed_qty} / ${inspection.failed_qty}`}
              />
              <Detail
                name="Created"
                value={new Date(inspection.created_at).toLocaleString("en-ID")}
              />
              <Detail
                name="Inspected at"
                value={
                  inspection.inspected_at
                    ? new Date(inspection.inspected_at).toLocaleString("en-ID")
                    : null
                }
              />
              <Detail name="Notes" value={inspection.notes} />
            </dl>
            {inspection.parent_inspection_id ? (
              <div className="text-sm text-slate-600">
                Replacement / reinspection of{" "}
                <Button
                  variant="ghost"
                  size="sm"
                  disabled={busy}
                  onClick={() => onSelect(inspection.parent_inspection_id!)}
                >
                  {inspection.parent_inspection_id}
                </Button>
              </div>
            ) : null}
            {inspection.putaway_task ? (
              <section className="rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-950">
                <h3 className="font-bold">Putaway task created</h3>
                <Button asChild variant="secondary" className="mt-3">
                  <Link
                    href={`/inbound/putaway?${new URLSearchParams({ owner: inspection.owner_id, warehouse: inspection.warehouse_id, task: inspection.putaway_task.putaway_task_id })}`}
                  >
                    Open putaway task
                  </Link>
                </Button>
                <p className="mt-1 break-all">
                  {inspection.putaway_task.putaway_task_id}
                </p>
                <p className="mt-1">
                  {inspection.putaway_task.planned_qty}{" "}
                  {inspection.base_uom_code} →{" "}
                  {inspection.putaway_task.target_location_code} ·{" "}
                  {inspection.putaway_task.task_status_code}
                </p>
              </section>
            ) : null}
            {inspection.quarantine_case ? (
              <section className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-950">
                <h3 className="font-bold">Quarantine case created</h3>
                <p className="mt-1 break-all">
                  {inspection.quarantine_case.quarantine_case_id}
                </p>
                <p className="mt-1">
                  {inspection.quarantine_case.quarantine_qty}{" "}
                  {inspection.base_uom_code} ·{" "}
                  {inspection.quarantine_case.status_code}
                </p>
                <Link
                  className="mt-3 inline-block text-sm font-semibold text-cyan-900 underline"
                  href={`/inbound/quarantine?${new URLSearchParams({ owner: inspection.owner_id, warehouse: inspection.warehouse_id, case: inspection.quarantine_case.quarantine_case_id })}`}
                >
                  Open quarantine case
                </Link>
              </section>
            ) : null}
            {inspection.cancelled_at ? (
              <section className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950">
                <h3 className="font-bold">
                  Cancelled{" "}
                  {new Date(inspection.cancelled_at).toLocaleString("en-ID")}
                </h3>
                <p className="mt-1 whitespace-pre-wrap">
                  {inspection.cancellation_reason}
                </p>
                <p className="mt-2">
                  QC has not been skipped. Stock is still QC pending and must be
                  inspected again.
                </p>
                {inspection.replacement_inspection_id ? (
                  <Button
                    variant="secondary"
                    className="mt-3"
                    onClick={() =>
                      onSelect(inspection.replacement_inspection_id!)
                    }
                  >
                    Open replacement inspection
                  </Button>
                ) : null}
              </section>
            ) : null}
            {pending && canQC && !cancelling ? (
              <InspectionCompletion
                key={`${inspection.inspection_id}:${inspection.version_no}:${inspection.source_balance_version_no}`}
                inspection={inspection}
                onBusyChange={setBusy}
                onDone={updated}
              />
            ) : null}
            {pending && !canQC ? (
              <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
                INBOUND.QC is required to record and complete this inspection.
              </p>
            ) : null}
            {pending && canCancel ? (
              cancelling ? (
                <form
                  className="space-y-3 rounded-xl border border-rose-200 bg-rose-50 p-4"
                  onSubmit={form.handleSubmit(
                    (values) => {
                      setBusy(true);
                      cancel.mutate(values);
                    },
                    () => toast.error("Enter a cancellation reason."),
                  )}
                >
                  <p className="text-sm text-rose-900">
                    Cancelling creates a replacement inspection. It does not
                    release stock or waive quality control.
                  </p>
                  <FormField
                    label="Cancellation reason"
                    htmlFor="qc-cancel-reason"
                    required
                    error={form.formState.errors.reason?.message}
                  >
                    <textarea
                      id="qc-cancel-reason"
                      maxLength={4000}
                      disabled={busy}
                      className="mt-2 min-h-24 w-full rounded-xl border border-slate-300 bg-white p-3 text-sm"
                      {...form.register("reason")}
                    />
                  </FormField>
                  <div className="flex flex-wrap gap-2">
                    <Button type="submit" disabled={busy}>
                      Confirm cancellation
                    </Button>
                    <Button
                      type="button"
                      variant="secondary"
                      disabled={busy}
                      onClick={() => setCancelling(false)}
                    >
                      Back
                    </Button>
                  </div>
                  {cancel.error ? (
                    <p role="alert" className="text-sm text-rose-900">
                      {cancel.error.message}
                    </p>
                  ) : null}
                </form>
              ) : (
                <Button
                  variant="secondary"
                  disabled={busy}
                  onClick={() => setCancelling(true)}
                >
                  Cancel inspection
                </Button>
              )
            ) : null}
          </>
        ) : null}
      </div>
    </InspectionDialog>
  );
}
