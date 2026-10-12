import { apiRequest } from "@/lib/api/client";
import type {
  CreateOutboundOrderRequest,
  OutboundHeaderRequest,
  OutboundLineRequest,
  OutboundOrder,
  OutboundOrderFilters,
  OutboundOrderPage,
  OutboundValidationRun,
} from "./outbound-order-types";

const basePath = "/api/v1/outbound/orders";
const orderPath = (id: string) => `${basePath}/${encodeURIComponent(id)}`;
export const outboundOrderKeys = {
  all: ["outbound-orders"] as const,
  list: (filters: OutboundOrderFilters) =>
    ["outbound-orders", "list", filters] as const,
  detail: (id: string) => ["outbound-orders", "detail", id] as const,
};
export function outboundOrderListPath(filters: OutboundOrderFilters) {
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
export const listOutboundOrders = (filters: OutboundOrderFilters) =>
  apiRequest<OutboundOrderPage>(outboundOrderListPath(filters));
export const getOutboundOrder = (id: string) =>
  apiRequest<OutboundOrder>(orderPath(id));
export const createOutboundOrder = (body: CreateOutboundOrderRequest) =>
  apiRequest<OutboundOrder>(basePath, { method: "POST", body });
export const updateOutboundOrder = (
  id: string,
  body: OutboundHeaderRequest & { expected_version: number },
) => apiRequest<OutboundOrder>(orderPath(id), { method: "PUT", body });
export const addOutboundOrderLine = (
  id: string,
  version: number,
  line: OutboundLineRequest,
) =>
  apiRequest<OutboundOrder>(`${orderPath(id)}/lines`, {
    method: "POST",
    body: { ...line, expected_version: version },
  });
export const updateOutboundOrderLine = (
  id: string,
  lineId: string,
  version: number,
  line: Omit<OutboundLineRequest, "item_id">,
) =>
  apiRequest<OutboundOrder>(
    `${orderPath(id)}/lines/${encodeURIComponent(lineId)}`,
    { method: "PUT", body: { ...line, expected_version: version } },
  );
export const deleteOutboundOrderLine = (
  id: string,
  lineId: string,
  version: number,
) =>
  apiRequest<OutboundOrder>(
    `${orderPath(id)}/lines/${encodeURIComponent(lineId)}`,
    { method: "DELETE", body: { expected_version: version } },
  );
export const validateOutboundOrder = (id: string, version: number) =>
  apiRequest<OutboundValidationRun>(`${orderPath(id)}/validate`, {
    method: "POST",
    body: { expected_version: version },
  });
export const transitionOutboundOrder = (
  id: string,
  action: "release" | "cancel",
  version: number,
  reason?: string,
) =>
  apiRequest<OutboundOrder>(`${orderPath(id)}/${action}`, {
    method: "POST",
    body: {
      expected_version: version,
      ...(action === "cancel" ? { reason } : {}),
    },
  });
