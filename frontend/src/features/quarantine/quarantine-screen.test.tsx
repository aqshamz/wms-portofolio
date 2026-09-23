import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { NuqsTestingAdapter } from "nuqs/adapters/testing";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  listWarehouses,
  listWarehouseOwners,
} from "@/features/warehouses/warehouse-api";
import { listQuarantineCases } from "./quarantine-api";
import { QuarantineScreen } from "./quarantine-screen";
import { testCase } from "./quarantine-test-fixtures";

vi.mock("@/features/warehouses/warehouse-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/warehouses/warehouse-api")
  >()),
  listWarehouses: vi.fn(),
  listWarehouseOwners: vi.fn(),
}));
vi.mock("./quarantine-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./quarantine-api")>()),
  listQuarantineCases: vi.fn(),
}));
vi.mock("./quarantine-detail-dialog", () => ({
  QuarantineDetailDialog: () => null,
}));
const warehouse = {
  warehouse_id: "warehouse-1",
  operator_id: "operator-1",
  code: "WH1",
  name: "Warehouse one",
  timezone_name: "Asia/Jakarta",
  is_active: true,
  created_at: "2026-09-17T00:00:00Z",
  updated_at: "2026-09-17T00:00:00Z",
};
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(listWarehouses).mockResolvedValue({
    items: [
      warehouse,
      {
        ...warehouse,
        warehouse_id: "warehouse-2",
        code: "WH2",
        name: "Warehouse two",
      },
    ],
    page: 1,
    page_size: 100,
    total_items: 2,
    total_pages: 1,
  });
  vi.mocked(listWarehouseOwners).mockImplementation(async (id) => [
    {
      warehouse_id: id,
      owner_id: id === "warehouse-1" ? "owner-1" : "owner-2",
      owner_code: "OWNER",
      owner_name: "Served owner",
      is_active: true,
      created_at: "2026-09-17T00:00:00Z",
    },
  ]);
  vi.mocked(listQuarantineCases).mockResolvedValue({
    items: [testCase],
    page: 1,
    page_size: 10,
    total_items: 1,
    total_pages: 1,
  });
});
afterEach(cleanup);
function mount(params = "") {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <NuqsTestingAdapter searchParams={params} hasMemory>
        <QuarantineScreen
          capabilities={{ canDispose: false, timezone: "Asia/Jakarta" }}
        />
      </NuqsTestingAdapter>
    </QueryClientProvider>,
  );
}
it("does not request operational cases before selecting a valid scope", async () => {
  mount();
  await screen.findByText("Select a warehouse and served owner");
  await waitFor(() => expect(listWarehouses).toHaveBeenCalled());
  expect(listQuarantineCases).not.toHaveBeenCalled();
  expect(listWarehouseOwners).not.toHaveBeenCalled();
});
it("applies status and submitted search within the selected scope", async () => {
  const user = userEvent.setup();
  mount("?warehouse=warehouse-1&owner=owner-1&status=OPEN");
  await waitFor(() =>
    expect(listQuarantineCases).toHaveBeenCalledWith(
      expect.objectContaining({
        warehouseId: "warehouse-1",
        ownerId: "owner-1",
        status: "OPEN",
        page: 1,
      }),
    ),
  );
  await user.type(
    screen.getByLabelText("Search quarantine cases"),
    "  damaged  ",
  );
  await user.click(screen.getByRole("button", { name: "Search" }));
  await waitFor(() =>
    expect(listQuarantineCases).toHaveBeenLastCalledWith(
      expect.objectContaining({
        warehouseId: "warehouse-1",
        ownerId: "owner-1",
        search: "damaged",
        page: 1,
      }),
    ),
  );
});
it("clears the old owner when switching warehouses and never requests a crossed scope", async () => {
  const user = userEvent.setup();
  mount("?warehouse=warehouse-1&owner=owner-1");
  await waitFor(() => expect(listQuarantineCases).toHaveBeenCalled());
  await user.click(
    screen.getByRole("combobox", { name: "Filter by warehouse" }),
  );
  await user.click(
    await screen.findByRole("option", { name: "Warehouse two (WH2)" }),
  );
  await waitFor(() =>
    expect(listQuarantineCases).toHaveBeenLastCalledWith(
      expect.objectContaining({
        warehouseId: "warehouse-2",
        ownerId: "owner-2",
      }),
    ),
  );
  expect(
    vi
      .mocked(listQuarantineCases)
      .mock.calls.every(([filters]) =>
        filters.warehouseId === "warehouse-1"
          ? filters.ownerId === "owner-1"
          : filters.ownerId === "owner-2",
      ),
  ).toBe(true);
});
