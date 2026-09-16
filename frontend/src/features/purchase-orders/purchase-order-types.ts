import type { PaginatedData } from "@/lib/api/types";

export type PurchaseOrderStatus =
  | "DRAFT"
  | "APPROVED"
  | "PARTIALLY_RECEIVED"
  | "RECEIVED"
  | "CLOSED"
  | "CANCELLED";

export interface PurchaseOrderLine {
  purchase_order_line_id: string;
  line_no: number;
  item_id: string;
  item_code: string;
  item_name: string;
  ordered_qty: string;
  scheduled_qty: string;
  completed_receipt_qty: string;
  over_receipt_tolerance_pct: string;
  under_receipt_tolerance_pct: string;
  uom_id: string;
  uom_code: string;
  vendor_item_code?: string;
  expected_lot_no?: string;
  expected_expiry_date?: string;
  notes?: string;
}

export interface PurchaseOrder {
  purchase_order_id: string;
  owner_id: string;
  owner_code: string;
  vendor_id: string;
  vendor_code: string;
  vendor_name: string;
  warehouse_id: string;
  warehouse_code: string;
  business_date: string;
  purchase_order_no: string;
  ordered_at: string;
  expected_arrival_at?: string;
  status_code: PurchaseOrderStatus;
  notes?: string;
  supersedes_purchase_order_id?: string;
  successor_purchase_order_id?: string;
  version_no: number;
  created_at: string;
  lines?: PurchaseOrderLine[];
}

export type PurchaseOrderPage = PaginatedData<PurchaseOrder>;

export interface PurchaseOrderListFilters {
  ownerId: string;
  warehouseId: string;
  status: string;
  search: string;
  page: number;
  pageSize: number;
}

export interface PurchaseOrderLineRequest {
  item_id: string;
  ordered_qty: string;
  uom_id: string;
  vendor_item_code?: string;
  expected_lot_no?: string;
  expected_expiry_date?: string;
  notes?: string;
  over_receipt_tolerance_pct?: string;
  under_receipt_tolerance_pct?: string;
}

export interface CreatePurchaseOrderRequest {
  owner_id: string;
  vendor_id: string;
  warehouse_id: string;
  business_date: string;
  purchase_order_no: string;
  ordered_at: string;
  expected_arrival_at?: string;
  notes?: string;
  lines: PurchaseOrderLineRequest[];
}

export interface UpdatePurchaseOrderRequest {
  expected_version: number;
  purchase_order_no: string;
  ordered_at: string;
  expected_arrival_at?: string;
  notes?: string;
}
