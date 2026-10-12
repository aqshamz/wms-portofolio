import { expect, it } from "vitest";
import type {
  Balance,
  SerialState,
} from "@/features/inventory/inventory-types";
import { statusChangeValidation } from "./inventory-status-change-validation";

const source: Balance = {
  balance_id: "BAL-1",
  owner_id: "owner-1",
  warehouse_id: "warehouse-1",
  location_id: "location-1",
  location_code: "STORAGE-1",
  location_type_code: "STORAGE",
  item_id: "item-1",
  item_code: "ITEM-1",
  item_name: "Test item",
  lot_id: null,
  lot_number: null,
  handling_unit_id: null,
  inventory_status_id: "available",
  inventory_status_code: "AVAILABLE",
  on_hand_qty: "10.000000",
  reserved_qty: "3.000000",
  available_qty: "7.000000",
  uom_id: "ea",
  uom_code: "EA",
  version_no: 9,
  updated_at: "2026-10-06",
};
const input = {
  source,
  target: {
    inventory_status_id: "hold",
    code: "HOLD",
    name: "Hold",
    is_active: true,
    is_allocatable: false,
    is_pickable: false,
  },
  reason: {
    reason_code_id: "reason-1",
    code: "STATUS_HOLD",
    name: "Place on hold",
    description: null,
    requires_note: true,
    is_active: true,
  },
  quantity: "2",
  serial: undefined as SerialState | undefined,
  date: "2026-10-06",
  reference: "ISC-1",
  notes: "Packaging review",
};

it("accepts a partial unreserved status change without altering location or HU", () => {
  expect(statusChangeValidation(input)).toBeUndefined();
  expect(statusChangeValidation({ ...input, quantity: " 2 " })).toBeUndefined();
  expect(
    statusChangeValidation({
      ...input,
      source: { ...source, handling_unit_id: "hu-1" },
    }),
  ).toBeUndefined();
  expect(
    statusChangeValidation({
      ...input,
      source: { ...source, location_type_code: "PICK_FACE" },
    }),
  ).toBeUndefined();
});
it("rejects reserved quantities, zero, negative and malformed base quantities", () => {
  for (const quantity of [
    "8",
    "0",
    "-1",
    "0.0000001",
    "100000000000000",
    "NaN",
  ]) {
    expect(statusChangeValidation({ ...input, quantity })).toBeDefined();
  }
});
it("keeps workflow-owned statuses and receiving locations out of manual changes", () => {
  for (const inventory_status_code of [
    "QC_PENDING",
    "PUTAWAY_PENDING",
    "QUARANTINE",
    "RETURN_PENDING",
  ]) {
    expect(
      statusChangeValidation({
        ...input,
        source: { ...source, inventory_status_code },
      }),
    ).toBeDefined();
    expect(
      statusChangeValidation({
        ...input,
        target: { ...input.target, code: inventory_status_code },
      }),
    ).toBeDefined();
  }
  expect(
    statusChangeValidation({
      ...input,
      source: { ...source, location_type_code: "RECEIVING" },
    }),
  ).toBeDefined();
  expect(
    statusChangeValidation({
      ...input,
      source: { ...source, location_is_locked: true },
    }),
  ).toBeDefined();
  expect(
    statusChangeValidation({
      ...input,
      target: {
        ...input.target,
        inventory_status_id: source.inventory_status_id,
      },
    }),
  ).toBeDefined();
  expect(
    statusChangeValidation({
      ...input,
      target: { ...input.target, is_active: false },
    }),
  ).toBeDefined();
});
it("requires a valid reason, its mandatory notes and document reference", () => {
  expect(statusChangeValidation({ ...input, reason: undefined })).toBeDefined();
  expect(statusChangeValidation({ ...input, notes: " " })).toContain(
    "Notes are required",
  );
  expect(statusChangeValidation({ ...input, reference: " " })).toBeDefined();
  expect(statusChangeValidation({ ...input, date: "" })).toBeDefined();
});
it("requires exactly one serial belonging to the source balance", () => {
  const serialized = {
    ...input,
    source: { ...source, serial_controlled: true },
    quantity: "1",
  };
  const serial = {
    serial_id: "serial-1",
    serial_no: "SN-1",
    balance: source,
    version_no: 1,
    updated_at: "2026-10-06",
  };
  expect(statusChangeValidation(serialized)).toBeDefined();
  expect(statusChangeValidation({ ...serialized, serial })).toBeUndefined();
  expect(
    statusChangeValidation({ ...serialized, serial, quantity: "2" }),
  ).toBeDefined();
  expect(
    statusChangeValidation({
      ...serialized,
      serial: { ...serial, balance: { ...source, balance_id: "BAL-2" } },
    }),
  ).toBeDefined();
});
