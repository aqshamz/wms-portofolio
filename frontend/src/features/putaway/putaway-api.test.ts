import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  assignPutaway,
  cancelPutaway,
  completePutaway,
  putawayListPath,
  putawayLookupPath,
  retargetPutaway,
  reversePutaway,
  startPutaway,
} from "./putaway-api";

vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));
beforeEach(() => vi.clearAllMocks());
describe("putaway requests", () => {
  it("includes the scoped list filters", () =>
    expect(
      putawayListPath({
        ownerId: "owner",
        warehouseId: "wh",
        status: "IN_PROGRESS",
        search: "  bulk 1  ",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inbound/putaway-tasks?owner_id=owner&warehouse_id=wh&page=2&page_size=10&status_code=IN_PROGRESS&search=bulk+1",
    ));
  it("uses task-specific paginated lookups without unrelated security permissions", () =>
    expect(
      putawayLookupPath("PUT/1", "assignees", {
        search: "  worker 1  ",
        page: 2,
        pageSize: 20,
      }),
    ).toBe(
      "/api/v1/inbound/putaway-tasks/PUT%2F1/assignees?page=2&page_size=20&search=worker+1",
    ));
  it("sends optimistic versions and the correct action payloads", async () => {
    await startPutaway("PUT-1", 2);
    expect(apiRequest).toHaveBeenLastCalledWith(
      "/api/v1/inbound/putaway-tasks/PUT-1/start",
      { method: "POST", body: { expected_version: 2 } },
    );
    await assignPutaway("PUT-1", 2, "worker");
    expect(apiRequest).toHaveBeenLastCalledWith(
      "/api/v1/inbound/putaway-tasks/PUT-1/assign",
      { method: "POST", body: { expected_version: 2, account_id: "worker" } },
    );
    await retargetPutaway("PUT-1", 2, "target");
    expect(apiRequest).toHaveBeenLastCalledWith(
      "/api/v1/inbound/putaway-tasks/PUT-1/retarget",
      {
        method: "POST",
        body: { expected_version: 2, target_location_id: "target" },
      },
    );
    const posting = {
      expected_version: 2,
      expected_balance_version: 7,
      business_date: "2026-09-17",
    };
    await completePutaway("PUT-1", posting);
    expect(apiRequest).toHaveBeenLastCalledWith(
      "/api/v1/inbound/putaway-tasks/PUT-1/complete",
      { method: "POST", body: posting },
    );
    await cancelPutaway("PUT-1", { ...posting, reason: "Wrong plan" });
    expect(apiRequest).toHaveBeenLastCalledWith(
      "/api/v1/inbound/putaway-tasks/PUT-1/cancel",
      { method: "POST", body: { ...posting, reason: "Wrong plan" } },
    );
    await reversePutaway("PUT-1", {
      ...posting,
      expected_balance_version: 9,
      reason: "Wrong slot",
    });
    expect(apiRequest).toHaveBeenLastCalledWith(
      "/api/v1/inbound/putaway-tasks/PUT-1/reverse",
      {
        method: "POST",
        body: { ...posting, expected_balance_version: 9, reason: "Wrong slot" },
      },
    );
  });
});
