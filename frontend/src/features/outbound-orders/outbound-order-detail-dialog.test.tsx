import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { OutboundOrderDetailDialog } from "./outbound-order-detail-dialog";
import {
  getOutboundOrder,
  transitionOutboundOrder,
  validateOutboundOrder,
} from "./outbound-order-api";
import type { OutboundOrder } from "./outbound-order-types";

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}));
vi.mock("./outbound-order-api", () => ({
  outboundOrderKeys: {
    all: ["outbound-orders"],
    detail: (id: string) => ["outbound-orders", "detail", id],
  },
  getOutboundOrder: vi.fn(),
  validateOutboundOrder: vi.fn(),
  transitionOutboundOrder: vi.fn(),
  addOutboundOrderLine: vi.fn(),
  updateOutboundOrderLine: vi.fn(),
  deleteOutboundOrderLine: vi.fn(),
  updateOutboundOrder: vi.fn(),
}));
vi.mock("@/components/ui/operation-dialog", () => ({
  OperationDialog: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
}));

const order: OutboundOrder = {
  outbound_id: "DO-1",
  owner_id: "owner",
  owner_code: "SIS",
  warehouse_id: "warehouse",
  warehouse_code: "BDG_WH",
  customer_id: "customer",
  customer_code: "STORE",
  customer_name: "Customer",
  ship_to_partner_id: "store",
  ship_to_code: "STORE",
  ship_to_name: "Store",
  ship_to_address_1: "Jl. Test",
  business_date: "2026-10-06",
  requested_ship_at: "2026-10-06T10:00:00+07:00",
  client_delivery_order_no: "CLIENT-DO-1",
  status_code: "DRAFT",
  version_no: 1,
  created_at: "2026-10-06T09:00:00+07:00",
  lines: [
    {
      outbound_line_id: "DO-1-L0001",
      line_no: 1,
      item_id: "item",
      item_code: "GEL",
      item_name: "Gel",
      ordered_qty: "2.000000",
      uom_id: "carton",
      uom_code: "CTN",
      uom_conversion_to_base: "12.000000",
      ordered_base_qty: "24.000000",
      base_uom_id: "ea",
      base_uom_code: "EA",
      allocated_qty: "0",
      picked_qty: "0",
      checked_qty: "0",
      packed_qty: "0",
      shipped_qty: "0",
    },
  ],
};
beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(getOutboundOrder).mockResolvedValue(order);
});
afterEach(cleanup);
function show(canPlan = true) {
  render(
    <QueryClientProvider
      client={
        new QueryClient({
          defaultOptions: {
            queries: { retry: false },
            mutations: { retry: false },
          },
        })
      }
    >
      <OutboundOrderDetailDialog
        id="DO-1"
        canPlan={canPlan}
        canCancel={false}
        onOpenChange={vi.fn()}
      />
    </QueryClientProvider>,
  );
}
it("shows frozen client/base quantities and hides write actions from viewers", async () => {
  show(false);
  expect(
    await screen.findByText(/Ordered 2.000000 CTN × 12.000000 = 24.000000 EA/),
  ).toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Validate DO" }),
  ).not.toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Edit header" }),
  ).not.toBeInTheDocument();
});
it("shows failed validation results without enabling release", async () => {
  vi.mocked(validateOutboundOrder).mockResolvedValue({
    validation_run_id: "VAL-1",
    status_code: "FAILED",
    results: [
      {
        rule_code: "BASE_UOM",
        rule_name: "Quantity snapshots",
        severity_code: "ERROR",
        passed: false,
        result_message: "Invalid snapshot",
      },
    ],
  });
  show();
  const user = userEvent.setup();
  await user.click(await screen.findByRole("button", { name: "Validate DO" }));
  expect(await screen.findByText(/Invalid snapshot/)).toBeInTheDocument();
  expect(validateOutboundOrder).toHaveBeenCalledWith("DO-1", 1);
  expect(
    screen.queryByRole("button", { name: "Release DO" }),
  ).not.toBeInTheDocument();
});
it("refreshes the validated document version before confirmed release", async () => {
  const validated = { ...order, status_code: "VALIDATED", version_no: 2 };
  vi.mocked(getOutboundOrder)
    .mockResolvedValueOnce(order)
    .mockResolvedValue(validated);
  vi.mocked(validateOutboundOrder).mockResolvedValue({
    validation_run_id: "VAL-1",
    status_code: "PASSED",
    results: [],
  });
  vi.mocked(transitionOutboundOrder).mockResolvedValue({
    ...validated,
    status_code: "RELEASED",
    version_no: 3,
  });
  show();
  const user = userEvent.setup();
  await user.click(await screen.findByRole("button", { name: "Validate DO" }));
  await user.click(await screen.findByRole("button", { name: "Release DO" }));
  expect(transitionOutboundOrder).not.toHaveBeenCalled();
  await user.click(screen.getByRole("button", { name: "Confirm" }));
  await waitFor(() =>
    expect(transitionOutboundOrder).toHaveBeenCalledWith(
      "DO-1",
      "release",
      2,
      "",
    ),
  );
});
