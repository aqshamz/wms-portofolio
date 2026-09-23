import { apiRequest } from "@/lib/api/client";
import type {
  CreateReceiptRequest,
  InventoryBalance,
  Receipt,
  ReceiptListFilters,
  ReceiptPage,
  UpdateReceiptRequest,
} from "@/features/receipts/receipt-types";

const basePath = "/api/v1/inbound/receipts";

export const receiptKeys = {
  all: ["receipts"] as const,
  lists: () => [...receiptKeys.all, "list"] as const,
  list: (filters: ReceiptListFilters) =>
    [...receiptKeys.lists(), filters] as const,
  details: () => [...receiptKeys.all, "detail"] as const,
  detail: (id: string) => [...receiptKeys.details(), id] as const,
  balance: (id: string) => [...receiptKeys.all, "balance", id] as const,
};

export function receiptListPath(filters: ReceiptListFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.status) query.set("status_code", filters.status);
  if (filters.inspectionEligible) query.set("inspection_eligible", "true");
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `${basePath}?${query}`;
}

export function listReceipts(filters: ReceiptListFilters) {
  return apiRequest<ReceiptPage>(receiptListPath(filters));
}

export function getReceipt(id: string) {
  return apiRequest<Receipt>(`${basePath}/${id}`);
}

export function createReceipt(request: CreateReceiptRequest) {
  return apiRequest<Receipt>(basePath, { method: "POST", body: request });
}

export function updateReceipt(id: string, request: UpdateReceiptRequest) {
  return apiRequest<Receipt>(`${basePath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function completeReceipt(id: string, expectedVersion: number) {
  return apiRequest<Receipt>(`${basePath}/${id}/complete`, {
    method: "POST",
    body: { expected_version: expectedVersion },
  });
}

export function cancelReceipt(
  id: string,
  expectedVersion: number,
  reason: string,
) {
  return apiRequest<Receipt>(`${basePath}/${id}/cancel`, {
    method: "POST",
    body: { expected_version: expectedVersion, reason },
  });
}

export function getInventoryBalance(id: string) {
  return apiRequest<InventoryBalance>(`/api/v1/inventory/balances/${id}`);
}

export function reverseReceipt(
  id: string,
  request: {
    expected_version: number;
    business_date: string;
    reason: string;
    balances: { balance_id: string; expected_version: number }[];
  },
) {
  return apiRequest<Receipt>(`${basePath}/${id}/reverse`, {
    method: "POST",
    body: request,
  });
}
