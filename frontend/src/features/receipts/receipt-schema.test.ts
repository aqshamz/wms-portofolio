import { describe, expect, it } from "vitest";

import { receiptFormSchema } from "@/features/receipts/receipt-schema";

const validReceipt = {
  inbound_id: "INB-100",
  business_date: "2026-09-16",
  received_at: "2026-09-16T09:00",
  dock_location_id: "location-1",
  vehicle_number: "B 1234 WMS",
  seal_number: "SEAL-1",
  delivery_note_no: "DN-100",
  notes: "",
  lines: [
    {
      inbound_line_id: "INB-100-L0001",
      item_id: "item-1",
      uom_id: "uom-ea",
      uom_conversion_to_base: "1",
      base_uom_id: "uom-ea",
      base_uom_code: "EA",
      received_qty: "10",
      rejected_qty: "1",
      exception_type_code: "DAMAGED" as const,
      exception_notes: "One damaged carton",
      lot_controlled: true,
      serial_controlled: false,
      batches: [
        {
          source_qty: "9",
          received_location_id: "location-1",
          lot_number: "LOT-100",
          manufacture_date: "2026-08-01",
          expiry_date: "2027-08-01",
          handling_unit_id: "",
          serial_no: "",
        },
      ],
    },
  ],
};

describe("receiptFormSchema", () => {
  it.each(["", "invalid", "-1"])(
    "reports invalid quantity %j without throwing",
    (quantity) => {
      for (const field of ["received_qty", "rejected_qty"] as const) {
        expect(
          receiptFormSchema.safeParse({
            ...validReceipt,
            lines: [{ ...validReceipt.lines[0], [field]: quantity }],
          }).success,
        ).toBe(false);
      }
      expect(
        receiptFormSchema.safeParse({
          ...validReceipt,
          lines: [
            {
              ...validReceipt.lines[0],
              batches: [
                { ...validReceipt.lines[0].batches[0], source_qty: quantity },
              ],
            },
          ],
        }).success,
      ).toBe(false);
    },
  );
  it("accepts a balanced lot-controlled receipt", () => {
    expect(receiptFormSchema.safeParse(validReceipt).success).toBe(true);
  });

  it("requires batch quantities to equal accepted quantity", () => {
    const result = receiptFormSchema.safeParse({
      ...validReceipt,
      lines: [
        {
          ...validReceipt.lines[0],
          batches: [{ ...validReceipt.lines[0].batches[0], source_qty: "8" }],
        },
      ],
    });
    expect(result.success).toBe(false);
  });

  it("requires exception notes when quantity is rejected", () => {
    const result = receiptFormSchema.safeParse({
      ...validReceipt,
      lines: [{ ...validReceipt.lines[0], exception_notes: "" }],
    });
    expect(result.success).toBe(false);
  });

  it("requires lot and serial identities for controlled items", () => {
    const result = receiptFormSchema.safeParse({
      ...validReceipt,
      lines: [
        {
          ...validReceipt.lines[0],
          serial_controlled: true,
          batches: [
            {
              ...validReceipt.lines[0].batches[0],
              lot_number: "",
              serial_no: "",
            },
          ],
        },
      ],
    });
    expect(result.success).toBe(false);
  });

  it("rejects a repeated serial number for the same item", () => {
    const batch = validReceipt.lines[0].batches[0];
    const result = receiptFormSchema.safeParse({
      ...validReceipt,
      lines: [
        {
          ...validReceipt.lines[0],
          received_qty: "2",
          rejected_qty: "0",
          exception_type_code: "NONE",
          exception_notes: "",
          serial_controlled: true,
          batches: [
            { ...batch, source_qty: "1", serial_no: "SN-100" },
            { ...batch, source_qty: "1", serial_no: "SN-100" },
          ],
        },
      ],
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            path: ["lines", 0, "batches", 1, "serial_no"],
          }),
        ]),
      );
    }
  });

  it("accepts one serial row per whole base unit", () => {
    const batch = validReceipt.lines[0].batches[0];
    const result = receiptFormSchema.safeParse({
      ...validReceipt,
      lines: [
        {
          ...validReceipt.lines[0],
          received_qty: "2",
          rejected_qty: "0",
          exception_type_code: "NONE",
          exception_notes: "",
          lot_controlled: false,
          serial_controlled: true,
          batches: [
            {
              ...batch,
              source_qty: "1",
              lot_number: "",
              serial_no: "SN-100",
            },
            {
              ...batch,
              source_qty: "1",
              lot_number: "",
              serial_no: "SN-101",
            },
          ],
        },
      ],
    });

    expect(result.success).toBe(true);
  });

  it("rejects a non-base receiving UOM for serialized items", () => {
    const batch = validReceipt.lines[0].batches[0];
    const result = receiptFormSchema.safeParse({
      ...validReceipt,
      lines: [
        {
          ...validReceipt.lines[0],
          uom_id: "uom-box",
          uom_conversion_to_base: "2",
          received_qty: "1",
          rejected_qty: "0",
          exception_type_code: "NONE",
          exception_notes: "",
          lot_controlled: false,
          serial_controlled: true,
          batches: [
            {
              ...batch,
              source_qty: "0.5",
              lot_number: "",
              serial_no: "SN-100",
            },
            {
              ...batch,
              source_qty: "0.5",
              lot_number: "",
              serial_no: "SN-101",
            },
          ],
        },
      ],
    });

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            path: ["lines", 0, "uom_id"],
          }),
        ]),
      );
    }
  });
});
