import { apiRequest } from "@/lib/api/client";
import type {
  ActiveFilter,
  CreateLocationRequest,
  CreateLocationTypeRequest,
  CreateZoneRequest,
  LocationListFilters,
  LocationType,
  UpdateLocationRequest,
  UpdateLocationTypeRequest,
  UpdateZoneRequest,
  WarehouseLocation,
  WarehouseLocationPage,
  WarehouseZone,
} from "@/features/storage-layout/storage-layout-types";

export const storageLayoutKeys = {
  all: ["storage-layout"] as const,
  locationTypes: (active: ActiveFilter) =>
    [...storageLayoutKeys.all, "location-types", active] as const,
  zones: (warehouseId: string, active: ActiveFilter, search = "") =>
    [...storageLayoutKeys.all, "zones", warehouseId, active, search] as const,
  locations: (filters: LocationListFilters) =>
    [...storageLayoutKeys.all, "locations", filters] as const,
  location: (id: string) => [...storageLayoutKeys.all, "location", id] as const,
};

function activeQuery(active: ActiveFilter) {
  return active === "all" ? "" : `?active=${active === "active"}`;
}

export function listLocationTypes(active: ActiveFilter) {
  return apiRequest<LocationType[]>(
    `/api/v1/master/location-types${activeQuery(active)}`,
  );
}

export function createLocationType(request: CreateLocationTypeRequest) {
  return apiRequest<LocationType>("/api/v1/master/location-types", {
    method: "POST",
    body: request,
  });
}

export function updateLocationType(
  id: string,
  request: UpdateLocationTypeRequest,
) {
  return apiRequest<LocationType>(`/api/v1/master/location-types/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function listZones(
  warehouseId: string,
  active: ActiveFilter,
  search = "",
) {
  const query = new URLSearchParams();
  if (active !== "all") query.set("active", String(active === "active"));
  if (search.trim()) query.set("search", search.trim());
  const suffix = query.size ? `?${query}` : "";
  return apiRequest<WarehouseZone[]>(
    `/api/v1/master/warehouses/${warehouseId}/zones${suffix}`,
  );
}

export function createZone(warehouseId: string, request: CreateZoneRequest) {
  return apiRequest<WarehouseZone>(
    `/api/v1/master/warehouses/${warehouseId}/zones`,
    { method: "POST", body: request },
  );
}

export function updateZone(id: string, request: UpdateZoneRequest) {
  return apiRequest<WarehouseZone>(`/api/v1/master/zones/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateZone(id: string) {
  return apiRequest<void>(`/api/v1/master/zones/${id}/deactivate`, {
    method: "PATCH",
  });
}

export function locationListPath(filters: LocationListFilters) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.zoneId) query.set("zone_id", filters.zoneId);
  if (filters.locationTypeId) {
    query.set("location_type_id", filters.locationTypeId);
  }
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all") {
    query.set("active", String(filters.active === "active"));
  }
  return `/api/v1/master/warehouses/${filters.warehouseId}/locations?${query}`;
}

export function listLocations(filters: LocationListFilters) {
  return apiRequest<WarehouseLocationPage>(locationListPath(filters));
}

export function getLocation(id: string) {
  return apiRequest<WarehouseLocation>(`/api/v1/master/locations/${id}`);
}

export function createLocation(
  warehouseId: string,
  request: CreateLocationRequest,
) {
  return apiRequest<WarehouseLocation>(
    `/api/v1/master/warehouses/${warehouseId}/locations`,
    { method: "POST", body: request },
  );
}

export function updateLocation(id: string, request: UpdateLocationRequest) {
  return apiRequest<WarehouseLocation>(`/api/v1/master/locations/${id}`, {
    method: "PUT",
    body: request,
  });
}

export function deactivateLocation(id: string) {
  return apiRequest<void>(`/api/v1/master/locations/${id}/deactivate`, {
    method: "PATCH",
  });
}
