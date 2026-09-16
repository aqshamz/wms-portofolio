import { apiRequest } from "@/lib/api/client";
import type {
  CreateInboundOrderRequest,
  InboundOrder,
  InboundOrderLineRequest,
  InboundOrderListFilters,
  InboundOrderPage,
  UpdateInboundOrderRequest,
} from "@/features/inbound-orders/inbound-order-types";

const basePath = "/api/v1/inbound/orders";

export const inboundOrderKeys = {
  all: ["inbound-orders"] as const,
  lists: () => [...inboundOrderKeys.all, "list"] as const,
  list: (filters: InboundOrderListFilters) =>
    [...inboundOrderKeys.lists(), filters] as const,
  details: () => [...inboundOrderKeys.all, "detail"] as const,
  detail: (id: string) => [...inboundOrderKeys.details(), id] as const,
};

export function inboundOrderListPath(filters: InboundOrderListFilters) {
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

export function listInboundOrders(filters: InboundOrderListFilters) {
  return apiRequest<InboundOrderPage>(inboundOrderListPath(filters));
}

export function getInboundOrder(id: string) {
  return apiRequest<InboundOrder>(`${basePath}/${id}`);
}

export function createInboundOrder(request: CreateInboundOrderRequest) {
  return apiRequest<InboundOrder>(basePath, { method: "POST", body: request });
}

export function updateInboundOrder(
  id: string,
  request: UpdateInboundOrderRequest,
) {
  return apiRequest<InboundOrder>(`${basePath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function addInboundOrderLine(
  id: string,
  expectedVersion: number,
  line: InboundOrderLineRequest,
) {
  return apiRequest<InboundOrder>(`${basePath}/${id}/lines`, {
    method: "POST",
    body: { expected_version: expectedVersion, ...line },
  });
}

export function updateInboundOrderLine(
  id: string,
  lineId: string,
  request: Omit<InboundOrderLineRequest, "purchase_order_line_id"> & {
    expected_version: number;
  },
) {
  return apiRequest<InboundOrder>(`${basePath}/${id}/lines/${lineId}`, {
    method: "PUT",
    body: request,
  });
}

export function deleteInboundOrderLine(
  id: string,
  lineId: string,
  expectedVersion: number,
) {
  return apiRequest<InboundOrder>(`${basePath}/${id}/lines/${lineId}`, {
    method: "DELETE",
    body: { expected_version: expectedVersion },
  });
}

export function transitionInboundOrder(
  id: string,
  action: "release" | "close" | "cancel",
  expectedVersion: number,
  reason?: string,
) {
  return apiRequest<InboundOrder>(`${basePath}/${id}/${action}`, {
    method: "POST",
    body:
      action === "release"
        ? { expected_version: expectedVersion }
        : { expected_version: expectedVersion, reason },
  });
}
