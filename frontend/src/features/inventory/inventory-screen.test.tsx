import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { NuqsTestingAdapter } from "nuqs/adapters/testing";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  listWarehouseOwners,
  listWarehouses,
} from "@/features/warehouses/warehouse-api";
import { listBalances } from "./inventory-api";
import { InventoryBalancesScreen } from "./inventory-screen";

vi.mock("@/features/warehouses/warehouse-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/warehouses/warehouse-api")
  >()),
  listWarehouses: vi.fn(),
  listWarehouseOwners: vi.fn(),
}));

vi.mock("./inventory-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./inventory-api")>()),
  listBalances: vi.fn(),
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
  vi.mocked(listBalances).mockResolvedValue({
    items: [
      {
        balance_id: "balance-1",
        owner_id: "owner-1",
        warehouse_id: "warehouse-1",
        location_id: "location-1",
        location_code: "A-01",
        item_id: "item-1",
        item_code: "COFFEE",
        item_name: "Coffee beans",
        lot_id: "lot-1",
        lot_number: "LOT-001",
        handling_unit_id: null,
        inventory_status_id: "status-1",
        inventory_status_code: "AVAILABLE",
        on_hand_qty: "100.000000",
        reserved_qty: "10.000000",
        available_qty: "90.000000",
        uom_id: "uom-1",
        uom_code: "KG",
        version_no: 3,
        updated_at: "2026-09-23T02:00:00Z",
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
        <InventoryBalancesScreen timezone="Asia/Jakarta" />
      </NuqsTestingAdapter>
    </QueryClientProvider>,
  );
}

it("does not request balances before an allowed scope is selected", async () => {
  mount();
  expect(
    await screen.findByText("Select a warehouse and served owner"),
  ).toBeInTheDocument();
  expect(listBalances).not.toHaveBeenCalled();
});

it("renders scoped balances for desktop and mobile layouts", async () => {
  mount("warehouse=warehouse-1&owner=owner-1");
  await waitFor(() =>
    expect(listBalances).toHaveBeenCalledWith(
      expect.objectContaining({
        warehouseId: "warehouse-1",
        ownerId: "owner-1",
      }),
    ),
  );
  expect(await screen.findAllByText("COFFEE")).toHaveLength(2);
  expect(
    await screen.findAllByText("100.000000 / 10.000000 / 90.000000 KG"),
  ).toHaveLength(2);
});
