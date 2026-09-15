import type { PaginatedData } from "@/lib/api/types";

export type ActiveFilter = "all" | "active" | "inactive";

export interface InventoryStatus {
  inventory_status_id: string;
  code: string;
  name: string;
  description?: string;
  is_allocatable: boolean;
  is_pickable: boolean;
  is_active: boolean;
}

export type InventoryStatusPage = PaginatedData<InventoryStatus>;

export interface InventoryStatusListFilters {
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface CreateInventoryStatusRequest {
  code: string;
  name: string;
  description?: string;
  is_allocatable: boolean;
  is_pickable: boolean;
}

export interface UpdateInventoryStatusRequest {
  name: string;
  description?: string;
  is_allocatable: boolean;
  is_pickable: boolean;
  is_active: boolean;
}
