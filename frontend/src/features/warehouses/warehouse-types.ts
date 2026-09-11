import type { PaginatedData } from "@/lib/api/types";

export interface Warehouse {
  warehouse_id: string;
  operator_id: string;
  operator_code?: string;
  operator_name?: string;
  code: string;
  name: string;
  timezone_name: string;
  address_line_1?: string;
  address_line_2?: string;
  city?: string;
  province?: string;
  postal_code?: string;
  country_code?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export type WarehousePage = PaginatedData<Warehouse>;

export interface WarehouseListFilters {
  search: string;
  active: "all" | "active" | "inactive";
  page: number;
  pageSize: number;
}

export interface CreateWarehouseRequest {
  operator_id: string;
  code: string;
  name: string;
  timezone_name: string;
  address_line_1?: string;
  address_line_2?: string;
  city?: string;
  province?: string;
  postal_code?: string;
  country_code?: string;
}

export interface UpdateWarehouseRequest extends Omit<
  CreateWarehouseRequest,
  "operator_id" | "code"
> {
  is_active: boolean;
  expected_updated_at: string;
}
