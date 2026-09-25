import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { NuqsTestingAdapter } from "nuqs/adapters/testing";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  listWarehouseOwners,
  listWarehouses,
} from "@/features/warehouses/warehouse-api";
import { getItem } from "@/features/item-catalog/item-catalog-api";
import { listSerials } from "./inventory-api";
import { InventorySerialsScreen } from "./serial-screen";

vi.mock("@/features/warehouses/warehouse-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/warehouses/warehouse-api")
  >()),
  listWarehouses: vi.fn(),
  listWarehouseOwners: vi.fn(),
}));
vi.mock("@/features/item-catalog/item-catalog-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/item-catalog/item-catalog-api")
  >()),
  getItem: vi.fn(),
}));
vi.mock("./inventory-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./inventory-api")>()),
  listSerials: vi.fn(),
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
  vi.mocked(listSerials).mockResolvedValue({
    items: [
      {
        serial_id: "serial-1",
        owner_id: "owner-1",
        item_id: "item-1",
        serial_no: "SN-0001",
        created_at: "2026-09-23T02:00:00Z",
        created_by: "account-1",
      },
    ],
    page: 1,
    page_size: 10,
    total_items: 1,
    total_pages: 1,
  });
  vi.mocked(getItem).mockResolvedValue({
    item_id: "item-1",
    owner_id: "owner-1",
    code: "SCANNER",
    name: "Barcode scanner",
    base_uom_id: "uom-1",
    lot_controlled: false,
    serial_controlled: true,
    is_active: true,
    created_at: "2026-09-23T00:00:00Z",
    updated_at: "2026-09-23T00:00:00Z",
    uoms: [],
    barcodes: [],
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
        <InventorySerialsScreen timezone="Asia/Jakarta" />
      </NuqsTestingAdapter>
    </QueryClientProvider>,
  );
}

it("does not request serial identities before scope selection", async () => {
  mount();
  expect(
    await screen.findByText("Select a warehouse and served owner"),
  ).toBeInTheDocument();
  expect(listSerials).not.toHaveBeenCalled();
});

it("requests owner-wide serial identities and resolves item names", async () => {
  mount("warehouse=warehouse-1&owner=owner-1");
  await waitFor(() =>
    expect(listSerials).toHaveBeenCalledWith(
      expect.objectContaining({ ownerId: "owner-1" }),
    ),
  );
  const request = vi.mocked(listSerials).mock.calls[0][0];
  expect(request).not.toHaveProperty("warehouseId");
  expect(await screen.findAllByText("SN-0001")).toHaveLength(2);
  expect(await screen.findAllByText("SCANNER · Barcode scanner")).toHaveLength(
    2,
  );
});
