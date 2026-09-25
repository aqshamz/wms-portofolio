import type { PaginatedData } from "@/lib/api/types";

export type InboundOrderStatus =
  | "DRAFT"
  | "RELEASED"
  | "PARTIALLY_RECEIVED"
  | "RECEIVED"
  | "CLOSED"
  | "CANCELLED";

export interface InboundOrderLine {
  inbound_line_id: string;
  purchase_order_line_id?: string;
  line_no: number;
  item_id: string;
  item_code: string;
  item_name: string;
  expected_qty: string;
  uom_conversion_to_base: string;
  expected_base_qty: string;
  base_uom_id: string;
  base_uom_code: string;
  completed_receipt_qty: string;
  completed_receipt_base_qty: string;
  uom_id: string;
  uom_code: string;
  expected_lot_no?: string;
  expected_expiry_date?: string;
  customer_line_reference?: string;
  notes?: string;
}

export interface InboundOrder {
  inbound_id: string;
  purchase_order_id: string;
  owner_id: string;
  owner_code: string;
  vendor_id: string;
  vendor_code: string;
  vendor_name: string;
  warehouse_id: string;
  warehouse_code: string;
  business_date: string;
  expected_arrival_at?: string;
  external_reference?: string;
  supplier_reference?: string;
  status_code: InboundOrderStatus;
  notes?: string;
  supersedes_inbound_id?: string;
  successor_inbound_id?: string;
  version_no: number;
  created_at: string;
  lines?: InboundOrderLine[];
}

export type InboundOrderPage = PaginatedData<InboundOrder>;

export interface InboundOrderListFilters {
  ownerId: string;
  warehouseId: string;
  status: string;
  search: string;
  page: number;
  pageSize: number;
}

export interface InboundOrderLineRequest {
  purchase_order_line_id: string;
  expected_qty: string;
  customer_line_reference?: string;
  notes?: string;
}

export interface CreateInboundOrderRequest {
  purchase_order_id: string;
  business_date: string;
  expected_arrival_at?: string;
  external_reference?: string;
  supplier_reference?: string;
  notes?: string;
  supersedes_inbound_id?: string;
  lines: InboundOrderLineRequest[];
}

export interface UpdateInboundOrderRequest {
  expected_version: number;
  expected_arrival_at?: string;
  external_reference?: string;
  supplier_reference?: string;
  notes?: string;
}
