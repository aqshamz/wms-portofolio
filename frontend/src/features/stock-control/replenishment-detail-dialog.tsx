"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import { Textarea } from "@/components/ui/textarea";
import {
  assignReplenishment,
  cancelReplenishment,
  completeReplenishment,
  getReplenishment,
  listReplenishmentAssignees,
  startReplenishment,
  stockControlKeys,
} from "./stock-control-api";
import { replenishmentLabel, replenishmentTone } from "./replenishment-types";

type Action = "assign" | "start" | "complete" | "cancel";

function today(timezone: string) {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}

function Detail({ label, value }: { label: string; value?: string | null }) {
  return (
    <div>
      <dt className="text-xs font-semibold text-slate-500">{label}</dt>
      <dd className="mt-1 text-sm break-words text-slate-950">
        {value || "—"}
      </dd>
    </div>
  );
}

export function ReplenishmentDetailDialog({
  taskId,
  accountId,
  timezone,
  onOpenChange,
}: {
  taskId: string;
  accountId: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [action, setAction] = useState<Action>();
  const [assigneeSearch, setAssigneeSearch] = useState("");
  const [account, setAccount] = useState("");
  const [businessDate, setBusinessDate] = useState(() => today(timezone));
  const [reason, setReason] = useState("");
  const query = useQuery({
    queryKey: stockControlKeys.replenishment(taskId),
    queryFn: () => getReplenishment(taskId),
    refetchOnWindowFocus: false,
  });
  const task = query.data;
  const assignees = useQuery({
    queryKey: stockControlKeys.replenishmentAssignees(taskId, assigneeSearch),
    queryFn: () => listReplenishmentAssignees(taskId, assigneeSearch),
    enabled: action === "assign",
  });
  const mutation = useMutation({
    mutationFn: async () => {
      if (!task || !action) throw new Error("Task action is unavailable.");
      if (action === "assign") {
        if (!account) throw new Error("Select an account.");
        return assignReplenishment(
          task.replenishment_task_id,
          task.version_no,
          account,
        );
      }
      if (action === "start")
        return startReplenishment(task.replenishment_task_id, task.version_no);
      if (action === "complete") {
        if (!businessDate) throw new Error("Business date is required.");
        return completeReplenishment(
          task.replenishment_task_id,
          task.version_no,
          task.source_balance_version_no,
          businessDate,
        );
      }
      if (!reason.trim()) throw new Error("Cancellation reason is required.");
      return cancelReplenishment(
        task.replenishment_task_id,
        task.version_no,
        task.source_balance_version_no,
        reason.trim(),
      );
    },
    onSuccess: (updated) => {
      queryClient.setQueryData(
        stockControlKeys.replenishment(updated.replenishment_task_id),
        updated,
      );
      void queryClient.invalidateQueries({ queryKey: stockControlKeys.all });
      void queryClient.invalidateQueries({ queryKey: ["inventory"] });
      toast.success(
        action === "complete"
          ? "Replenishment completed and stock moved into the pick face."
          : action === "cancel"
            ? "Replenishment cancelled and its reservation released."
            : action === "assign"
              ? "Replenishment assigned."
              : "Replenishment started.",
      );
      setAction(undefined);
    },
    onError: (error) => toast.error(error.message),
  });
  const beforeStart =
    task?.task_status_code === "OPEN" || task?.task_status_code === "ASSIGNED";
  const isAssignee = task?.assigned_to === accountId;
  const canStart =
    beforeStart && (task?.task_status_code === "OPEN" || isAssignee);
  const canComplete = task?.task_status_code === "IN_PROGRESS" && isAssignee;
  const canCancel =
    beforeStart || (task?.task_status_code === "IN_PROGRESS" && isAssignee);

  return (
    <OperationDialog
      title={taskId}
      description="Reserved stock, assignment and replenishment execution"
      busy={mutation.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close replenishment task"
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={query.isFetching}
          onClick={() => void query.refetch()}
        >
          <RefreshCw className="size-4" /> Refresh task
        </Button>
        {query.isPending ? (
          <p className="flex items-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-4 animate-spin" /> Loading task…
          </p>
        ) : null}
        {query.error ? (
          <p
            role="alert"
            className="rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
          >
            {query.error.message}
          </p>
        ) : null}
        {task ? (
          <>
            <div className="flex flex-wrap gap-2">
              <StatusBadge tone={replenishmentTone(task.task_status_code)}>
                {replenishmentLabel(task.task_status_code)}
              </StatusBadge>
              <StatusBadge>{task.task_priority_code}</StatusBadge>
              <StatusBadge tone="info">
                {task.inventory_status_code}
              </StatusBadge>
            </div>
            <section className="grid gap-3 rounded-xl border border-cyan-200 bg-cyan-50/50 p-4 sm:grid-cols-[1fr_auto_1fr]">
              <div>
                <p className="text-xs font-semibold text-slate-500">
                  Reserve source
                </p>
                <p className="mt-1 font-bold">{task.source_location_code}</p>
              </div>
              <p className="self-center font-semibold text-cyan-950">
                {task.planned_qty} {task.uom_code} →
              </p>
              <div>
                <p className="text-xs font-semibold text-slate-500">
                  Forward pick face
                </p>
                <p className="mt-1 font-bold">{task.target_location_code}</p>
              </div>
            </section>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2 lg:grid-cols-3">
              <Detail
                label="Item"
                value={`${task.item_code} · ${task.item_name}`}
              />
              <Detail label="Lot" value={task.lot_number} />
              <Detail label="Serial" value={task.serial_number} />
              <Detail
                label="Handling unit"
                value={task.handling_unit_barcode}
              />
              <Detail
                label="Assigned account"
                value={task.assigned_display_name || task.assigned_username}
              />
              <Detail
                label="Completed"
                value={`${task.completed_qty} ${task.uom_code}`}
              />
              <Detail label="Notes" value={task.notes} />
              <Detail label="Movement" value={task.inventory_movement_id} />
              <Detail
                label="Cancellation reason"
                value={task.cancellation_reason}
              />
            </dl>
            {!action ? (
              <div className="flex flex-wrap gap-2">
                {beforeStart ? (
                  <Button
                    variant="secondary"
                    onClick={() => setAction("assign")}
                  >
                    Assign
                  </Button>
                ) : null}
                {canStart ? (
                  <Button onClick={() => setAction("start")}>Start</Button>
                ) : null}
                {canComplete ? (
                  <Button onClick={() => setAction("complete")}>
                    Complete
                  </Button>
                ) : null}
                {canCancel ? (
                  <Button
                    variant="secondary"
                    className="text-rose-700"
                    onClick={() => setAction("cancel")}
                  >
                    Cancel
                  </Button>
                ) : null}
              </div>
            ) : (
              <form
                className="space-y-4 rounded-xl border border-slate-200 p-4"
                onSubmit={(event) => {
                  event.preventDefault();
                  mutation.mutate();
                }}
              >
                <h3 className="font-bold">
                  {action === "assign"
                    ? "Assign worker"
                    : action === "start"
                      ? "Start replenishment"
                      : action === "complete"
                        ? "Complete replenishment"
                        : "Cancel replenishment"}
                </h3>
                {action === "assign" ? (
                  <>
                    <FormField
                      label="Search accounts"
                      htmlFor="replenishment-assignee-search"
                    >
                      <Input
                        id="replenishment-assignee-search"
                        value={assigneeSearch}
                        onChange={(event) =>
                          setAssigneeSearch(event.target.value)
                        }
                      />
                    </FormField>
                    <FormField
                      label="Assigned account"
                      htmlFor="replenishment-assignee"
                      required
                    >
                      <Select
                        id="replenishment-assignee"
                        ariaLabel="Replenishment assigned account"
                        className="mt-2"
                        value={account}
                        options={(assignees.data?.items ?? []).map((row) => ({
                          value: row.account_id,
                          label: `${row.display_name} (@${row.username})`,
                        }))}
                        placeholder={
                          assignees.isPending
                            ? "Loading accounts…"
                            : "Select account"
                        }
                        onValueChange={setAccount}
                      />
                    </FormField>
                    <p className="text-xs text-slate-500">
                      Only active accounts with matching scope and
                      INVENTORY.MOVE are listed.
                    </p>
                  </>
                ) : null}
                {action === "start" ? (
                  <p className="text-sm text-slate-700">
                    Starting claims an open task for you. Only its assignee can
                    complete it.
                  </p>
                ) : null}
                {action === "complete" ? (
                  <>
                    <p className="text-sm text-slate-700">
                      Confirm the full planned quantity has physically moved to{" "}
                      {task.target_location_code}. Partial completion is not
                      supported.
                    </p>
                    <FormField
                      label="Business date"
                      htmlFor="replenishment-business-date"
                      required
                    >
                      <Input
                        id="replenishment-business-date"
                        type="date"
                        value={businessDate}
                        onChange={(event) =>
                          setBusinessDate(event.target.value)
                        }
                      />
                    </FormField>
                  </>
                ) : null}
                {action === "cancel" ? (
                  <FormField
                    label="Cancellation reason"
                    htmlFor="replenishment-cancel-reason"
                    required
                  >
                    <Textarea
                      id="replenishment-cancel-reason"
                      maxLength={4000}
                      value={reason}
                      onChange={(event) => setReason(event.target.value)}
                    />
                  </FormField>
                ) : null}
                {mutation.error ? (
                  <p
                    role="alert"
                    className="rounded-lg bg-rose-50 p-3 text-sm text-rose-900"
                  >
                    {mutation.error.message}
                  </p>
                ) : null}
                <div className="flex flex-wrap justify-end gap-2">
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => setAction(undefined)}
                  >
                    Back
                  </Button>
                  <Button
                    type="submit"
                    className={
                      action === "cancel"
                        ? "bg-rose-700 hover:bg-rose-600"
                        : undefined
                    }
                    disabled={mutation.isPending}
                  >
                    {mutation.isPending ? (
                      <LoaderCircle className="size-4 animate-spin" />
                    ) : null}
                    Confirm
                  </Button>
                </div>
              </form>
            )}
            {task.task_status_code === "ASSIGNED" && !isAssignee ? (
              <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
                This task is assigned to another account; they must start it, or
                you can reassign it.
              </p>
            ) : null}
          </>
        ) : null}
      </div>
    </OperationDialog>
  );
}
