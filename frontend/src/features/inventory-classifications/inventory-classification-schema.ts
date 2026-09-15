import { z } from "zod";

export const inventoryStatusFormSchema = z.object({
  code: z
    .string()
    .trim()
    .min(1, "Code is required.")
    .max(40, "Code must contain at most 40 characters.")
    .regex(
      /^[A-Za-z0-9][A-Za-z0-9_-]*$/,
      "Use letters, numbers, underscores, or hyphens.",
    ),
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(100, "Name must contain at most 100 characters."),
  description: z.string(),
  is_allocatable: z.boolean(),
  is_pickable: z.boolean(),
  is_active: z.boolean(),
});

export type InventoryStatusFormValues = z.infer<
  typeof inventoryStatusFormSchema
>;

export const emptyInventoryStatusForm: InventoryStatusFormValues = {
  code: "",
  name: "",
  description: "",
  is_allocatable: false,
  is_pickable: false,
  is_active: true,
};
