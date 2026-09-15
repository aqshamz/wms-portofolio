import { describe, expect, it } from "vitest";

import { putawayStrategyListPath } from "@/features/putaway-configuration/putaway-configuration-api";

describe("putawayStrategyListPath", () => {
  it("encodes scope and list filters", () => {
    expect(
      putawayStrategyListPath({
        search: "reserve storage",
        active: "inactive",
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        page: 1,
        pageSize: 100,
      }),
    ).toBe(
      "/api/v1/master/putaway-strategies?page=1&page_size=100&search=reserve+storage&active=false&owner_id=owner-1&warehouse_id=warehouse-1",
    );
  });
});
