import { apiRequest } from "@/lib/api/client";
import type {
  Balance,
  BalanceFilters,
  BalancePage,
  Movement,
  MovementFilters,
  MovementPage,
  MovementType,
  SerialState,
  SerialStateFilters,
  SerialStatePage,
} from "./inventory-types";

const root = "/api/v1/inventory";

export const balanceKeys = {
  list: (filters: BalanceFilters) =>
    ["inventory", "balances", "list", filters] as const,
  detail: (id: string) => ["inventory", "balances", "detail", id] as const,
};

export function balanceListPath(filters: BalanceFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.includeZero) query.set("include_zero", "true");
  return `${root}/balances?${query}`;
}

export function listBalances(filters: BalanceFilters) {
  return apiRequest<BalancePage>(balanceListPath(filters));
}

export function getBalance(id: string) {
  return apiRequest<Balance>(`${root}/balances/${id}`);
}

export const movementKeys = {
  list: (filters: MovementFilters) =>
    ["inventory", "movements", "list", filters] as const,
  detail: (id: string) => ["inventory", "movements", "detail", id] as const,
  types: ["inventory", "movement-types"] as const,
};

export function movementListPath(filters: MovementFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.movementTypeId)
    query.set("movement_type_id", filters.movementTypeId);
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `${root}/movements?${query}`;
}

export function listMovements(filters: MovementFilters) {
  return apiRequest<MovementPage>(movementListPath(filters));
}

export function getMovement(id: string) {
  return apiRequest<Movement>(`${root}/movements/${id}`);
}

export function listMovementTypes() {
  return apiRequest<MovementType[]>(`${root}/movement-types?active=true`);
}

export const serialStateKeys = {
  list: (filters: SerialStateFilters) =>
    ["inventory", "serial-states", "list", filters] as const,
  detail: (id: string) => ["inventory", "serial-states", "detail", id] as const,
};

export function serialStateListPath(filters: SerialStateFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `${root}/serial-states?${query}`;
}

export function listSerialStates(filters: SerialStateFilters) {
  return apiRequest<SerialStatePage>(serialStateListPath(filters));
}

export function getSerialState(id: string) {
  return apiRequest<SerialState>(`${root}/serial-states/${id}`);
}
