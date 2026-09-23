import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import Decimal from "decimal.js";
import { toast } from "sonner";
import { ApiError } from "@/lib/api/client";
import {
  createDisposition,
  getQuarantineCase,
  listDispositionTypes,
  listQuarantineTargets,
} from "./quarantine-api";
import { QuarantineDetailDialog } from "./quarantine-detail-dialog";
import { testCase, testTypes } from "./quarantine-test-fixtures";
import type { QuarantineCase } from "./quarantine-types";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("./quarantine-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./quarantine-api")>()),
  createDisposition: vi.fn(),
  getQuarantineCase: vi.fn(),
  listDispositionTypes: vi.fn(),
  listQuarantineTargets: vi.fn(),
}));
let serverCase: QuarantineCase;
beforeEach(() => {
  vi.clearAllMocks();
  serverCase = { ...testCase };
  vi.mocked(getQuarantineCase).mockImplementation(() =>
    Promise.resolve(serverCase),
  );
  vi.mocked(listDispositionTypes).mockResolvedValue(testTypes);
  vi.mocked(listQuarantineTargets).mockResolvedValue({
    items: [
      {
        location_id: "target-1",
        code: "BULK-01",
        zone_code: "AMBIENT",
        location_type_code: "STORAGE",
      },
    ],
    page: 1,
    page_size: 20,
    total_pages: 1,
    total_items: 1,
  });
  vi.mocked(createDisposition).mockImplementation((_id, request) => {
    const disposed = new Decimal(serverCase.disposed_qty).plus(
      request.disposition_qty,
    );
    serverCase = {
      ...serverCase,
      disposed_qty: disposed.toString(),
      available_qty: new Decimal(serverCase.available_qty!)
        .minus(request.disposition_qty)
        .toString(),
      status_code: disposed.eq(serverCase.quarantine_qty)
        ? "CLOSED"
        : "PARTIALLY_DECIDED",
      version_no: serverCase.version_no + 1,
      quarantine_balance_version_no:
        serverCase.quarantine_balance_version_no! + 1,
      dispositions: [
        ...serverCase.dispositions,
        {
          quarantine_disposition_id: "DISP-1",
          quarantine_case_id: serverCase.quarantine_case_id,
          disposition_type_code: request.disposition_type_code,
          disposition_qty: request.disposition_qty,
          status_code: "PROCESSED",
          uom_id: "uom-1",
          decided_at: request.decided_at,
          decided_by: "account-1",
          created_at: request.decided_at,
          decision_notes: request.decision_notes,
          client_decision_reference: request.client_decision_reference,
          target_location_id: request.target_location_id,
          target_location_code: request.target_location_id ? "BULK-01" : null,
          inventory_movement_id: "MOVE-1",
          ...(request.disposition_type_code === "REWORK"
            ? {
                rework_task: {
                  rework_task_id: "REWORK-1",
                  task_status_code: "OPEN",
                  planned_qty: request.disposition_qty,
                  completed_qty: "0",
                  work_instructions: request.work_instructions,
                },
              }
            : {}),
        },
      ],
    };
    return Promise.resolve(serverCase);
  });
});
afterEach(cleanup);
function setup(canDispose = true) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  render(
    <QueryClientProvider client={client}>
      <QuarantineDetailDialog
        caseId="QCASE-1"
        capabilities={{ canDispose, timezone: "Asia/Jakarta" }}
        onOpenChange={vi.fn()}
      />
    </QueryClientProvider>,
  );
  return userEvent.setup();
}
async function choose(user: ReturnType<typeof userEvent.setup>, code: string) {
  await user.click(
    await screen.findByRole("combobox", { name: "Disposition type" }),
  );
  await user.click(
    await screen.findByRole("option", {
      name: testTypes.find((row) => row.code === code)!.name + ` (${code})`,
    }),
  );
}
async function quantity(
  user: ReturnType<typeof userEvent.setup>,
  amount: string,
) {
  const input = screen.getByLabelText(/Quantity \(base units\)/);
  await user.clear(input);
  await user.type(input, amount);
}
describe("quarantine decisions", () => {
  it("keeps read-only accounts from making decisions or requesting mutation lookups", async () => {
    setup(false);
    await screen.findByText(/read-only quarantine access/);
    expect(
      screen.queryByRole("button", { name: "Review decision" }),
    ).not.toBeInTheDocument();
    expect(listDispositionTypes).not.toHaveBeenCalled();
    expect(listQuarantineTargets).not.toHaveBeenCalled();
  });
  it.each(["RETURN", "DISPOSE"])(
    "confirms %s stock removal and closes a fully decided case",
    async (code) => {
      const user = setup();
      await choose(user, code);
      await user.click(screen.getByRole("button", { name: "Review decision" }));
      expect(createDisposition).not.toHaveBeenCalled();
      await user.click(
        screen.getByRole("button", { name: "Confirm disposition" }),
      );
      await screen.findByText(/fully decided and closed/);
      expect(createDisposition).toHaveBeenCalledTimes(1);
      expect(createDisposition).toHaveBeenCalledWith("QCASE-1", {
        expected_case_version: 2,
        expected_balance_version: 7,
        disposition_type_code: code,
        disposition_qty: "10",
        business_date: expect.stringMatching(/^\d{4}-\d{2}-\d{2}$/),
        decided_at: expect.stringMatching(/Z$/),
      });
      expect(
        screen.queryByRole("button", { name: "Review decision" }),
      ).not.toBeInTheDocument();
      expect(
        screen.getByText("Target location").nextElementSibling,
      ).toHaveTextContent("—");
    },
  );
  it("requires a strategy-filtered target before accepting stock into storage", async () => {
    const user = setup();
    await choose(user, "ACCEPT");
    await screen.findByRole("combobox", { name: "Acceptance target" });
    await user.click(screen.getByRole("button", { name: "Review decision" }));
    await screen.findByText(/Select a strategy-eligible storage location/);
    expect(createDisposition).not.toHaveBeenCalled();
    expect(listQuarantineTargets).toHaveBeenCalledWith("QCASE-1", {
      search: "",
      page: 1,
      pageSize: 20,
    });
    await user.click(
      screen.getByRole("combobox", { name: "Acceptance target" }),
    );
    await user.click(
      await screen.findByRole("option", { name: "BULK-01 · AMBIENT" }),
    );
    await user.click(screen.getByRole("button", { name: "Review decision" }));
    expect(
      screen.getByText(/Confirm ACCEPT for 10 EA into BULK-01/),
    ).toBeInTheDocument();
    await user.click(
      screen.getByRole("button", { name: "Confirm disposition" }),
    );
    await waitFor(() =>
      expect(createDisposition).toHaveBeenCalledWith(
        "QCASE-1",
        expect.objectContaining({
          target_location_id: "target-1",
          disposition_type_code: "ACCEPT",
        }),
      ),
    );
    expect(await screen.findByText("BULK-01")).toBeInTheDocument();
    expect(
      screen.getByText("Target location").nextElementSibling,
    ).toHaveTextContent("BULK-01");
    expect(screen.queryByText("Target location ID")).not.toBeInTheDocument();
    expect(screen.queryByText("target-1")).not.toBeInTheDocument();
  });
  it("blocks excessive quantities inline and with a toast", async () => {
    const user = setup();
    await choose(user, "RETURN");
    await quantity(user, "11");
    await user.click(screen.getByRole("button", { name: "Review decision" }));
    await screen.findByText(/Quantity exceeds the undecided amount/);
    expect(toast.error).toHaveBeenCalledWith(
      "Please check the highlighted fields.",
    );
    expect(createDisposition).not.toHaveBeenCalled();
  });
  it("rejects split decisions for serial-numbered or handling-unit stock", async () => {
    serverCase = { ...testCase, is_indivisible: true };
    const user = setup();
    await choose(user, "RETURN");
    await quantity(user, "2");
    await user.click(screen.getByRole("button", { name: "Review decision" }));
    await screen.findByText(/must be processed in full/);
    expect(createDisposition).not.toHaveBeenCalled();
  });
  it("requires rework instructions and retains partial-case history and the rework task", async () => {
    const user = setup();
    await choose(user, "REWORK");
    await quantity(user, "2");
    await user.click(screen.getByRole("button", { name: "Review decision" }));
    await screen.findByText(/Work instructions are required for rework/);
    await user.type(
      screen.getByLabelText(/Work instructions/),
      " Replace seal ",
    );
    await user.click(screen.getByRole("button", { name: "Review decision" }));
    await user.click(
      screen.getByRole("button", { name: "Confirm disposition" }),
    );
    await screen.findByText(/Rework: REWORK-1/);
    expect(
      screen.getByRole("link", { name: "Open rework task" }),
    ).toHaveAttribute(
      "href",
      "/inbound/rework?owner=owner-1&warehouse=warehouse-1&task=REWORK-1",
    );
    expect(createDisposition).toHaveBeenCalledWith(
      "QCASE-1",
      expect.objectContaining({
        work_instructions: "Replace seal",
        disposition_qty: "2",
        expected_case_version: 2,
        expected_balance_version: 7,
      }),
    );
    expect(screen.getByText("Partially decided")).toBeInTheDocument();
    expect(screen.getByText(/Undecided: 8 EA/)).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Disposition history (1)" }),
    ).toBeInTheDocument();
  });
  it("omits stale target fields when switching from acceptance to removal", async () => {
    const user = setup();
    await choose(user, "ACCEPT");
    const target = await screen.findByRole("combobox", {
      name: "Acceptance target",
    });
    await waitFor(() => expect(target).not.toBeDisabled());
    await user.click(target);
    await user.click(
      await screen.findByRole("option", { name: "BULK-01 · AMBIENT" }),
    );
    await choose(user, "RETURN");
    expect(
      screen.queryByRole("combobox", { name: "Acceptance target" }),
    ).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Review decision" }));
    await user.click(
      screen.getByRole("button", { name: "Confirm disposition" }),
    );
    await waitFor(() => expect(createDisposition).toHaveBeenCalled());
    expect(vi.mocked(createDisposition).mock.calls[0][1]).not.toHaveProperty(
      "target_location_id",
    );
    expect(vi.mocked(createDisposition).mock.calls[0][1]).not.toHaveProperty(
      "work_instructions",
    );
  });
  it("keeps concurrency errors visible without exposing internal request IDs", async () => {
    vi.mocked(createDisposition).mockRejectedValue(
      new ApiError("Stock changed", 409, undefined, "internal-request-id"),
    );
    const user = setup();
    await choose(user, "RETURN");
    await user.click(screen.getByRole("button", { name: "Review decision" }));
    await user.click(
      screen.getByRole("button", { name: "Confirm disposition" }),
    );
    await screen.findByRole("alert");
    expect(
      screen.getByText(/Refresh this case and review the latest quantities/),
    ).toBeInTheDocument();
    expect(screen.queryByText(/internal-request-id/)).not.toBeInTheDocument();
    expect(toast.error).toHaveBeenCalledWith("Stock changed");
  });
  it("blocks recording when the backend stock snapshot is missing", async () => {
    serverCase = { ...testCase, quarantine_balance_version_no: undefined };
    setup();
    await screen.findByText(
      /Stock version or batch information is unavailable/,
    );
    expect(
      screen.getByRole("button", { name: "Review decision" }),
    ).toBeDisabled();
    expect(createDisposition).not.toHaveBeenCalled();
  });
  it("shows closed-case lineage and links to the completed rework's reinspection", async () => {
    serverCase = {
      ...testCase,
      status_code: "CLOSED",
      disposed_qty: "10",
      parent_quarantine_case_id: "PARENT-1",
      dispositions: [
        {
          quarantine_disposition_id: "DISP-1",
          quarantine_case_id: "QCASE-1",
          disposition_type_code: "REWORK",
          status_code: "PROCESSED",
          disposition_qty: "10",
          uom_id: "uom-1",
          decided_at: "2026-09-17T03:00:00Z",
          decided_by: "account-1",
          created_at: "2026-09-17T03:00:00Z",
          rework_task: {
            rework_task_id: "REWORK-1",
            task_status_code: "COMPLETED",
            planned_qty: "10",
            completed_qty: "10",
            reinspection_id: "QC-2",
          },
        },
      ],
    };
    setup();
    expect(
      await screen.findByRole("link", { name: "Open reinspection" }),
    ).toHaveAttribute(
      "href",
      "/inbound/quality-inspections?owner=owner-1&warehouse=warehouse-1&inspection=QC-2",
    );
    expect(
      screen.getByRole("link", { name: "Open parent quarantine case" }),
    ).toHaveAttribute(
      "href",
      "/inbound/quarantine?owner=owner-1&warehouse=warehouse-1&case=PARENT-1",
    );
    expect(listDispositionTypes).not.toHaveBeenCalled();
  });
});
