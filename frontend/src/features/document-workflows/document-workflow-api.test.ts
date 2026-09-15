import { describe, expect, it } from "vitest";

import { documentTypeListPath } from "@/features/document-workflows/document-workflow-api";

describe("documentTypeListPath", () => {
  it("encodes paging, search, and active filters", () => {
    expect(
      documentTypeListPath({
        search: "purchase order",
        active: "inactive",
        page: 2,
        pageSize: 25,
      }),
    ).toBe(
      "/api/v1/master/document-types?page=2&page_size=25&search=purchase+order&active=false",
    );
  });

  it("omits optional filters", () => {
    expect(documentTypeListPath({ search: " ", active: "all" })).toBe(
      "/api/v1/master/document-types?page=1&page_size=100",
    );
  });
});
