"use client";

import { useQuery } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { getMovement, movementKeys } from "./inventory-api";

function Detail({ label, value }: { label: string; value?: string | null }) {
  return (
    <div>
      <dt className="text-xs font-semibold text-slate-500">{label}</dt>
      <dd className="mt-1 text-sm break-words text-slate-900">
        {value === null || value === undefined || value === "" ? "—" : value}
      </dd>
    </div>
  );
}

export function InventoryMovementDetailDialog({
  id,
  timezone,
  onOpenChange,
}: {
  id: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const query = useQuery({
    queryKey: movementKeys.detail(id),
    queryFn: () => getMovement(id),
  });
  const movement = query.data;
  const time = (value: string) =>
    new Intl.DateTimeFormat("en-GB", {
      dateStyle: "medium",
      timeStyle: "short",
      timeZone: timezone,
    }).format(new Date(value));

  return (
    <OperationDialog
      title="Inventory movement"
      description="Immutable stock movement record"
      onOpenChange={onOpenChange}
      closeLabel="Close inventory movement details"
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={query.isFetching}
          onClick={() => void query.refetch()}
        >
          <RefreshCw className="size-4" /> Refresh movement
        </Button>
        {query.error ? (
          <div
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {query.error.message}
          </div>
        ) : null}
        {query.isPending ? (
          <p className="flex items-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-4 animate-spin" /> Loading movement…
          </p>
        ) : null}
        {movement ? (
          <>
            <div className="flex flex-wrap gap-2">
              <StatusBadge tone="info">
                {movement.movement_type_code}
              </StatusBadge>
              <StatusBadge tone="neutral">{movement.business_date}</StatusBadge>
            </div>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2 lg:grid-cols-3">
              <Detail label="Item" value={movement.item_code} />
              <Detail
                label="Quantity"
                value={`${movement.quantity} ${movement.uom_code}`}
              />
              <Detail label="Occurred" value={time(movement.occurred_at)} />
              <Detail
                label="From location"
                value={movement.from_location_code ?? "Outside WMS"}
              />
              <Detail
                label="To location"
                value={movement.to_location_code ?? "Outside WMS"}
              />
              <Detail
                label="Status change"
                value={`${movement.from_status_code ?? "Outside WMS"} → ${movement.to_status_code ?? "Outside WMS"}`}
              />
              <Detail label="Lot" value={movement.lot_number} />
              <Detail
                label="Serialized stock"
                value={movement.serial_id ? "Yes" : "No"}
              />
              <Detail
                label="Handling unit"
                value={movement.handling_unit_id ? "Assigned" : null}
              />
            </dl>
            <dl className="grid gap-4 rounded-xl border border-slate-200 p-4 sm:grid-cols-2">
              <Detail
                label="Source document"
                value={movement.source_document_id}
              />
              <Detail label="Source line" value={movement.source_line_id} />
              <Detail label="Notes" value={movement.notes} />
              <Detail
                label="Internal movement reference"
                value={movement.movement_id}
              />
            </dl>
          </>
        ) : null}
      </div>
    </OperationDialog>
  );
}
