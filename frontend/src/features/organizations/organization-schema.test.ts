import { describe, expect, it } from "vitest";

import {
  emptyOrganizationForm,
  organizationFormSchema,
} from "@/features/organizations/organization-schema";

describe("organizationFormSchema", () => {
  it("accepts the default local organization shape once required fields are set", () => {
    const result = organizationFormSchema.safeParse({
      ...emptyOrganizationForm,
      code: "JKT_DC",
      name: "Jakarta Distribution Center",
    });

    expect(result.success).toBe(true);
  });

  it("rejects invalid business and country codes", () => {
    const result = organizationFormSchema.safeParse({
      ...emptyOrganizationForm,
      code: "invalid code",
      name: "Test",
      country_code: "Indonesia",
    });

    expect(result.success).toBe(false);
  });
});
