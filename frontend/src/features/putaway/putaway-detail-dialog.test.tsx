import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";
import { ApiError } from "@/lib/api/client";
import {
  assignPutaway,
  cancelPutaway,
  completePutaway,
  getPutawayTask,
  listPutawayAssignees,
  listPutawayTargets,
  retargetPutaway,
  reversePutaway,
  startPutaway,
} from "./putaway-api";
import { PutawayDetailDialog } from "./putaway-detail-dialog";
import { testCapabilities, testTask } from "./putaway-test-fixtures";
import type { PutawayTask } from "./putaway-types";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("./putaway-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./putaway-api")>()),
  getPutawayTask: vi.fn(),
  listPutawayAssignees: vi.fn(),
  listPutawayTargets: vi.fn(),
  startPutaway: vi.fn(),
  assignPutaway: vi.fn(),
  retargetPutaway: vi.fn(),
  completePutaway: vi.fn(),
  cancelPutaway: vi.fn(),
  reversePutaway: vi.fn(),
}));

let serverTask: PutawayTask;
function update(values: Partial<PutawayTask>) {
  serverTask = {
    ...serverTask,
    ...values,
    version_no: serverTask.version_no + 1,
  };
  return Promise.resolve(serverTask);
}
beforeEach(() => {
  vi.clearAllMocks();
  serverTask = { ...testTask };
  vi.mocked(getPutawayTask).mockImplementation(() =>
    Promise.resolve(serverTask),
  );
  vi.mocked(startPutaway).mockImplementation(() =>
    update({
      task_status_code: "IN_PROGRESS",
      assigned_to: "account-1",
      assigned_display_name: "Worker One",
    }),
  );
  vi.mocked(assignPutaway).mockImplementation((_id, _version, account) =>
    update({
      task_status_code: "ASSIGNED",
      assigned_to: account,
      assigned_display_name: "Worker Two",
    }),
  );
  vi.mocked(retargetPutaway).mockImplementation((_id, _version, target) =>
    update({ target_location_id: target, target_location_code: "BULK-02" }),
  );
  vi.mocked(completePutaway).mockImplementation(() =>
    update({
      task_status_code: "COMPLETED",
      completed_qty: "10",
      resulting_balance_id: "balance-2",
      result_balance_version_no: 9,
      inventory_movement_id: "MOV-1",
    }),
  );
  vi.mocked(cancelPutaway).mockImplementation(() =>
    update({
      task_status_code: "CANCELLED",
      replacement_inspection_id: "QC-2",
    }),
  );
  vi.mocked(reversePutaway).mockImplementation(() =>
    update({
      task_status_code: "REVERSED",
      replacement_inspection_id: "QC-2",
      reversal_movement_id: "MOV-2",
      reversal_reason: "Wrong storage slot",
    }),
  );
  vi.mocked(listPutawayAssignees).mockResolvedValue({
    items: [
      {
        account_id: "worker-2",
        username: "worker.two",
        display_name: "Worker Two",
      },
    ],
    page: 1,
    page_size: 20,
    total_pages: 1,
    total_items: 1,
  });
  vi.mocked(listPutawayTargets).mockResolvedValue({
    items: [
      {
        location_id: "target-2",
        code: "BULK-02",
        zone_code: "AMBIENT",
        location_type_code: "BULK",
      },
    ],
    page: 1,
    page_size: 20,
    total_pages: 1,
    total_items: 1,
  });
});
afterEach(cleanup);
function setup(capabilities = testCapabilities) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  render(
    <QueryClientProvider client={client}>
      <PutawayDetailDialog
        taskId="PUT-1"
        capabilities={capabilities}
        onOpenChange={vi.fn()}
      />
    </QueryClientProvider>,
  );
  return userEvent.setup();
}

describe("putaway task workflow", () => {
  it("claims an open task then completes the full physical move using both versions", async () => {
    const user = setup({
      ...testCapabilities,
      canAssign: false,
      canCancel: false,
    });
    await user.click(await screen.findByRole("button", { name: "Start task" }));
    await user.click(
      screen.getByRole("button", { name: "Confirm start task" }),
    );
    await waitFor(() => expect(startPutaway).toHaveBeenCalledWith("PUT-1", 1));
    await user.click(
      await screen.findByRole("button", { name: "Complete putaway" }),
    );
    expect(
      screen.getByText(/Confirm that the full 10 EA has physically moved/),
    ).toBeInTheDocument();
    await user.click(
      screen.getByRole("button", { name: "Confirm complete putaway" }),
    );
    await waitFor(() =>
      expect(completePutaway).toHaveBeenCalledWith("PUT-1", {
        expected_version: 2,
        expected_balance_version: 7,
        business_date: expect.stringMatching(/^\d{4}-\d{2}-\d{2}$/),
      }),
    );
    expect(
      await screen.findByText(/The movement has been posted/),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Complete putaway" }),
    ).not.toBeInTheDocument();
  });
  it("does not let a putaway worker start a task assigned to someone else", async () => {
    serverTask = {
      ...testTask,
      task_status_code: "ASSIGNED",
      assigned_to: "worker-2",
    };
    setup({ ...testCapabilities, canAssign: false, canCancel: false });
    await screen.findByText(/No actions are available/);
    expect(
      screen.queryByRole("button", { name: "Start task" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Complete putaway" }),
    ).not.toBeInTheDocument();
    expect(startPutaway).not.toHaveBeenCalled();
    expect(completePutaway).not.toHaveBeenCalled();
  });
  it("explains when there are no permitted scoped assignees", async () => {
    vi.mocked(listPutawayAssignees).mockResolvedValue({
      items: [],
      page: 1,
      page_size: 20,
      total_pages: 0,
      total_items: 0,
    });
    const user = setup();
    await user.click(
      await screen.findByRole("button", { name: "Assign account" }),
    );
    expect(
      await screen.findByText(
        /No active accounts with INBOUND.PUTAWAY permission/,
      ),
    ).toBeInTheDocument();
    expect(assignPutaway).not.toHaveBeenCalled();
  });
  it("hides all mutations for a read-only account", async () => {
    setup({
      ...testCapabilities,
      canPutaway: false,
      canAssign: false,
      canCancel: false,
    });
    expect(
      await screen.findByText(/No actions are available/),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start task" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Assign account" }),
    ).not.toBeInTheDocument();
    expect(listPutawayAssignees).not.toHaveBeenCalled();
    expect(listPutawayTargets).not.toHaveBeenCalled();
  });
  it("does not allow another account to complete or cancel in-progress work", async () => {
    serverTask = {
      ...testTask,
      task_status_code: "IN_PROGRESS",
      assigned_to: "other",
    };
    setup();
    expect(
      await screen.findByText(/Only its assignee can complete or cancel/),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Complete putaway" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Cancel task" }),
    ).not.toBeInTheDocument();
  });
  it("requires a cancellation reason and opens the retained replacement inspection", async () => {
    const user = setup();
    await user.click(
      await screen.findByRole("button", { name: "Cancel task" }),
    );
    await user.click(
      screen.getByRole("button", { name: "Confirm cancel task" }),
    );
    expect(
      await screen.findByText("A reason is required."),
    ).toBeInTheDocument();
    expect(cancelPutaway).not.toHaveBeenCalled();
    await user.type(screen.getByLabelText(/Reason/), "Wrong plan");
    await user.click(
      screen.getByRole("button", { name: "Confirm cancel task" }),
    );
    await waitFor(() =>
      expect(cancelPutaway).toHaveBeenCalledWith(
        "PUT-1",
        expect.objectContaining({
          expected_version: 1,
          expected_balance_version: 7,
          reason: "Wrong plan",
        }),
      ),
    );
    const link = await screen.findByRole("link", {
      name: "Open replacement inspection",
    });
    expect(link).toHaveAttribute(
      "href",
      "/inbound/quality-inspections?owner=owner-1&warehouse=wh-1&inspection=QC-2",
    );
    expect(
      screen.queryByRole("button", { name: "Start task" }),
    ).not.toBeInTheDocument();
  });
  it("reverses a completed task using its result balance version, not the source version", async () => {
    serverTask = {
      ...testTask,
      task_status_code: "COMPLETED",
      assigned_to: "account-1",
      completed_qty: "10",
      resulting_balance_id: "balance-2",
      result_balance_version_no: 9,
    };
    const user = setup();
    await user.click(
      await screen.findByRole("button", { name: "Reverse putaway" }),
    );
    await user.type(screen.getByLabelText(/Reason/), "Wrong storage slot");
    await user.click(
      screen.getByRole("button", { name: "Confirm reverse putaway" }),
    );
    await waitFor(() =>
      expect(reversePutaway).toHaveBeenCalledWith(
        "PUT-1",
        expect.objectContaining({
          expected_balance_version: 9,
          reason: "Wrong storage slot",
        }),
      ),
    );
    expect(
      await screen.findByRole("link", { name: "Open replacement inspection" }),
    ).toBeInTheDocument();
  });
  it("assigns from the task-scoped lookup", async () => {
    const user = setup();
    await user.click(
      await screen.findByRole("button", { name: "Assign account" }),
    );
    const select = screen.getByRole("combobox", { name: "Assigned account" });
    await waitFor(() => expect(select).toBeEnabled());
    select.focus();
    await user.keyboard(" ");
    await screen.findByRole("option", { name: "Worker Two (@worker.two)" });
    await user.keyboard("[Enter]");
    await user.click(
      screen.getByRole("button", { name: "Confirm assign account" }),
    );
    await waitFor(() =>
      expect(assignPutaway).toHaveBeenCalledWith("PUT-1", 1, "worker-2"),
    );
    expect(
      await screen.findByText(/Only the assigned account can start/),
    ).toBeInTheDocument();
  });
  it("retargets only from the strategy-filtered lookup", async () => {
    const user = setup();
    await user.click(
      await screen.findByRole("button", { name: "Change target" }),
    );
    const select = screen.getByRole("combobox", { name: "New putaway target" });
    await waitFor(() => expect(select).toBeEnabled());
    select.focus();
    await user.keyboard(" ");
    await screen.findByRole("option", { name: "BULK-02 · AMBIENT · BULK" });
    await user.keyboard("[Enter]");
    await user.click(
      screen.getByRole("button", { name: "Confirm change target" }),
    );
    await waitFor(() =>
      expect(retargetPutaway).toHaveBeenCalledWith("PUT-1", 1, "target-2"),
    );
    expect(await screen.findByText("BULK-02")).toBeInTheDocument();
  });
  it("keeps stale-stock errors visible without exposing internal request IDs", async () => {
    serverTask = {
      ...testTask,
      task_status_code: "IN_PROGRESS",
      assigned_to: "account-1",
    };
    vi.mocked(completePutaway).mockRejectedValue(
      new ApiError("Stock version changed", 409, undefined, "internal-request"),
    );
    const user = setup();
    await user.click(
      await screen.findByRole("button", { name: "Complete putaway" }),
    );
    await user.click(
      screen.getByRole("button", { name: "Confirm complete putaway" }),
    );
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Stock version changed",
    );
    expect(toast.error).toHaveBeenCalled();
    expect(screen.queryByText(/internal-request/)).not.toBeInTheDocument();
  });
  it("blocks posting when the stock version is unavailable", async () => {
    serverTask = {
      ...testTask,
      task_status_code: "IN_PROGRESS",
      assigned_to: "account-1",
      source_balance_version_no: undefined,
    };
    const user = setup();
    await user.click(
      await screen.findByRole("button", { name: "Complete putaway" }),
    );
    expect(
      await screen.findByText(/Stock version is unavailable/),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Confirm complete putaway" }),
    ).toBeDisabled();
    expect(completePutaway).not.toHaveBeenCalled();
  });
});
