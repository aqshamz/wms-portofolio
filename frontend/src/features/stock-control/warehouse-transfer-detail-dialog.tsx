"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { Select } from "@/components/ui/select";
import { StatusBadge } from "@/components/ui/status-badge";
import { Textarea } from "@/components/ui/textarea";
import {
  listLocations,
  listLocationTypes,
  storageLayoutKeys,
} from "@/features/storage-layout/storage-layout-api";
import {
  approveWarehouseTransfer,
  cancelWarehouseTransfer,
  dispatchWarehouseTransfer,
  getWarehouseTransfer,
  putawayWarehouseTransfer,
  receiveWarehouseTransfer,
  stockControlKeys,
} from "./stock-control-api";

function today(timezone: string) {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}

export function WarehouseTransferDetailDialog({
  id,
  timezone,
  canTransfer,
  canReceive,
  canPutaway,
  onOpenChange,
}: {
  id: string;
  timezone: string;
  canTransfer: boolean;
  canReceive: boolean;
  canPutaway: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const client = useQueryClient();
  const [date, setDate] = useState(() => today(timezone));
  const [receiptLocationId, setReceiptLocationId] = useState("");
  const [targetLocationId, setTargetLocationId] = useState("");
  const [reason, setReason] = useState("");
  const detail = useQuery({
    queryKey: stockControlKeys.warehouseTransfer(id),
    queryFn: () => getWarehouseTransfer(id),
  });
  const transfer = detail.data;
  const types = useQuery({
    queryKey: storageLayoutKeys.locationTypes("active"),
    queryFn: () => listLocationTypes("active"),
    enabled: transfer?.status_code === "IN_TRANSIT",
  });
  const receivingTypeIds = useMemo(
    () =>
      new Set(
        (types.data ?? [])
          .filter((type) => type.is_active && type.allows_receiving)
          .map((type) => type.location_type_id),
      ),
    [types.data],
  );
  const storageTypeIds = useMemo(
    () =>
      new Set(
        (types.data ?? [])
          .filter((type) => type.is_active && type.allows_storage)
          .map((type) => type.location_type_id),
      ),
    [types.data],
  );
  const locationFilters = {
    warehouseId: transfer?.target_warehouse_id ?? "",
    search: "",
    active: "active" as const,
    page: 1,
    pageSize: 100,
  };
  const locations = useQuery({
    queryKey: storageLayoutKeys.locations(locationFilters),
    queryFn: () => listLocations(locationFilters),
    enabled: Boolean(transfer?.status_code === "IN_TRANSIT"),
  });
  const receiptLocations = (locations.data?.items ?? []).filter(
    (entry) => receivingTypeIds.has(entry.location_type_id) && !entry.is_locked,
  );
  const storageLocations = (locations.data?.items ?? []).filter(
    (entry) => storageTypeIds.has(entry.location_type_id) && !entry.is_locked,
  );
  const refresh = (next: unknown) => {
    client.setQueryData(stockControlKeys.warehouseTransfer(id), next);
    void client.invalidateQueries({
      queryKey: stockControlKeys.warehouseTransfers(),
    });
    void client.invalidateQueries({ queryKey: ["inventory", "balances"] });
    void client.invalidateQueries({ queryKey: ["inventory", "movements"] });
  };
  const action = useMutation({
    mutationFn: async (
      kind: "approve" | "dispatch" | "cancel" | "receive" | "putaway",
    ) => {
      if (!transfer) throw new Error("Transfer is unavailable.");
      if (kind === "approve")
        return approveWarehouseTransfer(id, transfer.version_no);
      if (kind === "dispatch")
        return dispatchWarehouseTransfer(id, transfer.version_no);
      if (kind === "cancel") {
        if (!reason.trim()) throw new Error("Cancellation reason is required.");
        return cancelWarehouseTransfer(id, transfer.version_no, reason.trim());
      }
      if (kind === "receive") {
        if (!receiptLocationId || !targetLocationId)
          throw new Error(
            "Select both the receiving and putaway target locations.",
          );
        return receiveWarehouseTransfer(id, {
          expected_version: transfer.version_no,
          business_date: date,
          receipt_location_id: receiptLocationId,
          putaway_target_location_id: targetLocationId,
        });
      }
      const line = transfer.lines[0];
      if (!line?.received_balance_version_no)
        throw new Error(
          "Received balance version is unavailable; refresh the transfer.",
        );
      return putawayWarehouseTransfer(id, {
        expected_version: transfer.version_no,
        expected_balance_version: line.received_balance_version_no,
        business_date: date,
      });
    },
    onSuccess: (next) => {
      refresh(next);
      toast.success("Warehouse transfer updated.");
    },
    onError: (error) => toast.error(error.message),
  });
  const line = transfer?.lines[0];
  return (
    <OperationDialog
      title={transfer?.warehouse_transfer_id ?? "Warehouse transfer"}
      description="Dispatch, destination receipt and putaway audit trail"
      busy={action.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close warehouse transfer"
    >
      {detail.isPending ? (
        <div className="flex min-h-48 items-center justify-center gap-2">
          <LoaderCircle className="size-5 animate-spin" /> Loading transfer…
        </div>
      ) : !transfer || !line ? (
        <p className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
          {detail.error?.message ?? "Transfer not found."}
        </p>
      ) : (
        <div className="space-y-5">
          <div className="flex items-center justify-between">
            <StatusBadge
              tone={
                transfer.status_code === "CANCELLED"
                  ? "danger"
                  : transfer.status_code === "RECEIVED"
                    ? "success"
                    : "info"
              }
            >
              {transfer.status_code}
            </StatusBadge>
            <span className="text-xs text-slate-500">
              Version {transfer.version_no}
            </span>
          </div>
          <section className="grid gap-3 rounded-xl bg-slate-50 p-4 text-sm sm:grid-cols-2">
            <Info
              label="Route"
              value={`${transfer.source_warehouse_name} → ${transfer.target_warehouse_name}`}
            />
            <Info label="Owner" value={transfer.owner_name} />
            <Info
              label="Item"
              value={`${line.item_code} · ${line.item_name}`}
            />
            <Info
              label="Quantity"
              value={`${line.quantity} ${line.uom_code}`}
            />
            <Info
              label="Source"
              value={`${line.source_location_code} · ${line.source_inventory_status_code}`}
            />
            <Info
              label="Lot / serial"
              value={line.serial_number ?? line.lot_number ?? "—"}
            />
          </section>
          {transfer.status_code === "IN_TRANSIT" && canReceive ? (
            <section className="grid gap-4 rounded-xl border border-cyan-200 p-4 sm:grid-cols-2">
              <FormField
                label="Receipt location"
                htmlFor="transfer-receipt-location"
                required
              >
                <Select
                  id="transfer-receipt-location"
                  ariaLabel="Transfer receipt location"
                  className="mt-2"
                  value={receiptLocationId}
                  options={receiptLocations.map((entry) => ({
                    value: entry.location_id,
                    label: entry.code,
                  }))}
                  placeholder="Select receiving-capable location"
                  onValueChange={setReceiptLocationId}
                />
              </FormField>
              <FormField
                label="Putaway target"
                htmlFor="transfer-putaway-target"
                required
              >
                <Select
                  id="transfer-putaway-target"
                  ariaLabel="Transfer putaway target"
                  className="mt-2"
                  value={targetLocationId}
                  options={storageLocations.map((entry) => ({
                    value: entry.location_id,
                    label: entry.code,
                  }))}
                  placeholder="Select storage location"
                  onValueChange={setTargetLocationId}
                />
              </FormField>
              <FormField label="Business date" htmlFor="transfer-receive-date">
                <Input
                  id="transfer-receive-date"
                  type="date"
                  value={date}
                  onChange={(event) => setDate(event.target.value)}
                />
              </FormField>
              <div className="flex items-end">
                <Button
                  className="w-full"
                  onClick={() => action.mutate("receive")}
                >
                  Receive transfer
                </Button>
              </div>
            </section>
          ) : null}
          {transfer.status_code === "RECEIVED" ? (
            <section className="rounded-xl border border-slate-200 p-4 text-sm">
              <p>
                Received at <strong>{line.receipt_location_code}</strong>.
                Putaway target:{" "}
                <strong>{line.putaway_target_location_code}</strong>.
              </p>
              {line.putaway_completed_at ? (
                <p className="mt-2 text-emerald-700">Putaway completed.</p>
              ) : canPutaway ? (
                <div className="mt-4 flex justify-end">
                  <Button onClick={() => action.mutate("putaway")}>
                    Complete putaway
                  </Button>
                </div>
              ) : null}
            </section>
          ) : null}
          {(transfer.status_code === "DRAFT" ||
            transfer.status_code === "APPROVED") &&
          canTransfer ? (
            <section className="space-y-3 border-t border-slate-200 pt-4">
              <div className="flex flex-wrap justify-end gap-2">
                {transfer.status_code === "DRAFT" ? (
                  <Button onClick={() => action.mutate("approve")}>
                    Approve
                  </Button>
                ) : (
                  <Button onClick={() => action.mutate("dispatch")}>
                    Dispatch
                  </Button>
                )}
              </div>
              <FormField
                label="Cancellation reason"
                htmlFor="transfer-cancel-reason"
              >
                <Textarea
                  id="transfer-cancel-reason"
                  value={reason}
                  onChange={(event) => setReason(event.target.value)}
                />
              </FormField>
              <div className="flex justify-end">
                <Button
                  className="bg-rose-700 hover:bg-rose-600"
                  onClick={() => action.mutate("cancel")}
                >
                  Cancel transfer
                </Button>
              </div>
            </section>
          ) : null}
          {action.error ? (
            <p
              role="alert"
              className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
            >
              {action.error.message}
            </p>
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
