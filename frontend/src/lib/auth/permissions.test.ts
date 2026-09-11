import { describe, expect, it } from "vitest";

import {
  hasPermission,
  isAuthorized,
  MODULE_ACCESS,
  PERMISSIONS,
} from "@/lib/auth/permissions";

describe("authorization helpers", () => {
  it("matches an exact permission", () => {
    expect(
      hasPermission([PERMISSIONS.REPORTING.READ], PERMISSIONS.REPORTING.READ),
    ).toBe(true);
    expect(
      hasPermission([PERMISSIONS.REPORTING.READ], PERMISSIONS.REPORTING.WRITE),
    ).toBe(false);
  });

  it("uses the same read and write codes enforced by backend modules", () => {
    expect(
      isAuthorized([PERMISSIONS.INBOUND.READ], MODULE_ACCESS.INBOUND.READ),
    ).toBe(true);
    expect(
      isAuthorized([PERMISSIONS.INBOUND.READ], MODULE_ACCESS.INBOUND.WRITE),
    ).toBe(false);
  });

  it("allows the backend wildcard permission", () => {
    expect(isAuthorized(["*"], MODULE_ACCESS.SECURITY.WRITE)).toBe(true);
  });
});
