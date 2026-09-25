import type { PaginatedData } from "@/lib/api/types";

export type ReceiptStatus = "OPEN" | "COMPLETED" | "CANCELLED" | "REVERSED";

export interface ReceiptBatch {
  receipt_inventory_id: string;
  item_id: string;
  source_qty: string;
  source_uom_id: string;
  source_uom_code: string;
  base_qty: string;
  base_uom_id: string;
  base_uom_code: string;
  lot_id?: string;
  lot_number?: string;
  handling_unit_id?: string;
  handling_unit_barcode?: string;
  serial_id?: string;
  serial_no?: string;
  received_location_id: string;
  received_location_code: string;
  initial_inventory_status_id: string;
  initial_inventory_status_code: string;
  initial_balance_id?: string;
}

export interface ReceiptLine {
  receipt_line_id: string;
  inbound_line_id?: string;
  line_no: number;
  item_id: string;
  item_code: string;
  item_name: string;
  received_qty: string;
  rejected_qty: string;
  uom_conversion_to_base: string;
  received_base_qty: string;
  rejected_base_qty: string;
  base_uom_id: string;
  base_uom_code: string;
  exception_notes?: string;
  exception_type_code?: string;
  accepted_qty: string;
  batched_qty: string;
  uom_id: string;
  uom_code: string;
  batches: ReceiptBatch[];
}

export interface Receipt {
  receipt_id: string;
  inbound_id?: string;
  owner_id: string;
  owner_code: string;
  warehouse_id: string;
  warehouse_code: string;
  business_date: string;
  received_at: string;
  dock_location_id?: string;
  vehicle_number?: string;
  seal_number?: string;
  delivery_note_no?: string;
  status_code: ReceiptStatus;
  notes?: string;
  supersedes_receipt_id?: string;
  successor_receipt_id?: string;
  version_no: number;
  created_at: string;
  lines?: ReceiptLine[];
}

export interface ReceiptLotRequest {
  lot_number: string;
  manufacture_date?: string;
  expiry_date?: string;
}

export interface ReceiptBatchRequest {
  source_qty: string;
  received_location_id: string;
  lot?: ReceiptLotRequest;
  handling_unit_id?: string;
  serial_no?: string;
}

export interface ReceiptLineRequest {
  inbound_line_id: string;
  uom_id?: string;
  received_qty: string;
  rejected_qty: string;
  batches: ReceiptBatchRequest[];
  exception_notes?: string;
  exception_type_code?: string;
}

export interface CreateReceiptRequest {
  inbound_id: string;
  business_date: string;
  received_at: string;
  dock_location_id: string;
  vehicle_number?: string;
  seal_number?: string;
  delivery_note_no?: string;
  notes?: string;
  supersedes_receipt_id?: string;
  lines: ReceiptLineRequest[];
}

export interface UpdateReceiptRequest {
  expected_version: number;
  received_at: string;
  dock_location_id: string;
  vehicle_number?: string;
  seal_number?: string;
  delivery_note_no?: string;
  notes?: string;
  lines: ReceiptLineRequest[];
}

export interface ReceiptListFilters {
  inspectionEligible?: boolean;
  ownerId: string;
  warehouseId: string;
  status: string;
  search: string;
  page: number;
  pageSize: number;
}

export type ReceiptPage = PaginatedData<Receipt>;

export interface InventoryBalance {
  balance_id: string;
  inventory_status_code: string;
  location_id: string;
  on_hand_qty: string;
  version_no: number;
}
