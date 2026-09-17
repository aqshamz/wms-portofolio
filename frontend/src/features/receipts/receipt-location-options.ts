import type {
  LocationType,
  WarehouseLocation,
} from "@/features/storage-layout/storage-layout-types";

export function receiptLocationOptions(
  locations: WarehouseLocation[],
  types: LocationType[],
) {
  const typesById = new Map(types.map((type) => [type.location_type_id, type]));
  const receiving = locations.filter((location) => {
    const type = typesById.get(location.location_type_id);
    return (
      location.is_active &&
      !location.is_locked &&
      type?.is_active &&
      type.allows_receiving
    );
  });
  return {
    receiving,
    docks: receiving.filter(
      (location) => typesById.get(location.location_type_id)?.code === "DOCK",
    ),
  };
}
