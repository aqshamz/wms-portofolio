import { apiRequest } from "@/lib/api/client";
import type {
  BusinessPartner,
  BusinessPartnerDetail,
  BusinessPartnerListFilters,
  BusinessPartnerPage,
  CreateBusinessPartnerRequest,
  CreatePartnerTypeRequest,
  PartnerType,
  PartnerTypeListFilters,
  PartnerTypePage,
  UpdateBusinessPartnerRequest,
  UpdatePartnerTypeRequest,
} from "@/features/business-partners/business-partner-types";

const partnerBasePath = "/api/v1/master/business-partners";
const partnerTypeBasePath = "/api/v1/master/partner-types";

export const businessPartnerKeys = {
  all: ["business-partners"] as const,
  partnerLists: () => [...businessPartnerKeys.all, "partners", "list"] as const,
  partnerList: (filters: BusinessPartnerListFilters) =>
    [...businessPartnerKeys.partnerLists(), filters] as const,
  partner: (partnerId: string) =>
    [...businessPartnerKeys.all, "partners", "detail", partnerId] as const,
  typeLists: () => [...businessPartnerKeys.all, "types", "list"] as const,
  typeList: (filters: PartnerTypeListFilters) =>
    [...businessPartnerKeys.typeLists(), filters] as const,
  type: (partnerTypeId: string) =>
    [...businessPartnerKeys.all, "types", "detail", partnerTypeId] as const,
};

function addActiveFilter(query: URLSearchParams, active: string) {
  if (active !== "all") query.set("active", String(active === "active"));
}

export function businessPartnerListPath(filters: BusinessPartnerListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.ownerId) query.set("owner_id", filters.ownerId);
  if (filters.partnerTypeCode) {
    query.set("partner_type_code", filters.partnerTypeCode);
  }
  if (filters.search.trim()) query.set("search", filters.search.trim());
  addActiveFilter(query, filters.active);
  return `${partnerBasePath}?${query}`;
}

export function partnerTypeListPath(filters: PartnerTypeListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  addActiveFilter(query, filters.active);
  return `${partnerTypeBasePath}?${query}`;
}

export function listBusinessPartners(filters: BusinessPartnerListFilters) {
  return apiRequest<BusinessPartnerPage>(businessPartnerListPath(filters));
}

export function getBusinessPartner(partnerId: string) {
  return apiRequest<BusinessPartnerDetail>(`${partnerBasePath}/${partnerId}`);
}

export function createBusinessPartner(request: CreateBusinessPartnerRequest) {
  return apiRequest<BusinessPartner>(partnerBasePath, {
    method: "POST",
    body: request,
  });
}

export function updateBusinessPartner(
  partnerId: string,
  request: UpdateBusinessPartnerRequest,
) {
  return apiRequest<BusinessPartner>(`${partnerBasePath}/${partnerId}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateBusinessPartner(
  partnerId: string,
  expectedUpdatedAt: string,
) {
  return apiRequest<BusinessPartner>(
    `${partnerBasePath}/${partnerId}/deactivate`,
    { method: "PATCH", body: { expected_updated_at: expectedUpdatedAt } },
  );
}

export function assignBusinessPartnerType(
  partnerId: string,
  partnerTypeId: string,
) {
  return apiRequest<PartnerType[]>(`${partnerBasePath}/${partnerId}/types`, {
    method: "POST",
    body: { partner_type_id: partnerTypeId },
  });
}

export function removeBusinessPartnerType(
  partnerId: string,
  partnerTypeId: string,
) {
  return apiRequest<void>(
    `${partnerBasePath}/${partnerId}/types/${partnerTypeId}`,
    { method: "DELETE" },
  );
}

export function listPartnerTypes(filters: PartnerTypeListFilters) {
  return apiRequest<PartnerTypePage>(partnerTypeListPath(filters));
}

export function getPartnerType(partnerTypeId: string) {
  return apiRequest<PartnerType>(`${partnerTypeBasePath}/${partnerTypeId}`);
}

export function createPartnerType(request: CreatePartnerTypeRequest) {
  return apiRequest<PartnerType>(partnerTypeBasePath, {
    method: "POST",
    body: request,
  });
}

export function updatePartnerType(
  partnerTypeId: string,
  request: UpdatePartnerTypeRequest,
) {
  return apiRequest<PartnerType>(`${partnerTypeBasePath}/${partnerTypeId}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivatePartnerType(partnerTypeId: string) {
  return apiRequest<PartnerType>(
    `${partnerTypeBasePath}/${partnerTypeId}/deactivate`,
    { method: "PATCH" },
  );
}
