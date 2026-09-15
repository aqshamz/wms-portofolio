import { describe, expect, it } from "vitest";

import {
  taskRecordFormSchema,
  taskTransitionFormSchema,
} from "@/features/task-workflows/task-workflow-schema";

describe("task workflow schemas", () => {
  it("coerces a valid priority value", () => {
    const result = taskRecordFormSchema.safeParse({
      code: "HIGH",
      name: "High",
      description: "",
      priority_value: "100",
      is_initial: false,
      is_final: false,
      is_cancelled: false,
      is_active: true,
    });
    expect(result.success).toBe(true);
  });

  it("rejects a self-transition", () => {
    expect(
      taskTransitionFormSchema.safeParse({
        from_status_id: "same",
        to_status_id: "same",
        required_permission_id: "none",
        is_active: true,
      }).success,
    ).toBe(false);
  });

  it("rejects a cancelled status that is not final", () => {
    expect(
      taskRecordFormSchema.safeParse({
        code: "CANCELLED",
        name: "Cancelled",
        description: "",
        priority_value: 0,
        is_initial: false,
        is_final: false,
        is_cancelled: true,
        is_active: true,
      }).success,
    ).toBe(false);
  });
});
