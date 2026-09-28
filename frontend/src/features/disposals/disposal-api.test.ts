import { beforeEach, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  cancelDisposal,
  completeDisposal,
  disposalKeys,
  disposalListPath,
  getDisposal,
  listDisposals,
} from "./disposal-api";

vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));

beforeEach(() => vi.clearAllMocks());

const filters = {
  ownerId: "owner-1",
  warehouseId: "warehouse-1",
  status: "PLANNED",
  search: "  item & lot  ",
  page: 2,
  pageSize: 10,
};

it("builds a scoped, encoded disposal list request", async () => {
  const path = disposalListPath(filters);
  expect(
    Object.fromEntries(new URL(path, "http://localhost").searchParams),
  ).toEqual({
    owner_id: "owner-1",
    warehouse_id: "warehouse-1",
    status_code: "PLANNED",
    search: "item & lot",
    page: "2",
    page_size: "10",
  });
  await listDisposals(filters);
  expect(apiRequest).toHaveBeenCalledWith(path);
  expect(disposalKeys.list(filters)).not.toEqual(
    disposalKeys.list({ ...filters, ownerId: "owner-2" }),
  );
});

it("posts optimistic document and balance versions to lifecycle actions", async () => {
  await getDisposal("DSP/1");
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/disposals/DSP%2F1",
  );
  await completeDisposal("DSP/1", {
    expected_version: 2,
    expected_balance_version: 7,
    completed_at: "2026-09-25T10:00:00Z",
  });
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/disposals/DSP%2F1/complete",
    {
      method: "POST",
      body: {
        expected_version: 2,
        expected_balance_version: 7,
        completed_at: "2026-09-25T10:00:00Z",
      },
    },
  );
  await cancelDisposal("DSP/1", { expected_version: 2, reason: "Retain" });
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/outbound/disposals/DSP%2F1/cancel",
    {
      method: "POST",
      body: { expected_version: 2, reason: "Retain" },
    },
  );
});
