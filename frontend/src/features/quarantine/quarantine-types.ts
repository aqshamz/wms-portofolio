import Decimal from "decimal.js";
import type { PaginatedData } from "@/lib/api/types";
import type {
  LookupFilters,
  PutawayFilters,
  PutawayTarget,
} from "@/features/putaway/putaway-types";

export type QuarantineStatus = "OPEN" | "PARTIALLY_DECIDED" | "CLOSED";
export interface DispositionType {
  quarantine_disposition_type_id: string;
  code: string;
  name: string;
  description?: string | null;
  releases_to_available: boolean;
  requires_reinspection: boolean;
  removes_inventory: boolean;
  is_active: boolean;
}
export interface QuarantineDisposition {
  quarantine_disposition_id: string;
  quarantine_case_id: string;
  disposition_type_code: string;
  status_code: string;
  disposition_qty: string;
  uom_id: string;
  client_decision_reference?: string | null;
  decision_notes?: string | null;
  decided_at: string;
  decided_by: string;
  processed_at?: string | null;
  inventory_movement_id?: string | null;
  resulting_balance_id?: string | null;
  target_location_id?: string | null;
  target_location_code?: string | null;
  created_at: string;
  rework_task?: {
    rework_task_id: string;
    task_status_code: string;
    planned_qty: string;
    completed_qty: string;
    work_instructions?: string | null;
    result_notes?: string | null;
    reinspection_id?: string | null;
  };
}
export interface QuarantineCase {
  quarantine_case_id: string;
  parent_quarantine_case_id?: string | null;
  inspection_id: string;
  receipt_inventory_id: string;
  quarantine_balance_id: string;
  quarantine_balance_version_no?: number;
  available_qty?: string;
  base_uom_code?: string;
  is_indivisible?: boolean;
  serial_no?: string | null;
  handling_unit_barcode?: string | null;
  owner_id: string;
  warehouse_id: string;
  item_id: string;
  item_code: string;
  lot_number?: string;
  location_code: string;
  status_code: QuarantineStatus;
  quarantine_qty: string;
  disposed_qty: string;
  uom_id: string;
  opened_at: string;
  closed_at?: string | null;
  notes?: string | null;
  version_no: number;
  dispositions: QuarantineDisposition[];
}
export type QuarantineFilters = PutawayFilters;
export type QuarantinePage = PaginatedData<QuarantineCase>;
export type QuarantineTarget = PutawayTarget;
export type QuarantineLookupFilters = LookupFilters;
export interface QuarantineCapabilities {
  canDispose: boolean;
  timezone: string;
}
export interface DispositionRequest {
  expected_case_version: number;
  expected_balance_version: number;
  disposition_type_code: string;
  disposition_qty: string;
  business_date: string;
  decided_at: string;
  target_location_id?: string;
  client_decision_reference?: string;
  decision_notes?: string;
  work_instructions?: string;
}
export function remainingQuantity(value: QuarantineCase) {
  return new Decimal(value.quarantine_qty).minus(value.disposed_qty).toString();
}
export function canDecide(
  value: QuarantineCase,
  capabilities: QuarantineCapabilities,
) {
  return (
    capabilities.canDispose &&
    ["OPEN", "PARTIALLY_DECIDED"].includes(value.status_code) &&
    new Decimal(remainingQuantity(value)).gt(0)
  );
}
export function quarantineTone(status: QuarantineStatus) {
  return status === "CLOSED"
    ? "success"
    : status === "PARTIALLY_DECIDED"
      ? "warning"
      : "danger";
}
export function quarantineLabel(status: string) {
  return status
    .toLowerCase()
    .replaceAll("_", " ")
    .replace(/^./, (first) => first.toUpperCase());
}
// Mirrors the backend's configured-type precedence instead of guessing by code.
export function dispositionEffect(type?: DispositionType) {
  if (type?.requires_reinspection) return "rework";
  if (type?.releases_to_available) return "accept";
  if (type?.removes_inventory) return "remove";
  return undefined;
}
