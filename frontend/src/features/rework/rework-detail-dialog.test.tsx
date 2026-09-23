import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { toast } from "sonner";
import { ApiError } from "@/lib/api/client";
import { getReworkTask, transitionReworkTask } from "./rework-api";
import { ReworkDetailDialog } from "./rework-detail-dialog";
import { capabilities, testTask } from "./rework-test-fixtures";
import type { ReworkTask } from "./rework-types";
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("./rework-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./rework-api")>()),
  getReworkTask: vi.fn(),
  transitionReworkTask: vi.fn(),
}));
let current: ReworkTask;
beforeEach(() => {
  vi.clearAllMocks();
  current = { ...testTask };
  vi.mocked(getReworkTask).mockImplementation(async () => current);
  vi.mocked(transitionReworkTask).mockImplementation(
    async (_id, action, fields) => {
      current = {
        ...current,
        version_no: current.version_no + 1,
        assigned_to: "worker-1",
        assigned_username: "godown.warehouse",
        task_status_code: action === "start" ? "IN_PROGRESS" : "COMPLETED",
        completed_qty: action === "complete" ? current.planned_qty : "0.000000",
        result_notes: fields.result_notes || null,
        reinspection_id: action === "complete" ? "QIN-2" : null,
      };
      return current;
    },
  );
});
afterEach(cleanup);
function mount(canRework = true) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const close = vi.fn();
  render(
    <QueryClientProvider client={client}>
      <ReworkDetailDialog
        taskId="RWK-1"
        capabilities={{ ...capabilities, canRework }}
        onOpenChange={close}
      />
    </QueryClientProvider>,
  );
  return { user: userEvent.setup(), close };
}
it("supports confirmation, claiming, completion notes and the generated reinspection link", async () => {
  const { user } = mount();
  await user.click(await screen.findByRole("button", { name: "Start rework" }));
  await user.click(screen.getByRole("button", { name: "Review start" }));
  expect(transitionReworkTask).not.toHaveBeenCalled();
  await user.click(screen.getByRole("button", { name: "Confirm start" }));
  await user.click(
    await screen.findByRole("button", { name: "Complete rework" }),
  );
  expect(transitionReworkTask).toHaveBeenLastCalledWith("RWK-1", "start", {
    expected_version: 1,
  });
  await user.type(screen.getByLabelText("Result notes"), " Seals replaced ");
  await user.click(screen.getByRole("button", { name: "Review completion" }));
  expect(transitionReworkTask).toHaveBeenCalledTimes(1);
  await user.click(screen.getByRole("button", { name: "Confirm completion" }));
  const link = await screen.findByRole("link", { name: "Open reinspection" });
  expect(link).toHaveAttribute(
    "href",
    "/inbound/quality-inspections?owner=owner-1&warehouse=warehouse-1&inspection=QIN-2",
  );
  expect(transitionReworkTask).toHaveBeenLastCalledWith("RWK-1", "complete", {
    expected_version: 2,
    result_notes: "Seals replaced",
  });
  expect(screen.getByText(/stock is not available yet/)).toBeInTheDocument();
  expect(screen.getByText("@godown.warehouse (you)")).toBeInTheDocument();
  expect(screen.queryByText("worker-1")).not.toBeInTheDocument();
});
it("keeps read-only users from execution but shows instructions and lineage", async () => {
  mount(false);
  await screen.findByText("Replace damaged seals");
  expect(
    screen.queryByRole("button", { name: "Start rework" }),
  ).not.toBeInTheDocument();
  expect(
    screen.getByRole("link", { name: "Open quarantine case" }),
  ).toHaveAttribute(
    "href",
    "/inbound/quarantine?owner=owner-1&warehouse=warehouse-1&case=QCASE-1",
  );
  expect(transitionReworkTask).not.toHaveBeenCalled();
});
it("shows another assignee's username without exposing their ID", async () => {
  current = {
    ...testTask,
    task_status_code: "COMPLETED",
    assigned_to: "other-account-id",
    assigned_username: "godown.warehouse",
  };
  mount(false);
  expect(await screen.findByText("@godown.warehouse")).toBeInTheDocument();
  expect(screen.queryByText("other-account-id")).not.toBeInTheDocument();
});
it.each(["ASSIGNED", "IN_PROGRESS"])(
  "blocks another worker's %s task",
  async (status) => {
    current = { ...testTask, task_status_code: status, assigned_to: "other" };
    mount();
    await screen.findByText(/Only its assigned worker can process it/);
    expect(
      screen.queryByRole("button", { name: /Start rework|Complete rework/ }),
    ).not.toBeInTheDocument();
  },
);
it("allows starting a task already assigned to the current worker", async () => {
  current = {
    ...testTask,
    task_status_code: "ASSIGNED",
    assigned_to: "worker-1",
  };
  mount();
  await screen.findByRole("button", { name: "Start rework" });
});
it("validates long notes inline and with a toast without posting", async () => {
  current = {
    ...testTask,
    task_status_code: "IN_PROGRESS",
    assigned_to: "worker-1",
  };
  const { user } = mount();
  await user.click(
    await screen.findByRole("button", { name: "Complete rework" }),
  );
  await user.click(screen.getByLabelText("Result notes"));
  await user.paste("a".repeat(4001));
  await user.click(screen.getByRole("button", { name: "Review completion" }));
  await screen.findByText("Maximum 4000 characters.");
  expect(toast.error).toHaveBeenCalled();
  expect(transitionReworkTask).not.toHaveBeenCalled();
});
it("omits optional blank completion notes and cannot partially complete quantity", async () => {
  current = {
    ...testTask,
    task_status_code: "IN_PROGRESS",
    assigned_to: "worker-1",
  };
  const { user } = mount();
  await user.click(
    await screen.findByRole("button", { name: "Complete rework" }),
  );
  expect(screen.queryByRole("spinbutton")).not.toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Review completion" }));
  await user.click(screen.getByRole("button", { name: "Confirm completion" }));
  await screen.findByRole("link", { name: "Open reinspection" });
  expect(transitionReworkTask).toHaveBeenLastCalledWith("RWK-1", "complete", {
    expected_version: 1,
  });
});
it("displays stale-version errors without request IDs and allows a fresh snapshot", async () => {
  vi.mocked(transitionReworkTask).mockRejectedValueOnce(
    new ApiError("Task changed", 409, undefined, "private-id"),
  );
  const { user } = mount();
  await user.click(await screen.findByRole("button", { name: "Start rework" }));
  await user.click(screen.getByRole("button", { name: "Review start" }));
  await user.click(screen.getByRole("button", { name: "Confirm start" }));
  await screen.findByText(/Refresh the task before trying again/);
  expect(screen.queryByText(/private-id/)).not.toBeInTheDocument();
  current = { ...testTask, version_no: 3 };
  await user.click(screen.getByRole("button", { name: "Refresh task" }));
  await user.click(await screen.findByRole("button", { name: "Start rework" }));
  await user.click(screen.getByRole("button", { name: "Review start" }));
  await user.click(screen.getByRole("button", { name: "Confirm start" }));
  await waitFor(() =>
    expect(transitionReworkTask).toHaveBeenLastCalledWith("RWK-1", "start", {
      expected_version: 3,
    }),
  );
});
it("prevents dismissal and duplicate submissions while processing", async () => {
  let finish!: (task: ReworkTask) => void;
  vi.mocked(transitionReworkTask).mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  );
  const { user, close } = mount();
  await user.click(await screen.findByRole("button", { name: "Start rework" }));
  await user.click(screen.getByRole("button", { name: "Review start" }));
  await user.dblClick(screen.getByRole("button", { name: "Confirm start" }));
  expect(
    screen.getByRole("button", { name: "Close rework task" }),
  ).toBeDisabled();
  await user.keyboard("{Escape}");
  expect(close).not.toHaveBeenCalled();
  expect(transitionReworkTask).toHaveBeenCalledTimes(1);
  current = {
    ...testTask,
    task_status_code: "IN_PROGRESS",
    assigned_to: "worker-1",
    version_no: 2,
  };
  finish(current);
  await screen.findByRole("button", { name: "Complete rework" });
});
