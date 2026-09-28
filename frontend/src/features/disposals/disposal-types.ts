import type { PaginatedData } from "@/lib/api/types";

export type DisposalStatus = "PLANNED" | "COMPLETED" | "CANCELLED";

export interface Disposal {
  disposal_id: string;
  quarantine_disposition_id: string;
  quarantine_case_id: string;
  owner_id: string;
  owner_code: string;
  owner_name: string;
  warehouse_id: string;
  warehouse_code: string;
  warehouse_name: string;
  business_date: string;
  source_balance_id: string;
  source_balance_version_no?: number;
  available_qty?: string;
  planned_balance_version_no: number;
  item_id: string;
  item_code: string;
  item_name: string;
  lot_id?: string | null;
  lot_number?: string;
  serial_id?: string | null;
  serial_no?: string;
  handling_unit_id?: string | null;
  handling_unit_barcode?: string;
  source_location_id: string;
  source_location_code: string;
  source_inventory_status_id: string;
  source_inventory_status_code: string;
  quantity: string;
  uom_id: string;
  uom_code: string;
  status_code: DisposalStatus;
  notes?: string | null;
  planned_at: string;
  completed_at?: string | null;
  completed_by?: string | null;
  completed_by_display_name?: string | null;
  cancelled_at?: string | null;
  cancelled_by?: string | null;
  cancelled_by_display_name?: string | null;
  cancellation_reason?: string | null;
  inventory_movement_id?: string | null;
  created_at: string;
  created_by: string;
  created_by_display_name: string;
  version_no: number;
}

export interface DisposalFilters {
  ownerId: string;
  warehouseId: string;
  status: string;
  search: string;
  page: number;
  pageSize: number;
}

export type DisposalPage = PaginatedData<Disposal>;

export interface DisposalCapabilities {
  canComplete: boolean;
  canCancel: boolean;
  timezone: string;
}

export function disposalLabel(status: string) {
  return status
    .toLowerCase()
    .replaceAll("_", " ")
    .replace(/^./, (first) => first.toUpperCase());
}

export function disposalTone(status: DisposalStatus) {
  return status === "COMPLETED"
    ? "success"
    : status === "CANCELLED"
      ? "neutral"
      : "warning";
}
