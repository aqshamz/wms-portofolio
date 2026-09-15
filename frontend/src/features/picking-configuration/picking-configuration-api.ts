import { apiRequest } from "@/lib/api/client";
import type {
  CreatePickingSortMethodRequest,
  CreatePickingStrategyRequest,
  CreatePickingStrategyRuleRequest,
  PickingListFilters,
  PickingSortMethod,
  PickingSortMethodPage,
  PickingStrategy,
  PickingStrategyPage,
  PickingStrategyRule,
  PickingStrategyRulePage,
  UpdatePickingSortMethodRequest,
  UpdatePickingStrategyRequest,
  UpdatePickingStrategyRuleRequest,
} from "@/features/picking-configuration/picking-configuration-types";

const methodsPath = "/api/v1/master/picking-sort-methods";
const strategiesPath = "/api/v1/master/picking-strategies";

export const pickingConfigurationKeys = {
  all: ["picking-configuration"] as const,
  methods: (filters: PickingListFilters) =>
    [...pickingConfigurationKeys.all, "methods", filters] as const,
  method: (id: string) =>
    [...pickingConfigurationKeys.all, "method", id] as const,
  strategies: (filters: PickingListFilters) =>
    [...pickingConfigurationKeys.all, "strategies", filters] as const,
  strategy: (id: string) =>
    [...pickingConfigurationKeys.all, "strategy", id] as const,
  rules: (strategyId: string, filters: PickingListFilters) =>
    [...pickingConfigurationKeys.all, strategyId, "rules", filters] as const,
  rule: (strategyId: string, id: string) =>
    [...pickingConfigurationKeys.all, strategyId, "rule", id] as const,
  organizations: () =>
    [...pickingConfigurationKeys.all, "organizations"] as const,
  warehouses: () => [...pickingConfigurationKeys.all, "warehouses"] as const,
  inventoryStatuses: () =>
    [...pickingConfigurationKeys.all, "inventory-statuses"] as const,
  zones: (warehouseId: string) =>
    [...pickingConfigurationKeys.all, "zones", warehouseId] as const,
};

function listPath(base: string, filters: PickingListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all")
    query.set("active", String(filters.active === "active"));
  if (filters.ownerId) query.set("owner_id", filters.ownerId);
  if (filters.warehouseId) query.set("warehouse_id", filters.warehouseId);
  return `${base}?${query}`;
}

export function pickingStrategyListPath(filters: PickingListFilters) {
  return listPath(strategiesPath, filters);
}

export function listPickingSortMethods(filters: PickingListFilters) {
  return apiRequest<PickingSortMethodPage>(listPath(methodsPath, filters));
}
export function getPickingSortMethod(id: string) {
  return apiRequest<PickingSortMethod>(`${methodsPath}/${id}`);
}
export function createPickingSortMethod(
  request: CreatePickingSortMethodRequest,
) {
  return apiRequest<PickingSortMethod>(methodsPath, {
    method: "POST",
    body: request,
  });
}
export function updatePickingSortMethod(
  id: string,
  request: UpdatePickingSortMethodRequest,
) {
  return apiRequest<PickingSortMethod>(`${methodsPath}/${id}`, {
    method: "PUT",
    body: request,
  });
}
export function deactivatePickingSortMethod(id: string) {
  return apiRequest<PickingSortMethod>(`${methodsPath}/${id}/deactivate`, {
    method: "PATCH",
  });
}

export function listPickingStrategies(filters: PickingListFilters) {
  return apiRequest<PickingStrategyPage>(pickingStrategyListPath(filters));
}
export function getPickingStrategy(id: string) {
  return apiRequest<PickingStrategy>(`${strategiesPath}/${id}`);
}
export function createPickingStrategy(request: CreatePickingStrategyRequest) {
  return apiRequest<PickingStrategy>(strategiesPath, {
    method: "POST",
    body: request,
  });
}
export function updatePickingStrategy(
  id: string,
  request: UpdatePickingStrategyRequest,
) {
  return apiRequest<PickingStrategy>(`${strategiesPath}/${id}`, {
    method: "PUT",
    body: request,
  });
}
export function deactivatePickingStrategy(id: string) {
  return apiRequest<PickingStrategy>(`${strategiesPath}/${id}/deactivate`, {
    method: "PATCH",
  });
}

function rulesPath(strategyId: string) {
  return `${strategiesPath}/${strategyId}/rules`;
}
export function listPickingStrategyRules(
  strategyId: string,
  filters: PickingListFilters,
) {
  return apiRequest<PickingStrategyRulePage>(
    listPath(rulesPath(strategyId), filters),
  );
}
export function getPickingStrategyRule(strategyId: string, id: string) {
  return apiRequest<PickingStrategyRule>(`${rulesPath(strategyId)}/${id}`);
}
export function createPickingStrategyRule(
  strategyId: string,
  request: CreatePickingStrategyRuleRequest,
) {
  return apiRequest<PickingStrategyRule>(rulesPath(strategyId), {
    method: "POST",
    body: request,
  });
}
export function updatePickingStrategyRule(
  strategyId: string,
  id: string,
  request: UpdatePickingStrategyRuleRequest,
) {
  return apiRequest<PickingStrategyRule>(`${rulesPath(strategyId)}/${id}`, {
    method: "PUT",
    body: request,
  });
}
export function deactivatePickingStrategyRule(strategyId: string, id: string) {
  return apiRequest<PickingStrategyRule>(
    `${rulesPath(strategyId)}/${id}/deactivate`,
    { method: "PATCH" },
  );
}
