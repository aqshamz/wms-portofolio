import { apiRequest } from "@/lib/api/client";
import type {
  VendorReturn,
  VendorReturnFilters,
  VendorReturnPage,
} from "./vendor-return-types";

const basePath = "/api/v1/outbound/vendor-returns";

export const vendorReturnKeys = {
  all: ["vendor-returns"] as const,
  list: (filters: VendorReturnFilters) =>
    ["vendor-returns", "list", filters] as const,
  detail: (id: string) => ["vendor-returns", "detail", id] as const,
};

export function vendorReturnListPath(filters: VendorReturnFilters) {
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

export const listVendorReturns = (filters: VendorReturnFilters) =>
  apiRequest<VendorReturnPage>(vendorReturnListPath(filters));

export const getVendorReturn = (id: string) =>
  apiRequest<VendorReturn>(`${basePath}/${encodeURIComponent(id)}`);

export const completeVendorReturn = (
  id: string,
  body: {
    expected_version: number;
    expected_balance_version: number;
    completed_at: string;
  },
) =>
  apiRequest<VendorReturn>(`${basePath}/${encodeURIComponent(id)}/complete`, {
    method: "POST",
    body,
  });

export const cancelVendorReturn = (
  id: string,
  body: { expected_version: number; reason: string },
) =>
  apiRequest<VendorReturn>(`${basePath}/${encodeURIComponent(id)}/cancel`, {
    method: "POST",
    body,
  });
