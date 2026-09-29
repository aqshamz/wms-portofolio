import { apiRequest } from "@/lib/api/client";
import type {
  AdjustmentFilters,
  AdjustmentPage,
  CreateAdjustmentRequest,
  InternalMoveRequest,
  InventoryAdjustment,
  StockControlCommandResponse,
  StockControlReason,
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
