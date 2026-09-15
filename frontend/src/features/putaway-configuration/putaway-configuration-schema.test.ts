import { describe, expect, it } from "vitest";

import { putawayRuleFormSchema } from "@/features/putaway-configuration/putaway-configuration-schema";

describe("putawayRuleFormSchema", () => {
  it("accepts a capacity-aware rule", () => {
    expect(
      putawayRuleFormSchema.safeParse({
        sequence_no: "1",
        category_id: "none",
        location_type_id: "location-type-1",
        zone_id: "none",
        minimum_empty_percent: "25.5000",
        is_active: true,
      }).success,
    ).toBe(true);
  });

  it("rejects capacity above 100 percent", () => {
    expect(
      putawayRuleFormSchema.safeParse({
        sequence_no: 1,
        category_id: "none",
        location_type_id: "none",
        zone_id: "none",
        minimum_empty_percent: "100.0001",
        is_active: true,
      }).success,
    ).toBe(false);
  });
});
