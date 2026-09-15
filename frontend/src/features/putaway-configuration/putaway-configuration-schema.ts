import { z } from "zod";

const code = z
  .string()
  .trim()
  .min(1, "Code is required.")
  .max(40, "Code must contain at most 40 characters.")
  .regex(
    /^[A-Za-z0-9][A-Za-z0-9_-]*$/,
    "Use letters, numbers, underscores, or hyphens.",
  );
const name = z
  .string()
  .trim()
  .min(1, "Name is required.")
  .max(100, "Name must contain at most 100 characters.");

export const putawayStrategyFormSchema = z.object({
  code,
  name,
  description: z.string(),
  owner_id: z.string(),
  warehouse_id: z.string(),
  is_active: z.boolean(),
});

export const putawayRuleFormSchema = z.object({
  sequence_no: z.coerce
    .number<number>()
    .int("Sequence must be a whole number.")
    .min(1, "Sequence starts at 1.")
    .max(2147483647, "Sequence is too large."),
  category_id: z.string(),
  location_type_id: z.string(),
  zone_id: z.string(),
  minimum_empty_percent: z
    .string()
    .trim()
    .refine(
      (value) =>
        value === "" || /^(0|[1-9][0-9]{0,2})(\.[0-9]{1,4})?$/.test(value),
      "Enter a percentage with up to 4 decimal places.",
    )
    .refine(
      (value) => value === "" || Number(value) <= 100,
      "Percentage cannot exceed 100.",
    ),
  is_active: z.boolean(),
});

export type PutawayStrategyFormValues = z.infer<
  typeof putawayStrategyFormSchema
>;
export type PutawayRuleFormValues = z.infer<typeof putawayRuleFormSchema>;
