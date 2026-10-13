import { apiRequest } from "@/lib/api/client";
import type { PaginatedData } from "@/lib/api/types";
import type { OutboundOrderFilters } from "@/features/outbound-orders/outbound-order-types";

export type FulfillmentMode = "packings" | "shipments";
export interface PackingLine {
  serial_no?: string;
  packing_line_id: string;
  outbound_check_line_id: string;
  item_code: string;
  lot_number?: string;
  source_balance_id: string;
  source_balance_version: number;
  handling_unit_id: string | null;
  handling_unit_barcode?: string;
  expected_qty: string;
  packed_qty: string;
  uom_code: string;
  movement_id?: string;
}
export interface Packing {
  packing_id: string;
  outbound_id: string;
  client_delivery_order_no: string;
  owner_id: string;
  warehouse_id: string;
  packing_location_id: string;
  packing_location_code: string;
  status_code: string;
  packed_at?: string;
  line_count: number;
  lines?: PackingLine[];
}
export interface ShipmentLine {
  serial_no?: string;
  shipment_line_id?: string;
  packing_line_id: string;
  outbound_id: string;
  item_code: string;
  lot_number?: string;
  source_balance_id: string;
  source_balance_version: number;
  handling_unit_id: string | null;
  handling_unit_barcode?: string;
  shipped_qty: string;
  uom_code: string;
  movement_id?: string;
}
export interface ShipmentDriver {
  driver_id: string;
  driver_code: string;
  driver_name: string;
  is_primary: boolean;
}
export interface Shipment {
  shipment_id: string;
  owner_id: string;
  warehouse_id: string;
  business_date: string;
  status_code: string;
  carrier_code?: string;
  carrier_service_id?: string;
  carrier_service_code?: string;
  vehicle_number?: string;
  seal_number?: string;
  tracking_number?: string;
  route_reference?: string;
  notes?: string;
  shipped_at?: string;
  order_count: number;
  line_count: number;
  lines?: ShipmentLine[];
  drivers?: ShipmentDriver[];
}
export interface PackingCheck {
  outbound_check_id: string;
  outbound_id: string;
  client_delivery_order_no: string;
  line_count: number;
}
export interface Carrier {
  carrier_id: string;
  code: string;
  name: string;
  is_active: boolean;
}
export interface CarrierDriver {
  driver_id: string;
  code: string;
  name: string;
  is_active: boolean;
}
export interface CarrierService {
  carrier_service_id: string;
  code: string;
  name: string;
  is_active: boolean;
}
export type FulfillmentDocument = Packing | Shipment;
export const fulfillmentId = (doc: FulfillmentDocument) =>
  "packing_id" in doc ? doc.packing_id : doc.shipment_id;
const root = "/api/v1/outbound";
const path = (mode: FulfillmentMode, id: string) =>
  `${root}/${mode}/${encodeURIComponent(id)}`;
export const fulfillmentKeys = {
  all: ["outbound-fulfillment"] as const,
  list: (mode: FulfillmentMode, filters: OutboundOrderFilters) =>
    ["outbound-fulfillment", mode, filters] as const,
  detail: (mode: FulfillmentMode, id: string) =>
    ["outbound-fulfillment", mode, "detail", id] as const,
};
export function fulfillmentListPath(
  resource: string,
  filters: OutboundOrderFilters,
) {
  const query = new URLSearchParams({
    owner_id: filters.ownerId,
    warehouse_id: filters.warehouseId,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.status) query.set("status_code", filters.status);
  if (filters.search.trim()) query.set("search", filters.search.trim());
  return `${root}/${resource}?${query}`;
}
export const listFulfillment = (
  mode: FulfillmentMode,
  filters: OutboundOrderFilters,
) =>
  apiRequest<PaginatedData<FulfillmentDocument>>(
    fulfillmentListPath(mode, filters),
  );
export const getFulfillment = (mode: FulfillmentMode, id: string) =>
  apiRequest<FulfillmentDocument>(path(mode, id));
export const createPacking = (body: {
  outbound_check_id: string;
  packing_location_id: string;
}) => apiRequest<Packing>(`${root}/packings`, { method: "POST", body });
export const createShipment = (body: {
  owner_id: string;
  warehouse_id: string;
  business_date: string;
  packing_ids: string[];
  carrier_service_id?: string;
  vehicle_number?: string;
  seal_number?: string;
  route_reference?: string;
  tracking_number?: string;
  notes?: string;
}) => apiRequest<Shipment>(`${root}/shipments`, { method: "POST", body });
export const packLine = (
  id: string,
  lineId: string,
  body: {
    expected_balance_version: number;
    handling_unit_id?: string;
    business_date: string;
    operation_key: string;
  },
) =>
  apiRequest<Packing>(
    `${path("packings", id)}/lines/${encodeURIComponent(lineId)}`,
    { method: "POST", body },
  );
export const dispatchLine = (
  id: string,
  lineId: string,
  body: {
    expected_balance_version: number;
    business_date: string;
    operation_key: string;
  },
) =>
  apiRequest<Shipment>(
    `${path("shipments", id)}/lines/${encodeURIComponent(lineId)}/dispatch`,
    { method: "POST", body },
  );
export const dispatchHandlingUnit = (
  id: string,
  huId: string,
  body: {
    business_date: string;
    operation_key: string;
    expected_balances: { balance_id: string; expected_version: number }[];
  },
) =>
  apiRequest<Shipment>(
    `${path("shipments", id)}/handling-units/${encodeURIComponent(huId)}/dispatch`,
    { method: "POST", body },
  );
export const transitionFulfillment = (
  mode: FulfillmentMode,
  id: string,
  action: "complete" | "cancel",
) =>
  apiRequest<FulfillmentDocument>(`${path(mode, id)}/${action}`, {
    method: "POST",
  });
export const assignDriver = (id: string, driverId: string) =>
  apiRequest<ShipmentDriver[]>(`${path("shipments", id)}/drivers`, {
    method: "POST",
    body: { driver_id: driverId, is_primary: true },
  });
export const listCarriers = () =>
  apiRequest<Carrier[]>(`${root}/carriers?active=true`);
export const listDrivers = (carrier: string) =>
  apiRequest<CarrierDriver[]>(
    `${root}/carriers/${encodeURIComponent(carrier)}/drivers?active=true`,
  );
export const listCarrierServices = (carrier: string) =>
  apiRequest<CarrierService[]>(
    `${root}/carriers/${encodeURIComponent(carrier)}/services?active=true`,
  );
export async function allPages<T>(
  load: (page: number) => Promise<PaginatedData<T>>,
) {
  const first = await load(1);
  const items = [...first.items];
  for (let page = 2; page <= first.total_pages; page++)
    items.push(...(await load(page)).items);
  return items;
}
export const packingCandidates = (filters: OutboundOrderFilters) =>
  allPages((page) =>
    apiRequest<PaginatedData<PackingCheck>>(
      fulfillmentListPath("packing-candidates", {
        ...filters,
        status: "",
        page,
        pageSize: 100,
      }),
    ),
  );
export const shipmentCandidates = (filters: OutboundOrderFilters) =>
  allPages((page) =>
    apiRequest<PaginatedData<Packing>>(
      fulfillmentListPath("shipment-candidates", {
        ...filters,
        status: "",
        page,
        pageSize: 100,
      }),
    ),
  );
