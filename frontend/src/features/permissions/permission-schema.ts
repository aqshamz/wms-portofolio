import { z } from "zod";

export const roleFormSchema = z.object({
  code: z
    .string()
    .trim()
    .min(2, "Code must contain at least 2 characters.")
    .max(60, "Code must contain at most 60 characters.")
    .regex(
      /^[A-Za-z0-9][A-Za-z0-9_-]*$/,
      "Use letters, numbers, underscores, or hyphens.",
    ),
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(120, "Name must contain at most 120 characters."),
  description: z.string(),
  permission_ids: z.array(z.string()),
});

export type RoleFormValues = z.infer<typeof roleFormSchema>;

export const emptyRoleForm: RoleFormValues = {
  code: "",
  name: "",
  description: "",
  permission_ids: [],
};
