import { z } from "zod";

const isoDate = z.string().regex(/^\d{4}-\d{2}-\d{2}$/, "Choose a valid date.");

export const numberRuleFormSchema = z.object({
  prefix: z
    .string()
    .trim()
    .min(1, "Prefix is required.")
    .max(20, "Prefix must contain at most 20 characters.")
    .regex(
      /^[A-Za-z0-9][A-Za-z0-9_-]*$/,
      "Use letters, numbers, underscores, or hyphens.",
    ),
  separator: z
    .string()
    .max(3, "Separator must contain at most 3 characters.")
    .regex(/^[./_-]{0,3}$/, "Use dots, slashes, underscores, or hyphens."),
  sequence_length: z.coerce
    .number<number>()
    .int("Sequence length must be a whole number.")
    .min(3, "Use at least 3 digits.")
    .max(18, "Use at most 18 digits."),
  effective_from: isoDate,
  include_partner_code: z.boolean(),
  include_warehouse_code: z.boolean(),
});

export const allocationFormSchema = z.object({
  business_date: isoDate,
  partner_id: z.string(),
  warehouse_id: z.string(),
});

export type NumberRuleFormValues = z.infer<typeof numberRuleFormSchema>;
export type AllocationFormValues = z.infer<typeof allocationFormSchema>;
