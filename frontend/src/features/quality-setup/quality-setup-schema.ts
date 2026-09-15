import { z } from "zod";

export const qualitySetupFormSchema = z.object({
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
  is_accepted: z.boolean(),
  is_active: z.boolean(),
});

export type QualitySetupFormValues = z.infer<typeof qualitySetupFormSchema>;

export const emptyQualitySetupForm: QualitySetupFormValues = {
  code: "",
  name: "",
  description: "",
  is_accepted: false,
  is_active: true,
};
