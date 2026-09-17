import type { PutawayCapabilities, PutawayTask } from "./putaway-types";

export const testTask: PutawayTask = {
  putaway_task_id: "PUT-1",
  inspection_id: "QC-1",
  receipt_inventory_id: "RCV-1-L0001-B0001",
  source_balance_id: "balance-1",
  source_balance_version_no: 7,
  base_uom_code: "EA",
  owner_id: "owner-1",
  warehouse_id: "wh-1",
  item_id: "item-1",
  item_code: "COFFEE",
  source_location_id: "rcv-1",
  source_location_code: "RCV-01",
  target_location_id: "target-1",
  target_location_code: "BULK-01",
  planned_qty: "10",
  completed_qty: "0",
  uom_id: "ea",
  task_status_code: "OPEN",
  task_priority_code: "NORMAL",
  version_no: 1,
  created_at: "2026-09-17T09:00:00Z",
};
export const testCapabilities: PutawayCapabilities = {
  accountId: "account-1",
  canPutaway: true,
  canAssign: true,
  canCancel: true,
  timezone: "Asia/Jakarta",
};
