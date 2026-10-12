import Decimal from "decimal.js";
import { z } from "zod";

const decimal = z
  .string()
  .trim()
  .regex(
    /^(0|[1-9]\d{0,13})(\.\d{1,6})?$/,
    "Use a positive quantity with up to 6 decimals.",
  )
  .refine((value) => Number(value) > 0, "Quantity must be greater than zero.");
const ExactDecimal = Decimal.clone({ precision: 50 });
export const outboundLineSchema = z.object({
  item_id: z.string().uuid("Select an item."),
  ordered_qty: decimal,
  uom_id: z.string().uuid("Select an order UOM."),
  requested_lot_no: z.string().max(100),
  customer_line_reference: z.string().max(100),
  notes: z.string().max(4000),
});
export type OutboundLineValues = z.infer<typeof outboundLineSchema>;
export const emptyOutboundLine = (): OutboundLineValues => ({
  item_id: "",
  ordered_qty: "",
  uom_id: "",
  requested_lot_no: "",
  customer_line_reference: "",
  notes: "",
});
export const outboundHeaderSchema = z.object({
  ship_to_partner_id: z.string().uuid("Select a ship-to partner."),
  requested_ship_at: z
    .string()
    .min(1, "Requested ship time is required.")
    .refine(
      (value) => !Number.isNaN(new Date(value).getTime()),
      "Enter a valid ship time.",
    ),
  customer_order_no: z.string().max(120),
  external_reference: z.string().max(120),
  notes: z.string().max(4000),
});
export type OutboundHeaderValues = z.infer<typeof outboundHeaderSchema>;
export const outboundCreateSchema = outboundHeaderSchema.extend({
  customer_id: z.string().uuid("Select a customer."),
  client_delivery_order_no: z
    .string()
    .trim()
    .min(1, "Client DO number is required.")
    .max(120),
  business_date: z
    .string()
    .regex(/^\d{4}-\d{2}-\d{2}$/, "Business date is required."),
  lines: z.array(outboundLineSchema).min(1).max(500),
});
export type OutboundCreateValues = z.infer<typeof outboundCreateSchema>;
export const optionalText = (value: string) => value.trim() || undefined;
export function linePayload(value: OutboundLineValues) {
  return {
    ...value,
    requested_lot_no: optionalText(value.requested_lot_no),
    customer_line_reference: optionalText(value.customer_line_reference),
    notes: optionalText(value.notes),
  };
}
export function headerPayload(value: OutboundHeaderValues) {
  return {
    ship_to_partner_id: value.ship_to_partner_id,
    requested_ship_at: new Date(value.requested_ship_at).toISOString(),
    customer_order_no: optionalText(value.customer_order_no),
    external_reference: optionalText(value.external_reference),
    notes: optionalText(value.notes),
  };
}
export function localDateTime(value = new Date()) {
  return new Date(value.getTime() - value.getTimezoneOffset() * 60_000)
    .toISOString()
    .slice(0, 16);
}
export function quantityPreview(
  qty: string,
  factor: string,
  serialized: boolean,
) {
  try {
    qty = qty.trim();
    if (!/^(0|[1-9]\d{0,13})(\.\d{1,6})?$/.test(qty))
      return { error: "Enter a valid order quantity." };
    const base = new ExactDecimal(qty).times(factor);
    if (
      !base.isPositive() ||
      base.decimalPlaces() > 6 ||
      base.greaterThanOrEqualTo("100000000000000")
    )
      return { error: "Converted quantity exceeds numeric(20,6) precision." };
    if (serialized && !base.isInteger())
      return { error: "Serialized items require whole base units." };
    return { base: base.toFixed(6) };
  } catch {
    return { error: "Conversion is unavailable." };
  }
}
