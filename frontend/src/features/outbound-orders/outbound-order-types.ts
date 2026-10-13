import type { PaginatedData } from "@/lib/api/types";

export interface OutboundOrderLine {
  outbound_line_id: string;
  line_no: number;
  item_id: string;
  item_code: string;
  item_name: string;
  ordered_qty: string;
  uom_id: string;
  uom_code: string;
  uom_conversion_to_base: string;
  ordered_base_qty: string;
  base_uom_id: string;
  base_uom_code: string;
  allocated_qty: string;
  picked_qty: string;
  checked_qty: string;
  packed_qty: string;
  shipped_qty: string;
  short_accepted_qty?: string;
  closed_short_qty?: string;
  open_qty?: string;
  unallocated_qty?: string;
  requested_lot_no?: string;
  customer_line_reference?: string;
  notes?: string;
}
export interface OutboundOrder {
  outbound_id: string;
  owner_id: string;
  owner_code: string;
  warehouse_id: string;
  warehouse_code: string;
  customer_id: string;
  customer_code: string;
  customer_name: string;
  ship_to_partner_id?: string;
  ship_to_code: string;
  ship_to_name: string;
  ship_to_address_1: string;
  ship_to_address_2?: string;
  ship_to_city?: string;
  business_date: string;
  requested_ship_at?: string;
  client_delivery_order_no: string;
  customer_order_no?: string;
  external_reference?: string;
  notes?: string;
  status_code: string;
  version_no: number;
  created_at: string;
  closed_short_at?: string;
  closed_short_reason?: string;
  lines?: OutboundOrderLine[];
}
export interface OutboundOrderFilters {
  ownerId: string;
  warehouseId: string;
  status: string;
  search: string;
  page: number;
  pageSize: number;
}
export type OutboundOrderPage = PaginatedData<OutboundOrder>;
export interface OutboundLineRequest {
  item_id: string;
  ordered_qty: string;
  uom_id?: string;
  requested_lot_no?: string;
  customer_line_reference?: string;
  notes?: string;
}
export interface OutboundHeaderRequest {
  ship_to_partner_id: string;
  requested_ship_at: string;
  customer_order_no?: string;
  external_reference?: string;
  notes?: string;
}
export interface CreateOutboundOrderRequest extends OutboundHeaderRequest {
  owner_id: string;
  warehouse_id: string;
  customer_id: string;
  business_date: string;
  client_delivery_order_no: string;
  lines: OutboundLineRequest[];
}
export interface OutboundValidationRun {
  validation_run_id: string;
  status_code: string;
  results: {
    rule_code: string;
    rule_name: string;
    severity_code: string;
    passed: boolean;
    result_message?: string;
  }[];
}
