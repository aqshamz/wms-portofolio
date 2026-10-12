"use client";

import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { addFoundStock, getFoundStockOptions } from "./stock-control-api";
import type { CycleCount } from "./stock-control-types";

export function FoundStockForm({
  count,
  onAdded,
  onCancel,
}: {
  count: Pick<CycleCount, "cycle_count_id" | "version_no">;
  onAdded: (updated: CycleCount) => void;
  onCancel: () => void;
}) {
  const [itemId, setItemId] = useState("");
  const [locationId, setLocationId] = useState("");
  const [statusId, setStatusId] = useState("");
  const [lotId, setLotId] = useState("none");
  const [huId, setHuId] = useState("none");
  const [quantity, setQuantity] = useState("");
  const [notes, setNotes] = useState("");
  const [itemSearch, setItemSearch] = useState("");
  const [locationSearch, setLocationSearch] = useState("");
  const [lotSearch, setLotSearch] = useState("");
  const [huSearch, setHuSearch] = useState("");
  const filters = {
    item_id: itemId,
    location_id: locationId,
    item_search: itemSearch,
    location_search: locationSearch,
    lot_search: lotSearch,
    hu_search: huSearch,
  };
  const options = useQuery({
    queryKey: [
      "stock-control",
      "found-stock-options",
      count.cycle_count_id,
      filters,
    ],
    queryFn: () => getFoundStockOptions(count.cycle_count_id, filters),
  });
  const item = options.data?.items.find((entry) => entry.item_id === itemId);
  const selectedStatus =
    statusId ||
    options.data?.statuses.find((entry) => entry.label === "AVAILABLE")?.id ||
    "";
  const add = useMutation({
    mutationFn: () =>
      addFoundStock(count.cycle_count_id, {
        expected_version: count.version_no,
        item_id: itemId,
        location_id: locationId,
        inventory_status_id: selectedStatus,
        lot_id: item?.lot_controlled && lotId !== "none" ? lotId : undefined,
        handling_unit_id: huId !== "none" ? huId : undefined,
        counted_quantity: quantity.trim(),
        notes: notes.trim(),
      }),
    onSuccess: (updated) => {
      onAdded(updated);
      toast.success(
        "Found stock recorded. Inventory changes only after approval.",
      );
    },
    onError: (error) => toast.error(error.message),
  });
  const validQuantity =
    /^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$/.test(quantity.trim()) &&
    Number(quantity) > 0 &&
    (!item?.serial_controlled ||
      (Number.isInteger(Number(quantity)) && Number(quantity) <= 5000));
  const canSubmit =
    !add.isPending &&
    !options.isFetching &&
    !!item &&
    !!locationId &&
    !!selectedStatus &&
    validQuantity &&
    !!notes.trim() &&
    (!item.lot_controlled || lotId !== "none");

  return (
    <form
      className="space-y-4 rounded-xl border border-cyan-200 bg-cyan-50 p-4"
      onSubmit={(event) => {
        event.preventDefault();
        if (canSubmit) add.mutate();
      }}
    >
      <div>
        <p className="font-semibold text-cyan-950">Record found stock</p>
        <p className="mt-1 text-sm text-cyan-900">
          Record stock with no positive balance at this location. If it already
          has a count line, use that line. Quantities use the item’s base unit;
          a recount is required when the variance exceeds this document’s
          tolerance. Approval creates the inventory balance and movement.
        </p>
      </div>
      <fieldset disabled={add.isPending} className="grid gap-4 sm:grid-cols-2">
        <FormField label="Search items" htmlFor="found-item-search">
          <Input
            id="found-item-search"
            value={itemSearch}
            maxLength={120}
            onChange={(event) => setItemSearch(event.target.value)}
            placeholder="Item code or name"
          />
        </FormField>
        <FormField label="Search locations" htmlFor="found-location-search">
          <Input
            id="found-location-search"
            value={locationSearch}
            maxLength={120}
            onChange={(event) => setLocationSearch(event.target.value)}
            placeholder="Location code"
          />
        </FormField>
        <FormField label="Item" htmlFor="found-item" required>
          <Select
            id="found-item"
            ariaLabel="Found stock item"
            value={itemId}
            options={
              options.data?.items.map((entry) => ({
                value: entry.item_id,
                label: `${entry.code} · ${entry.name}`,
              })) ?? []
            }
            placeholder="Select item"
            onValueChange={(value) => {
              setItemId(value);
              setQuantity("");
              setLotId("none");
              setLotSearch("");
            }}
          />
        </FormField>
        <FormField label="Found location" htmlFor="found-location" required>
          <Select
            id="found-location"
            ariaLabel="Found stock location"
            value={locationId}
            options={
              options.data?.locations.map((entry) => ({
                value: entry.id,
                label: entry.label,
              })) ?? []
            }
            placeholder="Select storage location"
            onValueChange={(value) => {
              setLocationId(value);
              setHuId("none");
              setHuSearch("");
            }}
          />
        </FormField>
        <FormField label="Inventory status" htmlFor="found-status" required>
          <Select
            id="found-status"
            ariaLabel="Found stock inventory status"
            value={selectedStatus}
            options={
              options.data?.statuses.map((entry) => ({
                value: entry.id,
                label: entry.label,
              })) ?? []
            }
            placeholder="Select stock condition"
            onValueChange={setStatusId}
          />
        </FormField>
        <FormField
          label={`Physical quantity${item ? ` (${item.uom_code})` : ""}`}
          htmlFor="found-quantity"
          required
        >
          <Input
            id="found-quantity"
            value={quantity}
            inputMode="decimal"
            onChange={(event) => setQuantity(event.target.value)}
          />
        </FormField>
        {item?.lot_controlled ? (
          <>
            <FormField label="Search lots" htmlFor="found-lot-search">
              <Input
                id="found-lot-search"
                value={lotSearch}
                maxLength={120}
                onChange={(event) => {
                  setLotSearch(event.target.value);
                  setLotId("none");
                }}
                placeholder="Lot number"
              />
            </FormField>
            <FormField label="Lot" htmlFor="found-lot" required>
              <Select
                id="found-lot"
                ariaLabel="Found stock lot"
                value={lotId}
                options={[
                  { value: "none", label: "Select lot" },
                  ...(options.data?.lots.map((entry) => ({
                    value: entry.id,
                    label: entry.label,
                  })) ?? []),
                ]}
                onValueChange={setLotId}
              />
              <p className="mt-1 text-xs text-cyan-900">
                If the lot is new, register it in Inventory → Lots first.
              </p>
            </FormField>
          </>
        ) : null}
        <FormField label="Search handling units" htmlFor="found-hu-search">
          <Input
            id="found-hu-search"
            value={huSearch}
            maxLength={120}
            onChange={(event) => {
              setHuSearch(event.target.value);
              setHuId("none");
            }}
            placeholder="Handling unit barcode"
            disabled={!locationId}
          />
        </FormField>
        <FormField label="Handling unit" htmlFor="found-hu">
          <Select
            id="found-hu"
            ariaLabel="Found stock handling unit"
            value={huId}
            options={[
              { value: "none", label: "Loose stock / no handling unit" },
              ...(options.data?.handling_units.map((entry) => ({
                value: entry.id,
                label: entry.label,
              })) ?? []),
            ]}
            onValueChange={setHuId}
            disabled={!locationId}
          />
        </FormField>
        <div className="sm:col-span-2">
          <FormField label="Discovery notes" htmlFor="found-notes" required>
            <Textarea
              id="found-notes"
              value={notes}
              maxLength={2000}
              onChange={(event) => setNotes(event.target.value)}
              placeholder="Where the stock was found and why it is not recorded"
            />
          </FormField>
        </div>
      </fieldset>
      {item?.serial_controlled ? (
        <p className="text-sm text-cyan-950">
          Enter a whole number of base units. The reviewer must identify every
          found serial through serial reconciliation before stock can post.
        </p>
      ) : null}
      {options.error || add.error ? (
        <p
          role="alert"
          className="rounded-lg bg-rose-50 p-3 text-sm text-rose-900"
        >
          {add.error?.message ?? options.error?.message}
        </p>
      ) : null}
      <div className="flex justify-end gap-3">
        <Button
          type="button"
          variant="secondary"
          disabled={add.isPending}
          onClick={onCancel}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={!canSubmit}>
          Record for review
        </Button>
      </div>
    </form>
  );
}
