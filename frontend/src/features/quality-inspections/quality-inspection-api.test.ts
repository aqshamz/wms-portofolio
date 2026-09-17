import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  cancelInspection,
  completeInspection,
  createInspection,
  inspectionListPath,
  listReceiptInspections,
} from "./quality-inspection-api";

vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));
beforeEach(() => vi.clearAllMocks());

describe("quality inspection requests", () => {
  it("encodes scope and filters", () => {
    expect(
      inspectionListPath({
        ownerId: "owner-1",
        warehouseId: "wh-1",
        status: "PENDING",
        search: "  lot 1  ",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inbound/quality-inspections?owner_id=owner-1&warehouse_id=wh-1&page=2&page_size=10&status_code=PENDING&search=lot+1",
    );
  });
  it("includes both optimistic versions on completion", async () => {
    const body = {
      expected_version: 3,
      expected_balance_version: 7,
      passed_qty: "8",
      failed_qty: "2",
      putaway_target_location_id: "target",
    };
    await completeInspection("QC/1", body);
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/inbound/quality-inspections/QC%2F1/complete",
      { method: "POST", body },
    );
    await cancelInspection("QC/1", 3, "Wrong sampling plan");
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/inbound/quality-inspections/QC%2F1/cancel",
      {
        method: "POST",
        body: { expected_version: 3, reason: "Wrong sampling plan" },
      },
    );
    await createInspection({ receipt_inventory_id: "batch" });
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/inbound/quality-inspections",
      { method: "POST", body: { receipt_inventory_id: "batch" } },
    );
  });
  it("checks every receipt inspection page and removes unrelated matches", async () => {
    vi.mocked(apiRequest)
      .mockResolvedValueOnce({
        items: [
          { receipt_id: "RCV-1", receipt_inventory_id: "batch1" },
          { receipt_id: "RCV-10", receipt_inventory_id: "other" },
        ],
        total_pages: 2,
      })
      .mockResolvedValueOnce({
        items: [{ receipt_id: "RCV-1", receipt_inventory_id: "batch2" }],
        total_pages: 2,
      });
    const result = await listReceiptInspections("owner", "warehouse", "RCV-1");
    expect(result.map((row) => row.receipt_inventory_id)).toEqual([
      "batch1",
      "batch2",
    ]);
    expect(apiRequest).toHaveBeenNthCalledWith(
      2,
      expect.stringContaining("page=2&page_size=100"),
    );
  });
});
