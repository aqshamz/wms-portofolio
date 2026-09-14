import { describe, expect, it } from "vitest";

import {
  businessPartnerListPath,
  partnerTypeListPath,
} from "@/features/business-partners/business-partner-api";

describe("business partner paths", () => {
  it("serializes owner and partner type filters", () => {
    expect(
      businessPartnerListPath({
        ownerId: "owner-id",
        partnerTypeCode: "SUPPLIER",
        search: " vendor ",
        active: "active",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/master/business-partners?page=2&page_size=10&owner_id=owner-id&partner_type_code=SUPPLIER&search=vendor&active=true",
    );
  });

  it("serializes global partner type filters", () => {
    expect(
      partnerTypeListPath({
        search: "",
        active: "all",
        page: 1,
        pageSize: 100,
      }),
    ).toBe("/api/v1/master/partner-types?page=1&page_size=100");
  });
});
