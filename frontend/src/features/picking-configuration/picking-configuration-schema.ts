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

export const pickingRecordFormSchema = z.object({
  code,
  name,
  description: z.string(),
  owner_id: z.string(),
  warehouse_id: z.string(),
  is_active: z.boolean(),
});

export const pickingRuleFormSchema = z.object({
  sequence_no: z.coerce
    .number<number>()
    .int("Sequence must be a whole number.")
    .min(1, "Sequence starts at 1.")
    .max(2147483647, "Sequence is too large."),
  inventory_status_id: z.string(),
  zone_id: z.string(),
  picking_sort_method_id: z.string().min(1, "Sort method is required."),
  is_active: z.boolean(),
});

export type PickingRecordFormValues = z.infer<typeof pickingRecordFormSchema>;
export type PickingRuleFormValues = z.infer<typeof pickingRuleFormSchema>;
