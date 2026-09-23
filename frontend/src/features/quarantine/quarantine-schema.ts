import Decimal from "decimal.js";
import { z } from "zod";
import {
  dispositionEffect,
  remainingQuantity,
  type DispositionType,
  type QuarantineCase,
} from "./quarantine-types";

const quantityPattern = /^\d{1,14}(\.\d{1,6})?$/;
export function dispositionSchema(
  value: QuarantineCase,
  types: DispositionType[],
) {
  return z
    .object({
      disposition_type_code: z.string().min(1, "Select a disposition type."),
      disposition_qty: z
        .string()
        .trim()
        .max(30)
        .refine(
          (quantity) => quantityPattern.test(quantity),
          "Enter a positive quantity with up to 6 decimals.",
        ),
      business_date: z
        .string()
        .refine(
          (date) => z.iso.date().safeParse(date).success,
          "Enter a valid business date.",
        ),
      target_location_id: z.string(),
      client_decision_reference: z
        .string()
        .trim()
        .max(120, "Maximum 120 characters."),
      decision_notes: z.string().trim().max(4000, "Maximum 4000 characters."),
      work_instructions: z
        .string()
        .trim()
        .max(4000, "Maximum 4000 characters."),
    })
    .superRefine((fields, context) => {
      const type = types.find(
        (row) => row.code === fields.disposition_type_code && row.is_active,
      );
      const effect = dispositionEffect(type);
      const issue = (path: keyof typeof fields, message: string) =>
        context.addIssue({ code: "custom", path: [path], message });
      if (!effect)
        issue(
          "disposition_type_code",
          "Select an active, supported disposition type.",
        );
      if (effect === "accept" && !fields.target_location_id)
        issue(
          "target_location_id",
          "Select a strategy-eligible storage location.",
        );
      if (effect === "rework" && !fields.work_instructions)
        issue(
          "work_instructions",
          "Work instructions are required for rework.",
        );
      if (!quantityPattern.test(fields.disposition_qty)) return;
      const quantity = new Decimal(fields.disposition_qty);
      if (quantity.lte(0))
        issue("disposition_qty", "Quantity must be greater than zero.");
      if (quantity.gt(remainingQuantity(value)))
        issue(
          "disposition_qty",
          `Quantity exceeds the undecided amount (${remainingQuantity(value)}).`,
        );
      if (value.available_qty !== undefined && quantity.gt(value.available_qty))
        issue(
          "disposition_qty",
          `Quantity exceeds unreserved quarantine stock (${value.available_qty}).`,
        );
      if (
        value.is_indivisible &&
        value.available_qty !== undefined &&
        !quantity.eq(value.available_qty)
      )
        issue(
          "disposition_qty",
          "Serial-numbered or handling-unit stock must be processed in full; it cannot be split.",
        );
    });
}
export type DispositionValues = z.infer<ReturnType<typeof dispositionSchema>>;
