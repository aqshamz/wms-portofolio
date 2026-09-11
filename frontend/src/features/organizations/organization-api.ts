import { apiRequest } from "@/lib/api/client";
import type {
  CreateOrganizationRequest,
  Organization,
  OrganizationListFilters,
  OrganizationPage,
  UpdateOrganizationRequest,
} from "@/features/organizations/organization-types";

const basePath = "/api/v1/master/organizations";

export const organizationKeys = {
  all: ["organizations"] as const,
  lists: () => [...organizationKeys.all, "list"] as const,
  list: (filters: OrganizationListFilters) =>
    [...organizationKeys.lists(), filters] as const,
  details: () => [...organizationKeys.all, "detail"] as const,
  detail: (id: string) => [...organizationKeys.details(), id] as const,
};

export function organizationListPath(filters: OrganizationListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all") {
    query.set("active", String(filters.active === "active"));
  }
  return `${basePath}?${query}`;
}

export function listOrganizations(filters: OrganizationListFilters) {
  return apiRequest<OrganizationPage>(organizationListPath(filters));
}

export function getOrganization(id: string) {
  return apiRequest<Organization>(`${basePath}/${id}`);
}

export function createOrganization(request: CreateOrganizationRequest) {
  return apiRequest<Organization>(basePath, {
    method: "POST",
    body: request,
  });
}

export function updateOrganization(
  id: string,
  request: UpdateOrganizationRequest,
) {
  return apiRequest<Organization>(`${basePath}/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateOrganization(id: string, expectedUpdatedAt: string) {
  return apiRequest<Organization>(`${basePath}/${id}/deactivate`, {
    method: "PATCH",
    body: { expected_updated_at: expectedUpdatedAt },
  });
}
