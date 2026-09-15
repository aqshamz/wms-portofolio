import { describe, expect, it } from "vitest";

import { numberRuleFormSchema } from "@/features/document-numbering/document-numbering-schema";

describe("numberRuleFormSchema", () => {
  it("accepts a valid numbering rule", () => {
    expect(
      numberRuleFormSchema.safeParse({
        prefix: "PO",
        separator: "-",
        sequence_length: "6",
        effective_from: "2026-09-15",
        include_partner_code: true,
        include_warehouse_code: true,
      }).success,
    ).toBe(true);
  });

  it("rejects unsafe separators and short sequences", () => {
    const result = numberRuleFormSchema.safeParse({
      prefix: "PO",
      separator: "@",
      sequence_length: 2,
      effective_from: "2026-09-15",
      include_partner_code: false,
      include_warehouse_code: false,
    });
    expect(result.success).toBe(false);
  });
});
