import Decimal from "decimal.js";
import { z } from "zod";

const quantityPattern = /^\d+(\.\d{1,6})?$/;

const positiveQuantity = z
  .string()
  .trim()
  .refine(
    (value) => quantityPattern.test(value) && new Decimal(value).gt(0),
    "Enter a quantity greater than zero with up to 6 decimals.",
  );

const nonNegativeQuantity = z
  .string()
  .trim()
  .refine(
    (value) => quantityPattern.test(value),
    "Enter zero or a positive quantity with up to 6 decimals.",
  );

export const receiptBatchSchema = z.object({
  source_qty: positiveQuantity,
  received_location_id: z.string().min(1, "Received location is required."),
  lot_number: z.string().max(100, "Maximum 100 characters."),
  manufacture_date: z.string(),
  expiry_date: z.string(),
  handling_unit_id: z.string().max(120, "Maximum 120 characters."),
  serial_no: z.string().max(120, "Maximum 120 characters."),
});

export const receiptLineSchema = z
  .object({
    inbound_line_id: z.string().min(1),
    item_id: z.string().min(1),
    received_qty: positiveQuantity,
    rejected_qty: nonNegativeQuantity,
    exception_type_code: z.enum([
      "NONE",
      "REJECTED_AT_DOCK",
      "DAMAGED",
      "WRONG_ITEM",
    ]),
    exception_notes: z.string().max(4000, "Maximum 4000 characters."),
    lot_controlled: z.boolean(),
    serial_controlled: z.boolean(),
    batches: z.array(receiptBatchSchema).max(1000),
  })
  .superRefine((value, context) => {
    // Field refinements report malformed quantities; do not let Decimal throw
    // before the form can render those validation errors.
    if (
      !quantityPattern.test(value.received_qty) ||
      !quantityPattern.test(value.rejected_qty) ||
      value.batches.some((batch) => !quantityPattern.test(batch.source_qty))
    )
      return;
    const received = new Decimal(value.received_qty);
    const rejected = new Decimal(value.rejected_qty);
    if (rejected.gt(received)) {
      context.addIssue({
        code: "custom",
        path: ["rejected_qty"],
        message: "Rejected quantity cannot exceed received quantity.",
      });
      return;
    }
    const accepted = received.minus(rejected);
    const batched = value.batches.reduce(
      (total, batch) => total.plus(batch.source_qty || 0),
      new Decimal(0),
    );
    if (!batched.eq(accepted)) {
      context.addIssue({
        code: "custom",
        path: ["batches"],
        message: "Batch quantities must equal the accepted quantity.",
      });
    }
    if (rejected.gt(0) && !value.exception_notes.trim()) {
      context.addIssue({
        code: "custom",
        path: ["exception_notes"],
        message: "Exception notes are required for rejected quantity.",
      });
    }
    value.batches.forEach((batch, index) => {
      if (value.lot_controlled && !batch.lot_number.trim()) {
        context.addIssue({
          code: "custom",
          path: ["batches", index, "lot_number"],
          message: "Lot number is required for this item.",
        });
      }
      if (!value.lot_controlled && batch.lot_number.trim()) {
        context.addIssue({
          code: "custom",
          path: ["batches", index, "lot_number"],
          message: "Lot data is not allowed for this item.",
        });
      }
      if (value.serial_controlled) {
        if (!batch.serial_no.trim()) {
          context.addIssue({
            code: "custom",
            path: ["batches", index, "serial_no"],
            message: "Serial number is required for this item.",
          });
        }
      } else if (batch.serial_no.trim()) {
        context.addIssue({
          code: "custom",
          path: ["batches", index, "serial_no"],
          message: "Serial number is not allowed for this item.",
        });
      }
    });
  });

export const receiptFormSchema = z.object({
  inbound_id: z.string().min(1, "Inbound Order is required."),
  business_date: z.string().min(1, "Business date is required."),
  received_at: z.string().min(1, "Received time is required."),
  dock_location_id: z.string().min(1, "Receiving dock is required."),
  vehicle_number: z.string().max(60, "Maximum 60 characters."),
  seal_number: z.string().max(60, "Maximum 60 characters."),
  delivery_note_no: z.string().max(100, "Maximum 100 characters."),
  notes: z.string().max(4000, "Maximum 4000 characters."),
  lines: z.array(receiptLineSchema).min(1, "Add at least one receipt line."),
});

export type ReceiptFormValues = z.infer<typeof receiptFormSchema>;

export const emptyReceiptBatch: ReceiptFormValues["lines"][number]["batches"][number] =
  {
    source_qty: "",
    received_location_id: "",
    lot_number: "",
    manufacture_date: "",
    expiry_date: "",
    handling_unit_id: "",
    serial_no: "",
  };
