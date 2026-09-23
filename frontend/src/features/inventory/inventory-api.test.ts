import { describe, expect, it } from "vitest";
import {
  balanceListPath,
  movementListPath,
  serialStateListPath,
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
