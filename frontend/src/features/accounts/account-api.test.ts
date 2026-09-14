import { describe, expect, it } from "vitest";

import { accountListPath } from "@/features/accounts/account-api";

describe("accountListPath", () => {
  it("serializes search, status, and pagination", () => {
    expect(
      accountListPath({
        search: "  warehouse user ",
        statusId: "status-id",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/security/accounts?page=2&page_size=10&search=warehouse+user&status_id=status-id",
    );
  });
});
