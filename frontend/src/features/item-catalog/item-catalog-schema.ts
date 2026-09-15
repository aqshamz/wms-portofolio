import { z } from "zod";

const code = z
  .string()
  .trim()
  .min(1, "Code is required.")
  .regex(
    /^[A-Za-z0-9][A-Za-z0-9_-]*$/,
    "Use letters, numbers, underscores, or hyphens.",
  );

const optionalDecimal = z
  .string()
  .trim()
  .refine(
    (value) => value === "" || /^(?:\d+)(?:\.\d{1,6})?$/.test(value),
    "Use a non-negative number with up to 6 decimal places.",
  )
  .refine(
    (value) => value === "" || value.split(".")[0].length <= 14,
    "The value is too large.",
  );

const optionalDays = z
  .string()
  .trim()
  .refine(
    (value) => value === "" || /^\d+$/.test(value),
    "Use a whole number of days.",
  );

export const itemCategoryFormSchema = z.object({
  code: code.max(40, "Code must contain at most 40 characters."),
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(100, "Name must contain at most 100 characters."),
  parent_category_id: z.string(),
  is_active: z.boolean(),
});

export type ItemCategoryFormValues = z.infer<typeof itemCategoryFormSchema>;

export const emptyItemCategoryForm: ItemCategoryFormValues = {
  code: "",
  name: "",
  parent_category_id: "",
  is_active: true,
};

export const itemFormSchema = z
  .object({
    code: code.max(60, "Code must contain at most 60 characters."),
    name: z
      .string()
      .trim()
      .min(1, "Name is required.")
      .max(200, "Name must contain at most 200 characters."),
    description: z.string(),
    category_id: z.string(),
    base_uom_id: z.string().min(1, "Base UOM is required."),
    weight: optionalDecimal,
    volume: optionalDecimal,
    lot_controlled: z.boolean(),
    serial_controlled: z.boolean(),
    shelf_life_days: optionalDays,
    minimum_receive_days: optionalDays,
    is_active: z.boolean(),
  })
  .superRefine((values, context) => {
    if (
      values.shelf_life_days !== "" &&
      values.minimum_receive_days !== "" &&
      Number(values.minimum_receive_days) > Number(values.shelf_life_days)
    ) {
      context.addIssue({
        code: "custom",
        path: ["minimum_receive_days"],
        message: "Minimum receiving life cannot exceed shelf life.",
      });
    }
  });

export type ItemFormValues = z.infer<typeof itemFormSchema>;

export const emptyItemForm: ItemFormValues = {
  code: "",
  name: "",
  description: "",
  category_id: "",
  base_uom_id: "",
  weight: "",
  volume: "",
  lot_controlled: false,
  serial_controlled: false,
  shelf_life_days: "",
  minimum_receive_days: "",
  is_active: true,
};
