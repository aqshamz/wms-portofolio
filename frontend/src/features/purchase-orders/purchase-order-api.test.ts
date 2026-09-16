import { describe, expect, it } from "vitest";

import { purchaseOrderListPath } from "@/features/purchase-orders/purchase-order-api";

describe("purchaseOrderListPath", () => {
  it("always sends the required owner and warehouse scope", () => {
    expect(
      purchaseOrderListPath({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        status: "DRAFT",
        search: "PO 100",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inbound/purchase-orders?owner_id=owner-1&warehouse_id=warehouse-1&page=2&page_size=10&status_code=DRAFT&search=PO+100",
    );
  });
});
