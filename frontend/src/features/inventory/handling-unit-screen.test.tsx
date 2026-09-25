import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { NuqsTestingAdapter } from "nuqs/adapters/testing";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  listWarehouseOwners,
  listWarehouses,
} from "@/features/warehouses/warehouse-api";
import { getLocation } from "@/features/storage-layout/storage-layout-api";
import { getHandlingUnitType } from "@/features/units-packaging/units-packaging-api";
import { listHandlingUnits } from "./inventory-api";
import { InventoryHandlingUnitsScreen } from "./handling-unit-screen";

vi.mock("@/features/warehouses/warehouse-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/warehouses/warehouse-api")
  >()),
  listWarehouses: vi.fn(),
  listWarehouseOwners: vi.fn(),
}));
vi.mock(
  "@/features/storage-layout/storage-layout-api",
  async (importOriginal) => ({
    ...(await importOriginal<
      typeof import("@/features/storage-layout/storage-layout-api")
    >()),
    getLocation: vi.fn(),
  }),
);
vi.mock(
  "@/features/units-packaging/units-packaging-api",
  async (importOriginal) => ({
    ...(await importOriginal<
      typeof import("@/features/units-packaging/units-packaging-api")
    >()),
    getHandlingUnitType: vi.fn(),
  }),
);
vi.mock("./inventory-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./inventory-api")>()),
  listHandlingUnits: vi.fn(),
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
  vi.mocked(listHandlingUnits).mockResolvedValue({
    items: [
      {
        handling_unit_id: "hu-1",
        warehouse_id: "warehouse-1",
        owner_id: "owner-1",
        handling_unit_type_id: "type-1",
        parent_handling_unit_id: null,
        current_location_id: "location-1",
        barcode: "PALLET-0001",
        is_closed: false,
        positive_balance_count: 0,
        child_count: 0,
        created_at: "2026-09-23T02:00:00Z",
        created_by: "account-1",
      },
    ],
    page: 1,
    page_size: 10,
    total_items: 1,
    total_pages: 1,
  });
  vi.mocked(getHandlingUnitType).mockResolvedValue({
    handling_unit_type_id: "type-1",
    code: "PALLET",
    name: "Pallet",
    is_active: true,
  });
  vi.mocked(getLocation).mockResolvedValue({
    location_id: "location-1",
    warehouse_id: "warehouse-1",
    zone_id: "zone-1",
    zone_code: "BULK",
    location_type_id: "location-type-1",
    code: "BULK-01",
    pick_sequence: 1,
    is_pick_face: false,
    is_locked: false,
    is_active: true,
    created_at: "2026-09-23T00:00:00Z",
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
        <InventoryHandlingUnitsScreen timezone="Asia/Jakarta" />
      </NuqsTestingAdapter>
    </QueryClientProvider>,
  );
}

it("does not request handling units before scope selection", async () => {
  mount();
  expect(
    await screen.findByText("Select a warehouse and served owner"),
  ).toBeInTheDocument();
  expect(listHandlingUnits).not.toHaveBeenCalled();
});

it("requests warehouse-scoped handling units and resolves references", async () => {
  mount("warehouse=warehouse-1&owner=owner-1&status=open");
  await waitFor(() =>
    expect(listHandlingUnits).toHaveBeenCalledWith(
      expect.objectContaining({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        status: "open",
      }),
    ),
  );
  expect(await screen.findAllByText("PALLET-0001")).toHaveLength(2);
  expect(await screen.findAllByText("PALLET · Pallet")).toHaveLength(2);
  expect(await screen.findAllByText("BULK-01 · BULK")).toHaveLength(2);
});
