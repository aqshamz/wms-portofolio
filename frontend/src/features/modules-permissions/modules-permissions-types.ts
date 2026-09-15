import type { PaginatedData } from "@/lib/api/types";

export type ActiveFilter = "all" | "active" | "inactive";
export type ModulesPermissionsView = "modules" | "permissions";

export interface AppModule {
  module_id: string;
  code: string;
  name: string;
  display_order: number;
  is_active: boolean;
}

export interface WorkflowPermission {
  permission_id: string;
  code: string;
  name: string;
  module_code: string;
  description?: string;
  is_active: boolean;
  created_at: string;
}

export type AppModulePage = PaginatedData<AppModule>;
export type WorkflowPermissionPage = PaginatedData<WorkflowPermission>;

export interface ModulesPermissionsListFilters {
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
  moduleCode?: string;
}

export interface CreateAppModuleRequest {
  code: string;
  name: string;
  display_order: number;
}

export interface UpdateAppModuleRequest {
  name: string;
  display_order: number;
  is_active: boolean;
}
