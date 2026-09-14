"use client";

import { useEffect, type ReactNode } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { LoaderCircle, X } from "lucide-react";
import {
  Controller,
  useForm,
  type UseFormRegisterReturn,
} from "react-hook-form";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  createLocation,
  createLocationType,
  createZone,
  storageLayoutKeys,
  updateLocation,
  updateLocationType,
  updateZone,
} from "@/features/storage-layout/storage-layout-api";
import {
  emptyLocationForm,
  emptyLocationTypeForm,
  emptyZoneForm,
  locationFormSchema,
  locationTypeFormSchema,
  zoneFormSchema,
  type LocationFormValues,
  type LocationTypeFormValues,
  type ZoneFormValues,
} from "@/features/storage-layout/storage-layout-schema";
import type {
  CreateLocationRequest,
  CreateLocationTypeRequest,
  CreateZoneRequest,
  LocationType,
  UpdateLocationRequest,
  UpdateLocationTypeRequest,
  UpdateZoneRequest,
  WarehouseLocation,
  WarehouseZone,
} from "@/features/storage-layout/storage-layout-types";

const textareaClassName =
  "mt-2 min-h-24 w-full resize-y rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-sm text-slate-950 outline-none transition placeholder:text-slate-400 focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100";

function optional(value: string) {
  const trimmed = value.trim();
  return trimmed || undefined;
}

function errorId(id: string, error?: string) {
  return error ? `${id}-error` : undefined;
}

function Toggle({
  label,
  registration,
}: {
  label: string;
  registration: UseFormRegisterReturn;
}) {
  return (
    <label className="flex min-h-11 items-center gap-3 rounded-xl border border-slate-200 px-3 text-sm font-semibold text-slate-800">
      <input
        type="checkbox"
        className="size-4 accent-cyan-600"
        {...registration}
      />
      {label}
    </label>
  );
}

function DialogFrame({
  open,
  busy,
  title,
  description,
  onOpenChange,
  children,
}: {
  open: boolean;
  busy: boolean;
  title: string;
  description: string;
  onOpenChange: (open: boolean) => void;
  children: ReactNode;
}) {
  return (
    <Dialog.Root
      open={open}
      onOpenChange={(nextOpen) => {
        if (!busy) onOpenChange(nextOpen);
      }}
    >
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[92vh] w-[calc(100%-2rem)] max-w-2xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-2xl bg-white shadow-2xl focus:outline-none">
          <div className="sticky top-0 z-10 flex items-start justify-between gap-4 border-b border-slate-200 bg-white/95 px-5 py-4 backdrop-blur sm:px-6">
            <div>
              <Dialog.Title className="text-lg font-bold text-slate-950">
                {title}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-slate-600">
                {description}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                className="grid size-10 shrink-0 place-items-center rounded-lg text-slate-500 hover:bg-slate-100 hover:text-slate-950"
                aria-label="Close form"
              >
                <X className="size-5" />
              </button>
            </Dialog.Close>
          </div>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

function FormActions({
  busy,
  editing,
  noun,
}: {
  busy: boolean;
  editing: boolean;
  noun: string;
}) {
  return (
    <div className="mt-7 flex flex-col-reverse gap-2 border-t border-slate-200 pt-5 sm:flex-row sm:justify-end">
      <Dialog.Close asChild>
        <Button type="button" variant="secondary">
          Cancel
        </Button>
      </Dialog.Close>
      <Button type="submit" disabled={busy}>
        {busy ? <LoaderCircle className="size-4 animate-spin" /> : null}
        {busy ? "Saving…" : editing ? "Save changes" : `Create ${noun}`}
      </Button>
    </div>
  );
}

export function LocationTypeDialog({
  open,
  locationType,
  onOpenChange,
}: {
  open: boolean;
  locationType?: LocationType;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = locationType !== undefined;
  const queryClient = useQueryClient();
  const {
    register,
    reset,
    handleSubmit,
    formState: { errors },
  } = useForm<LocationTypeFormValues>({
    resolver: zodResolver(locationTypeFormSchema),
    defaultValues: emptyLocationTypeForm,
  });
  const save = useMutation({
    mutationFn: (values: LocationTypeFormValues) => {
      const common = {
        name: values.name.trim(),
        description: optional(values.description),
        allows_receiving: values.allows_receiving,
        allows_storage: values.allows_storage,
        allows_picking: values.allows_picking,
        allows_shipping: values.allows_shipping,
      };
      if (locationType) {
        const request: UpdateLocationTypeRequest = {
          ...common,
          is_active: values.is_active,
        };
        return updateLocationType(locationType.location_type_id, request);
      }
      const request: CreateLocationTypeRequest = {
        code: values.code.trim().toUpperCase(),
        ...common,
      };
      return createLocationType(request);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: storageLayoutKeys.all });
      toast.success(
        editing ? "Location type updated." : "Location type created.",
      );
      onOpenChange(false);
    },
  });

  useEffect(() => {
    if (!open) return;
    reset(
      locationType
        ? {
            code: locationType.code,
            name: locationType.name,
            description: locationType.description ?? "",
            allows_receiving: locationType.allows_receiving,
            allows_storage: locationType.allows_storage,
            allows_picking: locationType.allows_picking,
            allows_shipping: locationType.allows_shipping,
            is_active: locationType.is_active,
          }
        : emptyLocationTypeForm,
    );
  }, [locationType, open, reset]);

  return (
    <DialogFrame
      open={open}
      busy={save.isPending}
      title={editing ? "Edit location type" : "Create location type"}
      description="Define which warehouse activities a physical location supports."
      onOpenChange={onOpenChange}
    >
      <form
        className="p-5 sm:p-6"
        noValidate
        onSubmit={handleSubmit((values) => save.mutate(values))}
      >
        {save.error ? (
          <p
            role="alert"
            className="mb-5 rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
          >
            {save.error.message}
          </p>
        ) : null}
        <div className="grid gap-5 sm:grid-cols-2">
          <FormField
            label="Code"
            htmlFor="type-code"
            required
            error={errors.code?.message}
          >
            <Input
              id="type-code"
              readOnly={editing}
              invalid={Boolean(errors.code)}
              aria-describedby={errorId("type-code", errors.code?.message)}
              {...register("code")}
            />
          </FormField>
          <FormField
            label="Name"
            htmlFor="type-name"
            required
            error={errors.name?.message}
          >
            <Input
              id="type-name"
              autoFocus
              invalid={Boolean(errors.name)}
              aria-describedby={errorId("type-name", errors.name?.message)}
              {...register("name")}
            />
          </FormField>
          <div className="sm:col-span-2">
            <FormField label="Description" htmlFor="type-description">
              <textarea
                id="type-description"
                className={textareaClassName}
                {...register("description")}
              />
            </FormField>
          </div>
          <Toggle
            label="Allows receiving"
            registration={register("allows_receiving")}
          />
          <Toggle
            label="Allows storage"
            registration={register("allows_storage")}
          />
          <Toggle
            label="Allows picking"
            registration={register("allows_picking")}
          />
          <Toggle
            label="Allows shipping"
            registration={register("allows_shipping")}
          />
          {editing ? (
            <Toggle
              label="Location type is active"
              registration={register("is_active")}
            />
          ) : null}
        </div>
        <FormActions
          busy={save.isPending}
          editing={editing}
          noun="location type"
        />
      </form>
    </DialogFrame>
  );
}

export function ZoneDialog({
  open,
  warehouseId,
  zone,
  onOpenChange,
}: {
  open: boolean;
  warehouseId: string;
  zone?: WarehouseZone;
  onOpenChange: (open: boolean) => void;
}) {
  const editing = zone !== undefined;
  const queryClient = useQueryClient();
  const {
    register,
    reset,
    handleSubmit,
    formState: { errors },
  } = useForm<ZoneFormValues>({
    resolver: zodResolver(zoneFormSchema),
    defaultValues: emptyZoneForm,
  });
  const save = useMutation({
    mutationFn: (values: ZoneFormValues) => {
      const common = {
        name: values.name.trim(),
        description: optional(values.description),
      };
      if (zone) {
        const request: UpdateZoneRequest = {
          ...common,
          is_active: values.is_active,
        };
        return updateZone(zone.zone_id, request);
      }
      const request: CreateZoneRequest = {
        code: values.code.trim().toUpperCase(),
        ...common,
      };
      return createZone(warehouseId, request);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: storageLayoutKeys.all });
      toast.success(editing ? "Zone updated." : "Zone created.");
      onOpenChange(false);
    },
  });

  useEffect(() => {
    if (!open) return;
    reset(
      zone
        ? {
            code: zone.code,
            name: zone.name,
            description: zone.description ?? "",
            is_active: zone.is_active,
          }
        : emptyZoneForm,
    );
  }, [open, reset, zone]);

  return (
    <DialogFrame
      open={open}
      busy={save.isPending}
      title={editing ? "Edit zone" : "Create zone"}
      description="Group warehouse locations into an operational area."
      onOpenChange={onOpenChange}
    >
      <form
        className="p-5 sm:p-6"
        noValidate
        onSubmit={handleSubmit((values) => save.mutate(values))}
      >
        {save.error ? (
          <p
            role="alert"
            className="mb-5 rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
          >
            {save.error.message}
          </p>
        ) : null}
        <div className="grid gap-5 sm:grid-cols-2">
          <FormField
            label="Code"
            htmlFor="zone-code"
            required
            error={errors.code?.message}
          >
            <Input
              id="zone-code"
              readOnly={editing}
              invalid={Boolean(errors.code)}
              {...register("code")}
            />
          </FormField>
          <FormField
            label="Name"
            htmlFor="zone-name"
            required
            error={errors.name?.message}
          >
            <Input
              id="zone-name"
              autoFocus
              invalid={Boolean(errors.name)}
              {...register("name")}
            />
          </FormField>
          <div className="sm:col-span-2">
            <FormField label="Description" htmlFor="zone-description">
              <textarea
                id="zone-description"
                className={textareaClassName}
                {...register("description")}
              />
            </FormField>
          </div>
          {editing ? (
            <Toggle
              label="Zone is active"
              registration={register("is_active")}
            />
          ) : null}
        </div>
        <FormActions busy={save.isPending} editing={editing} noun="zone" />
      </form>
    </DialogFrame>
  );
}

export function LocationDialog({
  open,
  warehouseId,
  location,
  zones,
  locationTypes,
  onOpenChange,
}: {
  open: boolean;
  warehouseId: string;
  location?: WarehouseLocation;
  zones: WarehouseZone[];
  locationTypes: LocationType[];
  onOpenChange: (open: boolean) => void;
}) {
  const editing = location !== undefined;
  const queryClient = useQueryClient();
  const {
    control,
    register,
    reset,
    handleSubmit,
    formState: { errors },
  } = useForm<LocationFormValues>({
    resolver: zodResolver(locationFormSchema),
    defaultValues: emptyLocationForm,
  });
  const save = useMutation({
    mutationFn: (values: LocationFormValues) => {
      const common = {
        zone_id: values.zone_id,
        location_type_id: values.location_type_id,
        barcode: optional(values.barcode),
        aisle: optional(values.aisle),
        bay: optional(values.bay),
        level_no: optional(values.level_no),
        position_no: optional(values.position_no),
        pick_sequence: values.pick_sequence,
        max_weight: optional(values.max_weight),
        max_volume: optional(values.max_volume),
        is_pick_face: values.is_pick_face,
      };
      if (location) {
        const request: UpdateLocationRequest = {
          ...common,
          is_locked: values.is_locked,
          is_active: values.is_active,
        };
        return updateLocation(location.location_id, request);
      }
      const request: CreateLocationRequest = {
        code: values.code.trim().toUpperCase(),
        ...common,
      };
      return createLocation(warehouseId, request);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: storageLayoutKeys.all });
      toast.success(editing ? "Location updated." : "Location created.");
      onOpenChange(false);
    },
  });

  useEffect(() => {
    if (!open) return;
    reset(
      location
        ? {
            zone_id: location.zone_id,
            location_type_id: location.location_type_id,
            code: location.code,
            barcode: location.barcode ?? "",
            aisle: location.aisle ?? "",
            bay: location.bay ?? "",
            level_no: location.level_no ?? "",
            position_no: location.position_no ?? "",
            pick_sequence: location.pick_sequence,
            max_weight: location.max_weight ?? "",
            max_volume: location.max_volume ?? "",
            is_pick_face: location.is_pick_face,
            is_locked: location.is_locked,
            is_active: location.is_active,
          }
        : emptyLocationForm,
    );
  }, [location, open, reset]);

  const zoneOptions = zones.map((item) => ({
    value: item.zone_id,
    label: `${item.name} (${item.code})`,
    disabled: !item.is_active && item.zone_id !== location?.zone_id,
  }));
  const typeOptions = locationTypes.map((item) => ({
    value: item.location_type_id,
    label: `${item.name} (${item.code})`,
    disabled:
      !item.is_active && item.location_type_id !== location?.location_type_id,
  }));

  return (
    <DialogFrame
      open={open}
      busy={save.isPending}
      title={editing ? "Edit location" : "Create location"}
      description="Define a scannable physical position inside the selected warehouse."
      onOpenChange={onOpenChange}
    >
      <form
        className="p-5 sm:p-6"
        noValidate
        onSubmit={handleSubmit((values) => save.mutate(values))}
      >
        {save.error ? (
          <p
            role="alert"
            className="mb-5 rounded-xl bg-rose-50 p-3 text-sm text-rose-900"
          >
            {save.error.message}
          </p>
        ) : null}
        <div className="grid gap-5 sm:grid-cols-2">
          <FormField
            label="Code"
            htmlFor="location-code"
            required
            error={errors.code?.message}
          >
            <Input
              id="location-code"
              readOnly={editing}
              invalid={Boolean(errors.code)}
              {...register("code")}
            />
          </FormField>
          <FormField
            label="Barcode"
            htmlFor="location-barcode"
            error={errors.barcode?.message}
          >
            <Input
              id="location-barcode"
              invalid={Boolean(errors.barcode)}
              {...register("barcode")}
            />
          </FormField>
          <FormField
            label="Zone"
            htmlFor="location-zone"
            required
            error={errors.zone_id?.message}
          >
            <Controller
              control={control}
              name="zone_id"
              render={({ field }) => (
                <Select
                  id="location-zone"
                  ariaLabel="Zone"
                  value={field.value}
                  onValueChange={field.onChange}
                  options={zoneOptions}
                  placeholder="Select a zone"
                  invalid={Boolean(errors.zone_id)}
                  className="mt-2"
                />
              )}
            />
          </FormField>
          <FormField
            label="Location type"
            htmlFor="location-type"
            required
            error={errors.location_type_id?.message}
          >
            <Controller
              control={control}
              name="location_type_id"
              render={({ field }) => (
                <Select
                  id="location-type"
                  ariaLabel="Location type"
                  value={field.value}
                  onValueChange={field.onChange}
                  options={typeOptions}
                  placeholder="Select a type"
                  invalid={Boolean(errors.location_type_id)}
                  className="mt-2"
                />
              )}
            />
          </FormField>
          <FormField
            label="Aisle"
            htmlFor="location-aisle"
            error={errors.aisle?.message}
          >
            <Input
              id="location-aisle"
              invalid={Boolean(errors.aisle)}
              {...register("aisle")}
            />
          </FormField>
          <FormField
            label="Bay"
            htmlFor="location-bay"
            error={errors.bay?.message}
          >
            <Input
              id="location-bay"
              invalid={Boolean(errors.bay)}
              {...register("bay")}
            />
          </FormField>
          <FormField
            label="Level"
            htmlFor="location-level"
            error={errors.level_no?.message}
          >
            <Input
              id="location-level"
              invalid={Boolean(errors.level_no)}
              {...register("level_no")}
            />
          </FormField>
          <FormField
            label="Position"
            htmlFor="location-position"
            error={errors.position_no?.message}
          >
            <Input
              id="location-position"
              invalid={Boolean(errors.position_no)}
              {...register("position_no")}
            />
          </FormField>
          <FormField
            label="Pick sequence"
            htmlFor="location-pick-sequence"
            error={errors.pick_sequence?.message}
          >
            <Input
              id="location-pick-sequence"
              type="number"
              min={0}
              invalid={Boolean(errors.pick_sequence)}
              {...register("pick_sequence", { valueAsNumber: true })}
            />
          </FormField>
          <FormField
            label="Maximum weight"
            htmlFor="location-max-weight"
            error={errors.max_weight?.message}
          >
            <Input
              id="location-max-weight"
              inputMode="decimal"
              invalid={Boolean(errors.max_weight)}
              {...register("max_weight")}
            />
          </FormField>
          <FormField
            label="Maximum volume"
            htmlFor="location-max-volume"
            error={errors.max_volume?.message}
          >
            <Input
              id="location-max-volume"
              inputMode="decimal"
              invalid={Boolean(errors.max_volume)}
              {...register("max_volume")}
            />
          </FormField>
          <Toggle label="Pick face" registration={register("is_pick_face")} />
          {editing ? (
            <Toggle
              label="Location is locked"
              registration={register("is_locked")}
            />
          ) : null}
          {editing ? (
            <Toggle
              label="Location is active"
              registration={register("is_active")}
            />
          ) : null}
        </div>
        <FormActions busy={save.isPending} editing={editing} noun="location" />
      </form>
    </DialogFrame>
  );
}
