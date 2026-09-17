"use client";

import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { ApiError } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  assignPutaway,
  cancelPutaway,
  completePutaway,
  listPutawayAssignees,
  listPutawayTargets,
  putawayKeys,
  retargetPutaway,
  reversePutaway,
  startPutaway,
} from "./putaway-api";
import {
  businessDateToday,
  putawayActionSchema,
  type PutawayActionValues,
} from "./putaway-schema";
import {
  allowedPutawayActions,
  type PutawayAction,
  type PutawayCapabilities,
  type PutawayTask,
} from "./putaway-types";

export const actionLabels: Record<PutawayAction, string> = {
  start: "Start task",
  assign: "Assign account",
  retarget: "Change target",
  complete: "Complete putaway",
  cancel: "Cancel task",
  reverse: "Reverse putaway",
};

function actionExplanation(action: PutawayAction, task: PutawayTask) {
  if (action === "start")
    return "Starting claims this task for your account. It does not move stock. Only the assignee can complete it.";
  if (action === "assign")
    return "Only active accounts with access to this owner and warehouse and effective INBOUND.PUTAWAY permission (from a role or a direct grant) are shown.";
  if (action === "retarget")
    return "Change the planned destination before work starts. Locations are filtered by storage eligibility and the applicable putaway strategy. The backend rechecks the rules when saving and completing.";
  if (action === "complete")
    return `Confirm that the full ${task.planned_qty} ${task.base_uom_code || "base units"} has physically moved from ${task.source_location_code} to ${task.target_location_code}. Completion posts the movement and makes stock AVAILABLE in storage. Partial completion is not supported.`;
  if (action === "cancel")
    return "Cancellation returns the task quantity to QC_PENDING at its source location and creates a replacement inspection. It does not make stock available or physically relocate it.";
  return `Confirm that stock is physically returning from ${task.target_location_code} to ${task.source_location_code}. Reversal is allowed only while the task quantity remains fully available and unreserved. It returns stock to QC_PENDING and creates a replacement inspection.`;
}

export function PutawayActionForm({
  action,
  task,
  capabilities,
  onBusyChange,
  onDone,
  onBack,
}: {
  action: PutawayAction;
  task: PutawayTask;
  capabilities: PutawayCapabilities;
  onBusyChange: (busy: boolean) => void;
  onDone: (task: PutawayTask) => void;
  onBack: () => void;
}) {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [selectedOption, setSelectedOption] = useState<{
    value: string;
    label: string;
  }>();
  const form = useForm<PutawayActionValues>({
    resolver: zodResolver(putawayActionSchema(action)),
    defaultValues: {
      business_date: businessDateToday(capabilities.timezone),
      reason: "",
      account_id: "",
      target_location_id: "",
    },
  });
  const accountId = useWatch({ control: form.control, name: "account_id" });
  const targetId = useWatch({
    control: form.control,
    name: "target_location_id",
  });
  const filters = { search, page, pageSize: 20 };
  const assignees = useQuery({
    queryKey: putawayKeys.assignees(task.putaway_task_id, filters),
    queryFn: () => listPutawayAssignees(task.putaway_task_id, filters),
    enabled: action === "assign" && capabilities.canAssign,
  });
  const targets = useQuery({
    queryKey: putawayKeys.targets(task.putaway_task_id, filters),
    queryFn: () => listPutawayTargets(task.putaway_task_id, filters),
    enabled: action === "retarget" && capabilities.canAssign,
  });
  const lookup = action === "assign" ? assignees : targets;
  const options =
    action === "assign"
      ? (assignees.data?.items ?? []).map((row) => ({
          value: row.account_id,
          label: `${row.display_name} (@${row.username})`,
        }))
      : (targets.data?.items ?? []).map((row) => ({
          value: row.location_id,
          label: `${row.code} · ${row.zone_code} · ${row.location_type_code}`,
        }));
  if (
    selectedOption &&
    !options.some((option) => option.value === selectedOption.value)
  )
    options.unshift(selectedOption);
  const posting =
    action === "complete" || action === "cancel" || action === "reverse";
  const balanceVersion =
    action === "reverse"
      ? task.result_balance_version_no
      : task.source_balance_version_no;
  const mutation = useMutation({
    mutationFn: (values: PutawayActionValues) => {
      if (!allowedPutawayActions(task, capabilities).includes(action))
        throw new Error(
          "This action is not allowed for this account and task status.",
        );
      const id = task.putaway_task_id;
      if (action === "start") return startPutaway(id, task.version_no);
      if (action === "assign")
        return assignPutaway(id, task.version_no, values.account_id);
      if (action === "retarget")
        return retargetPutaway(id, task.version_no, values.target_location_id);
      if (!balanceVersion)
        throw new Error(
          "Stock version is unavailable. Refresh the task before posting.",
        );
      const request = {
        expected_version: task.version_no,
        expected_balance_version: balanceVersion,
        business_date: values.business_date,
      };
      if (action === "complete") return completePutaway(id, request);
      if (action === "cancel")
        return cancelPutaway(id, { ...request, reason: values.reason });
      return reversePutaway(id, { ...request, reason: values.reason });
    },
    onSuccess: (updated) => {
      toast.success(
        action === "complete"
          ? "Putaway completed. Stock is now available in storage."
          : action === "cancel" || action === "reverse"
            ? "Stock returned to QC pending. Replacement inspection created."
            : action === "start"
              ? "Putaway task started."
              : action === "assign"
                ? "Putaway account assigned."
                : "Putaway target updated.",
      );
      onDone(updated);
    },
    onError: (error) => toast.error(error.message),
    onSettled: () => onBusyChange(false),
  });

  return (
    <form
      className={`space-y-4 rounded-xl border p-4 ${action === "cancel" || action === "reverse" ? "border-rose-200 bg-rose-50/50" : "border-cyan-200 bg-cyan-50/30"}`}
      onSubmit={form.handleSubmit(
        (values) => {
          onBusyChange(true);
          mutation.mutate(values);
        },
        () => toast.error("Please check the highlighted fields."),
      )}
    >
      <h3 className="font-bold text-slate-950">{actionLabels[action]}</h3>
      <p className="text-sm text-slate-700">
        {actionExplanation(action, task)}
      </p>
      {posting && !balanceVersion ? (
        <p
          role="alert"
          className="rounded-lg bg-amber-50 p-3 text-sm text-amber-900"
        >
          Stock version is unavailable. Restart the updated backend, then
          refresh the task before posting.
        </p>
      ) : null}
      <fieldset disabled={mutation.isPending} className="space-y-4">
        {action === "assign" || action === "retarget" ? (
          <>
            <FormField
              label={
                action === "assign"
                  ? "Search accounts"
                  : "Search eligible storage locations"
              }
              htmlFor="putaway-lookup-search"
            >
              <Input
                id="putaway-lookup-search"
                maxLength={160}
                value={search}
                placeholder={
                  action === "assign"
                    ? "Name or username"
                    : "Location or zone code"
                }
                onChange={(event) => {
                  setSearch(event.target.value);
                  setPage(1);
                }}
              />
            </FormField>
            <FormField
              label={
                action === "assign" ? "Assigned account" : "New putaway target"
              }
              htmlFor="putaway-selection"
              required
              error={
                action === "assign"
                  ? form.formState.errors.account_id?.message
                  : form.formState.errors.target_location_id?.message
              }
            >
              <Select
                id="putaway-selection"
                ariaLabel={
                  action === "assign"
                    ? "Assigned account"
                    : "New putaway target"
                }
                className="mt-2"
                value={action === "assign" ? accountId : targetId}
                options={options}
                placeholder={
                  lookup.isPending ? "Loading options…" : "Select an option"
                }
                disabled={
                  lookup.isPending || lookup.isError || mutation.isPending
                }
                invalid={Boolean(
                  action === "assign"
                    ? form.formState.errors.account_id
                    : form.formState.errors.target_location_id,
                )}
                onValueChange={(id) => {
                  setSelectedOption(
                    options.find((option) => option.value === id),
                  );
                  form.setValue(
                    action === "assign" ? "account_id" : "target_location_id",
                    id,
                    { shouldValidate: true },
                  );
                }}
              />
            </FormField>
            {lookup.isSuccess && !lookup.data.items.length ? (
              <p className="text-sm text-amber-900">
                {action === "assign"
                  ? "No active accounts with INBOUND.PUTAWAY permission and matching owner and warehouse access were found."
                  : "No storage locations match the active putaway strategy and your search. Check the owner/warehouse strategy’s category, location-type and zone rules."}
              </p>
            ) : null}
            {lookup.error ? (
              <p
                role="alert"
                className="rounded-lg bg-rose-50 p-3 text-sm text-rose-900"
              >
                {lookup.error.message}
              </p>
            ) : null}
            {lookup.data && lookup.data.total_pages > 1 ? (
              <div className="flex flex-wrap items-center gap-2 text-xs">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => setPage(page - 1)}
                >
                  Previous options
                </Button>
                <span>
                  Page {page} of {lookup.data.total_pages}
                </span>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  disabled={page >= lookup.data.total_pages}
                  onClick={() => setPage(page + 1)}
                >
                  Next options
                </Button>
              </div>
            ) : null}
          </>
        ) : null}
        {posting ? (
          <FormField
            label="Business date"
            htmlFor="putaway-business-date"
            required
            error={form.formState.errors.business_date?.message}
          >
            <Input
              id="putaway-business-date"
              type="date"
              invalid={Boolean(form.formState.errors.business_date)}
              {...form.register("business_date")}
            />
          </FormField>
        ) : null}
        {action === "cancel" || action === "reverse" ? (
          <FormField
            label="Reason"
            htmlFor="putaway-reason"
            required
            error={form.formState.errors.reason?.message}
          >
            <textarea
              id="putaway-reason"
              className="mt-2 min-h-24 w-full rounded-xl border border-slate-300 bg-white p-3 text-sm"
              maxLength={4000}
              aria-invalid={Boolean(form.formState.errors.reason)}
              {...form.register("reason")}
            />
          </FormField>
        ) : null}
      </fieldset>
      {mutation.error ? (
        <p
          role="alert"
          className="rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
        >
          {mutation.error.message}
          {mutation.error instanceof ApiError && mutation.error.status === 409
            ? " Refresh the task to check its latest status and stock version."
            : ""}
        </p>
      ) : null}
      <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
        <Button
          type="button"
          variant="secondary"
          disabled={mutation.isPending}
          onClick={onBack}
        >
          Back
        </Button>
        <Button
          type="submit"
          disabled={mutation.isPending || (posting && !balanceVersion)}
        >
          {mutation.isPending ? (
            <LoaderCircle className="size-4 animate-spin" />
          ) : null}
          Confirm {actionLabels[action].toLowerCase()}
        </Button>
      </div>
    </form>
  );
}
