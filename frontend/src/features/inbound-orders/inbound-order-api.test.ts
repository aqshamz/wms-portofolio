import { describe, expect, it } from "vitest";

import { inboundOrderListPath } from "@/features/inbound-orders/inbound-order-api";

describe("inboundOrderListPath", () => {
  it("encodes required operational scope and filters", () => {
    expect(
      inboundOrderListPath({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        status: "RELEASED",
        search: "ASN 100",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inbound/orders?owner_id=owner-1&warehouse_id=warehouse-1&page=2&page_size=10&status_code=RELEASED&search=ASN+100",
    );
  });
});
