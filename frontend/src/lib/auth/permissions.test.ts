import { describe, expect, it } from "vitest";

import { hasPermission, isAuthorized } from "@/lib/auth/permissions";

describe("authorization helpers", () => {
  it("matches an exact permission", () => {
    expect(hasPermission(["REPORT.INBOUND"], "REPORT.INBOUND")).toBe(true);
    expect(hasPermission(["REPORT.INBOUND"], "REPORT.OUTBOUND")).toBe(false);
  });

  it("matches one permission or one module prefix", () => {
    expect(isAuthorized(["INBOUND.PO.READ"], { anyPrefix: ["INBOUND."] })).toBe(
      true,
    );
    expect(
      isAuthorized(["REPORT.BILLING"], {
        anyOf: ["REPORT.INBOUND", "REPORT.BILLING"],
      }),
    ).toBe(true);
  });

  it("allows the backend wildcard permission", () => {
    expect(isAuthorized(["*"], { anyOf: ["SECURITY.WRITE"] })).toBe(true);
  });
});
