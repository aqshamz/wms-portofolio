import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { NuqsTestingAdapter } from "nuqs/adapters/testing";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  listWarehouseOwners,
  listWarehouses,
} from "@/features/warehouses/warehouse-api";
import { listMovements, listMovementTypes } from "./inventory-api";
import { InventoryMovementsScreen } from "./movement-screen";

vi.mock("@/features/warehouses/warehouse-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/warehouses/warehouse-api")
  >()),
  listWarehouses: vi.fn(),
  listWarehouseOwners: vi.fn(),
}));
vi.mock("./inventory-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./inventory-api")>()),
  listMovements: vi.fn(),
  listMovementTypes: vi.fn(),
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
  vi.mocked(listMovementTypes).mockResolvedValue([
    {
      movement_type_id: "type-1",
      code: "RECEIVE",
      name: "Receive",
      description: null,
      is_active: true,
    },
  ]);
  vi.mocked(listMovements).mockResolvedValue({
    items: [
      {
        movement_id: "movement-1",
        movement_type_id: "type-1",
        movement_type_code: "RECEIVE",
        owner_id: "owner-1",
        warehouse_id: "warehouse-1",
        business_date: "2026-09-23",
        occurred_at: "2026-09-23T02:00:00Z",
        item_id: "item-1",
        item_code: "COFFEE",
        lot_id: "lot-1",
        lot_number: "LOT-001",
        serial_id: null,
        handling_unit_id: null,
        from_location_id: null,
        from_location_code: null,
        to_location_id: "location-1",
        to_location_code: "RCV-01",
        from_status_id: null,
        from_status_code: null,
        to_status_id: "status-1",
        to_status_code: "QC_PENDING",
        quantity: "100.000000",
        uom_id: "uom-1",
        uom_code: "KG",
        source_document_id: "RCV-001",
        source_line_id: "line-1",
        reason_code_id: null,
        notes: null,
        operation_key: "receipt:1",
        created_by: "account-1",
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
        <InventoryMovementsScreen timezone="Asia/Jakarta" />
      </NuqsTestingAdapter>
    </QueryClientProvider>,
  );
}

it("waits for a complete owner and warehouse scope", async () => {
  mount();
  expect(
    await screen.findByText("Select a warehouse and served owner"),
  ).toBeInTheDocument();
  expect(listMovements).not.toHaveBeenCalled();
});

it("renders scoped movements in desktop and mobile layouts", async () => {
  mount("warehouse=warehouse-1&owner=owner-1&movementType=type-1");
  await waitFor(() =>
    expect(listMovements).toHaveBeenCalledWith(
      expect.objectContaining({
        warehouseId: "warehouse-1",
        ownerId: "owner-1",
        movementTypeId: "type-1",
      }),
    ),
  );
  expect(await screen.findAllByText("COFFEE")).toHaveLength(2);
  expect(screen.getAllByText("RCV-001")).toHaveLength(2);
  expect(screen.getAllByText("100.000000 KG")).toHaveLength(2);
});
