import { describe, expect, it } from "vitest";

import {
  businessPartnerFormSchema,
  emptyBusinessPartnerForm,
} from "@/features/business-partners/business-partner-schema";

describe("businessPartnerFormSchema", () => {
  it("accepts a valid partner profile", () => {
    expect(
      businessPartnerFormSchema.safeParse({
        ...emptyBusinessPartnerForm,
        code: "VENDOR_01",
        name: "Primary vendor",
        email: "vendor@example.com",
      }).success,
    ).toBe(true);
  });

  it("rejects malformed country codes and emails", () => {
    expect(
      businessPartnerFormSchema.safeParse({
        ...emptyBusinessPartnerForm,
        code: "VENDOR",
        name: "Vendor",
        country_code: "IDN",
        email: "invalid",
      }).success,
    ).toBe(false);
  });
});
