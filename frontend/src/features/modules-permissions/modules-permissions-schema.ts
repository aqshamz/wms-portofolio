import { z } from "zod";

export const appModuleFormSchema = z.object({
  code: z
    .string()
    .trim()
    .min(1, "Code is required.")
    .max(50, "Code must contain at most 50 characters.")
    .regex(
      /^[A-Za-z0-9][A-Za-z0-9_-]*$/,
      "Use letters, numbers, underscores, or hyphens.",
    ),
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(100, "Name must contain at most 100 characters."),
  display_order: z.coerce
    .number<number>()
    .int("Display order must be a whole number.")
    .min(0, "Display order cannot be negative.")
    .max(2_147_483_647, "Display order is too large."),
  is_active: z.boolean(),
});

export type AppModuleFormValues = z.infer<typeof appModuleFormSchema>;

export const emptyAppModuleForm: AppModuleFormValues = {
  code: "",
  name: "",
  display_order: 0,
  is_active: true,
};
