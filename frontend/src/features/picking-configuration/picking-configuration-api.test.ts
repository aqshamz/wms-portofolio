import { describe, expect, it } from "vitest";

import { pickingStrategyListPath } from "@/features/picking-configuration/picking-configuration-api";

describe("pickingStrategyListPath", () => {
  it("encodes scope and list filters", () => {
    expect(
      pickingStrategyListPath({
        search: "standard pick",
        active: "active",
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        page: 2,
        pageSize: 25,
      }),
    ).toBe(
      "/api/v1/master/picking-strategies?page=2&page_size=25&search=standard+pick&active=true&owner_id=owner-1&warehouse_id=warehouse-1",
    );
  });
});
