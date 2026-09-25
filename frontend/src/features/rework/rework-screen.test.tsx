import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { NuqsTestingAdapter } from "nuqs/adapters/testing";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import {
  listWarehouses,
  listWarehouseOwners,
} from "@/features/warehouses/warehouse-api";
import { listReworkTasks } from "./rework-api";
import { ReworkScreen } from "./rework-screen";
import { capabilities, testTask } from "./rework-test-fixtures";
vi.mock("@/features/warehouses/warehouse-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/warehouses/warehouse-api")
  >()),
  listWarehouses: vi.fn(),
  listWarehouseOwners: vi.fn(),
}));
vi.mock("./rework-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./rework-api")>()),
  listReworkTasks: vi.fn(),
}));
vi.mock("./rework-detail-dialog", () => ({
  ReworkDetailDialog: ({ taskId }: { taskId: string }) => (
    <p>Detail {taskId}</p>
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
  vi.mocked(listReworkTasks).mockImplementation(async (filters) => ({
    items: [testTask],
    page: filters.page,
    page_size: 10,
    total_items: 11,
    total_pages: 2,
  }));
});
afterEach(cleanup);
function mount(params = "", canRework = true) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <NuqsTestingAdapter searchParams={params} hasMemory>
        <ReworkScreen capabilities={{ ...capabilities, canRework }} />
      </NuqsTestingAdapter>
    </QueryClientProvider>,
  );
}
it("shows assignee usernames in desktop and mobile entries", async () => {
  vi.mocked(listReworkTasks).mockResolvedValue({
    items: [
      {
        ...testTask,
        assigned_to: "other-account-id",
        assigned_username: "godown.warehouse",
      },
    ],
    page: 1,
    page_size: 10,
    total_items: 1,
    total_pages: 1,
  });
  mount("?warehouse=warehouse-1&owner=owner-1");
  await waitFor(() =>
    expect(screen.getAllByText("@godown.warehouse")).toHaveLength(2),
  );
  expect(screen.queryByText("other-account-id")).not.toBeInTheDocument();
  expect(screen.getAllByText(/2\.000000 EA/).length).toBeGreaterThan(0);
});
it("waits for a valid scope and does not fetch owners for an unauthorized warehouse", async () => {
  mount("?warehouse=outside&owner=owner-1");
  await waitFor(() => expect(listWarehouses).toHaveBeenCalled());
  await screen.findByText("Select a warehouse and served owner");
  expect(listReworkTasks).not.toHaveBeenCalled();
  expect(listWarehouseOwners).not.toHaveBeenCalled();
});
it("applies status and submitted search, resetting the page", async () => {
  const user = userEvent.setup();
  mount("?warehouse=warehouse-1&owner=owner-1&status=OPEN&page=2");
  await waitFor(() =>
    expect(listReworkTasks).toHaveBeenCalledWith(
      expect.objectContaining({
        ownerId: "owner-1",
        warehouseId: "warehouse-1",
        status: "OPEN",
        page: 2,
      }),
    ),
  );
  await user.type(screen.getByLabelText("Search rework tasks"), "  seal  ");
  await user.click(screen.getByRole("button", { name: "Search" }));
  await waitFor(() =>
    expect(listReworkTasks).toHaveBeenLastCalledWith(
      expect.objectContaining({ search: "seal", page: 1 }),
    ),
  );
  await user.click(
    screen.getByRole("combobox", { name: "Filter rework status" }),
  );
  await user.click(await screen.findByRole("option", { name: "In progress" }));
  await waitFor(() =>
    expect(listReworkTasks).toHaveBeenLastCalledWith(
      expect.objectContaining({ status: "IN_PROGRESS", page: 1 }),
    ),
  );
});
it("clears the previous owner and detail when switching warehouses without crossed requests", async () => {
  const user = userEvent.setup();
  mount("?warehouse=warehouse-1&owner=owner-1&task=RWK-1");
  await screen.findByText("Detail RWK-1");
  await waitFor(() => expect(listReworkTasks).toHaveBeenCalled());
  await user.click(
    screen.getByRole("combobox", { name: "Filter by warehouse" }),
  );
  await user.click(
    await screen.findByRole("option", { name: "Warehouse two (WH2)" }),
  );
  await waitFor(() =>
    expect(listReworkTasks).toHaveBeenLastCalledWith(
      expect.objectContaining({
        warehouseId: "warehouse-2",
        ownerId: "owner-2",
      }),
    ),
  );
  expect(screen.queryByText("Detail RWK-1")).not.toBeInTheDocument();
  expect(
    vi
      .mocked(listReworkTasks)
      .mock.calls.every(([filters]) =>
        filters.warehouseId === "warehouse-1"
          ? filters.ownerId === "owner-1"
          : filters.ownerId === "owner-2",
      ),
  ).toBe(true);
});
it("supports pagination and read-only detail viewing on desktop and mobile", async () => {
  const user = userEvent.setup();
  mount("?warehouse=warehouse-1&owner=owner-1", false);
  const next = await screen.findByRole("button", { name: "Next" });
  await waitFor(() => expect(next).toBeEnabled());
  await user.click(next);
  await waitFor(() =>
    expect(listReworkTasks).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 2 }),
    ),
  );
  await user.click(screen.getByRole("button", { name: "View" }));
  await screen.findByText("Detail RWK-1");
  expect(screen.getByRole("button", { name: "View task" })).toBeInTheDocument();
  expect(
    screen.getByText(/INBOUND.REWORK is required to execute tasks/),
  ).toBeInTheDocument();
  expect(
    screen.queryByRole("button", {
      name: /create|assign|delete|cancel|complete/i,
    }),
  ).not.toBeInTheDocument();
});
it("explains automatic creation when the scoped list is empty", async () => {
  vi.mocked(listReworkTasks).mockResolvedValue({
    items: [],
    page: 1,
    page_size: 10,
    total_items: 0,
    total_pages: 0,
  });
  mount("?warehouse=warehouse-1&owner=owner-1");
  await screen.findByText("No rework tasks found");
  expect(
    screen.queryByRole("button", { name: "Next" }),
  ).not.toBeInTheDocument();
});
