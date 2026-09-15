import { describe, expect, it } from "vitest";

import {
  documentStatusFormSchema,
  documentTransitionFormSchema,
} from "@/features/document-workflows/document-workflow-schema";

describe("document workflow schemas", () => {
  it("accepts a valid status", () => {
    expect(
      documentStatusFormSchema.safeParse({
        code: "RELEASED",
        name: "Released",
        description: "",
        display_order: "20",
        is_initial: false,
        is_final: false,
        is_cancelled: false,
        is_active: true,
      }).success,
    ).toBe(true);
  });

  it("rejects a transition to the same status", () => {
    const result = documentTransitionFormSchema.safeParse({
      from_status_id: "status-a",
      to_status_id: "status-a",
      required_permission_id: "none",
      is_active: true,
    });
    expect(result.success).toBe(false);
  });
});
