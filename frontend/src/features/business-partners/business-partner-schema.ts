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

const optionalEmail = z
  .string()
  .trim()
  .max(254, "Email must contain at most 254 characters.")
  .refine(
    (value) => value === "" || z.email().safeParse(value).success,
    "Enter a valid email address.",
  );

export const partnerTypeFormSchema = z.object({
  code,
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(100, "Name must contain at most 100 characters."),
  description: z.string(),
  is_active: z.boolean(),
});

export type PartnerTypeFormValues = z.infer<typeof partnerTypeFormSchema>;

export const emptyPartnerTypeForm: PartnerTypeFormValues = {
  code: "",
  name: "",
  description: "",
  is_active: true,
};

export const businessPartnerFormSchema = z.object({
  code,
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(150, "Name must contain at most 150 characters."),
  legal_name: z.string().max(200, "Maximum 200 characters."),
  tax_number: z.string().max(100, "Maximum 100 characters."),
  email: optionalEmail,
  phone: z.string().max(50, "Maximum 50 characters."),
  address_line_1: z.string().max(255, "Maximum 255 characters."),
  address_line_2: z.string().max(255, "Maximum 255 characters."),
  city: z.string().max(100, "Maximum 100 characters."),
  province: z.string().max(100, "Maximum 100 characters."),
  postal_code: z.string().max(20, "Maximum 20 characters."),
  country_code: z
    .string()
    .refine(
      (value) => value === "" || /^[A-Za-z]{2}$/.test(value),
      "Use a 2-letter country code.",
    ),
  is_active: z.boolean(),
});

export type BusinessPartnerFormValues = z.infer<
  typeof businessPartnerFormSchema
>;

export const emptyBusinessPartnerForm: BusinessPartnerFormValues = {
  code: "",
  name: "",
  legal_name: "",
  tax_number: "",
  email: "",
  phone: "",
  address_line_1: "",
  address_line_2: "",
  city: "",
  province: "",
  postal_code: "",
  country_code: "ID",
  is_active: true,
};
