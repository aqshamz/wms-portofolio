"use client";

import { useId, useState, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
  businessPartnerKeys,
  listBusinessPartners,
} from "@/features/business-partners/business-partner-api";
import {
  getItem,
  itemCatalogKeys,
  listItems,
  listUOMs,
} from "@/features/item-catalog/item-catalog-api";
import type { OutboundOrderLine } from "./outbound-order-types";
import {
  quantityPreview,
  type OutboundHeaderValues,
  type OutboundLineValues,
} from "./outbound-order-schema";

const active = {
  search: "",
  active: "active" as const,
  page: 1,
  pageSize: 100,
};
export function TextField({
  label,
  value,
  onChange,
  type = "text",
  required = false,
  maxLength = 120,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  required?: boolean;
  maxLength?: number;
}) {
  const id = useId();
  return (
    <FormField label={label} htmlFor={id} required={required}>
      <Input
        id={id}
        aria-label={label}
        value={value}
        type={type}
        required={required}
        maxLength={maxLength}
        onChange={(event) => onChange(event.target.value)}
      />
    </FormField>
  );
}
export function useOutboundPartners(ownerId: string, search: string) {
  const customersFilter = {
    ...active,
    search,
    ownerId,
    partnerTypeCode: "CUSTOMER",
  };
  const storesFilter = { ...customersFilter, partnerTypeCode: "STORE" };
  const customers = useQuery({
    queryKey: businessPartnerKeys.partnerList(customersFilter),
    queryFn: () => listBusinessPartners(customersFilter),
  });
  const stores = useQuery({
    queryKey: businessPartnerKeys.partnerList(storesFilter),
    queryFn: () => listBusinessPartners(storesFilter),
  });
  const shipTo = [
    ...new Map(
      [...(customers.data?.items ?? []), ...(stores.data?.items ?? [])]
        .filter((partner) => partner.address_line_1?.trim())
        .map((partner) => [partner.partner_id, partner]),
    ).values(),
  ];
  return {
    customers: customers.data?.items ?? [],
    shipTo,
    pending: customers.isPending || stores.isPending,
    error: customers.error || stores.error,
  };
}
export function OutboundHeaderEditor({
  value,
  onChange,
  shipToOptions,
  children,
}: {
  value: OutboundHeaderValues;
  onChange: (value: OutboundHeaderValues) => void;
  shipToOptions: { value: string; label: string }[];
  children?: ReactNode;
}) {
  const id = useId();
  function set<K extends keyof OutboundHeaderValues>(
    key: K,
    next: OutboundHeaderValues[K],
  ) {
    onChange({ ...value, [key]: next });
  }
  return (
    <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
      {children}
      <TextField
        label="Requested ship time (device time)"
        type="datetime-local"
        required
        value={value.requested_ship_at}
        onChange={(next) => set("requested_ship_at", next)}
      />
      <FormField label="Ship to" htmlFor={`${id}-ship`} required>
        <Select
          className="mt-2"
          id={`${id}-ship`}
          ariaLabel="Ship to"
          value={value.ship_to_partner_id}
          options={shipToOptions}
          placeholder="Customer or store with an address"
          onValueChange={(next) => set("ship_to_partner_id", next)}
        />
      </FormField>
      <TextField
        label="Customer order number"
        value={value.customer_order_no}
        onChange={(next) => set("customer_order_no", next)}
      />
      <TextField
        label="External reference"
        value={value.external_reference}
        onChange={(next) => set("external_reference", next)}
      />
      <div className="sm:col-span-2 lg:col-span-3">
        <FormField label="Document notes" htmlFor={`${id}-notes`}>
          <Textarea
            rows={3}
            className="min-h-0"
            id={`${id}-notes`}
            aria-label="Document notes"
            maxLength={4000}
            value={value.notes}
            onChange={(event) => set("notes", event.target.value)}
          />
        </FormField>
      </div>
    </div>
  );
}
export function OutboundLineEditor({
  ownerId,
  value,
  onChange,
  existing,
}: {
  ownerId: string;
  value: OutboundLineValues;
  onChange: (value: OutboundLineValues) => void;
  existing?: OutboundOrderLine;
}) {
  const id = useId();
  const [search, setSearch] = useState("");
  const filters = { ...active, ownerId, categoryId: "", search };
  const items = useQuery({
    queryKey: itemCatalogKeys.itemList(filters),
    queryFn: () => listItems(filters),
    enabled: !existing,
  });
  const detail = useQuery({
    queryKey: itemCatalogKeys.item(value.item_id || "none"),
    queryFn: () => getItem(value.item_id),
    enabled: Boolean(value.item_id),
  });
  const uoms = useQuery({
    queryKey: [...itemCatalogKeys.uomLists(), "all-active"],
    queryFn: async () => {
      const first = await listUOMs(active);
      const units = [...first.items];
      for (let page = 2; page <= first.total_pages; page++) {
        const next = await listUOMs({ ...active, page });
        units.push(...next.items);
      }
      return units;
    },
  });
  const item = detail.data;
  const activeUoms = new Map(
    (uoms.data ?? [])
      .filter((unit) => unit.is_active)
      .map((unit) => [unit.uom_id, unit]),
  );
  const options = (item?.uoms ?? [])
    .filter((mapping) => mapping.is_active && activeUoms.has(mapping.uom_id))
    .map((mapping) => ({
      value: mapping.uom_id,
      label: activeUoms.get(mapping.uom_id)!.code,
    }));
  if (
    item &&
    activeUoms.has(item.base_uom_id) &&
    !options.some((option) => option.value === item.base_uom_id)
  )
    options.unshift({
      value: item.base_uom_id,
      label: activeUoms.get(item.base_uom_id)!.code,
    });
  if (existing && !options.some((option) => option.value === existing.uom_id))
    options.unshift({
      value: existing.uom_id,
      label: `${existing.uom_code} (saved snapshot)`,
    });
  const sameUom = existing?.uom_id === value.uom_id;
  const factor = sameUom
    ? existing.uom_conversion_to_base
    : value.uom_id === item?.base_uom_id
      ? "1"
      : item?.uoms.find((mapping) => mapping.uom_id === value.uom_id)
          ?.conversion_to_base;
  const baseCode = sameUom
    ? existing.base_uom_code
    : activeUoms.get(item?.base_uom_id ?? "")?.code;
  const preview =
    factor && value.ordered_qty
      ? quantityPreview(
          value.ordered_qty,
          factor,
          Boolean(item?.serial_controlled),
        )
      : undefined;
  function set<K extends keyof OutboundLineValues>(
    key: K,
    next: OutboundLineValues[K],
  ) {
    onChange({ ...value, [key]: next });
  }
  return (
    <div className="space-y-4">
      {!existing && (
        <TextField label="Search items" value={search} onChange={setSearch} />
      )}
      {existing && (
        <p className="font-semibold">
          {existing.item_name} ({existing.item_code})
        </p>
      )}
      {(items.error || detail.error || uoms.error) && (
        <p role="alert" className="text-sm text-rose-700">
          Unable to load item or UOM choices. Try reopening this form.
        </p>
      )}
      <div className="grid gap-4 sm:grid-cols-2">
        {!existing && (
          <FormField label="Item" htmlFor={`${id}-item`} required>
            <Select
              className="mt-2"
              id={`${id}-item`}
              ariaLabel="Item"
              value={value.item_id}
              options={(items.data?.items ?? []).map((item) => ({
                value: item.item_id,
                label: `${item.name} (${item.code})`,
              }))}
              placeholder="Select item"
              onValueChange={(next) =>
                onChange({ ...value, item_id: next, uom_id: "" })
              }
            />
          </FormField>
        )}
        {existing && (
          <TextField
            label="Ordered quantity"
            required
            value={value.ordered_qty}
            maxLength={30}
            onChange={(next) => set("ordered_qty", next)}
          />
        )}
        <FormField label="Order UOM" htmlFor={`${id}-uom`} required>
          <Select
            className="mt-2"
            id={`${id}-uom`}
            ariaLabel="Order UOM"
            value={value.uom_id}
            options={options}
            placeholder={
              value.item_id ? "Base or configured UOM" : "Select an item first"
            }
            disabled={!value.item_id}
            onValueChange={(next) => set("uom_id", next)}
          />
        </FormField>
        {!existing && (
          <TextField
            label="Ordered quantity"
            required
            value={value.ordered_qty}
            maxLength={30}
            onChange={(next) => set("ordered_qty", next)}
          />
        )}
        <TextField
          label="Customer line reference"
          value={value.customer_line_reference}
          maxLength={100}
          onChange={(next) => set("customer_line_reference", next)}
        />
        <TextField
          label="Requested lot (optional)"
          value={value.requested_lot_no}
          maxLength={100}
          onChange={(next) => set("requested_lot_no", next)}
        />
      </div>
      {preview && (
        <p
          role={preview.error ? "alert" : "status"}
          className={`rounded-lg p-3 text-sm ${preview.error ? "bg-rose-50 text-rose-700" : "bg-cyan-50 text-cyan-900"}`}
        >
          {preview.error ??
            `${value.ordered_qty} × ${factor} = ${preview.base} ${baseCode ?? "base units"}${sameUom ? " · Saved conversion retained" : " · Conversion frozen when saved"}`}
        </p>
      )}
      <FormField label="Line notes" htmlFor={`${id}-notes`}>
        <Textarea
          rows={3}
          className="min-h-0"
          id={`${id}-notes`}
          aria-label="Line notes"
          maxLength={4000}
          value={value.notes}
          onChange={(event) => set("notes", event.target.value)}
        />
      </FormField>
    </div>
  );
}
