import { apiRequest } from "@/lib/api/client";
import type {
  AdjustmentFilters,
  AdjustmentPage,
  CreateAdjustmentRequest,
  CycleCount,
  CycleCountFilters,
  CycleCountPage,
  InternalMoveRequest,
  InventoryAdjustment,
  StockControlCommandResponse,
  StockControlReason,
  WarehouseTransferRequest,
  WarehouseTransfer,
  WarehouseTransferFilters,
  WarehouseTransferPage,
} from "./stock-control-types";
import type {
  ReplenishmentAssignee,
  ReplenishmentFilters,
  ReplenishmentPage,
  ReplenishmentTarget,
  ReplenishmentTask,
} from "./replenishment-types";
import type { PaginatedData } from "@/lib/api/types";

const root = "/api/v1/stock-control";

export const stockControlKeys = {
  all: ["stock-control"] as const,
  reasons: () => [...stockControlKeys.all, "reasons"] as const,
  replenishments: () => [...stockControlKeys.all, "replenishments"] as const,
  replenishmentList: (filters: ReplenishmentFilters) =>
    [...stockControlKeys.replenishments(), "list", filters] as const,
  replenishment: (id: string) =>
    [...stockControlKeys.replenishments(), "detail", id] as const,
  replenishmentAssignees: (id: string, search: string) =>
    [...stockControlKeys.replenishments(), "assignees", id, search] as const,
  replenishmentTargets: (
    ownerId: string,
    warehouseId: string,
    search: string,
  ) =>
    [
      ...stockControlKeys.replenishments(),
      "targets",
      ownerId,
      warehouseId,
      search,
    ] as const,
  adjustments: () => [...stockControlKeys.all, "adjustments"] as const,
  adjustmentList: (filters: AdjustmentFilters) =>
    [...stockControlKeys.adjustments(), "list", filters] as const,
  adjustment: (id: string) =>
    [...stockControlKeys.adjustments(), "detail", id] as const,
  cycleCounts: () => [...stockControlKeys.all, "cycle-counts"] as const,
  cycleCountList: (filters: CycleCountFilters) =>
    [...stockControlKeys.cycleCounts(), "list", filters] as const,
  cycleCount: (id: string) =>
    [...stockControlKeys.cycleCounts(), "detail", id] as const,
  warehouseTransfers: () =>
    [...stockControlKeys.all, "warehouse-transfers"] as const,
  warehouseTransferList: (filters: WarehouseTransferFilters) =>
    [...stockControlKeys.warehouseTransfers(), "list", filters] as const,
  warehouseTransfer: (id: string) =>
    [...stockControlKeys.warehouseTransfers(), "detail", id] as const,
};

export function listStockControlReasons() {
  return apiRequest<StockControlReason[]>(`${root}/reason-codes?active=true`);
}

const replenishmentRoot = `${root}/replenishment-tasks`;

export function listReplenishments(filters: ReplenishmentFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.status) query.set("status_code", filters.status);
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return apiRequest<ReplenishmentPage>(`${replenishmentRoot}?${query}`);
}

export function getReplenishment(id: string) {
  return apiRequest<ReplenishmentTask>(
    `${replenishmentRoot}/${encodeURIComponent(id)}`,
  );
}

export function createReplenishment(request: {
  source_balance_id: string;
  target_location_id: string;
  serial_id?: string;
  quantity: string;
  priority_code: string;
  notes?: string;
  expected_balance_version: number;
}) {
  return apiRequest<ReplenishmentTask>(replenishmentRoot, {
    method: "POST",
    body: request,
  });
}

export function listReplenishmentAssignees(id: string, search: string) {
  const query = new URLSearchParams({ search, page: "1", page_size: "100" });
  return apiRequest<PaginatedData<ReplenishmentAssignee>>(
    `${replenishmentRoot}/${encodeURIComponent(id)}/assignees?${query}`,
  );
}

export function listReplenishmentTargets(
  ownerId: string,
  warehouseId: string,
  search: string,
) {
  const query = new URLSearchParams({
    owner_id: ownerId,
    warehouse_id: warehouseId,
    search,
    page: "1",
    page_size: "100",
  });
  return apiRequest<PaginatedData<ReplenishmentTarget>>(
    `${root}/replenishment-targets?${query}`,
  );
}

function replenishTransition(id: string, action: string, body: unknown) {
  return apiRequest<ReplenishmentTask>(
    `${replenishmentRoot}/${encodeURIComponent(id)}/${action}`,
    { method: "POST", body },
  );
}

export function assignReplenishment(
  id: string,
  expectedVersion: number,
  accountId: string,
) {
  return replenishTransition(id, "assign", {
    expected_version: expectedVersion,
    account_id: accountId,
  });
}

export function startReplenishment(id: string, expectedVersion: number) {
  return replenishTransition(id, "start", {
    expected_version: expectedVersion,
  });
}

export function completeReplenishment(
  id: string,
  expectedVersion: number,
  expectedBalanceVersion: number,
  businessDate: string,
) {
  return replenishTransition(id, "complete", {
    expected_version: expectedVersion,
    expected_balance_version: expectedBalanceVersion,
    business_date: businessDate,
  });
}

export function cancelReplenishment(
  id: string,
  expectedVersion: number,
  expectedBalanceVersion: number,
  reason: string,
) {
  return replenishTransition(id, "cancel", {
    expected_version: expectedVersion,
    expected_balance_version: expectedBalanceVersion,
    reason,
  });
}

export function postInternalMove(request: InternalMoveRequest) {
  return apiRequest<StockControlCommandResponse>(`${root}/internal-moves`, {
    method: "POST",
    body: request,
  });
}

export function postWarehouseTransfer(request: WarehouseTransferRequest) {
  return apiRequest<StockControlCommandResponse>(
    `${root}/warehouse-transfers`,
    {
      method: "POST",
      body: request,
    },
  );
}

const warehouseTransferRoot = `${root}/warehouse-transfer-documents`;
export function listWarehouseTransfers(filters: WarehouseTransferFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    side: filters.side,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.status) query.set("status_code", filters.status);
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return apiRequest<WarehouseTransferPage>(`${warehouseTransferRoot}?${query}`);
}
export function getWarehouseTransfer(id: string) {
  return apiRequest<WarehouseTransfer>(
    `${warehouseTransferRoot}/${encodeURIComponent(id)}`,
  );
}
export function createWarehouseTransfer(request: {
  business_date: string;
  source_balance_id: string;
  target_warehouse_id: string;
  quantity: string;
  serial_ids: string[];
  notes?: string;
}) {
  return apiRequest<WarehouseTransfer>(warehouseTransferRoot, {
    method: "POST",
    body: request,
  });
}
function warehouseTransferTransition(
  id: string,
  action: string,
  body: unknown,
) {
  return apiRequest<WarehouseTransfer>(
    `${warehouseTransferRoot}/${encodeURIComponent(id)}/${action}`,
    { method: "POST", body },
  );
}
export const approveWarehouseTransfer = (id: string, expectedVersion: number) =>
  warehouseTransferTransition(id, "approve", {
    expected_version: expectedVersion,
  });
export const dispatchWarehouseTransfer = (
  id: string,
  expectedVersion: number,
) =>
  warehouseTransferTransition(id, "dispatch", {
    expected_version: expectedVersion,
  });
export const cancelWarehouseTransfer = (
  id: string,
  expectedVersion: number,
  reason: string,
) =>
  warehouseTransferTransition(id, "cancel", {
    expected_version: expectedVersion,
    reason,
  });
export const receiveWarehouseTransfer = (
  id: string,
  body: {
    expected_version: number;
    business_date: string;
    receipt_location_id: string;
    putaway_target_location_id: string;
  },
) => warehouseTransferTransition(id, "receive", body);
export const putawayWarehouseTransfer = (
  id: string,
  body: {
    expected_version: number;
    expected_balance_version: number;
    business_date: string;
  },
) => warehouseTransferTransition(id, "putaway", body);

const adjustmentRoot = `${root}/adjustments`;

export function listAdjustments(filters: AdjustmentFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.status) query.set("status_code", filters.status);
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return apiRequest<AdjustmentPage>(`${adjustmentRoot}?${query}`);
}

export function getAdjustment(id: string) {
  return apiRequest<InventoryAdjustment>(
    `${adjustmentRoot}/${encodeURIComponent(id)}`,
  );
}

export function createAdjustment(request: CreateAdjustmentRequest) {
  return apiRequest<InventoryAdjustment>(adjustmentRoot, {
    method: "POST",
    body: request,
  });
}

function adjustmentTransition(id: string, action: string, body: unknown) {
  return apiRequest<InventoryAdjustment>(
    `${adjustmentRoot}/${encodeURIComponent(id)}/${action}`,
    { method: "POST", body },
  );
}

export function approveAdjustment(
  id: string,
  expectedVersion: number,
  lineIds: string[],
) {
  return adjustmentTransition(id, "approve", {
    expected_version: expectedVersion,
    line_ids: lineIds,
  });
}

export function cancelAdjustment(
  id: string,
  expectedVersion: number,
  reason: string,
) {
  return adjustmentTransition(id, "cancel", {
    expected_version: expectedVersion,
    reason,
  });
}

export function rejectAdjustment(
  id: string,
  expectedVersion: number,
  lineIds: string[],
  reason: string,
) {
  return adjustmentTransition(id, "reject", {
    expected_version: expectedVersion,
    line_ids: lineIds,
    reason,
  });
}

const cycleCountRoot = `${root}/cycle-counts`;
export function listCycleCounts(filters: CycleCountFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    count_type_code: filters.countType,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.status) query.set("status_code", filters.status);
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return apiRequest<CycleCountPage>(`${cycleCountRoot}?${query}`);
}
export function getCycleCount(id: string) {
  return apiRequest<CycleCount>(`${cycleCountRoot}/${encodeURIComponent(id)}`);
}
export function createCycleCount(request: {
  business_date: string;
  tolerance_quantity: string;
  notes?: string;
  balance_ids: string[];
}) {
  return apiRequest<CycleCount>(cycleCountRoot, {
    method: "POST",
    body: request,
  });
}
export function createGrandStockOpname(request: {
  owner_id: string;
  warehouse_id: string;
  business_date: string;
  tolerance_quantity: string;
  notes?: string;
}) {
  return apiRequest<CycleCount>(`${cycleCountRoot}/grand`, {
    method: "POST",
    body: request,
  });
}
function cycleCountTransition(id: string, action: string, body: unknown) {
  return apiRequest<CycleCount>(
    `${cycleCountRoot}/${encodeURIComponent(id)}/${action}`,
    { method: "POST", body },
  );
}
export function recordCycleCount(
  id: string,
  expectedVersion: number,
  lines: { line_id: string; counted_quantity: string; notes?: string }[],
) {
  return cycleCountTransition(id, "count", {
    expected_version: expectedVersion,
    lines,
  });
}
export function approveCycleCount(
  id: string,
  expectedVersion: number,
  lineIds: string[],
) {
  return cycleCountTransition(id, "approve", {
    expected_version: expectedVersion,
    line_ids: lineIds,
  });
}
export function rejectCycleCount(
  id: string,
  expectedVersion: number,
  lineIds: string[],
  reason: string,
) {
  return cycleCountTransition(id, "reject", {
    expected_version: expectedVersion,
    line_ids: lineIds,
    reason,
  });
}
export function cancelCycleCount(
  id: string,
  expectedVersion: number,
  reason: string,
) {
  return cycleCountTransition(id, "cancel", {
    expected_version: expectedVersion,
    reason,
  });
}
