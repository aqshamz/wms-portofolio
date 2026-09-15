import { describe, expect, it } from "vitest";

import { inventoryStatusListPath } from "@/features/inventory-classifications/inventory-classification-api";

describe("inventory status paths", () => {
  it("serializes search, status, and pagination", () => {
    expect(
      inventoryStatusListPath({
        search: " available ",
        active: "active",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/master/inventory-statuses?page=2&page_size=10&search=available&active=true",
    );
  });
});
