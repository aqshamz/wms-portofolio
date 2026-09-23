"use client";

import { useQuery } from "@tanstack/react-query";
import { LoaderCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { exceptionKeys, getInboundException } from "./inbound-exception-api";
import {
  exceptionLabel,
  exceptionTime,
  exceptionTone,
} from "./inbound-exception-types";

function Detail({ label, value }: { label: string; value: string | null }) {
  return (
    <div>
      <dt className="text-xs font-semibold text-slate-500">{label}</dt>
      <dd className="mt-1 text-sm break-words text-slate-900">
        {value ?? "—"}
      </dd>
    </div>
  );
}
export function ExceptionDetailDialog({
  exceptionId,
  timezone,
  onOpenChange,
}: {
  exceptionId: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const query = useQuery({
    queryKey: exceptionKeys.detail(exceptionId),
    queryFn: () => getInboundException(exceptionId),
  });
  const value = query.data;
  return (
    <OperationDialog
      title={exceptionId}
      description="Inbound exception audit record"
      closeLabel="Close exception details"
      onOpenChange={onOpenChange}
    >
      <div className="space-y-5">
        <Button
          variant="ghost"
          size="sm"
          disabled={query.isFetching}
          onClick={() => void query.refetch()}
        >
          <RefreshCw className="size-4" />
          Refresh exception
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
            Loading exception…
          </p>
        ) : null}
        {value ? (
          <>
            <StatusBadge tone={exceptionTone(value.exception_type_code)}>
              {exceptionLabel(value.exception_type_code)}
            </StatusBadge>
            <dl className="grid gap-4 rounded-xl bg-slate-50 p-4 sm:grid-cols-2">
              <Detail
                label="Source document"
                value={value.source_document_id}
              />
              <Detail label="Source line" value={value.source_line_id} />
              <Detail
                label="Recorded at"
                value={exceptionTime(value.created_at, timezone)}
              />
              <Detail
                label="Recorded by"
                value={value.created_by_display_name || null}
              />
              <Detail label="Owner" value={value.owner_name || null} />
              <Detail label="Warehouse" value={value.warehouse_name || null} />
            </dl>
            <section className="rounded-xl border border-slate-200 p-4">
              <h3 className="font-bold text-slate-950">Recorded quantities</h3>
              <p className="mt-1 text-xs text-slate-500">
                Quantities are recorded in the source document’s units.
                Cancellation and reversal records may not include quantities.
              </p>
              <dl className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-3">
                <Detail label="Expected quantity" value={value.expected_qty} />
                <Detail label="Actual quantity" value={value.actual_qty} />
                <Detail label="Variance quantity" value={value.variance_qty} />
              </dl>
            </section>
            <section>
              <h3 className="font-bold text-slate-950">Reason / notes</h3>
              <p className="mt-2 rounded-xl bg-slate-50 p-4 text-sm break-words whitespace-pre-wrap text-slate-700">
                {value.notes || "No notes recorded."}
              </p>
            </section>
            <p className="rounded-xl bg-amber-50 p-4 text-sm text-amber-900">
              This is an immutable audit record, not an open task. Any follow-up
              takes place on the source document; its workflow status is
              separate from this exception.
            </p>
          </>
        ) : null}
      </div>
    </OperationDialog>
  );
}
