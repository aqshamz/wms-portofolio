import type { PaginatedData } from "@/lib/api/types";

export type ActiveFilter = "all" | "active" | "inactive";
export type WorkflowView = "statuses" | "transitions";

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
}

export interface DocumentType {
  document_type_id: string;
  code: string;
  name: string;
  module_code: string;
  description?: string;
  is_active: boolean;
}

export interface DocumentStatus {
  status_id: string;
  document_type_id: string;
  code: string;
  name: string;
  description?: string;
  is_initial: boolean;
  is_final: boolean;
  is_cancelled: boolean;
  display_order: number;
  is_active: boolean;
}

export interface DocumentTransition {
  transition_id: string;
  document_type_id: string;
  from_status_id: string;
  to_status_id: string;
  required_permission_id?: string;
  is_active: boolean;
}

export type DocumentTypePage = PaginatedData<DocumentType>;
export type DocumentStatusPage = PaginatedData<DocumentStatus>;
export type DocumentTransitionPage = PaginatedData<DocumentTransition>;
export type AppModulePage = PaginatedData<AppModule>;
export type WorkflowPermissionPage = PaginatedData<WorkflowPermission>;

export interface WorkflowListFilters {
  search: string;
  active: ActiveFilter;
  page?: number;
  pageSize?: number;
}

export interface CreateDocumentTypeRequest {
  code: string;
  name: string;
  module_code: string;
  description?: string;
}

export interface UpdateDocumentTypeRequest {
  name: string;
  module_code: string;
  description?: string;
  is_active: boolean;
}

export interface CreateDocumentStatusRequest {
  code: string;
  name: string;
  description?: string;
  is_initial: boolean;
  is_final: boolean;
  is_cancelled: boolean;
  display_order: number;
}

export interface UpdateDocumentStatusRequest extends Omit<
  CreateDocumentStatusRequest,
  "code"
> {
  is_active: boolean;
}

export interface CreateDocumentTransitionRequest {
  from_status_id: string;
  to_status_id: string;
  required_permission_id?: string;
}

export interface UpdateDocumentTransitionRequest {
  required_permission_id?: string;
  is_active: boolean;
}
