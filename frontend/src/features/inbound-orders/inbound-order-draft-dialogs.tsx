"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Decimal from "decimal.js";
import { LoaderCircle, X } from "lucide-react";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  addInboundOrderLine,
  inboundOrderKeys,
  updateInboundOrder,
  updateInboundOrderLine,
} from "@/features/inbound-orders/inbound-order-api";
import {
  emptyInboundOrderLine,
  inboundOrderHeaderSchema,
  inboundOrderLineSchema,
  type InboundOrderHeaderValues,
  type InboundOrderLineValues,
} from "@/features/inbound-orders/inbound-order-schema";
import type {
  InboundOrder,
  InboundOrderLine,
} from "@/features/inbound-orders/inbound-order-types";
import {
  getPurchaseOrder,
  purchaseOrderKeys,
} from "@/features/purchase-orders/purchase-order-api";
import type { PurchaseOrderLine } from "@/features/purchase-orders/purchase-order-types";
import { ApiError } from "@/lib/api/client";

function optional(value: string) {
  return value.trim() || undefined;
}

function localDateTime(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  const offset = date.getTimezoneOffset() * 60_000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

function Frame({
  title,
  description,
  pending,
  onOpenChange,
  children,
}: {
  title: string;
  description: string;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
}) {
  return (
    <Dialog.Root open onOpenChange={(open) => !pending && onOpenChange(open)}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-[60] bg-slate-950/60" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-[70] max-h-[92vh] w-[calc(100%-2rem)] max-w-2xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="flex items-start justify-between gap-4 border-b border-slate-200 px-5 py-4 sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                {title}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                {description}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close draft editor"
                className="grid size-10 place-items-center rounded-lg text-slate-500 hover:bg-slate-100"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

function SaveError({ error }: { error: Error }) {
  return (
    <div
      role="alert"
      className="mb-5 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-900"
    >
      <p>{error.message}</p>
      {error instanceof ApiError && error.status === 409 ? (
        <p className="mt-2 font-semibold">
          Close this editor and reload the latest order version.
        </p>
      ) : null}
    </div>
  );
}

export function InboundOrderHeaderDialog({
  order,
  onOpenChange,
}: {
  order: InboundOrder;
  onOpenChange: (open: boolean) => void;
}) {
  const queryClient = useQueryClient();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<InboundOrderHeaderValues>({
    resolver: zodResolver(inboundOrderHeaderSchema),
    defaultValues: {
      expected_arrival_at: localDateTime(order.expected_arrival_at),
      external_reference: order.external_reference ?? "",
      supplier_reference: order.supplier_reference ?? "",
      notes: order.notes ?? "",
    },
  });
  const save = useMutation({
    mutationFn: (values: InboundOrderHeaderValues) =>
      updateInboundOrder(order.inbound_id, {
        expected_version: order.version_no,
        expected_arrival_at: values.expected_arrival_at
          ? new Date(values.expected_arrival_at).toISOString()
          : undefined,
        external_reference: optional(values.external_reference),
        supplier_reference: optional(values.supplier_reference),
        notes: optional(values.notes),
      }),
    onSuccess: async (updated) => {
      queryClient.setQueryData(
        inboundOrderKeys.detail(order.inbound_id),
        updated,
      );
      await queryClient.invalidateQueries({
        queryKey: inboundOrderKeys.lists(),
      });
      toast.success("Inbound Order header updated.");
      onOpenChange(false);
    },
  });

  return (
    <Frame
      title="Edit Inbound Order header"
      description="The source Purchase Order, owner, warehouse, and business date remain fixed."
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      <form
        className="p-5 sm:p-6"
        onSubmit={handleSubmit((values) => save.mutate(values))}
      >
        {save.error ? <SaveError error={save.error} /> : null}
        <div className="grid gap-5 sm:grid-cols-2">
          <FormField
            label="Expected arrival"
            htmlFor="edit-inbound-arrival"
            error={errors.expected_arrival_at?.message}
          >
            <Input
              id="edit-inbound-arrival"
              type="datetime-local"
              {...register("expected_arrival_at")}
            />
          </FormField>
          <FormField
            label="External reference"
            htmlFor="edit-inbound-external"
            error={errors.external_reference?.message}
          >
            <Input
              id="edit-inbound-external"
              {...register("external_reference")}
            />
          </FormField>
          <FormField
            label="Supplier reference"
            htmlFor="edit-inbound-supplier"
            error={errors.supplier_reference?.message}
          >
            <Input
              id="edit-inbound-supplier"
              {...register("supplier_reference")}
            />
          </FormField>
          <div className="sm:col-span-2">
            <FormField
              label="Notes"
              htmlFor="edit-inbound-notes"
              error={errors.notes?.message}
            >
              <textarea
                id="edit-inbound-notes"
                rows={4}
                className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                {...register("notes")}
              />
            </FormField>
          </div>
        </div>
        <div className="mt-6 flex justify-end gap-2 border-t border-slate-200 pt-5">
          <Dialog.Close asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </Dialog.Close>
          <Button type="submit" disabled={save.isPending}>
            {save.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            Save changes
          </Button>
        </div>
      </form>
    </Frame>
  );
}

function remaining(line: PurchaseOrderLine) {
  return Decimal.max(
    new Decimal(line.ordered_qty).minus(line.scheduled_qty),
    0,
  ).toFixed(6);
}

export function InboundOrderLineDialog({
  order,
  line,
  onOpenChange,
}: {
  order: InboundOrder;
  line?: InboundOrderLine;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = Boolean(line);
  const queryClient = useQueryClient();
  const purchaseOrder = useQuery({
    queryKey: purchaseOrderKeys.detail(order.purchase_order_id),
    queryFn: () => getPurchaseOrder(order.purchase_order_id),
  });
  const {
    control,
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<InboundOrderLineValues>({
    resolver: zodResolver(inboundOrderLineSchema),
    defaultValues: line
      ? {
          purchase_order_line_id: line.purchase_order_line_id ?? "",
          expected_qty: line.expected_qty,
          customer_line_reference: line.customer_line_reference ?? "",
          notes: line.notes ?? "",
        }
      : { ...emptyInboundOrderLine },
  });
  const usedLineIds = new Set(
    (order.lines ?? [])
      .map((item) => item.purchase_order_line_id)
      .filter(Boolean),
  );
  const availableLines = (purchaseOrder.data?.lines ?? []).filter(
    (item) =>
      !usedLineIds.has(item.purchase_order_line_id) &&
      new Decimal(remaining(item)).greaterThan(0),
  );
  const save = useMutation({
    mutationFn: (values: InboundOrderLineValues) =>
      editing && line
        ? updateInboundOrderLine(order.inbound_id, line.inbound_line_id, {
            expected_version: order.version_no,
            expected_qty: values.expected_qty.trim(),
            customer_line_reference: optional(values.customer_line_reference),
            notes: optional(values.notes),
          })
        : addInboundOrderLine(order.inbound_id, order.version_no, {
            purchase_order_line_id: values.purchase_order_line_id,
            expected_qty: values.expected_qty.trim(),
            customer_line_reference: optional(values.customer_line_reference),
            notes: optional(values.notes),
          }),
    onSuccess: async (updated) => {
      queryClient.setQueryData(
        inboundOrderKeys.detail(order.inbound_id),
        updated,
      );
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: inboundOrderKeys.lists() }),
        queryClient.invalidateQueries({ queryKey: purchaseOrderKeys.all }),
      ]);
      toast.success(
        editing ? "Inbound Order line updated." : "Inbound Order line added.",
      );
      onOpenChange(false);
    },
  });

  return (
    <Frame
      title={editing ? "Edit expected line" : "Add expected line"}
      description={
        editing
          ? "Update the expected quantity and references before release."
          : "Add another unscheduled line from the source Purchase Order."
      }
      pending={save.isPending}
      onOpenChange={onOpenChange}
    >
      <form
        className="p-5 sm:p-6"
        onSubmit={handleSubmit((values) => save.mutate(values))}
      >
        {save.error ? <SaveError error={save.error} /> : null}
        <div className="grid gap-5 sm:grid-cols-2">
          <div className="sm:col-span-2">
            <FormField
              label="Purchase Order line"
              htmlFor="edit-inbound-line-source"
              required
              error={errors.purchase_order_line_id?.message}
            >
              {editing ? (
                <Input
                  id="edit-inbound-line-source"
                  value={`${line?.item_name} (${line?.item_code}) · ${line?.uom_code}`}
                  readOnly
                />
              ) : (
                <Controller
                  control={control}
                  name="purchase_order_line_id"
                  render={({ field }) => (
                    <Select
                      id="edit-inbound-line-source"
                      ariaLabel="Purchase Order line"
                      value={field.value}
                      options={availableLines.map((item) => ({
                        value: item.purchase_order_line_id,
                        label: `${item.item_name} (${item.item_code}) · ${remaining(item)} ${item.uom_code} unscheduled`,
                      }))}
                      placeholder={
                        purchaseOrder.isPending
                          ? "Loading Purchase Order lines…"
                          : availableLines.length
                            ? "Select a line"
                            : "No unscheduled lines"
                      }
                      disabled={purchaseOrder.isPending}
                      invalid={Boolean(errors.purchase_order_line_id)}
                      className="mt-2"
                      onValueChange={field.onChange}
                    />
                  )}
                />
              )}
            </FormField>
          </div>
          <FormField
            label="Expected quantity"
            htmlFor="edit-inbound-line-qty"
            required
            error={errors.expected_qty?.message}
          >
            <Input
              id="edit-inbound-line-qty"
              inputMode="decimal"
              invalid={Boolean(errors.expected_qty)}
              {...register("expected_qty")}
            />
          </FormField>
          <FormField
            label="Customer line reference"
            htmlFor="edit-inbound-line-reference"
            error={errors.customer_line_reference?.message}
          >
            <Input
              id="edit-inbound-line-reference"
              {...register("customer_line_reference")}
            />
          </FormField>
          <div className="sm:col-span-2">
            <FormField
              label="Notes"
              htmlFor="edit-inbound-line-notes"
              error={errors.notes?.message}
            >
              <textarea
                id="edit-inbound-line-notes"
                rows={3}
                className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100"
                {...register("notes")}
              />
            </FormField>
          </div>
        </div>
        <div className="mt-6 flex justify-end gap-2 border-t border-slate-200 pt-5">
          <Dialog.Close asChild>
            <Button type="button" variant="secondary">
              Cancel
            </Button>
          </Dialog.Close>
          <Button type="submit" disabled={save.isPending}>
            {save.isPending ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : null}
            {editing ? "Save line" : "Add line"}
          </Button>
        </div>
      </form>
    </Frame>
  );
}
