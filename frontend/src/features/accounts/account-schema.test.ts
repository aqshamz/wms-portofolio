import { describe, expect, it } from "vitest";

import {
  accountFormSchema,
  emptyAccountForm,
  passwordResetSchema,
} from "@/features/accounts/account-schema";

describe("accountFormSchema", () => {
  it("accepts a valid account", () => {
    expect(
      accountFormSchema.safeParse({
        ...emptyAccountForm,
        username: "warehouse.user",
        display_name: "Warehouse User",
        password: "long-password-123",
      }).success,
    ).toBe(true);
  });

  it("rejects invalid usernames and short passwords", () => {
    expect(
      accountFormSchema.safeParse({
        ...emptyAccountForm,
        username: "bad user",
        display_name: "Bad User",
        password: "short",
      }).success,
    ).toBe(false);
  });
});

describe("passwordResetSchema", () => {
  it("requires matching passwords", () => {
    expect(
      passwordResetSchema.safeParse({
        password: "long-password-123",
        confirmation: "different-password",
      }).success,
    ).toBe(false);
  });
});
