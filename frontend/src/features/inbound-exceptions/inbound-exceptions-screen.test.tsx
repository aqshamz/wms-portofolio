import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { NuqsTestingAdapter } from "nuqs/adapters/testing";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  listWarehouses,
  listWarehouseOwners,
} from "@/features/warehouses/warehouse-api";
import { listInboundExceptions } from "./inbound-exception-api";
import { InboundExceptionsScreen } from "./inbound-exceptions-screen";
import { testException } from "./exception-test-fixtures";

vi.mock("@/features/warehouses/warehouse-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/warehouses/warehouse-api")
  >()),
  listWarehouses: vi.fn(),
  listWarehouseOwners: vi.fn(),
}));
vi.mock("./inbound-exception-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./inbound-exception-api")>()),
  listInboundExceptions: vi.fn(),
}));
vi.mock("./exception-detail-dialog", () => ({
  ExceptionDetailDialog: ({ exceptionId }: { exceptionId: string }) => (
    <p>Detail {exceptionId}</p>
  ),
}));
const warehouse = {
  warehouse_id: "warehouse-1",
  operator_id: "operator-1",
  code: "WH1",
  name: "Warehouse one",
  timezone_name: "Asia/Jakarta",
  is_active: true,
  created_at: "2026-09-18T00:00:00Z",
  updated_at: "2026-09-18T00:00:00Z",
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
      created_at: "2026-09-18T00:00:00Z",
    },
  ]);
  vi.mocked(listInboundExceptions).mockResolvedValue({
    items: [testException],
    page: 1,
    page_size: 10,
    total_items: 1,
    total_pages: 1,
  });
});
afterEach(cleanup);
function mount(params = "") {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <NuqsTestingAdapter searchParams={params} hasMemory>
        <InboundExceptionsScreen timezone="Asia/Jakarta" />
      </NuqsTestingAdapter>
    </QueryClientProvider>,
  );
}
it("waits for a valid scope without requesting broad exception data", async () => {
  mount();
  await screen.findByText("Select a warehouse and served owner");
  await waitFor(() => expect(listWarehouses).toHaveBeenCalled());
  expect(listInboundExceptions).not.toHaveBeenCalled();
  expect(listWarehouseOwners).not.toHaveBeenCalled();
});
it("does not fetch owners or exception lists for a warehouse outside the picker scope", async () => {
  mount("?warehouse=outside&owner=owner-1");
  await waitFor(() => expect(listWarehouses).toHaveBeenCalled());
  await screen.findByText("Select a warehouse and served owner");
  expect(listWarehouseOwners).not.toHaveBeenCalled();
  expect(listInboundExceptions).not.toHaveBeenCalled();
});
it("filters by exception type and trims submitted search, resetting pagination", async () => {
  const user = userEvent.setup();
  mount("?warehouse=warehouse-1&owner=owner-1&type=DAMAGED&page=2");
  await waitFor(() =>
    expect(listInboundExceptions).toHaveBeenCalledWith(
      expect.objectContaining({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        exceptionType: "DAMAGED",
        page: 2,
      }),
    ),
  );
  await user.type(
    screen.getByLabelText("Search inbound exceptions"),
    "  vendor  ",
  );
  await user.click(screen.getByRole("button", { name: "Search" }));
  await waitFor(() =>
    expect(listInboundExceptions).toHaveBeenLastCalledWith(
      expect.objectContaining({
        exceptionType: "DAMAGED",
        search: "vendor",
        page: 1,
      }),
    ),
  );
  await user.click(
    screen.getByRole("combobox", { name: "Filter exception type" }),
  );
  await user.click(await screen.findByRole("option", { name: "Reversal" }));
  await waitFor(() =>
    expect(listInboundExceptions).toHaveBeenLastCalledWith(
      expect.objectContaining({ exceptionType: "REVERSAL", page: 1 }),
    ),
  );
});
it("clears the old owner and selected detail when switching warehouses", async () => {
  const user = userEvent.setup();
  mount("?warehouse=warehouse-1&owner=owner-1&exception=IEX-1");
  await screen.findByText("Detail IEX-1");
  await waitFor(() => expect(listInboundExceptions).toHaveBeenCalled());
  await user.click(
    screen.getByRole("combobox", { name: "Filter by warehouse" }),
  );
  await user.click(
    await screen.findByRole("option", { name: "Warehouse two (WH2)" }),
  );
  await waitFor(() =>
    expect(listInboundExceptions).toHaveBeenLastCalledWith(
      expect.objectContaining({
        warehouseId: "warehouse-2",
        ownerId: "owner-2",
      }),
    ),
  );
  expect(screen.queryByText("Detail IEX-1")).not.toBeInTheDocument();
  expect(
    vi
      .mocked(listInboundExceptions)
      .mock.calls.every(([filters]) =>
        filters.warehouseId === "warehouse-1"
          ? filters.ownerId === "owner-1"
          : filters.ownerId === "owner-2",
      ),
  ).toBe(true);
});
it("opens details from desktop and mobile entries without edit controls", async () => {
  const user = userEvent.setup();
  mount("?warehouse=warehouse-1&owner=owner-1");
  await user.click(await screen.findByRole("button", { name: "View" }));
  await screen.findByText("Detail IEX-1");
  expect(
    screen.getByRole("button", { name: "View exception" }),
  ).toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: /create|edit|delete|resolve/i }),
  ).not.toBeInTheDocument();
});
it("shows an empty-state explanation when no automatic records exist", async () => {
  vi.mocked(listInboundExceptions).mockResolvedValue({
    items: [],
    page: 1,
    page_size: 10,
    total_items: 0,
    total_pages: 0,
  });
  mount("?warehouse=warehouse-1&owner=owner-1");
  await screen.findByText("No inbound exceptions found");
  expect(
    screen.queryByRole("button", { name: "Next" }),
  ).not.toBeInTheDocument();
});
