import { describe, expect, it } from "vitest";

import {
  emptyLocationForm,
  locationFormSchema,
  locationTypeFormSchema,
  zoneFormSchema,
} from "@/features/storage-layout/storage-layout-schema";

describe("storage layout schemas", () => {
  it("accepts valid location types and zones", () => {
    expect(
      locationTypeFormSchema.safeParse({
        code: "PICK_FACE",
        name: "Pick face",
        description: "Forward picking area",
        allows_receiving: false,
        allows_storage: true,
        allows_picking: true,
        allows_shipping: false,
        is_active: true,
      }).success,
    ).toBe(true);
    expect(
      zoneFormSchema.safeParse({
        code: "FAST_MOVERS",
        name: "Fast movers",
        description: "",
        is_active: true,
      }).success,
    ).toBe(true);
  });

  it("requires location relationships and validates capacities", () => {
    const result = locationFormSchema.safeParse({
      ...emptyLocationForm,
      code: "A-01-01",
      max_weight: "-10",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      const errors = result.error.flatten().fieldErrors;
      expect(errors.zone_id).toBeDefined();
      expect(errors.location_type_id).toBeDefined();
      expect(errors.max_weight).toBeDefined();
    }
  });
});
