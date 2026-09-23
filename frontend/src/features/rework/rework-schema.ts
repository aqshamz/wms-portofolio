import { z } from "zod";

export const reworkResultSchema = z.object({
  result_notes: z.string().trim().max(4000, "Maximum 4000 characters."),
});
export type ReworkResultValues = z.infer<typeof reworkResultSchema>;
