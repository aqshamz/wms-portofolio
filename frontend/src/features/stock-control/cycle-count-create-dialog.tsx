"use client";

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { Textarea } from "@/components/ui/textarea";
import { balanceKeys, listBalances } from "@/features/inventory/inventory-api";
import { createCycleCount, stockControlKeys } from "./stock-control-api";

function businessDate(timezone: string) {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}
export function CycleCountCreateDialog({
  ownerId,
  warehouseId,
  timezone,
  onOpenChange,
}: {
  ownerId: string;
  warehouseId: string;
  timezone: string;
  onOpenChange: (open: boolean) => void;
}) {
  const client = useQueryClient();
  const [search, setSearch] = useState("");
  const [selected, setSelected] = useState<string[]>([]);
  const [date, setDate] = useState(() => businessDate(timezone));
  const [tolerance, setTolerance] = useState("0");
  const [notes, setNotes] = useState("");
  const filters = useMemo(
    () => ({
      ownerId,
      warehouseId,
      search,
      includeZero: true,
      locationTypeCode: "STORAGE",
      page: 1,
      pageSize: 100,
    }),
    [ownerId, search, warehouseId],
  );
  const balances = useQuery({
    queryKey: balanceKeys.list(filters),
    queryFn: () => listBalances(filters),
  });
  const create = useMutation({
    mutationFn: () =>
      createCycleCount({
        business_date: date,
        tolerance_quantity: tolerance || "0",
        notes: notes.trim() || undefined,
        balance_ids: selected,
      }),
    onSuccess: (count) => {
      toast.success(
        `${count.cycle_count_id} created with ${count.total_lines} blind-count lines.`,
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
      title="Create cycle count"
      description="Snapshot selected STORAGE balances for blind physical counting."
      busy={create.isPending}
      onOpenChange={onOpenChange}
      closeLabel="Close cycle count creation"
    >
      <form
        className="space-y-5"
        onSubmit={(event) => {
          event.preventDefault();
          create.mutate();
        }}
      >
        <section className="rounded-xl border border-cyan-200 bg-cyan-50 p-4 text-sm text-cyan-950">
          Counters will not see system quantities until every line is counted
          and all required recounts are complete.
        </section>
        {balances.error || create.error ? (
          <p className="rounded-xl bg-rose-50 p-4 text-sm text-rose-900">
            {(balances.error ?? create.error)?.message}
          </p>
        ) : null}
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Business date" htmlFor="cycle-date" required>
            <Input
              id="cycle-date"
              type="date"
              value={date}
              onChange={(event) => setDate(event.target.value)}
            />
          </FormField>
          <FormField
            label="Variance tolerance (base units)"
            htmlFor="cycle-tolerance"
            required
          >
            <Input
              id="cycle-tolerance"
              inputMode="decimal"
              value={tolerance}
              onChange={(event) => setTolerance(event.target.value)}
            />
          </FormField>
          <div className="sm:col-span-2">
            <FormField label="Notes" htmlFor="cycle-notes">
              <Textarea
                id="cycle-notes"
                maxLength={4000}
                value={notes}
                onChange={(event) => setNotes(event.target.value)}
              />
            </FormField>
          </div>
        </div>
        <section className="overflow-hidden rounded-xl border border-slate-200">
          <div className="border-b border-slate-200 p-4">
            <FormField label="Search storage balances" htmlFor="cycle-search">
              <Input
                id="cycle-search"
                placeholder="Item, location or lot"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
              />
            </FormField>
          </div>
          <div className="max-h-80 divide-y divide-slate-200 overflow-y-auto">
            {(balances.data?.items ?? []).map((balance) => (
              <label
                key={balance.balance_id}
                className="flex cursor-pointer items-start gap-3 p-4 hover:bg-slate-50"
              >
                <input
                  className="mt-1"
                  type="checkbox"
                  checked={selected.includes(balance.balance_id)}
                  onChange={(event) =>
                    setSelected((current) =>
                      event.target.checked
                        ? [...current, balance.balance_id]
                        : current.filter((id) => id !== balance.balance_id),
                    )
                  }
                />
                <span className="text-sm">
                  <span className="font-semibold">
                    {balance.item_code} · {balance.location_code}
                  </span>
                  <span className="mt-1 block text-slate-600">
                    {balance.inventory_status_code} · {balance.uom_code}
                    {balance.lot_number ? ` · Lot ${balance.lot_number}` : ""}
                    {balance.serial_controlled ? " · Serial-controlled" : ""}
                  </span>
                </span>
              </label>
            ))}
            {!balances.isPending && !balances.data?.items.length ? (
              <p className="p-5 text-sm text-slate-500">
                No STORAGE balances match.
              </p>
            ) : null}
          </div>
        </section>
        <div className="flex justify-end gap-3 border-t border-slate-200 pt-4">
          <Button
            type="button"
            variant="secondary"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={!selected.length || create.isPending}>
            {create.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            Create {selected.length} line{selected.length === 1 ? "" : "s"}
          </Button>
        </div>
      </form>
    </OperationDialog>
  );
}
