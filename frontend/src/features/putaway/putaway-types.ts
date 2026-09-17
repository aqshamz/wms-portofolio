import type { PaginatedData } from "@/lib/api/types";

export type PutawayStatus =
  "OPEN" | "ASSIGNED" | "IN_PROGRESS" | "COMPLETED" | "CANCELLED" | "REVERSED";
export type PutawayAction =
  "start" | "assign" | "retarget" | "complete" | "cancel" | "reverse";

export interface PutawayTask {
  putaway_task_id: string;
  inspection_id: string;
  receipt_inventory_id: string;
  source_balance_id: string;
  source_balance_version_no?: number;
  result_balance_version_no?: number;
  base_uom_code?: string;
  owner_id: string;
  warehouse_id: string;
  item_id: string;
  item_code: string;
  lot_id?: string | null;
  lot_number?: string;
  handling_unit_id?: string | null;
  source_location_id: string;
  source_location_code: string;
  target_location_id: string;
  target_location_code: string;
  planned_qty: string;
  completed_qty: string;
  uom_id: string;
  task_status_code: PutawayStatus;
  task_priority_code: string;
  assigned_to?: string | null;
  assigned_display_name?: string | null;
  started_at?: string | null;
  completed_at?: string | null;
  inventory_movement_id?: string | null;
  resulting_balance_id?: string | null;
  reversal_movement_id?: string | null;
  reversed_at?: string | null;
  reversed_by?: string | null;
  reversal_reason?: string | null;
  replacement_inspection_id?: string | null;
  version_no: number;
  created_at: string;
}

export interface PutawayFilters {
  ownerId: string;
  warehouseId: string;
  status: string;
  search: string;
  page: number;
  pageSize: number;
}
export type PutawayPage = PaginatedData<PutawayTask>;
export interface LookupFilters {
  search: string;
  page: number;
  pageSize: number;
}
export interface PutawayAssignee {
  account_id: string;
  username: string;
  display_name: string;
}
export interface PutawayTarget {
  location_id: string;
  code: string;
  zone_code: string;
  location_type_code: string;
}
export interface PutawayCapabilities {
  accountId: string;
  canPutaway: boolean;
  canAssign: boolean;
  canCancel: boolean;
  timezone: string;
}
export interface PutawayPostingRequest {
  expected_version: number;
  expected_balance_version: number;
  business_date: string;
}

export function putawayTone(status: PutawayStatus) {
  if (status === "COMPLETED") return "success";
  if (status === "CANCELLED") return "danger";
  if (status === "IN_PROGRESS") return "warning";
  if (status === "ASSIGNED") return "info";
  return "neutral";
}
export function putawayLabel(status: PutawayStatus) {
  return status
    .toLowerCase()
    .replaceAll("_", " ")
    .replace(/^./, (first) => first.toUpperCase());
}
export function allowedPutawayActions(
  task: PutawayTask,
  permissions: PutawayCapabilities,
): PutawayAction[] {
  const actions: PutawayAction[] = [];
  const beforeStart =
    task.task_status_code === "OPEN" || task.task_status_code === "ASSIGNED";
  const isAssignee = task.assigned_to === permissions.accountId;
  if (beforeStart && permissions.canAssign) actions.push("assign", "retarget");
  if (
    beforeStart &&
    permissions.canPutaway &&
    (task.task_status_code === "OPEN" || isAssignee)
  )
    actions.push("start");
  if (
    task.task_status_code === "IN_PROGRESS" &&
    isAssignee &&
    permissions.canPutaway
  )
    actions.push("complete");
  if (
    permissions.canCancel &&
    (beforeStart || (task.task_status_code === "IN_PROGRESS" && isAssignee))
  )
    actions.push("cancel");
  if (permissions.canCancel && task.task_status_code === "COMPLETED")
    actions.push("reverse");
  return actions;
}
