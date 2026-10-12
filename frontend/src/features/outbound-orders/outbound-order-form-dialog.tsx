"use client";

import { useId, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { Select } from "@/components/ui/select";
import { createOutboundOrder, outboundOrderKeys } from "./outbound-order-api";
import {
  OutboundHeaderEditor,
  OutboundLineEditor,
  TextField,
  useOutboundPartners,
} from "./outbound-order-editors";
import {
  emptyOutboundLine,
  headerPayload,
  linePayload,
  localDateTime,
  outboundCreateSchema,
  type OutboundCreateValues,
} from "./outbound-order-schema";

export function OutboundOrderFormDialog({
  ownerId,
  warehouseId,
  ownerLabel,
  warehouseLabel,
  scopeLabel,
  onOpenChange,
  onCreated,
}: {
  ownerId: string;
  warehouseId: string;
  ownerLabel: string;
  warehouseLabel: string;
  scopeLabel: string;
  onOpenChange: (open: boolean) => void;
  onCreated: (id: string) => void;
}) {
  const client = useQueryClient();
  const customerId = useId();
  const [search, setSearch] = useState("");
  const [error, setError] = useState("");
  const [values, setValues] = useState<OutboundCreateValues>(() => ({
    customer_id: "",
    client_delivery_order_no: "",
    business_date: localDateTime().slice(0, 10),
    ship_to_partner_id: "",
    requested_ship_at: localDateTime(),
    customer_order_no: "",
    external_reference: "",
    notes: "",
    lines: [emptyOutboundLine()],
  }));
  const partners = useOutboundPartners(ownerId, search);
  const mutation = useMutation({
    mutationFn: () =>
      createOutboundOrder({
        owner_id: ownerId,
        warehouse_id: warehouseId,
        customer_id: values.customer_id,
        client_delivery_order_no: values.client_delivery_order_no.trim(),
        business_date: values.business_date,
        ...headerPayload(values),
        lines: values.lines.map(linePayload),
      }),
    onSuccess: async (order) => {
      await client.invalidateQueries({ queryKey: outboundOrderKeys.all });
      toast.success("Delivery Order draft created.");
      onCreated(order.outbound_id);
    },
    onError: (err: Error) => {
      setError(err.message);
      toast.error(err.message);
    },
  });
  return (
    <OperationDialog
      title="Create delivery order"
      className="max-w-5xl"
      description={scopeLabel}
      busy={mutation.isPending}
      onOpenChange={onOpenChange}
    >
      <form
        className="space-y-6"
        onSubmit={(event) => {
          event.preventDefault();
          const parsed = outboundCreateSchema.safeParse(values);
          if (!parsed.success) {
            setError(
              parsed.error.issues
                .map((issue) => `${issue.path.join(".")}: ${issue.message}`)
                .join(" · "),
            );
            return;
          }
          setError("");
          mutation.mutate();
        }}
      >
        <fieldset
          disabled={mutation.isPending}
          inert={mutation.isPending}
          className="space-y-5"
        >
          <h3 className="text-xs font-bold tracking-wide text-slate-500 uppercase">
            Order header
          </h3>
          <TextField
            label="Search customers / ship-to partners"
            value={search}
            onChange={setSearch}
          />
          {partners.error && (
            <p role="alert" className="text-rose-700">
              {partners.error.message}
            </p>
          )}
          <OutboundHeaderEditor
            value={values}
            onChange={(next) => setValues({ ...values, ...next })}
            shipToOptions={partners.shipTo.map((partner) => ({
              value: partner.partner_id,
              label: `${partner.name} (${partner.code})`,
            }))}
          >
            <FormField label="Owner" htmlFor={`${customerId}-owner`} required>
              <Input id={`${customerId}-owner`} value={ownerLabel} readOnly />
            </FormField>
            <FormField
              label="Warehouse"
              htmlFor={`${customerId}-warehouse`}
              required
            >
              <Input
                id={`${customerId}-warehouse`}
                value={warehouseLabel}
                readOnly
              />
            </FormField>
            <FormField label="Customer" htmlFor={customerId} required>
              <Select
                className="mt-2"
                id={customerId}
                ariaLabel="Customer"
                value={values.customer_id}
                options={partners.customers.map((partner) => ({
                  value: partner.partner_id,
                  label: `${partner.name} (${partner.code})`,
                }))}
                placeholder="Select customer"
                onValueChange={(next) =>
                  setValues({ ...values, customer_id: next })
                }
              />
            </FormField>
            <TextField
              label="Client DO number"
              required
              value={values.client_delivery_order_no}
              onChange={(next) =>
                setValues({ ...values, client_delivery_order_no: next })
              }
            />
            <TextField
              label="Business date"
              type="date"
              required
              value={values.business_date}
              onChange={(next) => setValues({ ...values, business_date: next })}
            />
          </OutboundHeaderEditor>
          <section className="mt-7 space-y-4 border-t border-slate-200 pt-6">
            <div className="flex items-start justify-between gap-3">
              <div>
                <h3 className="text-xs font-bold tracking-wide text-slate-500 uppercase">
                  Order lines
                </h3>
                <p className="mt-1 text-sm text-slate-600">
                  Only active items and their base UOM or active configured
                  conversions are available.
                </p>
              </div>
              <Button
                type="button"
                variant="secondary"
                size="sm"
                disabled={values.lines.length >= 500}
                onClick={() =>
                  setValues({
                    ...values,
                    lines: [...values.lines, emptyOutboundLine()],
                  })
                }
              >
                <Plus className="size-4" />
                Add line
              </Button>
            </div>
            <p className="rounded-xl bg-cyan-50 p-4 text-sm text-cyan-900">
              The client can order in packaging UOMs. Each line freezes its
              conversion and base quantity; validation and release do not move
              stock.
            </p>
            {values.lines.map((line, index) => (
              <section
                key={index}
                className="space-y-4 rounded-xl border border-slate-200 p-4"
              >
                <div className="flex items-center justify-between">
                  <h3 className="font-bold">Line {index + 1}</h3>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="text-rose-700 hover:text-rose-800"
                    disabled={values.lines.length === 1}
                    onClick={() =>
                      setValues({
                        ...values,
                        lines: values.lines.filter(
                          (_, position) => position !== index,
                        ),
                      })
                    }
                  >
                    <Trash2 className="size-4" />
                    Remove
                  </Button>
                </div>
                <OutboundLineEditor
                  ownerId={ownerId}
                  value={line}
                  onChange={(next) =>
                    setValues({
                      ...values,
                      lines: values.lines.map((value, position) =>
                        position === index ? next : value,
                      ),
                    })
                  }
                />
              </section>
            ))}
          </section>
        </fieldset>
        {error && (
          <p
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-700"
          >
            {error}
          </p>
        )}
        <div className="flex justify-end gap-2 border-t border-slate-200 pt-5">
          <Button
            type="button"
            variant="secondary"
            disabled={mutation.isPending}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button disabled={mutation.isPending || partners.pending}>
            {mutation.isPending ? "Saving…" : "Create draft"}
          </Button>
        </div>
      </form>
    </OperationDialog>
  );
}
