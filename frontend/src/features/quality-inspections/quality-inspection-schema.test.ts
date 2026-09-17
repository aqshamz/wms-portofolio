import { describe, expect, it } from "vitest";
import {
  cancelInspectionSchema,
  completeInspectionSchema,
  createInspectionSchema,
  quantityTotal,
} from "./quality-inspection-schema";

const values = {
  passed_qty: "8",
  failed_qty: "2",
  putaway_target_location_id: "storage-1",
  notes: "Checked",
};

describe("QC validation", () => {
  it("accepts full passes, failures and partial results in base units", () => {
    const schema = completeInspectionSchema("10.000000", false);
    expect(schema.safeParse(values).success).toBe(true);
    expect(
      schema.safeParse({ ...values, passed_qty: "10", failed_qty: "0" })
        .success,
    ).toBe(true);
    expect(
      schema.safeParse({
        ...values,
        passed_qty: "0",
        failed_qty: "10",
        putaway_target_location_id: "",
      }).success,
    ).toBe(true);
  });
  it.each(["1", "3"])(
    "reports a total mismatch for failed quantity %s",
    (failed_qty) => {
      const result = completeInspectionSchema("10", false).safeParse({
        ...values,
        failed_qty,
      });
      expect(result.success).toBe(false);
      if (!result.success)
        expect(result.error.issues).toContainEqual(
          expect.objectContaining({
            path: ["failed_qty"],
            message: expect.stringContaining("must equal 10"),
          }),
        );
    },
  );
  it("requires a storage target only when stock passes", () => {
    expect(
      completeInspectionSchema("10", false).safeParse({
        ...values,
        putaway_target_location_id: "",
      }).success,
    ).toBe(false);
  });
  it("cannot split serial or handling-unit batches", () => {
    const schema = completeInspectionSchema("10", true);
    expect(schema.safeParse(values).success).toBe(false);
    expect(
      schema.safeParse({ ...values, passed_qty: "10", failed_qty: "0" })
        .success,
    ).toBe(true);
  });
  it.each(["", "NaN", "-1", "1e2", "0.0000001", "100000000000000"])(
    "handles invalid quantity %s without throwing",
    (passed_qty) => {
      expect(
        completeInspectionSchema("10", false).safeParse({
          ...values,
          passed_qty,
        }).success,
      ).toBe(false);
      expect(quantityTotal(passed_qty, "0")).toBeUndefined();
    },
  );
  it("uses exact decimal addition", () => {
    expect(quantityTotal("0.1", "0.2")).toBe("0.3");
    expect(
      completeInspectionSchema("0.3", false).safeParse({
        ...values,
        passed_qty: "0.1",
        failed_qty: "0.2",
      }).success,
    ).toBe(true);
  });
  it("rejects zero inspection totals and empty reasons", () => {
    expect(
      completeInspectionSchema("0", false).safeParse({
        ...values,
        passed_qty: "0",
        failed_qty: "0",
      }).success,
    ).toBe(false);
    expect(cancelInspectionSchema.safeParse({ reason: "  " }).success).toBe(
      false,
    );
    expect(
      cancelInspectionSchema.parse({ reason: "  Wrong sampling plan  " })
        .reason,
    ).toBe("Wrong sampling plan");
  });
  it("requires a batch and bounds notes", () => {
    expect(
      createInspectionSchema.safeParse({ receipt_inventory_id: "", notes: "" })
        .success,
    ).toBe(false);
    expect(
      createInspectionSchema.safeParse({
        receipt_inventory_id: "batch",
        notes: "x".repeat(4001),
      }).success,
    ).toBe(false);
  });
});
