import type { PaginatedData } from "@/lib/api/types";

export type ActiveFilter = "all" | "active" | "inactive";
export type ItemCatalogView = "items" | "categories";

export interface ItemCategory {
  category_id: string;
  owner_id: string;
  parent_category_id?: string;
  code: string;
  name: string;
  is_active: boolean;
}

export interface UOM {
  uom_id: string;
  code: string;
  name: string;
  decimal_scale: number;
  is_active: boolean;
}

export interface ItemUOM {
  item_uom_id: string;
  item_id: string;
  uom_id: string;
  conversion_to_base: string;
  length?: string;
  width?: string;
  height?: string;
  weight?: string;
  is_receiving_uom: boolean;
  is_picking_uom: boolean;
  is_active: boolean;
}

export interface ItemBarcode {
  item_barcode_id: string;
  item_id: string;
  uom_id?: string;
  barcode: string;
  is_primary: boolean;
  is_active: boolean;
}

export interface CatalogItem {
  item_id: string;
  owner_id: string;
  category_id?: string;
  code: string;
  name: string;
  description?: string;
  base_uom_id: string;
  weight?: string;
  volume?: string;
  lot_controlled: boolean;
  serial_controlled: boolean;
  shelf_life_days?: number;
  minimum_receive_days?: number;
  is_active: boolean;
  created_at: string;
  created_by?: string;
  updated_at: string;
  updated_by?: string;
}

export interface CatalogItemDetail extends CatalogItem {
  uoms: ItemUOM[];
  barcodes: ItemBarcode[];
}

export type ItemPage = PaginatedData<CatalogItem>;
export type ItemCategoryPage = PaginatedData<ItemCategory>;
export type UOMPage = PaginatedData<UOM>;

export interface ItemListFilters {
  ownerId: string;
  categoryId: string;
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface ItemCategoryListFilters {
  ownerId: string;
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface UOMListFilters {
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface CreateItemCategoryRequest {
  owner_id: string;
  parent_category_id?: string;
  code: string;
  name: string;
}

export interface UpdateItemCategoryRequest {
  parent_category_id?: string;
  name: string;
  is_active: boolean;
}

export interface ItemProfileRequest {
  category_id?: string;
  name: string;
  description?: string;
  weight?: string;
  volume?: string;
  lot_controlled: boolean;
  serial_controlled: boolean;
  shelf_life_days?: number;
  minimum_receive_days?: number;
}

export interface CreateItemRequest extends ItemProfileRequest {
  owner_id: string;
  code: string;
  base_uom_id: string;
}

export interface UpdateItemRequest extends ItemProfileRequest {
  is_active: boolean;
  expected_updated_at: string;
}
