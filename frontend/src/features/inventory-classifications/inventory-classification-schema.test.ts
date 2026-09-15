import { describe, expect, it } from "vitest";

import {
  emptyInventoryStatusForm,
  inventoryStatusFormSchema,
} from "@/features/inventory-classifications/inventory-classification-schema";

describe("inventoryStatusFormSchema", () => {
  it("accepts a valid classification", () => {
    expect(
      inventoryStatusFormSchema.safeParse({
        ...emptyInventoryStatusForm,
        code: "AVAILABLE",
        name: "Available",
        is_allocatable: true,
        is_pickable: true,
      }).success,
    ).toBe(true);
  });

  it("rejects malformed codes", () => {
    expect(
      inventoryStatusFormSchema.safeParse({
        ...emptyInventoryStatusForm,
        code: "NOT AVAILABLE",
        name: "Not available",
      }).success,
    ).toBe(false);
  });
});
