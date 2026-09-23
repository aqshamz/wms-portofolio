import { expect, it } from "vitest";
import { reworkAssignee } from "./rework-types";
import { testTask } from "./rework-test-fixtures";

it("keeps unassigned tasks distinct from unavailable account names", () => {
  expect(reworkAssignee(testTask, "worker-1")).toBe("Unassigned");
  expect(
    reworkAssignee({ ...testTask, assigned_to: "private-id" }, "worker-1"),
  ).toBe("Account unavailable");
});
it("uses usernames and keeps the current-user hint without exposing IDs", () => {
  const task = {
    ...testTask,
    assigned_to: "worker-1",
    assigned_username: "godown.warehouse",
  };
  expect(reworkAssignee(task, "worker-1")).toBe("@godown.warehouse (you)");
  expect(reworkAssignee(task, "other")).toBe("@godown.warehouse");
});
