import Decimal from "decimal.js";
import { z } from "zod";

// Matches the backend's numeric(20,6) quantity precision.
const quantityPattern = /^\d{1,14}(\.\d{1,6})?$/;
const quantity = z
  .string()
  .trim()
  .max(30)
  .refine(
    (value) => quantityPattern.test(value),
    "Enter zero or a positive quantity with up to 6 decimals.",
  );

export const createInspectionSchema = z.object({
  receipt_inventory_id: z.string().min(1, "Select a receipt batch."),
  notes: z.string().trim().max(4000, "Maximum 4000 characters."),
});

export function completeInspectionSchema(
  inspectedQty: string,
  indivisible: boolean,
) {
  return z
    .object({
      passed_qty: quantity,
      failed_qty: quantity,
      putaway_target_location_id: z.string(),
      notes: z.string().trim().max(4000, "Maximum 4000 characters."),
    })
    .superRefine((value, context) => {
      if (
        !quantityPattern.test(value.passed_qty) ||
        !quantityPattern.test(value.failed_qty)
      )
        return;
      const passed = new Decimal(value.passed_qty);
      const failed = new Decimal(value.failed_qty);
      if (
        !passed.plus(failed).eq(inspectedQty) ||
        passed.plus(failed).isZero()
      ) {
        context.addIssue({
          code: "custom",
          path: ["failed_qty"],
          message: `Passed + failed must equal ${new Decimal(inspectedQty).toString()} (the inspected quantity).`,
        });
      }
      if (indivisible && passed.gt(0) && failed.gt(0)) {
        context.addIssue({
          code: "custom",
          path: ["failed_qty"],
          message:
            "Serial-numbered or handling-unit batches must pass or fail in full; they cannot be split.",
        });
      }
      if (passed.gt(0) && !value.putaway_target_location_id) {
        context.addIssue({
          code: "custom",
          path: ["putaway_target_location_id"],
          message: "Select a storage location for passed stock.",
        });
      }
    });
}

export const cancelInspectionSchema = z.object({
  reason: z
    .string()
    .trim()
    .min(1, "A cancellation reason is required.")
    .max(4000, "Maximum 4000 characters."),
});
export type CompleteInspectionValues = z.infer<
  ReturnType<typeof completeInspectionSchema>
>;

export function quantityTotal(passed: string, failed: string) {
  if (!quantityPattern.test(passed) || !quantityPattern.test(failed))
    return undefined;
  return new Decimal(passed).plus(failed).toString();
}
