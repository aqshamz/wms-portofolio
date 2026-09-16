import { describe, expect, it } from "vitest";

import { purchaseOrderFormSchema } from "@/features/purchase-orders/purchase-order-schema";

const validOrder = {
  owner_id: "owner-1",
  warehouse_id: "warehouse-1",
  vendor_id: "vendor-1",
  business_date: "2026-09-16",
  purchase_order_no: "PO-100",
  ordered_at: "2026-09-16T09:00",
  expected_arrival_at: "2026-09-17T09:00",
  notes: "",
  lines: [
    {
      item_id: "item-1",
      ordered_qty: "10.5",
      uom_id: "uom-1",
      vendor_item_code: "",
      expected_lot_no: "",
      expected_expiry_date: "",
      notes: "",
      over_receipt_tolerance_pct: "5",
      under_receipt_tolerance_pct: "0",
    },
  ],
};

describe("purchaseOrderFormSchema", () => {
  it("accepts a scoped order with a receiving line", () => {
    expect(purchaseOrderFormSchema.safeParse(validOrder).success).toBe(true);
  });

  it("requires at least one line", () => {
    expect(
      purchaseOrderFormSchema.safeParse({ ...validOrder, lines: [] }).success,
    ).toBe(false);
  });

  it("rejects arrival before the ordered time", () => {
    expect(
      purchaseOrderFormSchema.safeParse({
        ...validOrder,
        expected_arrival_at: "2026-09-15T09:00",
      }).success,
    ).toBe(false);
  });
});
