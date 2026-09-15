import type { PaginatedData } from "@/lib/api/types";

export type ActiveFilter = "all" | "active" | "inactive";
export type PickingView = "methods" | "strategies";

export interface PickingSortMethod {
  picking_sort_method_id: string;
  code: string;
  name: string;
  description?: string;
  is_active: boolean;
}

export interface PickingStrategy {
  picking_strategy_id: string;
  owner_id?: string;
  warehouse_id?: string;
  code: string;
  name: string;
  description?: string;
  is_active: boolean;
}

export interface PickingStrategyRule {
  rule_id: string;
  picking_strategy_id: string;
  sequence_no: number;
  inventory_status_id?: string;
  zone_id?: string;
  picking_sort_method_id: string;
  is_active: boolean;
}

export type PickingSortMethodPage = PaginatedData<PickingSortMethod>;
export type PickingStrategyPage = PaginatedData<PickingStrategy>;
export type PickingStrategyRulePage = PaginatedData<PickingStrategyRule>;

export interface PickingListFilters {
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
  ownerId?: string;
  warehouseId?: string;
}

export interface CreatePickingSortMethodRequest {
  code: string;
  name: string;
  description?: string;
}

export interface UpdatePickingSortMethodRequest {
  name: string;
  description?: string;
  is_active: boolean;
}

export interface CreatePickingStrategyRequest {
  owner_id?: string;
  warehouse_id?: string;
  code: string;
  name: string;
  description?: string;
}

export interface UpdatePickingStrategyRequest {
  name: string;
  description?: string;
  is_active: boolean;
}

export interface CreatePickingStrategyRuleRequest {
  sequence_no: number;
  inventory_status_id?: string;
  zone_id?: string;
  picking_sort_method_id: string;
}

export interface UpdatePickingStrategyRuleRequest extends CreatePickingStrategyRuleRequest {
  is_active: boolean;
}
