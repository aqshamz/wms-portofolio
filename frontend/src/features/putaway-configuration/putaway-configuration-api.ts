import { apiRequest } from "@/lib/api/client";
import type {
  CreatePutawayStrategyRequest,
  CreatePutawayStrategyRuleRequest,
  PutawayListFilters,
  PutawayStrategy,
  PutawayStrategyPage,
  PutawayStrategyRule,
  PutawayStrategyRulePage,
  UpdatePutawayStrategyRequest,
  UpdatePutawayStrategyRuleRequest,
} from "@/features/putaway-configuration/putaway-configuration-types";

const strategiesPath = "/api/v1/master/putaway-strategies";

export const putawayConfigurationKeys = {
  all: ["putaway-configuration"] as const,
  strategies: (filters: PutawayListFilters) =>
    [...putawayConfigurationKeys.all, "strategies", filters] as const,
  strategy: (id: string) =>
    [...putawayConfigurationKeys.all, "strategy", id] as const,
  rules: (strategyId: string, filters: PutawayListFilters) =>
    [...putawayConfigurationKeys.all, strategyId, "rules", filters] as const,
  rule: (strategyId: string, id: string) =>
    [...putawayConfigurationKeys.all, strategyId, "rule", id] as const,
  organizations: () =>
    [...putawayConfigurationKeys.all, "organizations"] as const,
  warehouses: () => [...putawayConfigurationKeys.all, "warehouses"] as const,
  categories: (ownerId: string) =>
    [...putawayConfigurationKeys.all, "categories", ownerId] as const,
  locationTypes: () =>
    [...putawayConfigurationKeys.all, "location-types"] as const,
  zones: (warehouseId: string) =>
    [...putawayConfigurationKeys.all, "zones", warehouseId] as const,
};

function listPath(base: string, filters: PutawayListFilters) {
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

export function putawayStrategyListPath(filters: PutawayListFilters) {
  return listPath(strategiesPath, filters);
}

export function listPutawayStrategies(filters: PutawayListFilters) {
  return apiRequest<PutawayStrategyPage>(putawayStrategyListPath(filters));
}
export function getPutawayStrategy(id: string) {
  return apiRequest<PutawayStrategy>(`${strategiesPath}/${id}`);
}
export function createPutawayStrategy(request: CreatePutawayStrategyRequest) {
  return apiRequest<PutawayStrategy>(strategiesPath, {
    method: "POST",
    body: request,
  });
}
export function updatePutawayStrategy(
  id: string,
  request: UpdatePutawayStrategyRequest,
) {
  return apiRequest<PutawayStrategy>(`${strategiesPath}/${id}`, {
    method: "PUT",
    body: request,
  });
}
export function deactivatePutawayStrategy(id: string) {
  return apiRequest<PutawayStrategy>(`${strategiesPath}/${id}/deactivate`, {
    method: "PATCH",
  });
}

function rulesPath(strategyId: string) {
  return `${strategiesPath}/${strategyId}/rules`;
}
export function listPutawayStrategyRules(
  strategyId: string,
  filters: PutawayListFilters,
) {
  return apiRequest<PutawayStrategyRulePage>(
    listPath(rulesPath(strategyId), filters),
  );
}
export function getPutawayStrategyRule(strategyId: string, id: string) {
  return apiRequest<PutawayStrategyRule>(`${rulesPath(strategyId)}/${id}`);
}
export function createPutawayStrategyRule(
  strategyId: string,
  request: CreatePutawayStrategyRuleRequest,
) {
  return apiRequest<PutawayStrategyRule>(rulesPath(strategyId), {
    method: "POST",
    body: request,
  });
}
export function updatePutawayStrategyRule(
  strategyId: string,
  id: string,
  request: UpdatePutawayStrategyRuleRequest,
) {
  return apiRequest<PutawayStrategyRule>(`${rulesPath(strategyId)}/${id}`, {
    method: "PUT",
    body: request,
  });
}
export function deactivatePutawayStrategyRule(strategyId: string, id: string) {
  return apiRequest<PutawayStrategyRule>(
    `${rulesPath(strategyId)}/${id}/deactivate`,
    { method: "PATCH" },
  );
}
