"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Textarea } from "@/components/ui/textarea";
import { ApiError } from "@/lib/api/client";
import { transitionReworkTask } from "./rework-api";
import { reworkResultSchema, type ReworkResultValues } from "./rework-schema";
import {
  allowedReworkActions,
  type ReworkAction,
  type ReworkCapabilities,
  type ReworkTask,
} from "./rework-types";

export function ReworkActionForm({
  task,
  action,
  capabilities,
  onDone,
  onCancel,
  onBusyChange,
}: {
  task: ReworkTask;
  action: ReworkAction;
  capabilities: ReworkCapabilities;
  onDone: (task: ReworkTask) => void;
  onCancel: () => void;
  onBusyChange: (busy: boolean) => void;
}) {
  const [decision, setDecision] = useState<ReworkResultValues>();
  const form = useForm<ReworkResultValues>({
    resolver: zodResolver(reworkResultSchema),
    defaultValues: { result_notes: task.result_notes || "" },
  });
  const mutation = useMutation({
    mutationFn: (fields: ReworkResultValues) => {
      if (!allowedReworkActions(task, capabilities).includes(action))
        throw new Error(
          "This task cannot be processed by your account. Refresh it first.",
        );
      return transitionReworkTask(task.rework_task_id, action, {
        expected_version: task.version_no,
        ...(action === "complete" && fields.result_notes
          ? { result_notes: fields.result_notes }
          : {}),
      });
    },
    onSuccess: (updated) => {
      toast.success(
        action === "start"
          ? "Rework started and assigned to you."
          : "Rework completed. Reinspection created.",
      );
      onDone(updated);
    },
    onError: (error) => toast.error(error.message),
    onSettled: () => onBusyChange(false),
  });
  const pending = mutation.isPending;
  return (
    <form
      className="space-y-4 rounded-xl border border-cyan-200 bg-cyan-50/30 p-4"
      onSubmit={form.handleSubmit(
        (fields) => {
          if (mutation.isPending || decision) return;
          mutation.reset();
          setDecision(fields);
        },
        () => toast.error("Please check the highlighted fields."),
      )}
    >
      <h3 className="font-bold text-slate-950">
        {action === "start" ? "Start rework" : "Complete rework"}
      </h3>
      <p className="text-sm text-slate-700">
        {action === "start"
          ? "Starting claims this task for you. Carry out the recorded work instructions before completing it."
          : `Completion records the entire planned quantity (${task.planned_qty} ${task.base_uom_code}) and creates a child quality inspection. Stock stays QC pending until the inspection and subsequent inbound steps are completed.`}
      </p>
      {action === "complete" ? (
        <FormField
          htmlFor="rework-result-notes"
          label="Result notes"
          error={form.formState.errors.result_notes?.message}
        >
          <Textarea
            id="rework-result-notes"
            disabled={pending || Boolean(decision)}
            aria-invalid={Boolean(form.formState.errors.result_notes)}
            {...form.register("result_notes")}
            placeholder="Describe the work completed and any observations"
          />
        </FormField>
      ) : null}
      {decision ? (
        <div className="space-y-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950">
          <p>
            {action === "start"
              ? "Start this task and assign it to yourself?"
              : `Confirm that all ${task.planned_qty} ${task.base_uom_code} have been reworked. Completion cannot be edited here; QC still needs to reinspect the stock.`}
          </p>
          {action === "complete" && decision.result_notes ? (
            <p className="break-words whitespace-pre-wrap">
              {decision.result_notes}
            </p>
          ) : null}
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              disabled={pending}
              onClick={() => {
                if (pending) return;
                onBusyChange(true);
                mutation.mutate(decision);
              }}
            >
              {pending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : null}
              {action === "start" ? "Confirm start" : "Confirm completion"}
            </Button>
            <Button
              type="button"
              variant="secondary"
              disabled={pending}
              onClick={() => setDecision(undefined)}
            >
              Back
            </Button>
          </div>
        </div>
      ) : (
        <div className="flex flex-wrap gap-2">
          <Button type="submit">
            {action === "start" ? "Review start" : "Review completion"}
          </Button>
          <Button type="button" variant="secondary" onClick={onCancel}>
            Cancel action
          </Button>
        </div>
      )}
      {mutation.error ? (
        <p
          role="alert"
          className="rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
        >
          {mutation.error.message}
          {mutation.error instanceof ApiError && mutation.error.status === 409
            ? " Refresh the task before trying again."
            : ""}
        </p>
      ) : null}
    </form>
  );
}
