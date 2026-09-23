"use client";

import { useQuery } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { getSerialState, serialStateKeys } from "./inventory-api";

function Detail({
  label,
  value,
}: {
  label: string;
  value?: string | number | null;
}) {
  return (
    <div>
      <dt className="text-xs font-semibold text-slate-500">{label}</dt>
      <dd className="mt-1 text-sm break-words text-slate-900">
        {value === null || value === undefined || value === "" ? "—" : value}
      </dd>
    </div>
  );
}

export function SerialStateDetailDialog({
  id,
  timezone,
  onOpenChange,
}: {
  id: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const query = useQuery({
    queryKey: serialStateKeys.detail(id),
    queryFn: () => getSerialState(id),
  });
  const state = query.data;

  return (
    <OperationDialog
      title="Serial state"
      description="Current warehouse position of a serialized unit"
      onOpenChange={onOpenChange}
      closeLabel="Close serial-state details"
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={query.isFetching}
          onClick={() => void query.refetch()}
        >
          <RefreshCw className="size-4" /> Refresh state
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
            <LoaderCircle className="size-4 animate-spin" /> Loading serial
            state…
          </p>
        ) : null}
        {state ? (
          <>
            <div className="flex flex-wrap gap-2">
              <StatusBadge tone="info">
                {state.balance.inventory_status_code}
              </StatusBadge>
              <StatusBadge tone="neutral">
                State version {state.version_no}
              </StatusBadge>
            </div>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2 lg:grid-cols-3">
              <Detail label="Serial number" value={state.serial_no} />
              <Detail
                label="Item"
                value={`${state.balance.item_code} · ${state.balance.item_name}`}
              />
              <Detail label="Location" value={state.balance.location_code} />
              <Detail label="Lot" value={state.balance.lot_number} />
              <Detail
                label="Handling unit"
                value={state.balance.handling_unit_id ? "Assigned" : null}
              />
              <Detail
                label="Updated"
                value={new Intl.DateTimeFormat("en-GB", {
                  dateStyle: "medium",
                  timeStyle: "short",
                  timeZone: timezone,
                }).format(new Date(state.updated_at))}
              />
            </dl>
            <section className="rounded-xl border border-slate-200 p-4">
              <h3 className="font-bold text-slate-950">Containing balance</h3>
              <p className="mt-1 text-xs text-slate-500">
                These quantities describe the full balance dimension containing
                this serial, not the quantity of this individual unit.
              </p>
              <dl className="mt-4 grid gap-4 sm:grid-cols-3">
                <Detail
                  label="On hand"
                  value={`${state.balance.on_hand_qty} ${state.balance.uom_code}`}
                />
                <Detail
                  label="Reserved"
                  value={`${state.balance.reserved_qty} ${state.balance.uom_code}`}
                />
                <Detail
                  label="Available"
                  value={`${state.balance.available_qty} ${state.balance.uom_code}`}
                />
              </dl>
            </section>
          </>
        ) : null}
      </div>
    </OperationDialog>
  );
}
