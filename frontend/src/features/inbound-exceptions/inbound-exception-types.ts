import type { PaginatedData } from "@/lib/api/types";

export const EXCEPTION_TYPES = [
  "OVER_RECEIPT",
  "UNDER_RECEIPT",
  "REJECTED_AT_DOCK",
  "DAMAGED",
  "WRONG_ITEM",
  "CANCELLATION",
  "REVERSAL",
] as const;

export interface InboundException {
  inbound_exception_id: string;
  owner_id: string;
  warehouse_id: string;
  source_document_id: string;
  source_line_id: string | null;
  exception_type_code: string;
  expected_qty: string | null;
  actual_qty: string | null;
  variance_qty: string | null;
  notes: string | null;
  created_at: string;
  created_by: string;
  created_by_display_name?: string | null;
  owner_name?: string | null;
  warehouse_name?: string | null;
}
export interface ExceptionFilters {
  ownerId: string;
  warehouseId: string;
  exceptionType: string;
  search: string;
  page: number;
  pageSize: number;
}
export type ExceptionPage = PaginatedData<InboundException>;

export function exceptionLabel(code: string) {
  return code
    .toLowerCase()
    .replaceAll("_", " ")
    .replace(/^./, (first) => first.toUpperCase());
}
export function exceptionTone(code: string) {
  return ["DAMAGED", "WRONG_ITEM", "REVERSAL"].includes(code)
    ? "danger"
    : "warning";
}
export function exceptionTime(value: string, timezone: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  const options: Intl.DateTimeFormatOptions = {
    dateStyle: "medium",
    timeStyle: "short",
  };
  try {
    return new Intl.DateTimeFormat("en-GB", {
      ...options,
      timeZone: timezone,
    }).format(date);
  } catch {
    return new Intl.DateTimeFormat("en-GB", {
      ...options,
      timeZone: "Asia/Jakarta",
    }).format(date);
  }
}
