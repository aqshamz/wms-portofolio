"use client";

import { useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Decimal from "decimal.js";
import {
  CircleAlert,
  LoaderCircle,
  Pencil,
  RotateCcw,
  ShieldCheck,
  X,
  XCircle,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { StatusBadge } from "@/components/ui/status-badge";
import { inboundOrderKeys } from "@/features/inbound-orders/inbound-order-api";
import {
  cancelReceipt,
  completeReceipt,
  getInventoryBalance,
  getReceipt,
  receiptKeys,
  reverseReceipt,
} from "@/features/receipts/receipt-api";
import { ReceiptFormDialog } from "@/features/receipts/receipt-form-dialog";
import type {
  ReceiptLine,
  ReceiptStatus,
} from "@/features/receipts/receipt-types";

function statusTone(status: ReceiptStatus) {
  if (status === "COMPLETED") return "success";
  if (status === "OPEN") return "info";
  if (status === "CANCELLED") return "danger";
  return "neutral";
}

function label(value: string) {
  return value
    .replaceAll("_", " ")
    .toLowerCase()
    .replace(/^./, (first) => first.toUpperCase());
}

function acceptedBaseQuantity(line: ReceiptLine) {
  try {
    return new Decimal(line.received_base_qty)
      .minus(line.rejected_base_qty)
      .toFixed(6);
  } catch {
    return "—";
  }
}

function Detail({ name, value }: { name: string; value?: string | null }) {
  return (
    <div>
      <dt className="text-xs font-semibold tracking-wide text-slate-500 uppercase">
        {name}
      </dt>
      <dd className="mt-1 text-sm font-medium text-slate-900">
        {value || "—"}
      </dd>
    </div>
  );
}

export function ReceiptDetailDialog({
  receiptId,
  canReceive,
  canCancel,
  canReadInventory,
  onOpenChange,
}: {
  receiptId: string;
  canReceive: boolean;
  canCancel: boolean;
  canReadInventory: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [editing, setEditing] = useState(false);
  const [replacing, setReplacing] = useState(false);
  const [action, setAction] = useState<"complete" | "cancel" | "reverse">();
  const [reason, setReason] = useState("");
  const [businessDate, setBusinessDate] = useState(
    new Date().toISOString().slice(0, 10),
  );
  const receipt = useQuery({
    queryKey: receiptKeys.detail(receiptId),
    queryFn: () => getReceipt(receiptId),
  });
  const refresh = async (id: string) => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: receiptKeys.lists() }),
      queryClient.invalidateQueries({ queryKey: receiptKeys.detail(id) }),
      queryClient.invalidateQueries({ queryKey: inboundOrderKeys.all }),
    ]);
  };
  const transition = useMutation({
    mutationFn: async () => {
      const current = receipt.data;
      if (!current || !action) throw new Error("Receipt is not ready.");
      if (action === "complete") {
        return completeReceipt(current.receipt_id, current.version_no);
      }
      if (!reason.trim()) throw new Error("A reason is required.");
      if (action === "cancel") {
        return cancelReceipt(
          current.receipt_id,
          current.version_no,
          reason.trim(),
        );
      }
      const balanceIds = Array.from(
        new Set(
          (current.lines ?? []).flatMap((line) =>
            line.batches.flatMap((batch) =>
              batch.initial_balance_id ? [batch.initial_balance_id] : [],
            ),
          ),
        ),
      );
      if (!balanceIds.length) {
        throw new Error(
          "This completed receipt has no posted balances to reverse.",
        );
      }
      const balances = await Promise.all(balanceIds.map(getInventoryBalance));
      return reverseReceipt(current.receipt_id, {
        expected_version: current.version_no,
        business_date: businessDate,
        reason: reason.trim(),
        balances: balances.map((balance) => ({
          balance_id: balance.balance_id,
          expected_version: balance.version_no,
        })),
      });
    },
    onSuccess: async (updated) => {
      await refresh(updated.receipt_id);
      toast.success(
        action === "complete"
          ? "Receipt completed and inventory posted to QC pending."
          : action === "cancel"
            ? "Open receipt cancelled."
            : "Receipt and untouched inventory reversed.",
      );
      setAction(undefined);
      setReason("");
    },
  });

  if ((editing || replacing) && receipt.data) {
    return (
      <ReceiptFormDialog
        ownerId={receipt.data.owner_id}
        warehouseId={receipt.data.warehouse_id}
        scopeLabel={`${receipt.data.owner_code} · ${receipt.data.warehouse_code}`}
        receipt={editing ? receipt.data : undefined}
        replacementFor={replacing ? receipt.data : undefined}
        onOpenChange={(open) => {
          if (!open) {
            setEditing(false);
            setReplacing(false);
          }
        }}
      />
    );
  }

  return (
    <Dialog.Root
      open
      onOpenChange={(open) => !transition.isPending && onOpenChange(open)}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[94vh] w-[calc(100%-1.5rem)] max-w-5xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                Receipt details
              </Dialog.Title>
              <Dialog.Description className="mt-1 font-mono text-xs text-slate-500">
                {receiptId}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close receipt details"
                className="grid size-10 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>
          {receipt.isPending ? (
            <div className="flex min-h-64 items-center justify-center gap-2 text-sm text-slate-600">
              <LoaderCircle className="size-5 animate-spin" /> Loading receipt…
            </div>
          ) : receipt.error ? (
            <p
              role="alert"
              className="m-6 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
            >
              {receipt.error.message}
            </p>
          ) : receipt.data ? (
            <div className="p-5 sm:p-6">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <div className="flex flex-wrap items-center gap-3">
                    <h2 className="text-xl font-bold text-slate-950">
                      {receipt.data.delivery_note_no || receipt.data.receipt_id}
                    </h2>
                    <StatusBadge tone={statusTone(receipt.data.status_code)}>
                      {label(receipt.data.status_code)}
                    </StatusBadge>
                  </div>
                  <p className="mt-1 text-sm text-slate-600">
                    {receipt.data.owner_code} · {receipt.data.warehouse_code}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  {canReceive && receipt.data.status_code === "OPEN" ? (
                    <Button
                      variant="secondary"
                      onClick={() => setEditing(true)}
                    >
                      <Pencil className="size-4" /> Edit draft
                    </Button>
                  ) : null}
                  {canReceive &&
                  receipt.data.status_code === "CANCELLED" &&
                  !receipt.data.successor_receipt_id ? (
                    <Button onClick={() => setReplacing(true)}>
                      <RotateCcw className="size-4" /> Create replacement
                    </Button>
                  ) : null}
                  {canReceive && receipt.data.status_code === "OPEN" ? (
                    <Button onClick={() => setAction("complete")}>
                      <ShieldCheck className="size-4" /> Complete
                    </Button>
                  ) : null}
                  {canCancel && receipt.data.status_code === "OPEN" ? (
                    <Button
                      variant="secondary"
                      className="text-rose-700"
                      onClick={() => setAction("cancel")}
                    >
                      <XCircle className="size-4" /> Cancel
                    </Button>
                  ) : null}
                  {canCancel &&
                  canReadInventory &&
                  receipt.data.status_code === "COMPLETED" ? (
                    <Button
                      variant="secondary"
                      className="text-amber-800"
                      onClick={() => setAction("reverse")}
                    >
                      <RotateCcw className="size-4" /> Reverse
                    </Button>
                  ) : null}
                </div>
              </div>
              <dl className="mt-6 grid gap-5 rounded-2xl bg-slate-50 p-4 sm:grid-cols-2 lg:grid-cols-4">
                <Detail name="Inbound Order" value={receipt.data.inbound_id} />
                <Detail
                  name="Business date"
                  value={receipt.data.business_date}
                />
                <Detail
                  name="Received at"
                  value={new Date(receipt.data.received_at).toLocaleString(
                    "en-ID",
                  )}
                />
                <Detail
                  name="Dock location ID"
                  value={receipt.data.dock_location_id}
                />
                <Detail name="Vehicle" value={receipt.data.vehicle_number} />
                <Detail name="Seal" value={receipt.data.seal_number} />
                <Detail
                  name="Delivery note"
                  value={receipt.data.delivery_note_no}
                />
                <Detail
                  name="Version"
                  value={String(receipt.data.version_no)}
                />
              </dl>
              {receipt.data.notes ? (
                <p className="mt-4 rounded-xl border border-slate-200 p-4 text-sm text-slate-700">
                  {receipt.data.notes}
                </p>
              ) : null}
              {canCancel &&
              !canReadInventory &&
              receipt.data.status_code === "COMPLETED" ? (
                <p className="mt-4 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
                  INVENTORY.READ is also required to load current balance
                  versions for a safe reversal.
                </p>
              ) : null}
              <section className="mt-7">
                <h3 className="text-sm font-bold text-slate-950">
                  Receipt lines
                </h3>
                <div className="mt-3 space-y-3">
                  {(receipt.data.lines ?? []).map((line) => (
                    <article
                      key={line.receipt_line_id}
                      className="rounded-2xl border border-slate-200 p-4"
                    >
                      <div className="flex flex-wrap items-start justify-between gap-3">
                        <div>
                          <p className="font-semibold text-slate-950">
                            {line.item_name}
                          </p>
                          <p className="font-mono text-xs text-slate-500">
                            {line.item_code} · {line.uom_code}
                          </p>
                        </div>
                        <p className="text-sm text-slate-700">
                          Received <strong>{line.received_qty}</strong>{" "}
                          {line.uom_code} · Rejected{" "}
                          <strong>{line.rejected_qty}</strong> {line.uom_code} ·
                          Accepted <strong>{line.accepted_qty}</strong>{" "}
                          {line.uom_code}
                        </p>
                      </div>
                      <p className="mt-2 text-xs text-slate-500">
                        Base quantity: received {line.received_base_qty}{" "}
                        {line.base_uom_code} · rejected {line.rejected_base_qty}{" "}
                        {line.base_uom_code} · accepted{" "}
                        {acceptedBaseQuantity(line)} {line.base_uom_code}
                      </p>
                      {line.exception_notes ? (
                        <p className="mt-3 rounded-lg bg-rose-50 p-3 text-sm text-rose-900">
                          <strong>
                            {label(line.exception_type_code || "exception")}:
                          </strong>{" "}
                          {line.exception_notes}
                        </p>
                      ) : null}
                      <div className="mt-3 overflow-x-auto">
                        <table className="min-w-full text-left text-sm">
                          <thead className="text-xs text-slate-500 uppercase">
                            <tr>
                              <th className="px-2 py-2">Quantity</th>
                              <th className="px-2 py-2">Location</th>
                              <th className="px-2 py-2">Lot</th>
                              <th className="px-2 py-2">Serial</th>
                              <th className="px-2 py-2">Inventory</th>
                            </tr>
                          </thead>
                          <tbody>
                            {line.batches.map((batch) => (
                              <tr
                                key={batch.receipt_inventory_id}
                                className="border-t border-slate-100"
                              >
                                <td className="px-2 py-2">
                                  {batch.source_qty} {batch.source_uom_code}
                                  {batch.source_uom_id !== batch.base_uom_id ? (
                                    <p className="mt-1 text-xs text-slate-500">
                                      {batch.base_qty} {batch.base_uom_code}
                                    </p>
                                  ) : null}
                                </td>
                                <td className="px-2 py-2">
                                  {batch.received_location_code}
                                </td>
                                <td className="px-2 py-2">
                                  {batch.lot_number || "—"}
                                </td>
                                <td className="px-2 py-2">
                                  {batch.serial_no || "—"}
                                </td>
                                <td className="px-2 py-2">
                                  {batch.initial_inventory_status_code}
                                  {batch.initial_balance_id
                                    ? " · posted"
                                    : " · draft"}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </article>
                  ))}
                </div>
              </section>
              {action ? (
                <section className="mt-6 rounded-2xl border border-amber-200 bg-amber-50 p-4">
                  <div className="flex items-start gap-3">
                    <CircleAlert className="mt-0.5 size-5 shrink-0 text-amber-700" />
                    <div>
                      <h3 className="font-bold text-amber-950">
                        {action === "complete"
                          ? "Complete this receipt?"
                          : action === "cancel"
                            ? "Cancel this open receipt?"
                            : "Reverse this completed receipt?"}
                      </h3>
                      <p className="mt-1 text-sm text-amber-900">
                        {action === "complete"
                          ? "Accepted batches will post to inventory as QC_PENDING and the Inbound Order progress will update."
                          : action === "cancel"
                            ? "The unposted draft will be cancelled. It can no longer be edited or completed."
                            : "Reversal only succeeds while every posted balance is untouched and quality inspection has not started."}
                      </p>
                    </div>
                  </div>
                  {action !== "complete" ? (
                    <textarea
                      rows={3}
                      value={reason}
                      onChange={(event) => setReason(event.target.value)}
                      placeholder="Reason (required)"
                      className="mt-4 w-full rounded-xl border border-amber-300 bg-white px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                    />
                  ) : null}
                  {action === "reverse" ? (
                    <div className="mt-3 max-w-xs">
                      <label
                        htmlFor="receipt-reversal-date"
                        className="text-sm font-semibold text-amber-950"
                      >
                        Reversal business date
                      </label>
                      <Input
                        id="receipt-reversal-date"
                        type="date"
                        value={businessDate}
                        onChange={(event) =>
                          setBusinessDate(event.target.value)
                        }
                      />
                    </div>
                  ) : null}
                  {transition.error ? (
                    <p
                      role="alert"
                      className="mt-3 text-sm font-medium text-rose-800"
                    >
                      {transition.error.message}
                    </p>
                  ) : null}
                  <div className="mt-4 flex flex-wrap justify-end gap-2">
                    <Button
                      type="button"
                      variant="secondary"
                      disabled={transition.isPending}
                      onClick={() => {
                        setAction(undefined);
                        setReason("");
                      }}
                    >
                      Keep receipt
                    </Button>
                    <Button
                      type="button"
                      disabled={
                        transition.isPending ||
                        (action !== "complete" && !reason.trim())
                      }
                      className={
                        action === "complete"
                          ? undefined
                          : "bg-rose-700 hover:bg-rose-800"
                      }
                      onClick={() => transition.mutate()}
                    >
                      {transition.isPending ? (
                        <LoaderCircle className="size-4 animate-spin" />
                      ) : null}
                      {transition.isPending
                        ? "Working…"
                        : action === "complete"
                          ? "Complete and post"
                          : action === "cancel"
                            ? "Cancel receipt"
                            : "Reverse inventory"}
                    </Button>
                  </div>
                </section>
              ) : null}
            </div>
          ) : null}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
