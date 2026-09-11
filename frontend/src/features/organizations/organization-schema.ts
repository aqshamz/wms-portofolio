import { z } from "zod";

const optionalField = (maximum: number, label: string) =>
  z
    .string()
    .trim()
    .max(maximum, `${label} must be ${maximum} characters or fewer.`);

export const organizationFormSchema = z.object({
  code: z
    .string()
    .trim()
    .min(1, "Code is required.")
    .max(40, "Code must be 40 characters or fewer.")
    .regex(
      /^[A-Za-z0-9][A-Za-z0-9_-]*$/,
      "Use letters, numbers, underscores, or hyphens.",
    ),
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(150, "Name must be 150 characters or fewer."),
  legal_name: optionalField(200, "Legal name"),
  tax_number: optionalField(100, "Tax number"),
  timezone_name: z
    .string()
    .trim()
    .min(1, "Timezone is required.")
    .max(50, "Timezone must be 50 characters or fewer."),
  address_line_1: optionalField(255, "Address"),
  address_line_2: optionalField(255, "Address"),
  city: optionalField(100, "City"),
  province: optionalField(100, "Province"),
  postal_code: optionalField(20, "Postal code"),
  country_code: z
    .string()
    .trim()
    .refine(
      (value) => value === "" || /^[A-Za-z]{2}$/.test(value),
      "Use a two-letter country code, such as ID.",
    ),
  is_active: z.boolean(),
});

export type OrganizationFormValues = z.infer<typeof organizationFormSchema>;

export const emptyOrganizationForm: OrganizationFormValues = {
  code: "",
  name: "",
  legal_name: "",
  tax_number: "",
  timezone_name: "Asia/Jakarta",
  address_line_1: "",
  address_line_2: "",
  city: "",
  province: "",
  postal_code: "",
  country_code: "ID",
  is_active: true,
};
