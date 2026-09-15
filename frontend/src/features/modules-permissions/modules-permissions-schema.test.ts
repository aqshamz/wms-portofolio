import { describe, expect, it } from "vitest";

import {
  appModuleFormSchema,
  emptyAppModuleForm,
} from "@/features/modules-permissions/modules-permissions-schema";

describe("appModuleFormSchema", () => {
  it("accepts a valid ordered module", () => {
    expect(
      appModuleFormSchema.safeParse({
        ...emptyAppModuleForm,
        code: "INBOUND",
        name: "Inbound",
        display_order: 10,
      }).success,
    ).toBe(true);
  });

  it("rejects malformed codes and negative display order", () => {
    expect(
      appModuleFormSchema.safeParse({
        ...emptyAppModuleForm,
        code: "INBOUND WORK",
        name: "Inbound",
        display_order: -1,
      }).success,
    ).toBe(false);
  });
});
