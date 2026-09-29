import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  approveAdjustment,
  completeReplenishment,
  createAdjustment,
  createReplenishment,
  listReplenishments,
  listReplenishmentTargets,
  listAdjustments,
  listStockControlReasons,
  postInternalMove,
} from "./stock-control-api";

vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));

describe("stock-control API", () => {
  beforeEach(() => vi.mocked(apiRequest).mockReset());

  it("loads active inventory reasons", async () => {
    vi.mocked(apiRequest).mockResolvedValue([]);
    await listStockControlReasons();
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/stock-control/reason-codes?active=true",
    );
  });

  it("posts an internal move command", async () => {
    vi.mocked(apiRequest).mockResolvedValue({});
    const request = {
      operation_key: "stock.internal-move.1",
      business_date: "2026-09-28",
      source_document_id: "IM-20260928",
      source_balance_id: "BAL-1",
      target_location_id: "11111111-1111-4111-8111-111111111111",
      quantity: "2",
      serial_ids: [],
      expected_version: 3,
    };
    await postInternalMove(request);
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/stock-control/internal-moves",
      { method: "POST", body: request },
    );
  });

  it("creates, lists and approves an inventory adjustment request", async () => {
    vi.mocked(apiRequest).mockResolvedValue({});
    const create = {
      business_date: "2026-09-29",
      direction: "DECREASE" as const,
      reason_code: "MANUAL_ADJUSTMENT",
      notes: "Physical verification",
      lines: [
        {
          balance_id: "BAL-1",
          quantity: "2",
          expected_balance_version: 4,
        },
        {
          balance_id: "BAL-2",
          quantity: "1",
          expected_balance_version: 7,
        },
      ],
    };
    await createAdjustment(create);
    expect(apiRequest).toHaveBeenNthCalledWith(
      1,
      "/api/v1/stock-control/adjustments",
      { method: "POST", body: create },
    );

    await listAdjustments({
      ownerId: "owner-1",
      warehouseId: "warehouse-1",
      status: "DRAFT",
      search: "gel",
      page: 1,
      pageSize: 10,
    });
    expect(apiRequest).toHaveBeenNthCalledWith(
      2,
      "/api/v1/stock-control/adjustments?owner_id=owner-1&warehouse_id=warehouse-1&page=1&page_size=10&status_code=DRAFT&search=gel",
    );

    await approveAdjustment("IADJ-1", 2, ["IADJ-1-L0002"]);
    expect(apiRequest).toHaveBeenNthCalledWith(
      3,
      "/api/v1/stock-control/adjustments/IADJ-1/approve",
      {
        method: "POST",
        body: { expected_version: 2, line_ids: ["IADJ-1-L0002"] },
      },
    );
  });

  it("lists scoped replenishment tasks", async () => {
    vi.mocked(apiRequest).mockResolvedValue({});
    await listReplenishments({
      ownerId: "owner-1",
      warehouseId: "warehouse-1",
      status: "IN_PROGRESS",
      search: "gel",
      page: 2,
      pageSize: 10,
    });
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/stock-control/replenishment-tasks?owner_id=owner-1&warehouse_id=warehouse-1&page=2&page_size=10&status_code=IN_PROGRESS&search=gel",
    );
  });

  it("creates and completes a replenishment task", async () => {
    vi.mocked(apiRequest).mockResolvedValue({});
    const create = {
      source_balance_id: "BAL-1",
      target_location_id: "11111111-1111-4111-8111-111111111111",
      quantity: "5",
      priority_code: "NORMAL",
      expected_balance_version: 3,
    };
    await createReplenishment(create);
    expect(apiRequest).toHaveBeenNthCalledWith(
      1,
      "/api/v1/stock-control/replenishment-tasks",
      { method: "POST", body: create },
    );
    await completeReplenishment("RPL-1", 2, 4, "2026-09-28");
    expect(apiRequest).toHaveBeenNthCalledWith(
      2,
      "/api/v1/stock-control/replenishment-tasks/RPL-1/complete",
      {
        method: "POST",
        body: {
          expected_version: 2,
          expected_balance_version: 4,
          business_date: "2026-09-28",
        },
      },
    );
  });

  it("loads scoped pick-face targets from stock control", async () => {
    vi.mocked(apiRequest).mockResolvedValue({});
    await listReplenishmentTargets("owner-1", "warehouse-1", "PICK");
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/stock-control/replenishment-targets?owner_id=owner-1&warehouse_id=warehouse-1&search=PICK&page=1&page_size=100",
    );
  });
});
