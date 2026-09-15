import { describe, expect, it } from "vitest";

import {
  emptyItemUOMForm,
  itemUOMFormSchema,
} from "@/features/units-packaging/units-packaging-schema";

describe("itemUOMFormSchema", () => {
  it("accepts a positive conversion", () => {
    expect(
      itemUOMFormSchema.safeParse({
        ...emptyItemUOMForm,
        uom_id: "uom-id",
        conversion_to_base: "12.5",
      }).success,
    ).toBe(true);
  });

  it("rejects zero conversion", () => {
    expect(
      itemUOMFormSchema.safeParse({
        ...emptyItemUOMForm,
        uom_id: "uom-id",
        conversion_to_base: "0",
      }).success,
    ).toBe(false);
  });
});
