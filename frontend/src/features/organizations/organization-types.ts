import type { PaginatedData } from "@/lib/api/types";

export interface Organization {
  organization_id: string;
  code: string;
  name: string;
  legal_name?: string;
  tax_number?: string;
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

export type OrganizationPage = PaginatedData<Organization>;

export interface OrganizationListFilters {
  search: string;
  active: "all" | "active" | "inactive";
  page: number;
  pageSize: number;
}

export interface CreateOrganizationRequest {
  code: string;
  name: string;
  legal_name?: string;
  tax_number?: string;
  timezone_name: string;
  address_line_1?: string;
  address_line_2?: string;
  city?: string;
  province?: string;
  postal_code?: string;
  country_code?: string;
}

export interface UpdateOrganizationRequest extends Omit<
  CreateOrganizationRequest,
  "code"
> {
  is_active: boolean;
  expected_updated_at: string;
}
