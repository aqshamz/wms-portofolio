"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/components/auth/auth-provider";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { Textarea } from "@/components/ui/textarea";
import { PERMISSIONS } from "@/lib/auth/permissions";
import {
  approveCycleCount,
  cancelCycleCount,
  getCycleCount,
  recordCycleCount,
  rejectCycleCount,
  stockControlKeys,
} from "./stock-control-api";

export function CycleCountDetailDialog({
  cycleCountId,
  onOpenChange,
}: {
  cycleCountId: string;
  onOpenChange: (open: boolean) => void;
}) {
  const { user, can } = useAuth();
  const client = useQueryClient();
  const [quantities, setQuantities] = useState<Record<string, string>>({});
  const [selected, setSelected] = useState<string[]>([]);
  const [reason, setReason] = useState("");
  const count = useQuery({
    queryKey: stockControlKeys.cycleCount(cycleCountId),
    queryFn: () => getCycleCount(cycleCountId),
  });
  const row = count.data;
  const countable = useMemo(
    () =>
      row?.lines.filter(
        (line) =>
          line.decision_code === "OPEN" ||
          (line.decision_code === "COUNTED" && line.requires_recount),
      ) ?? [],
    [row?.lines],
  );
  const reviewable = useMemo(
    () =>
      row?.lines.filter(
        (line) => line.decision_code === "COUNTED" && !line.requires_recount,
      ) ?? [],
    [row?.lines],
  );
  const isCreator = row?.created_by === user?.account_id;
  const canCount = can(PERMISSIONS.INVENTORY.COUNT);
  const hasApprovalPermission = can(PERMISSIONS.INVENTORY.COUNT_APPROVE);
  const canApprove = hasApprovalPermission && !isCreator;
  const update = useMutation({
    mutationFn: async (action: "count" | "approve" | "reject" | "cancel") => {
      if (!row) throw new Error("Cycle count is unavailable.");
      if (action === "count") {
        const lines = countable
          .filter((line) => quantities[line.cycle_count_line_id]?.trim())
          .map((line) => ({
            line_id: line.cycle_count_line_id,
            counted_quantity: quantities[line.cycle_count_line_id].trim(),
          }));
        if (!lines.length) throw new Error("Enter at least one count.");
        return recordCycleCount(row.cycle_count_id, row.version_no, lines);
      }
      if (action === "approve")
        return approveCycleCount(row.cycle_count_id, row.version_no, selected);
      if (action === "reject")
        return rejectCycleCount(
          row.cycle_count_id,
          row.version_no,
          selected,
          reason.trim(),
        );
      return cancelCycleCount(
        row.cycle_count_id,
        row.version_no,
        reason.trim(),
      );
    },
    onSuccess: (updated) => {
      client.setQueryData(stockControlKeys.cycleCount(cycleCountId), updated);
      void client.invalidateQueries({
        queryKey: stockControlKeys.cycleCounts(),
      });
      setSelected([]);
      setQuantities({});
      setReason("");
      toast.success("Cycle count updated.");
    },
    onError: (error) => toast.error(error.message),
  });
  return (
    <OperationDialog
      title={cycleCountId}
      description="Blind count, recount, variance review and posting audit."
      busy={update.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close cycle count details"
    >
      {count.isPending ? (
        <div className="grid min-h-56 place-items-center">
          <LoaderCircle className="size-6 animate-spin" />
        </div>
      ) : !row ? (
        <p className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
          {count.error?.message ?? "Cycle count could not be loaded."}
        </p>
      ) : (
        <div className="space-y-5">
          <section className="grid gap-4 rounded-xl bg-slate-50 p-4 text-sm sm:grid-cols-3">
            <Info label="Status" value={row.status_code} />
            <Info label="Business date" value={row.business_date} />
            <Info label="Tolerance" value={row.tolerance_quantity} />
            <Info label="Owner" value={row.owner_name} />
            <Info label="Warehouse" value={row.warehouse_name} />
            <Info label="Created by" value={row.created_by_display_name} />
            <Info
              label="Progress"
              value={`${row.final_lines} final · ${row.counted_lines} review · ${row.open_lines} open`}
            />
          </section>
          <section className="overflow-hidden rounded-xl border border-slate-200">
            <div className="border-b border-slate-200 bg-slate-50 px-4 py-3 font-semibold">
              Count lines ({row.total_lines})
            </div>
            <div className="divide-y divide-slate-200">
              {row.lines.map((line) => {
                const isCountable = countable.some(
                  (entry) =>
                    entry.cycle_count_line_id === line.cycle_count_line_id,
                );
                const isReviewable = reviewable.some(
                  (entry) =>
                    entry.cycle_count_line_id === line.cycle_count_line_id,
                );
                return (
                  <div
                    key={line.cycle_count_line_id}
                    className="grid gap-3 p-4 text-sm sm:grid-cols-[auto_1fr_12rem]"
                  >
                    {canApprove && isReviewable ? (
                      <input
                        className="mt-1"
                        type="checkbox"
                        checked={selected.includes(line.cycle_count_line_id)}
                        onChange={(event) =>
                          setSelected((current) =>
                            event.target.checked
                              ? [...current, line.cycle_count_line_id]
                              : current.filter(
                                  (id) => id !== line.cycle_count_line_id,
                                ),
                          )
                        }
                      />
                    ) : (
                      <span />
                    )}
                    <div>
                      <div className="flex flex-wrap items-center gap-2">
                        <p className="font-semibold">
                          {line.line_no}. {line.item_code} ·{" "}
                          {line.location_code}
                        </p>
                        <StatusBadge>{line.decision_code}</StatusBadge>
                        {line.requires_recount ? (
                          <StatusBadge tone="warning">
                            Recount required
                          </StatusBadge>
                        ) : null}
                      </div>
                      <p className="mt-1 text-slate-600">
                        {line.inventory_status_code} · {line.uom_code}
                        {line.lot_number ? ` · Lot ${line.lot_number}` : ""}
                      </p>
                      {line.system_quantity !== undefined ? (
                        <p className="mt-2 text-xs text-slate-600">
                          System {line.system_quantity}; counted{" "}
                          {line.counted_quantity ?? "—"}; variance{" "}
                          {line.variance_quantity ?? "—"}
                        </p>
                      ) : (
                        <p className="mt-2 text-xs font-semibold text-cyan-800">
                          Blind count — system quantity hidden
                        </p>
                      )}
                      {line.serial_controlled &&
                      line.variance_quantity &&
                      line.variance_quantity !== "0.000000" ? (
                        <p className="mt-2 text-xs text-amber-800">
                          Serial variance requires the dedicated serial
                          reconciliation workflow.
                        </p>
                      ) : null}
                    </div>
                    {canCount && isCountable ? (
                      <FormField
                        label={
                          line.requires_recount
                            ? "Recount quantity"
                            : "Physical quantity"
                        }
                        htmlFor={`count-${line.cycle_count_line_id}`}
                        required
                      >
                        <Input
                          id={`count-${line.cycle_count_line_id}`}
                          inputMode="decimal"
                          value={quantities[line.cycle_count_line_id] ?? ""}
                          onChange={(event) =>
                            setQuantities((current) => ({
                              ...current,
                              [line.cycle_count_line_id]: event.target.value,
                            }))
                          }
                        />
                      </FormField>
                    ) : null}
                  </div>
                );
              })}
            </div>
          </section>
          {canCount && countable.length ? (
            <div className="flex justify-end">
              <Button
                disabled={update.isPending}
                onClick={() => update.mutate("count")}
              >
                Record entered counts
              </Button>
            </div>
          ) : null}
          {reviewable.length && !canApprove ? (
            <section className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950">
              {isCreator
                ? `You created this cycle count. A different account with ${PERMISSIONS.INVENTORY.COUNT_APPROVE} must approve or reject it.`
                : `Your account does not currently have ${PERMISSIONS.INVENTORY.COUNT_APPROVE}. Refresh your session after the role permission is granted.`}
            </section>
          ) : null}
          {canApprove && reviewable.length ? (
            <section className="space-y-3 rounded-xl border border-amber-200 bg-amber-50 p-4">
              <p className="font-semibold">
                Review selected lines ({selected.length})
              </p>
              <Textarea
                aria-label="Cycle count rejection reason"
                placeholder="Reason required only when rejecting"
                value={reason}
                onChange={(event) => setReason(event.target.value)}
              />
              <div className="flex justify-end gap-3">
                <Button
                  variant="secondary"
                  disabled={!selected.length || !reason.trim()}
                  onClick={() => update.mutate("reject")}
                >
                  Reject selected
                </Button>
                <Button
                  disabled={!selected.length}
                  onClick={() => update.mutate("approve")}
                >
                  Approve selected
                </Button>
              </div>
            </section>
          ) : null}
          {isCreator &&
          row.status_code !== "POSTED" &&
          row.status_code !== "CANCELLED" ? (
            <section className="flex flex-col gap-3 border-t border-slate-200 pt-4 sm:flex-row sm:items-end">
              <div className="flex-1">
                <FormField
                  label="Cancellation reason"
                  htmlFor="cycle-cancel-reason"
                >
                  <Textarea
                    id="cycle-cancel-reason"
                    value={reason}
                    onChange={(event) => setReason(event.target.value)}
                  />
                </FormField>
              </div>
              <Button
                variant="secondary"
                disabled={!reason.trim()}
                onClick={() => update.mutate("cancel")}
              >
                Cancel remaining
              </Button>
            </section>
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
      <p className="mt-1 font-medium">{value}</p>
    </div>
  );
}
