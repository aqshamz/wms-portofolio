import { describe, expect, it } from "vitest";
import {
  balanceListPath,
  lotListPath,
  movementListPath,
  serialStateListPath,
  serialListPath,
  handlingUnitListPath,
} from "./inventory-api";

describe("inventory balance paths", () => {
  it("always includes owner and warehouse scope", () => {
    expect(
      balanceListPath({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        search: " coffee ",
        includeZero: false,
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inventory/balances?owner_id=owner-1&warehouse_id=warehouse-1&page=2&page_size=10&search=coffee",
    );
  });

  it("only sends include_zero when selected", () => {
    expect(
      balanceListPath({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        search: "",
        includeZero: true,
        page: 1,
        pageSize: 10,
      }),
    ).toContain("include_zero=true");
  });

  it("can scope balance inquiry to one handling unit", () => {
    expect(
      balanceListPath({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        handlingUnitId: "HU-1",
        search: "",
        includeZero: false,
        page: 1,
        pageSize: 100,
      }),
    ).toContain("handling_unit_id=HU-1");
  });
});

describe("inventory movement paths", () => {
  it("serializes scope, type and trimmed search", () => {
    expect(
      movementListPath({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        movementTypeId: "type-1",
        search: " receipt-1 ",
        page: 3,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inventory/movements?owner_id=owner-1&warehouse_id=warehouse-1&page=3&page_size=10&movement_type_id=type-1&search=receipt-1",
    );
  });
});

describe("inventory serial-state paths", () => {
  it("serializes warehouse scope and search", () => {
    expect(
      serialStateListPath({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        search: " serial-001 ",
        page: 1,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inventory/serial-states?owner_id=owner-1&warehouse_id=warehouse-1&page=1&page_size=10&search=serial-001",
    );
  });
});

describe("inventory lot paths", () => {
  it("serializes owner-wide scope without a warehouse", () => {
    expect(
      lotListPath({
        ownerId: "owner-1",
        search: " lot-001 ",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inventory/lots?owner_id=owner-1&page=2&page_size=10&search=lot-001",
    );
  });
});

describe("inventory serial identity paths", () => {
  it("serializes owner-wide scope without a warehouse", () => {
    expect(
      serialListPath({
        ownerId: "owner-1",
        search: " sn-001 ",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inventory/serials?owner_id=owner-1&page=2&page_size=10&search=sn-001",
    );
  });
});

describe("inventory handling-unit paths", () => {
  it("serializes warehouse scope, status and search", () => {
    expect(
      handlingUnitListPath({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        status: "open",
        search: " pallet-1 ",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inventory/handling-units?owner_id=owner-1&warehouse_id=warehouse-1&page=2&page_size=10&closed=false&search=pallet-1",
    );
  });

  it("can scope operational lookup by location and parent", () => {
    expect(
      handlingUnitListPath({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        status: "open",
        search: "",
        locationId: "location-1",
        parentId: "HU-PARENT",
        page: 1,
        pageSize: 100,
      }),
    ).toContain(
      "current_location_id=location-1&parent_handling_unit_id=HU-PARENT",
    );
  });
});
