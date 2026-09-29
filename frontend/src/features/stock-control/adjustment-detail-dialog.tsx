"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/components/auth/auth-provider";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { Textarea } from "@/components/ui/textarea";
import { PERMISSIONS } from "@/lib/auth/permissions";
import {
  approveAdjustment,
  cancelAdjustment,
  getAdjustment,
  rejectAdjustment,
  stockControlKeys,
} from "./stock-control-api";

type Action = "approve" | "reject" | "cancel" | null;
function tone(status: string) {
  if (status === "POSTED") return "success" as const;
  if (status === "CANCELLED" || status === "REJECTED") return "danger" as const;
  return "warning" as const;
}

export function AdjustmentDetailDialog({
  adjustmentId,
  onOpenChange,
}: {
  adjustmentId: string;
  onOpenChange: (open: boolean) => void;
}) {
  const { user, can } = useAuth();
  const queryClient = useQueryClient();
  const [action, setAction] = useState<Action>(null);
  const [reason, setReason] = useState("");
  const [selected, setSelected] = useState<string[]>([]);
  const adjustment = useQuery({
    queryKey: stockControlKeys.adjustment(adjustmentId),
    queryFn: () => getAdjustment(adjustmentId),
  });
  const row = adjustment.data;
  const pendingLines = useMemo(
    () => row?.lines.filter((line) => line.decision_code === "PENDING") ?? [],
    [row?.lines],
  );
  const isCreator = row?.created_by === user?.account_id;
  const canApprove =
    Boolean(row && !isCreator) && can(PERMISSIONS.INVENTORY.ADJUST_APPROVE);
  const selectedLines = pendingLines.filter((line) =>
    selected.includes(line.inventory_adjustment_line_id),
  );
  const selectedStale = selectedLines.some(
    (line) =>
      line.current_balance_version_no !== line.planned_balance_version_no,
  );
  const transition = useMutation({
    mutationFn: async () => {
      if (!row) throw new Error("Adjustment is unavailable.");
      if (action === "approve")
        return approveAdjustment(
          row.inventory_adjustment_id,
          row.version_no,
          selected,
        );
      if (action === "reject")
        return rejectAdjustment(
          row.inventory_adjustment_id,
          row.version_no,
          selected,
          reason.trim(),
        );
      return cancelAdjustment(
        row.inventory_adjustment_id,
        row.version_no,
        reason.trim(),
      );
    },
    onSuccess: (updated) => {
      toast.success(
        action === "approve"
          ? `${selected.length} selected line(s) posted.`
          : action === "reject"
            ? `${selected.length} selected line(s) rejected.`
            : "Remaining pending lines cancelled.",
      );
      queryClient.setQueryData(
        stockControlKeys.adjustment(adjustmentId),
        updated,
      );
      void queryClient.invalidateQueries({
        queryKey: stockControlKeys.adjustments(),
      });
      setAction(null);
      setReason("");
      setSelected([]);
    },
    onError: (error) => toast.error(error.message),
  });
  const error = adjustment.error ?? transition.error;

  return (
    <OperationDialog
      title={adjustmentId}
      description="Adjustment document, line decisions and inventory posting audit."
      busy={transition.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close adjustment details"
    >
      {adjustment.isPending ? (
        <div className="grid min-h-56 place-items-center">
          <LoaderCircle className="size-6 animate-spin" />
        </div>
      ) : !row ? (
        <p className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
          {error?.message ?? "Adjustment could not be loaded."}
        </p>
      ) : (
        <div className="space-y-5">
          <section className="grid gap-4 rounded-xl bg-slate-50 p-4 text-sm sm:grid-cols-3">
            <Info label="Status" value={row.status_code} />
            <Info label="Direction" value={row.direction} />
            <Info label="Business date" value={row.business_date} />
            <Info label="Owner" value={row.owner_name} />
            <Info label="Warehouse" value={row.warehouse_name} />
            <Info label="Reason" value={row.reason_name} />
            <Info label="Requested by" value={row.created_by_display_name} />
            <Info
              label="Progress"
              value={`${row.posted_lines} posted · ${row.pending_lines} pending · ${row.rejected_lines} rejected · ${row.cancelled_lines} cancelled`}
            />
            <Info label="Notes" value={row.notes ?? "—"} />
          </section>

          <section className="overflow-hidden rounded-xl border border-slate-200">
            <div className="flex items-center justify-between border-b border-slate-200 bg-slate-50 px-4 py-3">
              <p className="font-semibold">Lines ({row.total_lines})</p>
              {canApprove && pendingLines.length ? (
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={
                      selected.length === pendingLines.length &&
                      pendingLines.length > 0
                    }
                    onChange={(event) =>
                      setSelected(
                        event.target.checked
                          ? pendingLines.map(
                              (line) => line.inventory_adjustment_line_id,
                            )
                          : [],
                      )
                    }
                  />
                  Select all pending
                </label>
              ) : null}
            </div>
            <div className="divide-y divide-slate-200">
              {row.lines.map((line) => {
                const stale =
                  line.decision_code === "PENDING" &&
                  line.current_balance_version_no !==
                    line.planned_balance_version_no;
                return (
                  <div
                    key={line.inventory_adjustment_line_id}
                    className="flex items-start gap-3 p-4 text-sm"
                  >
                    {canApprove && line.decision_code === "PENDING" ? (
                      <input
                        className="mt-1"
                        type="checkbox"
                        aria-label={`Select line ${line.line_no}`}
                        checked={selected.includes(
                          line.inventory_adjustment_line_id,
                        )}
                        onChange={(event) =>
                          setSelected((current) =>
                            event.target.checked
                              ? [...current, line.inventory_adjustment_line_id]
                              : current.filter(
                                  (id) =>
                                    id !== line.inventory_adjustment_line_id,
                                ),
                          )
                        }
                      />
                    ) : null}
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <p className="font-semibold">
                          {line.line_no}. {line.item_code} · {line.quantity}{" "}
                          {line.uom_code}
                        </p>
                        <StatusBadge tone={tone(line.decision_code)}>
                          {line.decision_code}
                        </StatusBadge>
                      </div>
                      <p className="mt-1 text-slate-600">
                        {line.location_code} · {line.inventory_status_code}
                        {line.lot_number ? ` · Lot ${line.lot_number}` : ""}
                        {line.serial_number
                          ? ` · Serial ${line.serial_number}`
                          : ""}
                      </p>
                      <p className="mt-1 text-xs text-slate-500">
                        Captured version {line.planned_balance_version_no};
                        current {line.current_balance_version_no ?? "missing"}
                        {line.inventory_movement_id
                          ? ` · Movement ${line.inventory_movement_id}`
                          : ""}
                      </p>
                      {stale ? (
                        <p className="mt-2 flex items-center gap-1 text-xs font-semibold text-amber-800">
                          <AlertTriangle className="size-3" /> Balance changed;
                          this line cannot be approved.
                        </p>
                      ) : null}
                      {line.rejection_reason ? (
                        <p className="mt-2 text-xs text-rose-700">
                          Rejected by{" "}
                          {line.rejected_by_display_name ?? "reviewer"}:{" "}
                          {line.rejection_reason}
                        </p>
                      ) : null}
                    </div>
                  </div>
                );
              })}
            </div>
          </section>

          {error ? (
            <p
              role="alert"
              className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
            >
              {error.message}
            </p>
          ) : null}

          {action ? (
            <section className="rounded-xl border border-amber-200 bg-amber-50 p-4">
              <p className="font-semibold text-amber-950">
                {action === "approve"
                  ? `Post ${selected.length} selected line(s)?`
                  : action === "reject"
                    ? `Reject ${selected.length} selected line(s)?`
                    : `Cancel all ${row.pending_lines} pending line(s)?`}
              </p>
              <p className="mt-1 text-sm text-amber-900">
                {action === "approve"
                  ? "All selected lines post atomically. If one line is stale or invalid, none of them post."
                  : "Already posted lines are not reversed."}
              </p>
              {action !== "approve" ? (
                <div className="mt-4">
                  <FormField
                    label={
                      action === "reject"
                        ? "Rejection reason"
                        : "Cancellation reason"
                    }
                    htmlFor="adjustment-action-reason"
                    required
                  >
                    <Textarea
                      id="adjustment-action-reason"
                      maxLength={4000}
                      value={reason}
                      onChange={(event) => setReason(event.target.value)}
                    />
                  </FormField>
                </div>
              ) : null}
              <div className="mt-4 flex justify-end gap-3">
                <Button variant="secondary" onClick={() => setAction(null)}>
                  Back
                </Button>
                <Button
                  disabled={
                    transition.isPending ||
                    (action !== "cancel" && !selected.length) ||
                    (action !== "approve" && !reason.trim()) ||
                    (action === "approve" && selectedStale)
                  }
                  onClick={() => transition.mutate()}
                >
                  {transition.isPending ? (
                    <LoaderCircle className="size-4 animate-spin" />
                  ) : null}
                  Confirm
                </Button>
              </div>
            </section>
          ) : pendingLines.length ? (
            <div className="flex flex-wrap justify-end gap-3 border-t border-slate-200 pt-4">
              {isCreator ? (
                <Button variant="secondary" onClick={() => setAction("cancel")}>
                  Cancel remaining
                </Button>
              ) : null}
              {canApprove ? (
                <>
                  <Button
                    variant="secondary"
                    disabled={!selected.length}
                    onClick={() => setAction("reject")}
                  >
                    Reject selected
                  </Button>
                  <Button
                    disabled={!selected.length || selectedStale}
                    onClick={() => setAction("approve")}
                  >
                    Approve selected
                  </Button>
                </>
              ) : null}
            </div>
          ) : null}
        </div>
      )}
    </OperationDialog>
  );
}
function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs font-semibold text-slate-500 uppercase">{label}</p>
      <p className="mt-1 font-medium break-words text-slate-900">{value}</p>
    </div>
  );
}
