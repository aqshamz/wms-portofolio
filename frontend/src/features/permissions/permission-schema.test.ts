import { describe, expect, it } from "vitest";

import {
  emptyRoleForm,
  roleFormSchema,
} from "@/features/permissions/permission-schema";

describe("roleFormSchema", () => {
  it("accepts a role with a permission set", () => {
    expect(
      roleFormSchema.safeParse({
        ...emptyRoleForm,
        code: "WAREHOUSE_ADMIN",
        name: "Warehouse administrator",
        permission_ids: ["permission-id"],
      }).success,
    ).toBe(true);
  });

  it("rejects codes containing spaces", () => {
    expect(
      roleFormSchema.safeParse({
        ...emptyRoleForm,
        code: "BAD ROLE",
        name: "Bad role",
      }).success,
    ).toBe(false);
  });
});
