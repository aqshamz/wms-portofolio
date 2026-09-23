import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { ApiError } from "@/lib/api/client";
import { getInboundException } from "./inbound-exception-api";
import { ExceptionDetailDialog } from "./exception-detail-dialog";
import { testException } from "./exception-test-fixtures";

vi.mock("./inbound-exception-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./inbound-exception-api")>()),
  getInboundException: vi.fn(),
}));
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getInboundException).mockResolvedValue(testException);
});
afterEach(cleanup);
function mount(onOpenChange = vi.fn()) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <ExceptionDetailDialog
        exceptionId="IEX-1"
        timezone="Asia/Jakarta"
        onOpenChange={onOpenChange}
      />
    </QueryClientProvider>,
  );
  return onOpenChange;
}
it("renders source references, signed quantities, notes and read-only audit semantics", async () => {
  mount();
  await screen.findByText("RCV-1");
  expect(screen.getByText("RCV-1-L001")).toBeInTheDocument();
  expect(screen.getByText("WMS Administrator")).toBeInTheDocument();
  expect(screen.getByText("Science In Sport")).toBeInTheDocument();
  expect(screen.getByText("Bandung WH")).toBeInTheDocument();
  for (const id of [
    testException.created_by,
    testException.owner_id,
    testException.warehouse_id,
  ])
    expect(screen.queryByText(id)).not.toBeInTheDocument();
  expect(screen.getByText("-2.000000")).toBeInTheDocument();
  expect(screen.getByText(/Two damaged cartons/)).toBeInTheDocument();
  expect(
    screen.getByText(/immutable audit record, not an open task/),
  ).toBeInTheDocument();
  for (const name of [/create/i, /edit/i, /delete/i, /resolve/i])
    expect(screen.queryByRole("button", { name })).not.toBeInTheDocument();
});
it("shows absent cancellation quantities as unavailable, not zero", async () => {
  vi.mocked(getInboundException).mockResolvedValue({
    ...testException,
    exception_type_code: "CANCELLATION",
    source_line_id: null,
    expected_qty: null,
    actual_qty: null,
    variance_qty: null,
    notes: null,
  });
  mount();
  await screen.findByText("Cancellation");
  expect(screen.getAllByText("—")).toHaveLength(4);
  expect(screen.getByText("No notes recorded.")).toBeInTheDocument();
});
it("preserves recorded zero quantities", async () => {
  vi.mocked(getInboundException).mockResolvedValue({
    ...testException,
    actual_qty: "0.000000",
  });
  mount();
  await screen.findByText("0.000000");
});
it("shows unavailable names without exposing internal IDs", async () => {
  vi.mocked(getInboundException).mockResolvedValue({
    ...testException,
    created_by_display_name: null,
    owner_name: null,
    warehouse_name: null,
  });
  mount();
  await screen.findByText("RCV-1");
  for (const label of ["Recorded by", "Owner", "Warehouse"])
    expect(screen.getByText(label).nextElementSibling).toHaveTextContent("—");
  for (const id of [
    testException.created_by,
    testException.owner_id,
    testException.warehouse_id,
  ])
    expect(screen.queryByText(id)).not.toBeInTheDocument();
});
it("keeps errors visible, hides request IDs and allows retrying the scoped read", async () => {
  const user = userEvent.setup();
  vi.mocked(getInboundException).mockRejectedValueOnce(
    new ApiError("Access denied", 403, undefined, "private-request-id"),
  );
  mount();
  await screen.findByRole("alert");
  expect(screen.queryByText(/private-request-id/)).not.toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Refresh exception" }));
  await screen.findByText("RCV-1");
  expect(getInboundException).toHaveBeenCalledTimes(2);
  expect(getInboundException).toHaveBeenLastCalledWith("IEX-1");
});
it("lets users close the read-only dialog", async () => {
  const user = userEvent.setup();
  const close = mount();
  await screen.findByText("RCV-1");
  await user.click(
    screen.getByRole("button", { name: "Close exception details" }),
  );
  await waitFor(() => expect(close).toHaveBeenCalledWith(false));
});
