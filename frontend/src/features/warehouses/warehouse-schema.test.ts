import { describe, expect, it } from "vitest";

import {
  emptyWarehouseForm,
  warehouseFormSchema,
} from "@/features/warehouses/warehouse-schema";

describe("warehouseFormSchema", () => {
  it("accepts a valid warehouse", () => {
    expect(
      warehouseFormSchema.safeParse({
        ...emptyWarehouseForm,
        operator_id: "30aa3c17-9cb4-4b2f-9e91-e531c4156305",
        code: "JKT_DC",
        name: "Jakarta Distribution Center",
      }).success,
    ).toBe(true);
  });

  it("requires an operator and a valid country code", () => {
    const result = warehouseFormSchema.safeParse({
      ...emptyWarehouseForm,
      code: "JKT",
      name: "Jakarta",
      country_code: "IND",
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.flatten().fieldErrors.operator_id).toBeDefined();
      expect(result.error.flatten().fieldErrors.country_code).toBeDefined();
    }
  });
});
