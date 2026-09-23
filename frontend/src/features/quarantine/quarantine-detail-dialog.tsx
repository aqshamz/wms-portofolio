"use client";

import { useState } from "react";
import Link from "next/link";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { inspectionKeys } from "@/features/quality-inspections/quality-inspection-api";
import { reworkHref } from "@/features/rework/rework-types";
import {
  getQuarantineCase,
  listDispositionTypes,
  quarantineKeys,
} from "./quarantine-api";
import { DispositionForm } from "./disposition-form";
import {
  canDecide,
  quarantineLabel,
  quarantineTone,
  remainingQuantity,
  type QuarantineCapabilities,
  type QuarantineCase,
} from "./quarantine-types";

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
function inspectionHref(value: QuarantineCase, id: string) {
  return `/inbound/quality-inspections?${new URLSearchParams({ owner: value.owner_id, warehouse: value.warehouse_id, inspection: id })}`;
}
export function QuarantineDetailDialog({
  caseId,
  capabilities,
  onOpenChange,
}: {
  caseId: string;
  capabilities: QuarantineCapabilities;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [busy, setBusy] = useState(false);
  const query = useQuery({
    queryKey: quarantineKeys.detail(caseId),
    queryFn: () => getQuarantineCase(caseId),
    refetchOnWindowFocus: false,
  });
  const value = query.data;
  const editable = Boolean(value && canDecide(value, capabilities));
  const types = useQuery({
    queryKey: quarantineKeys.types,
    queryFn: listDispositionTypes,
    enabled: editable,
  });
  const supportedTypes = (types.data ?? []).filter(
    (type) =>
      type.is_active &&
      (type.requires_reinspection ||
        type.releases_to_available ||
        type.removes_inventory),
  );
  const done = (updated: QuarantineCase) => {
    queryClient.setQueryData(
      quarantineKeys.detail(updated.quarantine_case_id),
      updated,
    );
    void queryClient.invalidateQueries({ queryKey: quarantineKeys.all });
    void queryClient.invalidateQueries({ queryKey: inspectionKeys.all });
    void queryClient.invalidateQueries({ queryKey: ["inventory"] });
    void queryClient.invalidateQueries({ queryKey: ["rework-tasks"] });
    void queryClient.invalidateQueries({ queryKey: ["receipts", "balance"] });
    setBusy(false);
  };
  return (
    <OperationDialog
      title={caseId}
      description="Quarantine stock, disposition decisions and rework lineage"
      closeLabel="Close quarantine case"
      busy={busy}
      onOpenChange={onOpenChange}
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={busy || query.isFetching}
          onClick={() => {
            void query.refetch();
            if (editable) void types.refetch();
          }}
        >
          <RefreshCw className="size-4" />
          Refresh case
        </Button>
        {query.isPending ? (
          <p className="flex items-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-4 animate-spin" />
            Loading quarantine case…
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
        {value ? (
          <>
            <StatusBadge tone={quarantineTone(value.status_code)}>
              {quarantineLabel(value.status_code)}
            </StatusBadge>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2 lg:grid-cols-3">
              <Detail label="Item" value={value.item_code} />
              <Detail label="Lot" value={value.lot_number} />
              <Detail label="Current location" value={value.location_code} />
              <Detail
                label="Quarantined (base units)"
                value={`${value.quarantine_qty} ${value.base_uom_code || ""}`}
              />
              <Detail label="Decided quantity" value={value.disposed_qty} />
              <Detail
                label="Undecided quantity"
                value={remainingQuantity(value)}
              />
              <Detail
                label="Unreserved quarantine stock"
                value={value.available_qty}
              />
              <Detail
                label="Serial / handling unit"
                value={value.serial_no || value.handling_unit_barcode}
              />
              <Detail
                label="Receipt batch"
                value={value.receipt_inventory_id}
              />
              <Detail label="Opened" value={value.opened_at} />
              <Detail label="Closed" value={value.closed_at} />
              <Detail label="Case notes" value={value.notes} />
            </dl>
            <div className="flex flex-wrap gap-2" inert={busy || undefined}>
              <Link
                className="rounded-lg border border-slate-200 px-3 py-2 text-sm font-semibold text-cyan-900"
                href={inspectionHref(value, value.inspection_id)}
              >
                Open source inspection
              </Link>
              {value.parent_quarantine_case_id ? (
                <Link
                  className="rounded-lg border border-slate-200 px-3 py-2 text-sm font-semibold text-cyan-900"
                  href={`/inbound/quarantine?${new URLSearchParams({ owner: value.owner_id, warehouse: value.warehouse_id, case: value.parent_quarantine_case_id })}`}
                >
                  Open parent quarantine case
                </Link>
              ) : null}
            </div>
            {editable ? (
              types.isPending ? (
                <p className="text-sm text-slate-600">
                  Loading disposition types…
                </p>
              ) : types.error ? (
                <div
                  role="alert"
                  className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
                >
                  <p>{types.error.message}</p>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => void types.refetch()}
                  >
                    Retry disposition types
                  </Button>
                </div>
              ) : !supportedTypes.length ? (
                <p className="rounded-xl bg-amber-50 p-4 text-sm text-amber-900">
                  No active supported disposition types are configured.
                </p>
              ) : (
                <DispositionForm
                  key={`${value.quarantine_case_id}-${value.version_no}-${value.quarantine_balance_version_no}`}
                  value={value}
                  types={supportedTypes}
                  capabilities={capabilities}
                  onBusyChange={setBusy}
                  onDone={done}
                />
              )
            ) : (
              <p className="rounded-xl bg-slate-50 p-4 text-sm text-slate-600">
                {value.status_code === "CLOSED"
                  ? "This case is fully decided and closed. Rework and reinspection can still continue on the linked task."
                  : "This account has read-only quarantine access. INBOUND.QUARANTINE_DISPOSE is required to record decisions."}
              </p>
            )}
            <section className="space-y-3" inert={busy || undefined}>
              <h3 className="font-bold text-slate-950">
                Disposition history ({value.dispositions.length})
              </h3>
              {!value.dispositions.length ? (
                <p className="text-sm text-slate-500">
                  No disposition decisions recorded yet.
                </p>
              ) : (
                value.dispositions.map((disposition) => (
                  <article
                    key={disposition.quarantine_disposition_id}
                    className="space-y-3 rounded-xl border border-slate-200 p-4"
                  >
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <p className="min-w-0 text-sm font-bold break-all">
                        {disposition.quarantine_disposition_id}
                      </p>
                      <StatusBadge tone="neutral">
                        {disposition.disposition_type_code} ·{" "}
                        {disposition.status_code}
                      </StatusBadge>
                    </div>
                    <dl className="grid gap-3 sm:grid-cols-2">
                      <Detail
                        label="Quantity"
                        value={`${disposition.disposition_qty} ${value.base_uom_code || ""}`}
                      />
                      <Detail
                        label="Decision time"
                        value={disposition.decided_at}
                      />
                      <Detail
                        label="Decision reference"
                        value={disposition.client_decision_reference}
                      />
                      <Detail
                        label="Decision notes"
                        value={disposition.decision_notes}
                      />
                      <Detail
                        label="Target location"
                        value={disposition.target_location_code}
                      />
                      <Detail
                        label="Inventory movement"
                        value={disposition.inventory_movement_id}
                      />
                    </dl>
                    {disposition.rework_task ? (
                      <div className="space-y-2 rounded-lg bg-cyan-50 p-3 text-sm text-cyan-950">
                        <p className="font-semibold break-all">
                          Rework: {disposition.rework_task.rework_task_id} ·{" "}
                          {quarantineLabel(
                            disposition.rework_task.task_status_code,
                          )}
                        </p>
                        <Link
                          href={reworkHref({
                            ...value,
                            rework_task_id:
                              disposition.rework_task.rework_task_id,
                          })}
                          className="inline-block font-semibold text-cyan-800 underline underline-offset-4"
                        >
                          Open rework task
                        </Link>
                        <p>
                          Quantity {disposition.rework_task.planned_qty} ·
                          completed {disposition.rework_task.completed_qty}
                        </p>
                        <p className="whitespace-pre-wrap">
                          {disposition.rework_task.work_instructions}
                        </p>
                        {disposition.rework_task.result_notes ? (
                          <p>{disposition.rework_task.result_notes}</p>
                        ) : null}
                        {disposition.rework_task.reinspection_id ? (
                          <Link
                            className="inline-block font-semibold underline"
                            href={inspectionHref(
                              value,
                              disposition.rework_task.reinspection_id,
                            )}
                          >
                            Open reinspection
                          </Link>
                        ) : (
                          <p className="text-xs">
                            Complete the rework task before a new quality
                            inspection is created.
                          </p>
                        )}
                      </div>
                    ) : null}
                  </article>
                ))
              )}
            </section>
          </>
        ) : null}
      </div>
    </OperationDialog>
  );
}
