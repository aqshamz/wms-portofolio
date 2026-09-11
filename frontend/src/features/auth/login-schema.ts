import { z } from "zod";

export const loginSchema = z.object({
  identifier: z
    .string()
    .trim()
    .min(1, "Enter your username or email.")
    .max(254, "Username or email is too long."),
  password: z
    .string()
    .min(1, "Enter your password.")
    .max(1024, "Password is too long."),
});

export type LoginInput = z.infer<typeof loginSchema>;
