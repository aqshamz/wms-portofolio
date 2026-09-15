import { describe, expect, it } from "vitest";

import {
  emptyItemForm,
  itemFormSchema,
} from "@/features/item-catalog/item-catalog-schema";

describe("itemFormSchema", () => {
  it("accepts valid item control settings", () => {
    expect(
      itemFormSchema.safeParse({
        ...emptyItemForm,
        code: "COFFEE_01",
        name: "Roasted coffee",
        base_uom_id: "uom-id",
        weight: "1.250000",
        shelf_life_days: "365",
        minimum_receive_days: "90",
      }).success,
    ).toBe(true);
  });

  it("rejects minimum receiving life above shelf life", () => {
    expect(
      itemFormSchema.safeParse({
        ...emptyItemForm,
        code: "COFFEE_01",
        name: "Roasted coffee",
        base_uom_id: "uom-id",
        shelf_life_days: "30",
        minimum_receive_days: "31",
      }).success,
    ).toBe(false);
  });
});
