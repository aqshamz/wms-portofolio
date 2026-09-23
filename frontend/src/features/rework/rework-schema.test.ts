import { expect, it } from "vitest";
import { reworkResultSchema } from "./rework-schema";
import { allowedReworkActions, reworkHref } from "./rework-types";
import { capabilities, testTask } from "./rework-test-fixtures";
it.each([
  ["OPEN", null, ["start"]],
  ["ASSIGNED", "worker-1", ["start"]],
  ["ASSIGNED", "other", []],
  ["ASSIGNED", null, []],
  ["IN_PROGRESS", "worker-1", ["complete"]],
  ["IN_PROGRESS", "other", []],
  ["IN_PROGRESS", null, []],
  ["COMPLETED", "worker-1", []],
  ["CANCELLED", "worker-1", []],
])("gates %s with assignee %s", (status, assignee, expected) => {
  expect(
    allowedReworkActions(
      {
        ...testTask,
        task_status_code: status as string,
        assigned_to: assignee as string | null,
      },
      capabilities,
    ),
  ).toEqual(expected);
});
it("blocks read-only, anonymous and invalid-version execution", () => {
  expect(
    allowedReworkActions(testTask, { ...capabilities, canRework: false }),
  ).toEqual([]);
  expect(
    allowedReworkActions(testTask, { ...capabilities, accountId: "" }),
  ).toEqual([]);
  expect(
    allowedReworkActions({ ...testTask, version_no: 0 }, capabilities),
  ).toEqual([]);
});
it("trims optional notes while enforcing the backend length limit", () => {
  expect(
    reworkResultSchema.parse({ result_notes: "  Seals replaced  " }),
  ).toEqual({ result_notes: "Seals replaced" });
  expect(reworkResultSchema.safeParse({ result_notes: "" }).success).toBe(true);
  expect(
    reworkResultSchema.safeParse({ result_notes: "a".repeat(4001) }).success,
  ).toBe(false);
});
it("creates a scope-preserving direct task link", () => {
  const url = new URL(
    reworkHref({ ...testTask, rework_task_id: "RWK/1" }),
    "http://localhost",
  );
  expect(url.pathname).toBe("/inbound/rework");
  expect(Object.fromEntries(url.searchParams)).toEqual({
    owner: "owner-1",
    warehouse: "warehouse-1",
    task: "RWK/1",
  });
});
