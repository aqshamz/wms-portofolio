import type { InboundException } from "./inbound-exception-types";

export const testException: InboundException = {
  inbound_exception_id: "IEX-1",
  owner_id: "owner-1",
  warehouse_id: "warehouse-1",
  source_document_id: "RCV-1",
  source_line_id: "RCV-1-L001",
  exception_type_code: "DAMAGED",
  expected_qty: "10.000000",
  actual_qty: "8.000000",
  variance_qty: "-2.000000",
  notes: "Two damaged cartons\nVendor informed.",
  created_at: "2026-09-18T02:00:00Z",
  created_by: "account-1",
  created_by_display_name: "WMS Administrator",
  owner_name: "Science In Sport",
  warehouse_name: "Bandung WH",
};
