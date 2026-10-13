import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  FulfillmentDetailDialog,
  dispatchBalances,
} from "./fulfillment-detail-dialog";
import {
  getFulfillment,
  dispatchHandlingUnit,
  packLine,
  type Packing,
  type Shipment,
} from "./fulfillment-api";
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("./fulfillment-api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./fulfillment-api")>();
  return {
    ...actual,
    getFulfillment: vi.fn(),
    dispatchHandlingUnit: vi.fn(),
    dispatchLine: vi.fn(),
    packLine: vi.fn(),
    transitionFulfillment: vi.fn(),
    assignDriver: vi.fn(),
  };
});
vi.mock("@/features/inventory/inventory-api", () => ({
  handlingUnitKeys: { list: () => ["handling-units"] },
  listHandlingUnits: vi.fn().mockResolvedValue({ items: [], total_pages: 1 }),
}));
vi.mock("@/components/ui/operation-dialog", () => ({
  OperationDialog: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
}));
vi.mock("@/features/inventory/handling-unit-create-dialog", () => ({
  HandlingUnitCreateDialog: () => null,
}));
const shipment: Shipment = {
  shipment_id: "S1",
  owner_id: "owner",
  warehouse_id: "warehouse",
  status_code: "PLANNED",
  business_date: "2026-10-13",
  order_count: 1,
  line_count: 0,
  drivers: [
    {
      driver_id: "driver",
      driver_code: "D1",
      driver_name: "Driver",
      is_primary: true,
    },
  ],
  lines: [
    {
      packing_line_id: "P1",
      outbound_id: "DO1",
      item_code: "SKU1",
      source_balance_id: "B1",
      source_balance_version: 4,
      handling_unit_id: "HU1",
      handling_unit_barcode: "CARTON-001",
      shipped_qty: "5.000000",
      uom_code: "EA",
    },
    {
      packing_line_id: "P2",
      outbound_id: "DO1",
      item_code: "SKU2",
      source_balance_id: "B2",
      source_balance_version: 7,
      handling_unit_id: "HU1",
      handling_unit_barcode: "CARTON-001",
      shipped_qty: "1.000000",
      uom_code: "EA",
    },
  ],
};
const packing: Packing = {
  packing_id: "PK1",
  owner_id: "owner",
  warehouse_id: "warehouse",
  outbound_id: "DO1",
  client_delivery_order_no: "CLIENT-DO1",
  packing_location_id: "loc",
  packing_location_code: "PACK-01",
  status_code: "OPEN",
  line_count: 0,
  lines: [
    {
      packing_line_id: "",
      outbound_check_line_id: "C1",
      item_code: "SKU1",
      source_balance_id: "B1",
      source_balance_version: 3,
      handling_unit_id: null,
      expected_qty: "6.000000",
      packed_qty: "0.000000",
      uom_code: "EA",
    },
  ],
};
beforeEach(() => vi.clearAllMocks());
afterEach(cleanup);
function show(mode: "packings" | "shipments", canWrite: boolean) {
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
      <FulfillmentDetailDialog
        mode={mode}
        id={mode === "packings" ? "PK1" : "S1"}
        canWrite={canWrite}
        canCancel={false}
        canCreateHU={false}
        onOpenChange={vi.fn()}
      />
    </QueryClientProvider>,
  );
}
it("keeps viewers read-only and displays checked quantity rather than packed zero", async () => {
  vi.mocked(getFulfillment).mockResolvedValue(packing);
  show("packings", false);
  expect(await screen.findByText("SKU1 · 6.000000 EA")).toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Pack line" }),
  ).not.toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Complete packing" }),
  ).not.toBeInTheDocument();
});
it("requires confirmation before packing and preserves the source snapshot", async () => {
  vi.mocked(getFulfillment).mockResolvedValue(packing);
  vi.mocked(packLine).mockResolvedValue(packing);
  show("packings", true);
  const user = userEvent.setup();
  await user.click(await screen.findByRole("button", { name: "Pack line" }));
  expect(packLine).not.toHaveBeenCalled();
  await user.click(screen.getByRole("button", { name: "Confirm" }));
  await waitFor(() =>
    expect(packLine).toHaveBeenCalledWith(
      "PK1",
      "C1",
      expect.objectContaining({
        expected_balance_version: 3,
        operation_key: expect.any(String),
        handling_unit_id: undefined,
      }),
    ),
  );
});
it("dispatches one HU atomically, never one line of a mixed container", async () => {
  vi.mocked(getFulfillment).mockResolvedValue(shipment);
  vi.mocked(dispatchHandlingUnit).mockResolvedValue(shipment);
  show("shipments", true);
  const user = userEvent.setup();
  const button = await screen.findByRole("button", {
    name: "Dispatch entire HU",
  });
  expect(
    screen.queryByRole("button", { name: "Dispatch loose line" }),
  ).not.toBeInTheDocument();
  expect(
    screen.getByRole("button", { name: "Complete shipment" }),
  ).toBeDisabled();
  await user.click(button);
  expect(dispatchHandlingUnit).not.toHaveBeenCalled();
  await user.click(screen.getByRole("button", { name: "Confirm" }));
  await waitFor(() =>
    expect(dispatchHandlingUnit).toHaveBeenCalledWith("S1", "HU1", {
      business_date: "2026-10-13",
      operation_key: expect.any(String),
      expected_balances: [
        { balance_id: "B1", expected_version: 4 },
        { balance_id: "B2", expected_version: 7 },
      ],
    }),
  );
});
it("deduplicates shared-balance snapshots and rejects inconsistent versions", () => {
  const line = shipment.lines![0];
  expect(dispatchBalances([line, { ...line, packing_line_id: "P3" }])).toEqual([
    { balance_id: "B1", expected_version: 4 },
  ]);
  expect(() =>
    dispatchBalances([line, { ...line, source_balance_version: 5 }]),
  ).toThrow(/Refresh the HU/);
});
