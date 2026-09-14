import { describe, expect, it } from "vitest";

import { roleListPath } from "@/features/permissions/permission-api";

describe("roleListPath", () => {
  it("serializes role filters", () => {
    expect(
      roleListPath({
        search: " admin ",
        active: "inactive",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/security/roles?page=2&page_size=10&search=admin&active=false",
    );
  });
});
