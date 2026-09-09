# Outbound Part 2 API study

Part 2 continues a completed staging document through physical verification,
packing, dispatch, and store delivery:

```text
STAGED -> CHECKING -> CHECKED -> PACKING -> PACKED -> SHIPPED -> DELIVERED
                    \-> CHECK_FAILED
SHIPPED -> DELIVERY_FAILED -> PARTIALLY_DELIVERED | RETURNED
```

Every endpoint below requires:

```text
Authorization: Bearer <token>
Content-Type: application/json
```

The response body uses the common `success`, `message`, and `data` envelope.
IDs created by one step must be copied from that step's `data`; do not invent
document or line IDs.

## Current study records

The configured study database currently has:

```text
owner_id:            5e6a104a-51c2-4e94-8643-e339bcd5242f
warehouse_id:        59113433-3082-4208-87c1-330f4f56fcba
outbound_id:         OUT-STUDY_CUSTOMER-STUDY_WH-20260908-000001
staging_id:          STG-STUDY_WH-20260908-000001
packing_location_id: 7e446a79-9024-46d1-9c49-97fb4815f8e7
packing location:    STUDY_PACK_01
```

At the time this guide was generated, that staging document was still `OPEN`.
Complete it before starting the check:

```text
POST /api/v1/outbound/stagings/STG-STUDY_WH-20260908-000001/complete
```

If the study packing location is missing after rebuilding the database, rerun:

```powershell
.\scripts\seed-study-data.ps1
```

The seed is additive and idempotent.

## 1. Verify picked stock

Create the first check:

```http
POST /api/v1/outbound/checks
```

```json
{
  "staging_id": "STG-STUDY_WH-20260908-000001",
  "notes": "Outbound study verification"
}
```

The response contains an `outbound_check_id` and one or more lines. For every
line, send its `outbound_check_line_id` and copy `expected_qty` into
`checked_qty` for the normal case:

```http
POST /api/v1/outbound/checks/{outbound_check_id}/lines/{outbound_check_line_id}
```

```json
{
  "checked_qty": "12.000000",
  "result_code": "PASS"
}
```

Allowed results are `PASS`, `SHORT`, `OVER`, `WRONG_ITEM`, and
`DAMAGED`. All non-pass results require `notes`. Complete only after every
line has been recorded:

```text
POST /api/v1/outbound/checks/{outbound_check_id}/complete
```

A normal check becomes `PASSED` and the outbound order becomes `CHECKED`.
A failed check becomes `FAILED`, creates check-exception records, and moves
the order to `CHECK_FAILED`. Inspect them with:

```text
GET /api/v1/outbound/checks/{outbound_check_id}/exceptions
```

After the exception evidence is resolved, create a recheck by supplying both
the same staging ID and the failed check as `parent_check_id`.

## 2. Pack checked stock

Create one packing document from the passed check:

```http
POST /api/v1/outbound/packings
```

```json
{
  "outbound_check_id": "{outbound_check_id}",
  "packing_location_id": "7e446a79-9024-46d1-9c49-97fb4815f8e7"
}
```

The response lists every pack candidate. For each line, use its
`outbound_check_line_id` in the URL and its current
`source_balance_version` in the body:

```http
POST /api/v1/outbound/packings/{packing_id}/lines/{outbound_check_line_id}
```

```json
{
  "expected_balance_version": 13,
  "business_date": "2026-09-08",
  "operation_key": "study-pack-line-001"
}
```

`handling_unit_id` is optional. The operation key must be unique for the
logical stock movement; replaying the same payload is safe. Packing posts a
`PACK` movement from staging into `STUDY_PACK_01`.

After every candidate has a non-empty `packing_line_id`:

```text
POST /api/v1/outbound/packings/{packing_id}/complete
```

The packing document becomes `COMPLETED`; the order becomes `PACKED`.

## 3. Build and dispatch a shipment

One shipment can contain multiple completed packing documents, but only one
packing document per outbound order:

```http
POST /api/v1/outbound/shipments
```

```json
{
  "owner_id": "5e6a104a-51c2-4e94-8643-e339bcd5242f",
  "warehouse_id": "59113433-3082-4208-87c1-330f4f56fcba",
  "business_date": "2026-09-08",
  "packing_ids": ["{packing_id}"],
  "route_reference": "JKT-STUDY-01",
  "tracking_number": "TRACK-STUDY-001",
  "vehicle_number": "B 1234 WMS",
  "seal_number": "SEAL-STUDY-001",
  "notes": "Outbound Phase 2 study"
}
```

Dispatch each candidate using its `packing_line_id` and returned balance
version:

```http
POST /api/v1/outbound/shipments/{shipment_id}/lines/{packing_line_id}/dispatch
```

```json
{
  "expected_balance_version": 1,
  "business_date": "2026-09-08",
  "operation_key": "study-ship-line-001"
}
```

Dispatch posts a `SHIP` movement and removes the quantity from warehouse
on-hand stock. After all lines are dispatched:

```text
POST /api/v1/outbound/shipments/{shipment_id}/complete
```

The shipment and outbound order become `SHIPPED`.

## 4. Deliver and record proof

Create a delivery for one outbound order in the shipped manifest:

```http
POST /api/v1/outbound/deliveries
```

```json
{
  "shipment_id": "{shipment_id}",
  "outbound_id": "OUT-STUDY_CUSTOMER-STUDY_WH-20260908-000001",
  "business_date": "2026-09-08",
  "planned_delivery_at": "2026-09-09T13:00:00+07:00",
  "notes": "Deliver to study store"
}
```

Record departure and arrival in sequence:

```http
POST /api/v1/outbound/deliveries/{delivery_id}/depart

{
  "event_at": "2026-09-09T11:00:00+07:00",
  "notes": "Vehicle left warehouse"
}
```

```http
POST /api/v1/outbound/deliveries/{delivery_id}/arrive

{
  "event_at": "2026-09-09T12:30:00+07:00",
  "latitude": "-6.200000",
  "longitude": "106.816666",
  "notes": "Vehicle arrived at store"
}
```

For each delivery line, record the delivered quantity and proof:

```http
POST /api/v1/outbound/deliveries/{delivery_id}/lines/{delivery_line_id}/deliver
```

```json
{
  "delivered_qty": "12.000000",
  "event_at": "2026-09-09T12:40:00+07:00",
  "recipient_name": "Store Receiver",
  "recipient_reference": "EMP-001",
  "proof_reference": "POD-STUDY-001",
  "proof_uri": "https://example.invalid/study/pod-001",
  "latitude": "-6.200000",
  "longitude": "106.816666",
  "notes": "Boxes received in good condition"
}
```

Partial line confirmations are allowed. Once all planned quantities are
delivered:

```http
POST /api/v1/outbound/deliveries/{delivery_id}/complete

{
  "delivered_at": "2026-09-09T12:40:00+07:00"
}
```

The final delivery and order status is `DELIVERED`.

## Failed delivery and return to depot

Instead of completing the delivery, record a failed attempt while it is
`IN_TRANSIT` or `ARRIVED`:

```http
POST /api/v1/outbound/deliveries/{delivery_id}/fail

{
  "reason_code": "RECIPIENT_REJECTED",
  "event_at": "2026-09-09T12:45:00+07:00",
  "notes": "Store rejected the remaining quantity"
}
```

Failure reasons are `STORE_CLOSED`, `RECIPIENT_REJECTED`,
`ADDRESS_NOT_FOUND`, `DAMAGED_IN_TRANSIT`, `VEHICLE_ISSUE`, and `OTHER`.

Before posting a return, configure its controlled location and a non-allocatable
inventory status such as `HOLD`. Obtain its UUID from
`GET /api/v1/master/inventory-statuses`:

```http
PUT /api/v1/outbound/return-policy

{
  "owner_id": "5e6a104a-51c2-4e94-8643-e339bcd5242f",
  "warehouse_id": "59113433-3082-4208-87c1-330f4f56fcba",
  "return_location_id": "1719de47-52ab-4168-872d-48e2824a0f4d",
  "return_inventory_status_id": "{HOLD_inventory_status_id}",
  "is_active": true
}
```

Return each undelivered remainder:

```http
POST /api/v1/outbound/deliveries/{delivery_id}/lines/{delivery_line_id}/return

{
  "returned_qty": "11.000000",
  "event_at": "2026-09-09T14:00:00+07:00",
  "operation_key": "study-delivery-return-001",
  "notes": "Undelivered stock returned to depot"
}
```

This creates `DELIVERY_RETURN` movement evidence and places the goods in the
configured non-allocatable balance. Close after every line remainder is either
delivered or returned:

```http
POST /api/v1/outbound/deliveries/{delivery_id}/close-return

{
  "event_at": "2026-09-09T14:05:00+07:00",
  "notes": "Vehicle and stock returned"
}
```

The result is `RETURNED` when nothing was delivered, otherwise
`PARTIALLY_DELIVERED`.

## Useful reads and concurrency

```text
GET /api/v1/outbound/checks?owner_id={owner_uuid}&warehouse_id={warehouse_uuid}
GET /api/v1/outbound/packings?owner_id={owner_uuid}&warehouse_id={warehouse_uuid}
GET /api/v1/outbound/shipments?owner_id={owner_uuid}&warehouse_id={warehouse_uuid}
GET /api/v1/outbound/deliveries?owner_id={owner_uuid}&warehouse_id={warehouse_uuid}
GET /api/v1/outbound/return-policy?owner_id={owner_uuid}&warehouse_id={warehouse_uuid}
```

If packing or shipping returns `409 Conflict`, GET that document again and
copy the newly returned `source_balance_version`. A stale version means
another stock operation changed the same balance after it was read.
