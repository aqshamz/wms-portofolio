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

  it("uses the exact inbound action codes enforced by the backend", () => {
    expect(
      isAuthorized([PERMISSIONS.INBOUND.READ], MODULE_ACCESS.INBOUND.READ),
    ).toBe(true);
    expect(
      isAuthorized([PERMISSIONS.INBOUND.READ], MODULE_ACCESS.INBOUND.PLAN),
    ).toBe(false);
    expect(
      isAuthorized([PERMISSIONS.INBOUND.PLAN], MODULE_ACCESS.INBOUND.PLAN),
    ).toBe(true);
  });

  it("allows the backend wildcard permission", () => {
    expect(isAuthorized(["*"], MODULE_ACCESS.SECURITY.WRITE)).toBe(true);
  });
});
