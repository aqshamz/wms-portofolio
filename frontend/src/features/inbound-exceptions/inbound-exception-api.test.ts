import { beforeEach, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  exceptionKeys,
  exceptionListPath,
  getInboundException,
  listInboundExceptions,
} from "./inbound-exception-api";
import { exceptionLabel, exceptionTime } from "./inbound-exception-types";

vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));
beforeEach(() => vi.clearAllMocks());
const filters = {
  ownerId: "owner-1",
  warehouseId: "warehouse-1",
  exceptionType: "DAMAGED",
  search: "  source & notes  ",
  page: 2,
  pageSize: 10,
};
it("maps the exception type to status_code and sends both scopes with trimmed search", async () => {
  const path = exceptionListPath(filters);
  const params = new URL(path, "http://localhost").searchParams;
  expect(Object.fromEntries(params)).toEqual({
    owner_id: "owner-1",
    warehouse_id: "warehouse-1",
    status_code: "DAMAGED",
    search: "source & notes",
    page: "2",
    page_size: "10",
  });
  await listInboundExceptions(filters);
  expect(apiRequest).toHaveBeenCalledWith(path);
});
it("omits all-type and blank-search filters, with distinct caches for different scopes", () => {
  const path = exceptionListPath({
    ...filters,
    exceptionType: "",
    search: " ",
  });
  expect(path).not.toContain("status_code");
  expect(path).not.toContain("search=");
  expect(exceptionKeys.list(filters)).not.toEqual(
    exceptionKeys.list({ ...filters, warehouseId: "warehouse-2" }),
  );
});
it("loads a scoped detail by encoded identifier without any write API", async () => {
  await getInboundException("IEX/1");
  expect(apiRequest).toHaveBeenCalledWith("/api/v1/inbound/exceptions/IEX%2F1");
});
it("labels custom types without guessing workflow status and formats the account timezone", () => {
  expect(exceptionLabel("CUSTOM_REASON")).toBe("Custom reason");
  expect(exceptionTime("2026-09-18T02:00:00Z", "Asia/Jakarta")).toContain(
    "09:00",
  );
  expect(exceptionTime("2026-09-18T02:00:00Z", "invalid")).toContain("09:00");
  expect(exceptionTime("invalid", "Asia/Jakarta")).toBe("—");
});
