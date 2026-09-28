import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { toast } from "sonner";
import {
  cancelVendorReturn,
  completeVendorReturn,
  getVendorReturn,
} from "./vendor-return-api";
import { VendorReturnDetailDialog } from "./vendor-return-detail-dialog";
import {
  testVendorReturn,
  vendorReturnCapabilities,
} from "./vendor-return-test-fixtures";
import type { VendorReturn } from "./vendor-return-types";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("./vendor-return-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./vendor-return-api")>()),
  getVendorReturn: vi.fn(),
  completeVendorReturn: vi.fn(),
  cancelVendorReturn: vi.fn(),
}));

let current: VendorReturn;
beforeEach(() => {
  vi.clearAllMocks();
  current = { ...testVendorReturn };
  vi.mocked(getVendorReturn).mockImplementation(async () => current);
  vi.mocked(completeVendorReturn).mockImplementation(async () => {
    current = {
      ...current,
      status_code: "COMPLETED",
      inventory_movement_id: "MOV-RTV-1",
      version_no: 2,
    };
    return current;
  });
  vi.mocked(cancelVendorReturn).mockImplementation(async (_id, fields) => {
    current = {
      ...current,
      status_code: "CANCELLED",
      cancellation_reason: fields.reason,
      version_no: 2,
    };
    return current;
  });
});
afterEach(cleanup);

function mount() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <VendorReturnDetailDialog
        id={testVendorReturn.vendor_return_id}
        capabilities={vendorReturnCapabilities}
        onOpenChange={vi.fn()}
      />
    </QueryClientProvider>,
  );
  return userEvent.setup();
}

it("shows the inherited vendor and confirms the stock-removal action", async () => {
  const user = mount();
  expect(await screen.findByText("VENDOR · Source Vendor")).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Complete return" }));
  expect(completeVendorReturn).not.toHaveBeenCalled();
  await user.click(
    screen.getByRole("button", { name: "Confirm vendor return" }),
  );
  expect(await screen.findByText("Completed")).toBeInTheDocument();
  expect(completeVendorReturn).toHaveBeenCalledWith(
    testVendorReturn.vendor_return_id,
    {
      expected_version: 1,
      expected_balance_version: 7,
      completed_at: expect.stringMatching(/Z$/),
    },
  );
  expect(toast.success).toHaveBeenCalledWith("Return to vendor completed.");
});

it("requires and trims a cancellation reason", async () => {
  const user = mount();
  await user.click(
    await screen.findByRole("button", { name: "Cancel transaction" }),
  );
  const confirm = screen.getByRole("button", { name: "Confirm cancellation" });
  expect(confirm).toBeDisabled();
  await user.type(
    screen.getByLabelText("Cancellation reason"),
    " Vendor rejected request ",
  );
  await user.click(confirm);
  expect(await screen.findByText("Cancelled")).toBeInTheDocument();
  expect(cancelVendorReturn).toHaveBeenCalledWith(
    testVendorReturn.vendor_return_id,
    {
      expected_version: 1,
      reason: "Vendor rejected request",
    },
  );
});
