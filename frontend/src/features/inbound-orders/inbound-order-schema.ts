import { z } from "zod";

const positiveQuantity = z
  .string()
  .trim()
  .min(1, "Expected quantity is required.")
  .refine(
    (value) => /^\d+(\.\d{1,6})?$/.test(value) && Number(value) > 0,
    "Enter a quantity greater than zero with up to 6 decimals.",
  );

export const inboundOrderLineSchema = z.object({
  purchase_order_line_id: z.string().min(1, "Purchase Order line is required."),
  expected_qty: positiveQuantity,
  customer_line_reference: z.string().max(100, "Maximum 100 characters."),
  notes: z.string().max(4000, "Maximum 4000 characters."),
});

export const inboundOrderHeaderSchema = z.object({
  expected_arrival_at: z.string(),
  external_reference: z.string().max(120, "Maximum 120 characters."),
  supplier_reference: z.string().max(120, "Maximum 120 characters."),
  notes: z.string().max(4000, "Maximum 4000 characters."),
});

export const inboundOrderFormSchema = inboundOrderHeaderSchema.extend({
  purchase_order_id: z.string().min(1, "Purchase Order is required."),
  business_date: z.string().min(1, "Business date is required."),
  lines: z
    .array(inboundOrderLineSchema)
    .min(1, "Add at least one expected line.")
    .max(500),
});

export type InboundOrderLineValues = z.infer<typeof inboundOrderLineSchema>;
export type InboundOrderHeaderValues = z.infer<typeof inboundOrderHeaderSchema>;
export type InboundOrderFormValues = z.infer<typeof inboundOrderFormSchema>;

export const emptyInboundOrderLine: InboundOrderLineValues = {
  purchase_order_line_id: "",
  expected_qty: "",
  customer_line_reference: "",
  notes: "",
};
