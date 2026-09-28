import { beforeEach, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  cancelVendorReturn,
  completeVendorReturn,
  getVendorReturn,
  listVendorReturns,
  vendorReturnListPath,
} from "./vendor-return-api";

vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));
beforeEach(() => vi.clearAllMocks());

const filters = {
  ownerId: "owner-1",
  warehouseId: "warehouse-1",
  status: "PLANNED",
  search: " vendor & item ",
  page: 2,
  pageSize: 10,
};

it("builds a scoped return to vendor list request", async () => {
  const path = vendorReturnListPath(filters);
  expect(
    Object.fromEntries(new URL(path, "http://localhost").searchParams),
  ).toEqual({
    owner_id: "owner-1",
    warehouse_id: "warehouse-1",
    status_code: "PLANNED",
    search: "vendor & item",
    page: "2",
    page_size: "10",
  });
  await listVendorReturns(filters);
  expect(apiRequest).toHaveBeenCalledWith(path);
});

it("posts document and balance versions to lifecycle actions", async () => {
  await getVendorReturn("RTV/1");
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/vendor-returns/RTV%2F1",
  );
  await completeVendorReturn("RTV/1", {
    expected_version: 2,
    expected_balance_version: 7,
    completed_at: "2026-09-25T10:00:00Z",
  });
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/vendor-returns/RTV%2F1/complete",
    {
      method: "POST",
      body: {
        expected_version: 2,
        expected_balance_version: 7,
        completed_at: "2026-09-25T10:00:00Z",
      },
    },
  );
  await cancelVendorReturn("RTV/1", { expected_version: 2, reason: "Retain" });
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/vendor-returns/RTV%2F1/cancel",
    { method: "POST", body: { expected_version: 2, reason: "Retain" } },
  );
});
