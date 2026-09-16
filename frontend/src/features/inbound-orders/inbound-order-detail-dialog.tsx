"use client";

import { useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertTriangle,
  LoaderCircle,
  Pencil,
  Play,
  Plus,
  Trash2,
  X,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/ui/status-badge";
import {
  deleteInboundOrderLine,
  getInboundOrder,
  inboundOrderKeys,
  transitionInboundOrder,
} from "@/features/inbound-orders/inbound-order-api";
import {
  InboundOrderHeaderDialog,
  InboundOrderLineDialog,
} from "@/features/inbound-orders/inbound-order-draft-dialogs";
import type {
  InboundOrderLine,
  InboundOrderStatus,
} from "@/features/inbound-orders/inbound-order-types";
import { purchaseOrderKeys } from "@/features/purchase-orders/purchase-order-api";
import { ApiError } from "@/lib/api/client";

type Action = "release" | "close" | "cancel" | "delete-line";

const dateTimeFormatter = new Intl.DateTimeFormat("en-ID", {
  dateStyle: "medium",
  timeStyle: "short",
});

function statusTone(status: InboundOrderStatus) {
  if (status === "RELEASED" || status === "RECEIVED") return "success";
  if (status === "PARTIALLY_RECEIVED") return "warning";
  if (status === "CANCELLED") return "danger";
  if (status === "DRAFT") return "info";
  return "neutral";
}

function label(status: string) {
  return status
    .replaceAll("_", " ")
    .toLowerCase()
    .replace(/^./, (value) => value.toUpperCase());
}

export function InboundOrderDetailDialog({
  inboundOrderId,
  canPlan,
  canApprove,
  canCancel,
  onOpenChange,
}: {
  inboundOrderId: string;
  canPlan: boolean;
  canApprove: boolean;
  canCancel: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const [action, setAction] = useState<Action>();
  const [reason, setReason] = useState("");
  const [headerEditing, setHeaderEditing] = useState(false);
  const [lineEditing, setLineEditing] = useState<InboundOrderLine | null>();
  const [deleteLine, setDeleteLine] = useState<InboundOrderLine>();
  const detail = useQuery({
    queryKey: inboundOrderKeys.detail(inboundOrderId),
    queryFn: () => getInboundOrder(inboundOrderId),
  });
  const transition = useMutation({
    mutationFn: (nextAction: Action) => {
      if (!detail.data)
        throw new Error("Inbound Order details are not loaded.");
      if (nextAction === "delete-line") {
        if (!deleteLine) throw new Error("Select a line to remove.");
        return deleteInboundOrderLine(
          detail.data.inbound_id,
          deleteLine.inbound_line_id,
          detail.data.version_no,
        );
      }
      return transitionInboundOrder(
        detail.data.inbound_id,
        nextAction,
        detail.data.version_no,
        nextAction === "release" ? undefined : reason.trim(),
      );
    },
    onSuccess: async (updated, completedAction) => {
      queryClient.setQueryData(
        inboundOrderKeys.detail(updated.inbound_id),
        updated,
      );
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: inboundOrderKeys.lists() }),
        queryClient.invalidateQueries({ queryKey: purchaseOrderKeys.all }),
      ]);
      toast.success(
        completedAction === "release"
          ? "Inbound Order released."
          : completedAction === "close"
            ? "Inbound Order closed short."
            : completedAction === "cancel"
              ? "Inbound Order cancelled."
              : "Inbound Order line removed.",
      );
      setAction(undefined);
      setDeleteLine(undefined);
      setReason("");
    },
  });
  const order = detail.data;
  const canEditDraft = canPlan && order?.status_code === "DRAFT";
  const canRelease = canApprove && order?.status_code === "DRAFT";
  const canClose =
    canApprove &&
    (order?.status_code === "RELEASED" ||
      order?.status_code === "PARTIALLY_RECEIVED");
  const canCancelOrder =
    canCancel &&
    (order?.status_code === "DRAFT" ||
      order?.status_code === "RELEASED" ||
      order?.status_code === "PARTIALLY_RECEIVED");

  return (
    <>
      <Dialog.Root open onOpenChange={onOpenChange}>
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
          <Dialog.Content className="fixed inset-y-0 right-0 z-50 w-full max-w-3xl overflow-y-auto bg-white shadow-2xl focus:outline-none">
            <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
              <div>
                <Dialog.Title className="text-lg font-bold text-slate-950">
                  Inbound Order details
                </Dialog.Title>
                <Dialog.Description className="mt-1 text-sm text-slate-600">
                  Review the arrival plan, expected lines, and receiving
                  progress.
                </Dialog.Description>
              </div>
              <Dialog.Close asChild>
                <button
                  type="button"
                  aria-label="Close Inbound Order details"
                  className="grid size-10 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
                >
                  <X className="size-5" />
                </button>
              </Dialog.Close>
            </div>
            {detail.isPending ? (
              <div className="grid min-h-80 place-items-center text-sm text-slate-600">
                <div className="text-center">
                  <LoaderCircle className="mx-auto mb-3 size-7 animate-spin text-cyan-700" />
                  Loading Inbound Order…
                </div>
              </div>
            ) : detail.isError ? (
              <div className="p-5 sm:p-6">
                <div
                  role="alert"
                  className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
                >
                  {detail.error.message}
                </div>
                <Button
                  className="mt-4"
                  variant="secondary"
                  onClick={() => detail.refetch()}
                >
                  Try again
                </Button>
              </div>
            ) : order ? (
              <div className="p-5 sm:p-6">
                <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <div className="flex flex-wrap items-center gap-2">
                      <h2 className="font-mono text-xl font-bold text-slate-950 sm:text-2xl">
                        {order.inbound_id}
                      </h2>
                      <StatusBadge tone={statusTone(order.status_code)}>
                        {label(order.status_code)}
                      </StatusBadge>
                    </div>
                    <p className="mt-1 text-sm text-slate-500">
                      Version {order.version_no} · Created{" "}
                      {dateTimeFormatter.format(new Date(order.created_at))}
                    </p>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {canEditDraft ? (
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={() => setHeaderEditing(true)}
                      >
                        <Pencil className="size-4" />
                        Edit header
                      </Button>
                    ) : null}
                    {canRelease ? (
                      <Button size="sm" onClick={() => setAction("release")}>
                        <Play className="size-4" />
                        Release
                      </Button>
                    ) : null}
                    {canClose ? (
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={() => setAction("close")}
                      >
                        Close short
                      </Button>
                    ) : null}
                    {canCancelOrder ? (
                      <Button
                        size="sm"
                        variant="ghost"
                        className="text-rose-700"
                        onClick={() => setAction("cancel")}
                      >
                        Cancel
                      </Button>
                    ) : null}
                  </div>
                </div>
                <dl className="mt-6 grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2 lg:grid-cols-3">
                  <div>
                    <dt className="text-xs font-semibold text-slate-500">
                      Source Purchase Order
                    </dt>
                    <dd className="mt-1 font-mono text-xs text-slate-900">
                      {order.purchase_order_id}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-xs font-semibold text-slate-500">
                      Owner / warehouse
                    </dt>
                    <dd className="mt-1 font-medium text-slate-900">
                      {order.owner_code} · {order.warehouse_code}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-xs font-semibold text-slate-500">
                      Supplier
                    </dt>
                    <dd className="mt-1 font-medium text-slate-900">
                      {order.vendor_name}
                      <span className="block font-mono text-xs text-slate-500">
                        {order.vendor_code}
                      </span>
                    </dd>
                  </div>
                  <div>
                    <dt className="text-xs font-semibold text-slate-500">
                      Business date
                    </dt>
                    <dd className="mt-1 text-sm text-slate-800">
                      {order.business_date}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-xs font-semibold text-slate-500">
                      Expected arrival
                    </dt>
                    <dd className="mt-1 text-sm text-slate-800">
                      {order.expected_arrival_at
                        ? dateTimeFormatter.format(
                            new Date(order.expected_arrival_at),
                          )
                        : "Not set"}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-xs font-semibold text-slate-500">
                      References
                    </dt>
                    <dd className="mt-1 text-sm text-slate-800">
                      {order.external_reference ||
                        order.supplier_reference ||
                        "Not set"}
                    </dd>
                  </div>
                </dl>
                {order.notes ? (
                  <div className="mt-5 rounded-xl border border-slate-200 p-4">
                    <p className="text-xs font-semibold text-slate-500">
                      Notes
                    </p>
                    <p className="mt-2 text-sm whitespace-pre-wrap text-slate-700">
                      {order.notes}
                    </p>
                  </div>
                ) : null}
                <section className="mt-7">
                  <div className="flex items-end justify-between gap-3">
                    <div>
                      <h2 className="font-bold text-slate-950">
                        Expected lines
                      </h2>
                      <p className="mt-1 text-sm text-slate-500">
                        {order.lines?.length ?? 0} line
                        {order.lines?.length === 1 ? "" : "s"}
                      </p>
                    </div>
                    {canEditDraft ? (
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={() => setLineEditing(null)}
                      >
                        <Plus className="size-4" />
                        Add line
                      </Button>
                    ) : null}
                  </div>
                  <div className="mt-3 space-y-3">
                    {(order.lines ?? []).map((line) => (
                      <article
                        key={line.inbound_line_id}
                        className="rounded-xl border border-slate-200 p-4"
                      >
                        <div className="flex items-start justify-between gap-4">
                          <div>
                            <p className="font-semibold text-slate-950">
                              {line.item_name}
                            </p>
                            <p className="mt-0.5 font-mono text-xs text-slate-500">
                              Line {line.line_no} · {line.item_code}
                            </p>
                          </div>
                          <p className="text-right font-semibold text-slate-950">
                            {line.expected_qty}{" "}
                            <span className="text-sm text-slate-500">
                              {line.uom_code}
                            </span>
                          </p>
                        </div>
                        <dl className="mt-4 grid grid-cols-2 gap-3 text-sm sm:grid-cols-3">
                          <div>
                            <dt className="text-xs text-slate-500">Received</dt>
                            <dd className="mt-1 text-slate-800">
                              {line.completed_receipt_qty} {line.uom_code}
                            </dd>
                          </div>
                          <div>
                            <dt className="text-xs text-slate-500">
                              Expected lot
                            </dt>
                            <dd className="mt-1 text-slate-800">
                              {line.expected_lot_no || "Not set"}
                            </dd>
                          </div>
                          <div>
                            <dt className="text-xs text-slate-500">
                              Customer reference
                            </dt>
                            <dd className="mt-1 text-slate-800">
                              {line.customer_line_reference || "Not set"}
                            </dd>
                          </div>
                        </dl>
                        {canEditDraft ? (
                          <div className="mt-3 flex justify-end gap-1 border-t border-slate-100 pt-2">
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => setLineEditing(line)}
                            >
                              <Pencil className="size-4" />
                              Edit
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              className="text-rose-700"
                              disabled={(order.lines?.length ?? 0) <= 1}
                              onClick={() => {
                                setDeleteLine(line);
                                setAction("delete-line");
                              }}
                            >
                              <Trash2 className="size-4" />
                              Remove
                            </Button>
                          </div>
                        ) : null}
                      </article>
                    ))}
                  </div>
                </section>
              </div>
            ) : null}
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>

      {headerEditing && order ? (
        <InboundOrderHeaderDialog
          order={order}
          onOpenChange={(open) => !open && setHeaderEditing(false)}
        />
      ) : null}
      {lineEditing !== undefined && order ? (
        <InboundOrderLineDialog
          order={order}
          line={lineEditing ?? undefined}
          onOpenChange={(open) => !open && setLineEditing(undefined)}
        />
      ) : null}

      <Dialog.Root
        open={action !== undefined}
        onOpenChange={(open) => {
          if (!transition.isPending && !open) {
            setAction(undefined);
            setDeleteLine(undefined);
          }
        }}
      >
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-[60] bg-slate-950/60" />
          <Dialog.Content className="fixed top-1/2 left-1/2 z-[70] w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl bg-white p-6 shadow-2xl focus:outline-none">
            <div className="grid size-11 place-items-center rounded-xl bg-amber-100 text-amber-800">
              <AlertTriangle className="size-5" />
            </div>
            <Dialog.Title className="mt-4 text-lg font-bold text-slate-950">
              {action === "release"
                ? "Release Inbound Order?"
                : action === "close"
                  ? "Close Inbound Order short?"
                  : action === "delete-line"
                    ? "Remove expected line?"
                    : "Cancel Inbound Order?"}
            </Dialog.Title>
            <Dialog.Description className="mt-2 text-sm leading-6 text-slate-600">
              {action === "release"
                ? "Released orders are ready for receipt creation and can no longer be edited."
                : action === "delete-line"
                  ? `${deleteLine?.item_name ?? "This line"} will be removed from the draft.`
                  : "Open receipts must be completed or cancelled before this transition."}
            </Dialog.Description>
            {action === "close" || action === "cancel" ? (
              <label className="mt-4 block text-sm font-semibold text-slate-800">
                Reason *
                <textarea
                  value={reason}
                  onChange={(event) => setReason(event.target.value)}
                  rows={3}
                  maxLength={4000}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                />
              </label>
            ) : null}
            {transition.error ? (
              <div
                role="alert"
                className="mt-4 rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
              >
                <p>{transition.error.message}</p>
                {transition.error instanceof ApiError &&
                transition.error.status === 409 ? (
                  <Button
                    size="sm"
                    variant="secondary"
                    className="mt-3"
                    onClick={async () => {
                      transition.reset();
                      await detail.refetch();
                    }}
                  >
                    Reload latest data
                  </Button>
                ) : null}
              </div>
            ) : null}
            <div className="mt-6 flex justify-end gap-2">
              <Button
                variant="secondary"
                disabled={transition.isPending}
                onClick={() => setAction(undefined)}
              >
                Back
              </Button>
              <Button
                className={
                  action === "cancel" || action === "delete-line"
                    ? "bg-rose-700 hover:bg-rose-800"
                    : undefined
                }
                disabled={
                  transition.isPending ||
                  ((action === "close" || action === "cancel") &&
                    !reason.trim())
                }
                onClick={() => action && transition.mutate(action)}
              >
                {transition.isPending ? (
                  <LoaderCircle className="size-4 animate-spin" />
                ) : null}
                Confirm
              </Button>
            </div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </>
  );
}
