import { describe, expect, it } from "vitest";
import { receiptLocationOptions } from "@/features/receipts/receipt-location-options";
import type {
  LocationType,
  WarehouseLocation,
} from "@/features/storage-layout/storage-layout-types";

const dock: LocationType = {
  location_type_id: "dock-type",
  code: "DOCK",
  name: "Dock",
  allows_receiving: true,
  allows_shipping: true,
  allows_storage: false,
  allows_picking: false,
  is_active: true,
};
const receiving: LocationType = {
  ...dock,
  location_type_id: "receive-type",
  code: "RECEIVING",
  name: "Receiving",
  allows_shipping: false,
};
const storage: LocationType = {
  ...dock,
  location_type_id: "storage-type",
  code: "STORAGE",
  name: "Storage",
  allows_receiving: false,
  allows_storage: true,
};

function location(id: string, type: LocationType): WarehouseLocation {
  return {
    location_id: id,
    location_type_id: type.location_type_id,
    warehouse_id: "warehouse-1",
    zone_id: "zone-1",
    code: id,
    pick_sequence: 0,
    is_pick_face: false,
    is_locked: false,
    is_active: true,
    created_at: "2026-09-16",
  };
}

describe("receiptLocationOptions", () => {
  it("shows only Docks in the header, and Docks plus receiving areas for batches", () => {
    const result = receiptLocationOptions(
      [
        location("dock-1", dock),
        location("receive-1", receiving),
        location("storage-1", storage),
      ],
      [dock, receiving, storage],
    );
    expect(result.docks.map((item) => item.location_id)).toEqual(["dock-1"]);
    expect(result.receiving.map((item) => item.location_id)).toEqual([
      "dock-1",
      "receive-1",
    ]);
  });

  it("excludes locked/inactive locations and inactive location types", () => {
    const result = receiptLocationOptions(
      [
        { ...location("dock-1", dock), is_locked: true },
        { ...location("dock-2", dock), is_active: false },
        location("receive-1", receiving),
      ],
      [dock, { ...receiving, is_active: false }],
    );
    expect(result).toEqual({ docks: [], receiving: [] });
  });
});
