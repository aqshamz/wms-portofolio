import { describe, expect, it } from "vitest";

import {
  inspectionResultListPath,
  qualityStatusListPath,
} from "@/features/quality-setup/quality-setup-api";

const filters = {
  search: " pending ",
  active: "active" as const,
  page: 2,
  pageSize: 10,
};

describe("quality setup paths", () => {
  it("serializes quality status filters", () => {
    expect(qualityStatusListPath(filters)).toBe(
      "/api/v1/master/quality-statuses?page=2&page_size=10&search=pending&active=true",
    );
  });

  it("serializes inspection result filters", () => {
    expect(inspectionResultListPath(filters)).toBe(
      "/api/v1/master/inspection-results?page=2&page_size=10&search=pending&active=true",
    );
  });
});
