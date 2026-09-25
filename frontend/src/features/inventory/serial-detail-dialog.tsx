"use client";

import { useQuery } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import {
  getItem,
  itemCatalogKeys,
} from "@/features/item-catalog/item-catalog-api";
import { getSerial, serialKeys } from "./inventory-api";

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

export function SerialDetailDialog({
  id,
  timezone,
  onOpenChange,
}: {
  id: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const serial = useQuery({
    queryKey: serialKeys.detail(id),
    queryFn: () => getSerial(id),
  });
  const item = useQuery({
    queryKey: itemCatalogKeys.item(serial.data?.item_id ?? "none"),
    queryFn: () => getItem(serial.data!.item_id),
    enabled: Boolean(serial.data?.item_id),
  });

  return (
    <OperationDialog
      title="Serial number identity"
      description="Owner-wide immutable serial registration"
      onOpenChange={onOpenChange}
      closeLabel="Close serial-number details"
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={serial.isFetching}
          onClick={() => void serial.refetch()}
        >
          <RefreshCw className="size-4" /> Refresh serial
        </Button>
        {serial.error ? (
          <div
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {serial.error.message}
          </div>
        ) : null}
        {serial.isPending ? (
          <p className="flex items-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-4 animate-spin" /> Loading serial…
          </p>
        ) : null}
        {serial.data ? (
          <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2">
            <Detail label="Serial number" value={serial.data.serial_no} />
            <Detail
              label="Item"
              value={
                item.data
                  ? `${item.data.code} · ${item.data.name}`
                  : item.isPending
                    ? "Loading item…"
                    : "Item unavailable"
              }
            />
            <Detail
              label="Registered"
              value={new Intl.DateTimeFormat("en-GB", {
                dateStyle: "medium",
                timeStyle: "short",
                timeZone: timezone,
              }).format(new Date(serial.data.created_at))}
            />
            <Detail
              label="Internal serial reference"
              value={serial.data.serial_id}
            />
          </dl>
        ) : null}
      </div>
    </OperationDialog>
  );
}
