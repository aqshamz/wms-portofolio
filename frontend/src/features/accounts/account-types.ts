import type { PaginatedData } from "@/lib/api/types";

export interface AccountStatus {
  account_status_id: string;
  code: string;
  name: string;
  allows_login: boolean;
}

export interface AuthenticationPolicy {
  authentication_policy_id: string;
  code: string;
  name: string;
  max_failed_attempts: number;
  lockout_seconds: number;
  session_ttl_seconds: number;
  is_default: boolean;
}

export interface AccountSummary {
  account_id: string;
  username: string;
  email?: string;
  display_name: string;
  status: AccountStatus;
  authentication_policy_id?: string;
  authentication_policy_code?: string;
  preferred_timezone?: string;
  failed_login_count: number;
  locked_until?: string;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
  version_no: number;
}

export interface AccountRole {
  account_id: string;
  role_id: string;
  code: string;
  name: string;
  is_active: boolean;
  assigned_at: string;
  assigned_by?: string;
}

export interface AccountPermission {
  account_id: string;
  permission_id: string;
  code: string;
  name: string;
  module_code: string;
  granted_at: string;
  granted_by?: string;
}

export interface AccountOwnerAccess {
  owner_id: string;
  owner_code: string;
  owner_name: string;
  granted_at: string;
  granted_by?: string;
}

export interface AccountWarehouseAccess {
  warehouse_id: string;
  warehouse_code: string;
  warehouse_name: string;
  granted_at: string;
  granted_by?: string;
}

export interface AccountDetail extends AccountSummary {
  active_session_count: number;
  roles: AccountRole[];
  direct_permissions: AccountPermission[];
  effective_permissions: string[];
  owner_access: AccountOwnerAccess[];
  warehouse_access: AccountWarehouseAccess[];
}

export type AccountPage = PaginatedData<AccountSummary>;

export interface AccountListFilters {
  search: string;
  statusId: string;
  page: number;
  pageSize: number;
}

export interface CreateAccountRequest {
  username: string;
  email?: string;
  display_name: string;
  password: string;
  account_status_id?: string;
  authentication_policy_id?: string;
  preferred_timezone?: string;
}

export interface UpdateAccountRequest {
  username: string;
  email?: string;
  display_name: string;
  authentication_policy_id?: string;
  preferred_timezone?: string;
  expected_version: number;
}
