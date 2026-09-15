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

export const documentTypeFormSchema = z.object({
  code,
  name,
  module_code: z.string().min(1, "Module is required."),
  description: z.string(),
  is_active: z.boolean(),
});

export const documentStatusFormSchema = z.object({
  code,
  name,
  description: z.string(),
  display_order: z.coerce
    .number<number>()
    .int("Display order must be a whole number.")
    .min(0, "Display order cannot be negative.")
    .max(2147483647, "Display order is too large."),
  is_initial: z.boolean(),
  is_final: z.boolean(),
  is_cancelled: z.boolean(),
  is_active: z.boolean(),
});

export const documentTransitionFormSchema = z
  .object({
    from_status_id: z.string().min(1, "From status is required."),
    to_status_id: z.string().min(1, "To status is required."),
    required_permission_id: z.string(),
    is_active: z.boolean(),
  })
  .refine((value) => value.from_status_id !== value.to_status_id, {
    message: "From and to statuses must be different.",
    path: ["to_status_id"],
  });

export type DocumentTypeFormValues = z.infer<typeof documentTypeFormSchema>;
export type DocumentStatusFormValues = z.infer<typeof documentStatusFormSchema>;
export type DocumentTransitionFormValues = z.infer<
  typeof documentTransitionFormSchema
>;
