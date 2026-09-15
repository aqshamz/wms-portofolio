import { apiRequest } from "@/lib/api/client";
import type {
  AppModule,
  AppModulePage,
  CreateAppModuleRequest,
  ModulesPermissionsListFilters,
  UpdateAppModuleRequest,
  WorkflowPermissionPage,
} from "@/features/modules-permissions/modules-permissions-types";

const modulePath = "/api/v1/master/modules";
const permissionPath = "/api/v1/master/workflow-permissions";

export const modulesPermissionsKeys = {
  all: ["modules-permissions"] as const,
  modules: () => [...modulesPermissionsKeys.all, "modules"] as const,
  moduleList: (filters: ModulesPermissionsListFilters) =>
    [...modulesPermissionsKeys.modules(), "list", filters] as const,
  moduleDetail: (id: string) =>
    [...modulesPermissionsKeys.modules(), "detail", id] as const,
  permissions: () => [...modulesPermissionsKeys.all, "permissions"] as const,
  permissionList: (filters: ModulesPermissionsListFilters) =>
    [...modulesPermissionsKeys.permissions(), "list", filters] as const,
};

export function modulesPermissionsListPath(
  resource: "modules" | "permissions",
  filters: ModulesPermissionsListFilters,
) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all") {
    query.set("active", String(filters.active === "active"));
  }
  if (resource === "permissions" && filters.moduleCode?.trim()) {
    query.set("module_code", filters.moduleCode.trim());
  }
  const path = resource === "modules" ? modulePath : permissionPath;
  return `${path}?${query}`;
}

export function listAppModules(filters: ModulesPermissionsListFilters) {
  return apiRequest<AppModulePage>(
    modulesPermissionsListPath("modules", filters),
  );
}

export function getAppModule(id: string) {
  return apiRequest<AppModule>(`${modulePath}/${id}`);
}

export function createAppModule(request: CreateAppModuleRequest) {
  return apiRequest<AppModule>(modulePath, { method: "POST", body: request });
}

export function updateAppModule(id: string, request: UpdateAppModuleRequest) {
  return apiRequest<AppModule>(`${modulePath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateAppModule(id: string) {
  return apiRequest<AppModule>(`${modulePath}/${id}/deactivate`, {
    method: "PATCH",
  });
}

export function listWorkflowPermissions(
  filters: ModulesPermissionsListFilters,
) {
  return apiRequest<WorkflowPermissionPage>(
    modulesPermissionsListPath("permissions", filters),
  );
}
