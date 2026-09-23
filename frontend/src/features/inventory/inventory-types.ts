import type { PaginatedData } from "@/lib/api/types";

export interface Balance {
  balance_id: string;
  owner_id: string;
  warehouse_id: string;
  location_id: string;
  location_code: string;
  item_id: string;
  item_code: string;
  item_name: string;
  lot_id: string | null;
  lot_number: string | null;
  handling_unit_id: string | null;
  inventory_status_id: string;
  inventory_status_code: string;
  on_hand_qty: string;
  reserved_qty: string;
  available_qty: string;
  uom_id: string;
  uom_code: string;
  version_no: number;
  updated_at: string;
}

export type BalancePage = PaginatedData<Balance>;

export interface BalanceFilters {
  ownerId: string;
  warehouseId: string;
  search: string;
  includeZero: boolean;
  page: number;
  pageSize: number;
}

export interface Movement {
  movement_id: string;
  movement_type_id: string;
  movement_type_code: string;
  owner_id: string;
  warehouse_id: string;
  business_date: string;
  occurred_at: string;
  item_id: string;
  item_code: string;
  lot_id: string | null;
  lot_number: string | null;
  serial_id: string | null;
  handling_unit_id: string | null;
  from_location_id: string | null;
  from_location_code: string | null;
  to_location_id: string | null;
  to_location_code: string | null;
  from_status_id: string | null;
  from_status_code: string | null;
  to_status_id: string | null;
  to_status_code: string | null;
  quantity: string;
  uom_id: string;
  uom_code: string;
  source_document_id: string;
  source_line_id: string | null;
  reason_code_id: string | null;
  notes: string | null;
  operation_key: string;
  created_by: string;
}

export interface MovementType {
  movement_type_id: string;
  code: string;
  name: string;
  description: string | null;
  is_active: boolean;
}

export type MovementPage = PaginatedData<Movement>;

export interface MovementFilters {
  ownerId: string;
  warehouseId: string;
  movementTypeId: string;
  search: string;
  page: number;
  pageSize: number;
}

export interface SerialState {
  serial_id: string;
  serial_no: string;
  balance: Balance;
  version_no: number;
  updated_at: string;
}

export type SerialStatePage = PaginatedData<SerialState>;

export interface SerialStateFilters {
  ownerId: string;
  warehouseId: string;
  search: string;
  page: number;
  pageSize: number;
}
