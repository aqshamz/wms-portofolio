import { apiRequest } from "@/lib/api/client";
import type { Disposal, DisposalFilters, DisposalPage } from "./disposal-types";

const basePath = "/api/v1/outbound/disposals";

export const disposalKeys = {
  all: ["outbound-disposals"] as const,
  list: (filters: DisposalFilters) =>
    ["outbound-disposals", "list", filters] as const,
  detail: (id: string) => ["outbound-disposals", "detail", id] as const,
};

export function disposalListPath(filters: DisposalFilters) {
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

export const listDisposals = (filters: DisposalFilters) =>
  apiRequest<DisposalPage>(disposalListPath(filters));

export const getDisposal = (id: string) =>
  apiRequest<Disposal>(`${basePath}/${encodeURIComponent(id)}`);

export const completeDisposal = (
  id: string,
  body: {
    expected_version: number;
    expected_balance_version: number;
    completed_at: string;
  },
) =>
  apiRequest<Disposal>(`${basePath}/${encodeURIComponent(id)}/complete`, {
    method: "POST",
    body,
  });

export const cancelDisposal = (
  id: string,
  body: { expected_version: number; reason: string },
) =>
  apiRequest<Disposal>(`${basePath}/${encodeURIComponent(id)}/cancel`, {
    method: "POST",
    body,
  });
