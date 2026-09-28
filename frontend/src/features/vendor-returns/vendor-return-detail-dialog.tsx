"use client";

import { useState } from "react";
import Link from "next/link";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw, Undo2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { Textarea } from "@/components/ui/textarea";
import { quarantineKeys } from "@/features/quarantine/quarantine-api";
import {
  cancelVendorReturn,
  completeVendorReturn,
  getVendorReturn,
  vendorReturnKeys,
} from "./vendor-return-api";
import {
  vendorReturnLabel,
  vendorReturnTone,
  type VendorReturnCapabilities,
} from "./vendor-return-types";

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

export function VendorReturnDetailDialog({
  id,
  capabilities,
  onOpenChange,
}: {
  id: string;
  capabilities: VendorReturnCapabilities;
  onOpenChange: (open: boolean) => void;
}) {
  const client = useQueryClient();
  const [action, setAction] = useState<"complete" | "cancel">();
  const [reason, setReason] = useState("");
  const query = useQuery({
    queryKey: vendorReturnKeys.detail(id),
    queryFn: () => getVendorReturn(id),
    refetchOnWindowFocus: false,
  });
  const value = query.data;
  const finish = (updated: NonNullable<typeof value>, message: string) => {
    client.setQueryData(vendorReturnKeys.detail(id), updated);
    void client.invalidateQueries({ queryKey: vendorReturnKeys.all });
    void client.invalidateQueries({ queryKey: quarantineKeys.all });
    void client.invalidateQueries({ queryKey: ["inventory"] });
    setAction(undefined);
    setReason("");
    toast.success(message);
  };
  const complete = useMutation({
    mutationFn: () => {
      if (!value?.source_balance_version_no)
        throw new Error("Refresh the return before completing it.");
      return completeVendorReturn(id, {
        expected_version: value.version_no,
        expected_balance_version: value.source_balance_version_no,
        completed_at: new Date().toISOString(),
      });
    },
    onSuccess: (updated) => finish(updated, "Return to vendor completed."),
    onError: (error) => toast.error(error.message),
  });
  const cancel = useMutation({
    mutationFn: () => {
      if (!value) throw new Error("Refresh the return before cancelling it.");
      if (!reason.trim()) throw new Error("Enter a cancellation reason.");
      return cancelVendorReturn(id, {
        expected_version: value.version_no,
        reason: reason.trim(),
      });
    },
    onSuccess: (updated) => finish(updated, "Return to vendor cancelled."),
    onError: (error) => toast.error(error.message),
  });
  const busy = complete.isPending || cancel.isPending;

  return (
    <OperationDialog
      title={id}
      description="Controlled return of quarantined stock to its source vendor"
      busy={busy}
      closeLabel="Close return to vendor details"
      onOpenChange={onOpenChange}
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={busy || query.isFetching}
          onClick={() => void query.refetch()}
        >
          <RefreshCw className="size-4" /> Refresh return
        </Button>
        {query.isPending ? (
          <p className="flex items-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-4 animate-spin" /> Loading return…
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
            <StatusBadge tone={vendorReturnTone(value.status_code)}>
              {vendorReturnLabel(value.status_code)}
            </StatusBadge>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2 lg:grid-cols-3">
              <Detail
                label="Vendor"
                value={`${value.vendor_code} · ${value.vendor_name}`}
              />
              <Detail
                label="Item"
                value={`${value.item_code} · ${value.item_name}`}
              />
              <Detail label="Lot" value={value.lot_number} />
              <Detail
                label="Serial / handling unit"
                value={value.serial_no || value.handling_unit_barcode}
              />
              <Detail
                label="Quantity"
                value={`${value.quantity} ${value.uom_code}`}
              />
              <Detail
                label="Source location"
                value={value.source_location_code}
              />
              <Detail
                label="Inventory status"
                value={value.source_inventory_status_code}
              />
              <Detail label="Business date" value={value.business_date} />
              <Detail
                label="Planned by"
                value={value.created_by_display_name}
              />
              <Detail label="Planned at" value={value.planned_at} />
              <Detail label="Available now" value={value.available_qty} />
              <Detail label="Movement" value={value.inventory_movement_id} />
              <Detail label="Notes" value={value.notes} />
              <Detail
                label="Completed by"
                value={value.completed_by_display_name}
              />
              <Detail label="Completed at" value={value.completed_at} />
              <Detail
                label="Cancelled by"
                value={value.cancelled_by_display_name}
              />
              <Detail label="Cancelled at" value={value.cancelled_at} />
              <Detail
                label="Cancellation reason"
                value={value.cancellation_reason}
              />
            </dl>
            <Link
              className="inline-block rounded-lg border border-slate-200 px-3 py-2 text-sm font-semibold text-cyan-900"
              href={`/inbound/quarantine?${new URLSearchParams({ owner: value.owner_id, warehouse: value.warehouse_id, case: value.quarantine_case_id })}`}
            >
              Open quarantine case
            </Link>
            {value.status_code === "PLANNED" ? (
              <section className="space-y-3 rounded-xl border border-amber-200 bg-amber-50 p-4">
                <p className="text-sm text-amber-950">
                  Completing permanently removes {value.quantity}{" "}
                  {value.uom_code} from WMS inventory and records its return to{" "}
                  {value.vendor_name}. Confirm only after the physical handover
                  is authorized and performed.
                </p>
                {!action ? (
                  <div className="flex flex-wrap gap-2">
                    {capabilities.canComplete ? (
                      <Button onClick={() => setAction("complete")}>
                        <Undo2 className="size-4" /> Complete return
                      </Button>
                    ) : null}
                    {capabilities.canCancel ? (
                      <Button
                        variant="secondary"
                        onClick={() => setAction("cancel")}
                      >
                        Cancel transaction
                      </Button>
                    ) : null}
                  </div>
                ) : action === "complete" ? (
                  <div className="flex flex-wrap gap-2">
                    <Button
                      disabled={busy || !value.source_balance_version_no}
                      onClick={() => complete.mutate()}
                    >
                      {complete.isPending ? (
                        <LoaderCircle className="size-4 animate-spin" />
                      ) : null}
                      Confirm vendor return
                    </Button>
                    <Button
                      variant="secondary"
                      disabled={busy}
                      onClick={() => setAction(undefined)}
                    >
                      Back
                    </Button>
                  </div>
                ) : (
                  <div className="space-y-3">
                    <Textarea
                      aria-label="Cancellation reason"
                      rows={3}
                      maxLength={4000}
                      disabled={busy}
                      placeholder="Why is this return no longer required?"
                      value={reason}
                      onChange={(event) => setReason(event.target.value)}
                    />
                    <div className="flex flex-wrap gap-2">
                      <Button
                        disabled={busy || !reason.trim()}
                        onClick={() => cancel.mutate()}
                      >
                        {cancel.isPending ? (
                          <LoaderCircle className="size-4 animate-spin" />
                        ) : null}
                        Confirm cancellation
                      </Button>
                      <Button
                        variant="secondary"
                        disabled={busy}
                        onClick={() => setAction(undefined)}
                      >
                        Back
                      </Button>
                    </div>
                  </div>
                )}
                {complete.error || cancel.error ? (
                  <p
                    role="alert"
                    className="rounded-lg bg-rose-50 p-3 text-sm text-rose-900"
                  >
                    {(complete.error ?? cancel.error)?.message}
                  </p>
                ) : null}
              </section>
            ) : null}
          </>
        ) : null}
      </div>
    </OperationDialog>
  );
}
