import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { InventoryStatusChangeDialog } from "./inventory-status-change-dialog";
import {
  getBalance,
  listSerialStates,
} from "@/features/inventory/inventory-api";
import { listInventoryStatuses } from "@/features/inventory-classifications/inventory-classification-api";
import {
  listStockControlReasons,
  postInventoryStatusChange,
} from "./stock-control-api";
import type { Balance } from "@/features/inventory/inventory-types";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@/features/inventory/inventory-api", () => ({
  balanceKeys: {
    detail: (id: string) => ["inventory", "balances", "detail", id],
  },
  serialStateKeys: {
    list: (filters: unknown) => ["inventory", "serial-states", filters],
  },
  getBalance: vi.fn(),
  listSerialStates: vi.fn(),
}));
vi.mock(
  "@/features/inventory-classifications/inventory-classification-api",
  () => ({
    inventoryClassificationKeys: {
      list: (filters: unknown) => ["statuses", filters],
    },
    listInventoryStatuses: vi.fn(),
  }),
);
vi.mock("./stock-control-api", () => ({
  stockControlKeys: { reasons: () => ["stock-control", "reasons"] },
  listStockControlReasons: vi.fn(),
  postInventoryStatusChange: vi.fn(),
}));
vi.mock("@/components/ui/operation-dialog", () => ({
  OperationDialog: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
}));
vi.mock("@/components/ui/select", () => ({
  Select: ({
    value,
    options,
    onValueChange,
    ariaLabel,
    disabled,
  }: {
    value: string;
    options: { value: string; label: string }[];
    onValueChange: (value: string) => void;
    ariaLabel: string;
    disabled?: boolean;
  }) => (
    <select
      aria-label={ariaLabel}
      value={value}
      disabled={disabled}
      onChange={(event) => onValueChange(event.target.value)}
    >
      <option value="">Select</option>
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  ),
}));

const source: Balance = {
  balance_id: "BAL-1",
  owner_id: "owner-1",
  warehouse_id: "warehouse-1",
  location_id: "location-1",
  location_code: "STORAGE-1",
  location_type_code: "STORAGE",
  item_id: "item-1",
  item_code: "ITEM-1",
  item_name: "Test item",
  lot_id: null,
  lot_number: null,
  handling_unit_id: null,
  inventory_status_id: "available",
  inventory_status_code: "AVAILABLE",
  on_hand_qty: "10.000000",
  reserved_qty: "3.000000",
  available_qty: "7.000000",
  uom_id: "ea",
  uom_code: "EA",
  version_no: 9,
  updated_at: "2026-10-06",
};
beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(getBalance).mockResolvedValue(source);
  vi.mocked(listInventoryStatuses).mockResolvedValue({
    items: [
      {
        inventory_status_id: "available",
        code: "AVAILABLE",
        name: "Available",
        is_active: true,
        is_allocatable: true,
        is_pickable: true,
      },
      {
        inventory_status_id: "hold",
        code: "HOLD",
        name: "Hold",
        is_active: true,
        is_allocatable: false,
        is_pickable: false,
      },
      {
        inventory_status_id: "qc",
        code: "QC_PENDING",
        name: "QC pending",
        is_active: true,
        is_allocatable: false,
        is_pickable: false,
      },
    ],
    page: 1,
    page_size: 100,
    total_items: 3,
    total_pages: 1,
  });
  vi.mocked(listStockControlReasons).mockResolvedValue([
    {
      reason_code_id: "reason-1",
      code: "STATUS_HOLD",
      name: "Place on hold",
      description: null,
      is_active: true,
      requires_note: true,
    },
  ]);
  vi.mocked(postInventoryStatusChange).mockResolvedValue({
    operation: "STATUS_CHANGE",
    movements: [],
    idempotent_replay: false,
  });
});
afterEach(cleanup);

function mount() {
  const onOpenChange = vi.fn();
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <InventoryStatusChangeDialog
        balanceId="BAL-1"
        timezone="Asia/Jakarta"
        onOpenChange={onOpenChange}
      />
    </QueryClientProvider>,
  );
  return { user: userEvent.setup(), onOpenChange };
}
async function fill() {
  const result = mount();
  await screen.findByRole("option", { name: "Hold (HOLD)" });
  expect(
    screen.queryByRole("option", {
      name: /QC pending|Available \(AVAILABLE\)/,
    }),
  ).not.toBeInTheDocument();
  await result.user.selectOptions(
    screen.getByRole("combobox", { name: "Target inventory status" }),
    "hold",
  );
  await result.user.type(screen.getByLabelText(/Quantity/), "2");
  await result.user.selectOptions(
    screen.getByRole("combobox", { name: "Inventory status-change reason" }),
    "STATUS_HOLD",
  );
  await result.user.type(screen.getByLabelText(/Notes/), "Packaging review");
  return result;
}
it("requires confirmation and posts the source version and unreserved base quantity", async () => {
  const { user, onOpenChange } = await fill();
  await user.click(
    screen.getByRole("button", { name: "Review status change" }),
  );
  expect(postInventoryStatusChange).not.toHaveBeenCalled();
  await user.click(
    screen.getByRole("button", { name: "Confirm status change" }),
  );
  expect(postInventoryStatusChange).toHaveBeenCalledWith(
    expect.objectContaining({
      source_balance_id: "BAL-1",
      target_inventory_status_id: "hold",
      quantity: "2",
      serial_ids: [],
      expected_version: 9,
      reason_code: "STATUS_HOLD",
      notes: "Packaging review",
      operation_key: expect.stringMatching(/^stock\.status-change\./),
    }),
  );
  await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
});
it("retains the exact command and retry key when the API blocks posting", async () => {
  vi.mocked(postInventoryStatusChange).mockRejectedValue(
    new Error("Grand opname has frozen this stock."),
  );
  const { user, onOpenChange } = await fill();
  await user.click(
    screen.getByRole("button", { name: "Review status change" }),
  );
  await user.click(
    screen.getByRole("button", { name: "Confirm status change" }),
  );
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Grand opname has frozen this stock.",
  );
  expect(onOpenChange).not.toHaveBeenCalled();
  await user.click(
    screen.getByRole("button", { name: "Confirm status change" }),
  );
  await waitFor(() =>
    expect(postInventoryStatusChange).toHaveBeenCalledTimes(2),
  );
  expect(vi.mocked(postInventoryStatusChange).mock.calls[0][0]).toEqual(
    vi.mocked(postInventoryStatusChange).mock.calls[1][0],
  );
});
it("blocks quantities above unreserved stock before confirmation", async () => {
  const { user } = await fill();
  const quantity = screen.getByLabelText(/Quantity/);
  await user.clear(quantity);
  await user.type(quantity, "8");
  await user.click(
    screen.getByRole("button", { name: "Review status change" }),
  );
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "cannot exceed 7.000000",
  );
  expect(
    screen.queryByRole("button", { name: "Confirm status change" }),
  ).not.toBeInTheDocument();
  expect(postInventoryStatusChange).not.toHaveBeenCalled();
});

it("selects a scoped existing serial and posts exactly one base unit", async () => {
  vi.mocked(getBalance).mockResolvedValue({
    ...source,
    serial_controlled: true,
  });
  vi.mocked(listSerialStates).mockResolvedValue({
    items: [
      {
        serial_id: "serial-1",
        serial_no: "SN-1",
        balance: source,
        version_no: 1,
        updated_at: "2026-10-06",
      },
    ],
    page: 1,
    page_size: 100,
    total_items: 1,
    total_pages: 1,
  });
  const { user } = await fill();
  expect(screen.getByLabelText(/Quantity/)).toBeDisabled();
  expect(screen.getByLabelText(/Quantity/)).toHaveValue("1");
  await user.selectOptions(
    screen.getByRole("combobox", { name: "Serial number to change status" }),
    "serial-1",
  );
  await user.click(
    screen.getByRole("button", { name: "Review status change" }),
  );
  await user.click(
    screen.getByRole("button", { name: "Confirm status change" }),
  );
  expect(postInventoryStatusChange).toHaveBeenCalledWith(
    expect.objectContaining({
      quantity: "1",
      serial_ids: ["serial-1"],
      expected_version: 9,
    }),
  );
  expect(listSerialStates).toHaveBeenCalledWith(
    expect.objectContaining({
      ownerId: "owner-1",
      warehouseId: "warehouse-1",
      itemId: "item-1",
      locationId: "location-1",
      inventoryStatusId: "available",
    }),
  );
});
