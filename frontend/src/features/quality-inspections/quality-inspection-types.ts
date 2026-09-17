import type { PaginatedData } from "@/lib/api/types";

export interface QualityInspection {
  inspection_id: string;
  receipt_inventory_id: string;
  parent_inspection_id?: string | null;
  source_balance_id?: string | null;
  source_balance_version_no?: number;
  base_uom_code?: string;
  is_indivisible?: boolean;
  serial_no?: string;
  handling_unit_barcode?: string;
  receipt_id: string;
  owner_id: string;
  warehouse_id: string;
  item_id: string;
  item_code: string;
  lot_number?: string;
  location_code: string;
  quality_status_code: string;
  inspection_result_code?: string;
  inspected_qty: string;
  passed_qty: string;
  failed_qty: string;
  inspected_at?: string | null;
  inspected_by?: string | null;
  cancelled_at?: string | null;
  cancelled_by?: string | null;
  cancellation_reason?: string | null;
  replacement_inspection_id?: string | null;
  notes?: string | null;
  version_no: number;
  created_at: string;
  putaway_task?: {
    putaway_task_id: string;
    target_location_code: string;
    planned_qty: string;
    task_status_code: string;
  };
  quarantine_case?: {
    quarantine_case_id: string;
    quarantine_qty: string;
    status_code: string;
  };
}

export interface InspectionFilters {
  ownerId: string;
  warehouseId: string;
  status: string;
  search: string;
  page: number;
  pageSize: number;
}

export type InspectionPage = PaginatedData<QualityInspection>;

export interface CompleteInspectionRequest {
  expected_version: number;
  expected_balance_version: number;
  passed_qty: string;
  failed_qty: string;
  putaway_target_location_id?: string;
  notes?: string;
}

export function inspectionState(inspection: QualityInspection) {
  if (inspection.cancelled_at) return "Cancelled";
  if (!inspection.inspected_at) return "Pending";
  if (inspection.inspection_result_code === "ACCEPTED") return "Accepted";
  if (inspection.inspection_result_code === "REJECTED") return "Rejected";
  return "Partial";
}

export function inspectionTone(inspection: QualityInspection) {
  const state = inspectionState(inspection);
  if (state === "Accepted") return "success";
  if (state === "Rejected") return "danger";
  if (state === "Partial") return "warning";
  if (state === "Pending") return "info";
  return "neutral";
}
