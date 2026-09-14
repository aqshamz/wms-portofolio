import { z } from "zod";

const username = z
  .string()
  .trim()
  .min(3, "Username must contain at least 3 characters.")
  .max(100, "Username must contain at most 100 characters.")
  .regex(
    /^[A-Za-z0-9][A-Za-z0-9._-]*$/,
    "Use letters, numbers, dots, underscores, or hyphens.",
  );

const optionalEmail = z
  .string()
  .trim()
  .max(254, "Email must contain at most 254 characters.")
  .refine(
    (value) => value === "" || z.email().safeParse(value).success,
    "Enter a valid email address.",
  );

const profileFields = {
  username,
  email: optionalEmail,
  display_name: z
    .string()
    .trim()
    .min(1, "Display name is required.")
    .max(150, "Display name must contain at most 150 characters."),
  authentication_policy_id: z.string(),
  preferred_timezone: z
    .string()
    .trim()
    .max(50, "Timezone must contain at most 50 characters."),
};

export const accountFormSchema = z.discriminatedUnion("mode", [
  z.object({
    mode: z.literal("create"),
    ...profileFields,
    password: z
      .string()
      .min(12, "Password must contain at least 12 characters.")
      .max(72, "Password must contain at most 72 characters."),
    account_status_id: z.string(),
  }),
  z.object({
    mode: z.literal("edit"),
    ...profileFields,
    password: z.literal(""),
    account_status_id: z.string(),
  }),
]);

export type AccountFormValues = z.infer<typeof accountFormSchema>;

export const emptyAccountForm: AccountFormValues = {
  mode: "create",
  username: "",
  email: "",
  display_name: "",
  password: "",
  account_status_id: "",
  authentication_policy_id: "",
  preferred_timezone: "Asia/Jakarta",
};

export const passwordResetSchema = z
  .object({
    password: z
      .string()
      .min(12, "Password must contain at least 12 characters.")
      .max(72, "Password must contain at most 72 characters."),
    confirmation: z.string(),
  })
  .refine((value) => value.password === value.confirmation, {
    message: "Passwords do not match.",
    path: ["confirmation"],
  });

export type PasswordResetValues = z.infer<typeof passwordResetSchema>;
