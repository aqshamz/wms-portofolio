import { apiRequest } from "@/lib/api/client";
import type {
  CatalogItem,
  CatalogItemDetail,
  CreateItemCategoryRequest,
  CreateItemRequest,
  ItemCategory,
  ItemCategoryListFilters,
  ItemCategoryPage,
  ItemListFilters,
  ItemPage,
  UOMListFilters,
  UOMPage,
  UpdateItemCategoryRequest,
  UpdateItemRequest,
} from "@/features/item-catalog/item-catalog-types";

const itemPath = "/api/v1/master/items";
const categoryPath = "/api/v1/master/item-categories";
const uomPath = "/api/v1/master/uoms";

export const itemCatalogKeys = {
  all: ["item-catalog"] as const,
  itemLists: () => [...itemCatalogKeys.all, "items", "list"] as const,
  itemList: (filters: ItemListFilters) =>
    [...itemCatalogKeys.itemLists(), filters] as const,
  item: (id: string) =>
    [...itemCatalogKeys.all, "items", "detail", id] as const,
  categoryLists: () => [...itemCatalogKeys.all, "categories", "list"] as const,
  categoryList: (filters: ItemCategoryListFilters) =>
    [...itemCatalogKeys.categoryLists(), filters] as const,
  category: (id: string) =>
    [...itemCatalogKeys.all, "categories", "detail", id] as const,
  uomLists: () => [...itemCatalogKeys.all, "uoms", "list"] as const,
  uomList: (filters: UOMListFilters) =>
    [...itemCatalogKeys.uomLists(), filters] as const,
};

function listQuery(filters: {
  search: string;
  active: string;
  page: number;
  pageSize: number;
}) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all") {
    query.set("active", String(filters.active === "active"));
  }
  return query;
}

export function itemListPath(filters: ItemListFilters) {
  const query = listQuery(filters);
  if (filters.ownerId) query.set("owner_id", filters.ownerId);
  if (filters.categoryId) query.set("category_id", filters.categoryId);
  return `${itemPath}?${query}`;
}

export function itemCategoryListPath(filters: ItemCategoryListFilters) {
  const query = listQuery(filters);
  if (filters.ownerId) query.set("owner_id", filters.ownerId);
  return `${categoryPath}?${query}`;
}

export function uomListPath(filters: UOMListFilters) {
  return `${uomPath}?${listQuery(filters)}`;
}

export function listItems(filters: ItemListFilters) {
  return apiRequest<ItemPage>(itemListPath(filters));
}

export function getItem(id: string) {
  return apiRequest<CatalogItemDetail>(`${itemPath}/${id}`);
}

export function createItem(request: CreateItemRequest) {
  return apiRequest<CatalogItem>(itemPath, { method: "POST", body: request });
}

export function updateItem(id: string, request: UpdateItemRequest) {
  return apiRequest<CatalogItem>(`${itemPath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateItem(id: string, expectedUpdatedAt: string) {
  return apiRequest<CatalogItem>(`${itemPath}/${id}/deactivate`, {
    method: "PATCH",
    body: { expected_updated_at: expectedUpdatedAt },
  });
}

export function listItemCategories(filters: ItemCategoryListFilters) {
  return apiRequest<ItemCategoryPage>(itemCategoryListPath(filters));
}

export function getItemCategory(id: string) {
  return apiRequest<ItemCategory>(`${categoryPath}/${id}`);
}

export function createItemCategory(request: CreateItemCategoryRequest) {
  return apiRequest<ItemCategory>(categoryPath, {
    method: "POST",
    body: request,
  });
}

export function updateItemCategory(
  id: string,
  request: UpdateItemCategoryRequest,
) {
  return apiRequest<ItemCategory>(`${categoryPath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateItemCategory(id: string) {
  return apiRequest<ItemCategory>(`${categoryPath}/${id}/deactivate`, {
    method: "PATCH",
  });
}

export function listUOMs(filters: UOMListFilters) {
  return apiRequest<UOMPage>(uomListPath(filters));
}
