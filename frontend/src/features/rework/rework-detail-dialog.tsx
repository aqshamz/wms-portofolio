"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { exceptionTime } from "@/features/inbound-exceptions/inbound-exception-types";
import { quarantineKeys } from "@/features/quarantine/quarantine-api";
import { inspectionKeys } from "@/features/quality-inspections/quality-inspection-api";
import { getReworkTask, reworkKeys } from "./rework-api";
import { ReworkActionForm } from "./rework-action-form";
import {
  allowedReworkActions,
  reworkAssignee,
  reworkLabel,
  reworkTone,
  type ReworkAction,
  type ReworkCapabilities,
  type ReworkTask,
} from "./rework-types";

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
export function ReworkDetailDialog({
  taskId,
  capabilities,
  onOpenChange,
}: {
  taskId: string;
  capabilities: ReworkCapabilities;
  onOpenChange: (open: boolean) => void;
}) {
  const client = useQueryClient();
  const [action, setAction] = useState<ReworkAction>();
  const [busy, setBusy] = useState(false);
  const query = useQuery({
    queryKey: reworkKeys.detail(taskId),
    queryFn: () => getReworkTask(taskId),
    refetchOnWindowFocus: false,
  });
  const task = query.data;
  const actions = task ? allowedReworkActions(task, capabilities) : [];
  const done = (updated: ReworkTask) => {
    client.setQueryData(reworkKeys.detail(updated.rework_task_id), updated);
    void client.invalidateQueries({ queryKey: reworkKeys.all });
    void client.invalidateQueries({ queryKey: quarantineKeys.all });
    void client.invalidateQueries({ queryKey: inspectionKeys.all });
    setAction(undefined);
    setBusy(false);
  };
  return (
    <OperationDialog
      title={taskId}
      description="Rework instructions, execution and reinspection"
      closeLabel="Close rework task"
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
            Loading rework task…
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
            <StatusBadge tone={reworkTone(task.task_status_code)}>
              {reworkLabel(task.task_status_code)}
            </StatusBadge>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-3">
              <Detail label="Item" value={task.item_code} />
              <Detail
                label="Priority"
                value={reworkLabel(task.task_priority_code)}
              />
              <Detail
                label="Assigned account"
                value={reworkAssignee(task, capabilities.accountId)}
              />
              <Detail
                label="Planned quantity"
                value={`${task.planned_qty} ${task.base_uom_code}`}
              />
              <Detail
                label="Completed quantity"
                value={`${task.completed_qty} ${task.base_uom_code}`}
              />
              <Detail
                label="Created"
                value={exceptionTime(task.created_at, capabilities.timezone)}
              />
              <Detail
                label="Started"
                value={
                  task.started_at
                    ? exceptionTime(task.started_at, capabilities.timezone)
                    : null
                }
              />
              <Detail
                label="Completed at"
                value={
                  task.completed_at
                    ? exceptionTime(task.completed_at, capabilities.timezone)
                    : null
                }
              />
              <Detail label="Source balance" value={task.source_balance_id} />
              <Detail label="Quarantine case" value={task.quarantine_case_id} />
              <Detail
                label="Disposition"
                value={task.quarantine_disposition_id}
              />
            </dl>
            <section>
              <h3 className="font-bold text-slate-950">Work instructions</h3>
              <p className="mt-2 rounded-xl bg-cyan-50 p-4 text-sm break-words whitespace-pre-wrap text-cyan-950">
                {task.work_instructions || "No instructions recorded."}
              </p>
            </section>
            {task.result_notes ? (
              <section>
                <h3 className="font-bold text-slate-950">Result notes</h3>
                <p className="mt-2 rounded-xl bg-slate-50 p-4 text-sm break-words whitespace-pre-wrap text-slate-700">
                  {task.result_notes}
                </p>
              </section>
            ) : null}
            <div
              className="flex flex-wrap gap-3 text-sm font-semibold text-cyan-800"
              inert={busy || undefined}
            >
              <Link
                href={`/inbound/quarantine?${new URLSearchParams({ owner: task.owner_id, warehouse: task.warehouse_id, case: task.quarantine_case_id })}`}
                className="underline underline-offset-4"
              >
                Open quarantine case
              </Link>
              {task.reinspection_id ? (
                <Link
                  href={`/inbound/quality-inspections?${new URLSearchParams({ owner: task.owner_id, warehouse: task.warehouse_id, inspection: task.reinspection_id })}`}
                  className="underline underline-offset-4"
                >
                  Open reinspection
                </Link>
              ) : null}
            </div>
            {action && actions.includes(action) ? (
              <ReworkActionForm
                key={`${task.rework_task_id}:${task.version_no}:${action}`}
                task={task}
                action={action}
                capabilities={capabilities}
                onDone={done}
                onCancel={() => setAction(undefined)}
                onBusyChange={setBusy}
              />
            ) : actions.length ? (
              <div className="flex flex-wrap gap-2">
                {actions.map((name) => (
                  <Button
                    key={name}
                    disabled={query.isFetching || busy}
                    onClick={() => setAction(name)}
                  >
                    {name === "start" ? "Start rework" : "Complete rework"}
                  </Button>
                ))}
              </div>
            ) : (
              <p className="rounded-xl bg-amber-50 p-4 text-sm text-amber-950">
                {task.task_status_code === "COMPLETED"
                  ? "Rework is complete. Continue through the linked reinspection; stock is not available yet."
                  : !capabilities.canRework
                    ? "This account has read-only access. INBOUND.REWORK is required to process tasks."
                    : task.assigned_to &&
                        task.assigned_to !== capabilities.accountId
                      ? "This task is assigned to another account. Only its assigned worker can process it."
                      : "No execution action is available in this task status."}
              </p>
            )}
          </>
        ) : null}
      </div>
    </OperationDialog>
  );
}
