import { apiRequest } from "@/lib/api/client";
import type { PaginatedData } from "@/lib/api/types";
import type {
  DispositionRequest,
  DispositionType,
  QuarantineCase,
  QuarantineFilters,
  QuarantineLookupFilters,
  QuarantinePage,
  QuarantineTarget,
} from "./quarantine-types";

const basePath = "/api/v1/inbound/quarantine-cases";
export const quarantineKeys = {
  all: ["quarantine-cases"] as const,
  list: (filters: QuarantineFilters) =>
    ["quarantine-cases", "list", filters] as const,
  detail: (id: string) => ["quarantine-cases", "detail", id] as const,
  types: ["quarantine-disposition-types", "active"] as const,
  targets: (id: string, filters: QuarantineLookupFilters) =>
    ["quarantine-cases", "targets", id, filters] as const,
};
export function quarantineListPath(filters: QuarantineFilters) {
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
export function quarantineTargetPath(
  id: string,
  filters: QuarantineLookupFilters,
) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `${basePath}/${encodeURIComponent(id)}/targets?${query}`;
}
export const listQuarantineCases = (filters: QuarantineFilters) =>
  apiRequest<QuarantinePage>(quarantineListPath(filters));
export const getQuarantineCase = (id: string) =>
  apiRequest<QuarantineCase>(`${basePath}/${encodeURIComponent(id)}`);
export const listDispositionTypes = () =>
  apiRequest<DispositionType[]>(
    "/api/v1/inbound/quarantine-disposition-types?active=true",
  );
export const listQuarantineTargets = (
  id: string,
  filters: QuarantineLookupFilters,
) =>
  apiRequest<PaginatedData<QuarantineTarget>>(
    quarantineTargetPath(id, filters),
  );
export const createDisposition = (id: string, body: DispositionRequest) =>
  apiRequest<QuarantineCase>(
    `${basePath}/${encodeURIComponent(id)}/dispositions`,
    { method: "POST", body },
  );
