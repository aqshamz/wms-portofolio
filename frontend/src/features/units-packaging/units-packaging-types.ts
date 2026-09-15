import type { PaginatedData } from "@/lib/api/types";
import type {
  ItemBarcode,
  ItemUOM,
  UOM,
} from "@/features/item-catalog/item-catalog-types";

export type ActiveFilter = "all" | "active" | "inactive";
export type UnitsPackagingView =
  "uoms" | "handling-units" | "item-uoms" | "barcodes";

export type { ItemBarcode, ItemUOM, UOM };

export interface HandlingUnitType {
  handling_unit_type_id: string;
  code: string;
  name: string;
  max_weight?: string;
  max_volume?: string;
  is_active: boolean;
}

export type HandlingUnitTypePage = PaginatedData<HandlingUnitType>;
export type UOMPage = PaginatedData<UOM>;

export interface ReferenceListFilters {
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface CreateUOMRequest {
  code: string;
  name: string;
  decimal_scale: number;
}

export interface UpdateUOMRequest {
  name: string;
  decimal_scale: number;
  is_active: boolean;
}

export interface CreateHandlingUnitTypeRequest {
  code: string;
  name: string;
  max_weight?: string;
  max_volume?: string;
}

export interface UpdateHandlingUnitTypeRequest {
  name: string;
  max_weight?: string;
  max_volume?: string;
  is_active: boolean;
}

export interface ItemUOMProfileRequest {
  conversion_to_base: string;
  length?: string;
  width?: string;
  height?: string;
  weight?: string;
  is_receiving_uom: boolean;
  is_picking_uom: boolean;
}

export interface CreateItemUOMRequest extends ItemUOMProfileRequest {
  uom_id: string;
}

export interface UpdateItemUOMRequest extends ItemUOMProfileRequest {
  is_active: boolean;
}

export interface CreateItemBarcodeRequest {
  uom_id?: string;
  barcode: string;
  is_primary: boolean;
}

export interface UpdateItemBarcodeRequest {
  uom_id?: string;
  barcode: string;
  is_active: boolean;
}
