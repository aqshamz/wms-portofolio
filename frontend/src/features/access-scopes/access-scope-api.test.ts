import { describe, expect, it } from "vitest";

import { accountListPath } from "@/features/access-scopes/access-scope-api";

describe("accountListPath", () => {
  it("serializes pagination and a trimmed search", () => {
    expect(
      accountListPath({ search: "  operator  ", page: 2, pageSize: 10 }),
    ).toBe("/api/v1/security/accounts?page=2&page_size=10&search=operator");
  });

  it("omits an empty search", () => {
    expect(accountListPath({ search: " ", page: 1, pageSize: 10 })).toBe(
      "/api/v1/security/accounts?page=1&page_size=10",
    );
  });
});
