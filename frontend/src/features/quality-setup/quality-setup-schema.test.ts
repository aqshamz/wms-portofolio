import { describe, expect, it } from "vitest";

import {
  emptyQualitySetupForm,
  qualitySetupFormSchema,
} from "@/features/quality-setup/quality-setup-schema";

describe("qualitySetupFormSchema", () => {
  it("accepts an inspection result", () => {
    expect(
      qualitySetupFormSchema.safeParse({
        ...emptyQualitySetupForm,
        code: "PARTIAL",
        name: "Partially accepted",
        is_accepted: true,
      }).success,
    ).toBe(true);
  });

  it("rejects malformed codes", () => {
    expect(
      qualitySetupFormSchema.safeParse({
        ...emptyQualitySetupForm,
        code: "NOT ACCEPTED",
        name: "Not accepted",
      }).success,
    ).toBe(false);
  });
});
