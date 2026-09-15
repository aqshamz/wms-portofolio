import { describe, expect, it } from "vitest";

import { modulesPermissionsListPath } from "@/features/modules-permissions/modules-permissions-api";

describe("module and workflow permission paths", () => {
  it("serializes module search, status, and pagination", () => {
    expect(
      modulesPermissionsListPath("modules", {
        search: " inbound ",
        active: "active",
        page: 2,
        pageSize: 12,
      }),
    ).toBe(
      "/api/v1/master/modules?page=2&page_size=12&search=inbound&active=true",
    );
  });

  it("serializes and encodes the permission module filter", () => {
    expect(
      modulesPermissionsListPath("permissions", {
        search: "release stock",
        active: "inactive",
        moduleCode: "QUALITY & HOLD",
        page: 1,
        pageSize: 12,
      }),
    ).toBe(
      "/api/v1/master/workflow-permissions?page=1&page_size=12&search=release+stock&active=false&module_code=QUALITY+%26+HOLD",
    );
  });
});
