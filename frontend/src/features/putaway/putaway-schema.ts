import { z } from "zod";
import type { PutawayAction } from "./putaway-types";

export function putawayActionSchema(action: PutawayAction) {
  return z
    .object({
      business_date: z.string(),
      reason: z.string().trim().max(4000, "Maximum 4000 characters."),
      account_id: z.string(),
      target_location_id: z.string(),
    })
    .superRefine((value, context) => {
      if (
        ["complete", "cancel", "reverse"].includes(action) &&
        !z.iso.date().safeParse(value.business_date).success
      )
        context.addIssue({
          code: "custom",
          path: ["business_date"],
          message: "Enter a valid business date.",
        });
      if ((action === "cancel" || action === "reverse") && !value.reason)
        context.addIssue({
          code: "custom",
          path: ["reason"],
          message: "A reason is required.",
        });
      if (action === "assign" && !value.account_id)
        context.addIssue({
          code: "custom",
          path: ["account_id"],
          message: "Select an account.",
        });
      if (action === "retarget" && !value.target_location_id)
        context.addIssue({
          code: "custom",
          path: ["target_location_id"],
          message: "Select an eligible target location.",
        });
    });
}
export type PutawayActionValues = z.infer<
  ReturnType<typeof putawayActionSchema>
>;
export function businessDateToday(timezone: string, now = new Date()) {
  // An old/invalid account preference should not prevent opening a task form.
  try {
    new Intl.DateTimeFormat("en-CA", { timeZone: timezone });
  } catch {
    timezone = "Asia/Jakarta";
  }
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(now);
  const part = (name: string) =>
    parts.find((value) => value.type === name)?.value;
  return `${part("year")}-${part("month")}-${part("day")}`;
}
