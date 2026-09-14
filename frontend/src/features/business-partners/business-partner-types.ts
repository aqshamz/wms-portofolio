import type { PaginatedData } from "@/lib/api/types";

export type ActiveFilter = "all" | "active" | "inactive";
export type BusinessPartnerView = "partners" | "types";

export interface PartnerType {
  partner_type_id: string;
  code: string;
  name: string;
  description?: string;
  is_active: boolean;
}

export interface BusinessPartner {
  partner_id: string;
  owner_id: string;
  code: string;
  name: string;
  legal_name?: string;
  tax_number?: string;
  email?: string;
  phone?: string;
  address_line_1?: string;
  address_line_2?: string;
  city?: string;
  province?: string;
  postal_code?: string;
  country_code?: string;
  is_active: boolean;
  created_at: string;
  created_by?: string;
  updated_at: string;
  updated_by?: string;
}

export interface BusinessPartnerDetail extends BusinessPartner {
  partner_types: PartnerType[];
}

export type PartnerTypePage = PaginatedData<PartnerType>;
export type BusinessPartnerPage = PaginatedData<BusinessPartner>;

export interface PartnerTypeListFilters {
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface BusinessPartnerListFilters {
  ownerId: string;
  partnerTypeCode: string;
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface CreatePartnerTypeRequest {
  code: string;
  name: string;
  description?: string;
}

export interface UpdatePartnerTypeRequest {
  name: string;
  description?: string;
  is_active: boolean;
}

export interface BusinessPartnerProfileRequest {
  name: string;
  legal_name?: string;
  tax_number?: string;
  email?: string;
  phone?: string;
  address_line_1?: string;
  address_line_2?: string;
  city?: string;
  province?: string;
  postal_code?: string;
  country_code?: string;
}

export interface CreateBusinessPartnerRequest extends BusinessPartnerProfileRequest {
  owner_id: string;
  code: string;
}

export interface UpdateBusinessPartnerRequest extends BusinessPartnerProfileRequest {
  is_active: boolean;
  expected_updated_at: string;
}
