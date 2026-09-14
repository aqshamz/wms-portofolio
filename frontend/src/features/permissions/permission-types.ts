import type { PaginatedData } from "@/lib/api/types";

export interface Permission {
  permission_id: string;
  code: string;
  name: string;
  module_code: string;
}

export interface Role {
  role_id: string;
  code: string;
  name: string;
  description?: string;
  is_active: boolean;
  permissions: Permission[];
  created_at: string;
  updated_at: string;
  version_no: number;
}

export type RolePage = PaginatedData<Role>;
export type ActiveFilter = "all" | "active" | "inactive";

export interface RoleListFilters {
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface CreateRoleRequest {
  code: string;
  name: string;
  description?: string;
  permission_ids: string[];
}

export interface UpdateRoleRequest {
  name: string;
  description?: string;
  expected_version: number;
}

export type PermissionView = "roles" | "accounts";
