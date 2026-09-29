import type { PaginatedData } from "@/lib/api/types";

export type ReplenishmentStatus =
  "OPEN" | "ASSIGNED" | "IN_PROGRESS" | "COMPLETED" | "CANCELLED";

export interface ReplenishmentTask {
  replenishment_task_id: string;
  source_balance_id: string;
  source_balance_version_no: number;
  owner_id: string;
  warehouse_id: string;
  item_id: string;
  item_code: string;
  item_name: string;
  lot_id: string | null;
  lot_number: string | null;
  serial_id: string | null;
  serial_number: string | null;
  handling_unit_id: string | null;
  handling_unit_barcode: string | null;
  inventory_status_id: string;
  inventory_status_code: string;
  source_location_id: string;
  source_location_code: string;
  target_location_id: string;
  target_location_code: string;
  planned_qty: string;
  completed_qty: string;
  uom_id: string;
  uom_code: string;
  task_status_code: ReplenishmentStatus;
  task_priority_code: string;
  assigned_to: string | null;
  assigned_username: string | null;
  assigned_display_name: string | null;
  notes: string | null;
  started_at: string | null;
  completed_at: string | null;
  cancelled_at: string | null;
  cancelled_by: string | null;
  cancellation_reason: string | null;
  inventory_movement_id: string | null;
  resulting_balance_id: string | null;
  version_no: number;
  created_at: string;
}

export interface ReplenishmentFilters {
  ownerId: string;
  warehouseId: string;
  status: string;
  search: string;
  page: number;
  pageSize: number;
}

export interface ReplenishmentAssignee {
  account_id: string;
  username: string;
  display_name: string;
}

export interface ReplenishmentTarget {
  location_id: string;
  code: string;
  zone_code: string;
  location_type_code: string;
}

export type ReplenishmentPage = PaginatedData<ReplenishmentTask>;

export function replenishmentLabel(status: ReplenishmentStatus) {
  return status
    .toLowerCase()
    .replaceAll("_", " ")
    .replace(/^./, (value) => value.toUpperCase());
}

export function replenishmentTone(status: ReplenishmentStatus) {
  if (status === "COMPLETED") return "success" as const;
  if (status === "CANCELLED") return "danger" as const;
  if (status === "IN_PROGRESS") return "warning" as const;
  if (status === "ASSIGNED") return "info" as const;
  return "neutral" as const;
}
