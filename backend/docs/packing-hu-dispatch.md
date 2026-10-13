# Packing and handling-unit dispatch

UI: **Outbound → Packing** and **Outbound → Shipments / HU dispatch**.

Both menus require `OUTBOUND.READ` and an accessible warehouse + active served owner. Packing writes require `OUTBOUND.PACK`; shipment/driver/dispatch writes require `OUTBOUND.SHIP`; cancellation requires `OUTBOUND.CANCEL`. In-form empty-HU creation uses the existing `INVENTORY.IDENTITY` permission. Listing available HUs uses `INVENTORY.READ`.

## Packing

Create from a passed outbound check that belongs to a CHECKED order and still needs packing. Use an active, unlocked location of the active `PACKING` type. The server rechecks location eligibility when posting each line.

Each checked line is posted in full in its base UOM. It can remain loose or enter an open, unnested HU at the packing area. Existing HU stock retains its identity and must be moved in full. A loaded container elsewhere cannot be moved by attaching one loose line. Container selection is scoped to the same owner and warehouse; barcodes are displayed on packing and shipment lines.

Packing records PACK movements and updates serial identity; it does not deduct warehouse stock. Completion requires every positive passed-check line to be posted. Cancellation is possible only before any line is posted.

Serialized lines follow the exact serial recorded on their originating pick/pack movement. This preserves individual serial lineage even when several units share a pooled staging or packed-HU balance. Serial numbers are shown on the document lines.

## Shipment and dispatch

Create a PLANNED manifest from one or more completed, unassigned packings in the same scope. Vehicle, seal, tracking, route, notes and carrier service are optional. A carrier-service shipment requires an active primary driver before completion; self-managed transport does not.

Dispatch loose stock by packing line. For HUs use:

`POST /api/v1/outbound/shipments/:id/handling-units/:hu_id/dispatch`

```json
{
  "business_date": "2026-10-13",
  "operation_key": "stable-client-request-key",
  "expected_balances": [
    {"balance_id": "BAL-1", "expected_version": 4},
    {"balance_id": "BAL-2", "expected_version": 7}
  ]
}
```

Supply one version per distinct positive stock balance, even when multiple packing lines share that balance. The server locks the shipment, HU and source balances. All positive HU stock must be present on this manifest, at the HU's current location, with matching owner/warehouse and no reserved quantity. Already-dispatched lines, stale versions, omitted stock, stock assigned to another shipment and nested HUs are rejected. Grand-opname stock freezes still apply.

All shipment lines, SHIP movements, serial-state changes, quantity deductions and HU location clearing commit together. Any error rolls everything back. Retrying the same operation key and manifest/date is idempotent; changing its meaning conflicts. Dispatch deducts stock immediately; completion then requires every manifest line to have been posted and advances the document to SHIPPED. Cancellation is only possible before any dispatch.

The existing delivery-line return API cannot split a multi-line HU; those returns are rejected until a whole-HU return transaction is implemented. Single-line HU and loose-stock return behavior is unchanged.

Candidate endpoints (`GET /packing-candidates`, `GET /shipment-candidates`) require `owner_id` and `warehouse_id`, support normal paging/search filters, and exclude already-consumed documents. The UI loads all pages for selectors.

No schema migration or reseeding is required. Restart the API to load the new routes and validation rules. Upstream allocation, picking, staging and checking continue to use their existing APIs; this change adds only the packing and shipment UI.
