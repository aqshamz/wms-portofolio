import { apiRequest } from "@/lib/api/client";
import type {
  AppModulePage,
  CreateDocumentStatusRequest,
  CreateDocumentTransitionRequest,
  CreateDocumentTypeRequest,
  DocumentStatus,
  DocumentStatusPage,
  DocumentTransition,
  DocumentTransitionPage,
  DocumentType,
  DocumentTypePage,
  UpdateDocumentStatusRequest,
  UpdateDocumentTransitionRequest,
  UpdateDocumentTypeRequest,
  WorkflowListFilters,
  WorkflowPermissionPage,
} from "@/features/document-workflows/document-workflow-types";

const typesPath = "/api/v1/master/document-types";

export const documentWorkflowKeys = {
  all: ["document-workflows"] as const,
  typeLists: () => [...documentWorkflowKeys.all, "types", "list"] as const,
  typeList: (filters: WorkflowListFilters) =>
    [...documentWorkflowKeys.typeLists(), filters] as const,
  type: (id: string) =>
    [...documentWorkflowKeys.all, "types", "detail", id] as const,
  statuses: (typeId: string, filters: WorkflowListFilters) =>
    [...documentWorkflowKeys.all, typeId, "statuses", filters] as const,
  status: (typeId: string, id: string) =>
    [...documentWorkflowKeys.all, typeId, "statuses", "detail", id] as const,
  transitions: (typeId: string, filters: WorkflowListFilters) =>
    [...documentWorkflowKeys.all, typeId, "transitions", filters] as const,
  transition: (typeId: string, id: string) =>
    [...documentWorkflowKeys.all, typeId, "transitions", "detail", id] as const,
  modules: () => [...documentWorkflowKeys.all, "modules"] as const,
  permissions: () => [...documentWorkflowKeys.all, "permissions"] as const,
};

function listPath(base: string, filters: WorkflowListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page ?? 1),
    page_size: String(filters.pageSize ?? 100),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all") {
    query.set("active", String(filters.active === "active"));
  }
  return `${base}?${query}`;
}

export function documentTypeListPath(filters: WorkflowListFilters) {
  return listPath(typesPath, filters);
}

export function listDocumentTypes(filters: WorkflowListFilters) {
  return apiRequest<DocumentTypePage>(documentTypeListPath(filters));
}

export function getDocumentType(id: string) {
  return apiRequest<DocumentType>(`${typesPath}/${id}`);
}

export function createDocumentType(request: CreateDocumentTypeRequest) {
  return apiRequest<DocumentType>(typesPath, { method: "POST", body: request });
}

export function updateDocumentType(
  id: string,
  request: UpdateDocumentTypeRequest,
) {
  return apiRequest<DocumentType>(`${typesPath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateDocumentType(id: string) {
  return apiRequest<DocumentType>(`${typesPath}/${id}/deactivate`, {
    method: "PATCH",
  });
}

function statusesPath(typeId: string) {
  return `${typesPath}/${typeId}/statuses`;
}

export function listDocumentStatuses(
  typeId: string,
  filters: WorkflowListFilters,
) {
  return apiRequest<DocumentStatusPage>(
    listPath(statusesPath(typeId), filters),
  );
}

export function getDocumentStatus(typeId: string, id: string) {
  return apiRequest<DocumentStatus>(`${statusesPath(typeId)}/${id}`);
}

export function createDocumentStatus(
  typeId: string,
  request: CreateDocumentStatusRequest,
) {
  return apiRequest<DocumentStatus>(statusesPath(typeId), {
    method: "POST",
    body: request,
  });
}

export function updateDocumentStatus(
  typeId: string,
  id: string,
  request: UpdateDocumentStatusRequest,
) {
  return apiRequest<DocumentStatus>(`${statusesPath(typeId)}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateDocumentStatus(typeId: string, id: string) {
  return apiRequest<DocumentStatus>(
    `${statusesPath(typeId)}/${id}/deactivate`,
    {
      method: "PATCH",
    },
  );
}

function transitionsPath(typeId: string) {
  return `${typesPath}/${typeId}/transitions`;
}

export function listDocumentTransitions(
  typeId: string,
  filters: WorkflowListFilters,
) {
  return apiRequest<DocumentTransitionPage>(
    listPath(transitionsPath(typeId), filters),
  );
}

export function getDocumentTransition(typeId: string, id: string) {
  return apiRequest<DocumentTransition>(`${transitionsPath(typeId)}/${id}`);
}

export function createDocumentTransition(
  typeId: string,
  request: CreateDocumentTransitionRequest,
) {
  return apiRequest<DocumentTransition>(transitionsPath(typeId), {
    method: "POST",
    body: request,
  });
}

export function updateDocumentTransition(
  typeId: string,
  id: string,
  request: UpdateDocumentTransitionRequest,
) {
  return apiRequest<DocumentTransition>(`${transitionsPath(typeId)}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateDocumentTransition(typeId: string, id: string) {
  return apiRequest<DocumentTransition>(
    `${transitionsPath(typeId)}/${id}/deactivate`,
    { method: "PATCH" },
  );
}

export function listWorkflowModules() {
  return apiRequest<AppModulePage>(
    "/api/v1/master/modules?page=1&page_size=100",
  );
}

export function listWorkflowPermissions() {
  return apiRequest<WorkflowPermissionPage>(
    "/api/v1/master/workflow-permissions?page=1&page_size=100",
  );
}
