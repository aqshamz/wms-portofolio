import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { NuqsTestingAdapter } from "nuqs/adapters/testing";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  listWarehouseOwners,
  listWarehouses,
} from "@/features/warehouses/warehouse-api";
import { listDisposals } from "./disposal-api";
import { DisposalScreen } from "./disposal-screen";
import { disposalCapabilities, testDisposal } from "./disposal-test-fixtures";

vi.mock("@/features/warehouses/warehouse-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/warehouses/warehouse-api")
  >()),
  listWarehouses: vi.fn(),
  listWarehouseOwners: vi.fn(),
}));
vi.mock("./disposal-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./disposal-api")>()),
  listDisposals: vi.fn(),
}));
vi.mock("./disposal-detail-dialog", () => ({
  DisposalDetailDialog: ({ id }: { id: string }) => <p>Detail {id}</p>,
}));

const warehouse = {
  warehouse_id: "warehouse-1",
  operator_id: "operator-1",
  code: "WH1",
  name: "Warehouse one",
  timezone_name: "Asia/Jakarta",
  is_active: true,
  created_at: "2026-09-25T00:00:00Z",
  updated_at: "2026-09-25T00:00:00Z",
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(listWarehouses).mockResolvedValue({
    items: [warehouse],
    page: 1,
    page_size: 100,
    total_items: 1,
    total_pages: 1,
  });
  vi.mocked(listWarehouseOwners).mockResolvedValue([
    {
      warehouse_id: "warehouse-1",
      owner_id: "owner-1",
      owner_code: "OWNER",
      owner_name: "Owner one",
      is_active: true,
      created_at: "2026-09-25T00:00:00Z",
    },
  ]);
  vi.mocked(listDisposals).mockImplementation(async (filters) => ({
    items: [testDisposal],
    page: filters.page,
    page_size: 10,
    total_items: 1,
    total_pages: 1,
  }));
});

afterEach(cleanup);

function mount(params = "?warehouse=warehouse-1&owner=owner-1") {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <NuqsTestingAdapter searchParams={params} hasMemory>
        <DisposalScreen capabilities={disposalCapabilities} />
      </NuqsTestingAdapter>
    </QueryClientProvider>,
  );
}

it("lists scoped planned transactions and opens their detail", async () => {
  const user = userEvent.setup();
  mount();
  await waitFor(() =>
    expect(listDisposals).toHaveBeenCalledWith(
      expect.objectContaining({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
      }),
    ),
  );
  expect(await screen.findAllByText(testDisposal.disposal_id)).toHaveLength(2);
  await user.click(screen.getByRole("button", { name: "View" }));
  expect(
    await screen.findByText(`Detail ${testDisposal.disposal_id}`),
  ).toBeInTheDocument();
});

it("does not query transactions outside the accessible warehouse list", async () => {
  mount("?warehouse=outside&owner=owner-1");
  await screen.findByText("Select a warehouse and served owner");
  expect(listWarehouseOwners).not.toHaveBeenCalled();
  expect(listDisposals).not.toHaveBeenCalled();
});
