import { describe, expect, it } from "vitest";

import {
  itemCategoryListPath,
  itemListPath,
  uomListPath,
} from "@/features/item-catalog/item-catalog-api";

describe("item catalog paths", () => {
  it("serializes owner and category item filters", () => {
    expect(
      itemListPath({
        ownerId: "owner-id",
        categoryId: "category-id",
        search: " coffee ",
        active: "active",
        page: 2,
        pageSize: 10,
      }),
    ).toBe(
      "/api/v1/master/items?page=2&page_size=10&search=coffee&active=true&owner_id=owner-id&category_id=category-id",
    );
  });

  it("serializes category and UOM lookups", () => {
    expect(
      itemCategoryListPath({
        ownerId: "owner-id",
        search: "",
        active: "all",
        page: 1,
        pageSize: 100,
      }),
    ).toBe(
      "/api/v1/master/item-categories?page=1&page_size=100&owner_id=owner-id",
    );
    expect(
      uomListPath({
        search: "",
        active: "active",
        page: 1,
        pageSize: 100,
      }),
    ).toBe("/api/v1/master/uoms?page=1&page_size=100&active=true");
  });
});
