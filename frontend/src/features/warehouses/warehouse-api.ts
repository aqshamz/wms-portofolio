import { apiRequest } from "@/lib/api/client";
import type {
  CreateWarehouseRequest,
  UpdateWarehouseRequest,
  Warehouse,
  WarehouseListFilters,
  WarehousePage,
} from "@/features/warehouses/warehouse-types";

const basePath = "/api/v1/master/warehouses";

export const warehouseKeys = {
  all: ["warehouses"] as const,
  lists: () => [...warehouseKeys.all, "list"] as const,
  list: (filters: WarehouseListFilters) =>
    [...warehouseKeys.lists(), filters] as const,
  details: () => [...warehouseKeys.all, "detail"] as const,
  detail: (id: string) => [...warehouseKeys.details(), id] as const,
};

export function warehouseListPath(filters: WarehouseListFilters) {
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

export function listWarehouses(filters: WarehouseListFilters) {
  return apiRequest<WarehousePage>(warehouseListPath(filters));
}

export function getWarehouse(id: string) {
  return apiRequest<Warehouse>(`${basePath}/${id}`);
}

export function createWarehouse(request: CreateWarehouseRequest) {
  return apiRequest<Warehouse>(basePath, { method: "POST", body: request });
}

export function updateWarehouse(id: string, request: UpdateWarehouseRequest) {
  return apiRequest<Warehouse>(`${basePath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateWarehouse(id: string, expectedUpdatedAt: string) {
  return apiRequest<Warehouse>(`${basePath}/${id}/deactivate`, {
    method: "PATCH",
    body: { expected_updated_at: expectedUpdatedAt },
  });
}
