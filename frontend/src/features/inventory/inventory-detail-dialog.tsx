"use client";

import { useQuery } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { balanceKeys, getBalance } from "./inventory-api";

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

export function InventoryBalanceDetailDialog({
  id,
  timezone,
  onOpenChange,
}: {
  id: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const query = useQuery({
    queryKey: balanceKeys.detail(id),
    queryFn: () => getBalance(id),
  });
  const balance = query.data;

  return (
    <OperationDialog
      title="Inventory balance"
      description="Read-only stock record"
      onOpenChange={onOpenChange}
      closeLabel="Close inventory balance details"
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={query.isFetching}
          onClick={() => void query.refetch()}
        >
          <RefreshCw className="size-4" />
          Refresh balance
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
            <LoaderCircle className="size-4 animate-spin" />
            Loading inventory balance…
          </p>
        ) : null}

        {balance ? (
          <>
            <div className="flex flex-wrap gap-2">
              <StatusBadge tone="info">
                {balance.inventory_status_code}
              </StatusBadge>
              <StatusBadge tone="neutral">
                Version {balance.version_no}
              </StatusBadge>
            </div>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2 lg:grid-cols-3">
              <Detail
                label="Item"
                value={`${balance.item_code} · ${balance.item_name}`}
              />
              <Detail label="Location" value={balance.location_code} />
              <Detail label="Lot" value={balance.lot_number} />
              <Detail
                label="Handling unit"
                value={balance.handling_unit_id ? "Assigned" : null}
              />
              <Detail
                label="Updated"
                value={new Intl.DateTimeFormat("en-GB", {
                  dateStyle: "medium",
                  timeStyle: "short",
                  timeZone: timezone,
                }).format(new Date(balance.updated_at))}
              />
              <Detail
                label="Internal balance reference"
                value={balance.balance_id}
              />
            </dl>
            <dl className="grid gap-4 rounded-xl border border-slate-200 p-4 sm:grid-cols-3">
              <Detail
                label="On hand"
                value={`${balance.on_hand_qty} ${balance.uom_code}`}
              />
              <Detail
                label="Reserved"
                value={`${balance.reserved_qty} ${balance.uom_code}`}
              />
              <Detail
                label="Available"
                value={`${balance.available_qty} ${balance.uom_code}`}
              />
            </dl>
          </>
        ) : null}
      </div>
    </OperationDialog>
  );
}
