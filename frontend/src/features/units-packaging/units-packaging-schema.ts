import { z } from "zod";

const code = z
  .string()
  .trim()
  .min(1, "Code is required.")
  .regex(
    /^[A-Za-z0-9][A-Za-z0-9_-]*$/,
    "Use letters, numbers, underscores, or hyphens.",
  );

const decimal = (positive: boolean) =>
  z
    .string()
    .trim()
    .refine(
      (value) => value === "" || /^(?:\d+)(?:\.\d{1,6})?$/.test(value),
      "Use a number with up to 6 decimal places.",
    )
    .refine(
      (value) => value === "" || value.split(".")[0].length <= 14,
      "The value is too large.",
    )
    .refine(
      (value) => !positive || value === "" || Number(value) > 0,
      "The value must be greater than zero.",
    );

export const uomFormSchema = z.object({
  code: code.max(20, "Code must contain at most 20 characters."),
  name: z.string().trim().min(1, "Name is required.").max(100),
  decimal_scale: z.number().int().min(0).max(6),
  is_active: z.boolean(),
});

export type UOMFormValues = z.infer<typeof uomFormSchema>;
export const emptyUOMForm: UOMFormValues = {
  code: "",
  name: "",
  decimal_scale: 0,
  is_active: true,
};

export const handlingUnitFormSchema = z.object({
  code: code.max(40, "Code must contain at most 40 characters."),
  name: z.string().trim().min(1, "Name is required.").max(100),
  max_weight: decimal(false),
  max_volume: decimal(false),
  is_active: z.boolean(),
});

export type HandlingUnitFormValues = z.infer<typeof handlingUnitFormSchema>;
export const emptyHandlingUnitForm: HandlingUnitFormValues = {
  code: "",
  name: "",
  max_weight: "",
  max_volume: "",
  is_active: true,
};

export const itemUOMFormSchema = z.object({
  uom_id: z.string().min(1, "UOM is required."),
  conversion_to_base: decimal(true).refine(
    (value) => value !== "",
    "Conversion is required.",
  ),
  length: decimal(false),
  width: decimal(false),
  height: decimal(false),
  weight: decimal(false),
  is_receiving_uom: z.boolean(),
  is_picking_uom: z.boolean(),
  is_active: z.boolean(),
});

export type ItemUOMFormValues = z.infer<typeof itemUOMFormSchema>;
export const emptyItemUOMForm: ItemUOMFormValues = {
  uom_id: "",
  conversion_to_base: "",
  length: "",
  width: "",
  height: "",
  weight: "",
  is_receiving_uom: true,
  is_picking_uom: true,
  is_active: true,
};

export const barcodeFormSchema = z.object({
  uom_id: z.string(),
  barcode: z
    .string()
    .trim()
    .min(1, "Barcode is required.")
    .max(100, "Barcode must contain at most 100 characters."),
  is_primary: z.boolean(),
  is_active: z.boolean(),
});

export type BarcodeFormValues = z.infer<typeof barcodeFormSchema>;
export const emptyBarcodeForm: BarcodeFormValues = {
  uom_id: "",
  barcode: "",
  is_primary: false,
  is_active: true,
};
