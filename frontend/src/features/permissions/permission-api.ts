import { apiRequest } from "@/lib/api/client";
import type {
  CreateRoleRequest,
  Permission,
  Role,
  RoleListFilters,
  RolePage,
  UpdateRoleRequest,
} from "@/features/permissions/permission-types";

const roleBasePath = "/api/v1/security/roles";

export const permissionKeys = {
  all: ["permissions"] as const,
  available: () => [...permissionKeys.all, "available"] as const,
  roleLists: () => [...permissionKeys.all, "roles", "list"] as const,
  roleList: (filters: RoleListFilters) =>
    [...permissionKeys.roleLists(), filters] as const,
  role: (roleId: string) =>
    [...permissionKeys.all, "roles", "detail", roleId] as const,
};

export function roleListPath(filters: RoleListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all") {
    query.set("active", String(filters.active === "active"));
  }
  return `${roleBasePath}?${query}`;
}

export function listPermissions() {
  return apiRequest<Permission[]>("/api/v1/security/permissions");
}

export function listRoles(filters: RoleListFilters) {
  return apiRequest<RolePage>(roleListPath(filters));
}

export function getRole(roleId: string) {
  return apiRequest<Role>(`${roleBasePath}/${roleId}`);
}

export function createRole(request: CreateRoleRequest) {
  return apiRequest<Role>(roleBasePath, { method: "POST", body: request });
}

export function updateRole(roleId: string, request: UpdateRoleRequest) {
  return apiRequest<Role>(`${roleBasePath}/${roleId}`, {
    method: "PUT",
    body: request,
  });
}

export function replaceRolePermissions(
  roleId: string,
  permissionIds: string[],
  expectedVersion: number,
) {
  return apiRequest<Role>(`${roleBasePath}/${roleId}/permissions`, {
    method: "PUT",
    body: {
      permission_ids: permissionIds,
      expected_version: expectedVersion,
    },
  });
}

export function changeRoleStatus(
  roleId: string,
  isActive: boolean,
  expectedVersion: number,
) {
  return apiRequest<Role>(`${roleBasePath}/${roleId}/status`, {
    method: "PATCH",
    body: { is_active: isActive, expected_version: expectedVersion },
  });
}

export function assignAccountRole(accountId: string, roleId: string) {
  return apiRequest<void>(`/api/v1/security/accounts/${accountId}/roles`, {
    method: "POST",
    body: { role_id: roleId },
  });
}

export function revokeAccountRole(accountId: string, roleId: string) {
  return apiRequest<void>(
    `/api/v1/security/accounts/${accountId}/roles/${roleId}`,
    { method: "DELETE" },
  );
}

export function grantAccountPermission(
  accountId: string,
  permissionId: string,
) {
  return apiRequest<void>(
    `/api/v1/security/accounts/${accountId}/permissions`,
    { method: "POST", body: { permission_id: permissionId } },
  );
}

export function revokeAccountPermission(
  accountId: string,
  permissionId: string,
) {
  return apiRequest<void>(
    `/api/v1/security/accounts/${accountId}/permissions/${permissionId}`,
    { method: "DELETE" },
  );
}
