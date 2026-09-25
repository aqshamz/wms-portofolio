import { apiRequest } from "@/lib/api/client";
import type {
  Balance,
  BalanceFilters,
  BalancePage,
  Movement,
  MovementFilters,
  MovementPage,
  MovementType,
  InventoryLot,
  LotFilters,
  LotPage,
  SerialState,
  SerialStateFilters,
  SerialStatePage,
  InventorySerial,
  SerialFilters,
  SerialPage,
  HandlingUnit,
  HandlingUnitFilters,
  HandlingUnitPage,
  CreateHandlingUnitRequest,
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
  if (filters.handlingUnitId) {
    query.set("handling_unit_id", filters.handlingUnitId);
  }
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

export const lotKeys = {
  list: (filters: LotFilters) =>
    ["inventory", "lots", "list", filters] as const,
  detail: (id: string) => ["inventory", "lots", "detail", id] as const,
};

export function lotListPath(filters: LotFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `${root}/lots?${query}`;
}

export function listLots(filters: LotFilters) {
  return apiRequest<LotPage>(lotListPath(filters));
}

export function getLot(id: string) {
  return apiRequest<InventoryLot>(`${root}/lots/${id}`);
}

export const serialKeys = {
  list: (filters: SerialFilters) =>
    ["inventory", "serials", "list", filters] as const,
  detail: (id: string) => ["inventory", "serials", "detail", id] as const,
};

export function serialListPath(filters: SerialFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `${root}/serials?${query}`;
}

export function listSerials(filters: SerialFilters) {
  return apiRequest<SerialPage>(serialListPath(filters));
}

export function getSerial(id: string) {
  return apiRequest<InventorySerial>(`${root}/serials/${id}`);
}

export const handlingUnitKeys = {
  list: (filters: HandlingUnitFilters) =>
    ["inventory", "handling-units", "list", filters] as const,
  detail: (id: string) =>
    ["inventory", "handling-units", "detail", id] as const,
};

export function createHandlingUnit(request: CreateHandlingUnitRequest) {
  return apiRequest<HandlingUnit>(`${root}/handling-units`, {
    method: "POST",
    body: request,
  });
}

export function handlingUnitListPath(filters: HandlingUnitFilters) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.status !== "all") {
    query.set("closed", String(filters.status === "closed"));
  }
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.locationId) query.set("current_location_id", filters.locationId);
  if (filters.parentId) query.set("parent_handling_unit_id", filters.parentId);
  return `${root}/handling-units?${query}`;
}

export function listHandlingUnits(filters: HandlingUnitFilters) {
  return apiRequest<HandlingUnitPage>(handlingUnitListPath(filters));
}

export function getHandlingUnit(id: string) {
  return apiRequest<HandlingUnit>(`${root}/handling-units/${id}`);
}
