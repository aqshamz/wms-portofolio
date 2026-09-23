import { describe, expect, it } from "vitest";

import { receiptListPath } from "@/features/receipts/receipt-api";

describe("receiptListPath", () => {
  it("requests inspection eligibility only for the inspection picker", () => {
    const filters = {
      ownerId: "owner-1",
      warehouseId: "wh-1",
      status: "COMPLETED",
      search: "",
      page: 1,
      pageSize: 10,
    };
    expect(receiptListPath({ ...filters, inspectionEligible: true })).toContain(
      "inspection_eligible=true",
    );
    expect(receiptListPath(filters)).not.toContain("inspection_eligible");
    expect(
      receiptListPath({ ...filters, inspectionEligible: false }),
    ).not.toContain("inspection_eligible");
  });
  it("encodes operational scope and receipt filters", () => {
    expect(
      receiptListPath({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        status: "OPEN",
        search: "delivery 100",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inbound/receipts?owner_id=owner-1&warehouse_id=warehouse-1&page=2&page_size=10&status_code=OPEN&search=delivery+100",
    );
  });
});
