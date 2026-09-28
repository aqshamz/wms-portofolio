import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { toast } from "sonner";
import { cancelDisposal, completeDisposal, getDisposal } from "./disposal-api";
import { DisposalDetailDialog } from "./disposal-detail-dialog";
import { disposalCapabilities, testDisposal } from "./disposal-test-fixtures";
import type { Disposal } from "./disposal-types";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("./disposal-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./disposal-api")>()),
  getDisposal: vi.fn(),
  completeDisposal: vi.fn(),
  cancelDisposal: vi.fn(),
}));

let current: Disposal;

beforeEach(() => {
  vi.clearAllMocks();
  current = { ...testDisposal };
  vi.mocked(getDisposal).mockImplementation(async () => current);
  vi.mocked(completeDisposal).mockImplementation(async () => {
    current = {
      ...current,
      status_code: "COMPLETED",
      completed_at: "2026-09-25T11:00:00+07:00",
      completed_by_display_name: "Warehouse Admin",
      inventory_movement_id: "MOV-1",
      source_balance_version_no: undefined,
      version_no: current.version_no + 1,
    };
    return current;
  });
  vi.mocked(cancelDisposal).mockImplementation(async (_id, fields) => {
    current = {
      ...current,
      status_code: "CANCELLED",
      cancelled_at: "2026-09-25T11:00:00+07:00",
      cancelled_by_display_name: "Warehouse Admin",
      cancellation_reason: fields.reason,
      source_balance_version_no: undefined,
      version_no: current.version_no + 1,
    };
    return current;
  });
});

afterEach(cleanup);

function mount(capabilities = disposalCapabilities) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <DisposalDetailDialog
        id={testDisposal.disposal_id}
        capabilities={capabilities}
        onOpenChange={vi.fn()}
      />
    </QueryClientProvider>,
  );
  return userEvent.setup();
}

it("requires a second confirmation and posts both optimistic versions", async () => {
  const user = mount();
  await user.click(
    await screen.findByRole("button", { name: "Complete disposal" }),
  );
  expect(completeDisposal).not.toHaveBeenCalled();
  await user.click(
    screen.getByRole("button", { name: "Confirm permanent disposal" }),
  );
  expect(await screen.findByText("Completed")).toBeInTheDocument();
  expect(completeDisposal).toHaveBeenCalledWith(testDisposal.disposal_id, {
    expected_version: 1,
    expected_balance_version: 7,
    completed_at: expect.stringMatching(/Z$/),
  });
  expect(toast.success).toHaveBeenCalledWith("Disposal completed.");
  expect(screen.getByText("MOV-1")).toBeInTheDocument();
});

it("requires a reason to cancel and hides lifecycle actions from read-only users", async () => {
  const user = mount();
  await user.click(
    await screen.findByRole("button", { name: "Cancel transaction" }),
  );
  const confirm = screen.getByRole("button", { name: "Confirm cancellation" });
  expect(confirm).toBeDisabled();
  await user.type(
    screen.getByLabelText("Cancellation reason"),
    " Retain stock ",
  );
  await user.click(confirm);
  expect(await screen.findByText("Cancelled")).toBeInTheDocument();
  expect(cancelDisposal).toHaveBeenCalledWith(testDisposal.disposal_id, {
    expected_version: 1,
    reason: "Retain stock",
  });

  cleanup();
  current = { ...testDisposal };
  mount({ ...disposalCapabilities, canComplete: false, canCancel: false });
  await screen.findByText(
    testDisposal.item_code + " · " + testDisposal.item_name,
  );
  expect(
    screen.queryByRole("button", { name: "Complete disposal" }),
  ).not.toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Cancel transaction" }),
  ).not.toBeInTheDocument();
});
