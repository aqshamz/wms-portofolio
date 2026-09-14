import { z } from "zod";

const code = (maximum: number) =>
  z
    .string()
    .trim()
    .min(1, "Code is required.")
    .max(maximum, `Code must be ${maximum} characters or fewer.`)
    .regex(
      /^[A-Za-z0-9][A-Za-z0-9_-]*$/,
      "Use letters, numbers, underscores, or hyphens.",
    );
const name = z
  .string()
  .trim()
  .min(1, "Name is required.")
  .max(100, "Name must be 100 characters or fewer.");
const optionalText = (maximum: number, label: string) =>
  z.string().trim().max(maximum, `${label} is too long.`);
const decimal = z
  .string()
  .trim()
  .refine(
    (value) =>
      value === "" || /^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$/.test(value),
    "Use a positive number with up to 6 decimal places.",
  );

export const locationTypeFormSchema = z.object({
  code: code(40),
  name,
  description: z.string().trim(),
  allows_receiving: z.boolean(),
  allows_storage: z.boolean(),
  allows_picking: z.boolean(),
  allows_shipping: z.boolean(),
  is_active: z.boolean(),
});

export const zoneFormSchema = z.object({
  code: code(40),
  name,
  description: z.string().trim(),
  is_active: z.boolean(),
});

export const locationFormSchema = z.object({
  zone_id: z.string().min(1, "Zone is required."),
  location_type_id: z.string().min(1, "Location type is required."),
  code: code(60),
  barcode: optionalText(100, "Barcode"),
  aisle: optionalText(20, "Aisle"),
  bay: optionalText(20, "Bay"),
  level_no: optionalText(20, "Level"),
  position_no: optionalText(20, "Position"),
  pick_sequence: z.number().int().min(0, "Pick sequence cannot be negative."),
  max_weight: decimal,
  max_volume: decimal,
  is_pick_face: z.boolean(),
  is_locked: z.boolean(),
  is_active: z.boolean(),
});

export type LocationTypeFormValues = z.infer<typeof locationTypeFormSchema>;
export type ZoneFormValues = z.infer<typeof zoneFormSchema>;
export type LocationFormValues = z.infer<typeof locationFormSchema>;

export const emptyLocationTypeForm: LocationTypeFormValues = {
  code: "",
  name: "",
  description: "",
  allows_receiving: false,
  allows_storage: false,
  allows_picking: false,
  allows_shipping: false,
  is_active: true,
};

export const emptyZoneForm: ZoneFormValues = {
  code: "",
  name: "",
  description: "",
  is_active: true,
};

export const emptyLocationForm: LocationFormValues = {
  zone_id: "",
  location_type_id: "",
  code: "",
  barcode: "",
  aisle: "",
  bay: "",
  level_no: "",
  position_no: "",
  pick_sequence: 0,
  max_weight: "",
  max_volume: "",
  is_pick_face: false,
  is_locked: false,
  is_active: true,
};
