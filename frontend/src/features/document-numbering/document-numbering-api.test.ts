import { describe, expect, it } from "vitest";

import { counterListPath } from "@/features/document-numbering/document-numbering-api";

describe("counterListPath", () => {
  it("encodes the required business-date range and paging", () => {
    expect(
      counterListPath("document-type-1", {
        dateFrom: "2026-09-01",
        dateTo: "2026-09-15",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/master/document-types/document-type-1/daily-counters?date_from=2026-09-01&date_to=2026-09-15&page=2&page_size=10",
    );
  });
});
