import { describe, expect, it } from "vitest";
import { dispositionSchema } from "./quarantine-schema";
import {
  canDecide,
  dispositionEffect,
  remainingQuantity,
} from "./quarantine-types";
import { testCase, testTypes } from "./quarantine-test-fixtures";
const fields = {
  disposition_type_code: "RETURN",
  disposition_qty: "2",
  business_date: "2026-09-17",
  target_location_id: "",
  client_decision_reference: "",
  decision_notes: "",
  work_instructions: "",
};
describe("quarantine validation", () => {
  it.each(["0", "-1", "11", "1.0000001", "1e1", "abc", "100000000000000"])(
    "rejects invalid or excessive quantity %s",
    (quantity) =>
      expect(
        dispositionSchema(testCase, testTypes).safeParse({
          ...fields,
          disposition_qty: quantity,
        }).success,
      ).toBe(false),
  );
  it("accepts partial decisions and exact decimal limits", () => {
    expect(
      dispositionSchema(testCase, testTypes).safeParse(fields).success,
    ).toBe(true);
    const value = {
      ...testCase,
      quarantine_qty: "0.3",
      disposed_qty: "0.1",
      available_qty: "0.2",
    };
    expect(remainingQuantity(value)).toBe("0.2");
    expect(
      dispositionSchema(value, testTypes).safeParse({
        ...fields,
        disposition_qty: "0.2",
      }).success,
    ).toBe(true);
    expect(
      dispositionSchema(value, testTypes).safeParse({
        ...fields,
        disposition_qty: "0.200001",
      }).success,
    ).toBe(false);
  });
  it("checks both undecided and unreserved stock", () => {
    expect(
      dispositionSchema(
        { ...testCase, disposed_qty: "9" },
        testTypes,
      ).safeParse(fields).success,
    ).toBe(false);
    expect(
      dispositionSchema(
        { ...testCase, available_qty: "1" },
        testTypes,
      ).safeParse(fields).success,
    ).toBe(false);
  });
  it("requires acceptance targets and rework instructions", () => {
    expect(
      dispositionSchema(testCase, testTypes).safeParse({
        ...fields,
        disposition_type_code: "ACCEPT",
      }).success,
    ).toBe(false);
    expect(
      dispositionSchema(testCase, testTypes).safeParse({
        ...fields,
        disposition_type_code: "ACCEPT",
        target_location_id: "target",
      }).success,
    ).toBe(true);
    expect(
      dispositionSchema(testCase, testTypes).safeParse({
        ...fields,
        disposition_type_code: "REWORK",
        work_instructions: "   ",
      }).success,
    ).toBe(false);
    expect(
      dispositionSchema(testCase, testTypes).safeParse({
        ...fields,
        disposition_type_code: "REWORK",
        work_instructions: "Replace seal",
      }).success,
    ).toBe(true);
  });
  it("blocks indivisible stock splitting", () => {
    const schema = dispositionSchema(
      { ...testCase, is_indivisible: true },
      testTypes,
    );
    expect(schema.safeParse(fields).success).toBe(false);
    expect(schema.safeParse({ ...fields, disposition_qty: "10" }).success).toBe(
      true,
    );
  });
  it("rejects inactive/unsupported types and invalid business dates", () => {
    expect(
      dispositionSchema(
        testCase,
        testTypes.map((type) => ({ ...type, is_active: false })),
      ).safeParse(fields).success,
    ).toBe(false);
    expect(
      dispositionSchema(testCase, testTypes).safeParse({
        ...fields,
        business_date: "2026-02-30",
      }).success,
    ).toBe(false);
  });
  it("uses the backend type flags rather than code guesses", () => {
    expect(dispositionEffect({ ...testTypes[0], code: "CUSTOM" })).toBe(
      "accept",
    );
    expect(
      dispositionEffect({ ...testTypes[0], requires_reinspection: true }),
    ).toBe("rework");
  });
  it("checks state and exact decision permission", () => {
    expect(
      canDecide(testCase, { canDispose: false, timezone: "Asia/Jakarta" }),
    ).toBe(false);
    expect(
      canDecide(testCase, { canDispose: true, timezone: "Asia/Jakarta" }),
    ).toBe(true);
    expect(
      canDecide(
        { ...testCase, status_code: "PARTIALLY_DECIDED", disposed_qty: "2" },
        { canDispose: true, timezone: "Asia/Jakarta" },
      ),
    ).toBe(true);
    expect(
      canDecide(
        { ...testCase, status_code: "CLOSED" },
        { canDispose: true, timezone: "Asia/Jakarta" },
      ),
    ).toBe(false);
  });
});
