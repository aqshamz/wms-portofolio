import { describe, expect, it } from "vitest";

import { locationListPath } from "@/features/storage-layout/storage-layout-api";

describe("locationListPath", () => {
  it("serializes warehouse location filters for the backend", () => {
    expect(
      locationListPath({
        warehouseId: "warehouse-id",
        zoneId: "zone-id",
        locationTypeId: "type-id",
        search: " A-01 ",
        active: "active",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/master/warehouses/warehouse-id/locations?page=2&page_size=10&zone_id=zone-id&location_type_id=type-id&search=A-01&active=true",
    );
  });
});
