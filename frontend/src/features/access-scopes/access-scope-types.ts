import type { PaginatedData } from "@/lib/api/types";

export interface AccountStatus {
  account_status_id: string;
  code: string;
  name: string;
  allows_login: boolean;
}

export interface AccountSummary {
  account_id: string;
  username: string;
  email?: string;
  display_name: string;
  status: AccountStatus;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
  version_no: number;
}

export type AccountPage = PaginatedData<AccountSummary>;

export interface AccountListFilters {
  search: string;
  page: number;
  pageSize: number;
}

export interface OwnerAccess {
  account_id: string;
  owner_id: string;
  owner_code: string;
  owner_name: string;
  granted_by?: string;
  granted_at: string;
}

export interface WarehouseAccess {
  account_id: string;
  warehouse_id: string;
  warehouse_code: string;
  warehouse_name: string;
  granted_by?: string;
  granted_at: string;
}

export type ScopeKind = "owners" | "warehouses";
