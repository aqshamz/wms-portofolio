import { apiRequest } from "@/lib/api/client";
import type {
  CreateHandlingUnitTypeRequest,
  CreateItemBarcodeRequest,
  CreateItemUOMRequest,
  CreateUOMRequest,
  HandlingUnitType,
  HandlingUnitTypePage,
  ItemBarcode,
  ItemUOM,
  ReferenceListFilters,
  UOM,
  UOMPage,
  UpdateHandlingUnitTypeRequest,
  UpdateItemBarcodeRequest,
  UpdateItemUOMRequest,
  UpdateUOMRequest,
} from "@/features/units-packaging/units-packaging-types";

const uomPath = "/api/v1/master/uoms";
const handlingUnitPath = "/api/v1/master/handling-unit-types";
const itemPath = "/api/v1/master/items";

export const unitsPackagingKeys = {
  all: ["units-packaging"] as const,
  uomLists: () => [...unitsPackagingKeys.all, "uoms", "list"] as const,
  uomList: (filters: ReferenceListFilters) =>
    [...unitsPackagingKeys.uomLists(), filters] as const,
  uom: (id: string) => [...unitsPackagingKeys.all, "uoms", id] as const,
  handlingLists: () =>
    [...unitsPackagingKeys.all, "handling-units", "list"] as const,
  handlingList: (filters: ReferenceListFilters) =>
    [...unitsPackagingKeys.handlingLists(), filters] as const,
  handling: (id: string) =>
    [...unitsPackagingKeys.all, "handling-units", id] as const,
  itemUOMs: (itemId: string) =>
    [...unitsPackagingKeys.all, "items", itemId, "uoms"] as const,
  barcodes: (itemId: string) =>
    [...unitsPackagingKeys.all, "items", itemId, "barcodes"] as const,
};

function referenceListPath(base: string, filters: ReferenceListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all") {
    query.set("active", String(filters.active === "active"));
  }
  return `${base}?${query}`;
}

export function uomListPath(filters: ReferenceListFilters) {
  return referenceListPath(uomPath, filters);
}

export function handlingUnitListPath(filters: ReferenceListFilters) {
  return referenceListPath(handlingUnitPath, filters);
}

export function listUOMs(filters: ReferenceListFilters) {
  return apiRequest<UOMPage>(uomListPath(filters));
}

export function getUOM(id: string) {
  return apiRequest<UOM>(`${uomPath}/${id}`);
}

export function createUOM(request: CreateUOMRequest) {
  return apiRequest<UOM>(uomPath, { method: "POST", body: request });
}

export function updateUOM(id: string, request: UpdateUOMRequest) {
  return apiRequest<UOM>(`${uomPath}/${id}`, { method: "PUT", body: request });
}

export function deactivateUOM(id: string) {
  return apiRequest<UOM>(`${uomPath}/${id}/deactivate`, { method: "PATCH" });
}

export function listHandlingUnitTypes(filters: ReferenceListFilters) {
  return apiRequest<HandlingUnitTypePage>(handlingUnitListPath(filters));
}

export function getHandlingUnitType(id: string) {
  return apiRequest<HandlingUnitType>(`${handlingUnitPath}/${id}`);
}

export function createHandlingUnitType(request: CreateHandlingUnitTypeRequest) {
  return apiRequest<HandlingUnitType>(handlingUnitPath, {
    method: "POST",
    body: request,
  });
}

export function updateHandlingUnitType(
  id: string,
  request: UpdateHandlingUnitTypeRequest,
) {
  return apiRequest<HandlingUnitType>(`${handlingUnitPath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateHandlingUnitType(id: string) {
  return apiRequest<HandlingUnitType>(`${handlingUnitPath}/${id}/deactivate`, {
    method: "PATCH",
  });
}

export function listItemUOMs(itemId: string) {
  return apiRequest<ItemUOM[]>(`${itemPath}/${itemId}/uoms`);
}

export function createItemUOM(itemId: string, request: CreateItemUOMRequest) {
  return apiRequest<ItemUOM>(`${itemPath}/${itemId}/uoms`, {
    method: "POST",
    body: request,
  });
}

export function updateItemUOM(
  itemId: string,
  itemUOMId: string,
  request: UpdateItemUOMRequest,
) {
  return apiRequest<ItemUOM>(`${itemPath}/${itemId}/uoms/${itemUOMId}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateItemUOM(itemId: string, itemUOMId: string) {
  return apiRequest<ItemUOM>(
    `${itemPath}/${itemId}/uoms/${itemUOMId}/deactivate`,
    { method: "PATCH" },
  );
}

export function listItemBarcodes(itemId: string) {
  return apiRequest<ItemBarcode[]>(`${itemPath}/${itemId}/barcodes`);
}

export function createItemBarcode(
  itemId: string,
  request: CreateItemBarcodeRequest,
) {
  return apiRequest<ItemBarcode>(`${itemPath}/${itemId}/barcodes`, {
    method: "POST",
    body: request,
  });
}

export function updateItemBarcode(
  itemId: string,
  barcodeId: string,
  request: UpdateItemBarcodeRequest,
) {
  return apiRequest<ItemBarcode>(
    `${itemPath}/${itemId}/barcodes/${barcodeId}`,
    { method: "PUT", body: request },
  );
}

export function deactivateItemBarcode(itemId: string, barcodeId: string) {
  return apiRequest<ItemBarcode>(
    `${itemPath}/${itemId}/barcodes/${barcodeId}/deactivate`,
    { method: "PATCH" },
  );
}

export function setPrimaryItemBarcode(itemId: string, barcodeId: string) {
  return apiRequest<ItemBarcode>(
    `${itemPath}/${itemId}/barcodes/${barcodeId}/primary`,
    { method: "PATCH" },
  );
}
