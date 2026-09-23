import { beforeEach, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  getReworkTask,
  listReworkTasks,
  reworkKeys,
  reworkListPath,
  transitionReworkTask,
} from "./rework-api";
vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));
beforeEach(() => vi.clearAllMocks());
const filters = {
  ownerId: "owner-1",
  warehouseId: "warehouse-1",
  status: "OPEN",
  search: "  seal & item  ",
  page: 2,
  pageSize: 10,
};
it("sends both scopes, status, pagination and safely encoded trimmed search", async () => {
  const path = reworkListPath(filters);
  expect(
    Object.fromEntries(new URL(path, "http://localhost").searchParams),
  ).toEqual({
    owner_id: "owner-1",
    warehouse_id: "warehouse-1",
    status_code: "OPEN",
    search: "seal & item",
    page: "2",
    page_size: "10",
  });
  await listReworkTasks(filters);
  expect(apiRequest).toHaveBeenCalledWith(path);
});
it("omits empty filters and isolates caches by scope", () => {
  expect(reworkListPath({ ...filters, status: "", search: " " })).not.toMatch(
    /status_code|search=/,
  );
  expect(reworkKeys.list(filters)).not.toEqual(
    reworkKeys.list({ ...filters, ownerId: "owner-2" }),
  );
});
it("encodes identifiers and posts the current document version to supported actions", async () => {
  await getReworkTask("RWK/1");
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/inbound/rework-tasks/RWK%2F1",
  );
  await transitionReworkTask("RWK/1", "start", { expected_version: 1 });
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/inbound/rework-tasks/RWK%2F1/start",
    { method: "POST", body: { expected_version: 1 } },
  );
  await transitionReworkTask("RWK/1", "complete", {
    expected_version: 2,
    result_notes: "Seal replaced",
  });
  expect(apiRequest).toHaveBeenLastCalledWith(
    "/api/v1/inbound/rework-tasks/RWK%2F1/complete",
    {
      method: "POST",
      body: { expected_version: 2, result_notes: "Seal replaced" },
    },
  );
});
