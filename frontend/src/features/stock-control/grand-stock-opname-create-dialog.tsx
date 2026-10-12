"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { Textarea } from "@/components/ui/textarea";
import { createGrandStockOpname, stockControlKeys } from "./stock-control-api";

function businessDate(timezone: string) {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}

export function GrandStockOpnameCreateDialog({
  ownerId,
  ownerName,
  warehouseId,
  warehouseName,
  timezone,
  onOpenChange,
}: {
  ownerId: string;
  ownerName: string;
  warehouseId: string;
  warehouseName: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const client = useQueryClient();
  const [date, setDate] = useState(() => businessDate(timezone));
  const [tolerance, setTolerance] = useState("0");
  const [notes, setNotes] = useState("");
  const create = useMutation({
    mutationFn: () =>
      createGrandStockOpname({
        owner_id: ownerId,
        warehouse_id: warehouseId,
        business_date: date,
        tolerance_quantity: tolerance || "0",
        notes: notes.trim() || undefined,
      }),
    onSuccess: (count) => {
      toast.success(
        `${count.cycle_count_id} created with ${count.total_lines} blind-count lines. Stock is now frozen for this customer and warehouse.`,
      );
      void client.invalidateQueries({
        queryKey: stockControlKeys.cycleCounts(),
      });
      onOpenChange(false);
    },
    onError: (error) => toast.error(error.message),
  });
  return (
    <OperationDialog
      title="Create grand stock opname"
      description="Snapshot all eligible stock for one warehouse and served customer."
      busy={create.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close grand stock opname creation"
    >
      <form
        className="space-y-5"
        onSubmit={(event) => {
          event.preventDefault();
          create.mutate();
        }}
      >
        <section className="rounded-xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950">
          <p className="font-semibold">{ownerName}</p>
          <p className="mt-1">{warehouseName}</p>
          <p className="mt-3">
            The document includes all positive balances in active, unlocked,
            STORAGE and PICK_FACE locations for this customer only. Receiving,
            QC, quarantine and dock stock are excluded.
          </p>
        </section>
        {create.error ? (
          <p className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
            {create.error.message}
          </p>
        ) : null}
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Business date" htmlFor="grand-date" required>
            <Input
              id="grand-date"
              type="date"
              value={date}
              onChange={(event) => setDate(event.target.value)}
            />
          </FormField>
          <FormField
            label="Variance tolerance (base units)"
            htmlFor="grand-tolerance"
            required
          >
            <Input
              id="grand-tolerance"
              inputMode="decimal"
              value={tolerance}
              onChange={(event) => setTolerance(event.target.value)}
            />
          </FormField>
          <div className="sm:col-span-2">
            <FormField label="Notes" htmlFor="grand-notes">
              <Textarea
                id="grand-notes"
                maxLength={4000}
                value={notes}
                onChange={(event) => setNotes(event.target.value)}
              />
            </FormField>
          </div>
        </div>
        <section className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-950">
          <p className="font-semibold">
            Creating this opname freezes stock immediately.
          </p>
          <p className="mt-2">
            All stock movements and reservation changes for this customer in
            this warehouse are blocked, including receiving, QC, quarantine and
            dock stock outside the count. Only approved corrections from this
            opname can change stock.
          </p>
          <p className="mt-2">
            The freeze remains during partial posting and ends automatically
            when every line is finalized or the document is cancelled. Other
            customers and warehouses remain available. Only one active grand
            opname is allowed for this warehouse/customer.
          </p>
          <p className="mt-2">
            Complete or cancel outstanding stock reservations and replenishments
            before starting. Creation is blocked while any stock is reserved for
            this customer and warehouse.
          </p>
        </section>
        <div className="flex justify-end gap-3 border-t border-slate-200 pt-4">
          <Button
            type="button"
            variant="secondary"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={!date || create.isPending}>
            {create.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            Create and freeze stock
          </Button>
        </div>
      </form>
    </OperationDialog>
  );
}
