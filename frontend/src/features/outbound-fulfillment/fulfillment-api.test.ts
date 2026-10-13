import { beforeEach, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  allPages,
  dispatchHandlingUnit,
  dispatchLine,
  fulfillmentListPath,
  packLine,
  packingCandidates,
  shipmentCandidates,
  listCarriers,
  listDrivers,
} from "./fulfillment-api";
vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));
beforeEach(() => vi.resetAllMocks());
const filters = {
  ownerId: "owner",
  warehouseId: "warehouse",
  status: "OPEN",
  search: "  DO & 1  ",
  page: 2,
  pageSize: 10,
};
it("keeps owner and warehouse in every list and eligible candidate request", async () => {
  const path = fulfillmentListPath("packings", filters);
  expect(
    Object.fromEntries(new URL(path, "http://localhost").searchParams),
  ).toEqual({
    owner_id: "owner",
    warehouse_id: "warehouse",
    status_code: "OPEN",
    search: "DO & 1",
    page: "2",
    page_size: "10",
  });
  vi.mocked(apiRequest).mockResolvedValue({ items: [], total_pages: 1 });
  await packingCandidates(filters);
  expect(apiRequest).toHaveBeenLastCalledWith(
    expect.stringMatching(
      /packing-candidates\?owner_id=owner&warehouse_id=warehouse/,
    ),
  );
  await shipmentCandidates(filters);
  expect(apiRequest).toHaveBeenLastCalledWith(
    expect.stringMatching(
      /shipment-candidates\?owner_id=owner&warehouse_id=warehouse/,
    ),
  );
});
it("preserves idempotency keys and each expected balance version", async () => {
  const body = {
    business_date: "2026-10-13",
    operation_key: "retry-key",
    expected_balances: [
      { balance_id: "B1", expected_version: 3 },
      { balance_id: "B2", expected_version: 8 },
    ],
  };
  await dispatchHandlingUnit("S/1", "HU/1", body);
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/shipments/S%2F1/handling-units/HU%2F1/dispatch",
    { method: "POST", body },
  );
  const line = {
    business_date: body.business_date,
    operation_key: body.operation_key,
    expected_balance_version: 3,
  };
  await dispatchLine("S/1", "P/1", line);
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/shipments/S%2F1/lines/P%2F1/dispatch",
    { method: "POST", body: line },
  );
  await packLine("P/1", "C/1", { ...line, handling_unit_id: "HU1" });
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/packings/P%2F1/lines/C%2F1",
    { method: "POST", body: { ...line, handling_unit_id: "HU1" } },
  );
});
it("loads later pages rather than silently truncating eligible choices", async () => {
  const load = vi
    .fn()
    .mockResolvedValueOnce({ items: ["first"], total_pages: 2 })
    .mockResolvedValueOnce({ items: ["second"], total_pages: 2 });
  expect(await allPages(load)).toEqual(["first", "second"]);
  expect(load).toHaveBeenNthCalledWith(2, 2);
});
it("uses the unpaginated transport contracts", async () => {
  await listCarriers();
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/carriers?active=true",
  );
  await listDrivers("C/1");
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/carriers/C%2F1/drivers?active=true",
  );
});
