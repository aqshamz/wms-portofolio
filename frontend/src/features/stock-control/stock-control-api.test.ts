import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  approveAdjustment,
  approveCycleCount,
  completeReplenishment,
  createAdjustment,
  createCycleCount,
  createGrandStockOpname,
  createReplenishment,
  listReplenishments,
  listReplenishmentTargets,
  listAdjustments,
  listCycleCounts,
  listStockControlReasons,
  postInternalMove,
  postWarehouseTransfer,
  recordCycleCount,
  createWarehouseTransfer,
  listWarehouseTransfers,
  receiveWarehouseTransfer,
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

  it("posts an atomic warehouse transfer command", async () => {
    vi.mocked(apiRequest).mockResolvedValue({});
    const request = {
      operation_key: "stock.warehouse-transfer.1",
      business_date: "2026-10-02",
      source_document_id: "WT-20261002",
      source_balance_id: "BAL-1",
      target_warehouse_id: "11111111-1111-4111-8111-111111111111",
      target_location_id: "22222222-2222-4222-8222-222222222222",
      target_inventory_status_id: "33333333-3333-4333-8333-333333333333",
      quantity: "2",
      serial_ids: [],
      expected_version: 3,
    };
    await postWarehouseTransfer(request);
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/stock-control/warehouse-transfers",
      { method: "POST", body: request },
    );
  });

  it("creates, lists and receives a warehouse transfer document", async () => {
    vi.mocked(apiRequest).mockResolvedValue({});
    const create = {
      business_date: "2026-10-02",
      source_balance_id: "BAL-1",
      target_warehouse_id: "warehouse-2",
      quantity: "2",
      serial_ids: [],
    };
    await createWarehouseTransfer(create);
    expect(apiRequest).toHaveBeenNthCalledWith(
      1,
      "/api/v1/stock-control/warehouse-transfer-documents",
      { method: "POST", body: create },
    );
    await listWarehouseTransfers({
      ownerId: "owner-1",
      warehouseId: "warehouse-2",
      side: "TARGET",
      status: "IN_TRANSIT",
      search: "gel",
      page: 1,
      pageSize: 20,
    });
    expect(apiRequest).toHaveBeenNthCalledWith(
      2,
      "/api/v1/stock-control/warehouse-transfer-documents?owner_id=owner-1&warehouse_id=warehouse-2&side=TARGET&page=1&page_size=20&status_code=IN_TRANSIT&search=gel",
    );
    const receive = {
      expected_version: 3,
      business_date: "2026-10-03",
      receipt_location_id: "dock-1",
      putaway_target_location_id: "storage-1",
    };
    await receiveWarehouseTransfer("TRF-1", receive);
    expect(apiRequest).toHaveBeenNthCalledWith(
      3,
      "/api/v1/stock-control/warehouse-transfer-documents/TRF-1/receive",
      { method: "POST", body: receive },
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

  it("creates, records and partially approves a cycle count", async () => {
    vi.mocked(apiRequest).mockResolvedValue({});
    const create = {
      business_date: "2026-09-29",
      tolerance_quantity: "1",
      notes: "Aisle A weekly count",
      balance_ids: ["BAL-1", "BAL-2"],
    };
    await createCycleCount(create);
    expect(apiRequest).toHaveBeenNthCalledWith(
      1,
      "/api/v1/stock-control/cycle-counts",
      { method: "POST", body: create },
    );

    await recordCycleCount("CCNT-1", 1, [
      { line_id: "CCNT-1-L0001", counted_quantity: "9" },
    ]);
    expect(apiRequest).toHaveBeenNthCalledWith(
      2,
      "/api/v1/stock-control/cycle-counts/CCNT-1/count",
      {
        method: "POST",
        body: {
          expected_version: 1,
          lines: [{ line_id: "CCNT-1-L0001", counted_quantity: "9" }],
        },
      },
    );

    await approveCycleCount("CCNT-1", 2, ["CCNT-1-L0001"]);
    expect(apiRequest).toHaveBeenNthCalledWith(
      3,
      "/api/v1/stock-control/cycle-counts/CCNT-1/approve",
      {
        method: "POST",
        body: {
          expected_version: 2,
          line_ids: ["CCNT-1-L0001"],
        },
      },
    );
  });

  it("lists and creates customer-scoped grand stock opnames", async () => {
    vi.mocked(apiRequest).mockResolvedValue({});
    await listCycleCounts({
      ownerId: "owner-1",
      warehouseId: "warehouse-1",
      countType: "GRAND",
      status: "DRAFT",
      search: "GSO",
      page: 1,
      pageSize: 20,
    });
    expect(apiRequest).toHaveBeenNthCalledWith(
      1,
      "/api/v1/stock-control/cycle-counts?owner_id=owner-1&warehouse_id=warehouse-1&count_type_code=GRAND&page=1&page_size=20&status_code=DRAFT&search=GSO",
    );

    const request = {
      owner_id: "owner-1",
      warehouse_id: "warehouse-1",
      business_date: "2026-10-05",
      tolerance_quantity: "0",
      notes: "Annual count",
    };
    await createGrandStockOpname(request);
    expect(apiRequest).toHaveBeenNthCalledWith(
      2,
      "/api/v1/stock-control/cycle-counts/grand",
      { method: "POST", body: request },
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
