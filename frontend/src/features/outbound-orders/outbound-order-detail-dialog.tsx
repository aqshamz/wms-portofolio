"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { StatusBadge } from "@/components/ui/status-badge";
import { Textarea } from "@/components/ui/textarea";
import {
  addOutboundOrderLine,
  allocateAvailableOutboundStock,
  deleteOutboundOrderLine,
  getOutboundOrder,
  outboundOrderKeys,
  transitionOutboundOrder,
  updateOutboundOrder,
  updateOutboundOrderLine,
  validateOutboundOrder,
} from "./outbound-order-api";
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
  outboundHeaderSchema,
  outboundLineSchema,
  type OutboundHeaderValues,
  type OutboundLineValues,
} from "./outbound-order-schema";
import type {
  OutboundOrder,
  OutboundOrderLine,
  OutboundValidationRun,
} from "./outbound-order-types";

type Edit = { kind: "header" } | { kind: "line"; line?: OutboundOrderLine };
type Action =
  | { kind: "validate" }
  | { kind: "release" }
  | { kind: "cancel" }
  | { kind: "close-short" }
  | { kind: "allocate" }
  | { kind: "delete"; line: OutboundOrderLine }
  | { kind: "header"; value: OutboundHeaderValues }
  | { kind: "line"; value: OutboundLineValues; line?: OutboundOrderLine };

function DraftEditor({
  order,
  edit,
  busy,
  onSave,
  onBack,
}: {
  order: OutboundOrder;
  edit: Edit;
  busy: boolean;
  onSave: (action: Action) => void;
  onBack: () => void;
}) {
  const [search, setSearch] = useState("");
  const partners = useOutboundPartners(order.owner_id, search);
  const [header, setHeader] = useState<OutboundHeaderValues>(() => ({
    ship_to_partner_id: order.ship_to_partner_id ?? "",
    requested_ship_at: order.requested_ship_at
      ? localDateTime(new Date(order.requested_ship_at))
      : "",
    customer_order_no: order.customer_order_no ?? "",
    external_reference: order.external_reference ?? "",
    notes: order.notes ?? "",
  }));
  const savedLine = edit.kind === "line" ? edit.line : undefined;
  const [line, setLine] = useState<OutboundLineValues>(() =>
    savedLine
      ? {
          item_id: savedLine.item_id,
          ordered_qty: savedLine.ordered_qty,
          uom_id: savedLine.uom_id,
          requested_lot_no: savedLine.requested_lot_no ?? "",
          customer_line_reference: savedLine.customer_line_reference ?? "",
          notes: savedLine.notes ?? "",
        }
      : emptyOutboundLine(),
  );
  const [error, setError] = useState("");
  const options = partners.shipTo.map((partner) => ({
    value: partner.partner_id,
    label: `${partner.name} (${partner.code})`,
  }));
  if (
    order.ship_to_partner_id &&
    !options.some((option) => option.value === order.ship_to_partner_id)
  )
    options.unshift({
      value: order.ship_to_partner_id,
      label: order.ship_to_name,
    });
  return (
    <form
      className="space-y-4 rounded-xl border border-cyan-200 p-4"
      onSubmit={(event) => {
        event.preventDefault();
        if (busy) return;
        const parsed =
          edit.kind === "header"
            ? outboundHeaderSchema.safeParse(header)
            : outboundLineSchema.safeParse(line);
        if (!parsed.success) {
          setError(
            parsed.error.issues.map((issue) => issue.message).join(" · "),
          );
          return;
        }
        setError("");
        onSave(
          edit.kind === "header"
            ? { kind: "header", value: header }
            : { kind: "line", line: savedLine, value: line },
        );
      }}
    >
      <h3 className="font-bold">
        {edit.kind === "header"
          ? "Edit draft header"
          : savedLine
            ? `Edit line ${savedLine.line_no}`
            : "Add draft line"}
      </h3>
      <fieldset disabled={busy} inert={busy} className="space-y-4">
        {edit.kind === "header" ? (
          <>
            <TextField
              label="Search ship-to partners"
              value={search}
              onChange={setSearch}
            />
            <OutboundHeaderEditor
              value={header}
              onChange={setHeader}
              shipToOptions={options}
            />
          </>
        ) : (
          <OutboundLineEditor
            ownerId={order.owner_id}
            value={line}
            onChange={setLine}
            existing={savedLine}
          />
        )}
      </fieldset>
      {error && (
        <p role="alert" className="text-sm text-rose-700">
          {error}
        </p>
      )}
      <div className="flex justify-end gap-3">
        <Button
          variant="secondary"
          type="button"
          disabled={busy}
          onClick={onBack}
        >
          Back
        </Button>
        <Button disabled={busy}>Save draft changes</Button>
      </div>
    </form>
  );
}

export function OutboundOrderDetailDialog({
  id,
  canPlan,
  canCancel,
  canCloseShort = false,
  onOpenChange,
}: {
  id: string;
  canPlan: boolean;
  canCancel: boolean;
  canCloseShort?: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const client = useQueryClient();
  const query = useQuery({
    queryKey: outboundOrderKeys.detail(id),
    queryFn: () => getOutboundOrder(id),
  });
  const order = query.data;
  const [edit, setEdit] = useState<Edit>();
  const [confirmation, setConfirmation] = useState<Action>();
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");
  const [validation, setValidation] = useState<OutboundValidationRun>();
  const mutation = useMutation({
    mutationFn: async (action: Action) => {
      if (!order) throw new Error("Refresh the Delivery Order first.");
      const version = order.version_no;
      if (action.kind === "validate") {
        const run = await validateOutboundOrder(id, version);
        setValidation(run);
        if (run.status_code === "PASSED")
          toast.success("Delivery Order validated.");
        else toast.warning("Validation failed. Review the rule results.");
        return getOutboundOrder(id);
      }
      if (action.kind === "allocate") {
        const result = await allocateAvailableOutboundStock(id, version);
        return result.order;
      }
      if (
        action.kind === "release" ||
        action.kind === "cancel" ||
        action.kind === "close-short"
      )
        return transitionOutboundOrder(id, action.kind, version, reason.trim());
      if (action.kind === "delete")
        return deleteOutboundOrderLine(
          id,
          action.line.outbound_line_id,
          version,
        );
      if (action.kind === "header")
        return updateOutboundOrder(id, {
          ...headerPayload(action.value),
          expected_version: version,
        });
      const payload = linePayload(action.value);
      if (action.line)
        return updateOutboundOrderLine(
          id,
          action.line.outbound_line_id,
          version,
          {
            ordered_qty: payload.ordered_qty,
            uom_id: payload.uom_id,
            requested_lot_no: payload.requested_lot_no,
            customer_line_reference: payload.customer_line_reference,
            notes: payload.notes,
          },
        );
      return addOutboundOrderLine(id, version, payload);
    },
    onSuccess: async (next, action) => {
      client.setQueryData(outboundOrderKeys.detail(id), next);
      await client.invalidateQueries({ queryKey: outboundOrderKeys.all });
      if (action.kind === "allocate") await client.invalidateQueries();
      setEdit(undefined);
      setConfirmation(undefined);
      setError("");
      setReason("");
      if (action.kind !== "validate") {
        setValidation(undefined);
        if (action.kind === "allocate" && next.version_no === order?.version_no)
          toast.warning(
            "No additional stock available. Remaining demand stays open.",
          );
        else
          toast.success(
            action.kind === "allocate"
              ? "Available stock reserved. Shortages stay open on this DO."
              : "Delivery Order updated.",
          );
      }
    },
    onError: (err: Error) => {
      setError(err.message);
      toast.error(err.message);
    },
  });
  const busy = mutation.isPending;
  return (
    <OperationDialog
      title={id}
      description="Client Delivery Order, frozen quantities and validation"
      busy={busy}
      onOpenChange={onOpenChange}
    >
      <div className="space-y-5">
        <Button
          variant="secondary"
          disabled={busy || query.isFetching}
          onClick={() => {
            setEdit(undefined);
            setConfirmation(undefined);
            void query.refetch();
          }}
        >
          Refresh Delivery Order
        </Button>
        {query.isPending && <p>Loading Delivery Order…</p>}
        {(error || query.error) && (
          <p
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-700"
          >
            {error || query.error?.message}
          </p>
        )}
        {order && (
          <>
            <div className="rounded-xl bg-slate-50 p-4">
              <StatusBadge
                tone={order.status_code === "CANCELLED" ? "danger" : "info"}
              >
                {order.status_code.replaceAll("_", " ")}
              </StatusBadge>
              <dl className="mt-4 grid gap-4 text-sm sm:grid-cols-2">
                <div>
                  <dt className="text-slate-500">Client DO</dt>
                  <dd>{order.client_delivery_order_no}</dd>
                </div>
                <div>
                  <dt className="text-slate-500">Customer</dt>
                  <dd>
                    {order.customer_name} ({order.customer_code})
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">Owner / warehouse</dt>
                  <dd>
                    {order.owner_code} / {order.warehouse_code}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">Business date</dt>
                  <dd>{order.business_date}</dd>
                </div>
                <div>
                  <dt className="text-slate-500">Ship to (saved address)</dt>
                  <dd>
                    {order.ship_to_name} · {order.ship_to_address_1}{" "}
                    {order.ship_to_address_2} {order.ship_to_city}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">Requested ship time</dt>
                  <dd>
                    {order.requested_ship_at
                      ? new Date(order.requested_ship_at).toLocaleString()
                      : "—"}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">
                    Customer order / external reference
                  </dt>
                  <dd>
                    {order.customer_order_no ?? "—"} /{" "}
                    {order.external_reference ?? "—"}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">Notes</dt>
                  <dd>{order.notes ?? "—"}</dd>
                </div>
              </dl>
            </div>
            {!edit && !confirmation && (
              <div className="flex flex-wrap gap-3">
                {canPlan && order.status_code === "DRAFT" && (
                  <>
                    <Button
                      variant="secondary"
                      disabled={busy}
                      onClick={() => setEdit({ kind: "header" })}
                    >
                      Edit header
                    </Button>
                    <Button
                      variant="secondary"
                      disabled={busy}
                      onClick={() => setEdit({ kind: "line" })}
                    >
                      Add line
                    </Button>
                    <Button
                      disabled={busy}
                      onClick={() => {
                        setError("");
                        mutation.mutate({ kind: "validate" });
                      }}
                    >
                      Validate DO
                    </Button>
                  </>
                )}
                {canPlan && order.status_code === "VALIDATED" && (
                  <Button
                    disabled={busy}
                    onClick={() => setConfirmation({ kind: "release" })}
                  >
                    Release DO
                  </Button>
                )}
                {canPlan &&
                  [
                    "RELEASED",
                    "PARTIALLY_ALLOCATED",
                    "PARTIALLY_FULFILLED",
                  ].includes(order.status_code) && (
                    <Button
                      disabled={busy}
                      onClick={() => setConfirmation({ kind: "allocate" })}
                    >
                      Allocate available stock
                    </Button>
                  )}
                {canCloseShort &&
                  ["RELEASED", "PARTIALLY_FULFILLED"].includes(
                    order.status_code,
                  ) && (
                    <Button
                      variant="secondary"
                      disabled={busy}
                      onClick={() => {
                        setReason("");
                        setConfirmation({ kind: "close-short" });
                      }}
                    >
                      Close remaining demand short
                    </Button>
                  )}
                {canCancel &&
                  ["DRAFT", "VALIDATED", "RELEASED"].includes(
                    order.status_code,
                  ) && (
                    <Button
                      variant="secondary"
                      disabled={busy}
                      onClick={() => setConfirmation({ kind: "cancel" })}
                    >
                      Cancel DO
                    </Button>
                  )}
              </div>
            )}
            {validation && (
              <section className="rounded-xl border border-slate-200 p-4">
                <h3 className="font-bold">
                  Validation: {validation.status_code}
                </h3>
                <ul className="mt-3 space-y-2 text-sm">
                  {validation.results.map((result) => (
                    <li
                      key={result.rule_code}
                      className={
                        result.passed ? "text-emerald-800" : "text-rose-700"
                      }
                    >
                      {result.passed ? "Passed" : "Failed"} · {result.rule_name}
                      : {result.result_message}
                    </li>
                  ))}
                </ul>
              </section>
            )}
            {confirmation && (
              <section className="space-y-4 rounded-xl border border-amber-200 bg-amber-50 p-4">
                <p className="text-sm">
                  {confirmation.kind === "release"
                    ? "Release this validated DO for allocation? Draft quantities cannot be edited after release. No stock moves yet."
                    : confirmation.kind === "cancel"
                      ? "Cancel this Delivery Order? It remains in history and cannot be processed."
                      : confirmation.kind === "close-short"
                        ? "Permanently close ALL remaining uncommitted demand on this DO? Original quantities and shipment history are preserved. Reserved or in-process stock must be finished or released first. Delivery tracking for dispatched stock continues separately. A reason is required."
                        : confirmation.kind === "allocate"
                          ? "Reserve available stock using the configured picking strategy? Shortages stay open on this same DO for a later allocation and shipment. No physical stock moves yet."
                          : confirmation.kind === "delete"
                            ? `Delete draft line ${confirmation.line.line_no}?`
                            : "Confirm change?"}
                </p>
                {(confirmation.kind === "cancel" ||
                  confirmation.kind === "close-short") && (
                  <Textarea
                    aria-label={
                      confirmation.kind === "close-short"
                        ? "Close-short reason"
                        : "Cancellation reason"
                    }
                    placeholder="Required reason"
                    maxLength={4000}
                    value={reason}
                    onChange={(event) => setReason(event.target.value)}
                    disabled={busy}
                  />
                )}
                <div className="flex gap-3">
                  <Button
                    disabled={
                      busy ||
                      (["cancel", "close-short"].includes(confirmation.kind) &&
                        !reason.trim())
                    }
                    onClick={() => {
                      setError("");
                      mutation.mutate(confirmation);
                    }}
                  >
                    Confirm
                  </Button>
                  <Button
                    variant="secondary"
                    disabled={busy}
                    onClick={() => setConfirmation(undefined)}
                  >
                    Back
                  </Button>
                </div>
              </section>
            )}
            {edit && (
              <DraftEditor
                key={
                  edit.kind === "header"
                    ? "header"
                    : (edit.line?.outbound_line_id ?? "new")
                }
                order={order}
                edit={edit}
                busy={busy}
                onSave={(action) => {
                  setError("");
                  mutation.mutate(action);
                }}
                onBack={() => setEdit(undefined)}
              />
            )}
            {order.closed_short_reason && (
              <p className="rounded-xl bg-amber-50 p-4 text-sm text-amber-950">
                Closed short: {order.closed_short_reason} ·{" "}
                {order.closed_short_at
                  ? new Date(order.closed_short_at).toLocaleString()
                  : ""}
              </p>
            )}
            <section className="space-y-4">
              <h3 className="font-bold">
                Order lines ({order.lines?.length ?? 0})
              </h3>
              {order.lines?.map((line) => (
                <article
                  key={line.outbound_line_id}
                  className="space-y-3 rounded-xl border border-slate-200 p-4 text-sm"
                >
                  <div className="flex flex-wrap justify-between gap-3">
                    <h4 className="font-bold">
                      {line.line_no}. {line.item_name} ({line.item_code})
                    </h4>
                    {canPlan &&
                      order.status_code === "DRAFT" &&
                      !edit &&
                      !confirmation && (
                        <div className="flex gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={busy}
                            onClick={() => setEdit({ kind: "line", line })}
                          >
                            Edit line
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            disabled={busy || (order.lines?.length ?? 0) <= 1}
                            onClick={() =>
                              setConfirmation({ kind: "delete", line })
                            }
                          >
                            Delete line
                          </Button>
                        </div>
                      )}
                  </div>
                  <p className="rounded-lg bg-cyan-50 p-3 text-cyan-900">
                    Ordered {line.ordered_qty} {line.uom_code} ×{" "}
                    {line.uom_conversion_to_base} = {line.ordered_base_qty}{" "}
                    {line.base_uom_code}
                  </p>
                  <p className="text-slate-600">
                    Allocated {line.allocated_qty} · Picked {line.picked_qty} ·
                    Checked {line.checked_qty} · Packed {line.packed_qty} ·
                    Shipped {line.shipped_qty} ({line.base_uom_code})
                  </p>
                  {line.open_qty !== undefined && (
                    <p className="text-slate-600">
                      Open demand {line.open_qty} · Awaiting allocation{" "}
                      {line.unallocated_qty} · Closed short{" "}
                      {line.closed_short_qty ?? "0"} · Accepted checking
                      shortage {line.short_accepted_qty ?? "0"} (
                      {line.base_uom_code})
                    </p>
                  )}
                  <p className="text-slate-600">
                    Requested lot: {line.requested_lot_no ?? "—"} · Client line:{" "}
                    {line.customer_line_reference ?? "—"}
                  </p>
                  {line.notes && <p>{line.notes}</p>}
                </article>
              ))}
            </section>
          </>
        )}
      </div>
    </OperationDialog>
  );
}
