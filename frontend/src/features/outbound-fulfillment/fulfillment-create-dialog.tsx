"use client";

import { useId, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { OperationDialog } from "@/components/ui/operation-dialog";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
  listLocationTypes,
  listLocations,
  storageLayoutKeys,
} from "@/features/storage-layout/storage-layout-api";
import { localDateTime } from "@/features/outbound-orders/outbound-order-schema";
import {
  allPages,
  createPacking,
  createShipment,
  fulfillmentId,
  listCarriers,
  listCarrierServices,
  packingCandidates,
  shipmentCandidates,
  type FulfillmentMode,
} from "./fulfillment-api";

export function FulfillmentCreateDialog({
  mode,
  ownerId,
  warehouseId,
  scopeLabel,
  onCreated,
  onOpenChange,
}: {
  mode: FulfillmentMode;
  ownerId: string;
  warehouseId: string;
  scopeLabel: string;
  onCreated: (id: string) => void;
  onOpenChange: (open: boolean) => void;
}) {
  const id = useId();
  const client = useQueryClient();
  const packing = mode === "packings";
  const [search, setSearch] = useState("");
  const [checkId, setCheckId] = useState("");
  const [locationId, setLocationId] = useState("");
  const [selected, setSelected] = useState<string[]>([]);
  const [businessDate, setBusinessDate] = useState(
    localDateTime().slice(0, 10),
  );
  const [vehicle, setVehicle] = useState("");
  const [seal, setSeal] = useState("");
  const [tracking, setTracking] = useState("");
  const [route, setRoute] = useState("");
  const [notes, setNotes] = useState("");
  const [carrierId, setCarrierId] = useState("");
  const [serviceId, setServiceId] = useState("");
  const filters = {
    ownerId,
    warehouseId,
    search: "",
    status: "",
    page: 1,
    pageSize: 100,
  };
  const checks = useQuery({
    queryKey: ["outbound-fulfillment", "packing-candidates", filters],
    queryFn: () => packingCandidates(filters),
    enabled: packing,
  });
  const packings = useQuery({
    queryKey: ["outbound-fulfillment", "shipment-candidates", filters],
    queryFn: () => shipmentCandidates(filters),
    enabled: !packing,
  });
  const locationFilters = {
    warehouseId,
    search: "",
    active: "active" as const,
    page: 1,
    pageSize: 100,
  };
  const locations = useQuery({
    queryKey: [...storageLayoutKeys.locations(locationFilters), "all-pages"],
    queryFn: () =>
      allPages((page) => listLocations({ ...locationFilters, page })),
    enabled: packing,
  });
  const locationTypes = useQuery({
    queryKey: storageLayoutKeys.locationTypes("active"),
    queryFn: () => listLocationTypes("active"),
    enabled: packing,
  });
  const packingTypeIds = new Set(
    (locationTypes.data ?? [])
      .filter((type) => type.code === "PACKING")
      .map((type) => type.location_type_id),
  );
  const packingLocations = (locations.data ?? []).filter(
    (location) =>
      !location.is_locked && packingTypeIds.has(location.location_type_id),
  );
  const carriers = useQuery({
    queryKey: ["outbound-fulfillment", "carriers"],
    queryFn: listCarriers,
    enabled: !packing,
  });
  const services = useQuery({
    queryKey: ["outbound-fulfillment", "carrier-services", carrierId],
    queryFn: () => listCarrierServices(carrierId),
    enabled: !packing && Boolean(carrierId),
  });
  const mutation = useMutation({
    mutationFn: async () => {
      if (packing) {
        if (!checkId || !locationId)
          throw new Error("Select a passed check and packing location.");
        return createPacking({
          outbound_check_id: checkId,
          packing_location_id: locationId,
        });
      }
      if (!selected.length)
        throw new Error("Select at least one completed packing.");
      if (carrierId && !serviceId)
        throw new Error(
          "Select a carrier service or use self-managed transport.",
        );
      return createShipment({
        owner_id: ownerId,
        warehouse_id: warehouseId,
        business_date: businessDate,
        packing_ids: selected,
        carrier_service_id: serviceId || undefined,
        vehicle_number: vehicle.trim() || undefined,
        seal_number: seal.trim() || undefined,
        tracking_number: tracking.trim() || undefined,
        route_reference: route.trim() || undefined,
        notes: notes.trim() || undefined,
      });
    },
    onSuccess: async (doc) => {
      await client.invalidateQueries({ queryKey: ["outbound-fulfillment"] });
      await client.invalidateQueries({ queryKey: ["outbound-orders"] });
      toast.success(
        packing
          ? "Packing opened."
          : "Shipment planned. No stock has been dispatched yet.",
      );
      onCreated(fulfillmentId(doc));
    },
    onError: (error: Error) => toast.error(error.message),
  });
  const error =
    mutation.error ||
    (packing
      ? checks.error || locations.error || locationTypes.error
      : packings.error || carriers.error || services.error);
  const pending = packing
    ? checks.isPending || locations.isPending || locationTypes.isPending
    : packings.isPending ||
      carriers.isPending ||
      (Boolean(carrierId) && services.isPending);
  const term = search.trim().toLowerCase();
  const visiblePackings = (packings.data ?? []).filter((doc) =>
    `${doc.packing_id} ${doc.client_delivery_order_no}`
      .toLowerCase()
      .includes(term),
  );
  return (
    <OperationDialog
      title={packing ? "Create packing" : "Create shipment"}
      description={scopeLabel}
      busy={mutation.isPending}
      className="max-w-5xl"
      onOpenChange={onOpenChange}
    >
      <form
        className="space-y-6"
        onSubmit={(e) => {
          e.preventDefault();
          if (!mutation.isPending) mutation.mutate();
        }}
      >
        <p className="rounded-xl bg-cyan-50 p-4 text-sm text-cyan-900">
          {packing
            ? "Choose a passed check that still needs packing. Checked quantities are packed in full in their base UOM."
            : "Select completed packings in this scope. Creating the shipment records its manifest only; stock leaves the warehouse when dispatch is confirmed."}
        </p>
        {error && (
          <p
            role="alert"
            className="rounded-xl bg-rose-50 p-4 text-sm text-rose-700"
          >
            {error.message}
          </p>
        )}
        <fieldset
          disabled={mutation.isPending}
          inert={mutation.isPending}
          className="space-y-5"
        >
          <FormField
            label={
              packing ? "Search passed checks" : "Search completed packings"
            }
            htmlFor={`${id}-search`}
          >
            <Input
              id={`${id}-search`}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Document or client DO"
            />
          </FormField>
          {packing ? (
            <div className="grid gap-5 sm:grid-cols-2">
              <FormField label="Passed check" htmlFor={`${id}-check`} required>
                <Select
                  className="mt-2"
                  id={`${id}-check`}
                  ariaLabel="Passed check"
                  value={checkId}
                  onValueChange={setCheckId}
                  options={(checks.data ?? [])
                    .filter((c) =>
                      `${c.outbound_check_id} ${c.client_delivery_order_no}`
                        .toLowerCase()
                        .includes(term),
                    )
                    .map((c) => ({
                      value: c.outbound_check_id,
                      label: `${c.outbound_check_id} · ${c.client_delivery_order_no}`,
                    }))}
                  placeholder={
                    checks.isPending
                      ? "Loading eligible checks…"
                      : "Select passed check"
                  }
                />
              </FormField>
              <FormField
                label="Packing location"
                htmlFor={`${id}-location`}
                required
              >
                <Select
                  className="mt-2"
                  id={`${id}-location`}
                  ariaLabel="Packing location"
                  value={locationId}
                  onValueChange={setLocationId}
                  options={packingLocations.map((l) => ({
                    value: l.location_id,
                    label: `${l.code} · ${l.zone_code ?? ""}`,
                  }))}
                  placeholder="Active, unlocked packing area"
                />
              </FormField>
              {!locations.isPending &&
                !locationTypes.isPending &&
                !packingLocations.length && (
                  <p className="text-sm text-amber-800 sm:col-span-2">
                    Create or unlock a location of the active PACKING type in
                    this warehouse’s storage layout before starting packing.
                  </p>
                )}
              {!checks.isPending && !checks.data?.length && (
                <p className="text-sm text-amber-800 sm:col-span-2">
                  No passed checks are waiting for packing in this scope.
                </p>
              )}
            </div>
          ) : (
            <>
              <section className="rounded-xl border border-slate-200 p-4">
                <h3 className="font-bold">
                  Completed packings{" "}
                  <span className="text-sm font-normal text-slate-500">
                    ({selected.length} selected)
                  </span>
                </h3>
                <div className="mt-3 max-h-64 space-y-3 overflow-y-auto">
                  {visiblePackings.map((doc) => (
                    <label
                      key={doc.packing_id}
                      className="flex items-start gap-3 text-sm"
                    >
                      <input
                        type="checkbox"
                        className="mt-1 size-4"
                        checked={selected.includes(doc.packing_id)}
                        onChange={(e) =>
                          setSelected(
                            e.target.checked
                              ? [...selected, doc.packing_id]
                              : selected.filter(
                                  (value) => value !== doc.packing_id,
                                ),
                          )
                        }
                      />
                      <span>
                        <span className="font-semibold">{doc.packing_id}</span>
                        <span className="block text-slate-500">
                          {doc.client_delivery_order_no} · {doc.line_count}{" "}
                          packed lines · {doc.packing_location_code}
                        </span>
                      </span>
                    </label>
                  ))}
                  {!packings.isPending && !visiblePackings.length && (
                    <p className="text-sm text-slate-500">
                      No unassigned completed packings match.
                    </p>
                  )}
                </div>
              </section>
              <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
                <FormField
                  label="Business date"
                  htmlFor={`${id}-date`}
                  required
                >
                  <Input
                    id={`${id}-date`}
                    type="date"
                    required
                    value={businessDate}
                    onChange={(e) => setBusinessDate(e.target.value)}
                  />
                </FormField>
                <FormField label="Vehicle number" htmlFor={`${id}-vehicle`}>
                  <Input
                    id={`${id}-vehicle`}
                    maxLength={60}
                    value={vehicle}
                    onChange={(e) => setVehicle(e.target.value)}
                  />
                </FormField>
                <FormField label="Seal number" htmlFor={`${id}-seal`}>
                  <Input
                    id={`${id}-seal`}
                    maxLength={60}
                    value={seal}
                    onChange={(e) => setSeal(e.target.value)}
                  />
                </FormField>
                <FormField label="Carrier" htmlFor={`${id}-carrier`}>
                  <Select
                    className="mt-2"
                    id={`${id}-carrier`}
                    ariaLabel="Carrier"
                    value={carrierId || "none"}
                    options={[
                      { value: "none", label: "Self-managed transport" },
                      ...(carriers.data ?? []).map((c) => ({
                        value: c.carrier_id,
                        label: `${c.name} (${c.code})`,
                      })),
                    ]}
                    onValueChange={(next) => {
                      setCarrierId(next === "none" ? "" : next);
                      setServiceId("");
                    }}
                  />
                </FormField>
                <FormField
                  label="Carrier service"
                  htmlFor={`${id}-service`}
                  required={Boolean(carrierId)}
                >
                  <Select
                    className="mt-2"
                    id={`${id}-service`}
                    ariaLabel="Carrier service"
                    value={serviceId}
                    disabled={!carrierId}
                    options={(services.data ?? []).map((s) => ({
                      value: s.carrier_service_id,
                      label: `${s.name} (${s.code})`,
                    }))}
                    placeholder={
                      carrierId ? "Select service" : "No carrier service"
                    }
                    onValueChange={setServiceId}
                  />
                </FormField>
                <FormField label="Tracking number" htmlFor={`${id}-tracking`}>
                  <Input
                    id={`${id}-tracking`}
                    maxLength={150}
                    value={tracking}
                    onChange={(e) => setTracking(e.target.value)}
                  />
                </FormField>
                <FormField label="Route reference" htmlFor={`${id}-route`}>
                  <Input
                    id={`${id}-route`}
                    maxLength={100}
                    value={route}
                    onChange={(e) => setRoute(e.target.value)}
                  />
                </FormField>
                <div className="sm:col-span-2">
                  <FormField label="Notes" htmlFor={`${id}-notes`}>
                    <Textarea
                      id={`${id}-notes`}
                      rows={3}
                      className="min-h-0"
                      maxLength={4000}
                      value={notes}
                      onChange={(e) => setNotes(e.target.value)}
                    />
                  </FormField>
                </div>
              </div>
            </>
          )}
        </fieldset>
        <div className="flex justify-end gap-2 border-t border-slate-200 pt-5">
          <Button
            type="button"
            variant="secondary"
            disabled={mutation.isPending}
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            disabled={
              mutation.isPending || pending || Boolean(error && !mutation.error)
            }
          >
            {mutation.isPending
              ? "Saving…"
              : packing
                ? "Open packing"
                : "Create manifest"}
          </Button>
        </div>
      </form>
    </OperationDialog>
  );
}
