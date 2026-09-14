import type { PaginatedData } from "@/lib/api/types";

export interface LocationType {
  location_type_id: string;
  code: string;
  name: string;
  description?: string;
  allows_receiving: boolean;
  allows_storage: boolean;
  allows_picking: boolean;
  allows_shipping: boolean;
  is_active: boolean;
}

export interface WarehouseZone {
  zone_id: string;
  warehouse_id: string;
  code: string;
  name: string;
  description?: string;
  is_active: boolean;
  location_count: number;
  created_at: string;
}

export interface WarehouseLocation {
  location_id: string;
  warehouse_id: string;
  warehouse_code?: string;
  warehouse_name?: string;
  zone_id: string;
  zone_code?: string;
  zone_name?: string;
  location_type_id: string;
  location_type_code?: string;
  location_type_name?: string;
  code: string;
  barcode?: string;
  aisle?: string;
  bay?: string;
  level_no?: string;
  position_no?: string;
  pick_sequence: number;
  max_weight?: string;
  max_volume?: string;
  is_pick_face: boolean;
  is_locked: boolean;
  is_active: boolean;
  created_at: string;
}

export type WarehouseLocationPage = PaginatedData<WarehouseLocation>;
export type ActiveFilter = "all" | "active" | "inactive";

export interface LocationListFilters {
  warehouseId: string;
  zoneId?: string;
  locationTypeId?: string;
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface CreateLocationTypeRequest {
  code: string;
  name: string;
  description?: string;
  allows_receiving: boolean;
  allows_storage: boolean;
  allows_picking: boolean;
  allows_shipping: boolean;
}

export interface UpdateLocationTypeRequest extends Omit<
  CreateLocationTypeRequest,
  "code"
> {
  is_active: boolean;
}

export interface CreateZoneRequest {
  code: string;
  name: string;
  description?: string;
}

export interface UpdateZoneRequest extends Omit<CreateZoneRequest, "code"> {
  is_active: boolean;
}

export interface CreateLocationRequest {
  zone_id: string;
  location_type_id: string;
  code: string;
  barcode?: string;
  aisle?: string;
  bay?: string;
  level_no?: string;
  position_no?: string;
  pick_sequence: number;
  max_weight?: string;
  max_volume?: string;
  is_pick_face: boolean;
}

export interface UpdateLocationRequest extends Omit<
  CreateLocationRequest,
  "code"
> {
  is_locked: boolean;
  is_active: boolean;
}
