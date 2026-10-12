# Client Delivery Orders

UI: **Outbound → Delivery orders** (`/outbound/orders`). The first step covers scoped listing, creation, draft header/line edits, draft line deletion, validation, release and cancellation. Allocation/waves/picking are separate subsequent steps.

Permissions: `OUTBOUND.READ` to view, `OUTBOUND.PLAN` to create/edit/validate/release, and `OUTBOUND.CANCEL` to cancel. API owner/warehouse scope checks still apply. Select an accessible warehouse and an active served owner; partners and items belong to that owner.

## Quantity contract

Each create/add/update line accepts `ordered_qty` and optional `uom_id`. Base UOM is the default for legacy create/add clients. Delivery Orders allow the base UOM plus all active UOMs with active conversions configured on the item, regardless of their receiving/picking flags. Unrelated or inactive UOMs/conversions cannot be selected for a new snapshot. Serialized demand may be in cartons, but the converted base quantity must be a whole number.

The server captures:

- `uom_id`: document/order UOM.
- `ordered_qty`: client quantity in that UOM.
- `uom_conversion_to_base`: frozen conversion factor.
- `base_uom_id`: frozen item base UOM.
- `ordered_base_qty`: exact product, representable as numeric(20,6).

Responses also include `uom_code` and `base_uom_code`. Allocated, picked, checked, packed, shipped and delivered quantities remain in **base units**, as do reservations, movements and check-exception resolutions.

Example: 2 CTN × 12 EA/CTN = 24 EA. Allocation reserves 24 EA, not 2 EA.

Changing the catalog factor does not rewrite a document. Editing a draft quantity while retaining or omitting its existing UOM preserves its factor. Explicitly switching to a different UOM snapshots that UOM's current configured factor. If the item's base UOM changes, the old document is rejected rather than reinterpreted.

Migration **27** backfills legacy outbound lines at factor 1 with their original UOM and quantity, since the previous API accepted base-UOM-only orders. Existing allocations and history are preserved. Restart the backend to apply its versioned migration; no manual reseeding is required.

## Lifecycle

Draft edits use `expected_version`; stale writes fail. Header customer/owner/warehouse/client DO number and business date are immutable after creation. Ship-to address is captured on the document. Validate returns a validation run; refresh the order to get its updated version. Failed validation remains draft and displays rule results. Release is only available after validation. Cancelled documents remain in audit history; there is no destructive header-delete action.

Creating, validating or releasing a DO does **not** move or deduct stock. Stock reservation begins at allocation and deduction happens at dispatch. The current UI stops before allocation.

Requested ship time is entered and displayed in the user's device timezone, labelled accordingly; requests carry an explicit timezone via ISO timestamp.
