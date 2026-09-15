import { apiRequest } from "@/lib/api/client";
import type {
  CreateInspectionResultRequest,
  CreateQualityStatusRequest,
  InspectionResult,
  InspectionResultPage,
  QualityListFilters,
  QualityStatus,
  QualityStatusPage,
  UpdateInspectionResultRequest,
  UpdateQualityStatusRequest,
} from "@/features/quality-setup/quality-setup-types";

const statusPath = "/api/v1/master/quality-statuses";
const resultPath = "/api/v1/master/inspection-results";

export const qualitySetupKeys = {
  all: ["quality-setup"] as const,
  statusLists: () => [...qualitySetupKeys.all, "statuses", "list"] as const,
  statusList: (filters: QualityListFilters) =>
    [...qualitySetupKeys.statusLists(), filters] as const,
  status: (id: string) =>
    [...qualitySetupKeys.all, "statuses", "detail", id] as const,
  resultLists: () => [...qualitySetupKeys.all, "results", "list"] as const,
  resultList: (filters: QualityListFilters) =>
    [...qualitySetupKeys.resultLists(), filters] as const,
  result: (id: string) =>
    [...qualitySetupKeys.all, "results", "detail", id] as const,
};

function listPath(base: string, filters: QualityListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all") {
    query.set("active", String(filters.active === "active"));
  }
  return `${base}?${query}`;
}

export function qualityStatusListPath(filters: QualityListFilters) {
  return listPath(statusPath, filters);
}

export function inspectionResultListPath(filters: QualityListFilters) {
  return listPath(resultPath, filters);
}

export function listQualityStatuses(filters: QualityListFilters) {
  return apiRequest<QualityStatusPage>(qualityStatusListPath(filters));
}

export function getQualityStatus(id: string) {
  return apiRequest<QualityStatus>(`${statusPath}/${id}`);
}

export function createQualityStatus(request: CreateQualityStatusRequest) {
  return apiRequest<QualityStatus>(statusPath, {
    method: "POST",
    body: request,
  });
}

export function updateQualityStatus(
  id: string,
  request: UpdateQualityStatusRequest,
) {
  return apiRequest<QualityStatus>(`${statusPath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateQualityStatus(id: string) {
  return apiRequest<QualityStatus>(`${statusPath}/${id}/deactivate`, {
    method: "PATCH",
  });
}

export function listInspectionResults(filters: QualityListFilters) {
  return apiRequest<InspectionResultPage>(inspectionResultListPath(filters));
}

export function getInspectionResult(id: string) {
  return apiRequest<InspectionResult>(`${resultPath}/${id}`);
}

export function createInspectionResult(request: CreateInspectionResultRequest) {
  return apiRequest<InspectionResult>(resultPath, {
    method: "POST",
    body: request,
  });
}

export function updateInspectionResult(
  id: string,
  request: UpdateInspectionResultRequest,
) {
  return apiRequest<InspectionResult>(`${resultPath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateInspectionResult(id: string) {
  return apiRequest<InspectionResult>(`${resultPath}/${id}/deactivate`, {
    method: "PATCH",
  });
}
