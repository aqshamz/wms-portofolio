import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { cancelCycleCount, getCycleCount } from "./stock-control-api";
import { CycleCountDetailDialog } from "./cycle-count-detail-dialog";
import type { CycleCount } from "./stock-control-types";

vi.mock("@/components/auth/auth-provider", () => ({
  useAuth: () => ({ user: { account_id: "counter-1" }, can: () => true }),
}));
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("./stock-control-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./stock-control-api")>()),
  getCycleCount: vi.fn(),
  cancelCycleCount: vi.fn(),
}));

let current: CycleCount;
beforeEach(() => {
  vi.clearAllMocks();
  current = {
    cycle_count_id: "GSO-1",
    count_type_code: "GRAND",
    stock_frozen: true,
    status_code: "PARTIALLY_POSTED",
    owner_id: "owner-1",
    owner_code: "SIS",
    owner_name: "Science In Sport",
    warehouse_id: "warehouse-1",
    warehouse_code: "BDG_WH",
    warehouse_name: "Bandung WH",
    business_date: "2026-10-06",
    tolerance_quantity: "0.000000",
    blind_count: false,
    notes: null,
    total_lines: 2,
    open_lines: 1,
    counted_lines: 0,
    recount_lines: 0,
    final_lines: 1,
    lines: [],
    completed_at: null,
    cancelled_at: null,
    cancellation_reason: null,
    created_at: "2026-10-06T09:00:00+07:00",
    created_by: "counter-1",
    created_by_display_name: "Counter",
    version_no: 3,
  };
  vi.mocked(getCycleCount).mockImplementation(async () => current);
  vi.mocked(cancelCycleCount).mockImplementation(
    async (_id, _version, reason) => {
      current = {
        ...current,
        stock_frozen: false,
        status_code: "CANCELLED",
        cancelled_at: "2026-10-06T10:00:00+07:00",
        cancellation_reason: reason,
        version_no: 4,
      };
      return current;
    },
  );
});
afterEach(cleanup);

function mount() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <CycleCountDetailDialog
        cycleCountId={current.cycle_count_id}
        onOpenChange={vi.fn()}
      />
    </QueryClientProvider>,
  );
  return userEvent.setup();
}

it("keeps the customer and warehouse freeze visible during partial posting, then reflects cancellation", async () => {
  const user = mount();
  const status = await screen.findByRole("status");
  expect(status).toHaveTextContent("Stock frozen for this grand opname");
  expect(status).toHaveTextContent("Science In Sport in Bandung WH");
  expect(status).toHaveTextContent("Partial posting keeps the freeze active");

  await user.type(
    screen.getByLabelText("Cancellation reason"),
    "Recount later",
  );
  await user.click(screen.getByRole("button", { name: "Cancel remaining" }));

  expect(
    await screen.findByText("This grand opname’s freeze is released"),
  ).toBeInTheDocument();
  expect(screen.getByRole("status")).toHaveTextContent(
    "Any corrections already posted remain in inventory",
  );
  expect(cancelCycleCount).toHaveBeenCalledWith("GSO-1", 3, "Recount later");
});

it("shows release after all grand-opname lines are finalized", async () => {
  current = {
    ...current,
    stock_frozen: false,
    status_code: "POSTED",
    completed_at: "2026-10-06T10:00:00+07:00",
    final_lines: 2,
    open_lines: 0,
  };
  mount();
  expect(await screen.findByRole("status")).toHaveTextContent(
    "All count lines are finalized",
  );
  expect(
    screen.queryByRole("button", { name: "Cancel remaining" }),
  ).not.toBeInTheDocument();
});

it("does not present ordinary cycle counts as warehouse freezes", async () => {
  current = { ...current, count_type_code: "CYCLE", stock_frozen: false };
  mount();
  await screen.findByText("Bandung WH");
  expect(screen.queryByRole("status")).not.toBeInTheDocument();
});

it("preserves discovery notes and prevents a found-stock recorder from reconciling their own line", async () => {
  current = {
    ...current,
    created_by: "other-counter",
    status_code: "REVIEW",
    open_lines: 0,
    counted_lines: 1,
    lines: [
      {
        cycle_count_line_id: "GSO-1-L0001",
        line_no: 1,
        decision_code: "COUNTED",
        balance_id: null,
        found_stock: true,
        created_by: "counter-1",
        counted_by: "counter-1",
        created_by_display_name: "Counter",
        discovery_notes: "Found behind rack",
        snapshot_version_no: 0,
        item_code: "SERIAL-1",
        item_name: "Found serial item",
        location_code: "STORAGE-1",
        inventory_status_code: "AVAILABLE",
        uom_code: "EA",
        lot_number: null,
        handling_unit_barcode: null,
        serial_controlled: true,
        system_quantity: "0.000000",
        counted_quantity: "2.000000",
        variance_quantity: "2.000000",
        count_attempts: 2,
        requires_recount: false,
        counted_at: null,
        counted_by_display_name: "Counter",
        count_notes: null,
        decided_at: null,
        decided_by_display_name: null,
        decision_reason: null,
        inventory_movement_id: null,
        resulting_balance_id: null,
        current_serials: [],
        serial_reconciliations: [],
      },
    ],
  };
  mount();
  expect(await screen.findByText("Found behind rack")).toBeInTheDocument();
  expect(
    screen.getByText(
      "A different account must review the stock you recorded or counted.",
    ),
  ).toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Reconcile serials" }),
  ).not.toBeInTheDocument();
  expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
});
