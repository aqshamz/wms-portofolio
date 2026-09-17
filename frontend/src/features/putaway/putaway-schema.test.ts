import { describe, expect, it } from "vitest";
import { businessDateToday, putawayActionSchema } from "./putaway-schema";
import { allowedPutawayActions } from "./putaway-types";
import { testCapabilities, testTask } from "./putaway-test-fixtures";

const values = {
  business_date: "2026-09-17",
  reason: "Wrong destination",
  account_id: "worker-1",
  target_location_id: "target-2",
};

describe("putaway action validation", () => {
  it("requires an account or target for planning changes", () => {
    expect(
      putawayActionSchema("assign").safeParse({ ...values, account_id: "" })
        .success,
    ).toBe(false);
    expect(
      putawayActionSchema("retarget").safeParse({
        ...values,
        target_location_id: "",
      }).success,
    ).toBe(false);
    expect(
      putawayActionSchema("start").safeParse({
        business_date: "",
        reason: "",
        account_id: "",
        target_location_id: "",
      }).success,
    ).toBe(true);
  });
  it.each(["complete", "cancel", "reverse"] as const)(
    "validates the posting date for %s",
    (action) => {
      expect(putawayActionSchema(action).safeParse(values).success).toBe(true);
      expect(
        putawayActionSchema(action).safeParse({
          ...values,
          business_date: "2026-02-30",
        }).success,
      ).toBe(false);
      expect(
        putawayActionSchema(action).safeParse({ ...values, business_date: "" })
          .success,
      ).toBe(false);
    },
  );
  it.each(["cancel", "reverse"] as const)(
    "requires a bounded reason for %s",
    (action) => {
      expect(
        putawayActionSchema(action).safeParse({ ...values, reason: "  " })
          .success,
      ).toBe(false);
      expect(
        putawayActionSchema(action).safeParse({
          ...values,
          reason: "x".repeat(4001),
        }).success,
      ).toBe(false);
      expect(
        putawayActionSchema(action).parse({
          ...values,
          reason: "  Wrong slot  ",
        }).reason,
      ).toBe("Wrong slot");
    },
  );
  it("uses the user's timezone for business date", () => {
    const now = new Date("2026-09-16T18:00:00Z");
    expect(businessDateToday("Asia/Jakarta", now)).toBe("2026-09-17");
    expect(businessDateToday("UTC", now)).toBe("2026-09-16");
  });
  it("falls back safely for an invalid legacy timezone preference", () =>
    expect(
      businessDateToday("invalid-timezone", new Date("2026-09-16T18:00:00Z")),
    ).toBe("2026-09-17"));
});

describe("putaway permission and assignee rules", () => {
  it("allows planning, claiming and cancellation on an open task", () =>
    expect(allowedPutawayActions(testTask, testCapabilities)).toEqual([
      "assign",
      "retarget",
      "start",
      "cancel",
    ]));
  it("does not let another worker start an assigned task", () =>
    expect(
      allowedPutawayActions(
        { ...testTask, task_status_code: "ASSIGNED", assigned_to: "other" },
        { ...testCapabilities, canAssign: false, canCancel: false },
      ),
    ).toEqual([]));
  it("only permits the assignee to complete or cancel in-progress tasks", () => {
    const task = {
      ...testTask,
      task_status_code: "IN_PROGRESS" as const,
      assigned_to: "account-1",
    };
    expect(allowedPutawayActions(task, testCapabilities)).toEqual([
      "complete",
      "cancel",
    ]);
    expect(
      allowedPutawayActions(
        { ...task, assigned_to: "other" },
        testCapabilities,
      ),
    ).toEqual([]);
  });
  it("treats completed tasks as reversal-only and recovery states as final", () => {
    expect(
      allowedPutawayActions(
        { ...testTask, task_status_code: "COMPLETED" },
        testCapabilities,
      ),
    ).toEqual(["reverse"]);
    expect(
      allowedPutawayActions(
        { ...testTask, task_status_code: "CANCELLED" },
        testCapabilities,
      ),
    ).toEqual([]);
    expect(
      allowedPutawayActions(
        { ...testTask, task_status_code: "REVERSED" },
        testCapabilities,
      ),
    ).toEqual([]);
  });
  it("keeps start/complete, planning, and recovery permissions separate", () => {
    expect(
      allowedPutawayActions(testTask, {
        ...testCapabilities,
        canAssign: false,
        canCancel: false,
      }),
    ).toEqual(["start"]);
    expect(
      allowedPutawayActions(testTask, {
        ...testCapabilities,
        canPutaway: false,
        canCancel: false,
      }),
    ).toEqual(["assign", "retarget"]);
    expect(
      allowedPutawayActions(testTask, {
        ...testCapabilities,
        canPutaway: false,
        canAssign: false,
        canCancel: false,
      }),
    ).toEqual([]);
  });
});
