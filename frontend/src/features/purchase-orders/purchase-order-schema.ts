import { z } from "zod";

const decimal = z
  .string()
  .trim()
  .min(1, "Quantity is required.")
  .refine(
    (value) => /^\d+(\.\d{1,6})?$/.test(value) && Number(value) > 0,
    "Enter a quantity greater than zero with up to 6 decimals.",
  );

const optionalPercentage = z
  .string()
  .trim()
  .refine(
    (value) =>
      value === "" ||
      (/^\d+(\.\d{1,6})?$/.test(value) &&
        Number(value) >= 0 &&
        Number(value) <= 100),
    "Enter a percentage from 0 to 100.",
  );

export const purchaseOrderLineSchema = z.object({
  item_id: z.string().min(1, "Item is required."),
  ordered_qty: decimal,
  uom_id: z.string().min(1, "Receiving UOM is required."),
  vendor_item_code: z.string().max(100, "Maximum 100 characters."),
  expected_lot_no: z.string().max(100, "Maximum 100 characters."),
  expected_expiry_date: z.string(),
  notes: z.string().max(4000, "Maximum 4000 characters."),
  over_receipt_tolerance_pct: optionalPercentage,
  under_receipt_tolerance_pct: optionalPercentage,
});

export const purchaseOrderHeaderSchema = z
  .object({
    purchase_order_no: z
      .string()
      .trim()
      .min(1, "Purchase order number is required.")
      .max(120, "Maximum 120 characters."),
    ordered_at: z.string().min(1, "Ordered time is required."),
    expected_arrival_at: z.string(),
    notes: z.string().max(4000, "Maximum 4000 characters."),
  })
  .refine(
    (value) =>
      !value.expected_arrival_at ||
      new Date(value.expected_arrival_at) >= new Date(value.ordered_at),
    {
      path: ["expected_arrival_at"],
      message: "Expected arrival cannot be before the ordered time.",
    },
  );

export type PurchaseOrderHeaderValues = z.infer<
  typeof purchaseOrderHeaderSchema
>;

export const purchaseOrderFormSchema = z
  .object({
    owner_id: z.string().min(1, "Owner is required."),
    warehouse_id: z.string().min(1, "Warehouse is required."),
    vendor_id: z.string().min(1, "Supplier is required."),
    business_date: z.string().min(1, "Business date is required."),
    purchase_order_no: z
      .string()
      .trim()
      .min(1, "Purchase order number is required.")
      .max(120, "Maximum 120 characters."),
    ordered_at: z.string().min(1, "Ordered time is required."),
    expected_arrival_at: z.string(),
    notes: z.string().max(4000, "Maximum 4000 characters."),
    lines: z
      .array(purchaseOrderLineSchema)
      .min(1, "Add at least one purchase order line.")
      .max(500),
  })
  .refine(
    (value) =>
      !value.expected_arrival_at ||
      new Date(value.expected_arrival_at) >= new Date(value.ordered_at),
    {
      path: ["expected_arrival_at"],
      message: "Expected arrival cannot be before the ordered time.",
    },
  );

export type PurchaseOrderFormValues = z.infer<typeof purchaseOrderFormSchema>;

export const emptyPurchaseOrderLine: PurchaseOrderFormValues["lines"][number] =
  {
    item_id: "",
    ordered_qty: "",
    uom_id: "",
    vendor_item_code: "",
    expected_lot_no: "",
    expected_expiry_date: "",
    notes: "",
    over_receipt_tolerance_pct: "0",
    under_receipt_tolerance_pct: "0",
  };
