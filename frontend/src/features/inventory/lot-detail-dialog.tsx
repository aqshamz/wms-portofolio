"use client";

import { useQuery } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import {
  getItem,
  itemCatalogKeys,
} from "@/features/item-catalog/item-catalog-api";
import { getLot, lotKeys } from "./inventory-api";

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

export function LotDetailDialog({
  id,
  timezone,
  onOpenChange,
}: {
  id: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const lot = useQuery({
    queryKey: lotKeys.detail(id),
    queryFn: () => getLot(id),
  });
  const item = useQuery({
    queryKey: itemCatalogKeys.item(lot.data?.item_id ?? "none"),
    queryFn: () => getItem(lot.data!.item_id),
    enabled: Boolean(lot.data?.item_id),
  });

  return (
    <OperationDialog
      title="Lot identity"
      description="Owner-wide immutable lot registration"
      onOpenChange={onOpenChange}
      closeLabel="Close lot details"
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={lot.isFetching}
          onClick={() => void lot.refetch()}
        >
          <RefreshCw className="size-4" /> Refresh lot
        </Button>
        {lot.error ? (
          <div
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900"
          >
            {lot.error.message}
          </div>
        ) : null}
        {lot.isPending ? (
          <p className="flex items-center gap-2 text-sm text-slate-600">
            <LoaderCircle className="size-4 animate-spin" /> Loading lot…
          </p>
        ) : null}
        {lot.data ? (
          <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2">
            <Detail label="Lot number" value={lot.data.lot_number} />
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
              label="Manufacture date"
              value={lot.data.manufacture_date}
            />
            <Detail label="Expiry date" value={lot.data.expiry_date} />
            <Detail
              label="Registered"
              value={new Intl.DateTimeFormat("en-GB", {
                dateStyle: "medium",
                timeStyle: "short",
                timeZone: timezone,
              }).format(new Date(lot.data.created_at))}
            />
            <Detail label="Internal lot reference" value={lot.data.lot_id} />
          </dl>
        ) : null}
      </div>
    </OperationDialog>
  );
}
