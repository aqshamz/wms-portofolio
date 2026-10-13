import { beforeEach, describe, expect, it, vi } from "vitest";
import { apiRequest } from "@/lib/api/client";
import {
  createOutboundOrder,
  allocateAvailableOutboundStock,
  deleteOutboundOrderLine,
  outboundOrderListPath,
  transitionOutboundOrder,
  updateOutboundOrderLine,
} from "./outbound-order-api";
vi.mock("@/lib/api/client", () => ({ apiRequest: vi.fn() }));
beforeEach(() => vi.resetAllMocks());
describe("Delivery Order API", () => {
  it("explicitly allows available-stock allocation with a version guard", async () => {
    await allocateAvailableOutboundStock("DO/1", 8);
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/outbound/orders/DO%2F1/allocate",
      { method: "POST", body: { expected_version: 8, allow_partial: true } },
    );
  });
  it("sends close-short reason and version to the separate action", async () => {
    await transitionOutboundOrder(
      "DO-1",
      "close-short",
      9,
      "Client agreed to shortage",
    );
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/outbound/orders/DO-1/close-short",
      {
        method: "POST",
        body: { expected_version: 9, reason: "Client agreed to shortage" },
      },
    );
  });
  it("encodes owner, warehouse, status and search", () => {
    expect(
      outboundOrderListPath({
        ownerId: "owner",
        warehouseId: "warehouse",
        status: "DRAFT",
        search: " DO 10 ",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/outbound/orders?owner_id=owner&warehouse_id=warehouse&page=2&page_size=10&status_code=DRAFT&search=DO+10",
    );
  });
  it("sends the chosen document UOM when editing a line", async () => {
    await updateOutboundOrderLine("DO-1", "DO-1-L0001", 9, {
      ordered_qty: "2",
      uom_id: "carton",
    });
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/outbound/orders/DO-1/lines/DO-1-L0001",
      {
        method: "PUT",
        body: { ordered_qty: "2", uom_id: "carton", expected_version: 9 },
      },
    );
  });
  it("keeps omitted UOM edits compatible", async () => {
    await updateOutboundOrderLine("DO-1", "line", 2, { ordered_qty: "3" });
    expect(apiRequest).toHaveBeenCalledWith(expect.any(String), {
      method: "PUT",
      body: { ordered_qty: "3", expected_version: 2 },
    });
  });
  it("guards draft deletion with the document version", async () => {
    await deleteOutboundOrderLine("DO-1", "line", 7);
    expect(apiRequest).toHaveBeenCalledWith(expect.any(String), {
      method: "DELETE",
      body: { expected_version: 7 },
    });
  });
  it("retains cancellation reason and version", async () => {
    await transitionOutboundOrder("DO-1", "cancel", 4, "Client cancelled");
    expect(apiRequest).toHaveBeenCalledWith(
      "/api/v1/outbound/orders/DO-1/cancel",
      {
        method: "POST",
        body: { expected_version: 4, reason: "Client cancelled" },
      },
    );
  });
  it("creates demand without sending computed base quantities", async () => {
    const request = {
      owner_id: "owner",
      warehouse_id: "warehouse",
      customer_id: "customer",
      ship_to_partner_id: "store",
      client_delivery_order_no: "CLIENT-1",
      business_date: "2026-10-06",
      requested_ship_at: "2026-10-06T10:00:00+07:00",
      lines: [{ item_id: "item", ordered_qty: "2", uom_id: "carton" }],
    };
    await createOutboundOrder(request);
    expect(apiRequest).toHaveBeenCalledWith("/api/v1/outbound/orders", {
      method: "POST",
      body: request,
    });
  });
});
