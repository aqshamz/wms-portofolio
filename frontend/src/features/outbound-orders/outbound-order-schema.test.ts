import { describe, expect, it } from "vitest";
import {
  outboundLineSchema,
  quantityPreview,
  linePayload,
} from "./outbound-order-schema";

describe("outbound quantity snapshots", () => {
  it("previews cartons in base units", () => {
    expect(quantityPreview("2", "12", false)).toEqual({ base: "24.000000" });
  });
  it("allows fractional cartons for whole serialized units", () => {
    expect(quantityPreview("0.5", "12", true)).toEqual({ base: "6.000000" });
    expect(quantityPreview("0.5", "3", true).error).toMatch(/whole/);
  });
  it("does not round or overflow the snapshot", () => {
    expect(quantityPreview("0.000001", "0.000001", false).error).toMatch(
      /precision/,
    );
    expect(quantityPreview("99999999999999", "2", false).error).toMatch(
      /precision/,
    );
  });
  it.each(["bad", "", "0", "-1", "1e2", "1.1234567"])(
    "rejects invalid quantity %s without throwing",
    (qty) => {
      expect(
        outboundLineSchema.safeParse({
          item_id: "8e228260-70f1-4f1d-87c1-c272e39bf659",
          uom_id: "39eee9a0-c820-421b-8ecb-30163e47b3ac",
          ordered_qty: qty,
          requested_lot_no: "",
          customer_line_reference: "",
          notes: "",
        }).success,
      ).toBe(false);
    },
  );
  it("sends document quantities and the selected UOM, not a browser-computed snapshot", () => {
    expect(
      linePayload({
        item_id: "item",
        uom_id: "carton",
        ordered_qty: "2",
        requested_lot_no: " LOT ",
        customer_line_reference: "",
        notes: "",
      }),
    ).toEqual({
      item_id: "item",
      uom_id: "carton",
      ordered_qty: "2",
      requested_lot_no: "LOT",
      customer_line_reference: undefined,
      notes: undefined,
    });
  });
});
