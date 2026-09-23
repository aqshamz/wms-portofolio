import type { PaginatedData } from "@/lib/api/types";

export const REWORK_STATUSES = [
  "OPEN",
  "ASSIGNED",
  "IN_PROGRESS",
  "COMPLETED",
] as const;
export type ReworkAction = "start" | "complete";
export interface ReworkTask {
  rework_task_id: string;
  quarantine_disposition_id: string;
  quarantine_case_id: string;
  source_balance_id: string;
  owner_id: string;
  warehouse_id: string;
  item_id: string;
  item_code: string;
  task_status_code: string;
  task_priority_code: string;
  planned_qty: string;
  completed_qty: string;
  uom_id: string;
  assigned_to: string | null;
  assigned_username?: string | null;
  work_instructions: string | null;
  result_notes: string | null;
  started_at: string | null;
  completed_at: string | null;
  reinspection_id: string | null;
  version_no: number;
  created_at: string;
}
export interface ReworkFilters {
  ownerId: string;
  warehouseId: string;
  status: string;
  search: string;
  page: number;
  pageSize: number;
}
export type ReworkPage = PaginatedData<ReworkTask>;
export interface ReworkCapabilities {
  accountId: string;
  canRework: boolean;
  timezone: string;
}
export function allowedReworkActions(
  task: ReworkTask,
  capabilities: ReworkCapabilities,
): ReworkAction[] {
  if (!capabilities.canRework || !capabilities.accountId || task.version_no < 1)
    return [];
  if (task.task_status_code === "OPEN") return ["start"];
  if (task.assigned_to !== capabilities.accountId) return [];
  if (task.task_status_code === "ASSIGNED") return ["start"];
  if (task.task_status_code === "IN_PROGRESS") return ["complete"];
  return [];
}
export function reworkAssignee(task: ReworkTask, accountId: string) {
  if (!task.assigned_to) return "Unassigned";
  const username = task.assigned_username
    ? `@${task.assigned_username}`
    : "Account unavailable";
  return task.assigned_to === accountId ? `${username} (you)` : username;
}
export function reworkLabel(code: string) {
  return code
    .toLowerCase()
    .replaceAll("_", " ")
    .replace(/^./, (first) => first.toUpperCase());
}
export function reworkTone(status: string) {
  return status === "COMPLETED"
    ? "success"
    : status === "IN_PROGRESS"
      ? "info"
      : status === "ASSIGNED"
        ? "warning"
        : "neutral";
}
export function reworkHref(task: {
  owner_id: string;
  warehouse_id: string;
  rework_task_id: string;
}) {
  return `/inbound/rework?${new URLSearchParams({ owner: task.owner_id, warehouse: task.warehouse_id, task: task.rework_task_id })}`;
}
