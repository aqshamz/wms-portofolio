import { describe, expect, it } from "vitest";

import {
  handlingUnitListPath,
  uomListPath,
} from "@/features/units-packaging/units-packaging-api";

const filters = {
  search: " box ",
  active: "active" as const,
  page: 2,
  pageSize: 10,
};

describe("units and packaging paths", () => {
  it("serializes UOM filters", () => {
    expect(uomListPath(filters)).toBe(
      "/api/v1/master/uoms?page=2&page_size=10&search=box&active=true",
    );
  });

  it("serializes handling unit filters", () => {
    expect(handlingUnitListPath(filters)).toBe(
      "/api/v1/master/handling-unit-types?page=2&page_size=10&search=box&active=true",
    );
  });
});
