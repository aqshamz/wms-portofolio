import { beforeEach, describe, expect, it, vi } from "vitest";

import { apiRequest } from "@/lib/api/client";
import {
  assignWarehouseOwner,
  deactivateWarehouseOwner,
  listWarehouseOwners,
  warehouseListPath,
} from "./warehouse-api";

vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));

describe("warehouse owner API", () => {
  beforeEach(() => vi.mocked(apiRequest).mockReset());

  it("loads served owners", async () => {
    vi.mocked(apiRequest).mockResolvedValue([]);

    await listWarehouseOwners("WH-1");

    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/master/warehouses/WH-1/owners",
    );
  });

  it("filters warehouse choices by served owner", () => {
    expect(
      warehouseListPath({
        search: "",
        active: "active",
        ownerId: "OWNER-1",
        page: 1,
        pageSize: 100,
      }),
    ).toBe(
      "/api/v1/master/warehouses?page=1&page_size=100&owner_id=OWNER-1&active=true",
    );
  });

  it("assigns a served owner", async () => {
    vi.mocked(apiRequest).mockResolvedValue(undefined);

    await assignWarehouseOwner("WH-1", "OWNER-1");

    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/master/warehouses/WH-1/owners",
      { method: "POST", body: { owner_id: "OWNER-1" } },
    );
  });

  it("deactivates a served owner", async () => {
    vi.mocked(apiRequest).mockResolvedValue(undefined);

    await deactivateWarehouseOwner("WH-1", "OWNER-1");

    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/master/warehouses/WH-1/owners/OWNER-1/deactivate",
      { method: "PATCH" },
    );
  });
});
