import { apiRequest } from "@/lib/api/client";
import type {
  CreatePurchaseOrderRequest,
  PurchaseOrder,
  PurchaseOrderLineRequest,
  PurchaseOrderListFilters,
  PurchaseOrderPage,
  UpdatePurchaseOrderRequest,
} from "@/features/purchase-orders/purchase-order-types";

const basePath = "/api/v1/inbound/purchase-orders";

export const purchaseOrderKeys = {
  all: ["purchase-orders"] as const,
  lists: () => [...purchaseOrderKeys.all, "list"] as const,
  list: (filters: PurchaseOrderListFilters) =>
    [...purchaseOrderKeys.lists(), filters] as const,
  details: () => [...purchaseOrderKeys.all, "detail"] as const,
  detail: (id: string) => [...purchaseOrderKeys.details(), id] as const,
};

export function purchaseOrderListPath(filters: PurchaseOrderListFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.status) query.set("status_code", filters.status);
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `${basePath}?${query}`;
}

export function listPurchaseOrders(filters: PurchaseOrderListFilters) {
  return apiRequest<PurchaseOrderPage>(purchaseOrderListPath(filters));
}

export function getPurchaseOrder(id: string) {
  return apiRequest<PurchaseOrder>(`${basePath}/${id}`);
}

export function createPurchaseOrder(request: CreatePurchaseOrderRequest) {
  return apiRequest<PurchaseOrder>(basePath, { method: "POST", body: request });
}

export function updatePurchaseOrder(
  id: string,
  request: UpdatePurchaseOrderRequest,
) {
  return apiRequest<PurchaseOrder>(`${basePath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function addPurchaseOrderLine(
  id: string,
  expectedVersion: number,
  line: PurchaseOrderLineRequest,
) {
  return apiRequest<PurchaseOrder>(`${basePath}/${id}/lines`, {
    method: "POST",
    body: { expected_version: expectedVersion, ...line },
  });
}

export function updatePurchaseOrderLine(
  id: string,
  lineId: string,
  request: Omit<PurchaseOrderLineRequest, "item_id" | "uom_id"> & {
    expected_version: number;
  },
) {
  return apiRequest<PurchaseOrder>(`${basePath}/${id}/lines/${lineId}`, {
    method: "PUT",
    body: request,
  });
}

export function deletePurchaseOrderLine(
  id: string,
  lineId: string,
  expectedVersion: number,
) {
  return apiRequest<PurchaseOrder>(`${basePath}/${id}/lines/${lineId}`, {
    method: "DELETE",
    body: { expected_version: expectedVersion },
  });
}

export function transitionPurchaseOrder(
  id: string,
  action: "approve" | "close" | "cancel",
  expectedVersion: number,
  reason?: string,
) {
  return apiRequest<PurchaseOrder>(`${basePath}/${id}/${action}`, {
    method: "POST",
    body:
      action === "approve"
        ? { expected_version: expectedVersion }
        : { expected_version: expectedVersion, reason },
  });
}
