import { describe, expect, it } from "vitest";

import { inboundOrderFormSchema } from "@/features/inbound-orders/inbound-order-schema";

const validOrder = {
  purchase_order_id: "po-1",
  business_date: "2026-09-16",
  expected_arrival_at: "2026-09-17T09:00",
  external_reference: "ASN-100",
  supplier_reference: "DEL-100",
  notes: "",
  lines: [
    {
      purchase_order_line_id: "po-line-1",
      expected_qty: "10.5",
      customer_line_reference: "LINE-1",
      notes: "",
    },
  ],
};

describe("inboundOrderFormSchema", () => {
  it("accepts a Purchase Order-backed arrival plan", () => {
    expect(inboundOrderFormSchema.safeParse(validOrder).success).toBe(true);
  });

  it("requires at least one expected line", () => {
    expect(
      inboundOrderFormSchema.safeParse({ ...validOrder, lines: [] }).success,
    ).toBe(false);
  });

  it("rejects zero expected quantity", () => {
    expect(
      inboundOrderFormSchema.safeParse({
        ...validOrder,
        lines: [{ ...validOrder.lines[0], expected_qty: "0" }],
      }).success,
    ).toBe(false);
  });
});
