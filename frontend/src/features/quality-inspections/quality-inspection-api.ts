import { apiRequest } from "@/lib/api/client";
import type {
  CompleteInspectionRequest,
  InspectionFilters,
  InspectionPage,
  QualityInspection,
} from "./quality-inspection-types";

const basePath = "/api/v1/inbound/quality-inspections";

export const inspectionKeys = {
  all: ["quality-inspections"] as const,
  lists: () => [...inspectionKeys.all, "list"] as const,
  list: (filters: InspectionFilters) =>
    [...inspectionKeys.lists(), filters] as const,
  detail: (id: string) => [...inspectionKeys.all, "detail", id] as const,
  receipt: (owner: string, warehouse: string, receipt: string) =>
    [...inspectionKeys.all, "receipt", owner, warehouse, receipt] as const,
};

export function inspectionListPath(filters: InspectionFilters) {
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

export function listInspections(filters: InspectionFilters) {
  return apiRequest<InspectionPage>(inspectionListPath(filters));
}

// Fetch every matching page so an already-inspected batch cannot reappear just
// because a receipt has more than 100 inspections (including replacements).
export async function listReceiptInspections(
  ownerId: string,
  warehouseId: string,
  receiptId: string,
) {
  const inspections: QualityInspection[] = [];
  let page = 1;
  let totalPages = 1;
  do {
    const result = await listInspections({
      ownerId,
      warehouseId,
      search: receiptId,
      status: "",
      page,
      pageSize: 100,
    });
    inspections.push(
      ...result.items.filter((row) => row.receipt_id === receiptId),
    );
    totalPages = result.total_pages;
    page += 1;
  } while (page <= totalPages);
  return inspections;
}

export function getInspection(id: string) {
  return apiRequest<QualityInspection>(`${basePath}/${encodeURIComponent(id)}`);
}

export function createInspection(request: {
  receipt_inventory_id: string;
  notes?: string;
}) {
  return apiRequest<QualityInspection>(basePath, {
    method: "POST",
    body: request,
  });
}

export function completeInspection(
  id: string,
  request: CompleteInspectionRequest,
) {
  return apiRequest<QualityInspection>(
    `${basePath}/${encodeURIComponent(id)}/complete`,
    { method: "POST", body: request },
  );
}

export function cancelInspection(
  id: string,
  expectedVersion: number,
  reason: string,
) {
  return apiRequest<QualityInspection>(
    `${basePath}/${encodeURIComponent(id)}/cancel`,
    { method: "POST", body: { expected_version: expectedVersion, reason } },
  );
}
