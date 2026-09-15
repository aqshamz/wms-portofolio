import type { PaginatedData } from "@/lib/api/types";

export type ActiveFilter = "all" | "active" | "inactive";

export interface PutawayStrategy {
  putaway_strategy_id: string;
  owner_id?: string;
  warehouse_id?: string;
  code: string;
  name: string;
  description?: string;
  is_active: boolean;
}

export interface PutawayStrategyRule {
  rule_id: string;
  putaway_strategy_id: string;
  sequence_no: number;
  category_id?: string;
  location_type_id?: string;
  zone_id?: string;
  minimum_empty_percent?: string;
  is_active: boolean;
}

export type PutawayStrategyPage = PaginatedData<PutawayStrategy>;
export type PutawayStrategyRulePage = PaginatedData<PutawayStrategyRule>;

export interface PutawayListFilters {
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
  ownerId?: string;
  warehouseId?: string;
}

export interface CreatePutawayStrategyRequest {
  owner_id?: string;
  warehouse_id?: string;
  code: string;
  name: string;
  description?: string;
}

export interface UpdatePutawayStrategyRequest {
  name: string;
  description?: string;
  is_active: boolean;
}

export interface CreatePutawayStrategyRuleRequest {
  sequence_no: number;
  category_id?: string;
  location_type_id?: string;
  zone_id?: string;
  minimum_empty_percent?: string;
}

export interface UpdatePutawayStrategyRuleRequest extends CreatePutawayStrategyRuleRequest {
  is_active: boolean;
}
