import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  createDisposition,
  listDispositionTypes,
  quarantineListPath,
  quarantineTargetPath,
} from "./quarantine-api";
vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));
beforeEach(() => vi.clearAllMocks());
describe("quarantine requests", () => {
  it("includes scope, status, search and pagination", () =>
    expect(
      quarantineListPath({
        ownerId: "owner",
        warehouseId: "wh",
        status: "PARTIALLY_DECIDED",
        search: "  LOT 1 ",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/inbound/quarantine-cases?owner_id=owner&warehouse_id=wh&page=2&page_size=10&status_code=PARTIALLY_DECIDED&search=LOT+1",
    ));
  it("uses case-scoped target lookup and active decision types", async () => {
    expect(
      quarantineTargetPath("Q/1", { search: " BULK ", page: 2, pageSize: 20 }),
    ).toBe(
      "/api/v1/inbound/quarantine-cases/Q%2F1/targets?page=2&page_size=20&search=BULK",
    );
    await listDispositionTypes();
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/inbound/quarantine-disposition-types?active=true",
    );
  });
  it("posts a decision with both optimistic versions", async () => {
    const body = {
      expected_case_version: 2,
      expected_balance_version: 7,
      disposition_type_code: "RETURN",
      disposition_qty: "2",
      business_date: "2026-09-17",
      decided_at: "2026-09-17T03:00:00.000Z",
    };
    await createDisposition("Q/1", body);
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/inbound/quarantine-cases/Q%2F1/dispositions",
      { method: "POST", body },
    );
  });
});
