import { apiRequest } from "@/lib/api/client";
import type {
  ReworkAction,
  ReworkFilters,
  ReworkPage,
  ReworkTask,
} from "./rework-types";

const basePath = "/api/v1/inbound/rework-tasks";
export const reworkKeys = {
  all: ["rework-tasks"] as const,
  list: (filters: ReworkFilters) => ["rework-tasks", "list", filters] as const,
  detail: (id: string) => ["rework-tasks", "detail", id] as const,
};
export function reworkListPath(filters: ReworkFilters) {
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
export const listReworkTasks = (filters: ReworkFilters) =>
  apiRequest<ReworkPage>(reworkListPath(filters));
export const getReworkTask = (id: string) =>
  apiRequest<ReworkTask>(`${basePath}/${encodeURIComponent(id)}`);
export const transitionReworkTask = (
  id: string,
  action: ReworkAction,
  body: { expected_version: number; result_notes?: string },
) =>
  apiRequest<ReworkTask>(`${basePath}/${encodeURIComponent(id)}/${action}`, {
    method: "POST",
    body,
  });
