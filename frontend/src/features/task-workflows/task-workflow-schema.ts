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

export const taskRecordFormSchema = z
  .object({
    code,
    name,
    description: z.string(),
    priority_value: z.coerce
      .number<number>()
      .int("Priority must be a whole number.")
      .min(0, "Priority cannot be negative.")
      .max(2147483647, "Priority is too large."),
    is_initial: z.boolean(),
    is_final: z.boolean(),
    is_cancelled: z.boolean(),
    is_active: z.boolean(),
  })
  .superRefine((value, context) => {
    if (value.is_cancelled && !value.is_final) {
      context.addIssue({
        code: "custom",
        message: "A cancelled status must also be final.",
        path: ["is_final"],
      });
    }
    if (
      value.is_initial &&
      (!value.is_active || value.is_final || value.is_cancelled)
    ) {
      context.addIssue({
        code: "custom",
        message: "An initial status must be active and non-terminal.",
        path: ["is_initial"],
      });
    }
  });

export const taskTransitionFormSchema = z
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

export type TaskRecordFormValues = z.infer<typeof taskRecordFormSchema>;
export type TaskTransitionFormValues = z.infer<typeof taskTransitionFormSchema>;
