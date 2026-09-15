import { describe, expect, it } from "vitest";

import { pickingRuleFormSchema } from "@/features/picking-configuration/picking-configuration-schema";

describe("pickingRuleFormSchema", () => {
  it("accepts an ordered rule", () => {
    expect(
      pickingRuleFormSchema.safeParse({
        sequence_no: "1",
        inventory_status_id: "none",
        zone_id: "none",
        picking_sort_method_id: "method-1",
        is_active: true,
      }).success,
    ).toBe(true);
  });

  it("rejects a zero sequence", () => {
    expect(
      pickingRuleFormSchema.safeParse({
        sequence_no: 0,
        inventory_status_id: "none",
        zone_id: "none",
        picking_sort_method_id: "method-1",
        is_active: true,
      }).success,
    ).toBe(false);
  });
});
