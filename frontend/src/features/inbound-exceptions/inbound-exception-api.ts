import { apiRequest } from "@/lib/api/client";
import type {
  ExceptionFilters,
  ExceptionPage,
  InboundException,
} from "./inbound-exception-types";

const basePath = "/api/v1/inbound/exceptions";
export const exceptionKeys = {
  all: ["inbound-exceptions"] as const,
  list: (filters: ExceptionFilters) =>
    ["inbound-exceptions", "list", filters] as const,
  detail: (id: string) => ["inbound-exceptions", "detail", id] as const,
};
export function exceptionListPath(filters: ExceptionFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  // This API reuses status_code for exception type, not a workflow status.
  if (filters.exceptionType) query.set("status_code", filters.exceptionType);
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `${basePath}?${query}`;
}
export const listInboundExceptions = (filters: ExceptionFilters) =>
  apiRequest<ExceptionPage>(exceptionListPath(filters));
export const getInboundException = (id: string) =>
  apiRequest<InboundException>(`${basePath}/${encodeURIComponent(id)}`);
