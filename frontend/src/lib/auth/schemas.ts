import { z } from "zod";

export const authenticatedUserSchema = z.object({
  account_id: z.string().min(1),
  username: z.string().min(1),
  email: z.string().email().optional(),
  display_name: z.string().min(1),
  preferred_timezone: z.string().optional(),
  last_login_at: z.string().optional(),
  permissions: z.array(z.string()),
});

export const loginResultSchema = z.object({
  token: z.string().min(1),
  token_type: z.literal("Bearer"),
  expires_at: z.string().datetime({ offset: true }),
  user: authenticatedUserSchema,
});

export type AuthenticatedUser = z.infer<typeof authenticatedUserSchema>;
export type LoginResult = z.infer<typeof loginResultSchema>;
