import type { Balance, Movement } from "@/features/inventory/inventory-types";
import type { PaginatedData } from "@/lib/api/types";

export interface StockControlReason {
  reason_code_id: string;
  code: string;
  name: string;
  description: string | null;
  requires_note: boolean;
  is_active: boolean;
}

export interface InternalMoveRequest {
  operation_key: string;
  business_date: string;
  source_document_id: string;
  reason_code?: string;
  notes?: string;
  source_balance_id: string;
  target_location_id: string;
  quantity: string;
  serial_ids: string[];
  expected_version: number;
}

export interface StockControlCommandResponse {
  operation: string;
  movements: Movement[];
  source_balance?: Balance;
  destination_balance?: Balance;
  idempotent_replay: boolean;
}

export type AdjustmentDirection = "INCREASE" | "DECREASE";
export type AdjustmentStatus =
  "DRAFT" | "PARTIALLY_POSTED" | "POSTED" | "CANCELLED";

export interface InventoryAdjustmentLine {
  inventory_adjustment_line_id: string;
  line_no: number;
  decision_code: "PENDING" | "POSTED" | "REJECTED" | "CANCELLED";
  balance_id: string;
  planned_balance_version_no: number;
  current_balance_version_no?: number;
  item_id: string;
  item_code: string;
  item_name: string;
  lot_id: string | null;
  lot_number: string | null;
  serial_id: string | null;
  serial_number: string | null;
  handling_unit_id: string | null;
  handling_unit_barcode: string | null;
  location_id: string;
  location_code: string;
  inventory_status_id: string;
  inventory_status_code: string;
  quantity: string;
  uom_id: string;
  uom_code: string;
  approved_at: string | null;
  approved_by: string | null;
  approved_by_display_name: string | null;
  rejected_at: string | null;
  rejected_by: string | null;
  rejected_by_display_name: string | null;
  rejection_reason: string | null;
  cancelled_at: string | null;
  cancelled_by: string | null;
  cancellation_reason: string | null;
  inventory_movement_id: string | null;
  resulting_balance_id: string | null;
  version_no: number;
}

export interface InventoryAdjustment {
  inventory_adjustment_id: string;
  status_code: AdjustmentStatus;
  owner_id: string;
  owner_code: string;
  owner_name: string;
  warehouse_id: string;
  warehouse_code: string;
  warehouse_name: string;
  business_date: string;
  direction: AdjustmentDirection;
  reason_code: string;
  reason_name: string;
  notes: string | null;
  total_lines: number;
  pending_lines: number;
  posted_lines: number;
  rejected_lines: number;
  cancelled_lines: number;
  lines: InventoryAdjustmentLine[];
  completed_at: string | null;
  cancelled_at: string | null;
  cancelled_by: string | null;
  cancelled_by_display_name: string | null;
  cancellation_reason: string | null;
  created_at: string;
  created_by: string;
  created_by_username: string;
  created_by_display_name: string;
  version_no: number;
}

export interface AdjustmentFilters {
  ownerId: string;
  warehouseId: string;
  status: string;
  search: string;
  page: number;
  pageSize: number;
}

export type AdjustmentPage = PaginatedData<InventoryAdjustment>;

export interface CreateAdjustmentRequest {
  business_date: string;
  direction: AdjustmentDirection;
  reason_code: string;
  notes?: string;
  lines: {
    balance_id: string;
    quantity: string;
    serial_id?: string;
    expected_balance_version: number;
  }[];
}
