import type { ReworkCapabilities, ReworkTask } from "./rework-types";

export const capabilities: ReworkCapabilities = {
  accountId: "worker-1",
  canRework: true,
  timezone: "Asia/Jakarta",
};
export const testTask: ReworkTask = {
  rework_task_id: "RWK-1",
  quarantine_disposition_id: "QDIS-1",
  quarantine_case_id: "QCASE-1",
  source_balance_id: "balance-1",
  owner_id: "owner-1",
  warehouse_id: "warehouse-1",
  item_id: "item-1",
  item_code: "ITEM1",
  task_status_code: "OPEN",
  task_priority_code: "NORMAL",
  planned_qty: "2.000000",
  completed_qty: "0.000000",
  uom_id: "uom-1",
  assigned_to: null,
  work_instructions: "Replace damaged seals",
  result_notes: null,
  started_at: null,
  completed_at: null,
  reinspection_id: null,
  version_no: 1,
  created_at: "2026-09-18T02:00:00Z",
};
