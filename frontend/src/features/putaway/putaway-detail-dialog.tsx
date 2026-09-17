"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { inspectionKeys } from "@/features/quality-inspections/quality-inspection-api";
import { getPutawayTask, putawayKeys } from "./putaway-api";
import { actionLabels, PutawayActionForm } from "./putaway-action-form";
import {
  allowedPutawayActions,
  putawayLabel,
  putawayTone,
  type PutawayAction,
  type PutawayCapabilities,
  type PutawayTask,
} from "./putaway-types";

function Detail({ label, value }: { label: string; value?: string | null }) {
  return (
    <div>
      <dt className="text-xs font-semibold text-slate-500">{label}</dt>
      <dd className="mt-1 text-sm break-words text-slate-900">
        {value || "—"}
      </dd>
    </div>
  );
}
function inspectionHref(task: PutawayTask, id: string) {
  return `/inbound/quality-inspections?${new URLSearchParams({ owner: task.owner_id, warehouse: task.warehouse_id, inspection: id })}`;
}

export function PutawayDetailDialog({
  taskId,
  capabilities,
  onOpenChange,
}: {
  taskId: string;
  capabilities: PutawayCapabilities;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [action, setAction] = useState<PutawayAction>();
  const [busy, setBusy] = useState(false);
  const query = useQuery({
    queryKey: putawayKeys.detail(taskId),
    queryFn: () => getPutawayTask(taskId),
    refetchOnWindowFocus: false,
  });
  const task = query.data;
  const actions = task ? allowedPutawayActions(task, capabilities) : [];
  const done = (updated: PutawayTask) => {
    queryClient.setQueryData(
      putawayKeys.detail(updated.putaway_task_id),
      updated,
    );
    void queryClient.invalidateQueries({ queryKey: putawayKeys.all });
    void queryClient.invalidateQueries({ queryKey: inspectionKeys.all });
    void queryClient.invalidateQueries({ queryKey: ["inventory"] });
    void queryClient.invalidateQueries({ queryKey: ["receipts", "balance"] });
    setAction(undefined);
    setBusy(false);
  };
  return (
    <OperationDialog
      title={taskId}
      description="Putaway task, movement details and recovery workflow"
      closeLabel="Close putaway task"
      busy={busy}
      onOpenChange={onOpenChange}
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={busy || query.isFetching}
          onClick={() => {
            setAction(undefined);
            void query.refetch();
          }}
        >
          <RefreshCw className="size-4" />
          Refresh task
        </Button>
        {query.isPending ? (
          <p className="flex items-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-4 animate-spin" />
            Loading putaway task…
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
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge tone={putawayTone(task.task_status_code)}>
                {putawayLabel(task.task_status_code)}
              </StatusBadge>
              <StatusBadge>{task.task_priority_code}</StatusBadge>
            </div>
            <section className="grid gap-3 rounded-xl border border-cyan-200 bg-cyan-50/50 p-4 sm:grid-cols-[1fr_auto_1fr]">
              <div>
                <p className="text-xs font-semibold text-slate-500">
                  From received / QC location
                </p>
                <p className="mt-1 font-bold break-words">
                  {task.source_location_code}
                </p>
              </div>
              <p className="self-center text-sm font-semibold text-cyan-900">
                {task.planned_qty} {task.base_uom_code || "base units"} →
              </p>
              <div>
                <p className="text-xs font-semibold text-slate-500">
                  To storage location
                </p>
                <p className="mt-1 font-bold break-words">
                  {task.target_location_code}
                </p>
              </div>
            </section>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2 lg:grid-cols-3">
              <Detail label="Item" value={task.item_code} />
              <Detail label="Lot" value={task.lot_number} />
              <Detail label="Batch" value={task.receipt_inventory_id} />
              <Detail
                label="Assigned account"
                value={task.assigned_display_name || task.assigned_to}
              />
              <Detail label="Handling unit" value={task.handling_unit_id} />
              <Detail
                label="Completed quantity"
                value={`${task.completed_qty} ${task.base_uom_code || ""}`}
              />
              <Detail
                label="Created"
                value={new Date(task.created_at).toLocaleString("en-ID")}
              />
              <Detail
                label="Started"
                value={
                  task.started_at
                    ? new Date(task.started_at).toLocaleString("en-ID")
                    : null
                }
              />
              <Detail
                label="Completed"
                value={
                  task.completed_at
                    ? new Date(task.completed_at).toLocaleString("en-ID")
                    : null
                }
              />
              <Detail
                label="Inventory movement"
                value={task.inventory_movement_id}
              />
              <Detail
                label="Resulting balance"
                value={task.resulting_balance_id}
              />
            </dl>
            <Button
              asChild
              variant="secondary"
              className={busy ? "pointer-events-none opacity-50" : undefined}
            >
              <Link
                href={inspectionHref(task, task.inspection_id)}
                aria-disabled={busy}
                tabIndex={busy ? -1 : undefined}
                onClick={(event) => {
                  if (busy) event.preventDefault();
                }}
              >
                View source inspection
              </Link>
            </Button>
            {task.task_status_code === "ASSIGNED" &&
            task.assigned_to !== capabilities.accountId ? (
              <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
                Only the assigned account can start this task. An account with
                INBOUND.ASSIGN can reassign it before work starts.
              </p>
            ) : null}
            {task.task_status_code === "IN_PROGRESS" &&
            task.assigned_to !== capabilities.accountId ? (
              <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
                This task is being worked by another account. Only its assignee
                can complete or cancel it while in progress.
              </p>
            ) : null}
            {task.task_status_code === "COMPLETED" ? (
              <p className="rounded-xl bg-emerald-50 p-3 text-sm text-emerald-900">
                The movement has been posted and stock was made available at the
                target location. Any later recovery must use reversal, not
                editing.
              </p>
            ) : null}
            {task.task_status_code === "CANCELLED" ||
            task.task_status_code === "REVERSED" ? (
              <section className="space-y-3 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950">
                <h3 className="font-bold">Stock returned to QC pending</h3>
                <p>
                  This task is final. Continue through the replacement
                  inspection; stock was not released as available.
                </p>
                {task.reversal_reason ? (
                  <p className="whitespace-pre-wrap">{task.reversal_reason}</p>
                ) : null}
                {task.reversed_at ? (
                  <p>
                    Reversed{" "}
                    {new Date(task.reversed_at).toLocaleString("en-ID")}
                  </p>
                ) : null}
                {task.reversal_movement_id ? (
                  <p className="break-all">
                    Reversal movement: {task.reversal_movement_id}
                  </p>
                ) : null}
                {task.replacement_inspection_id ? (
                  <Button asChild variant="secondary">
                    <Link
                      href={inspectionHref(
                        task,
                        task.replacement_inspection_id,
                      )}
                    >
                      Open replacement inspection
                    </Link>
                  </Button>
                ) : (
                  <p>
                    For older cancelled tasks, find the replacement in Quality
                    inspections using this batch or source inspection.
                  </p>
                )}
              </section>
            ) : null}
            {!action ? (
              <div className="flex flex-wrap gap-2">
                {actions.map((available) => (
                  <Button
                    key={available}
                    variant={
                      available === "start" || available === "complete"
                        ? "primary"
                        : "secondary"
                    }
                    onClick={() => setAction(available)}
                  >
                    {actionLabels[available]}
                  </Button>
                ))}
                {!actions.length &&
                !["CANCELLED", "REVERSED", "COMPLETED"].includes(
                  task.task_status_code,
                ) ? (
                  <p className="text-sm text-slate-600">
                    No actions are available for this account and task status.
                  </p>
                ) : null}
              </div>
            ) : actions.includes(action) ? (
              <PutawayActionForm
                key={`${task.putaway_task_id}:${task.version_no}:${task.source_balance_version_no}:${task.result_balance_version_no}:${action}`}
                task={task}
                action={action}
                capabilities={capabilities}
                onBusyChange={setBusy}
                onDone={done}
                onBack={() => setAction(undefined)}
              />
            ) : (
              <div className="space-y-3 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
                <p>
                  This task changed and the selected action is no longer
                  available. Check its latest status before continuing.
                </p>
                <Button
                  variant="secondary"
                  onClick={() => setAction(undefined)}
                >
                  Back to task
                </Button>
              </div>
            )}
          </>
        ) : null}
      </div>
    </OperationDialog>
  );
}
