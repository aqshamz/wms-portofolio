import { apiRequest } from "@/lib/api/client";
import type {
  CreateInventoryStatusRequest,
  InventoryStatus,
  InventoryStatusListFilters,
  InventoryStatusPage,
  UpdateInventoryStatusRequest,
} from "@/features/inventory-classifications/inventory-classification-types";

const basePath = "/api/v1/master/inventory-statuses";

export const inventoryClassificationKeys = {
  all: ["inventory-classifications"] as const,
  lists: () => [...inventoryClassificationKeys.all, "list"] as const,
  list: (filters: InventoryStatusListFilters) =>
    [...inventoryClassificationKeys.lists(), filters] as const,
  detail: (id: string) =>
    [...inventoryClassificationKeys.all, "detail", id] as const,
};

export function inventoryStatusListPath(filters: InventoryStatusListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all") {
    query.set("active", String(filters.active === "active"));
  }
  return `${basePath}?${query}`;
}

export function listInventoryStatuses(filters: InventoryStatusListFilters) {
  return apiRequest<InventoryStatusPage>(inventoryStatusListPath(filters));
}

export function getInventoryStatus(id: string) {
  return apiRequest<InventoryStatus>(`${basePath}/${id}`);
}

export function createInventoryStatus(request: CreateInventoryStatusRequest) {
  return apiRequest<InventoryStatus>(basePath, {
    method: "POST",
    body: request,
  });
}

export function updateInventoryStatus(
  id: string,
  request: UpdateInventoryStatusRequest,
) {
  return apiRequest<InventoryStatus>(`${basePath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateInventoryStatus(id: string) {
  return apiRequest<InventoryStatus>(`${basePath}/${id}/deactivate`, {
    method: "PATCH",
  });
}
