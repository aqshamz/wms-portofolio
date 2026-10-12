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
  reconcileCycleCountSerials,
  recordCycleCount,
  rejectCycleCount,
  stockControlKeys,
} from "./stock-control-api";
import type { CycleCountLine } from "./stock-control-types";
import { FoundStockForm } from "./found-stock-form";

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
  const [showFoundStock, setShowFoundStock] = useState(false);
  const [missingSerials, setMissingSerials] = useState<
    Record<string, string[]>
  >({});
  const [foundSerials, setFoundSerials] = useState<Record<string, string>>({});
  const count = useQuery({
    queryKey: stockControlKeys.cycleCount(cycleCountId),
    queryFn: () => getCycleCount(cycleCountId),
  });
  const row = count.data;
  const documentLabel =
    row?.count_type_code === "GRAND" ? "Grand stock opname" : "Cycle count";
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
  const serialReviewable = useMemo(
    () =>
      reviewable.filter(
        (line) =>
          line.serial_controlled &&
          Number(line.variance_quantity ?? 0) !== 0 &&
          (!line.found_stock ||
            (line.created_by !== user?.account_id &&
              line.counted_by !== user?.account_id)),
      ),
    [reviewable, user?.account_id],
  );
  const selectedHasSerialVariance = selected.some((lineId) =>
    serialReviewable.some((line) => line.cycle_count_line_id === lineId),
  );
  const isCreator = row?.created_by === user?.account_id;
  const canCount = can(PERMISSIONS.INVENTORY.COUNT);
  const hasApprovalPermission = can(PERMISSIONS.INVENTORY.COUNT_APPROVE);
  const canApprove =
    hasApprovalPermission &&
    !isCreator &&
    (row?.status_code === "REVIEW" || row?.status_code === "PARTIALLY_POSTED");
  const update = useMutation({
    mutationFn: async (action: "count" | "approve" | "reject" | "cancel") => {
      if (!row) throw new Error(`${documentLabel} is unavailable.`);
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
      toast.success(`${documentLabel} updated.`);
    },
    onError: (error) => toast.error(error.message),
  });
  const reconcile = useMutation({
    mutationFn: async (line: CycleCountLine) => {
      if (!row) throw new Error(`${documentLabel} is unavailable.`);
      const variance = Number(line.variance_quantity ?? 0);
      const missing = missingSerials[line.cycle_count_line_id] ?? [];
      const found = (foundSerials[line.cycle_count_line_id] ?? "")
        .split(/\r?\n/)
        .map((value) => value.trim())
        .filter(Boolean);
      return reconcileCycleCountSerials(row.cycle_count_id, row.version_no, {
        line_id: line.cycle_count_line_id,
        missing_serial_ids: variance < 0 ? missing : [],
        found_serial_numbers: variance > 0 ? found : [],
      });
    },
    onSuccess: (updated, line) => {
      client.setQueryData(stockControlKeys.cycleCount(cycleCountId), updated);
      void client.invalidateQueries({
        queryKey: stockControlKeys.cycleCounts(),
      });
      setSelected((current) =>
        current.filter((id) => id !== line.cycle_count_line_id),
      );
      setMissingSerials((current) => {
        const next = { ...current };
        delete next[line.cycle_count_line_id];
        return next;
      });
      setFoundSerials((current) => {
        const next = { ...current };
        delete next[line.cycle_count_line_id];
        return next;
      });
      toast.success("Serial variance reconciled.");
    },
    onError: (error) => toast.error(error.message),
  });
  return (
    <OperationDialog
      title={cycleCountId}
      description={`${documentLabel}: blind count, recount, variance review and posting audit.`}
      busy={update.isPending || reconcile.isPending}
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
            <Info label="Count type" value={documentLabel} />
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
          {row.count_type_code === "GRAND" ? (
            row.stock_frozen ? (
              <section
                role="status"
                className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950"
              >
                <p className="font-semibold">
                  Stock frozen for this grand opname
                </p>
                <p className="mt-2">
                  All stock movements and reservation changes for{" "}
                  {row.owner_name} in {row.warehouse_name} are blocked,
                  including receiving, QC, quarantine and dock stock outside the
                  count. Only approved corrections from this opname can change
                  stock.
                </p>
                <p className="mt-2">
                  Partial posting keeps the freeze active. It is released
                  automatically when every line is finalized or the remaining
                  count is cancelled. Other customers and warehouses are
                  unaffected.
                </p>
              </section>
            ) : (
              <section
                role="status"
                className="rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-950"
              >
                <p className="font-semibold">
                  This grand opname’s freeze is released
                </p>
                <p className="mt-2">
                  {row.cancelled_at
                    ? "The remaining count was cancelled. Any corrections already posted remain in inventory."
                    : "All count lines are finalized."}
                </p>
              </section>
            )
          ) : null}
          {row.count_type_code === "GRAND" &&
          row.total_lines === 0 &&
          !row.completed_at &&
          !row.cancelled_at ? (
            <p className="rounded-xl bg-cyan-50 p-4 text-sm text-cyan-950">
              No positive storage balances were registered when this opname
              started. Record any stock found during the count. If nothing is
              found, cancel the empty document with an explanatory reason.
            </p>
          ) : null}
          {canCount && !row.completed_at && !row.cancelled_at ? (
            showFoundStock ? (
              <FoundStockForm
                count={row}
                onCancel={() => setShowFoundStock(false)}
                onAdded={(updated) => {
                  client.setQueryData(
                    stockControlKeys.cycleCount(cycleCountId),
                    updated,
                  );
                  void client.invalidateQueries({
                    queryKey: stockControlKeys.cycleCounts(),
                  });
                  setShowFoundStock(false);
                  setSelected([]);
                }}
              />
            ) : (
              <div className="flex justify-end">
                <Button
                  variant="secondary"
                  onClick={() => setShowFoundStock(true)}
                >
                  Record found stock
                </Button>
              </div>
            )
          ) : null}
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
                const isReviewable =
                  (!line.found_stock ||
                    (line.created_by !== user?.account_id &&
                      line.counted_by !== user?.account_id)) &&
                  reviewable.some(
                    (entry) =>
                      entry.cycle_count_line_id === line.cycle_count_line_id,
                  );
                const isSerialReviewable = serialReviewable.some(
                  (entry) =>
                    entry.cycle_count_line_id === line.cycle_count_line_id,
                );
                const serialUnits = Math.abs(
                  Number(line.variance_quantity ?? 0),
                );
                const selectedMissing =
                  missingSerials[line.cycle_count_line_id] ?? [];
                const enteredFound = (
                  foundSerials[line.cycle_count_line_id] ?? ""
                )
                  .split(/\r?\n/)
                  .map((value) => value.trim())
                  .filter(Boolean);
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
                        {line.found_stock ? (
                          <StatusBadge tone="warning">Found stock</StatusBadge>
                        ) : null}
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
                      {line.found_stock ? (
                        <div className="mt-2 text-xs text-cyan-900">
                          <p>
                            Found stock · recorded by{" "}
                            {line.created_by_display_name} · system quantity at
                            discovery: 0
                          </p>
                          <p className="mt-1">{line.discovery_notes}</p>
                          {line.created_by === user?.account_id ||
                          line.counted_by === user?.account_id ? (
                            <p className="mt-1">
                              A different account must review the stock you
                              recorded or counted.
                            </p>
                          ) : null}
                        </div>
                      ) : null}
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
                          Reconcile the exact serial identities below before
                          this line can post.
                        </p>
                      ) : null}
                      {line.handling_unit_barcode ? (
                        <p className="mt-1 text-xs text-slate-600">
                          Handling unit: {line.handling_unit_barcode}
                        </p>
                      ) : null}
                      {line.inventory_movement_id ? (
                        <p className="mt-2 text-xs text-slate-600">
                          Inventory movement: {line.inventory_movement_id}
                        </p>
                      ) : null}
                      {line.serial_reconciliations.length ? (
                        <div className="mt-3 rounded-lg bg-emerald-50 p-3 text-xs text-emerald-950">
                          <p className="font-semibold">Serial audit</p>
                          {line.serial_reconciliations.map((entry) => (
                            <p
                              key={entry.cycle_count_serial_reconciliation_id}
                              className="mt-1"
                            >
                              {entry.direction_code}: {entry.serial_no} ·{" "}
                              {entry.inventory_movement_id} ·{" "}
                              {entry.created_by_display_name}
                            </p>
                          ))}
                        </div>
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
                    {canApprove && isSerialReviewable ? (
                      <div className="space-y-3 rounded-xl border border-cyan-200 bg-cyan-50 p-4 sm:col-span-3">
                        <div>
                          <p className="font-semibold text-cyan-950">
                            Serial reconciliation · {serialUnits} unit
                            {serialUnits === 1 ? "" : "s"}
                          </p>
                          <p className="mt-1 text-xs text-cyan-900">
                            {Number(line.variance_quantity) < 0
                              ? "Select the serials missing from this balance. Each selected identity will be removed with its own movement."
                              : "Enter each newly found serial number on a separate line. Each identity will be added with its own movement."}
                          </p>
                        </div>
                        {Number(line.variance_quantity) < 0 ? (
                          <div className="grid gap-2 sm:grid-cols-2">
                            {line.current_serials.map((serial) => (
                              <label
                                key={serial.serial_id}
                                className="flex items-center gap-2 rounded-lg border border-cyan-200 bg-white p-3"
                              >
                                <input
                                  type="checkbox"
                                  checked={selectedMissing.includes(
                                    serial.serial_id,
                                  )}
                                  onChange={(event) =>
                                    setMissingSerials((current) => ({
                                      ...current,
                                      [line.cycle_count_line_id]: event.target
                                        .checked
                                        ? [
                                            ...(current[
                                              line.cycle_count_line_id
                                            ] ?? []),
                                            serial.serial_id,
                                          ]
                                        : (
                                            current[line.cycle_count_line_id] ??
                                            []
                                          ).filter(
                                            (id) => id !== serial.serial_id,
                                          ),
                                    }))
                                  }
                                />
                                <span className="font-medium">
                                  {serial.serial_no}
                                </span>
                              </label>
                            ))}
                          </div>
                        ) : (
                          <Textarea
                            aria-label={`Found serial numbers for line ${line.line_no}`}
                            placeholder="One serial number per line"
                            value={foundSerials[line.cycle_count_line_id] ?? ""}
                            onChange={(event) =>
                              setFoundSerials((current) => ({
                                ...current,
                                [line.cycle_count_line_id]: event.target.value,
                              }))
                            }
                          />
                        )}
                        <div className="flex items-center justify-between gap-3">
                          <p className="text-xs text-cyan-900">
                            {Number(line.variance_quantity) < 0
                              ? `${selectedMissing.length} of ${serialUnits} selected`
                              : `${enteredFound.length} of ${serialUnits} entered`}
                          </p>
                          <Button
                            disabled={
                              reconcile.isPending ||
                              (Number(line.variance_quantity) < 0
                                ? selectedMissing.length !== serialUnits
                                : enteredFound.length !== serialUnits)
                            }
                            onClick={() => reconcile.mutate(line)}
                          >
                            Reconcile serials
                          </Button>
                        </div>
                      </div>
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
              {row.status_code === "COUNTING" || row.status_code === "DRAFT"
                ? "Finish the remaining counts and recounts before approval."
                : isCreator
                  ? `You created this cycle count. A different account with ${PERMISSIONS.INVENTORY.COUNT_APPROVE} must approve or reject it.`
                  : `Your account does not currently have ${PERMISSIONS.INVENTORY.COUNT_APPROVE}. Refresh your session after the role permission is granted.`}
            </section>
          ) : null}
          {canApprove && reviewable.length ? (
            <section className="space-y-3 rounded-xl border border-amber-200 bg-amber-50 p-4">
              <p className="font-semibold">
                Review selected lines ({selected.length})
              </p>
              <p className="text-xs text-amber-900">
                Serial variance lines can be rejected here, but must use their
                line-level reconciliation control before approval.
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
                  disabled={!selected.length || selectedHasSerialVariance}
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
