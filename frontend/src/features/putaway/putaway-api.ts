import { apiRequest } from "@/lib/api/client";
import type { PaginatedData } from "@/lib/api/types";
import type {
  LookupFilters,
  PutawayAssignee,
  PutawayFilters,
  PutawayPage,
  PutawayPostingRequest,
  PutawayTarget,
  PutawayTask,
} from "./putaway-types";

const basePath = "/api/v1/inbound/putaway-tasks";
export const putawayKeys = {
  all: ["putaway-tasks"] as const,
  lists: () => [...putawayKeys.all, "list"] as const,
  list: (filters: PutawayFilters) => [...putawayKeys.lists(), filters] as const,
  detail: (id: string) => [...putawayKeys.all, "detail", id] as const,
  assignees: (id: string, filters: LookupFilters) =>
    [...putawayKeys.all, "assignees", id, filters] as const,
  targets: (id: string, filters: LookupFilters) =>
    [...putawayKeys.all, "targets", id, filters] as const,
};
export function putawayListPath(filters: PutawayFilters) {
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
export function putawayLookupPath(
  id: string,
  kind: "targets" | "assignees",
  filters: LookupFilters,
) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `${basePath}/${encodeURIComponent(id)}/${kind}?${query}`;
}
export function listPutawayTasks(filters: PutawayFilters) {
  return apiRequest<PutawayPage>(putawayListPath(filters));
}
export function getPutawayTask(id: string) {
  return apiRequest<PutawayTask>(`${basePath}/${encodeURIComponent(id)}`);
}
export function listPutawayAssignees(id: string, filters: LookupFilters) {
  return apiRequest<PaginatedData<PutawayAssignee>>(
    putawayLookupPath(id, "assignees", filters),
  );
}
export function listPutawayTargets(id: string, filters: LookupFilters) {
  return apiRequest<PaginatedData<PutawayTarget>>(
    putawayLookupPath(id, "targets", filters),
  );
}
function transition(id: string, action: string, body: unknown) {
  return apiRequest<PutawayTask>(
    `${basePath}/${encodeURIComponent(id)}/${action}`,
    { method: "POST", body },
  );
}
export function startPutaway(id: string, expectedVersion: number) {
  return transition(id, "start", { expected_version: expectedVersion });
}
export function assignPutaway(
  id: string,
  expectedVersion: number,
  accountId: string,
) {
  return transition(id, "assign", {
    expected_version: expectedVersion,
    account_id: accountId,
  });
}
export function retargetPutaway(
  id: string,
  expectedVersion: number,
  locationId: string,
) {
  return transition(id, "retarget", {
    expected_version: expectedVersion,
    target_location_id: locationId,
  });
}
export function completePutaway(id: string, request: PutawayPostingRequest) {
  return transition(id, "complete", request);
}
export function cancelPutaway(
  id: string,
  request: PutawayPostingRequest & { reason: string },
) {
  return transition(id, "cancel", request);
}
export function reversePutaway(
  id: string,
  request: PutawayPostingRequest & { reason: string },
) {
  return transition(id, "reverse", request);
}
