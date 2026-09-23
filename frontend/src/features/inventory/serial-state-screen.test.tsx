import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { NuqsTestingAdapter } from "nuqs/adapters/testing";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  listWarehouseOwners,
  listWarehouses,
} from "@/features/warehouses/warehouse-api";
import { listSerialStates } from "./inventory-api";
import { InventorySerialStatesScreen } from "./serial-state-screen";

vi.mock("@/features/warehouses/warehouse-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/warehouses/warehouse-api")
  >()),
  listWarehouses: vi.fn(),
  listWarehouseOwners: vi.fn(),
}));
vi.mock("./inventory-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./inventory-api")>()),
  listSerialStates: vi.fn(),
}));

const warehouse = {
  warehouse_id: "warehouse-1",
  operator_id: "operator-1",
  code: "WH1",
  name: "Warehouse one",
  timezone_name: "Asia/Jakarta",
  is_active: true,
  created_at: "2026-09-23T00:00:00Z",
  updated_at: "2026-09-23T00:00:00Z",
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(listWarehouses).mockResolvedValue({
    items: [
      warehouse,
      { ...warehouse, warehouse_id: "warehouse-2", code: "WH2" },
    ],
    page: 1,
    page_size: 100,
    total_items: 2,
    total_pages: 1,
  });
  vi.mocked(listWarehouseOwners).mockResolvedValue([
    {
      warehouse_id: "warehouse-1",
      owner_id: "owner-1",
      owner_code: "OWNER",
      owner_name: "Served owner",
      is_active: true,
      created_at: "2026-09-23T00:00:00Z",
    },
  ]);
  vi.mocked(listSerialStates).mockResolvedValue({
    items: [
      {
        serial_id: "serial-1",
        serial_no: "SN-0001",
        version_no: 4,
        updated_at: "2026-09-23T02:00:00Z",
        balance: {
          balance_id: "balance-1",
          owner_id: "owner-1",
          warehouse_id: "warehouse-1",
          location_id: "location-1",
          location_code: "PICK-01",
          item_id: "item-1",
          item_code: "SCANNER",
          item_name: "Barcode scanner",
          lot_id: null,
          lot_number: null,
          handling_unit_id: null,
          inventory_status_id: "status-1",
          inventory_status_code: "AVAILABLE",
          on_hand_qty: "2.000000",
          reserved_qty: "0.000000",
          available_qty: "2.000000",
          uom_id: "uom-1",
          uom_code: "EA",
          version_no: 2,
          updated_at: "2026-09-23T02:00:00Z",
        },
      },
    ],
    page: 1,
    page_size: 10,
    total_items: 1,
    total_pages: 1,
  });
});

afterEach(cleanup);

function mount(searchParams = "") {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <NuqsTestingAdapter searchParams={searchParams} hasMemory>
        <InventorySerialStatesScreen timezone="Asia/Jakarta" />
      </NuqsTestingAdapter>
    </QueryClientProvider>,
  );
}

it("does not request serial states without a complete scope", async () => {
  mount();
  expect(
    await screen.findByText("Select a warehouse and served owner"),
  ).toBeInTheDocument();
  expect(listSerialStates).not.toHaveBeenCalled();
});

it("renders current serial state for desktop and mobile layouts", async () => {
  mount("warehouse=warehouse-1&owner=owner-1");
  await waitFor(() =>
    expect(listSerialStates).toHaveBeenCalledWith(
      expect.objectContaining({
        warehouseId: "warehouse-1",
        ownerId: "owner-1",
      }),
    ),
  );
  expect(await screen.findAllByText("SN-0001")).toHaveLength(2);
  expect(screen.getByText("SCANNER")).toBeInTheDocument();
  expect(screen.getByText("SCANNER · Barcode scanner")).toBeInTheDocument();
  expect(screen.getAllByText("PICK-01")).toHaveLength(2);
});
